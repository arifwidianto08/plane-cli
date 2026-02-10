package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"plane-cli/internal/config"
	"plane-cli/internal/plane"
)

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage project labels",
	Long: `List, create, update, and delete labels in your Plane projects.

Examples:
  # List all labels in a project
  plane-cli label list --project c20fcc54-c675-47c4-85db-a4acdde3c9e1

  # Create a new label
  plane-cli label create --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 --name "bug" --color "#ff0000"

  # Update a label
  plane-cli label update --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 --id <label-id> --name "Bug"

  # Delete a label
  plane-cli label delete --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 --id <label-id>

  # Interactive label management
  plane-cli label interactive`,
}

var labelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all labels in a project",
	RunE:  runLabelList,
}

var labelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new label",
	RunE:  runLabelCreate,
}

var labelUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing label",
	RunE:  runLabelUpdate,
}

var labelDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a label",
	RunE:  runLabelDelete,
}

var labelInteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive label management",
	Long:  `Interactive workflow for managing labels - select project, then create, update, or delete labels.`,
	RunE:  runLabelInteractive,
}

func init() {
	rootCmd.AddCommand(labelCmd)
	labelCmd.AddCommand(labelListCmd)
	labelCmd.AddCommand(labelCreateCmd)
	labelCmd.AddCommand(labelUpdateCmd)
	labelCmd.AddCommand(labelDeleteCmd)
	labelCmd.AddCommand(labelInteractiveCmd)

	// List flags
	labelListCmd.Flags().String("project", "", "Project identifier (required)")
	labelListCmd.MarkFlagRequired("project")

	// Create flags
	labelCreateCmd.Flags().String("project", "", "Project identifier (required)")
	labelCreateCmd.Flags().String("name", "", "Label name (required)")
	labelCreateCmd.Flags().String("color", "", "Label color (hex code, e.g., #ff0000)")
	labelCreateCmd.MarkFlagRequired("project")
	labelCreateCmd.MarkFlagRequired("name")

	// Update flags
	labelUpdateCmd.Flags().String("project", "", "Project identifier (required)")
	labelUpdateCmd.Flags().String("id", "", "Label ID (required)")
	labelUpdateCmd.Flags().String("name", "", "New label name")
	labelUpdateCmd.Flags().String("color", "", "New label color")
	labelUpdateCmd.MarkFlagRequired("project")
	labelUpdateCmd.MarkFlagRequired("id")

	// Delete flags
	labelDeleteCmd.Flags().String("project", "", "Project identifier (required)")
	labelDeleteCmd.Flags().String("id", "", "Label ID (required)")
	labelDeleteCmd.MarkFlagRequired("project")
	labelDeleteCmd.MarkFlagRequired("id")
}

func runLabelList(cmd *cobra.Command, args []string) error {
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

	labels, err := client.GetLabels(projectID)
	if err != nil {
		return fmt.Errorf("failed to get labels: %w", err)
	}

	if len(labels) == 0 {
		fmt.Println("No labels found in this project.")
		return nil
	}

	fmt.Printf("\n🏷️  Labels (%d):\n\n", len(labels))
	fmt.Printf("%-5s %-36s %-20s %s\n", "#", "ID", "NAME", "COLOR")
	fmt.Println(strings.Repeat("-", 70))

	for i, l := range labels {
		color := l.Color
		if color == "" {
			color = "-"
		}
		fmt.Printf("%-5d %-36s %-20s %s\n", i+1, l.ID, l.Name, color)
	}

	fmt.Println()
	return nil
}

func runLabelCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	name, _ := cmd.Flags().GetString("name")
	color, _ := cmd.Flags().GetString("color")
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

	create := &plane.LabelCreate{
		Name:  name,
		Color: color,
	}

	label, err := client.CreateLabel(projectID, create)
	if err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}

	fmt.Printf("\n✅ Created label:\n")
	fmt.Printf("   ID: %s\n", label.ID)
	fmt.Printf("   Name: %s\n", label.Name)
	if label.Color != "" {
		fmt.Printf("   Color: %s\n", label.Color)
	}

	return nil
}

func runLabelUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	labelID, _ := cmd.Flags().GetString("id")
	name, _ := cmd.Flags().GetString("name")
	color, _ := cmd.Flags().GetString("color")
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

	update := &plane.LabelUpdate{}
	if name != "" {
		update.Name = name
	}
	if color != "" {
		update.Color = color
	}

	label, err := client.UpdateLabel(projectID, labelID, update)
	if err != nil {
		return fmt.Errorf("failed to update label: %w", err)
	}

	fmt.Printf("\n✅ Updated label:\n")
	fmt.Printf("   ID: %s\n", label.ID)
	fmt.Printf("   Name: %s\n", label.Name)

	return nil
}

