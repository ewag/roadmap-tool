package models

import (
	"fmt"
	"time"
)

// RoadmapStatus represents the status of a roadmap item
type RoadmapStatus string

const (
	StatusPlanned    RoadmapStatus = "planned"
	StatusInProgress RoadmapStatus = "in-progress"
	StatusCompleted  RoadmapStatus = "completed"
	StatusBlocked    RoadmapStatus = "blocked"
)

// ValidateStatus checks if a status string is valid
func ValidateStatus(status string) error {
	switch RoadmapStatus(status) {
	case StatusPlanned, StatusInProgress, StatusCompleted, StatusBlocked:
		return nil
	default:
		return fmt.Errorf("invalid status: %s (must be planned, in-progress, completed, or blocked)", status)
	}
}

// RoadmapItem represents a single item on a roadmap
type RoadmapItem struct {
	ID           string        `yaml:"id" json:"id"`
	Name         string        `yaml:"name" json:"name"`
	Start        string        `yaml:"start" json:"start"`
	End          string        `yaml:"end" json:"end"`
	Status       RoadmapStatus `yaml:"status" json:"status"`
	Description  string        `yaml:"description,omitempty" json:"description,omitempty"`
	Notes        string        `yaml:"notes,omitempty" json:"notes,omitempty"`
	Dependencies []string      `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
}

// Validate checks if a roadmap item has all required fields
func (r *RoadmapItem) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("item id is required")
	}
	if r.Name == "" {
		return fmt.Errorf("item name is required")
	}
	if r.Start == "" {
		return fmt.Errorf("item start is required")
	}
	if r.End == "" {
		return fmt.Errorf("item end is required")
	}
	if err := ValidateStatus(string(r.Status)); err != nil {
		return err
	}
	return nil
}

// Roadmap represents a complete roadmap
type Roadmap struct {
	Name        string         `yaml:"name" json:"name"`
	ServiceLine string         `yaml:"service_line" json:"service_line"`
	Owner       string         `yaml:"owner,omitempty" json:"owner,omitempty"`
	Notes       string         `yaml:"notes,omitempty" json:"notes,omitempty"`
	Items       []RoadmapItem  `yaml:"items" json:"items"`
}

// Validate checks if a roadmap has all required fields and valid items
func (r *Roadmap) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("roadmap name is required")
	}
	if r.ServiceLine == "" {
		return fmt.Errorf("service_line is required")
	}
	if len(r.Items) == 0 {
		return fmt.Errorf("roadmap must have at least one item")
	}

	// Validate each item
	itemIDs := make(map[string]bool)
	for i, item := range r.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("item %d: %w", i, err)
		}
		// Check for duplicate IDs
		if itemIDs[item.ID] {
			return fmt.Errorf("duplicate item id: %s", item.ID)
		}
		itemIDs[item.ID] = true
	}

	// Validate dependencies reference existing items
	for _, item := range r.Items {
		for _, depID := range item.Dependencies {
			if !itemIDs[depID] {
				return fmt.Errorf("item %s: dependency %s does not exist", item.ID, depID)
			}
		}
	}

	return nil
}

// RoadmapFile represents the top-level structure of a roadmap YAML file
type RoadmapFile struct {
	Roadmap Roadmap `yaml:"roadmap" json:"roadmap"`
}

// StoredRoadmap represents a roadmap as stored in the system
type StoredRoadmap struct {
	ID          string    `json:"id"`
	Roadmap     Roadmap   `json:"roadmap"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	FileName    string    `json:"file_name"`
}
