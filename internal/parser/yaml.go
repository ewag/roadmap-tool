package parser

import (
	"fmt"
	"roadmap-visualizer/internal/models"

	"gopkg.in/yaml.v3"
)

// ParseRoadmap parses a YAML byte slice into a Roadmap struct
func ParseRoadmap(data []byte) (*models.Roadmap, error) {
	var roadmapFile models.RoadmapFile

	if err := yaml.Unmarshal(data, &roadmapFile); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the parsed roadmap
	if err := roadmapFile.Roadmap.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &roadmapFile.Roadmap, nil
}

// SerializeRoadmap converts a Roadmap to YAML bytes
func SerializeRoadmap(roadmap *models.Roadmap) ([]byte, error) {
	roadmapFile := models.RoadmapFile{
		Roadmap: *roadmap,
	}

	data, err := yaml.Marshal(&roadmapFile)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize YAML: %w", err)
	}

	return data, nil
}
