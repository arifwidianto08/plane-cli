package commands

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"plane-cli/internal/config"
	"plane-cli/internal/fuzzy"
	"plane-cli/internal/plane"
	"plane-cli/internal/templates"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update existing work items",
	Long: `Update work items by ID or fuzzy title matching.

Examples:
  # Update by work item ID
  plane-cli update --id PROJ-123 --state "In Progress"

  # Fuzzy search and update
  plane-cli update --title-fuzzy "admin tenant" --template feature

  # Interactive mode
  plane-cli update --title-fuzzy "api" --interactive

  # Bulk update with auto-apply
  plane-cli update --title-fuzzy "bug" --template bug --auto`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Identification flags (one required)
	updateCmd.Flags().String("id", "", "Work item ID (e.g., PROJ-123)")
	updateCmd.Flags().String("title-fuzzy", "", "Fuzzy search by title")
	updateCmd.Flags().String("project", "", "Project identifier (required with title-fuzzy)")

	// Update flags
	updateCmd.Flags().String("title", "", "New title")
	updateCmd.Flags().String("description", "", "New description")
	updateCmd.Flags().String("description-file", "", "Read description from file")
	updateCmd.Flags().String("template", "", "Template name for description")
	updateCmd.Flags().StringToString("vars", nil, "Template variables")
	updateCmd.Flags().String("state", "", "New state")
	updateCmd.Flags().String("priority", "", "New priority (urgent, high, medium, low)")
	updateCmd.Flags().StringSlice("assignees", nil, "Assignee user IDs")
	updateCmd.Flags().StringSlice("labels", nil, "Label IDs")
	updateCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	updateCmd.Flags().String("target-date", "", "Target date (YYYY-MM-DD)")
	updateCmd.Flags().Float64("estimate", 0, "Estimate points")
	updateCmd.Flags().String("module", "", "Module ID")
	updateCmd.Flags().String("cycle", "", "Cycle ID")
	updateCmd.Flags().String("parent", "", "Parent work item ID")

	// Behavior flags
	updateCmd.Flags().Bool("interactive", false, "Interactive mode for selecting matches")
	updateCmd.Flags().Bool("auto", false, "Auto-apply to all matches")
	updateCmd.Flags().Bool("dry-run", false, "Preview changes without applying")
	updateCmd.Flags().Int("min-score", 60, "Minimum fuzzy match score (0-100)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n\n💡 To configure the CLI, run: plane-cli configure", err)
	}

	// Parse flags
	id, _ := cmd.Flags().GetString("id")
	titleFuzzy, _ := cmd.Flags().GetString("title-fuzzy")
	project, _ := cmd.Flags().GetString("project")
	newTitle, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	descriptionFile, _ := cmd.Flags().GetString("description-file")
	templateName, _ := cmd.Flags().GetString("template")

	// Read description from file if specified
	if descriptionFile != "" {
		content, err := os.ReadFile(descriptionFile)
		if err != nil {
			return fmt.Errorf("failed to read description file: %w", err)
		}
		description = string(content)
	}
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
	interactive, _ := cmd.Flags().GetBool("interactive")
	auto, _ := cmd.Flags().GetBool("auto")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	minScore, _ := cmd.Flags().GetInt("min-score")
	workspace, _ := cmd.Flags().GetString("workspace")

	// Validate input
	if id == "" && titleFuzzy == "" {
		return fmt.Errorf("either --id or --title-fuzzy is required")
	}
	if titleFuzzy != "" && project == "" {
		return fmt.Errorf("--project is required when using --title-fuzzy")
	}

	// Get workspace - priority: flag > env > extract from URL
	if workspace == "" {
		if cfg.PlaneWorkspace != "" {
			workspace = cfg.PlaneWorkspace
		} else {
			workspace = extractWorkspaceFromURL(cfg.PlaneBaseURL)
		}
	}

	// Initialize template manager
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

	// Build update payload
	update := &plane.WorkItemUpdate{}
	if newTitle != "" {
		update.Name = newTitle
	}
	if description != "" {
		// Send description as description_html
		update.DescriptionHTML = description
	}
	if state != "" {
		update.State = state
	}
	if priorityStr != "" {
		update.Priority = priorityStr
	}
	if len(assignees) > 0 {
		update.Assignees = assignees
	}
	if len(labels) > 0 {
		update.Labels = labels
	}
	if startDate != "" {
		update.StartDate = startDate
	}
	if targetDate != "" {
		update.TargetDate = targetDate
	}
	if estimate > 0 {
		update.EstimatePoint = estimate
	}
	if module != "" {
		update.Module = module
	}
	if cycle != "" {
		update.Cycle = cycle
	}
	if parent != "" {
		update.Parent = parent
	}

	// Execute update based on mode
	if id != "" {
		// Direct ID update
		return updateByID(client, project, id, update, dryRun)
	}

	// Fuzzy title search
	return updateByFuzzyTitle(client, project, titleFuzzy, update, minScore, interactive, auto, dryRun)
}

