package plane

import (
	"fmt"
	"net/url"
	"strconv"
)

// GetWorkItems retrieves a list of work items for a project
func (c *Client) GetWorkItems(projectID string, options map[string]string) (*ListResponse, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	// Build query parameters
	params := url.Values{}
	for key, value := range options {
		params.Add(key, value)
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/work-items/", c.workspace, projectID)

	var response ListResponse
	if err := c.getWithQuery(endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get work items: %w", err)
	}

	return &response, nil
}

// GetWorkItem retrieves a single work item by ID
func (c *Client) GetWorkItem(projectID, workItemID string) (*WorkItem, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if workItemID == "" {
		return nil, fmt.Errorf("work item ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/work-items/%s/", c.workspace, projectID, workItemID)

	var workItem WorkItem
	if err := c.get(endpoint, &workItem); err != nil {
		return nil, fmt.Errorf("failed to get work item: %w", err)
	}

	return &workItem, nil
}

// CreateWorkItem creates a new work item
func (c *Client) CreateWorkItem(projectID string, create *WorkItemCreate) (*WorkItem, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if create == nil {
		return nil, fmt.Errorf("work item data is required")
	}
	if create.Name == "" {
		return nil, fmt.Errorf("work item name is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/work-items/", c.workspace, projectID)

	var workItem WorkItem
	if err := c.post(endpoint, create, &workItem); err != nil {
		return nil, fmt.Errorf("failed to create work item: %w", err)
	}

	return &workItem, nil
}

// UpdateWorkItem updates an existing work item
func (c *Client) UpdateWorkItem(projectID, workItemID string, update *WorkItemUpdate) (*WorkItem, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if workItemID == "" {
		return nil, fmt.Errorf("work item ID is required")
	}
	if update == nil {
		return nil, fmt.Errorf("update data is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/work-items/%s/", c.workspace, projectID, workItemID)

	var workItem WorkItem
	if err := c.patch(endpoint, update, &workItem); err != nil {
		return nil, fmt.Errorf("failed to update work item: %w", err)
	}

	return &workItem, nil
}

// DeleteWorkItem deletes a work item
func (c *Client) DeleteWorkItem(projectID, workItemID string) error {
	if c.workspace == "" {
		return fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if workItemID == "" {
		return fmt.Errorf("work item ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/work-items/%s/", c.workspace, projectID, workItemID)

	if err := c.delete(endpoint); err != nil {
		return fmt.Errorf("failed to delete work item: %w", err)
	}

	return nil
}

// SearchWorkItems searches work items by title (client-side filtering)
// Note: This fetches all work items and filters locally. For large projects,
// consider implementing server-side search if Plane API supports it
func (c *Client) SearchWorkItems(projectID, query string) ([]WorkItem, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	// Get all work items (handle pagination if needed)
	var allItems []WorkItem
	nextURL := ""

	for {
		options := map[string]string{}
		if nextURL != "" {
			// Parse next URL to extract cursor/page
			// This is a simplified version - actual implementation may vary
		}

		response, err := c.GetWorkItems(projectID, options)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, response.Results...)

		// Check if there are more results using cursor-based pagination
		if !response.NextPageResults || response.NextCursor == nil {
			break
		}
		// Use cursor for next page (stored in options for next iteration)
		options["cursor"] = *response.NextCursor
	}

	return allItems, nil
}

// Helper function to convert int to string
func intToString(i int) string {
	return strconv.Itoa(i)
}

// BuildCreatePayload builds a WorkItemCreate from a map of values
func BuildCreatePayload(values map[string]interface{}) (*WorkItemCreate, error) {
	payload := &WorkItemCreate{}

	if name, ok := values["name"].(string); ok {
		payload.Name = name
	}
	if desc, ok := values["description"].(string); ok {
		payload.Description = desc
	}
	if state, ok := values["state"].(string); ok {
		payload.State = state
	}
	if priority, ok := values["priority"].(string); ok {
		payload.Priority = priority
	}
	if assignees, ok := values["assignees"].([]string); ok {
		payload.Assignees = assignees
	}
	if labels, ok := values["labels"].([]string); ok {
		payload.Labels = labels
	}
	if startDate, ok := values["start_date"].(string); ok {
		payload.StartDate = startDate
	}
	if targetDate, ok := values["target_date"].(string); ok {
		payload.TargetDate = targetDate
	}
	if estimate, ok := values["estimate_point"].(string); ok {
		payload.EstimatePoint = estimate
	}
	if module, ok := values["module"].(string); ok {
		payload.Module = module
	}
	if cycle, ok := values["cycle"].(string); ok {
		payload.Cycle = cycle
	}
	if parent, ok := values["parent"].(string); ok {
		payload.Parent = parent
	}

	return payload, nil
}
