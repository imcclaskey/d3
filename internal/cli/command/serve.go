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
			workdirFlag, _ := cmd.Flags().GetString("workdir") // Error can be ignored, defaults to ""
			return runServe(workdirFlag)
		},
	}

	// Add a persistent flag for the working directory
	cmd.PersistentFlags().StringP("workdir", "w", "", "Specify the working directory (project root)")

	return cmd
}

// runServe handles the serve command execution
func runServe(workdirFlag string) error {
	command := &ServeCommand{}

	var workspaceRoot string
	var err error

	if workdirFlag != "" {
		workspaceRoot = workdirFlag
		// Check if the directory exists and is accessible
		_, statErr := os.Stat(workspaceRoot)
		if os.IsNotExist(statErr) {
			return fmt.Errorf("specified working directory '%s' does not exist", workspaceRoot)
		} else if statErr != nil {
			// This will catch permission errors and other issues
			return fmt.Errorf("cannot access working directory '%s': %w", workspaceRoot, statErr)
		}
	} else {
		// Use the current working directory as the default
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	result, err := command.Run(context.Background(), workspaceRoot)

	if err != nil {
		return err
	}

	fmt.Println(result.Message)

	return nil
}

// Run implements a modified Command interface for serve
func (s *ServeCommand) Run(ctx context.Context, workspaceRoot string) (Result, error) {
	server := mcp.NewServer(workspaceRoot)

	err := mcp.ServeStdio(server)

	if err != nil {
		return Result{}, fmt.Errorf("failed to serve MCP: %w", err)
	}

	return NewResult("MCP server started", nil, nil), nil
}
