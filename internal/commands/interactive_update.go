package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"plane-cli/internal/config"
	"plane-cli/internal/fuzzy"
	"plane-cli/internal/plane"

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

	reader := bufio.NewReader(os.Stdin)

	// Step 1: Select Project
	var project *plane.Project
	if projectID == "" {
		project, err = selectProjectInteractive(client, reader)
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
	workItem, err := searchAndSelectWorkItem(client, projectID, reader, minScore)
	if err != nil {
		return err
	}

	// Step 3: Choose what to update
	update, err := chooseUpdateFields(client, projectID, reader)
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

	fmt.Print("\nApply these changes? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
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

func selectProjectInteractive(client *plane.Client, reader *bufio.Reader) (*plane.Project, error) {
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
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("%-5s %-20s %s\n", "#", "IDENTIFIER", "NAME")
	fmt.Println(strings.Repeat("-", 70))

	for i, p := range projects {
		fmt.Printf("%-5d %-20s %s\n", i+1, p.Identifier, truncate(p.Name, 40))
	}

	fmt.Println(strings.Repeat("-", 70))

	for {
		fmt.Print("\nEnter project number: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(projects) {
			fmt.Println("❌ Invalid selection. Please try again.")
			continue
		}

		selected := &projects[num-1]
		fmt.Printf("✓ Selected: %s\n", selected.Name)
		return selected, nil
	}
}

func searchAndSelectWorkItem(client *plane.Client, projectID string, reader *bufio.Reader, minScore int) (*plane.WorkItem, error) {
	fmt.Println("\n🔍 Step 2: Find Work Item")
	fmt.Println(strings.Repeat("-", 70))

	for {
		fmt.Print("\nEnter search term (or part of the title): ")
		searchTerm, _ := reader.ReadString('\n')
		searchTerm = strings.TrimSpace(searchTerm)

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

		if len(matches) == 0 {
			fmt.Printf("❌ No work items found matching '%s'.\n", searchTerm)
			fmt.Print("Try again? (y/n): ")
			retry, _ := reader.ReadString('\n')
			retry = strings.TrimSpace(strings.ToLower(retry))
			if retry == "y" || retry == "yes" {
				continue
			}
			return nil, fmt.Errorf("no matches found")
		}

		// Show matches
		fmt.Printf("\nFound %d match(es):\n\n", len(matches))
		fmt.Printf("%-5s %-10s %-40s %s\n", "#", "ID", "TITLE", "SCORE")
		fmt.Println(strings.Repeat("-", 70))

		for i, match := range matches {
			item := workItems[match.Index]
			fmt.Printf("%-5d %-10d %-40s %d%%\n", i+1, item.SequenceID, truncate(item.Name, 38), match.Score)
		}

		fmt.Println(strings.Repeat("-", 70))

		// Get selection
		fmt.Print("\nSelect work item number: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(matches) {
			fmt.Println("❌ Invalid selection.")
			continue
		}

		selected := &workItems[matches[num-1].Index]
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

func chooseUpdateFields(client *plane.Client, projectID string, reader *bufio.Reader) (*plane.WorkItemUpdate, error) {
	fmt.Println("\n✏️  Step 3: What would you like to update?")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("1. Description")
	fmt.Println("2. Title")
	fmt.Println("3. State")
	fmt.Println("4. Priority")
	fmt.Println("5. Assignees")
	fmt.Println("6. Estimate Points")
	fmt.Println("7. Module")
	fmt.Println("8. Multiple fields")
	fmt.Println("9. Cancel")
	fmt.Println(strings.Repeat("-", 70))

	fmt.Print("\nEnter your choice (1-9): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	update := &plane.WorkItemUpdate{}

	switch choice {
	case "1":
		// Description - choose between file or direct text
		desc, err := selectDescriptionSource(reader)
		if err != nil {
			return nil, err
		}
		update.DescriptionHTML = desc

	case "2":
		// Title
		fmt.Print("\nEnter new title: ")
		title, _ := reader.ReadString('\n')
		update.Name = strings.TrimSpace(title)

	case "3":
		// State
		state, err := selectState(reader)
		if err != nil {
			return nil, err
		}
		update.State = state

	case "4":
		// Priority
		priority, err := selectPriority(reader)
		if err != nil {
			return nil, err
		}
		update.Priority = priority

	case "5":
		// Assignees
		assignees, err := selectAssignees(client, projectID, reader)
		if err != nil {
			return nil, err
		}
		update.Assignees = assignees

	case "6":
		// Estimate Points
		estimate, err := selectEstimate(reader)
		if err != nil {
			return nil, err
		}
		update.EstimatePoint = estimate

	case "7":
		// Module
		module, err := selectModule(client, projectID, reader)
		if err != nil {
			return nil, err
		}
		update.Module = module

	case "8":
		// Multiple fields
		return chooseMultipleFields(client, projectID, reader)

	case "9", "cancel", "c":
		return nil, nil

	default:
		fmt.Println("❌ Invalid choice.")
		return nil, nil
	}

	return update, nil
}

func selectDescriptionSource(reader *bufio.Reader) (string, error) {
	fmt.Println("\n📝 Update Description")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("1. Load from file (markdown or text file)")
	fmt.Println("2. Enter text directly")
	fmt.Println("3. Cancel")
	fmt.Println(strings.Repeat("-", 70))

	for {
		fmt.Print("\nEnter your choice (1-3): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			return selectDescriptionFile(reader)
		case "2":
			return enterDescriptionDirectly(reader)
		case "3", "cancel", "c":
			return "", fmt.Errorf("description update cancelled")
		default:
			fmt.Println("❌ Invalid choice. Please enter 1, 2, or 3.")
		}
	}
}

func enterDescriptionDirectly(reader *bufio.Reader) (string, error) {
	fmt.Println("\n✏️  Enter Description")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("Enter your description below (supports multiple lines).")
	fmt.Println("Type \":done\" on a new line to finish, or press Enter twice.")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println()

	var lines []string
	emptyLineCount := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		// Check for :done sentinel
		if strings.TrimSpace(line) == ":done" {
			break
		}

		// Check for double empty line (Enter twice)
		if strings.TrimSpace(line) == "" {
			emptyLineCount++
			if emptyLineCount >= 2 {
				break
			}
		} else {
			emptyLineCount = 0
		}

		lines = append(lines, line)
	}

	description := strings.Join(lines, "")
	description = strings.TrimSpace(description)

	if description == "" {
		return "", fmt.Errorf("no description entered")
	}

	fmt.Printf("\n✓ Description entered: %d characters\n", len(description))
	return description, nil
}

func selectDescriptionFile(reader *bufio.Reader) (string, error) {
	fmt.Println("\n📝 Select Description File")
	fmt.Println(strings.Repeat("-", 70))

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
		for i, file := range mdFiles {
			fmt.Printf("%d. %s\n", i+1, file)
		}
		fmt.Println("0. Enter custom path")
		fmt.Println(strings.Repeat("-", 70))

		for {
			fmt.Print("\nSelect file number (or 0 for custom path): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			num, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("❌ Invalid selection.")
				continue
			}

			if num == 0 {
				break // Go to custom path input
			}

			if num < 1 || num > len(mdFiles) {
				fmt.Println("❌ Invalid selection.")
				continue
			}

			content, err := os.ReadFile(mdFiles[num-1])
			if err != nil {
				return "", fmt.Errorf("failed to read file: %w", err)
			}

			fmt.Printf("✓ Loaded %d characters from %s\n", len(content), mdFiles[num-1])
			return string(content), nil
		}
	}

	// Custom path input
	for {
		fmt.Print("\nEnter path to description file: ")
		path, _ := reader.ReadString('\n')
		path = strings.TrimSpace(path)

		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("❌ Failed to read file: %v\n", err)
			fmt.Print("Try again? (y/n): ")
			retry, _ := reader.ReadString('\n')
			retry = strings.TrimSpace(strings.ToLower(retry))
			if retry == "y" || retry == "yes" {
				continue
			}
			return "", fmt.Errorf("file selection cancelled")
		}

		fmt.Printf("✓ Loaded %d characters from %s\n", len(content), path)
		return string(content), nil
	}
}

func selectState(reader *bufio.Reader) (string, error) {
	fmt.Println("\n📊 Select State")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("1. Backlog")
	fmt.Println("2. Todo")
	fmt.Println("3. In Progress")
	fmt.Println("4. Done")
	fmt.Println("5. Cancelled")
	fmt.Println(strings.Repeat("-", 70))

	for {
		fmt.Print("\nEnter state number (1-5): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			return "Backlog", nil
		case "2":
			return "Todo", nil
		case "3":
			return "In Progress", nil
		case "4":
			return "Done", nil
		case "5":
			return "Cancelled", nil
		default:
			fmt.Println("❌ Invalid selection.")
		}
	}
}

func selectPriority(reader *bufio.Reader) (string, error) {
	fmt.Println("\n🎯 Select Priority")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("1. Urgent")
	fmt.Println("2. High")
	fmt.Println("3. Medium")
	fmt.Println("4. Low")
	fmt.Println(strings.Repeat("-", 70))

	for {
		fmt.Print("\nEnter priority number (1-4): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			return "urgent", nil
		case "2":
			return "high", nil
		case "3":
			return "medium", nil
		case "4":
			return "low", nil
		default:
			fmt.Println("❌ Invalid selection.")
		}
	}
}

func chooseMultipleFields(client *plane.Client, projectID string, reader *bufio.Reader) (*plane.WorkItemUpdate, error) {
	update := &plane.WorkItemUpdate{}

	for {
		fmt.Println("\n✏️  Select fields to update (choose one at a time, 'done' when finished):")
		fmt.Println("1. Description (from file or enter text)")
		fmt.Println("2. Title")
		fmt.Println("3. State")
		fmt.Println("4. Priority")
		fmt.Println("5. Assignees")
		fmt.Println("6. Estimate Points")
		fmt.Println("7. Module")
		fmt.Println("8. Done")
		fmt.Println(strings.Repeat("-", 70))

		fmt.Print("\nChoice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			desc, err := selectDescriptionSource(reader)
			if err != nil {
				continue
			}
			update.DescriptionHTML = desc
			fmt.Println("✓ Description added to update")

		case "2":
			fmt.Print("\nEnter new title: ")
			title, _ := reader.ReadString('\n')
			update.Name = strings.TrimSpace(title)
			fmt.Println("✓ Title added to update")

		case "3":
			state, err := selectState(reader)
			if err != nil {
				continue
			}
			update.State = state
			fmt.Printf("✓ State set to: %s\n", state)

		case "4":
			priority, err := selectPriority(reader)
			if err != nil {
				continue
			}
			update.Priority = priority
			fmt.Printf("✓ Priority set to: %s\n", priority)

		case "5":
			assignees, err := selectAssignees(client, projectID, reader)
			if err != nil {
				continue
			}
			update.Assignees = assignees
			fmt.Printf("✓ Assignees set: %v\n", assignees)

		case "6":
			estimate, err := selectEstimate(reader)
			if err != nil {
				continue
			}
			update.EstimatePoint = estimate
			fmt.Printf("✓ Estimate set to: %.1f\n", estimate)

		case "7":
			module, err := selectModule(client, projectID, reader)
			if err != nil {
				continue
			}
			update.Module = module
			fmt.Printf("✓ Module set to: %s\n", module)

		case "8", "done":
			return update, nil

		default:
			fmt.Println("❌ Invalid choice.")
		}
	}
}

func selectAssignees(client *plane.Client, projectID string, reader *bufio.Reader) ([]string, error) {
	fmt.Println("\n👥 Select Assignees")
	fmt.Println(strings.Repeat("-", 70))

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

	// Display members
	fmt.Printf("\n%-5s %-30s %s\n", "#", "NAME", "EMAIL")
	fmt.Println(strings.Repeat("-", 70))
	for i, m := range members {
		name := m.GetDisplayName()
		if len(name) > 28 {
			name = name[:25] + "..."
		}
		email := m.Email
		if len(email) > 25 {
			email = email[:22] + "..."
		}
		fmt.Printf("%-5d %-30s %s\n", i+1, name, email)
	}
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("Enter numbers separated by commas (e.g., 1,3,5) or 'clear' to remove all:")

	fmt.Print("\nSelection: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "clear" || input == "none" {
		return []string{}, nil
	}

	var selectedIDs []string
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		num, err := strconv.Atoi(part)
		if err != nil || num < 1 || num > len(members) {
			fmt.Printf("❌ Invalid selection: %s\n", part)
			continue
		}
		selectedIDs = append(selectedIDs, members[num-1].ID)
	}

	if len(selectedIDs) == 0 {
		return nil, fmt.Errorf("no valid selections")
	}

	fmt.Printf("✓ Selected %d assignee(s)\n", len(selectedIDs))
	return selectedIDs, nil
}

func selectEstimate(reader *bufio.Reader) (float64, error) {
	fmt.Println("\n📊 Enter Estimate Points")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("Enter a number (e.g., 1, 2, 3, 5, 8, 13) or 0 to clear:")

	for {
		fmt.Print("\nEstimate: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" || input == "0" {
			return 0, nil
		}

		estimate, err := strconv.ParseFloat(input, 64)
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

func selectModule(client *plane.Client, projectID string, reader *bufio.Reader) (string, error) {
	fmt.Println("\n📦 Select Module")
	fmt.Println(strings.Repeat("-", 70))

	modules, err := client.GetProjectModules(projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get modules: %w", err)
	}

	if len(modules) == 0 {
		return "", fmt.Errorf("no modules found in this project")
	}

	// Display modules
	fmt.Printf("\n%-5s %s\n", "#", "NAME")
	fmt.Println(strings.Repeat("-", 50))
	for i, m := range modules {
		fmt.Printf("%-5d %s\n", i+1, m.Name)
	}
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("0. Clear module (remove from work item)")

	for {
		fmt.Print("\nEnter module number: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		num, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("❌ Please enter a valid number.")
			continue
		}

		if num == 0 {
			return "", nil
		}

		if num < 1 || num > len(modules) {
			fmt.Println("❌ Invalid selection.")
			continue
		}

		return modules[num-1].ID, nil
	}
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
