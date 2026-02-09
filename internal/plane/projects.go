package plane

import (
	"fmt"
	"strings"
)

// GetProjects retrieves all projects in the workspace
func (c *Client) GetProjects() ([]Project, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/", c.workspace)

	var response struct {
		Count    int       `json:"count"`
		Next     *string   `json:"next"`
		Previous *string   `json:"previous"`
		Results  []Project `json:"results"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return response.Results, nil
}

// GetProject retrieves a single project by identifier
func (c *Client) GetProject(projectID string) (*Project, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/", c.workspace, projectID)

	var project Project
	if err := c.get(endpoint, &project); err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// GetProjectStates retrieves all workflow states for a project
func (c *Client) GetProjectStates(projectID string) ([]State, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/states/", c.workspace, projectID)

	var response struct {
		Results []State `json:"results"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get project states: %w", err)
	}

	return response.Results, nil
}

// GetProjectLabels retrieves all labels for a project
func (c *Client) GetProjectLabels(projectID string) ([]Label, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/labels/", c.workspace, projectID)

	var response struct {
		Results []Label `json:"results"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get project labels: %w", err)
	}

	return response.Results, nil
}

// GetProjectCycles retrieves all cycles/sprints for a project
func (c *Client) GetProjectCycles(projectID string) ([]Cycle, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/cycles/", c.workspace, projectID)

	var response struct {
		Results []Cycle `json:"results"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get project cycles: %w", err)
	}

	return response.Results, nil
}

// GetProjectModules retrieves all modules for a project
func (c *Client) GetProjectModules(projectID string) ([]Module, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/modules/", c.workspace, projectID)

	var response struct {
		Results []Module `json:"results"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get project modules: %w", err)
	}

	return response.Results, nil
}

// SearchProjects searches projects by name (client-side filtering)
func (c *Client) SearchProjects(query string) ([]Project, error) {
	projects, err := c.GetProjects()
	if err != nil {
		return nil, err
	}

	if query == "" {
		return projects, nil
	}

	// Simple case-insensitive search
	var results []Project
	query = strings.ToLower(query)
	for _, p := range projects {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Identifier), query) {
			results = append(results, p)
		}
	}

	return results, nil
}

// Helper to check if project exists
func (c *Client) ProjectExists(projectID string) (bool, error) {
	_, err := c.GetProject(projectID)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
