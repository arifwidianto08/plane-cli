package plane

import (
	"fmt"
	"strconv"
)

// GetEstimates retrieves all estimate configurations for a project
func (c *Client) GetEstimates(projectID string) ([]Estimate, error) {
	if c.workspace == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/estimates/", c.workspace, projectID)

	var estimates []Estimate
	if err := c.get(endpoint, &estimates); err != nil {
		return nil, fmt.Errorf("failed to get estimates: %w", err)
	}

	return estimates, nil
}

// GetEstimatePointByValue finds an estimate point UUID by its numeric value
func (c *Client) GetEstimatePointByValue(projectID string, value float64) (string, error) {
	estimates, err := c.GetEstimates(projectID)
	if err != nil {
		return "", err
	}

	// Convert value to string for comparison
	valueStr := strconv.FormatFloat(value, 'f', -1, 64)
	// Also try without decimal for integers
	valueStrInt := strconv.FormatFloat(value, 'f', 0, 64)

	for _, estimate := range estimates {
		for _, point := range estimate.Points {
			if point.Value == valueStr || point.Value == valueStrInt {
				return point.ID, nil
			}
		}
	}

	return "", fmt.Errorf("no estimate point found for value %v", value)
}

// GetStateByName finds a state ID by its name
func (c *Client) GetStateByName(projectID, name string) (string, error) {
	states, err := c.GetProjectStates(projectID)
	if err != nil {
		return "", err
	}

	nameLower := ""
	for _, s := range states {
		if s.Name == name {
			return s.ID, nil
		}
		// Case-insensitive fallback
		if nameLower == "" {
			nameLower = toLower(name)
		}
		if toLower(s.Name) == nameLower {
			return s.ID, nil
		}
	}

	return "", fmt.Errorf("state '%s' not found", name)
}

func toLower(s string) string {
	// Simple lowercase conversion
	result := []rune(s)
	for i, r := range result {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + ('a' - 'A')
		}
	}
	return string(result)
}
