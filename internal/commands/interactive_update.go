package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"plane-cli/internal/config"
	"plane-cli/internal/fuzzy"
	"plane-cli/internal/plane"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var interactiveUpdateCmd = &cobra.Command{
	Use:   "interactive-update",
	Short: "Interactively update work items with guided workflow",
	Long: `Interactive work item updater with step-by-step workflow:

1. Select a project from list
2. Search for work item by name (fuzzy matching)
3. Select the work item to update
4. Choose what to update (description from file, title, state, etc.)
5. Apply the update

Examples:
  # Start interactive update workflow
  plane-cli interactive-update

  # Start with specific project pre-selected
  plane-cli interactive-update --project c20fcc54-c675-47c4-85db-a4acdde3c9e1`,
	RunE: runInteractiveUpdate,
}

func init() {
	rootCmd.AddCommand(interactiveUpdateCmd)

	// Optional flags
	interactiveUpdateCmd.Flags().String("project", "", "Pre-select a project (optional)")
	interactiveUpdateCmd.Flags().String("workspace", "", "Workspace identifier")
	interactiveUpdateCmd.Flags().Int("min-score", 60, "Minimum fuzzy match score (0-100)")
}

func runInteractiveUpdate(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	workspace, _ := cmd.Flags().GetString("workspace")
	minScore, _ := cmd.Flags().GetInt("min-score")

	// Get workspace
	if workspace == "" {
		if cfg.PlaneWorkspace != "" {
			workspace = cfg.PlaneWorkspace
		} else {
			workspace = extractWorkspaceFromURL(cfg.PlaneBaseURL)
		}
	}

	// Create Plane client
	client, err := plane.NewClient(cfg.PlaneBaseURL, cfg.PlaneAPIToken)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	client.SetWorkspace(workspace)

	// Step 1: Select Project
	var project *plane.Project
	if projectID == "" {
		project, err = selectProjectInteractive(client)
		if err != nil {
			return err
		}
		projectID = project.ID
	} else {
		// Verify project exists
		project, err = client.GetProject(projectID)
		if err != nil {
			return fmt.Errorf("failed to get project: %w", err)
		}
		fmt.Printf("\n✓ Using project: %s (%s)\n", project.Name, project.Identifier)
	}

	// Step 2: Search for Work Item
	workItem, err := searchAndSelectWorkItem(client, projectID, minScore)
	if err != nil {
		return err
	}

	// Step 3: Choose what to update
	update, err := chooseUpdateFields(client, projectID)
	if err != nil {
		return err
	}

	// If nothing selected, exit
	if update == nil {
		fmt.Println("\nNo changes selected. Exiting.")
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
		fmt.Println("Update cancelled.")
		return nil
	}

	// Apply update
	updated, err := client.UpdateWorkItem(projectID, workItem.ID, update)
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

func selectProjectInteractive(client *plane.Client) (*plane.Project, error) {
	projects, err := client.GetProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects found in workspace")
	}

	if len(projects) == 1 {
		fmt.Printf("\n✓ Auto-selected only project: %s\n", projects[0].Name)
		return &projects[0], nil
	}

	fmt.Println("\n📁 Step 1: Select a Project")

	// Build options list
	var options []string
	for _, p := range projects {
		options = append(options, fmt.Sprintf("%s (%s)", p.Name, p.Identifier))
	}

	idx, err := selectOption("Select a project:", options)
	if err != nil {
		return nil, err
	}

	selected := &projects[idx]
	fmt.Printf("✓ Selected: %s\n", selected.Name)
	return selected, nil
}

