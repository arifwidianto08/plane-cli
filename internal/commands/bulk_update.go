package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"plane-cli/internal/config"
	"plane-cli/internal/fuzzy"
	"plane-cli/internal/plane"
)

var bulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Bulk update multiple work items at once",
	Long: `Update multiple work items simultaneously with the same values.

You can bulk update:
- Assignees (add or replace)
- Estimate points
- Labels (add or replace)
- Module
- State
- Priority

Examples:
  # Interactive bulk update
  plane-cli bulk-update --project c20fcc54-c675-47c4-85db-a4acdde3c9e1

  # Bulk update by search pattern
  plane-cli bulk-update --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 --search "BE" --assignees user-id-1,user-id-2

  # Bulk update with confirmation
  plane-cli bulk-update --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 --search "SaaS" --state "In Progress" --dry-run`,
	RunE: runBulkUpdate,
}

func init() {
	rootCmd.AddCommand(bulkUpdateCmd)

	// Required flags
	bulkUpdateCmd.Flags().String("project", "", "Project identifier (required)")
	bulkUpdateCmd.MarkFlagRequired("project")

	// Search/Selection flags
	bulkUpdateCmd.Flags().String("search", "", "Search term to find work items (if not provided, uses interactive selection)")
	bulkUpdateCmd.Flags().Int("min-score", 60, "Minimum fuzzy match score (0-100)")

	// Update flags
	bulkUpdateCmd.Flags().StringSlice("assignees", nil, "Assignee user IDs (comma-separated)")
	bulkUpdateCmd.Flags().Bool("replace-assignees", false, "Replace existing assignees instead of adding")
	bulkUpdateCmd.Flags().Float64("estimate", -1, "Estimate points (use -1 to skip)")
	bulkUpdateCmd.Flags().StringSlice("labels", nil, "Label IDs (comma-separated)")
	bulkUpdateCmd.Flags().Bool("replace-labels", false, "Replace existing labels instead of adding")
	bulkUpdateCmd.Flags().String("module", "", "Module ID")
	bulkUpdateCmd.Flags().String("state", "", "State name")
	bulkUpdateCmd.Flags().String("priority", "", "Priority (urgent, high, medium, low)")

	// Behavior flags
	bulkUpdateCmd.Flags().Bool("dry-run", false, "Preview changes without applying")
	bulkUpdateCmd.Flags().Bool("interactive", false, "Force interactive mode even with flags")
}

func runBulkUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n\n💡 To configure the CLI, run: plane-cli configure", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	searchTerm, _ := cmd.Flags().GetString("search")
	minScore, _ := cmd.Flags().GetInt("min-score")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	forceInteractive, _ := cmd.Flags().GetBool("interactive")

	// Get update values from flags
	assignees, _ := cmd.Flags().GetStringSlice("assignees")
	replaceAssignees, _ := cmd.Flags().GetBool("replace-assignees")
	estimate, _ := cmd.Flags().GetFloat64("estimate")
	labels, _ := cmd.Flags().GetStringSlice("labels")
	replaceLabels, _ := cmd.Flags().GetBool("replace-labels")
	moduleID, _ := cmd.Flags().GetString("module")
	state, _ := cmd.Flags().GetString("state")
	priorityStr, _ := cmd.Flags().GetString("priority")

	workspace := cfg.PlaneWorkspace
	if workspace == "" {
		workspace = extractWorkspaceFromURL(cfg.PlaneBaseURL)
	}

	client, err := plane.NewClient(cfg.PlaneBaseURL, cfg.PlaneAPIToken)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	client.SetWorkspace(workspace)

	// Fetch all work items
	fmt.Printf("📥 Fetching work items from project '%s'...\n", projectID)
	allWorkItems, err := fetchAllWorkItemsForProject(client, projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch work items: %w", err)
	}

	if len(allWorkItems) == 0 {
		return fmt.Errorf("no work items found in this project")
	}

	// Select work items to update
	var selectedWorkItems []plane.WorkItem

	if searchTerm != "" && !forceInteractive {
		// Use search pattern
		fmt.Printf("🔍 Searching for work items matching '%s'...\n", searchTerm)
		titles := make([]string, len(allWorkItems))
		for i, item := range allWorkItems {
			titles[i] = item.Name
		}

		matcher := fuzzy.NewMatcher(minScore)
		matches := matcher.FindMatches(searchTerm, titles)

		// Fallback to substring matching
		if len(matches) == 0 {
			searchLower := strings.ToLower(searchTerm)
			for i, title := range titles {
				if strings.Contains(strings.ToLower(title), searchLower) {
					matches = append(matches, fuzzy.MatchResult{
						Index: i,
						Score: 50,
					})
				}
			}
		}

		if len(matches) == 0 {
			return fmt.Errorf("no work items found matching '%s'", searchTerm)
		}

		for _, match := range matches {
			selectedWorkItems = append(selectedWorkItems, allWorkItems[match.Index])
		}

		fmt.Printf("✓ Found %d matching work items\n", len(selectedWorkItems))
	} else {
		// Interactive selection
		selectedWorkItems, err = selectMultipleWorkItemsInteractive(allWorkItems)
		if err != nil {
			return err
		}
	}

	if len(selectedWorkItems) == 0 {
		return fmt.Errorf("no work items selected")
	}

	// Build update payload
	update := &plane.WorkItemUpdate{}
	hasUpdates := false

	// Determine what to update
	if len(assignees) > 0 || forceInteractive {
		if forceInteractive && len(assignees) == 0 {
			// Interactive assignee selection
			newAssignees, replace, err := selectAssigneesInteractive(client, projectID, selectedWorkItems)
			if err != nil {
				return err
			}
			if len(newAssignees) > 0 {
				assignees = newAssignees
				replaceAssignees = replace
			}
		}

		if len(assignees) > 0 {
			if replaceAssignees {
				update.Assignees = assignees
			} else {
				// Merge with existing assignees
				update.Assignees = mergeSlices(getAllAssignees(selectedWorkItems), assignees)
			}
			hasUpdates = true
		}
	}

	if estimate >= 0 || forceInteractive {
		if forceInteractive && estimate < 0 {
			newEstimate, err := selectEstimateInteractive()
			if err != nil {
				return err
			}
			if newEstimate >= 0 {
				estimate = newEstimate
			}
		}
		if estimate >= 0 {
			update.EstimatePoint = estimate
			hasUpdates = true
		}
	}

	if len(labels) > 0 || forceInteractive {
		if forceInteractive && len(labels) == 0 {
			newLabels, replace, err := selectLabelsInteractive(client, projectID)
			if err != nil {
				return err
			}
			if len(newLabels) > 0 {
				labels = newLabels
				replaceLabels = replace
			}
		}

		if len(labels) > 0 {
			if replaceLabels {
				update.Labels = labels
			} else {
				// Merge with existing labels
				update.Labels = mergeSlices(getAllLabels(selectedWorkItems), labels)
			}
			hasUpdates = true
		}
	}

	if moduleID != "" || forceInteractive {
		if forceInteractive && moduleID == "" {
			newModule, err := selectModuleInteractive(client, projectID)
			if err != nil {
				return err
			}
			moduleID = newModule
		}
		if moduleID != "" {
			update.Module = moduleID
			hasUpdates = true
		}
	}

	if state != "" {
		update.State = state
		hasUpdates = true
	}

	if priorityStr != "" {
		update.Priority = priorityStr
		hasUpdates = true
	}

	if !hasUpdates {
		fmt.Println("\n⚠️  No updates specified. Use flags or --interactive mode.")
		return nil
	}

	// Preview changes
	fmt.Printf("\n📋 Bulk Update Preview:\n")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Project: %s\n", projectID)
	fmt.Printf("Work items to update: %d\n", len(selectedWorkItems))
	fmt.Println("\nSelected work items:")
	for _, item := range selectedWorkItems {
		fmt.Printf("  • [%d] %s\n", item.SequenceID, truncate(item.Name, 50))
	}
	fmt.Println("\nUpdates to apply:")
	printUpdatePreview(update)
	fmt.Println(strings.Repeat("-", 70))

	if dryRun {
		fmt.Println("\n📝 Dry run mode - no changes made.")
		return nil
	}

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
		_, err := client.UpdateWorkItem(projectID, item.ID, update)
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

func selectMultipleWorkItemsInteractive(workItems []plane.WorkItem) ([]plane.WorkItem, error) {
	fmt.Println("\n🔍 Select Work Items to Update")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("You can select multiple work items using SPACE, then press ENTER")
	fmt.Println(strings.Repeat("-", 70))

	// Build options
	var options []string
	for _, item := range workItems {
		options = append(options, fmt.Sprintf("[%d] %s", item.SequenceID, truncate(item.Name, 60)))
	}

	// Use multi-select
	indices, err := selectMultiOption("Select work items (SPACE to select, ENTER to confirm):", options)
	if err != nil {
		return nil, err
	}

	var selected []plane.WorkItem
	for _, idx := range indices {
		selected = append(selected, workItems[idx])
	}

	fmt.Printf("✓ Selected %d work items\n", len(selected))
	return selected, nil
}

