package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/d3/internal/project"
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
		Short: "Initialize d3 in the current workspace",
		Long:  "Initialize d3 in the current workspace and create base project files",
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

	// Get project root (using os.Getwd as a fallback for now)
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine project root: %w", err)
	}

	// Execute the command
	cfg := NewConfig(projectRoot)
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
	// Initialize project instance
	proj := project.New(cfg.WorkspaceRoot)

	// Run the project initialization logic
	result, err := proj.Init(c.clean)
	if err != nil {
		return Result{}, fmt.Errorf("failed to initialize project: %w", err)
	}

	// Convert project Result to CLI Result
	return NewResult(result.FormatCLI(), nil, nil), nil
}
