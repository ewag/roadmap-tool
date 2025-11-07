package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	} else if strings.HasPrefix(path, "/api/roadmaps/") {
		switch r.Method {
		case http.MethodGet:
			h.GetRoadmap(w, r)
		case http.MethodDelete:
			h.DeleteRoadmap(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
