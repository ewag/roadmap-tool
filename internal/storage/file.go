package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"roadmap-visualizer/internal/models"
	"roadmap-visualizer/internal/parser"
	"sync"
	"time"

	"github.com/google/uuid"
)

// FileStorage implements file-based storage for roadmaps
type FileStorage struct {
	dataDir string
	mu      sync.RWMutex
}

// NewFileStorage creates a new file storage instance
func NewFileStorage(dataDir string) (*FileStorage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create subdirectories for YAML and metadata
	yamlDir := filepath.Join(dataDir, "yaml")
	metaDir := filepath.Join(dataDir, "meta")

	if err := os.MkdirAll(yamlDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create yaml directory: %w", err)
	}
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create meta directory: %w", err)
	}

	return &FileStorage{
		dataDir: dataDir,
	}, nil
}

// Create stores a new roadmap
func (fs *FileStorage) Create(roadmap *models.Roadmap, originalFileName string) (*models.StoredRoadmap, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	id := uuid.New().String()
	now := time.Now()

	stored := &models.StoredRoadmap{
		ID:        id,
		Roadmap:   *roadmap,
		CreatedAt: now,
		UpdatedAt: now,
		FileName:  originalFileName,
	}

	// Serialize roadmap to YAML
	yamlData, err := parser.SerializeRoadmap(roadmap)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize roadmap: %w", err)
	}

	// Write YAML file
	yamlPath := filepath.Join(fs.dataDir, "yaml", fmt.Sprintf("%s.yaml", id))
	if err := os.WriteFile(yamlPath, yamlData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write yaml file: %w", err)
	}

	// Write metadata file
	metaData, err := json.Marshal(stored)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize metadata: %w", err)
	}

	metaPath := filepath.Join(fs.dataDir, "meta", fmt.Sprintf("%s.json", id))
	if err := os.WriteFile(metaPath, metaData, 0644); err != nil {
		// Clean up YAML file if metadata write fails
		os.Remove(yamlPath)
		return nil, fmt.Errorf("failed to write metadata file: %w", err)
	}

	return stored, nil
}

// Get retrieves a roadmap by ID
func (fs *FileStorage) Get(id string) (*models.StoredRoadmap, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	metaPath := filepath.Join(fs.dataDir, "meta", fmt.Sprintf("%s.json", id))
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("roadmap not found")
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var stored models.StoredRoadmap
	if err := json.Unmarshal(metaData, &stored); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &stored, nil
}

// List returns all stored roadmaps
func (fs *FileStorage) List() ([]*models.StoredRoadmap, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	metaDir := filepath.Join(fs.dataDir, "meta")
	entries, err := os.ReadDir(metaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata directory: %w", err)
	}

	var roadmaps []*models.StoredRoadmap
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		metaPath := filepath.Join(metaDir, entry.Name())
		metaData, err := os.ReadFile(metaPath)
		if err != nil {
			continue // Skip files we can't read
		}

		var stored models.StoredRoadmap
		if err := json.Unmarshal(metaData, &stored); err != nil {
			continue // Skip files we can't parse
		}

		roadmaps = append(roadmaps, &stored)
	}

	return roadmaps, nil
}

// Delete removes a roadmap by ID
func (fs *FileStorage) Delete(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	yamlPath := filepath.Join(fs.dataDir, "yaml", fmt.Sprintf("%s.yaml", id))
	metaPath := filepath.Join(fs.dataDir, "meta", fmt.Sprintf("%s.json", id))

	// Check if metadata exists
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return fmt.Errorf("roadmap not found")
	}

	// Delete both files
	if err := os.Remove(yamlPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete yaml file: %w", err)
	}

	if err := os.Remove(metaPath); err != nil {
		return fmt.Errorf("failed to delete metadata file: %w", err)
	}

	return nil
}

// ValidateExternalDependencies validates all external dependencies across roadmaps
func ValidateExternalDependencies(roadmaps []*models.StoredRoadmap) []models.ExternalDependencyValidation {
	// Convert to slice of values for models function
	rmValues := make([]models.StoredRoadmap, len(roadmaps))
	for i, rm := range roadmaps {
		rmValues[i] = *rm
	}
	return models.ValidateExternalDependencies(rmValues)
}

// GetExternalDependents returns all items that depend on items in the given roadmap
func GetExternalDependents(roadmapID string, allRoadmaps []*models.StoredRoadmap) []struct {
	RoadmapID   string
	RoadmapName string
	ItemID      string
	ItemName    string
	DependsOn   string
} {
	// Convert to slice of values for models function
	rmValues := make([]models.StoredRoadmap, len(allRoadmaps))
	for i, rm := range allRoadmaps {
		rmValues[i] = *rm
	}
	return models.GetExternalDependents(roadmapID, rmValues)
}
