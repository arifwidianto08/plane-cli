package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

// IsConfigured checks if the essential configuration is present
func IsConfigured() bool {
	// Try to load .env file
	godotenv.Load(".env")

	baseURL := os.Getenv("PLANE_BASE_URL")
	apiToken := os.Getenv("PLANE_API_TOKEN")
	workspace := os.Getenv("PLANE_WORKSPACE")

	return baseURL != "" && apiToken != "" && workspace != ""
}

// CheckAndPromptConfig checks if config exists and prompts user if not
// Returns the config and a boolean indicating if it was just configured
func CheckAndPromptConfig() (*Config, bool, error) {
	if IsConfigured() {
		cfg, err := Load()
		return cfg, false, err
	}

	// Configuration missing, prompt user
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("       🔧 Configuration Required")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("\nWelcome to Plane CLI! Let's set up your configuration.")
	fmt.Println("You'll need your Plane.so API credentials.")
	fmt.Println("\nYou can find these in your Plane workspace settings:")
	fmt.Println("  1. Go to your Plane workspace")
	fmt.Println("  2. Navigate to Settings → API")
	fmt.Println("  3. Copy your API Token and note your workspace slug")
	fmt.Println(strings.Repeat("-", 70))

	reader := bufio.NewReader(os.Stdin)

	baseURL, err := promptForBaseURL(reader)
	if err != nil {
		return nil, false, err
	}

	apiToken, err := promptForAPIToken(reader)
	if err != nil {
		return nil, false, err
	}

	workspace, err := promptForWorkspace(reader)
	if err != nil {
		return nil, false, err
	}

	// Save configuration
	envData := map[string]string{
		"PLANE_BASE_URL":  baseURL,
		"PLANE_API_TOKEN": apiToken,
		"PLANE_WORKSPACE": workspace,
	}

	if err := SaveToEnv(envData); err != nil {
		return nil, false, fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("\n✅ Configuration saved to .env")
	fmt.Println(strings.Repeat("=", 70))

	// Load and return the newly saved config
	cfg, err := Load()
	return cfg, true, err
}

// InteractiveSetup prompts user for all configuration values
func InteractiveSetup() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("       🔧 Plane CLI Configuration")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Show current values if they exist
	currentBaseURL := os.Getenv("PLANE_BASE_URL")
	currentToken := os.Getenv("PLANE_API_TOKEN")
	currentWorkspace := os.Getenv("PLANE_WORKSPACE")

	if currentBaseURL != "" || currentToken != "" || currentWorkspace != "" {
		fmt.Println("Current Configuration:")
		fmt.Println(strings.Repeat("-", 70))
		if currentBaseURL != "" {
			fmt.Printf("Base URL:   %s\n", currentBaseURL)
		}
		if currentToken != "" {
			fmt.Printf("API Token:  %s\n", maskToken(currentToken))
		}
		if currentWorkspace != "" {
			fmt.Printf("Workspace:  %s\n", currentWorkspace)
		}
		fmt.Println(strings.Repeat("-", 70))
		fmt.Println()
	}

	// Ask what to update
	fmt.Println("What would you like to do?")
	fmt.Println("1. Update Base URL")
	fmt.Println("2. Update API Token")
	fmt.Println("3. Update Workspace")
	fmt.Println("4. Update All Settings")
	fmt.Println("5. Cancel")
	fmt.Println()
	fmt.Print("Enter choice (1-5): ")

	choice, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}
	choice = strings.TrimSpace(choice)

	envData := make(map[string]string)

	// Load existing values
	if currentBaseURL != "" {
		envData["PLANE_BASE_URL"] = currentBaseURL
	}
	if currentToken != "" {
		envData["PLANE_API_TOKEN"] = currentToken
	}
	if currentWorkspace != "" {
		envData["PLANE_WORKSPACE"] = currentWorkspace
	}

	switch choice {
	case "1":
		baseURL, err := promptForBaseURL(reader)
		if err != nil {
			return err
		}
		envData["PLANE_BASE_URL"] = baseURL

	case "2":
		apiToken, err := promptForAPIToken(reader)
		if err != nil {
			return err
		}
		envData["PLANE_API_TOKEN"] = apiToken

	case "3":
		workspace, err := promptForWorkspace(reader)
		if err != nil {
			return err
		}
		envData["PLANE_WORKSPACE"] = workspace

	case "4":
		baseURL, err := promptForBaseURL(reader)
		if err != nil {
			return err
		}
		envData["PLANE_BASE_URL"] = baseURL

		apiToken, err := promptForAPIToken(reader)
		if err != nil {
			return err
		}
		envData["PLANE_API_TOKEN"] = apiToken

		workspace, err := promptForWorkspace(reader)
		if err != nil {
			return err
		}
		envData["PLANE_WORKSPACE"] = workspace

	case "5", "cancel", "c":
		fmt.Println("\n❌ Configuration cancelled.")
		return nil

	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}

	if len(envData) == 0 {
		return fmt.Errorf("no configuration to save")
	}

	if err := SaveToEnv(envData); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("\n✅ Configuration updated successfully!")
	return nil
}

