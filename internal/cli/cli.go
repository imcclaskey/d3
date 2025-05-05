// Package cli implements the d3 command-line interface
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/d3/internal/cli/command"
	"github.com/imcclaskey/d3/internal/version"
)

// CLI represents the d3 command-line interface
type CLI struct {
	rootCmd *cobra.Command
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	rootCmd := &cobra.Command{
		Use:   "d3",
		Short: "d3 Framework CLI",
		Long:  "Define, Design, Deliver (d3) Framework CLI",
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
		Short: "Print the version of d3",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("d3 version %s\n", version.Version)
		},
	})
}

// Execute executes the CLI
func (c *CLI) Execute() error {
	return c.rootCmd.Execute()
}
