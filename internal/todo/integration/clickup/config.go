package clickup

import (
	"fmt"
)

// Sync metadata constants
const (
	SyncName       = "clickup"
	SyncKeyTaskID  = "task_id"
	SyncKeySyncedAt = "synced_at"
)

// Config holds ClickUp-specific settings parsed from cfg.SyncConfig("clickup").
type Config struct {
	ListID          string
	Assignee        *int
	StatusMapping   map[string]string
	PriorityMapping map[string]int
	TypeMapping     map[string]int
	CustomFields    *CustomFieldsMap
	SyncFilter      *SyncFilter
}

// CustomFieldsMap maps issue fields to ClickUp custom field UUIDs.
type CustomFieldsMap struct {
	IssueID   string
	CreatedAt string
	UpdatedAt string
}

// SyncFilter defines which issues to sync.
type SyncFilter struct {
	ExcludeStatus []string
}

// DefaultStatusMapping provides standard issue status → ClickUp status mapping.
var DefaultStatusMapping = map[string]string{
	"draft":       "backlog",
	"ready":       "to do",
	"in-progress": "in progress",
	"review":      "review",
	"completed":   "complete",
	"scrapped":    "closed",
}

// ClickUp priority levels.
const (
	PriorityUrgent = 1
	PriorityHigh   = 2
	PriorityNormal = 3
	PriorityLow    = 4
)

// DefaultPriorityMapping provides standard issue priority → ClickUp priority mapping.
var DefaultPriorityMapping = map[string]int{
	"critical": PriorityUrgent,
	"high":     PriorityHigh,
	"normal":   PriorityNormal,
	"low":      PriorityLow,
	"deferred": PriorityLow,
}

// ParseConfig parses ClickUp configuration from a map[string]any (from cfg.SyncConfig("clickup")).
// Returns nil if the map is nil or has no list_id.
func ParseConfig(m map[string]any) (*Config, error) {
	if m == nil {
		return nil, nil
	}

	listID, _ := m["list_id"].(string)
	if listID == "" {
		return nil, nil
	}

	cfg := &Config{
		ListID:          listID,
		StatusMapping:   DefaultStatusMapping,
		PriorityMapping: DefaultPriorityMapping,
	}

	// Parse assignee
	if v, ok := m["assignee"]; ok {
		switch a := v.(type) {
		case int:
			cfg.Assignee = &a
		case float64:
			i := int(a)
			cfg.Assignee = &i
		}
	}

	// Parse status_mapping
	if v, ok := m["status_mapping"]; ok {
		if sm, ok := v.(map[string]any); ok {
			mapping := make(map[string]string, len(sm))
			for k, val := range sm {
				if s, ok := val.(string); ok {
					mapping[k] = s
				}
			}
			if len(mapping) > 0 {
				cfg.StatusMapping = mapping
			}
		}
	}

	// Parse priority_mapping
	if v, ok := m["priority_mapping"]; ok {
		if pm, ok := v.(map[string]any); ok {
			if mapping := parseIntMapping(pm); len(mapping) > 0 {
				cfg.PriorityMapping = mapping
			}
		}
	}

	// Parse type_mapping
	if v, ok := m["type_mapping"]; ok {
		if tm, ok := v.(map[string]any); ok {
			if mapping := parseIntMapping(tm); len(mapping) > 0 {
				cfg.TypeMapping = mapping
			}
		}
	}

	// Parse custom_fields
	if v, ok := m["custom_fields"]; ok {
		if cf, ok := v.(map[string]any); ok {
			fields := &CustomFieldsMap{}
			fields.IssueID, _ = cf["issue_id"].(string)
			fields.CreatedAt, _ = cf["created_at"].(string)
			fields.UpdatedAt, _ = cf["updated_at"].(string)
			if fields.IssueID != "" || fields.CreatedAt != "" || fields.UpdatedAt != "" {
				cfg.CustomFields = fields
			}
		}
	}

	// Parse sync_filter
	if v, ok := m["sync_filter"]; ok {
		if sf, ok := v.(map[string]any); ok {
			filter := &SyncFilter{}
			if es, ok := sf["exclude_status"]; ok {
				switch s := es.(type) {
				case []any:
					for _, item := range s {
						if str, ok := item.(string); ok {
							filter.ExcludeStatus = append(filter.ExcludeStatus, str)
						}
					}
				}
			}
			if len(filter.ExcludeStatus) > 0 {
				cfg.SyncFilter = filter
			}
		}
	}

	return cfg, nil
}

// parseIntMapping converts a map[string]any (from YAML/JSON) into a map[string]int,
// accepting both int and float64 values.
func parseIntMapping(m map[string]any) map[string]int {
	result := make(map[string]int, len(m))
	for k, val := range m {
		switch n := val.(type) {
		case int:
			result[k] = n
		case float64:
			result[k] = int(n)
		}
	}
	return result
}

// GetStatusMapping returns the effective status mapping.
func (c *Config) GetStatusMapping() map[string]string {
	if c.StatusMapping != nil {
		return c.StatusMapping
	}
	return DefaultStatusMapping
}

// GetPriorityMapping returns the effective priority mapping.
func (c *Config) GetPriorityMapping() map[string]int {
	if c.PriorityMapping != nil {
		return c.PriorityMapping
	}
	return DefaultPriorityMapping
}

// Validate checks the config for issues.
func (c *Config) Validate() error {
	if c.ListID == "" {
		return fmt.Errorf("list_id is required")
	}
	return nil
}
