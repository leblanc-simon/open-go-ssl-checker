package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"leblanc.io/open-go-ssl-checker/internal/template"
	"leblanc.io/open-go-ssl-checker/internal/types"
)

func (ac *AppContext) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)

		return
	}

	summaries, err := ac.Store.GetLatestChecksSummary()
	if err != nil {
		log.Printf("IndexHandler error - GetLatestChecksSummary: %v", err)
		http.Error(
			w,
			"Unable to retrieve project information.",
			http.StatusInternalServerError,
		)

		return
	}

	template.Execute(w, "index", r.Header.Get("Accept-Language"), summaries)
}

func (ac *AppContext) AddProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		template.Execute(w, "add", r.Header.Get("Accept-Language"), "nil")

		return
	}

	name := r.FormValue("name")
	host := r.FormValue("host")
	port := r.FormValue("port")
	projectType := r.FormValue("type")
	allowInsecure := r.FormValue("allow_insecure") == "true"

	if name == "" || host == "" || port == "" || projectType == "" {
		http.Error(w, "All fields are required.", http.StatusBadRequest)

		return
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		http.Error(w, "Port must be an integer.", http.StatusBadRequest)

		return
	}

	if portInt < 1 || portInt > 65535 {
		http.Error(w, "Port must be between 1 and 65535.", http.StatusBadRequest)

		return
	}

	project := types.Project{
		ID:            uuid.New().String(),
		Name:          name,
		Host:          host,
		Port:          port,
		Type:          projectType,
		AllowInsecure: allowInsecure,
	}

	if err := ac.Store.AddProject(project); err != nil {
		log.Printf("AddProjectHandler error - AddProject: %v", err)
		http.Error(w, "Unable to add project.", http.StatusInternalServerError)

		return
	}

	go ac.Checker.CheckAndStoreCertificate(
		project.ID,
		project.Host,
		project.Port,
		project.Type,
		project.AllowInsecure,
	)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (ac *AppContext) ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	projects, err := ac.Store.ListProjects()
	if err != nil {
		log.Printf("ProjectsHandler error - ListProjects: %v", err)
		http.Error(
			w,
			"Unable to retrieve projects list.",
			http.StatusInternalServerError,
		)

		return
	}

	template.Execute(w, "projects", r.Header.Get("Accept-Language"), projects)
}

func (ac *AppContext) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["uuid"]

	if projectID == "" {
		http.Error(w, "ID de projet manquant.", http.StatusBadRequest)

		return
	}

	// Vérifier si le projet existe avant de tenter de le supprimer (optionnel mais propre)
	proj, err := ac.Store.GetProject(projectID)
	if err != nil {
		log.Printf("Erreur DeleteProjectHandler - GetProject %s: %v", projectID, err)
		http.Error(w, "Erreur lors de la vérification du projet.", http.StatusInternalServerError)

		return
	}

	if proj == nil {
		http.NotFound(w, r)

		return
	}

	if err := ac.Store.DeleteProject(projectID); err != nil {
		log.Printf("DeleteProjectHandler error - DeleteProject %s: %v", projectID, err)
		http.Error(w, "Unable to delete project.", http.StatusInternalServerError)

		return
	}

	log.Printf("Project %s successfully deleted.", projectID)
	http.Redirect(w, r, "/projects", http.StatusSeeOther)
}
