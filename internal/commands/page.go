package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"plane-cli/internal/config"
	"plane-cli/internal/plane"
)

var pageCmd = &cobra.Command{
	Use:   "page",
	Short: "Manage project pages",
	Long: `List, create, update, and delete pages in your Plane projects.

Examples:
  # List all pages in a project
  plane-cli page list --project c20fcc54-c675-47c4-85db-a4acdde3c9e1

  # Create a new page
  plane-cli page create --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 --name "Documentation" --description-file docs.md

  # Update a page
  plane-cli page update --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 --id <page-id> --name "API Documentation"

  # Delete a page
  plane-cli page delete --project c20fcc54-c675-47c4-85db-a4acdde3c9e1 --id <page-id>

  # Interactive page management
  plane-cli page interactive`,
}

var pageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pages in a project",
	RunE:  runPageList,
}

var pageCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new page",
	RunE:  runPageCreate,
}

var pageUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing page",
	RunE:  runPageUpdate,
}

var pageDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a page",
	RunE:  runPageDelete,
}

var pageInteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive page management",
	Long:  `Interactive workflow for managing pages - select project, then create, update, or delete pages.`,
	RunE:  runPageInteractive,
}

func init() {
	rootCmd.AddCommand(pageCmd)
	pageCmd.AddCommand(pageListCmd)
	pageCmd.AddCommand(pageCreateCmd)
	pageCmd.AddCommand(pageUpdateCmd)
	pageCmd.AddCommand(pageDeleteCmd)
	pageCmd.AddCommand(pageInteractiveCmd)

	// List flags
	pageListCmd.Flags().String("project", "", "Project identifier (required)")
	pageListCmd.MarkFlagRequired("project")

	// Create flags
	pageCreateCmd.Flags().String("project", "", "Project identifier (required)")
	pageCreateCmd.Flags().String("name", "", "Page name (required)")
	pageCreateCmd.Flags().String("description", "", "Page content/description")
	pageCreateCmd.Flags().String("description-file", "", "Read page content from file")
	pageCreateCmd.Flags().String("parent", "", "Parent page ID")
	pageCreateCmd.Flags().String("access", "public", "Page access (public, private)")
	pageCreateCmd.MarkFlagRequired("project")
	pageCreateCmd.MarkFlagRequired("name")

	// Update flags
	pageUpdateCmd.Flags().String("project", "", "Project identifier (required)")
	pageUpdateCmd.Flags().String("id", "", "Page ID (required)")
	pageUpdateCmd.Flags().String("name", "", "New page name")
	pageUpdateCmd.Flags().String("description", "", "New page content")
	pageUpdateCmd.Flags().String("description-file", "", "Read new content from file")
	pageUpdateCmd.Flags().String("parent", "", "New parent page ID")
	pageUpdateCmd.Flags().String("access", "", "New access level")
	pageUpdateCmd.MarkFlagRequired("project")
	pageUpdateCmd.MarkFlagRequired("id")

	// Delete flags
	pageDeleteCmd.Flags().String("project", "", "Project identifier (required)")
	pageDeleteCmd.Flags().String("id", "", "Page ID (required)")
	pageDeleteCmd.MarkFlagRequired("project")
	pageDeleteCmd.MarkFlagRequired("id")
}

func runPageList(cmd *cobra.Command, args []string) error {
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

	pages, err := client.GetPages(projectID)
	if err != nil {
		return fmt.Errorf("failed to get pages: %w", err)
	}

	if len(pages) == 0 {
		fmt.Println("No pages found in this project.")
		return nil
	}

	fmt.Printf("\n📄 Pages (%d):\n\n", len(pages))
	fmt.Printf("%-5s %-36s %-30s %-10s\n", "#", "ID", "NAME", "ACCESS")
	fmt.Println(strings.Repeat("-", 85))

	for i, p := range pages {
		name := truncate(p.Name, 28)
		access := p.Access
		if access == "" {
			access = "public"
		}
		fmt.Printf("%-5d %-36s %-30s %-10s\n", i+1, p.ID, name, access)
	}

	fmt.Println()
	return nil
}

func runPageCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	descriptionFile, _ := cmd.Flags().GetString("description-file")
	parent, _ := cmd.Flags().GetString("parent")
	access, _ := cmd.Flags().GetString("access")
	workspace, _ := cmd.Flags().GetString("workspace")

	// Read from file if specified
	if descriptionFile != "" {
		content, err := os.ReadFile(descriptionFile)
		if err != nil {
			return fmt.Errorf("failed to read description file: %w", err)
		}
		description = string(content)
	}

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

	create := &plane.PageCreate{
		Name:            name,
		Description:     description,
		DescriptionHTML: description,
		ParentID:        parent,
		Access:          access,
	}

	page, err := client.CreatePage(projectID, create)
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	fmt.Printf("\n✅ Created page:\n")
	fmt.Printf("   ID: %s\n", page.ID)
	fmt.Printf("   Name: %s\n", page.Name)
	if description != "" {
		fmt.Printf("   Content: %d characters\n", len(description))
	}

	return nil
}

func runPageUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	pageID, _ := cmd.Flags().GetString("id")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	descriptionFile, _ := cmd.Flags().GetString("description-file")
	parent, _ := cmd.Flags().GetString("parent")
	access, _ := cmd.Flags().GetString("access")
	workspace, _ := cmd.Flags().GetString("workspace")

	// Read from file if specified
	if descriptionFile != "" {
		content, err := os.ReadFile(descriptionFile)
		if err != nil {
			return fmt.Errorf("failed to read description file: %w", err)
		}
		description = string(content)
	}

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

	update := &plane.PageUpdate{}
	if name != "" {
		update.Name = name
	}
	if description != "" {
		update.Description = description
		update.DescriptionHTML = description
	}
	if parent != "" {
		update.ParentID = parent
	}
	if access != "" {
		update.Access = access
	}

	page, err := client.UpdatePage(projectID, pageID, update)
	if err != nil {
		return fmt.Errorf("failed to update page: %w", err)
	}

	fmt.Printf("\n✅ Updated page:\n")
	fmt.Printf("   ID: %s\n", page.ID)
	fmt.Printf("   Name: %s\n", page.Name)

	return nil
}

func runPageDelete(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	projectID, _ := cmd.Flags().GetString("project")
	pageID, _ := cmd.Flags().GetString("id")
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

	// Get page info for confirmation
	page, err := client.GetPage(projectID, pageID)
	if err != nil {
		return fmt.Errorf("failed to get page: %w", err)
	}

	confirmed, err := confirm(fmt.Sprintf("Are you sure you want to delete page '%s'?", page.Name))
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("❌ Deletion cancelled.")
		return nil
	}

	if err := client.DeletePage(projectID, pageID); err != nil {
		return fmt.Errorf("failed to delete page: %w", err)
	}

	fmt.Println("\n✅ Page deleted successfully.")
	return nil
}

func runPageInteractive(cmd *cobra.Command, args []string) error {
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
		fmt.Println("\n📄 Page Management")

		options := []string{
			"List all pages",
			"Create new page",
			"Update page",
			"Delete page",
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
			fmt.Println("\n👋 Goodbye!")
			return nil
		}
	}
}

func listPagesInteractive(client *plane.Client, projectID string) error {
	pages, err := client.GetPages(projectID)
	if err != nil {
		return err
	}

	if len(pages) == 0 {
		fmt.Println("\nNo pages found.")
		return nil
	}

	fmt.Printf("\n📄 Pages (%d):\n\n", len(pages))
	fmt.Printf("%-5s %-36s %-30s %-10s\n", "#", "ID", "NAME", "ACCESS")
	fmt.Println(strings.Repeat("-", 85))

	for i, p := range pages {
		name := truncate(p.Name, 28)
		access := p.Access
		if access == "" {
			access = "public"
		}
		fmt.Printf("%-5d %-36s %-30s %-10s\n", i+1, p.ID, name, access)
	}

	fmt.Println()
	return nil
}

