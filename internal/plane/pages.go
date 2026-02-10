package plane

import (
	"fmt"
	"strings"
)

// GetPages retrieves all pages for a project
func (c *Client) GetPages(projectID string) ([]Page, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/pages/", c.workspace, projectID)

	var response PageListResponse
	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}

	return response.Results, nil
}

// GetPage retrieves a single page by ID
func (c *Client) GetPage(projectID, pageID string) (*Page, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if pageID == "" {
		return nil, fmt.Errorf("page ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/pages/%s/", c.workspace, projectID, pageID)

	var page Page
	if err := c.get(endpoint, &page); err != nil {
		return nil, fmt.Errorf("failed to get page: %w", err)
	}

	return &page, nil
}

// CreatePage creates a new page
func (c *Client) CreatePage(projectID string, create *PageCreate) (*Page, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if create == nil {
		return nil, fmt.Errorf("page data is required")
	}
	if create.Name == "" {
		return nil, fmt.Errorf("page name is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/pages/", c.workspace, projectID)

	var page Page
	if err := c.post(endpoint, create, &page); err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	return &page, nil
}

// UpdatePage updates an existing page
func (c *Client) UpdatePage(projectID, pageID string, update *PageUpdate) (*Page, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if pageID == "" {
		return nil, fmt.Errorf("page ID is required")
	}
	if update == nil {
		return nil, fmt.Errorf("update data is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/pages/%s/", c.workspace, projectID, pageID)

	var page Page
	if err := c.patch(endpoint, update, &page); err != nil {
		return nil, fmt.Errorf("failed to update page: %w", err)
	}

	return &page, nil
}

// DeletePage deletes a page
func (c *Client) DeletePage(projectID, pageID string) error {
	if c.workspace == "" {
		return fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if pageID == "" {
		return fmt.Errorf("page ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/pages/%s/", c.workspace, projectID, pageID)

	if err := c.delete(endpoint); err != nil {
		return fmt.Errorf("failed to delete page: %w", err)
	}

	return nil
}

// SearchPages searches pages by name (client-side filtering)
func (c *Client) SearchPages(projectID, query string) ([]Page, error) {
	pages, err := c.GetPages(projectID)
	if err != nil {
		return nil, err
	}

	if query == "" {
		return pages, nil
	}

	// Simple case-insensitive search
	var results []Page
	query = strings.ToLower(query)
	for _, p := range pages {
		if strings.Contains(strings.ToLower(p.Name), query) {
			results = append(results, p)
		}
	}

	return results, nil
}

// GetPageChildren retrieves child pages of a page
func (c *Client) GetPageChildren(projectID, pageID string) ([]Page, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if pageID == "" {
		return nil, fmt.Errorf("page ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/pages/%s/children/", c.workspace, projectID, pageID)

	var response PageListResponse
	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get page children: %w", err)
	}

	return response.Results, nil
}
