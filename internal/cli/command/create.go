package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/d3/internal/project"
)

// CreateCommand represents the create command implementation
type CreateCommand struct {
	featureName string
}

// NewCreateCommand creates a new cobra command for the create functionality
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <feature>",
		Short: "Create a new feature and set it as the current context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(args[0])
		},
	}

	return cmd
}

// runCreate handles the create command execution
func runCreate(featureName string) error {
	// Create command instance
	command := &CreateCommand{
		featureName: featureName,
	}

	// Get workspace root (using os.Getwd as fallback)
	workspaceRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine workspace root: %w", err)
	}

	// Execute the command
	cfg := NewConfig(workspaceRoot)
	result, err := command.Run(context.Background(), cfg)

	if err != nil {
		return err
	}

	// Print the result message
	fmt.Println(result.Message)

	return nil
}

// Run implements the Command interface
func (c *CreateCommand) Run(ctx context.Context, cfg Config) (Result, error) {
	// Initialize project
	proj := project.New(cfg.WorkspaceRoot)

	// Create the feature
	result, err := proj.CreateFeature(ctx, c.featureName)
	if err != nil {
		return Result{}, err
	}

	// Convert project Result to CLI Result
	return NewResult(result.FormatCLI(), nil, nil), nil
}
