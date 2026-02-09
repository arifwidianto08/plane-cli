package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"plane-cli/internal/templates"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize plane-cli configuration",
	Long: `Initialize plane-cli by creating configuration files and directories.

This command will:
- Create .env file with API configuration
- Create config.yaml with defaults
- Create templates directory with default templates
- Guide you through setup`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Welcome to Plane CLI!")
	fmt.Println("Let's set up your configuration.\n")

	// Check if already initialized
	if _, err := os.Stat(".env"); err == nil {
		fmt.Print("Configuration files already exist. Overwrite? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(response)) != "y" {
			fmt.Println("Setup cancelled.")
			return nil
		}
	}

	// Get configuration from user
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Please provide the following information:\n")

	// Base URL
	fmt.Print("Plane Base URL (e.g., https://plane.your-domain.com): ")
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)

	// API Token
	fmt.Print("API Token (from Plane settings): ")
	apiToken, _ := reader.ReadString('\n')
	apiToken = strings.TrimSpace(apiToken)

	// Workspace
	fmt.Print("Workspace slug (optional): ")
	workspace, _ := reader.ReadString('\n')
	workspace = strings.TrimSpace(workspace)

	// Default project
	fmt.Print("Default project identifier (optional): ")
	defaultProject, _ := reader.ReadString('\n')
	defaultProject = strings.TrimSpace(defaultProject)

	// Validate required fields
	if baseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	// Create .env file
	envContent := fmt.Sprintf(`# Plane CLI Configuration
PLANE_BASE_URL=%s
PLANE_API_TOKEN=%s
`, baseURL, apiToken)

	if err := os.WriteFile(".env", []byte(envContent), 0600); err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}
	fmt.Println("✓ Created .env")

	// Create config.yaml
	configContent := `defaults:
  project: """

# Project shortcuts (optional)
# projects:
#   short-name: "actual-project-identifier"

templates:
  directory: "./templates"
  default: "feature"

fuzzy:
  min_score: 60
  max_results: 10
`

	if defaultProject != "" {
		configContent = fmt.Sprintf(`defaults:
  project: "%s"

# Project shortcuts (optional)
# projects:
#   short-name: "actual-project-identifier"

templates:
  directory: "./templates"
  default: "feature"

fuzzy:
  min_score: 60
  max_results: 10
`, defaultProject)
	}

	if err := os.WriteFile("config.yaml", []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create config.yaml: %w", err)
	}
	fmt.Println("✓ Created config.yaml")

	// Create templates directory and default templates
	templatesDir := "./templates"
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Create default templates
	if err := templates.CreateDefaultTemplates(templatesDir); err != nil {
		return fmt.Errorf("failed to create default templates: %w", err)
	}
	fmt.Println("✓ Created templates directory with default templates")

	// Create .gitignore
	gitignoreContent := `# Plane CLI
.env
config.yaml
templates/custom/
`
	if err := os.WriteFile(".gitignore", []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}
	fmt.Println("✓ Created .gitignore")

	// Success message
	fmt.Println("\n🎉 Setup complete!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review and customize .env and config.yaml")
	fmt.Println("  2. Add your templates to ./templates/")
	fmt.Println("  3. Run 'plane-cli --help' to see available commands")
	fmt.Println("\nQuick start:")
	fmt.Println("  plane-cli list --project your-project")
	fmt.Println("  plane-cli create --project your-project --title \"My Task\" --template feature")

	return nil
}

// Helper function to get absolute path
func getAbsolutePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Abs(path)
}
