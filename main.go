package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ilyakaznacheev/cleanenv"
	"leblanc.io/open-go-ssl-checker/internal/checker"
	"leblanc.io/open-go-ssl-checker/internal/config"
	"leblanc.io/open-go-ssl-checker/internal/handlers"
	"leblanc.io/open-go-ssl-checker/internal/middleware"
	"leblanc.io/open-go-ssl-checker/internal/scheduler"
	"leblanc.io/open-go-ssl-checker/internal/store"
	"leblanc.io/open-go-ssl-checker/internal/websocket"
)

//go:embed all:static
var staticFs embed.FS

type args struct {
	ConfigPath string
}

var cfg config.Config

var (
	version = "develop"
	appName = "OpenGoSSLChecker"
)

const (
	defaultPeriodicCheckInterval = 24 * time.Hour
)

func showVersion() {
	fmt.Fprintf(flag.CommandLine.Output(), "%s (%s)\n", appName, version)
}

func main() {
	loadConfig(&cfg)

	// Initialize the Store (Database)
	dbStore, err := store.NewStore(cfg.Database.Driver, cfg.Database.Dsn)
	if err != nil {
		log.Fatalf("Error initializing store: %v", err)
	}
	defer dbStore.Close()

	if err := dbStore.InitSchema(); err != nil {
		log.Fatalf("Error initializing DB schema: %v", err)
	}

	log.Println("Database schema initialized/verified.")

	wsHub := websocket.NewHub(dbStore)
	go wsHub.Run() // Start the hub in a goroutine
	log.Println("WebSocket hub started.")

	// Initialize the certificate checking service
	certCheckerService := checker.NewCertificateService(dbStore, wsHub)

	periodicCertChecker := scheduler.NewPeriodicChecker(
		certCheckerService,
		defaultPeriodicCheckInterval,
	)
	periodicCertChecker.Start()
	defer periodicCertChecker.Stop()

	// Initialize the application context
	appCtx := &handlers.AppContext{
		Store:   dbStore,
		Checker: certCheckerService,
	}

	// Configure the routes
	router := mux.NewRouter()
	router.Handle("/", middleware.LinkMiddleware(http.HandlerFunc(appCtx.IndexHandler))).
		Methods("GET")
	router.HandleFunc("/add", appCtx.AddProjectHandler).Methods("GET", "POST")
	router.Handle("/projects", middleware.LinkMiddleware(http.HandlerFunc(appCtx.ProjectsHandler))).
		Methods("GET")
	router.HandleFunc("/delete/{uuid}", appCtx.DeleteProjectHandler).Methods("POST")
	router.Handle("/history/{uuid}", middleware.LinkMiddleware(http.HandlerFunc(appCtx.HistoryHandler))).
		Methods("GET")
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHub.ServeWs(w, r)
	})

	router.PathPrefix("/static/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, staticFs, strings.TrimLeft(r.RequestURI, "/"))
	}).Methods("GET")

	// Start the web server
	serverPort := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))

	log.Printf("Starting server on %s", serverPort)

	if err := http.ListenAndServe(serverPort, router); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}

func loadConfig(cfg *config.Config) {
	args := processArgs(&cfg)
	// read configuration from the file and environment variables
	if _, err := os.Stat(args.ConfigPath); errors.Is(err, os.ErrNotExist) {
		if err := cleanenv.ReadEnv(cfg); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	} else {
		if err := cleanenv.ReadConfig(args.ConfigPath, &cfg); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}
}

func processArgs(cfg interface{}) args {
	var arguments args

	flag := flag.NewFlagSet(appName, 1)

	flag.StringVar(&arguments.ConfigPath, "c", "config.yaml", "Path to configuration file")
	versionFlag := flag.Bool("version", false, "Show version")

	fu := flag.Usage
	flag.Usage = func() {
		fu()

		envHelp, _ := cleanenv.GetDescription(cfg, nil)

		fmt.Fprintln(flag.Output())
		fmt.Fprintln(flag.Output(), envHelp)
	}

	flag.Parse(os.Args[1:])

	if *versionFlag {
		showVersion()
		os.Exit(0)
	}

	return arguments
}
