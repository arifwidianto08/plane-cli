package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"plane-cli/internal/config"
	"plane-cli/internal/plane"
)

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Manage project modules",
	Long: `List, create, update, and delete modules in your Plane projects.

Examples:
  # List all modules in a project
  plane-cli module list --project c20fcc54-c675-47c4-85db-a4acdde3c9e1

  # Create a new module
  plane-cli module create --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 --name "Frontend"

  # Update a module
  plane-cli module update --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 --id <module-id> --name "Frontend v2"

  # Delete a module
  plane-cli module delete --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 --id <module-id>

  # Interactive module management
  plane-cli module interactive`,
}

var moduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all modules in a project",
	RunE:  runModuleList,
}

var moduleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new module",
	RunE:  runModuleCreate,
}

var moduleUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing module",
	RunE:  runModuleUpdate,
}

var moduleDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a module",
	RunE:  runModuleDelete,
}

var moduleInteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive module management",
	Long:  `Interactive workflow for managing modules - select project, then create, update, or delete modules.`,
	RunE:  runModuleInteractive,
}

func init() {
	rootCmd.AddCommand(moduleCmd)
	moduleCmd.AddCommand(moduleListCmd)
	moduleCmd.AddCommand(moduleCreateCmd)
	moduleCmd.AddCommand(moduleUpdateCmd)
	moduleCmd.AddCommand(moduleDeleteCmd)
	moduleCmd.AddCommand(moduleInteractiveCmd)

	// List flags
	moduleListCmd.Flags().String("project", "", "Project identifier (required)")
	moduleListCmd.MarkFlagRequired("project")

	// Create flags
	moduleCreateCmd.Flags().String("project", "", "Project identifier (required)")
	moduleCreateCmd.Flags().String("name", "", "Module name (required)")
	moduleCreateCmd.Flags().String("description", "", "Module description")
	moduleCreateCmd.Flags().String("color", "", "Module color (hex code)")
	moduleCreateCmd.Flags().String("status", "backlog", "Module status (backlog, started, paused, completed, cancelled)")
	moduleCreateCmd.MarkFlagRequired("project")
	moduleCreateCmd.MarkFlagRequired("name")

	// Update flags
	moduleUpdateCmd.Flags().String("project", "", "Project identifier (required)")
	moduleUpdateCmd.Flags().String("id", "", "Module ID (required)")
	moduleUpdateCmd.Flags().String("name", "", "New module name")
	moduleUpdateCmd.Flags().String("description", "", "New module description")
	moduleUpdateCmd.Flags().String("color", "", "New module color")
	moduleUpdateCmd.Flags().String("status", "", "New module status")
	moduleUpdateCmd.MarkFlagRequired("project")
	moduleUpdateCmd.MarkFlagRequired("id")

	// Delete flags
	moduleDeleteCmd.Flags().String("project", "", "Project identifier (required)")
	moduleDeleteCmd.Flags().String("id", "", "Module ID (required)")
	moduleDeleteCmd.MarkFlagRequired("project")
	moduleDeleteCmd.MarkFlagRequired("id")
}

func runModuleList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
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

	modules, err := client.GetModules(projectID)
	if err != nil {
		return fmt.Errorf("failed to get modules: %w", err)
	}

	if len(modules) == 0 {
		fmt.Println("No modules found in this project.")
		return nil
	}

	fmt.Printf("\n📦 Modules (%d):\n\n", len(modules))
	fmt.Printf("%-5s %-36s %-20s %-10s %s\n", "#", "ID", "NAME", "STATUS", "DESCRIPTION")
	fmt.Println(strings.Repeat("-", 100))

	for i, m := range modules {
		desc := truncate(m.Description, 30)
		if desc == "" {
			desc = "-"
		}
		name := truncate(m.Name, 18)
		status := m.Status
		if status == "" {
			status = "backlog"
		}
		fmt.Printf("%-5d %-36s %-20s %-10s %s\n", i+1, m.ID, name, status, desc)
	}

	fmt.Println()
	return nil
}

func runModuleCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	color, _ := cmd.Flags().GetString("color")
	status, _ := cmd.Flags().GetString("status")
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

	create := &plane.ModuleCreate{
		Name:        name,
		Description: description,
		Color:       color,
		Status:      status,
	}

	module, err := client.CreateModule(projectID, create)
	if err != nil {
		return fmt.Errorf("failed to create module: %w", err)
	}

	fmt.Printf("\n✅ Created module:\n")
	fmt.Printf("   ID: %s\n", module.ID)
	fmt.Printf("   Name: %s\n", module.Name)
	if module.Description != "" {
		fmt.Printf("   Description: %s\n", module.Description)
	}

	return nil
}

func runModuleUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	moduleID, _ := cmd.Flags().GetString("id")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	color, _ := cmd.Flags().GetString("color")
	status, _ := cmd.Flags().GetString("status")
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

	update := &plane.ModuleUpdate{}
	if name != "" {
		update.Name = name
	}
	if description != "" {
		update.Description = description
	}
	if color != "" {
		update.Color = color
	}
	if status != "" {
		update.Status = status
	}

	module, err := client.UpdateModule(projectID, moduleID, update)
	if err != nil {
		return fmt.Errorf("failed to update module: %w", err)
	}

	fmt.Printf("\n✅ Updated module:\n")
	fmt.Printf("   ID: %s\n", module.ID)
	fmt.Printf("   Name: %s\n", module.Name)

	return nil
}

