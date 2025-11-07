package parser

import (
	"bytes"
	"fmt"
	"io"
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

// ParseMultipleRoadmaps parses a YAML file containing multiple roadmap documents
// separated by --- into a slice of Roadmap structs
func ParseMultipleRoadmaps(data []byte) ([]*models.Roadmap, error) {
	var roadmaps []*models.Roadmap

	// Create a YAML decoder to handle multiple documents
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var roadmapFile models.RoadmapFile

		// Decode the next document
		err := decoder.Decode(&roadmapFile)
		if err == io.EOF {
			// No more documents
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse YAML document %d: %w", len(roadmaps)+1, err)
		}

		// Validate the parsed roadmap
		if err := roadmapFile.Roadmap.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed for roadmap %d (%s): %w", len(roadmaps)+1, roadmapFile.Roadmap.Name, err)
		}

		roadmaps = append(roadmaps, &roadmapFile.Roadmap)
	}

	if len(roadmaps) == 0 {
		return nil, fmt.Errorf("no roadmaps found in file")
	}

	return roadmaps, nil
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