func searchAndSelectWorkItem(client *plane.Client, projectID string, minScore int) (*plane.WorkItem, error) {
	fmt.Println("\n🔍 Step 2: Find Work Item")

	for {
		searchTerm, err := input("Enter search term (or part of the title):")
		if err != nil {
			return nil, err
		}

		if searchTerm == "" {
			fmt.Println("❌ Please enter a search term.")
			continue
		}

		fmt.Println("\nSearching...")

		// Fetch all work items
		workItems, err := fetchAllWorkItemsForProject(client, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch work items: %w", err)
		}

		if len(workItems) == 0 {
			return nil, fmt.Errorf("no work items found in this project")
		}

		// Extract titles for fuzzy matching
		titles := make([]string, len(workItems))
		for i, item := range workItems {
			titles[i] = item.Name
		}

		// Find fuzzy matches
		matcher := fuzzy.NewMatcher(minScore)
		matches := matcher.FindMatches(searchTerm, titles)

		// If no fuzzy matches, try substring matching as fallback
		if len(matches) == 0 {
			searchLower := strings.ToLower(searchTerm)
			for i, title := range titles {
				if strings.Contains(strings.ToLower(title), searchLower) {
					matches = append(matches, fuzzy.MatchResult{
						Index: i,
						Score: 50, // Substring match gets 50%
					})
				}
			}
		}

		if len(matches) == 0 {
			fmt.Printf("❌ No work items found matching '%s'.\n", searchTerm)
			retry, err := confirm("Try again?")
			if err != nil {
				return nil, err
			}
			if retry {
				continue
			}
			return nil, fmt.Errorf("no matches found")
		}

		// Build options from matches
		fmt.Printf("\nFound %d match(es):\n", len(matches))
		var options []string
		for _, match := range matches {
			item := workItems[match.Index]
			options = append(options, fmt.Sprintf("[%d] %s (Score: %d%%)", item.SequenceID, truncate(item.Name, 40), match.Score))
		}

		// Get selection
		idx, err := selectOption("Select work item:", options)
		if err != nil {
			if err.Error() == "cancelled by user" {
				continue
			}
			return nil, err
		}

		selected := &workItems[matches[idx].Index]
		fmt.Printf("✓ Selected: %s (ID: %d)\n", selected.Name, selected.SequenceID)
		return selected, nil
	}
}

func fetchAllWorkItemsForProject(client *plane.Client, projectID string) ([]plane.WorkItem, error) {
	var allItems []plane.WorkItem
	offset := 0
	limit := 100

	for {
		options := map[string]string{
			"offset": fmt.Sprintf("%d", offset),
			"limit":  fmt.Sprintf("%d", limit),
		}

		response, err := client.GetWorkItems(projectID, options)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, response.Results...)

		if !response.NextPageResults || response.NextCursor == nil {
			break
		}
		break // TODO: Implement cursor pagination
	}

	return allItems, nil
}

func chooseUpdateFields(client *plane.Client, projectID string) (*plane.WorkItemUpdate, error) {
	fmt.Println("\n✏️  Step 3: What would you like to update?")

	options := []string{
		"Description",
		"Title",
		"State",
		"Priority",
		"Assignees",
		"Estimate Points",
		"Module",
		"Multiple fields",
		"Cancel",
	}

	idx, err := selectOption("Select an option:", options)
	if err != nil {
		return nil, err
	}

	update := &plane.WorkItemUpdate{}

	switch idx {
	case 0:
		// Description - choose between file or direct text
		desc, err := selectDescriptionSource()
		if err != nil {
			return nil, err
		}
		update.DescriptionHTML = desc

	case 1:
		// Title
		title, err := input("Enter new title:")
		if err != nil {
			return nil, err
		}
		update.Name = title

	case 2:
		// State
		state, err := selectState()
		if err != nil {
			return nil, err
		}
		update.State = state

	case 3:
		// Priority
		priority, err := selectPriority()
		if err != nil {
			return nil, err
		}
		update.Priority = priority

	case 4:
		// Assignees
		assignees, err := selectAssignees(client, projectID)
		if err != nil {
			return nil, err
		}
		update.Assignees = assignees

	case 5:
		// Estimate Points
		estimate, err := selectEstimate()
		if err != nil {
			return nil, err
		}
		update.EstimatePoint = estimate

	case 6:
		// Module
		module, err := selectModule(client, projectID)
		if err != nil {
			return nil, err
		}
		update.Module = module

	case 7:
		// Multiple fields
		return chooseMultipleFields(client, projectID)

	case 8:
		// Cancel
		return nil, nil
	}

	return update, nil
}