func selectAssigneesInteractive(client *plane.Client, projectID string, workItems []plane.WorkItem) ([]string, bool, error) {
	fmt.Println("\n👥 Update Assignees")
	fmt.Println(strings.Repeat("-", 70))

	// Show current assignees
	currentAssignees := getAllAssignees(workItems)
	if len(currentAssignees) > 0 {
		fmt.Printf("Current assignees across selected items: %v\n", currentAssignees)
	}

	// Get available members
	members, err := client.GetProjectMembers(projectID)
	if err != nil {
		members, err = client.GetWorkspaceMembers()
		if err != nil {
			return nil, false, fmt.Errorf("failed to get members: %w", err)
		}
	}

	if len(members) == 0 {
		return nil, false, fmt.Errorf("no members found")
	}

	// Build options
	var options []string
	for _, m := range members {
		name := m.GetDisplayName()
		if len(name) > 30 {
			name = name[:27] + "..."
		}
		options = append(options, fmt.Sprintf("%s (%s)", name, m.Email))
	}

	// Ask for action
	actionIdx, err := selectOption("What would you like to do?", []string{
		"Add assignees to existing ones",
		"Replace all assignees",
		"Clear all assignees",
		"Cancel",
	})
	if err != nil {
		return nil, false, err
	}

	switch actionIdx {
	case 0: // Add
		indices, err := selectMultiOption("Select assignees to add:", options)
		if err != nil {
			return nil, false, err
		}
		var selectedIDs []string
		for _, idx := range indices {
			selectedIDs = append(selectedIDs, members[idx].ID)
		}
		return selectedIDs, false, nil

	case 1: // Replace
		indices, err := selectMultiOption("Select new assignees:", options)
		if err != nil {
			return nil, false, err
		}
		var selectedIDs []string
		for _, idx := range indices {
			selectedIDs = append(selectedIDs, members[idx].ID)
		}
		return selectedIDs, true, nil

	case 2: // Clear
		return []string{}, true, nil

	case 3: // Cancel
		return nil, false, fmt.Errorf("cancelled")
	}

	return nil, false, nil
}

func selectEstimateInteractive() (float64, error) {
	fmt.Println("\n📊 Update Estimate Points")
	fmt.Println(strings.Repeat("-", 70))

	result, err := input("Enter estimate points (e.g., 1, 2, 3, 5, 8, 13) or press Enter to skip:")
	if err != nil {
		return -1, err
	}

	if result == "" {
		return -1, nil
	}

	estimate, err := parseFloat(result)
	if err != nil {
		return -1, fmt.Errorf("invalid number: %w", err)
	}

	return estimate, nil
}

func selectLabelsInteractive(client *plane.Client, projectID string) ([]string, bool, error) {
	fmt.Println("\n🏷️  Update Labels")
	fmt.Println(strings.Repeat("-", 70))

	labels, err := client.GetLabels(projectID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get labels: %w", err)
	}

	if len(labels) == 0 {
		return nil, false, fmt.Errorf("no labels found in this project")
	}

	// Build options
	var options []string
	for _, l := range labels {
		options = append(options, l.Name)
	}

	// Ask for action
	actionIdx, err := selectOption("What would you like to do?", []string{
		"Add labels to existing ones",
		"Replace all labels",
		"Clear all labels",
		"Cancel",
	})
	if err != nil {
		return nil, false, err
	}

	switch actionIdx {
	case 0: // Add
		indices, err := selectMultiOption("Select labels to add:", options)
		if err != nil {
			return nil, false, err
		}
		var selectedIDs []string
		for _, idx := range indices {
			selectedIDs = append(selectedIDs, labels[idx].ID)
		}
		return selectedIDs, false, nil

	case 1: // Replace
		indices, err := selectMultiOption("Select new labels:", options)
		if err != nil {
			return nil, false, err
		}
		var selectedIDs []string
		for _, idx := range indices {
			selectedIDs = append(selectedIDs, labels[idx].ID)
		}
		return selectedIDs, true, nil

	case 2: // Clear
		return []string{}, true, nil

	case 3: // Cancel
		return nil, false, fmt.Errorf("cancelled")
	}

	return nil, false, nil
}

func selectModuleInteractive(client *plane.Client, projectID string) (string, error) {
	fmt.Println("\n📦 Update Module")
	fmt.Println(strings.Repeat("-", 70))

	modules, err := client.GetModules(projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get modules: %w", err)
	}

	if len(modules) == 0 {
		return "", fmt.Errorf("no modules found in this project")
	}

	// Build options
	options := []string{"Clear module (remove from work items)"}
	for _, m := range modules {
		options = append(options, m.Name)
	}

	idx, err := selectOption("Select module:", options)
	if err != nil {
		return "", err
	}

	if idx == 0 {
		return "", nil // Clear module
	}

	return modules[idx-1].ID, nil
}

// Helper functions
func getAllAssignees(workItems []plane.WorkItem) []string {
	assigneeMap := make(map[string]bool)
	for _, item := range workItems {
		for _, a := range item.Assignees {
			assigneeMap[a] = true
		}
	}
	var result []string
	for a := range assigneeMap {
		result = append(result, a)
	}
	return result
}

func getAllLabels(workItems []plane.WorkItem) []string {
	labelMap := make(map[string]bool)
	for _, item := range workItems {
		for _, l := range item.Labels {
			labelMap[l] = true
		}
	}
	var result []string
	for l := range labelMap {
		result = append(result, l)
	}
	return result
}

func mergeSlices(existing, new []string) []string {
	seen := make(map[string]bool)
	var result []string

	// Add all existing
	for _, item := range existing {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	// Add new items
	for _, item := range new {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}
