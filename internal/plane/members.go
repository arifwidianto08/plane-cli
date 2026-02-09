package plane

import (
	"fmt"
)

// GetWorkspaceMembers retrieves all members in the workspace
func (c *Client) GetWorkspaceMembers() ([]Member, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/members/", c.workspace)

	var response struct {
		Count   int      `json:"count"`
		Results []Member `json:"results"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get workspace members: %w", err)
	}

	return response.Results, nil
}

// GetProjectMembers retrieves all members assigned to a project
func (c *Client) GetProjectMembers(projectID string) ([]Member, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/members/", c.workspace, projectID)

	var response struct {
		Count   int      `json:"count"`
		Results []Member `json:"results"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get project members: %w", err)
	}

	return response.Results, nil
}

// Helper to get display name for a member
func (m *Member) GetDisplayName() string {
	if m.DisplayName != "" {
		return m.DisplayName
	}
	if m.FirstName != "" && m.LastName != "" {
		return fmt.Sprintf("%s %s", m.FirstName, m.LastName)
	}
	if m.FirstName != "" {
		return m.FirstName
	}
	return m.Email
}