func updateByID(client *plane.Client, project, id string, update *plane.WorkItemUpdate, dryRun bool) error {
	// Get current work item
	workItem, err := client.GetWorkItem(project, id)
	if err != nil {
		return fmt.Errorf("failed to get work item: %w", err)
	}

	if dryRun {
		fmt.Printf("DRY RUN - Would update work item %s-\n", project, id)
		fmt.Printf("  Title: %s\n", workItem.Name)
		printUpdateDetails(update)
		return nil
	}

	// Store description length before sending
	sentDescLen := len(update.DescriptionHTML)

	// Apply update
	updated, err := client.UpdateWorkItem(project, id, update)
	if err != nil {
		return fmt.Errorf("failed to update work item: %w", err)
	}

	fmt.Printf("✓ Updated work item: %s-%d\n", project, updated.SequenceID)
	fmt.Printf("  Title: %s\n", updated.Name)
	fmt.Printf("  Desc sent: %d chars | Desc received: %d chars\n", sentDescLen, len(updated.DescriptionHTML))
	return nil
}

func updateByFuzzyTitle(client *plane.Client, project, pattern string, update *plane.WorkItemUpdate, minScore int, interactive, auto, dryRun bool) error {
	// Fetch all work items
	fmt.Printf("Fetching work items from project '%s'...\n", project)
	workItems, err := fetchAllWorkItems(client, project)
	if err != nil {
		return fmt.Errorf("failed to fetch work items: %w", err)
	}

	// Extract titles for fuzzy matching
	titles := make([]string, len(workItems))
	for i, item := range workItems {
		titles[i] = item.Name
	}

	// Find fuzzy matches
	matcher := fuzzy.NewMatcher(minScore)
	matches := matcher.FindMatches(pattern, titles)

	if len(matches) == 0 {
		fmt.Println("No matching work items found.")
		return nil
	}

	// Filter matches
	var matchedItems []*plane.WorkItem
	for _, match := range matches {
		matchedItems = append(matchedItems, &workItems[match.Index])
	}

	// Handle different modes
	if dryRun {
		printDryRun(matchedItems, update, matcher)
		return nil
	}

	if interactive {
		return updateInteractive(client, project, matchedItems, update)
	}

	if auto {
		return updateAll(client, project, matchedItems, update)
	}

	// Default: show matches and ask
	fmt.Printf("\nFound %d matching work items:\n\n", len(matchedItems))
	for i, item := range matchedItems {
		fmt.Printf("  %d. [%s-%d] %s\n", i+1, project, item.SequenceID, item.Name)
	}
	fmt.Printf("\nUpdate all? (y/n/list): ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	switch response {
	case "y", "yes":
		return updateAll(client, project, matchedItems, update)
	case "list", "l":
		return updateInteractive(client, project, matchedItems, update)
	default:
		fmt.Println("Update cancelled.")
		return nil
	}
}

