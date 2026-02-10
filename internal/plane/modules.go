package plane

import (
	"fmt"
)

// GetModules retrieves all modules for a project
func (c *Client) GetModules(projectID string) ([]Module, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/modules/", c.workspace, projectID)

	var response ModuleListResponse
	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get modules: %w", err)
	}

	return response.Results, nil
}

// GetModule retrieves a single module by ID
func (c *Client) GetModule(projectID, moduleID string) (*Module, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if moduleID == "" {
		return nil, fmt.Errorf("module ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/modules/%s/", c.workspace, projectID, moduleID)

	var module Module
	if err := c.get(endpoint, &module); err != nil {
		return nil, fmt.Errorf("failed to get module: %w", err)
	}

	return &module, nil
}

// CreateModule creates a new module
func (c *Client) CreateModule(projectID string, create *ModuleCreate) (*Module, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if create == nil {
		return nil, fmt.Errorf("module data is required")
	}
	if create.Name == "" {
		return nil, fmt.Errorf("module name is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/modules/", c.workspace, projectID)

	var module Module
	if err := c.post(endpoint, create, &module); err != nil {
		return nil, fmt.Errorf("failed to create module: %w", err)
	}

	return &module, nil
}

// UpdateModule updates an existing module
func (c *Client) UpdateModule(projectID, moduleID string, update *ModuleUpdate) (*Module, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if moduleID == "" {
		return nil, fmt.Errorf("module ID is required")
	}
	if update == nil {
		return nil, fmt.Errorf("update data is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/modules/%s/", c.workspace, projectID, moduleID)

	var module Module
	if err := c.patch(endpoint, update, &module); err != nil {
		return nil, fmt.Errorf("failed to update module: %w", err)
	}

	return &module, nil
}

// DeleteModule deletes a module
func (c *Client) DeleteModule(projectID, moduleID string) error {
	if c.workspace == "" {
		return fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if moduleID == "" {
		return fmt.Errorf("module ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/modules/%s/", c.workspace, projectID, moduleID)

	if err := c.delete(endpoint); err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}

	return nil
}

// GetModuleWorkItems retrieves work items associated with a module
func (c *Client) GetModuleWorkItems(projectID, moduleID string) ([]WorkItem, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if moduleID == "" {
		return nil, fmt.Errorf("module ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/modules/%s/work-items/", c.workspace, projectID, moduleID)

	var response ListResponse
	if err := c.get(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get module work items: %w", err)
	}

	return response.Results, nil
}