func selectDescriptionSource() (string, error) {
	fmt.Println("\n📝 Update Description")

	options := []string{
		"Load from file (markdown or text file)",
		"Enter text directly",
		"Cancel",
	}

	for {
		idx, err := selectOption("How would you like to enter the description?", options)
		if err != nil {
			return "", err
		}

		switch idx {
		case 0:
			return selectDescriptionFile()
		case 1:
			return enterDescriptionDirectly()
		case 2:
			return "", fmt.Errorf("description update cancelled")
		}
	}
}

func enterDescriptionDirectly() (string, error) {
	fmt.Println("\n✏️  Enter Description")

	var description string
	prompt := &survey.Multiline{
		Message: "Enter your description (supports multiple lines):",
	}
	err := survey.AskOne(prompt, &description)
	if err != nil {
		if err.Error() == "interrupt" {
			return "", fmt.Errorf("description entry cancelled")
		}
		return "", err
	}

	description = strings.TrimSpace(description)

	if description == "" {
		return "", fmt.Errorf("no description entered")
	}

	fmt.Printf("\n✓ Description entered: %d characters\n", len(description))
	return description, nil
}

func selectDescriptionFile() (string, error) {
	fmt.Println("\n📝 Select Description File")

	// Check for markdown files in common locations
	searchDirs := []string{
		"work_items",
		"descriptions",
		"docs",
		".",
	}

	var mdFiles []string
	for _, dir := range searchDirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.md"))
		if err == nil {
			mdFiles = append(mdFiles, files...)
		}
	}

	if len(mdFiles) > 0 {
		fmt.Println("\nAvailable markdown files:")
		var options []string
		for _, file := range mdFiles {
			options = append(options, file)
		}
		options = append(options, "Enter custom path")

		idx, err := selectOption("Select a file:", options)
		if err != nil {
			return "", err
		}

		if idx < len(mdFiles) {
			content, err := os.ReadFile(mdFiles[idx])
			if err != nil {
				return "", fmt.Errorf("failed to read file: %w", err)
			}

			fmt.Printf("✓ Loaded %d characters from %s\n", len(content), mdFiles[idx])
			return string(content), nil
		}
		// Fall through to custom path if "Enter custom path" selected
	}

	// Custom path input
	for {
		path, err := input("Enter path to description file:")
		if err != nil {
			return "", err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("❌ Failed to read file: %v\n", err)
			retry, err := confirm("Try again?")
			if err != nil {
				return "", err
			}
			if retry {
				continue
			}
			return "", fmt.Errorf("file selection cancelled")
		}

		fmt.Printf("✓ Loaded %d characters from %s\n", len(content), path)
		return string(content), nil
	}
}

func selectState() (string, error) {
	fmt.Println("\n📊 Select State")

	options := []string{
		"Backlog",
		"Todo",
		"In Progress",
		"Done",
		"Cancelled",
	}

	idx, err := selectOption("Select state:", options)
	if err != nil {
		return "", err
	}

	return options[idx], nil
}

func selectPriority() (string, error) {
	fmt.Println("\n🎯 Select Priority")

	options := []string{
		"urgent",
		"high",
		"medium",
		"low",
	}

	labels := []string{
		"Urgent",
		"High",
		"Medium",
		"Low",
	}

	idx, err := selectOption("Select priority:", labels)
	if err != nil {
		return "", err
	}

	return options[idx], nil
}

func chooseMultipleFields(client *plane.Client, projectID string) (*plane.WorkItemUpdate, error) {
	update := &plane.WorkItemUpdate{}

	for {
		fmt.Println("\n✏️  Select fields to update:")

		options := []string{
			"Description (from file or enter text)",
			"Title",
			"State",
			"Priority",
			"Assignees",
			"Estimate Points",
			"Module",
			"Done - Finish selection",
		}

		idx, err := selectOption("Select a field to update:", options)
		if err != nil {
			return nil, err
		}

		switch idx {
		case 0:
			desc, err := selectDescriptionSource()
			if err != nil {
				continue
			}
			update.DescriptionHTML = desc
			fmt.Println("✓ Description added to update")

		case 1:
			title, err := input("Enter new title:")
			if err != nil {
				continue
			}
			update.Name = title
			fmt.Println("✓ Title added to update")

		case 2:
			state, err := selectState()
			if err != nil {
				continue
			}
			update.State = state
			fmt.Printf("✓ State set to: %s\n", state)

		case 3:
			priority, err := selectPriority()
			if err != nil {
				continue
			}
			update.Priority = priority
			fmt.Printf("✓ Priority set to: %s\n", priority)

		case 4:
			assignees, err := selectAssignees(client, projectID)
			if err != nil {
				continue
			}
			update.Assignees = assignees
			fmt.Printf("✓ Assignees set: %v\n", assignees)

		case 5:
			estimate, err := selectEstimate()
			if err != nil {
				continue
			}
			update.EstimatePoint = estimate
			fmt.Printf("✓ Estimate set to: %.1f\n", estimate)

		case 6:
			module, err := selectModule(client, projectID)
			if err != nil {
				continue
			}
			update.Module = module
			fmt.Printf("✓ Module set to: %s\n", module)

		case 7:
			return update, nil
		}
	}
}

