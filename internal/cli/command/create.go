package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/i3/internal/common"
	"github.com/imcclaskey/i3/internal/core"
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

	// Execute the command
	workspaceRoot, err := common.GetWorkspaceRoot()
	if err != nil {
		return fmt.Errorf("could not determine workspace root: %w", err)
	}
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
	// Create core services
	services := core.NewServices(cfg.WorkspaceRoot)

	// Call the feature service to create a new feature
	result, err := services.Feature.Create(ctx, c.featureName)
	if err != nil {
		return Result{}, err
	}

	return NewResult(result.Message, result, nil), nil
}
