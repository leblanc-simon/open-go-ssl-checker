package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"leblanc.io/open-go-ssl-checker/internal/types"
)

// AddProjectAPIHandler handles POST /api/projects
// It expects a JSON body with fields: name, host, port (int), type, allow_insecure (bool)
// Authentication: header X-API-Key must match the configured API key.
func (ac *AppContext) AddProjectAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	// API key check
	apiKey := r.Header.Get("X-API-Key")
	if ac.ApiKey == "" || apiKey == "" || apiKey != ac.ApiKey {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	type addProjectRequest struct {
		Name          string `json:"name"`
		Host          string `json:"host"`
		Port          int    `json:"port"`
		Type          string `json:"type"`
		AllowInsecure bool   `json:"allow_insecure"`
	}

	var req addProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	if req.Name == "" || req.Host == "" || req.Port == 0 || req.Type == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "name, host, port and type are required"})
		return
	}

	if req.Port < 1 || req.Port > 65535 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "port must be between 1 and 65535"})
		return
	}

	project := types.Project{
		ID:            uuid.New().String(),
		Name:          req.Name,
		Host:          req.Host,
		Port:          strconv.Itoa(req.Port),
		Type:          req.Type,
		AllowInsecure: req.AllowInsecure,
	}

	if err := ac.Store.AddProject(project); err != nil {
		log.Printf("AddProjectAPIHandler error - AddProject: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unable to add project"})
		return
	}

	// Trigger immediate check in background
	go ac.Checker.CheckAndStoreCertificate(
		project.ID,
		project.Host,
		project.Port,
		project.Type,
		project.AllowInsecure,
	)

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"id":     project.ID,
		"status": "created",
	})
}
