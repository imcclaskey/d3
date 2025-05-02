// Package cli implements the i3 command-line interface
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/i3/internal/cli/command"
	"github.com/imcclaskey/i3/internal/common"
	"github.com/imcclaskey/i3/internal/version"
)

// CLI represents the i3 command-line interface
type CLI struct {
	rootCmd *cobra.Command
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	rootCmd := &cobra.Command{
		Use:   "i3",
		Short: "i3 Framework CLI",
		Long:  "Interactive Intelligent Interface (i3) Framework CLI",
	}

	return &CLI{
		rootCmd: rootCmd,
	}
}

// InitCommands initializes all CLI commands
func (c *CLI) InitCommands() {
	// Add commands to the root command
	c.rootCmd.AddCommand(command.NewCreateCommand())
	c.rootCmd.AddCommand(command.NewServeCommand())
	c.rootCmd.AddCommand(command.NewInitCommand())

	// Version command
	c.rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version of i3",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("i3 version %s\n", version.Version)
		},
	})
}

// Execute executes the CLI
func (c *CLI) Execute() error {
	return c.rootCmd.Execute()
}

// GetWorkspaceRoot gets the workspace root directory
// It returns an error if the current working directory cannot be determined.
func GetWorkspaceRoot() (string, error) {
	// Use the common implementation to ensure consistency and handle errors
	return common.GetWorkspaceRoot()
}
