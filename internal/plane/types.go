package plane

import (
	"time"
)

// WorkItem represents a Plane.so work item (issue)
type WorkItem struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	DescriptionHTML string    `json:"description_html,omitempty"`
	State           string    `json:"state"`
	StateID         string    `json:"state_id"`
	Priority        string    `json:"priority"`
	Assignees       []string  `json:"assignees,omitempty"`
	AssigneeIDs     []string  `json:"assignee_ids,omitempty"`
	Labels          []string  `json:"labels,omitempty"`
	LabelIDs        []string  `json:"label_ids,omitempty"`
	ProjectID       string    `json:"project_id"`
	Project         string    `json:"project"`
	WorkspaceID     string    `json:"workspace_id"`
	SequenceID      int       `json:"sequence_id"`
	StartDate       *string   `json:"start_date,omitempty"`
	TargetDate      *string   `json:"target_date,omitempty"`
	EstimatePoint   *string   `json:"estimate_point,omitempty"`
	Module          string    `json:"module,omitempty"`
	ModuleID        string    `json:"module_id,omitempty"`
	Cycle           string    `json:"cycle,omitempty"`
	CycleID         string    `json:"cycle_id,omitempty"`
	ParentID        string    `json:"parent,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// WorkItemCreate represents the payload for creating a work item
type WorkItemCreate struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	State         string   `json:"state,omitempty"`
	Priority      int      `json:"priority,omitempty"`
	Assignees     []string `json:"assignees,omitempty"`
	Labels        []string `json:"labels,omitempty"`
	StartDate     string   `json:"start_date,omitempty"`
	TargetDate    string   `json:"target_date,omitempty"`
	EstimatePoint float64  `json:"estimate_point,omitempty"`
	Module        string   `json:"module,omitempty"`
	Cycle         string   `json:"cycle,omitempty"`
	Parent        string   `json:"parent,omitempty"`
}

// WorkItemUpdate represents the payload for updating a work item
type WorkItemUpdate struct {
	Name            string   `json:"name,omitempty"`
	DescriptionHTML string   `json:"description_html,omitempty"`
	State           string   `json:"state,omitempty"`
	Priority        string   `json:"priority,omitempty"`
	Assignees       []string `json:"assignees,omitempty"`
	Labels          []string `json:"labels,omitempty"`
	StartDate       string   `json:"start_date,omitempty"`
	TargetDate      string   `json:"target_date,omitempty"`
	EstimatePoint   float64  `json:"estimate_point,omitempty"`
	Module          string   `json:"module,omitempty"`
	Cycle           string   `json:"cycle,omitempty"`
	Parent          string   `json:"parent,omitempty"`
}

// Project represents a Plane.so project
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description,omitempty"`
	WorkspaceID string `json:"workspace_id"`
}

// State represents a workflow state in a project
type State struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Group       string `json:"group"`
	Color       string `json:"color"`
	ProjectID   string `json:"project_id"`
	WorkspaceID string `json:"workspace_id"`
}

// Label represents a label/tag in a project
type Label struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Color       string    `json:"color,omitempty"`
	ProjectID   string    `json:"project_id"`
	WorkspaceID string    `json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LabelCreate represents payload for creating a label
type LabelCreate struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// LabelUpdate represents payload for updating a label
type LabelUpdate struct {
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

// Cycle represents a sprint/cycle in a project
type Cycle struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ProjectID   string `json:"project_id"`
	WorkspaceID string `json:"workspace_id"`
}

// Module represents a module in a project
type Module struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color,omitempty"`
	Status      string    `json:"status,omitempty"`
	ProjectID   string    `json:"project_id"`
	WorkspaceID string    `json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ModuleCreate represents payload for creating a module
type ModuleCreate struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Status      string `json:"status,omitempty"`
}

// ModuleUpdate represents payload for updating a module
type ModuleUpdate struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Status      string `json:"status,omitempty"`
}

// Page represents a page/document in a project
type Page struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	DescriptionHTML string    `json:"description_html,omitempty"`
	ProjectID       string    `json:"project_id"`
	WorkspaceID     string    `json:"workspace_id"`
	ParentID        string    `json:"parent,omitempty"`
	Access          string    `json:"access,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PageCreate represents payload for creating a page
type PageCreate struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	DescriptionHTML string `json:"description_html,omitempty"`
	ParentID        string `json:"parent,omitempty"`
	Access          string `json:"access,omitempty"`
}

// PageUpdate represents payload for updating a page
type PageUpdate struct {
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	DescriptionHTML string `json:"description_html,omitempty"`
	ParentID        string `json:"parent,omitempty"`
	Access          string `json:"access,omitempty"`
}

// Member represents a workspace member/user
type Member struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// ListResponse represents a paginated API response
type ListResponse struct {
	TotalCount      int        `json:"total_count"`
	Count           int        `json:"count"`
	NextCursor      *string    `json:"next_cursor"`
	PrevCursor      *string    `json:"prev_cursor"`
	NextPageResults bool       `json:"next_page_results"`
	PrevPageResults bool       `json:"prev_page_results"`
	TotalPages      int        `json:"total_pages"`
	TotalResults    int        `json:"total_results"`
	Results         []WorkItem `json:"results"`
}

// ModuleListResponse represents paginated modules response
type ModuleListResponse struct {
	Count    int      `json:"count"`
	Results  []Module `json:"results"`
	Next     *string  `json:"next"`
	Previous *string  `json:"previous"`
}

// LabelListResponse represents paginated labels response
type LabelListResponse struct {
	Count    int     `json:"count"`
	Results  []Label `json:"results"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
}

// PageListResponse represents paginated pages response
type PageListResponse struct {
	Count    int     `json:"count"`
	Results  []Page  `json:"results"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
}

// Priority levels
const (
	PriorityUrgent = 0
	PriorityHigh   = 1
	PriorityMedium = 2
	PriorityLow    = 3
)

// PriorityNames maps priority levels to names
var PriorityNames = map[int]string{
	PriorityUrgent: "Urgent",
	PriorityHigh:   "High",
	PriorityMedium: "Medium",
	PriorityLow:    "Low",
}

// GetPriorityName returns the name for a priority level
func GetPriorityName(priority int) string {
	if name, ok := PriorityNames[priority]; ok {
		return name
	}
	return "Unknown"
}

// ParsePriority parses a priority string to int
func ParsePriority(s string) int {
	switch s {
	case "urgent", "Urgent", "0":
		return PriorityUrgent
	case "high", "High", "1":
		return PriorityHigh
	case "low", "Low", "3":
		return PriorityLow
	default:
		return PriorityMedium
	}
}
