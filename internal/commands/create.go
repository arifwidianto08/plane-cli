package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"plane-cli/internal/config"
	"plane-cli/internal/plane"
	"plane-cli/internal/templates"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new work item",
	Long: `Create a new work item in Plane.so with optional template-based description.

Examples:
  # Create a simple work item
  plane-cli create --project my-project --title "Fix login bug"

  # Create with template
  plane-cli create --project my-project --title "User auth" --template feature

  # Create with template variables
  plane-cli create --project my-project --title "Dashboard" \
    --template feature \
    --vars feature_name="Analytics Dashboard" \
    --vars notes="High priority feature"`,
	RunE: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Required flags
	createCmd.Flags().StringP("project", "p", "", "Project identifier (required)")
	createCmd.Flags().StringP("title", "t", "", "Work item title (required)")
	createCmd.MarkFlagRequired("project")
	createCmd.MarkFlagRequired("title")

	// Optional flags
	createCmd.Flags().StringP("description", "d", "", "Work item description")
	createCmd.Flags().String("template", "", "Template name for description")
	createCmd.Flags().StringToString("vars", nil, "Template variables (key=value pairs)")
	createCmd.Flags().String("state", "", "Initial state")
	createCmd.Flags().String("priority", "medium", "Priority (urgent, high, medium, low)")
	createCmd.Flags().StringSlice("assignees", nil, "Assignee user IDs")
	createCmd.Flags().StringSlice("labels", nil, "Label IDs")
	createCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	createCmd.Flags().String("target-date", "", "Target date (YYYY-MM-DD)")
	createCmd.Flags().Float64("estimate", 0, "Estimate points")
	createCmd.Flags().String("module", "", "Module ID")
	createCmd.Flags().String("cycle", "", "Cycle ID")
	createCmd.Flags().String("parent", "", "Parent work item ID")
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse flags
	project, _ := cmd.Flags().GetString("project")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	templateName, _ := cmd.Flags().GetString("template")
	vars, _ := cmd.Flags().GetStringToString("vars")
	state, _ := cmd.Flags().GetString("state")
	priorityStr, _ := cmd.Flags().GetString("priority")
	assignees, _ := cmd.Flags().GetStringSlice("assignees")
	labels, _ := cmd.Flags().GetStringSlice("labels")
	startDate, _ := cmd.Flags().GetString("start-date")
	targetDate, _ := cmd.Flags().GetString("target-date")
	estimate, _ := cmd.Flags().GetFloat64("estimate")
	module, _ := cmd.Flags().GetString("module")
	cycle, _ := cmd.Flags().GetString("cycle")
	parent, _ := cmd.Flags().GetString("parent")
	workspace, _ := cmd.Flags().GetString("workspace")

	// Get workspace - priority: flag > env > extract from URL
	if workspace == "" {
		if cfg.PlaneWorkspace != "" {
			workspace = cfg.PlaneWorkspace
		} else {
			workspace = extractWorkspaceFromURL(cfg.PlaneBaseURL)
		}
	}

	// Initialize template manager if template is specified
	var tmplManager *templates.Manager
	if templateName != "" {
		tmplManager, err = templates.NewManager(cfg.TemplatesDir)
		if err != nil {
			return fmt.Errorf("failed to initialize template manager: %w", err)
		}
	}

	// Build description from template if specified
	if templateName != "" {
		rendered, err := tmplManager.Render(templateName, vars)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
		description = rendered
	}

	// Create Plane client
	client, err := plane.NewClient(cfg.PlaneBaseURL, cfg.PlaneAPIToken)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	client.SetWorkspace(workspace)

	// Build work item create payload
	create := &plane.WorkItemCreate{
		Name:          title,
		Description:   description,
		State:         state,
		Priority:      plane.ParsePriority(priorityStr),
		Assignees:     assignees,
		Labels:        labels,
		StartDate:     startDate,
		TargetDate:    targetDate,
		EstimatePoint: estimate,
		Module:        module,
		Cycle:         cycle,
		Parent:        parent,
	}

	// Create work item
	fmt.Printf("Creating work item in project '%s'...\n", project)
	workItem, err := client.CreateWorkItem(project, create)
	if err != nil {
		return fmt.Errorf("failed to create work item: %w", err)
	}

	fmt.Printf("✓ Created work item: %s-%d\n", project, workItem.SequenceID)
	fmt.Printf("  Title: %s\n", workItem.Name)
	if workItem.Description != "" {
		fmt.Printf("  Description: [set using template '%s']\n", templateName)
	}
	fmt.Printf("  Priority: %s\n", workItem.Priority)

	return nil
}

// extractWorkspaceFromURL extracts workspace slug from Plane URL
func extractWorkspaceFromURL(baseURL string) string {
	// This is a simplified extraction
	// You might need to adjust based on your Plane setup
	parts := strings.Split(baseURL, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return ""
}
