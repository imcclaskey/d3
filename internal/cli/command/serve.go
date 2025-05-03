package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/d3/internal/common"
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
			return runServe()
		},
	}

	return cmd
}

// runServe handles the serve command execution
func runServe() error {
	// Create command instance
	command := &ServeCommand{}

	// Execute the command
	workspaceRoot, err := common.GetWorkspaceRoot()
	if err != nil {
		return fmt.Errorf("could not determine workspace root for server: %w", err)
	}
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

	// Start serving over stdio
	err := server.Serve()

	if err != nil {
		return Result{}, fmt.Errorf("failed to serve MCP: %w", err)
	}

	return NewResult("MCP server started", nil, nil), nil
}
