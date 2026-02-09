package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"plane-cli/internal/config"
	"plane-cli/internal/templates"
)

// getTemplatesDir returns the templates directory path
func getTemplatesDir() string {
	templatesDir := "./templates"
	cfg, err := config.Load()
	if err == nil && cfg.TemplatesDir != "" {
		templatesDir = cfg.TemplatesDir
	}
	return templatesDir
}

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage description templates",
	Long: `Manage JSON templates for work item descriptions.

Examples:
  # List all templates
  plane-cli template list

  # Show template details
  plane-cli template show feature

  # Create new template
  plane-cli template create my-template

  # Delete template
  plane-cli template delete my-template`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all templates",
	RunE:  runTemplateList,
}

var templateShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show template details",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateShow,
}

var templateCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new template",
	Long: `Create a new template interactively or with flags.

Example:
  plane-cli template create my-feature`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateCreate,
}

var templateDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a template",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateDelete,
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateShowCmd)
	templateCmd.AddCommand(templateCreateCmd)
	templateCmd.AddCommand(templateDeleteCmd)

	// Create flags
	templateCreateCmd.Flags().String("description", "", "Template description")
	templateCreateCmd.Flags().String("content", "", "Template content")
	templateCreateCmd.Flags().StringSlice("vars", nil, "Template variables")
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	mgr, err := templates.NewManager(getTemplatesDir())
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	templateNames := mgr.List()
	if len(templateNames) == 0 {
		fmt.Println("No templates found.")
		fmt.Printf("Templates directory: %s\n", getTemplatesDir())
		return nil
	}

	fmt.Printf("Available templates (%d):\n\n", len(templateNames))
	for _, name := range templateNames {
		tmpl, err := mgr.Get(name)
		if err != nil {
			fmt.Printf("  - %s (error loading)\n", name)
			continue
		}
		fmt.Printf("  - %s: %s\n", name, tmpl.Description)
	}

	return nil
}

func runTemplateShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	mgr, err := templates.NewManager(getTemplatesDir())
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	tmpl, err := mgr.Get(name)
	if err != nil {
		return err
	}

	fmt.Printf("Template: %s\n", tmpl.Name)
	fmt.Printf("Description: %s\n", tmpl.Description)
	fmt.Printf("Variables: %v\n", tmpl.Variables)
	fmt.Printf("\nContent:\n%s\n", tmpl.Content)

	return nil
}

func runTemplateCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	mgr, err := templates.NewManager(getTemplatesDir())
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	// Check if template already exists
	if _, err := mgr.Get(name); err == nil {
		return fmt.Errorf("template '%s' already exists", name)
	}

	description, _ := cmd.Flags().GetString("description")
	content, _ := cmd.Flags().GetString("content")
	vars, _ := cmd.Flags().GetStringSlice("vars")

	// Interactive mode if content not provided
	if content == "" {
		fmt.Println("Enter template content (press Ctrl+D when done):")
		content = readMultiLineInput()
	}

	if content == "" {
		return fmt.Errorf("template content cannot be empty")
	}

	tmpl := &templates.Template{
		Name:        name,
		Description: description,
		Content:     content,
		Variables:   vars,
	}

	if err := mgr.Save(tmpl); err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	fmt.Printf("✓ Created template: %s\n", name)
	return nil
}

func runTemplateDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	mgr, err := templates.NewManager(getTemplatesDir())
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete template '%s'? (y/n): ", name)
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	if err := mgr.Delete(name); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	fmt.Printf("✓ Deleted template: %s\n", name)
	return nil
}

func readMultiLineInput() string {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return strings.Join(lines, "\n")
}
