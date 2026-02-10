package plane

import (
	"fmt"
	"strings"
)

// GetLabels retrieves all labels for a project
func (c *Client) GetLabels(projectID string) ([]Label, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/labels/", c.workspace, projectID)

	var response LabelListResponse
	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}

	return response.Results, nil
}

// GetLabel retrieves a single label by ID
func (c *Client) GetLabel(projectID, labelID string) (*Label, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if labelID == "" {
		return nil, fmt.Errorf("label ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/labels/%s/", c.workspace, projectID, labelID)

	var label Label
	if err := c.get(endpoint, &label); err != nil {
		return nil, fmt.Errorf("failed to get label: %w", err)
	}

	return &label, nil
}

// CreateLabel creates a new label
func (c *Client) CreateLabel(projectID string, create *LabelCreate) (*Label, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if create == nil {
		return nil, fmt.Errorf("label data is required")
	}
	if create.Name == "" {
		return nil, fmt.Errorf("label name is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/labels/", c.workspace, projectID)

	var label Label
	if err := c.post(endpoint, create, &label); err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}

	return &label, nil
}

// UpdateLabel updates an existing label
func (c *Client) UpdateLabel(projectID, labelID string, update *LabelUpdate) (*Label, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if labelID == "" {
		return nil, fmt.Errorf("label ID is required")
	}
	if update == nil {
		return nil, fmt.Errorf("update data is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/labels/%s/", c.workspace, projectID, labelID)

	var label Label
	if err := c.patch(endpoint, update, &label); err != nil {
		return nil, fmt.Errorf("failed to update label: %w", err)
	}

	return &label, nil
}

// DeleteLabel deletes a label
func (c *Client) DeleteLabel(projectID, labelID string) error {
	if c.workspace == "" {
		return fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if labelID == "" {
		return fmt.Errorf("label ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/labels/%s/", c.workspace, projectID, labelID)

	if err := c.delete(endpoint); err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	return nil
}

// SearchLabels searches labels by name (client-side filtering)
func (c *Client) SearchLabels(projectID, query string) ([]Label, error) {
	labels, err := c.GetLabels(projectID)
	if err != nil {
		return nil, err
	}

	if query == "" {
		return labels, nil
	}

	// Simple case-insensitive search
	var results []Label
	query = strings.ToLower(query)
	for _, l := range labels {
		if strings.Contains(strings.ToLower(l.Name), query) {
			results = append(results, l)
		}
	}

	return results, nil
}
