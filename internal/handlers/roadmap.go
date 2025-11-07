package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"roadmap-visualizer/internal/models"
	"roadmap-visualizer/internal/parser"
	"roadmap-visualizer/internal/storage"
	"strings"
)

// RoadmapHandler handles roadmap-related HTTP requests
type RoadmapHandler struct {
	storage *storage.FileStorage
}

// NewRoadmapHandler creates a new roadmap handler
func NewRoadmapHandler(storage *storage.FileStorage) *RoadmapHandler {
	return &RoadmapHandler{
		storage: storage,
	}
}

// CreateRoadmap handles POST /api/roadmaps
func (h *RoadmapHandler) CreateRoadmap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse YAML
	roadmap, err := parser.ParseRoadmap(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid roadmap: %v", err), http.StatusBadRequest)
		return
	}

	// Store roadmap
	fileName := "uploaded.yaml"
	if fileNameHeader := r.Header.Get("X-File-Name"); fileNameHeader != "" {
		fileName = fileNameHeader
	}

	stored, err := h.storage.Create(roadmap, fileName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to store roadmap: %v", err), http.StatusInternalServerError)
		return
	}

	// Return created roadmap
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(stored)
}

// CreateMultipleRoadmaps handles POST /api/roadmaps/batch
// This endpoint parses files with multiple roadmap documents separated by ---
func (h *RoadmapHandler) CreateMultipleRoadmaps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse multiple roadmaps from YAML
	roadmaps, err := parser.ParseMultipleRoadmaps(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid roadmap file: %v", err), http.StatusBadRequest)
		return
	}

	// Get base filename from header
	baseFileName := "uploaded.yaml"
	if fileNameHeader := r.Header.Get("X-File-Name"); fileNameHeader != "" {
		baseFileName = fileNameHeader
	}

	// Store each roadmap
	var storedRoadmaps []interface{}
	for i, roadmap := range roadmaps {
		// Create unique filename for each roadmap
		fileName := fmt.Sprintf("%s-part%d.yaml", strings.TrimSuffix(baseFileName, ".yaml"), i+1)

		stored, err := h.storage.Create(roadmap, fileName)
		if err != nil {
			// If we fail partway through, we've already stored some roadmaps
			// Return an error but also include what was stored
			http.Error(w, fmt.Sprintf("Failed to store roadmap %d (%s): %v", i+1, roadmap.Name, err), http.StatusInternalServerError)
			return
		}
		storedRoadmaps = append(storedRoadmaps, stored)
	}

	// Return all created roadmaps
	response := map[string]interface{}{
		"count":    len(storedRoadmaps),
		"roadmaps": storedRoadmaps,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ListRoadmaps handles GET /api/roadmaps
func (h *RoadmapHandler) ListRoadmaps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roadmaps, err := h.storage.List()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list roadmaps: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roadmaps)
}

// GetRoadmap handles GET /api/roadmaps/{id}
func (h *RoadmapHandler) GetRoadmap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/roadmaps/")
	if id == "" || strings.Contains(id, "/") {
		http.Error(w, "Invalid roadmap ID", http.StatusBadRequest)
		return
	}

	stored, err := h.storage.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Roadmap not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get roadmap: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stored)
}

// DeleteRoadmap handles DELETE /api/roadmaps/{id}
func (h *RoadmapHandler) DeleteRoadmap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/roadmaps/")
	if id == "" || strings.Contains(id, "/") {
		http.Error(w, "Invalid roadmap ID", http.StatusBadRequest)
		return
	}

	err := h.storage.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Roadmap not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete roadmap: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetRoadmapDependencies handles GET /api/roadmaps/{id}/dependencies
// Returns all external dependencies for items in the roadmap
func (h *RoadmapHandler) GetRoadmapDependencies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/roadmaps/")
	id = strings.TrimSuffix(id, "/dependencies")
	if id == "" || strings.Contains(id, "/") {
		http.Error(w, "Invalid roadmap ID", http.StatusBadRequest)
		return
	}

	stored, err := h.storage.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Roadmap not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get roadmap: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Collect all external dependencies
	type DependencyInfo struct {
		ItemID               string                       `json:"item_id"`
		ItemName             string                       `json:"item_name"`
		ExternalDependencies []models.ExternalDependency `json:"external_dependencies"`
	}

	var dependencies []DependencyInfo
	for _, item := range stored.Roadmap.Items {
		if len(item.ExternalDependencies) > 0 {
			dependencies = append(dependencies, DependencyInfo{
				ItemID:               item.ID,
				ItemName:             item.Name,
				ExternalDependencies: item.ExternalDependencies,
			})
		}
	}

	response := map[string]interface{}{
		"roadmap_id":   stored.ID,
		"roadmap_name": stored.Roadmap.Name,
		"dependencies": dependencies,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRoadmapDependents handles GET /api/roadmaps/{id}/dependents
// Returns all roadmap items that depend on this roadmap
func (h *RoadmapHandler) GetRoadmapDependents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/roadmaps/")
	id = strings.TrimSuffix(id, "/dependents")
	if id == "" || strings.Contains(id, "/") {
		http.Error(w, "Invalid roadmap ID", http.StatusBadRequest)
		return
	}

	// Get all roadmaps
	allRoadmaps, err := h.storage.List()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list roadmaps: %v", err), http.StatusInternalServerError)
		return
	}

	// Find dependents
	dependents := storage.GetExternalDependents(id, allRoadmaps)

	stored, err := h.storage.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Roadmap not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get roadmap: %v", err), http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"roadmap_id":   stored.ID,
		"roadmap_name": stored.Roadmap.Name,
		"dependents":   dependents,
		"count":        len(dependents),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateDependencies handles GET /api/dependencies/validate
// Validates all external dependencies across all roadmaps
func (h *RoadmapHandler) ValidateDependencies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all roadmaps
	allRoadmaps, err := h.storage.List()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list roadmaps: %v", err), http.StatusInternalServerError)
		return
	}

	// Validate external dependencies
	validations := storage.ValidateExternalDependencies(allRoadmaps)

	// Count valid and invalid
	validCount := 0
	invalidCount := 0
	for _, v := range validations {
		if v.Valid {
			validCount++
		} else {
			invalidCount++
		}
	}

	response := map[string]interface{}{
		"total":    len(validations),
		"valid":    validCount,
		"invalid":  invalidCount,
		"results":  validations,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleRoadmaps routes roadmap requests
func (h *RoadmapHandler) HandleRoadmaps(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-File-Name")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := r.URL.Path

	if path == "/api/roadmaps" {
		switch r.Method {
		case http.MethodPost:
			h.CreateRoadmap(w, r)
		case http.MethodGet:
			h.ListRoadmaps(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if path == "/api/roadmaps/batch" {
		// Handle batch upload of multiple roadmaps
		if r.Method == http.MethodPost {
			h.CreateMultipleRoadmaps(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if strings.HasPrefix(path, "/api/roadmaps/") {
		// Check for sub-endpoints
		if strings.HasSuffix(path, "/dependencies") {
			h.GetRoadmapDependencies(w, r)
		} else if strings.HasSuffix(path, "/dependents") {
			h.GetRoadmapDependents(w, r)
		} else {
			// Regular roadmap GET/DELETE
			switch r.Method {
			case http.MethodGet:
				h.GetRoadmap(w, r)
			case http.MethodDelete:
				h.DeleteRoadmap(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// HandleDependencies routes dependency validation requests
func (h *RoadmapHandler) HandleDependencies(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := r.URL.Path

	if path == "/api/dependencies/validate" {
		h.ValidateDependencies(w, r)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
