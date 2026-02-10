package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"plane-cli/internal/config"
	"plane-cli/internal/plane"
)

var bulkCreateCmd = &cobra.Command{
	Use:   "bulk-create",
	Short: "Create multiple work items at once with shared specifications",
	Long: `Create multiple work items simultaneously with the same attributes.

Perfect for creating batches of related work items like:
- [BE] Purchase Order
- [BE] Sales Order
- [BE] Inventory Management

All work items will share the same:
- Assignees
- Estimate points
- Labels
- Module
- State
- Priority

Examples:
  # Interactive bulk create
  plane-cli bulk-create --project c20fcc54-c675-47c4-85db-a4acdde3c9e1

  # Bulk create with titles from command line
  plane-cli bulk-create \
    --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 \
    --titles "[BE] Purchase Order,[BE] Sales Order,[BE] Inventory" \
    --assignees user-id-1 \
    --estimate 5 \
    --state "Backlog"

  # Create from file (one title per line)
  plane-cli bulk-create \
    --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 \
    --titles-file work-items.txt \
    --module module-id \
    --labels label-1,label-2`,
	RunE: runBulkCreate,
}

func init() {
	rootCmd.AddCommand(bulkCreateCmd)

	// Required flags
	bulkCreateCmd.Flags().String("project", "", "Project identifier (required)")
	bulkCreateCmd.MarkFlagRequired("project")

	// Titles input
	bulkCreateCmd.Flags().StringSlice("titles", nil, "Work item titles (comma-separated)")
	bulkCreateCmd.Flags().String("titles-file", "", "File containing titles (one per line)")

	// Common attributes
	bulkCreateCmd.Flags().StringSlice("assignees", nil, "Assignee user IDs (comma-separated)")
	bulkCreateCmd.Flags().Float64("estimate", 0, "Estimate points for all work items")
	bulkCreateCmd.Flags().StringSlice("labels", nil, "Label IDs (comma-separated)")
	bulkCreateCmd.Flags().String("module", "", "Module ID")
	bulkCreateCmd.Flags().String("state", "Backlog", "Initial state (default: Backlog)")
	bulkCreateCmd.Flags().String("priority", "medium", "Priority: urgent, high, medium, low (default: medium)")
	bulkCreateCmd.Flags().String("description", "", "Description for all work items")
	bulkCreateCmd.Flags().String("description-file", "", "Read description from file")

	// Behavior flags
	bulkCreateCmd.Flags().Bool("dry-run", false, "Preview what would be created without actually creating")
	bulkCreateCmd.Flags().Bool("interactive", false, "Force interactive mode")
}

func runBulkCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n\n💡 To configure the CLI, run: plane-cli configure", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	titlesFlag, _ := cmd.Flags().GetStringSlice("titles")
	titlesFile, _ := cmd.Flags().GetString("titles-file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	forceInteractive, _ := cmd.Flags().GetBool("interactive")

	// Get common attributes
	assignees, _ := cmd.Flags().GetStringSlice("assignees")
	estimate, _ := cmd.Flags().GetFloat64("estimate")
	labels, _ := cmd.Flags().GetStringSlice("labels")
	moduleID, _ := cmd.Flags().GetString("module")
	state, _ := cmd.Flags().GetString("state")
	priorityStr, _ := cmd.Flags().GetString("priority")
	description, _ := cmd.Flags().GetString("description")
	descriptionFile, _ := cmd.Flags().GetString("description-file")

	// Read description from file if specified
	if descriptionFile != "" {
		content, err := readFileContent(descriptionFile)
		if err != nil {
			return fmt.Errorf("failed to read description file: %w", err)
		}
		description = content
	}

	workspace := cfg.PlaneWorkspace
	if workspace == "" {
		workspace = extractWorkspaceFromURL(cfg.PlaneBaseURL)
	}

	client, err := plane.NewClient(cfg.PlaneBaseURL, cfg.PlaneAPIToken)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	client.SetWorkspace(workspace)

	// Get project info
	project, err := client.GetProject(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Collect titles
	var titles []string

	if len(titlesFlag) > 0 && !forceInteractive {
		// Use titles from command line
		titles = titlesFlag
	} else if titlesFile != "" && !forceInteractive {
		// Read titles from file
		content, err := readFileContent(titlesFile)
		if err != nil {
			return fmt.Errorf("failed to read titles file: %w", err)
		}
		// Split by lines and filter empty ones
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				titles = append(titles, line)
			}
		}
	} else {
		// Interactive mode - collect titles
		titles, err = collectTitlesInteractive()
		if err != nil {
			return err
		}
	}

	if len(titles) == 0 {
		return fmt.Errorf("no titles provided")
	}

	// If in interactive mode or missing attributes, prompt for them
	if forceInteractive || (len(assignees) == 0 && estimate == 0 && len(labels) == 0 && moduleID == "" && description == "") {
		// Get common attributes interactively
		attrs, err := selectCommonAttributes(client, projectID)
		if err != nil {
			return err
		}

		// Merge with command line flags (CLI flags take precedence)
		if len(assignees) == 0 && len(attrs.Assignees) > 0 {
			assignees = attrs.Assignees
		}
		if estimate == 0 && attrs.EstimatePoint > 0 {
			estimate = attrs.EstimatePoint
		}
		if len(labels) == 0 && len(attrs.Labels) > 0 {
			labels = attrs.Labels
		}
		if moduleID == "" && attrs.Module != "" {
			moduleID = attrs.Module
		}
		if description == "" && attrs.Description != "" {
			description = attrs.Description
		}
		if state == "Backlog" && attrs.State != "" {
			state = attrs.State
		}
		if priorityStr == "medium" && attrs.Priority != "" {
			priorityStr = attrs.Priority
		}
	}

	// Parse priority
	priority := plane.ParsePriority(priorityStr)

	// Preview
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("                    📋 BULK CREATE PREVIEW")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Project: %s (%s)\n", project.Name, project.Identifier)
	fmt.Printf("Number of work items to create: %d\n\n", len(titles))

	fmt.Println("Titles:")
	for i, title := range titles {
		fmt.Printf("  %d. %s\n", i+1, title)
	}

	fmt.Println("\nCommon attributes:")
	if len(assignees) > 0 {
		fmt.Printf("  • Assignees: %d selected\n", len(assignees))
	}
	if estimate > 0 {
		fmt.Printf("  • Estimate: %.1f points\n", estimate)
	}
	if len(labels) > 0 {
		fmt.Printf("  • Labels: %d selected\n", len(labels))
	}
	if moduleID != "" {
		fmt.Printf("  • Module: %s\n", moduleID)
	}
	if state != "" {
		fmt.Printf("  • State: %s\n", state)
	}
	fmt.Printf("  • Priority: %s\n", plane.GetPriorityName(priority))
	if description != "" {
		fmt.Printf("  • Description: %d characters\n", len(description))
	}

	fmt.Println(strings.Repeat("=", 70))

	if dryRun {
		fmt.Println("\n📝 Dry run mode - no work items created.")
		return nil
	}

	// Confirm
	confirmed, err := confirm("\nCreate these work items?")
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("\n❌ Creation cancelled.")
		return nil
	}

	// Create work items
	fmt.Printf("\n🔄 Creating %d work items...\n\n", len(titles))

	successCount := 0
	failCount := 0
	var createdItems []plane.WorkItem

	for _, title := range titles {
		create := &plane.WorkItemCreate{
			Name:        title,
			Description: description,
			Priority:    plane.ParsePriorityString(priorityStr),
			Assignees:   assignees,
			Labels:      labels,
			Module:      moduleID,
		}

		// Convert state name to UUID if provided
		if state != "" {
			stateID, err := client.GetStateByName(projectID, state)
			if err == nil {
				create.State = stateID
			} else {
				fmt.Printf("  ⚠️  Warning: Could not convert state '%s': %v\n", state, err)
			}
		}

		// Convert estimate to UUID if provided
		if estimate > 0 {
			estimateID, err := client.GetEstimatePointByValue(projectID, estimate)
			if err == nil {
				create.EstimatePoint = estimateID
				fmt.Printf("  ℹ️  Converted estimate %.0f to UUID: %s\n", estimate, estimateID)
			} else {
				fmt.Printf("  ⚠️  Warning: Could not find estimate UUID for value %.0f: %v\n", estimate, err)
			}
		}

		// Debug: Print what we're sending
		fmt.Printf("  ℹ️  Creating with Module: %s, Estimate: %s\n", create.Module, create.EstimatePoint)

		workItem, err := client.CreateWorkItem(projectID, create)
		if err != nil {
			fmt.Printf("  ❌ Failed: %s - %v\n", title, err)
			failCount++
		} else {
			fmt.Printf("  ✅ Created: [%d] %s\n", workItem.SequenceID, title)

			// If module was set but didn't apply during creation, update it separately
			if moduleID != "" && workItem.ModuleID == "" {
				update := &plane.WorkItemUpdate{
					Module: moduleID,
				}
				_, err := client.UpdateWorkItem(projectID, workItem.ID, update)
				if err != nil {
					fmt.Printf("  ⚠️  Warning: Created but couldn't set module: %v\n", err)
				} else {
					fmt.Printf("  ✅ Module updated for: [%d] %s\n", workItem.SequenceID, title)
				}
			}

			createdItems = append(createdItems, *workItem)
			successCount++
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Printf("✅ Completed: %d/%d work items created successfully\n", successCount, len(titles))
	if failCount > 0 {
		fmt.Printf("❌ Failed: %d work items\n", failCount)
	}

	// Show summary of created items
	if len(createdItems) > 0 {
		fmt.Println("\nCreated work items:")
		for _, item := range createdItems {
			fmt.Printf("  • %s-%d: %s\n", project.Identifier, item.SequenceID, item.Name)
		}
	}

	return nil
}

func collectTitlesInteractive() ([]string, error) {
	fmt.Println("\n📝 Enter Work Item Titles")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("Enter each title on a new line.")
	fmt.Println("Type ':done' on a new line when finished, or press Enter twice.")
	fmt.Println(strings.Repeat("-", 70))

	var titles []string
	emptyLineCount := 0

	for {
		result, err := input("Enter title (or :done):")
		if err != nil {
			return nil, err
		}

		result = strings.TrimSpace(result)

		if result == ":done" {
			break
		}

		if result == "" {
			emptyLineCount++
			if emptyLineCount >= 2 {
				break
			}
			continue
		}

		emptyLineCount = 0
		titles = append(titles, result)
	}

	if len(titles) == 0 {
		return nil, fmt.Errorf("no titles entered")
	}

	fmt.Printf("\n✓ Collected %d titles\n", len(titles))
	return titles, nil
}

type commonAttributes struct {
	Assignees     []string
	EstimatePoint float64
	Labels        []string
	Module        string
	State         string
	Priority      string
	Description   string
}

func selectCommonAttributes(client *plane.Client, projectID string) (*commonAttributes, error) {
	attrs := &commonAttributes{}

	fmt.Println("\n⚙️  Select Common Attributes")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("These attributes will be applied to ALL work items.")
	fmt.Println(strings.Repeat("-", 70))

	for {
		options := []string{
			"Assignees",
			"Estimate Points",
			"Labels",
			"Module",
			"State",
			"Priority",
			"Description",
			"Done - Continue to create",
		}

		idx, err := selectOption("What would you like to set?", options)
		if err != nil {
			return nil, err
		}

		switch idx {
		case 0: // Assignees
			members, err := client.GetProjectMembers(projectID)
			if err != nil {
				members, err = client.GetWorkspaceMembers()
				if err != nil {
					fmt.Printf("❌ Error fetching members: %v\n", err)
					continue
				}
			}

			if len(members) == 0 {
				fmt.Println("❌ No members found")
				continue
			}

			var options []string
			for _, m := range members {
				options = append(options, fmt.Sprintf("%s (%s)", m.GetDisplayName(), m.Email))
			}

			indices, err := selectMultiOption("Select assignees:", options)
			if err != nil {
				continue
			}

			for _, idx := range indices {
				attrs.Assignees = append(attrs.Assignees, members[idx].ID)
			}
			fmt.Printf("✓ Selected %d assignees\n", len(attrs.Assignees))

		case 1: // Estimate
			estimate, err := askFloat("Enter estimate points:")
			if err != nil {
				continue
			}
			if estimate > 0 {
				attrs.EstimatePoint = estimate
				fmt.Printf("✓ Estimate set to: %.1f\n", estimate)
			}

		case 2: // Labels
			labels, err := client.GetLabels(projectID)
			if err != nil {
				fmt.Printf("❌ Error fetching labels: %v\n", err)
				continue
			}

			if len(labels) == 0 {
				fmt.Println("❌ No labels found in this project")
				continue
			}

			var options []string
			for _, l := range labels {
				options = append(options, l.Name)
			}

			indices, err := selectMultiOption("Select labels:", options)
			if err != nil {
				continue
			}

			for _, idx := range indices {
				attrs.Labels = append(attrs.Labels, labels[idx].ID)
			}
			fmt.Printf("✓ Selected %d labels\n", len(attrs.Labels))

		case 3: // Module
			modules, err := client.GetModules(projectID)
			if err != nil {
				fmt.Printf("❌ Error fetching modules: %v\n", err)
				continue
			}

			if len(modules) == 0 {
				fmt.Println("❌ No modules found in this project")
				continue
			}

			options := []string{"No module"}
			for _, m := range modules {
				options = append(options, m.Name)
			}

			idx, err := selectOption("Select module:", options)
			if err != nil {
				continue
			}

			if idx > 0 {
				attrs.Module = modules[idx-1].ID
				fmt.Printf("✓ Module set to: %s\n", modules[idx-1].Name)
			}

		case 4: // State
			state, err := selectState()
			if err != nil {
				continue
			}
			attrs.State = state
			fmt.Printf("✓ State set to: %s\n", state)

		case 5: // Priority
			priority, err := selectPriority()
			if err != nil {
				continue
			}
			attrs.Priority = priority
			fmt.Printf("✓ Priority set to: %s\n", plane.GetPriorityName(plane.ParsePriority(priority)))

		case 6: // Description
			fmt.Println("\nEnter description source:")
			srcIdx, err := selectOption("Select source:", []string{
				"Enter text directly",
				"Load from file",
				"Skip description",
			})
			if err != nil {
				continue
			}

			switch srcIdx {
			case 0: // Direct text
				text, err := input("Enter description (supports markdown):")
				if err != nil {
					continue
				}
				attrs.Description = text
				fmt.Printf("✓ Description set (%d chars)\n", len(text))

			case 1: // File
				path, err := input("Enter file path:")
				if err != nil {
					continue
				}
				content, err := readFileContent(path)
				if err != nil {
					fmt.Printf("❌ Error reading file: %v\n", err)
					continue
				}
				attrs.Description = content
				fmt.Printf("✓ Description loaded from file (%d chars)\n", len(content))
			}

		case 7: // Done
			return attrs, nil
		}
	}
}
