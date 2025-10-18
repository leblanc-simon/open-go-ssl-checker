package scheduler

import (
	"log"
	"time"

	"leblanc.io/open-go-ssl-checker/internal/checker"
)

type PeriodicChecker struct {
	cs       *checker.CertificateService
	interval time.Duration
	stopChan chan struct{}
}

func NewPeriodicChecker(cs *checker.CertificateService, interval time.Duration) *PeriodicChecker {
	return &PeriodicChecker{
		cs:       cs,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start launches the periodic checking goroutine.
func (pc *PeriodicChecker) Start() {
	log.Printf("Starting periodic checker with interval %v.", pc.interval)
	ticker := time.NewTicker(pc.interval)

	go func() {
		// Run an initial check immediately at startup (optional)
		log.Println("Running first set of periodic checks at startup...")
		pc.runChecks()

		for {
			select {
			case <-ticker.C:
				log.Println("Triggering periodic certificate checks...")
				pc.runChecks()
			case <-pc.stopChan:
				ticker.Stop()
				log.Println("Periodic checker stopped.")

				return
			}
		}
	}()
}

// Stop stops the periodic checking goroutine.
func (pc *PeriodicChecker) Stop() {
	close(pc.stopChan)
}

// RunOnce triggers a single execution of all certificate checks immediately.
func (pc *PeriodicChecker) RunOnce() {
	pc.runChecks()
}

// runChecks retrieves all projects and triggers verification for each.
func (pc *PeriodicChecker) runChecks() {
	log.Println("Starting periodic check series.")

	projects, err := pc.cs.Store.ListProjects()
	if err != nil {
		log.Printf(
			"Error retrieving projects for periodic check: %v",
			err,
		)

		return
	}

	if len(projects) == 0 {
		log.Println("No projects to check periodically.")

		return
	}

	log.Printf("Checking %d project(s)...", len(projects))

	for _, project := range projects {
		log.Printf(
			"Periodic check for: %s (ID: %s, Host: %s:%s)",
			project.Name,
			project.ID,
			project.Host,
			project.Port,
		)
		pc.cs.CheckAndStoreCertificate(
			project.ID,
			project.Host,
			project.Port,
			project.Type,
			project.AllowInsecure,
		)
	}

	log.Println("Periodic check series completed.")
}
