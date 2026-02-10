package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"plane-cli/internal/config"
	"plane-cli/internal/plane"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive mode for all Plane CLI features",
	Long: `Launch interactive mode with a menu to access all features:
- Work Items: Update work items with guided workflow
- Modules: Create, update, delete project modules  
- Labels: Manage project labels
- Pages: Create and manage project pages

This is the easiest way to use the CLI without remembering all commands.`,
	RunE: runInteractive,
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
	interactiveCmd.Flags().String("workspace", "", "Workspace identifier")
}

func runInteractive(cmd *cobra.Command, args []string) error {
	// Check and prompt for configuration if missing
	cfg, wasConfigured, err := config.CheckAndPromptConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	if wasConfigured {
		// User just configured the CLI, show success message
		fmt.Println("\n✨ Configuration complete! Continuing to interactive mode...\n")
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

	for {
		fmt.Println("\n" + strings.Repeat("=", 70))
		fmt.Println("                    🚀 PLANE CLI - INTERACTIVE MODE")
		fmt.Println(strings.Repeat("=", 70))

		options := []string{
			"📋 Work Items - Update work items (description, assignees, state, etc.)",
			"📦 Modules - Create, update, delete project modules",
			"🏷️  Labels - Manage project labels and tags",
			"📄 Pages - Create and manage project documentation pages",
			"🚪 Exit",
		}

		idx, err := selectOption("Select an option:", options)
		if err != nil {
			if err.Error() == "cancelled by user" {
				fmt.Println("\n👋 Goodbye!")
				return nil
			}
			return err
		}

		switch idx {
		case 0:
			if err := runWorkItemInteractive(client); err != nil {
				fmt.Printf("\n❌ Error: %v\n", err)
			}

		case 1:
			if err := runModuleInteractiveSubmenu(client); err != nil {
				fmt.Printf("\n❌ Error: %v\n", err)
			}

		case 2:
			if err := runLabelInteractiveSubmenu(client); err != nil {
				fmt.Printf("\n❌ Error: %v\n", err)
			}

		case 3:
			if err := runPageInteractiveSubmenu(client); err != nil {
				fmt.Printf("\n❌ Error: %v\n", err)
			}

		case 4:
			fmt.Println("\n👋 Goodbye!")
			return nil
		}

		fmt.Println("\nPress Enter to continue...")
		input("")
	}
}

// Work Items Interactive
func runWorkItemInteractive(client *plane.Client) error {
	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("                    📋 WORK ITEMS")
	fmt.Println(strings.Repeat("-", 70))

	// Step 1: Select Project
	project, err := selectProjectInteractive(client)
	if err != nil {
		return err
	}

	// Step 2: Search for Work Item
	workItem, err := searchAndSelectWorkItem(client, project.ID, 60)
	if err != nil {
		return err
	}

	// Step 3: Choose what to update
	update, err := chooseUpdateFields(client, project.ID)
	if err != nil {
		return err
	}

	if update == nil {
		fmt.Println("\nNo changes selected.")
		return nil
	}

	// Step 4: Confirm and apply
	fmt.Printf("\n📋 Update Summary:\n")
	fmt.Printf("   Work Item: %s-%d (%s)\n", project.Identifier, workItem.SequenceID, workItem.Name)
	printUpdatePreview(update)

	confirmed, err := confirm("\nApply these changes?")
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("❌ Update cancelled.")
		return nil
	}

	updated, err := client.UpdateWorkItem(project.ID, workItem.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update work item: %w", err)
	}

	fmt.Printf("\n✅ Successfully updated work item!\n")
	fmt.Printf("   ID: %s-%d\n", project.Identifier, updated.SequenceID)
	fmt.Printf("   Title: %s\n", updated.Name)
	if update.DescriptionHTML != "" {
		fmt.Printf("   Description: %d characters\n", len(updated.DescriptionHTML))
	}

	return nil
}

// Module Interactive Submenu
func runModuleInteractiveSubmenu(client *plane.Client) error {
	// Step 1: Select Project
	project, err := selectProjectInteractive(client)
	if err != nil {
		return err
	}

	for {
		fmt.Println("\n" + strings.Repeat("-", 70))
		fmt.Println("                    📦 MODULES")
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("Project: %s\n\n", project.Name)

		options := []string{
			"List all modules",
			"Create new module",
			"Update module",
			"Delete module",
			"Back to main menu",
		}

		idx, err := selectOption("Select an action:", options)
		if err != nil {
			if err.Error() == "cancelled by user" {
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
			return nil
		}
	}
}

// Label Interactive Submenu
func runLabelInteractiveSubmenu(client *plane.Client) error {
	// Step 1: Select Project
	project, err := selectProjectInteractive(client)
	if err != nil {
		return err
	}

	for {
		fmt.Println("\n" + strings.Repeat("-", 70))
		fmt.Println("                    🏷️  LABELS")
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("Project: %s\n\n", project.Name)

		options := []string{
			"List all labels",
			"Create new label",
			"Update label",
			"Delete label",
			"Back to main menu",
		}

		idx, err := selectOption("Select an action:", options)
		if err != nil {
			if err.Error() == "cancelled by user" {
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
			return nil
		}
	}
}

// Page Interactive Submenu
func runPageInteractiveSubmenu(client *plane.Client) error {
	// Step 1: Select Project
	project, err := selectProjectInteractive(client)
	if err != nil {
		return err
	}

	for {
		fmt.Println("\n" + strings.Repeat("-", 70))
		fmt.Println("                    📄 PAGES")
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("Project: %s\n\n", project.Name)

		options := []string{
			"List all pages",
			"Create new page",
			"Update page",
			"Delete page",
			"Back to main menu",
		}

		idx, err := selectOption("Select an action:", options)
		if err != nil {
			if err.Error() == "cancelled by user" {
				return nil
			}
			return err
		}

		switch idx {
		case 0:
			if err := listPagesInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 1:
			if err := createPageInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 2:
			if err := updatePageInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 3:
			if err := deletePageInteractive(client, project.ID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case 4:
			return nil
		}
	}
}
