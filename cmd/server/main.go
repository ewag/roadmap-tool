package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"roadmap-visualizer/internal/handlers"
	"roadmap-visualizer/internal/storage"
)

func main() {
	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	// Initialize storage
	fileStorage, err := storage.NewFileStorage(dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize handlers
	roadmapHandler := handlers.NewRoadmapHandler(fileStorage)

	// Set up routes
	http.HandleFunc("/api/roadmaps", roadmapHandler.HandleRoadmaps)
	http.HandleFunc("/api/roadmaps/", roadmapHandler.HandleRoadmaps)

	// Health check endpoints
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Serve HTML templates
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "web/templates/index.html")
		} else if r.URL.Path == "/list" {
			http.ServeFile(w, r, "web/templates/list.html")
		} else if r.URL.Path == "/view" {
			http.ServeFile(w, r, "web/templates/view.html")
		} else if r.URL.Path == "/compare" {
			http.ServeFile(w, r, "web/templates/compare.html")
		} else {
			http.NotFound(w, r)
		}
	})

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Data directory: %s", dataDir)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
