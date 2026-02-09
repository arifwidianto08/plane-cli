package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"plane-cli/internal/config"
	"plane-cli/internal/plane"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List work items",
	Long: `List work items in a project with optional filtering.

Examples:
  # List all work items in a project
  plane-cli list --project my-project

  # Filter by state
  plane-cli list --project my-project --state "In Progress"

  # Filter by priority
  plane-cli list --project my-project --priority high

  # Limit results
  plane-cli list --project my-project --limit 20`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Required flags
	listCmd.Flags().StringP("project", "p", "", "Project identifier (required)")
	listCmd.MarkFlagRequired("project")

	// Filter flags
	listCmd.Flags().String("state", "", "Filter by state")
	listCmd.Flags().String("priority", "", "Filter by priority (urgent, high, medium, low)")
	listCmd.Flags().StringSlice("labels", nil, "Filter by label IDs")
	listCmd.Flags().String("assignee", "", "Filter by assignee ID")

	// Pagination
	listCmd.Flags().Int("limit", 50, "Maximum number of results")
	listCmd.Flags().Int("offset", 0, "Offset for pagination")

	// Display options
	listCmd.Flags().Bool("show-description", false, "Show descriptions (may be truncated)")
}

func runList(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse flags
	project, _ := cmd.Flags().GetString("project")
	state, _ := cmd.Flags().GetString("state")
	priorityStr, _ := cmd.Flags().GetString("priority")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	showDescription, _ := cmd.Flags().GetBool("show-description")
	workspace, _ := cmd.Flags().GetString("workspace")

	// Get workspace - priority: flag > env > extract from URL
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

	// Build query options
	options := map[string]string{
		"limit":  fmt.Sprintf("%d", limit),
		"offset": fmt.Sprintf("%d", offset),
	}

	if state != "" {
		options["state"] = state
	}

	if priorityStr != "" {
		priority := plane.ParsePriority(priorityStr)
		options["priority"] = fmt.Sprintf("%d", priority)
	}

	// Note: Labels and assignee filtering may need custom handling
	// depending on Plane API capabilities

	// Fetch work items
	fmt.Printf("Fetching work items from project '%s'...\n\n", project)
	response, err := client.GetWorkItems(project, options)
	if err != nil {
		return fmt.Errorf("failed to fetch work items: %w", err)
	}

	if len(response.Results) == 0 {
		fmt.Println("No work items found.")
		return nil
	}

	// Display results
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header
	if showDescription {
		fmt.Fprintln(w, "ID\tTITLE\tSTATE\tPRIORITY\tASSIGNEES\tDESCRIPTION")
	} else {
		fmt.Fprintln(w, "ID\tTITLE\tSTATE\tPRIORITY\tASSIGNEES")
	}

	// Rows
	for _, item := range response.Results {
		id := fmt.Sprintf("%s-%d", project, item.SequenceID)
		title := truncate(item.Name, 40)
		state := item.State
		priority := item.Priority
		assignees := fmt.Sprintf("%d", len(item.Assignees))

		if showDescription {
			desc := ""
			if item.Description != "" {
				desc = truncate(stripHTML(item.Description), 50)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", id, title, state, priority, assignees, desc)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, title, state, priority, assignees)
		}
	}

	w.Flush()

	// Show pagination info
	fmt.Printf("\nShowing %d of %d work items\n", len(response.Results), response.TotalCount)
	if response.NextPageResults && response.NextCursor != nil {
		fmt.Printf("More results available. Use cursor-based pagination.\n")
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func stripHTML(s string) string {
	// Simple HTML tag removal
	result := ""
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result += string(r)
		}
	}
	return result
}