func createPageInteractive(client *plane.Client, projectID string) error {
	fmt.Println("\n➕ Create New Page")

	name, err := input("Page name:")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("page name is required")
	}

	// Content selection
	contentOptions := []string{
		"Load from file",
		"Enter text directly",
	}

	contentIdx, err := selectOption("Content source:", contentOptions)
	if err != nil {
		return err
	}

	var content string
	switch contentIdx {
	case 0:
		// Find markdown files
		searchDirs := []string{
			"pages",
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
			var options []string
			for _, file := range mdFiles {
				options = append(options, file)
			}
			options = append(options, "Enter custom path")

			idx, err := selectOption("Select a file:", options)
			if err != nil {
				return err
			}

			if idx < len(mdFiles) {
				fileContent, err := os.ReadFile(mdFiles[idx])
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				content = string(fileContent)
			} else {
				// Custom path
				path, err := input("Enter file path:")
				if err != nil {
					return err
				}
				fileContent, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				content = string(fileContent)
			}
		} else {
			path, err := input("Enter file path:")
			if err != nil {
				return err
			}
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			content = string(fileContent)
		}

	case 1:
		var lines []string
		prompt := &survey.Multiline{
			Message: "Enter content (supports multiple lines):",
		}
		err := survey.AskOne(prompt, &content)
		if err != nil {
			if err.Error() == "interrupt" {
				return fmt.Errorf("content entry cancelled")
			}
			return err
		}
		content = strings.TrimSpace(content)
		_ = lines // Suppress unused variable warning
	}

	accessOptions := []string{
		"Public",
		"Private",
	}

	accessIdx, err := selectOption("Access level:", accessOptions)
	if err != nil {
		return err
	}

	accessValues := []string{"public", "private"}
	access := accessValues[accessIdx]

	create := &plane.PageCreate{
		Name:            name,
		Description:     content,
		DescriptionHTML: content,
		Access:          access,
	}

	page, err := client.CreatePage(projectID, create)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Created page: %s (ID: %s)\n", page.Name, page.ID)
	if content != "" {
		fmt.Printf("   Content: %d characters\n", len(content))
	}
	return nil
}

func updatePageInteractive(client *plane.Client, projectID string) error {
	pages, err := client.GetPages(projectID)
	if err != nil {
		return err
	}

	if len(pages) == 0 {
		return fmt.Errorf("no pages found")
	}

	// Build options
	var options []string
	for _, p := range pages {
		options = append(options, p.Name)
	}

	idx, err := selectOption("Select page to update:", options)
	if err != nil {
		return err
	}

	page := pages[idx]

	fmt.Printf("\n✏️  Update Page: %s\n", page.Name)

	update := &plane.PageUpdate{}

	name, err := inputWithDefault(fmt.Sprintf("New name (current: %s):", page.Name), "")
	if err != nil {
		return err
	}
	if name != "" {
		update.Name = name
	}

	updateContent, err := confirm("Update content?")
	if err != nil {
		return err
	}

	if updateContent {
		contentOptions := []string{
			"Load from file",
			"Enter text directly",
		}

		contentIdx, err := selectOption("Content source:", contentOptions)
		if err != nil {
			return err
		}

		var content string
		switch contentIdx {
		case 0:
			path, err := input("Enter file path:")
			if err != nil {
				return err
			}
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			content = string(fileContent)

		case 1:
			prompt := &survey.Multiline{
				Message: "Enter new content:",
			}
			err := survey.AskOne(prompt, &content)
			if err != nil {
				if err.Error() == "interrupt" {
					return fmt.Errorf("content entry cancelled")
				}
				return err
			}
			content = strings.TrimSpace(content)
		}

		update.Description = content
		update.DescriptionHTML = content
	}

	updated, err := client.UpdatePage(projectID, page.ID, update)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Updated page: %s\n", updated.Name)
	return nil
}

func deletePageInteractive(client *plane.Client, projectID string) error {
	pages, err := client.GetPages(projectID)
	if err != nil {
		return err
	}

	if len(pages) == 0 {
		return fmt.Errorf("no pages found")
	}

	// Build options
	var options []string
	for _, p := range pages {
		options = append(options, p.Name)
	}

	idx, err := selectOption("Select page to delete:", options)
	if err != nil {
		return err
	}

	page := pages[idx]

	confirmed, err := confirm(fmt.Sprintf("Delete page '%s'?", page.Name))
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("❌ Deletion cancelled.")
		return nil
	}

	if err := client.DeletePage(projectID, page.ID); err != nil {
		return err
	}

	fmt.Println("\n✅ Page deleted.")
	return nil
}
