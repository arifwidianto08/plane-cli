package commands

import (
	"github.com/spf13/cobra"
	"plane-cli/internal/config"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Plane CLI settings",
	Long: `Interactive configuration for Plane CLI.

This command allows you to view and update your Plane.so API credentials:
- Base URL (e.g., https://project.lazuardy.tech)
- API Token (your Plane API key)
- Workspace slug (e.g., lazuardy-tech)

Configuration is saved to .env file in the current directory.

Examples:
  # View current configuration
  plane-cli configure --show

  # Update configuration interactively
  plane-cli configure`,
	RunE: runConfigure,
}

func init() {
	rootCmd.AddCommand(configureCmd)

	configureCmd.Flags().Bool("show", false, "Show current configuration without interactive prompts")
}

func runConfigure(cmd *cobra.Command, args []string) error {
	showOnly, _ := cmd.Flags().GetBool("show")

	if showOnly {
		config.ShowCurrentConfig()
		return nil
	}

	return config.InteractiveSetup()
}
