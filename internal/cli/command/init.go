package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/i3/internal/common"
	"github.com/imcclaskey/i3/internal/core"
)

// InitCommand represents the init command implementation
type InitCommand struct {
	clean bool
}

// NewInitCommand creates a new cobra command for the init functionality
func NewInitCommand() *cobra.Command {
	// Create command instance to store flags
	cmd := &InitCommand{}

	// Create cobra command
	cobraCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize i3 in the current workspace",
		Long:  "Initialize i3 in the current workspace and create base project files",
		Args:  cobra.NoArgs,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return runInit(cmd.clean)
		},
	}

	// Add flags
	cobraCmd.Flags().BoolVar(&cmd.clean, "clean", false, "Perform a clean initialization (remove existing files)")

	return cobraCmd
}

// runInit handles the init command execution
func runInit(clean bool) error {
	// Create command instance
	command := &InitCommand{
		clean: clean,
	}

	// Get workspace root
	workspaceRoot, err := common.GetWorkspaceRoot()
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
func (c *InitCommand) Run(ctx context.Context, cfg Config) (Result, error) {
	// Create core services
	services := core.NewServices(cfg.WorkspaceRoot)

	// Initialize workspace
	message, newlyCreated, err := services.Files.InitWorkspace(c.clean)
	if err != nil {
		return Result{}, fmt.Errorf("failed to initialize workspace: %w", err)
	}

	// Prepare a more detailed message for the user
	resultMsg := message
	if len(newlyCreated) > 0 {
		resultMsg = fmt.Sprintf("%s\nCreated files: %v", message, newlyCreated)
	}

	return NewResult(resultMsg, nil, nil), nil
}
