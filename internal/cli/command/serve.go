package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/d3/internal/mcp"
)

// ServeCommand represents the serve command implementation
type ServeCommand struct{}

// NewServeCommand creates a new cobra command for the serve functionality
func NewServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start an MCP server for d3",
		Long:  "Start a Model Context Protocol server that exposes d3 functionality to LLM clients",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the working directory flag value here
			workdirFlag, _ := cmd.Flags().GetString("workdir") // Error can be ignored, defaults to ""
			return runServe(workdirFlag)
		},
	}

	// Add a persistent flag for the working directory
	cmd.PersistentFlags().StringP("workdir", "w", "", "Specify the working directory (project root)")

	return cmd
}

// runServe handles the serve command execution, now accepting workdir flag
func runServe(workdirFlag string) error {
	// Create command instance
	command := &ServeCommand{}

	var workspaceRoot string
	var err error

	if workdirFlag != "" {
		// Use the flag value if provided
		workspaceRoot = workdirFlag
		// Optional: Add validation to check if the directory exists
		if _, statErr := os.Stat(workspaceRoot); os.IsNotExist(statErr) {
			return fmt.Errorf("specified working directory '%s' does not exist", workspaceRoot)
		}
	} else {
		// Use the current working directory as the default
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Execute the command
	result, err := command.Run(context.Background(), workspaceRoot)

	if err != nil {
		return err
	}

	// Print the result message (although for stdio server, this may never be seen)
	fmt.Println(result.Message)

	return nil
}

// Run implements a modified Command interface for serve
func (s *ServeCommand) Run(ctx context.Context, workspaceRoot string) (Result, error) {
	// Create MCP server
	server := mcp.NewServer(workspaceRoot)

	// Start the stdio server
	err := mcp.ServeStdio(server)

	if err != nil {
		return Result{}, fmt.Errorf("failed to serve MCP: %w", err)
	}

	return NewResult("MCP server started", nil, nil), nil
}
