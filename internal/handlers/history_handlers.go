package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"leblanc.io/open-go-ssl-checker/internal/template"
	"leblanc.io/open-go-ssl-checker/internal/types"
)

func (ac *AppContext) HistoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["uuid"]

	if projectID == "" {
		http.Error(w, "Missing project ID.", http.StatusBadRequest)

		return
	}

	project, err := ac.Store.GetProject(projectID)
	if err != nil {
		log.Printf("HistoryHandler error - GetProject %s: %v", projectID, err)
		http.Error(w, "Error retrieving project.", http.StatusInternalServerError)

		return
	}

	if project == nil {
		http.NotFound(w, r)

		return
	}

	checks, err := ac.Store.GetCertificateChecksForProject(projectID)
	if err != nil {
		log.Printf("HistoryHandler error - GetCertificateChecksForProject %s: %v", projectID, err)
		http.Error(
			w,
			"Unable to retrieve verification history.",
			http.StatusInternalServerError,
		)

		return
	}

	// Pour passer le nom du projet au template, on peut l'encapsuler
	data := struct {
		ProjectName string
		Checks      []types.CertificateCheck
	}{
		ProjectName: project.Name,
		Checks:      checks,
	}

	template.Execute(w, "history", r.Header.Get("Accept-Language"), data)
}