// promptForBaseURL prompts user for base URL with validation
func promptForBaseURL(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("\nEnter Plane Base URL (e.g., https://project.lazuardy.tech): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		baseURL := strings.TrimSpace(input)

		// Validate URL format
		if baseURL == "" {
			fmt.Println("❌ Base URL is required.")
			continue
		}

		// Check if it looks like a URL
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			fmt.Println("❌ URL must start with http:// or https://")
			continue
		}

		// Remove trailing slash
		baseURL = strings.TrimSuffix(baseURL, "/")

		return baseURL, nil
	}
}

// promptForAPIToken prompts user for API token
func promptForAPIToken(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("\nEnter Plane API Token: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		apiToken := strings.TrimSpace(input)

		if apiToken == "" {
			fmt.Println("❌ API Token is required.")
			continue
		}

		return apiToken, nil
	}
}

// promptForWorkspace prompts user for workspace slug
func promptForWorkspace(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("\nEnter Workspace slug (e.g., lazuardy-tech): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		workspace := strings.TrimSpace(input)

		if workspace == "" {
			fmt.Println("❌ Workspace slug is required.")
			continue
		}

		return workspace, nil
	}
}

// SaveToEnv saves configuration to .env file
func SaveToEnv(data map[string]string) error {
	envPath := filepath.Join(".", ".env")

	// Read existing file content if it exists
	existingContent := ""
	if _, err := os.Stat(envPath); err == nil {
		content, err := os.ReadFile(envPath)
		if err == nil {
			existingContent = string(content)
		}
	}

	// Parse existing content into map
	lines := strings.Split(existingContent, "\n")
	envMap := make(map[string]string)
	var lineOrder []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			lineOrder = append(lineOrder, line)
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			envMap[key] = value
			lineOrder = append(lineOrder, key)
		} else {
			lineOrder = append(lineOrder, line)
		}
	}

	// Update with new values
	for key, value := range data {
		envMap[key] = value
		// Add to order if not exists
		found := false
		for _, k := range lineOrder {
			if k == key {
				found = true
				break
			}
		}
		if !found {
			lineOrder = append(lineOrder, key)
		}
	}

	// Build new content
	var newLines []string
	writtenKeys := make(map[string]bool)

	for _, item := range lineOrder {
		if item == "" || strings.HasPrefix(item, "#") {
			newLines = append(newLines, item)
			continue
		}

		if value, exists := envMap[item]; exists && !writtenKeys[item] {
			newLines = append(newLines, fmt.Sprintf("%s=%s", item, value))
			writtenKeys[item] = true
		}
	}

	// Write to file
	content := strings.Join(newLines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}

// maskToken masks the API token for display
func maskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:8] + strings.Repeat("*", len(token)-12) + token[len(token)-4:]
}

// ShowCurrentConfig displays the current configuration
func ShowCurrentConfig() {
	// Load .env file first
	godotenv.Load(".env")

	baseURL := os.Getenv("PLANE_BASE_URL")
	apiToken := os.Getenv("PLANE_API_TOKEN")
	workspace := os.Getenv("PLANE_WORKSPACE")

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("       📋 Current Configuration")
	fmt.Println(strings.Repeat("=", 70))

	if baseURL == "" && apiToken == "" && workspace == "" {
		fmt.Println("\n❌ No configuration found.")
		fmt.Println("Run 'plane-cli configure' to set up your configuration.")
	} else {
		fmt.Println()
		if baseURL != "" {
			fmt.Printf("Base URL:   %s\n", baseURL)
		} else {
			fmt.Println("Base URL:   ❌ Not set")
		}

		if apiToken != "" {
			fmt.Printf("API Token:  %s\n", maskToken(apiToken))
		} else {
			fmt.Println("API Token:  ❌ Not set")
		}

		if workspace != "" {
			fmt.Printf("Workspace:  %s\n", workspace)
		} else {
			fmt.Println("Workspace:  ❌ Not set")
		}
	}

	fmt.Println(strings.Repeat("=", 70))
}

// ValidateConfig validates that all required configuration is present
func ValidateConfig() error {
	godotenv.Load(".env")

	missing := []string{}

	if os.Getenv("PLANE_BASE_URL") == "" {
		missing = append(missing, "PLANE_BASE_URL")
	}
	if os.Getenv("PLANE_API_TOKEN") == "" {
		missing = append(missing, "PLANE_API_TOKEN")
	}
	if os.Getenv("PLANE_WORKSPACE") == "" {
		missing = append(missing, "PLANE_WORKSPACE")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	return nil
}

// extractWorkspaceFromURL extracts workspace slug from Plane URL
func extractWorkspaceSlug(url string) string {
	// Pattern: https://project.example.com/api/v1/workspaces/{workspace}/...
	re := regexp.MustCompile(`/workspaces/([^/]+)/`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