func runModuleDelete(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	moduleID, _ := cmd.Flags().GetString("id")
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

	// Get module info for confirmation
	module, err := client.GetModule(projectID, moduleID)
	if err != nil {
		return fmt.Errorf("failed to get module: %w", err)
	}

	confirmed, err := confirm(fmt.Sprintf("Are you sure you want to delete module '%s'?", module.Name))
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("❌ Deletion cancelled.")
		return nil
	}

	if err := client.DeleteModule(projectID, moduleID); err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}

	fmt.Println("\n✅ Module deleted successfully.")
	return nil
}

func runModuleInteractive(cmd *cobra.Command, args []string) error {
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

	// Step 1: Select Project
	project, err := selectProjectInteractive(client)
	if err != nil {
		return err
	}

	// Step 2: Choose action
	for {
		fmt.Println("\n📦 Module Management")

		options := []string{
			"List all modules",
			"Create new module",
			"Update module",
			"Delete module",
			"Exit",
		}

		idx, err := selectOption("Select an action:", options)
		if err != nil {
			if err.Error() == "cancelled by user" {
				fmt.Println("\n👋 Goodbye!")
				return nil
			}
			return err
		}

		switch idx {
		case 0:
			if err := listModulesInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 1:
			if err := createModuleInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 2:
			if err := updateModuleInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 3:
			if err := deleteModuleInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 4:
			fmt.Println("\n👋 Goodbye!")
			return nil
		}
	}
}

func listModulesInteractive(client *plane.Client, projectID string) error {
	modules, err := client.GetModules(projectID)
	if err != nil {
		return err
	}

	if len(modules) == 0 {
		fmt.Println("\nNo modules found.")
		return nil
	}

	fmt.Printf("\n📦 Modules (%d):\n\n", len(modules))
	fmt.Printf("%-5s %-36s %-25s %-12s\n", "#", "ID", "NAME", "STATUS")
	fmt.Println(strings.Repeat("-", 80))

	for i, m := range modules {
		name := truncate(m.Name, 23)
		status := m.Status
		if status == "" {
			status = "backlog"
		}
		fmt.Printf("%-5d %-36s %-25s %-12s\n", i+1, m.ID, name, status)
	}

	fmt.Println()
	return nil
}

func createModuleInteractive(client *plane.Client, projectID string) error {
	fmt.Println("\n➕ Create New Module")

	name, err := input("Module name:")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("module name is required")
	}

	description, err := inputWithDefault("Description (optional):", "")
	if err != nil {
		return err
	}

	statusOptions := []string{
		"Backlog",
		"Started",
		"Paused",
		"Completed",
		"Cancelled",
	}

	statusIdx, err := selectOption("Select status:", statusOptions)
	if err != nil {
		return err
	}

	statusValues := []string{"backlog", "started", "paused", "completed", "cancelled"}
	status := statusValues[statusIdx]

	create := &plane.ModuleCreate{
		Name:        name,
		Description: description,
		Status:      status,
	}

	module, err := client.CreateModule(projectID, create)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Created module: %s (ID: %s)\n", module.Name, module.ID)
	return nil
}

func updateModuleInteractive(client *plane.Client, projectID string) error {
	modules, err := client.GetModules(projectID)
	if err != nil {
		return err
	}

	if len(modules) == 0 {
		return fmt.Errorf("no modules found")
	}

	// Build options
	var options []string
	for _, m := range modules {
		options = append(options, m.Name)
	}

	idx, err := selectOption("Select module to update:", options)
	if err != nil {
		return err
	}

	module := modules[idx]

	fmt.Printf("\n✏️  Update Module: %s\n", module.Name)

	update := &plane.ModuleUpdate{}

	name, err := inputWithDefault(fmt.Sprintf("New name (current: %s):", module.Name), "")
	if err != nil {
		return err
	}
	if name != "" {
		update.Name = name
	}

	desc, err := inputWithDefault(fmt.Sprintf("New description (current: %s):", truncate(module.Description, 20)), "")
	if err != nil {
		return err
	}
	if desc != "" {
		update.Description = desc
	}

	statusOptions := []string{
		"Keep current (" + module.Status + ")",
		"Backlog",
		"Started",
		"Paused",
		"Completed",
		"Cancelled",
	}

	statusIdx, err := selectOption("Select status:", statusOptions)
	if err != nil {
		return err
	}

	if statusIdx > 0 {
		statusValues := []string{"", "backlog", "started", "paused", "completed", "cancelled"}
		update.Status = statusValues[statusIdx]
	}

	updated, err := client.UpdateModule(projectID, module.ID, update)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Updated module: %s\n", updated.Name)
	return nil
}

func deleteModuleInteractive(client *plane.Client, projectID string) error {
	modules, err := client.GetModules(projectID)
	if err != nil {
		return err
	}

	if len(modules) == 0 {
		return fmt.Errorf("no modules found")
	}

	// Build options
	var options []string
	for _, m := range modules {
		options = append(options, m.Name)
	}

	idx, err := selectOption("Select module to delete:", options)
	if err != nil {
		return err
	}

	module := modules[idx]

	confirmed, err := confirm(fmt.Sprintf("Delete module '%s'?", module.Name))
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("❌ Deletion cancelled.")
		return nil
	}

	if err := client.DeleteModule(projectID, module.ID); err != nil {
		return err
	}

	fmt.Println("\n✅ Module deleted.")
	return nil
}