func fetchAllWorkItems(client *plane.Client, project string) ([]plane.WorkItem, error) {
	var allItems []plane.WorkItem
	offset := 0
	limit := 100

	for {
		options := map[string]string{
			"offset": fmt.Sprintf("%d", offset),
			"limit":  fmt.Sprintf("%d", limit),
		}

		response, err := client.GetWorkItems(project, options)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, response.Results...)

		// Check if there are more results using cursor-based pagination
		if !response.NextPageResults || response.NextCursor == nil {
			break
		}
		// For now, just break to avoid infinite loop with cursor pagination
		// TODO: Implement cursor-based pagination properly
		break
	}

	return allItems, nil
}

func updateInteractive(client *plane.Client, project string, items []*plane.WorkItem, update *plane.WorkItemUpdate) error {
	fmt.Println("\nSelect items to update (comma-separated numbers, 'all', or 'cancel'):")
	for i, item := range items {
		fmt.Printf("  %d. [%s-%d] %s\n", i+1, project, item.SequenceID, item.Name)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nSelection: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "cancel" || input == "c" {
		fmt.Println("Update cancelled.")
		return nil
	}

	if input == "all" || input == "a" {
		return updateAll(client, project, items, update)
	}

	// Parse selection
	selected := parseSelection(input, len(items))
	if len(selected) == 0 {
		fmt.Println("No items selected.")
		return nil
	}

	var selectedItems []*plane.WorkItem
	for _, idx := range selected {
		selectedItems = append(selectedItems, items[idx-1])
	}

	return updateAll(client, project, selectedItems, update)
}

func updateAll(client *plane.Client, project string, items []*plane.WorkItem, update *plane.WorkItemUpdate) error {
	fmt.Printf("\nUpdating %d work items...\n", len(items))

	successCount := 0
	for _, item := range items {
		_, err := client.UpdateWorkItem(project, item.ID, update)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to update %s-%d: %v\n", project, item.SequenceID, err)
			continue
		}
		fmt.Printf("✓ Updated %s-%d: %s\n", project, item.SequenceID, item.Name)
		successCount++
	}

	fmt.Printf("\nUpdated %d/%d work items.\n", successCount, len(items))
	return nil
}

func parseSelection(input string, max int) []int {
	var selected []int
	parts := strings.Split(input, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		num, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		if num > 0 && num <= max {
			selected = append(selected, num)
		}
	}

	return selected
}

func printDryRun(items []*plane.WorkItem, update *plane.WorkItemUpdate, matcher *fuzzy.Matcher) {
	fmt.Println("DRY RUN - No changes will be made\n")
	for _, item := range items {
		fmt.Printf("  [%s] %s\n", item.ID, item.Name)
		printUpdateDetails(update)
		fmt.Println()
	}
	fmt.Println("Run without --dry-run to apply changes.")
}

func printUpdateDetails(update *plane.WorkItemUpdate) {
	if update.Name != "" {
		fmt.Printf("  → Title: %s\n", update.Name)
	}
	if update.DescriptionHTML != "" {
		fmt.Printf("  → Description: [updated - %d chars]\n", len(update.DescriptionHTML))
	}
	if update.State != "" {
		fmt.Printf("  → State: %s\n", update.State)
	}
	if update.Priority != "" {
		fmt.Printf("  → Priority: %s\n", update.Priority)
	}
	if len(update.Assignees) > 0 {
		fmt.Printf("  → Assignees: %v\n", update.Assignees)
	}
	if len(update.Labels) > 0 {
		fmt.Printf("  → Labels: %v\n", update.Labels)
	}
}

// markdownToHTML converts basic markdown to HTML
func markdownToHTML(markdown string) string {
	// For Plane, we can wrap markdown in a div and it will render properly
	// Plane's editor handles markdown conversion internally
	html := strings.TrimSpace(markdown)

	// Wrap the content in a div container
	return "<div>" + html + "</div>"
}