func selectAssignees(client *plane.Client, projectID string) ([]string, error) {
	fmt.Println("\n👥 Select Assignees")

	// Try to get project members first, fall back to workspace members
	members, err := client.GetProjectMembers(projectID)
	if err != nil || len(members) == 0 {
		members, err = client.GetWorkspaceMembers()
		if err != nil {
			return nil, fmt.Errorf("failed to get members: %w", err)
		}
	}

	if len(members) == 0 {
		return nil, fmt.Errorf("no members found")
	}

	// Build options
	var options []string
	for _, m := range members {
		name := m.GetDisplayName()
		options = append(options, fmt.Sprintf("%s (%s)", name, m.Email))
	}

	indices, err := selectMultiOption("Select assignees (use arrow keys and space to select, 'clear' to remove all):", options)
	if err != nil {
		if err.Error() == "cancelled by user" {
			return nil, err
		}
		// Check if user wants to clear
		return []string{}, nil
	}

	var selectedIDs []string
	for _, idx := range indices {
		selectedIDs = append(selectedIDs, members[idx].ID)
	}

	if len(selectedIDs) == 0 {
		return []string{}, nil
	}

	fmt.Printf("✓ Selected %d assignee(s)\n", len(selectedIDs))
	return selectedIDs, nil
}

func selectEstimate() (float64, error) {
	fmt.Println("\n📊 Enter Estimate Points")
	fmt.Println("Enter a number (e.g., 1, 2, 3, 5, 8, 13) or 0 to clear:")

	for {
		result, err := input("Estimate:")
		if err != nil {
			return 0, err
		}

		if result == "" || result == "0" {
			return 0, nil
		}

		estimate, err := strconv.ParseFloat(result, 64)
		if err != nil {
			fmt.Println("❌ Please enter a valid number.")
			continue
		}

		if estimate < 0 {
			fmt.Println("❌ Estimate cannot be negative.")
			continue
		}

		return estimate, nil
	}
}

func selectModule(client *plane.Client, projectID string) (string, error) {
	fmt.Println("\n📦 Select Module")

	modules, err := client.GetProjectModules(projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get modules: %w", err)
	}

	if len(modules) == 0 {
		return "", fmt.Errorf("no modules found in this project")
	}

	// Build options
	var options []string
	for _, m := range modules {
		options = append(options, m.Name)
	}
	options = append(options, "Clear module (remove from work item)")

	idx, err := selectOption("Select module:", options)
	if err != nil {
		return "", err
	}

	if idx == len(modules) {
		return "", nil
	}

	return modules[idx].ID, nil
}

func printUpdatePreview(update *plane.WorkItemUpdate) {
	if update.Name != "" {
		fmt.Printf("   → Title: %s\n", update.Name)
	}
	if update.DescriptionHTML != "" {
		fmt.Printf("   → Description: %d characters\n", len(update.DescriptionHTML))
	}
	if update.State != "" {
		fmt.Printf("   → State: %s\n", update.State)
	}
	if update.Priority != "" {
		fmt.Printf("   → Priority: %s\n", update.Priority)
	}
	if len(update.Assignees) > 0 {
		fmt.Printf("   → Assignees: %d selected\n", len(update.Assignees))
	}
	if update.EstimatePoint > 0 {
		fmt.Printf("   → Estimate: %.1f points\n", update.EstimatePoint)
	}
	if update.Module != "" {
		fmt.Printf("   → Module: %s\n", update.Module)
	}
}
