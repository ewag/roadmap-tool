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

// ExternalDependency represents a dependency on an item in another roadmap
type ExternalDependency struct {
	RoadmapName string `yaml:"roadmap" json:"roadmap"`
	RoadmapID   string `yaml:"roadmap_id,omitempty" json:"roadmap_id,omitempty"`
	ItemID      string `yaml:"item" json:"item"`
	Reason      string `yaml:"reason,omitempty" json:"reason,omitempty"`
	Criticality string `yaml:"criticality,omitempty" json:"criticality,omitempty"`
}

// RoadmapItem represents a single item on a roadmap
type RoadmapItem struct {
	ID                   string               `yaml:"id" json:"id"`
	Name                 string               `yaml:"name" json:"name"`
	Start                string               `yaml:"start" json:"start"`
	End                  string               `yaml:"end" json:"end"`
	Status               RoadmapStatus        `yaml:"status" json:"status"`
	Description          string               `yaml:"description,omitempty" json:"description,omitempty"`
	Notes                string               `yaml:"notes,omitempty" json:"notes,omitempty"`
	Dependencies         []string             `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	ExternalDependencies []ExternalDependency `yaml:"external_dependencies,omitempty" json:"external_dependencies,omitempty"`
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

	// Validate external dependencies structure
	for i, extDep := range r.ExternalDependencies {
		if extDep.RoadmapName == "" && extDep.RoadmapID == "" {
			return fmt.Errorf("external dependency %d: either roadmap name or roadmap_id is required", i)
		}
		if extDep.ItemID == "" {
			return fmt.Errorf("external dependency %d: item id is required", i)
		}
		// Validate criticality if provided
		if extDep.Criticality != "" {
			switch extDep.Criticality {
			case "low", "medium", "high", "critical":
				// valid
			default:
				return fmt.Errorf("external dependency %d: invalid criticality '%s' (must be low, medium, high, or critical)", i, extDep.Criticality)
			}
		}
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

// ExternalDependencyValidation represents validation result for an external dependency
type ExternalDependencyValidation struct {
	Valid          bool   `json:"valid"`
	RoadmapItemID  string `json:"roadmap_item_id"`
	DependencyDesc string `json:"dependency_desc"`
	Error          string `json:"error,omitempty"`
}

// ValidateExternalDependencies validates all external dependencies across roadmaps
func ValidateExternalDependencies(roadmaps []StoredRoadmap) []ExternalDependencyValidation {
	var results []ExternalDependencyValidation

	// Build lookup maps
	roadmapsByName := make(map[string]*StoredRoadmap)
	roadmapsByID := make(map[string]*StoredRoadmap)
	itemsByRoadmap := make(map[string]map[string]bool)

	for i := range roadmaps {
		rm := &roadmaps[i]
		roadmapsByName[rm.Roadmap.Name] = rm
		roadmapsByID[rm.ID] = rm
		itemsByRoadmap[rm.ID] = make(map[string]bool)
		for _, item := range rm.Roadmap.Items {
			itemsByRoadmap[rm.ID][item.ID] = true
		}
	}

	// Validate each external dependency
	for _, rm := range roadmaps {
		for _, item := range rm.Roadmap.Items {
			for _, extDep := range item.ExternalDependencies {
				validation := ExternalDependencyValidation{
					RoadmapItemID:  fmt.Sprintf("%s:%s", rm.Roadmap.Name, item.ID),
					DependencyDesc: fmt.Sprintf("%s:%s", extDep.RoadmapName, extDep.ItemID),
					Valid:          false,
				}

				// Find the target roadmap
				var targetRoadmap *StoredRoadmap
				if extDep.RoadmapID != "" {
					targetRoadmap = roadmapsByID[extDep.RoadmapID]
					if targetRoadmap == nil {
						validation.Error = fmt.Sprintf("roadmap with ID '%s' not found", extDep.RoadmapID)
						results = append(results, validation)
						continue
					}
				} else {
					targetRoadmap = roadmapsByName[extDep.RoadmapName]
					if targetRoadmap == nil {
						validation.Error = fmt.Sprintf("roadmap named '%s' not found", extDep.RoadmapName)
						results = append(results, validation)
						continue
					}
				}

				// Check if the target item exists
				if !itemsByRoadmap[targetRoadmap.ID][extDep.ItemID] {
					validation.Error = fmt.Sprintf("item '%s' not found in roadmap '%s'", extDep.ItemID, targetRoadmap.Roadmap.Name)
					results = append(results, validation)
					continue
				}

				validation.Valid = true
				results = append(results, validation)
			}
		}
	}

	return results
}

// GetExternalDependents returns all items that depend on items in the given roadmap
func GetExternalDependents(roadmapID string, allRoadmaps []StoredRoadmap) []struct {
	RoadmapID   string
	RoadmapName string
	ItemID      string
	ItemName    string
	DependsOn   string // The item ID in the target roadmap
} {
	var dependents []struct {
		RoadmapID   string
		RoadmapName string
		ItemID      string
		ItemName    string
		DependsOn   string
	}

	// Find the target roadmap
	var targetRoadmap *StoredRoadmap
	for i := range allRoadmaps {
		if allRoadmaps[i].ID == roadmapID {
			targetRoadmap = &allRoadmaps[i]
			break
		}
	}
	if targetRoadmap == nil {
		return dependents
	}

	// Find items that depend on this roadmap
	for _, rm := range allRoadmaps {
		if rm.ID == roadmapID {
			continue // Skip the roadmap itself
		}
		for _, item := range rm.Roadmap.Items {
			for _, extDep := range item.ExternalDependencies {
				// Check if this dependency points to our target roadmap
				if extDep.RoadmapID == roadmapID ||
					extDep.RoadmapName == targetRoadmap.Roadmap.Name {
					dependents = append(dependents, struct {
						RoadmapID   string
						RoadmapName string
						ItemID      string
						ItemName    string
						DependsOn   string
					}{
						RoadmapID:   rm.ID,
						RoadmapName: rm.Roadmap.Name,
						ItemID:      item.ID,
						ItemName:    item.Name,
						DependsOn:   extDep.ItemID,
					})
				}
			}
		}
	}

	return dependents
}
