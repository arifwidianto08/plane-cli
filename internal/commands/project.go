package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"plane-cli/internal/config"
	"plane-cli/internal/plane"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long: `List and select projects from your Plane workspace.

Examples:
  # List all projects
  plane-cli project list

  # Search projects
  plane-cli project list --search "admin"

  # Select project for commands
  plane-cli project select`,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE:  runProjectList,
}

var projectSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Interactively select a project",
	Long: `Interactively select a project from your workspace.

This is useful for setting a default project for subsequent commands.`,
	RunE: runProjectSelect,
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectSelectCmd)

	// List flags
	projectListCmd.Flags().String("search", "", "Search projects by name")
}

func runProjectList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	search, _ := cmd.Flags().GetString("search")
	workspace, _ := cmd.Flags().GetString("workspace")

	if workspace == "" {
		if cfg.PlaneWorkspace != "" {
			workspace = cfg.PlaneWorkspace
		} else {
			workspace = extractWorkspaceFromURL(cfg.PlaneBaseURL)
		}
	}

	client, err := plane.NewClient(cfg.PlaneBaseURL, cfg.PlaneAPIToken)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	client.SetWorkspace(workspace)

	var projects []plane.Project
	if search != "" {
		projects, err = client.SearchProjects(search)
	} else {
		projects, err = client.GetProjects()
	}

	if err != nil {
		return fmt.Errorf("failed to fetch projects: %w", err)
	}

	if len(projects) == 0 {
		if search != "" {
			fmt.Printf("No projects found matching '%s'.\n", search)
		} else {
			fmt.Println("No projects found in workspace.")
		}
		return nil
	}

	fmt.Printf("\nAvailable projects (%d):\n\n", len(projects))
	fmt.Printf("%-5s %-20s %-30s %s\n", "#", "IDENTIFIER", "NAME", "DESCRIPTION")
	fmt.Println(strings.Repeat("-", 90))

	for i, p := range projects {
		desc := truncate(p.Description, 30)
		if desc == "" {
			desc = "-"
		}
		fmt.Printf("%-5d %-20s %-30s %s\n", i+1, p.Identifier, truncate(p.Name, 30), desc)
	}

	fmt.Println()
	return nil
}

func runProjectSelect(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	workspace, _ := cmd.Flags().GetString("workspace")
	if workspace == "" {
		if cfg.PlaneWorkspace != "" {
			workspace = cfg.PlaneWorkspace
		} else {
			workspace = extractWorkspaceFromURL(cfg.PlaneBaseURL)
		}
	}

	client, err := plane.NewClient(cfg.PlaneBaseURL, cfg.PlaneAPIToken)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	client.SetWorkspace(workspace)

	// Fetch projects
	projects, err := client.GetProjects()
	if err != nil {
		return fmt.Errorf("failed to fetch projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	// Build options
	var options []string
	for _, p := range projects {
		options = append(options, fmt.Sprintf("%s (%s)", p.Name, p.Identifier))
	}

	idx, err := selectOption("\nSelect a project:", options)
	if err != nil {
		if err.Error() == "cancelled by user" {
			fmt.Println("Selection cancelled.")
			return nil
		}
		return err
	}

	selected := projects[idx]
	fmt.Printf("\n✓ Selected project: %s (%s)\n", selected.Name, selected.Identifier)
	fmt.Printf("\nUse this project with: --project %s\n", selected.Identifier)
	fmt.Printf("Or set as default in config.yaml:\n  defaults:\n    project: \"%s\"\n", selected.Identifier)

	return nil
}

// InteractiveProjectSelector allows selecting a project interactively
func InteractiveProjectSelector(client *plane.Client) (*plane.Project, error) {
	projects, err := client.GetProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects found")
	}

	if len(projects) == 1 {
		return &projects[0], nil
	}

	// Build options list
	var options []string
	for _, p := range projects {
		options = append(options, fmt.Sprintf("%s (%s)", p.Name, p.Identifier))
	}

	// Use survey for selection
	idx, err := selectOption("Select a project:", options)
	if err != nil {
		return nil, err
	}

	return &projects[idx], nil
}
