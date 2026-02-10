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
			"📋 Work Items - Update single work item",
			"⚡ Bulk Update - Update multiple work items at once",
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
			if err := runBulkUpdateInteractive(client); err != nil {
				fmt.Printf("\n❌ Error: %v\n", err)
			}

		case 2:
			if err := runModuleInteractiveSubmenu(client); err != nil {
				fmt.Printf("\n❌ Error: %v\n", err)
			}

		case 3:
			if err := runLabelInteractiveSubmenu(client); err != nil {
				fmt.Printf("\n❌ Error: %v\n", err)
			}

		case 4:
			if err := runPageInteractiveSubmenu(client); err != nil {
				fmt.Printf("\n❌ Error: %v\n", err)
			}

		case 5:
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

// Bulk Update Interactive
func runBulkUpdateInteractive(client *plane.Client) error {
	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("                    ⚡ BULK UPDATE")
	fmt.Println(strings.Repeat("-", 70))

	// Step 1: Select Project
	project, err := selectProjectInteractive(client)
	if err != nil {
		return err
	}

	// Fetch all work items
	fmt.Printf("\n📥 Fetching work items from project '%s'...\n", project.Name)
	allWorkItems, err := fetchAllWorkItemsForProject(client, project.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch work items: %w", err)
	}

	if len(allWorkItems) == 0 {
		return fmt.Errorf("no work items found in this project")
	}

	// Select work items
	fmt.Printf("\nFound %d work items. Select which ones to update:\n", len(allWorkItems))
	selectedWorkItems, err := selectMultipleWorkItemsInteractive(allWorkItems)
	if err != nil {
		return err
	}

	if len(selectedWorkItems) == 0 {
		return fmt.Errorf("no work items selected")
	}

	// Choose what to update
	update, err := chooseBulkUpdateFields(client, project.ID, selectedWorkItems)
	if err != nil {
		return err
	}

	if update == nil {
		fmt.Println("\nNo changes selected.")
		return nil
	}

	// Preview changes
	fmt.Printf("\n📋 Bulk Update Preview:\n")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Project: %s\n", project.Name)
	fmt.Printf("Work items to update: %d\n", len(selectedWorkItems))
	fmt.Println("\nSelected work items:")
	for _, item := range selectedWorkItems {
		fmt.Printf("  • [%d] %s\n", item.SequenceID, truncate(item.Name, 50))
	}
	fmt.Println("\nUpdates to apply:")
	printUpdatePreview(update)
	fmt.Println(strings.Repeat("-", 70))

	// Confirm
	confirmed, err := confirm("\nApply these updates to all selected work items?")
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("\n❌ Update cancelled.")
		return nil
	}

	// Apply updates
	fmt.Printf("\n🔄 Updating %d work items...\n\n", len(selectedWorkItems))

	successCount := 0
	failCount := 0

	for _, item := range selectedWorkItems {
		_, err := client.UpdateWorkItem(project.ID, item.ID, update)
		if err != nil {
			fmt.Printf("  ❌ Failed: [%d] %s - %v\n", item.SequenceID, truncate(item.Name, 40), err)
			failCount++
		} else {
			fmt.Printf("  ✅ Updated: [%d] %s\n", item.SequenceID, truncate(item.Name, 40))
			successCount++
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("-", 70))
	fmt.Printf("✅ Completed: %d/%d work items updated successfully\n", successCount, len(selectedWorkItems))
	if failCount > 0 {
		fmt.Printf("❌ Failed: %d work items\n", failCount)
	}

	return nil
}

// chooseBulkUpdateFields allows selecting which fields to bulk update
func chooseBulkUpdateFields(client *plane.Client, projectID string, workItems []plane.WorkItem) (*plane.WorkItemUpdate, error) {
	update := &plane.WorkItemUpdate{}
	hasUpdates := false

	for {
		fmt.Println("\n" + strings.Repeat("-", 70))
		fmt.Println("Select fields to update (choose one at a time, 'done' when finished):")
		fmt.Println(strings.Repeat("-", 70))

		options := []string{
			"Assignees",
			"Estimate Points",
			"Labels",
			"Module",
			"State",
			"Priority",
			"Done - Apply changes",
			"Cancel",
		}

		idx, err := selectOption("What would you like to update?", options)
		if err != nil {
			return nil, err
		}

		switch idx {
		case 0: // Assignees
			assignees, replace, err := selectAssigneesInteractive(client, projectID, workItems)
			if err != nil {
				if err.Error() == "cancelled" {
					continue
				}
				return nil, err
			}
			if len(assignees) > 0 {
				if replace {
					update.Assignees = assignees
				} else {
					// Merge with existing
					allExisting := getAllAssignees(workItems)
					update.Assignees = mergeSlices(allExisting, assignees)
				}
				hasUpdates = true
				fmt.Println("✓ Assignees updated")
			}

		case 1: // Estimate
			estimate, err := selectEstimateInteractive()
			if err != nil {
				continue
			}
			if estimate >= 0 {
				update.EstimatePoint = estimate
				hasUpdates = true
				fmt.Printf("✓ Estimate set to: %.1f\n", estimate)
			}

		case 2: // Labels
			labels, replace, err := selectLabelsInteractive(client, projectID)
			if err != nil {
				if err.Error() == "cancelled" {
					continue
				}
				return nil, err
			}
			if len(labels) > 0 {
				if replace {
					update.Labels = labels
				} else {
					// Merge with existing
					allExisting := getAllLabels(workItems)
					update.Labels = mergeSlices(allExisting, labels)
				}
				hasUpdates = true
				fmt.Println("✓ Labels updated")
			}

		case 3: // Module
			moduleID, err := selectModuleInteractive(client, projectID)
			if err != nil {
				if err.Error() == "cancelled" {
					continue
				}
				return nil, err
			}
			update.Module = moduleID
			hasUpdates = true
			if moduleID == "" {
				fmt.Println("✓ Module cleared")
			} else {
				fmt.Println("✓ Module updated")
			}

		case 4: // State
			state, err := selectState()
			if err != nil {
				continue
			}
			update.State = state
			hasUpdates = true
			fmt.Printf("✓ State set to: %s\n", state)

		case 5: // Priority
			priority, err := selectPriority()
			if err != nil {
				continue
			}
			update.Priority = priority
			hasUpdates = true
			fmt.Printf("✓ Priority set to: %s\n", priority)

		case 6: // Done
			if !hasUpdates {
				fmt.Println("⚠️  No updates selected. Please select at least one field to update.")
				continue
			}
			return update, nil

		case 7: // Cancel
			return nil, nil
		}
	}
}