func runLabelDelete(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	labelID, _ := cmd.Flags().GetString("id")
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

	// Get label info for confirmation
	label, err := client.GetLabel(projectID, labelID)
	if err != nil {
		return fmt.Errorf("failed to get label: %w", err)
	}

	confirmed, err := confirm(fmt.Sprintf("Are you sure you want to delete label '%s'?", label.Name))
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("❌ Deletion cancelled.")
		return nil
	}

	if err := client.DeleteLabel(projectID, labelID); err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	fmt.Println("\n✅ Label deleted successfully.")
	return nil
}

func runLabelInteractive(cmd *cobra.Command, args []string) error {
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
		fmt.Println("\n🏷️  Label Management")

		options := []string{
			"List all labels",
			"Create new label",
			"Update label",
			"Delete label",
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
			if err := listLabelsInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 1:
			if err := createLabelInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 2:
			if err := updateLabelInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 3:
			if err := deleteLabelInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 4:
			fmt.Println("\n👋 Goodbye!")
			return nil
		}
	}
}

func listLabelsInteractive(client *plane.Client, projectID string) error {
	labels, err := client.GetLabels(projectID)
	if err != nil {
		return err
	}

	if len(labels) == 0 {
		fmt.Println("\nNo labels found.")
		return nil
	}

	fmt.Printf("\n🏷️  Labels (%d):\n\n", len(labels))
	fmt.Printf("%-5s %-36s %-20s %s\n", "#", "ID", "NAME", "COLOR")
	fmt.Println(strings.Repeat("-", 70))

	for i, l := range labels {
		color := l.Color
		if color == "" {
			color = "-"
		}
		fmt.Printf("%-5d %-36s %-20s %s\n", i+1, l.ID, l.Name, color)
	}

	fmt.Println()
	return nil
}

func createLabelInteractive(client *plane.Client, projectID string) error {
	fmt.Println("\n➕ Create New Label")

	name, err := input("Label name:")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("label name is required")
	}

	color, err := inputWithDefault("Color (hex code, e.g., #ff0000):", "")
	if err != nil {
		return err
	}

	create := &plane.LabelCreate{
		Name:  name,
		Color: color,
	}

	label, err := client.CreateLabel(projectID, create)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Created label: %s (ID: %s)\n", label.Name, label.ID)
	return nil
}

func updateLabelInteractive(client *plane.Client, projectID string) error {
	labels, err := client.GetLabels(projectID)
	if err != nil {
		return err
	}

	if len(labels) == 0 {
		return fmt.Errorf("no labels found")
	}

	// Build options
	var options []string
	for _, l := range labels {
		options = append(options, l.Name)
	}

	idx, err := selectOption("Select label to update:", options)
	if err != nil {
		return err
	}

	label := labels[idx]

	fmt.Printf("\n✏️  Update Label: %s\n", label.Name)

	update := &plane.LabelUpdate{}

	name, err := inputWithDefault(fmt.Sprintf("New name (current: %s):", label.Name), "")
	if err != nil {
		return err
	}
	if name != "" {
		update.Name = name
	}

	color, err := inputWithDefault(fmt.Sprintf("New color (current: %s):", label.Color), "")
	if err != nil {
		return err
	}
	if color != "" {
		update.Color = color
	}

	updated, err := client.UpdateLabel(projectID, label.ID, update)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Updated label: %s\n", updated.Name)
	return nil
}

func deleteLabelInteractive(client *plane.Client, projectID string) error {
	labels, err := client.GetLabels(projectID)
	if err != nil {
		return err
	}

	if len(labels) == 0 {
		return fmt.Errorf("no labels found")
	}

	// Build options
	var options []string
	for _, l := range labels {
		options = append(options, l.Name)
	}

	idx, err := selectOption("Select label to delete:", options)
	if err != nil {
		return err
	}

	label := labels[idx]

	confirmed, err := confirm(fmt.Sprintf("Delete label '%s'?", label.Name))
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("❌ Deletion cancelled.")
		return nil
	}

	if err := client.DeleteLabel(projectID, label.ID); err != nil {
		return err
	}

	fmt.Println("\n✅ Label deleted.")
	return nil
}
