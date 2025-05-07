package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	// Import necessary core packages for service instantiation
	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/phase"
	"github.com/imcclaskey/d3/internal/core/ports"
	"github.com/imcclaskey/d3/internal/core/rules"
	"github.com/imcclaskey/d3/internal/core/session"
	"github.com/imcclaskey/d3/internal/project"
)

// ExitCommand holds dependencies for the exit command.
type ExitCommand struct {
	projectSvc project.ProjectService // Dependency for the project service
}

// NewExitCommand creates a new cobra command for exiting the current feature context.
func NewExitCommand() *cobra.Command {
	cmdRunner := &ExitCommand{}
	cmd := &cobra.Command{
		Use:   "exit",
		Short: "Exit the current feature context, clearing active feature state.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Instantiate ProjectService
			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not determine workspace root: %w", err)
			}
			cfg := NewConfig(projectRoot)

			fs := ports.RealFileSystem{}
			sessionSvc := session.NewStorage(cfg.D3Dir, fs)
			featureSvc := feature.NewService(cfg.WorkspaceRoot, cfg.FeaturesDir, cfg.D3Dir, fs)
			ruleGenerator := rules.NewRuleGenerator()
			rulesSvc := rules.NewService(cfg.WorkspaceRoot, cfg.CursorRulesDir, ruleGenerator, fs)
			phaseSvc := phase.NewService(fs)

			cmdRunner.projectSvc = project.New(cfg.WorkspaceRoot, fs, sessionSvc, featureSvc, rulesSvc, phaseSvc)

			return cmdRunner.run(context.Background())
		},
	}
	return cmd
}

// run executes the logic to directly call ProjectService.ExitFeature.
func (c *ExitCommand) run(ctx context.Context) error {
	if c.projectSvc == nil {
		return fmt.Errorf("project service not initialized in ExitCommand")
	}

	result, err := c.projectSvc.ExitFeature(ctx)
	if err != nil {
		return err // Return the error from the service call
	}

	// Print the result message from the project service
	fmt.Println(result.FormatCLI())

	return nil
}
