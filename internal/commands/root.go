package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "plane-cli",
	Short: "CLI tool for managing Plane.so work items",
	Long: `Plane CLI is a command-line tool to automate work item management
in Plane.so project management software.

Features:
- Create and update work items
- Fuzzy title matching for bulk operations
- Template-based description generation
- Multi-project support

For more information, visit: https://plane.so`,
	Version: "1.0.0",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().String("config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().String("workspace", "", "Plane workspace slug")
}
