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

// FeatureEnterCommand holds dependencies for the feature enter command.
type FeatureEnterCommand struct {
	featureName string
	projectSvc  project.ProjectService // Dependency for the project service
}

// NewFeatureEnterCommand creates a new cobra command for entering features.
func NewFeatureEnterCommand() *cobra.Command {
	cmdRunner := &FeatureEnterCommand{}
	cmd := &cobra.Command{
		Use:   "enter <name>",
		Short: "Enter a feature context, resuming its last known phase",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdRunner.featureName = args[0]

			// Instantiate ProjectService (similar to create.go)
			// Consider refactoring this setup logic into a shared helper if used often.
			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not determine workspace root: %w", err)
			}
			// Assuming NewConfig helper exists or is accessible
			cfg := NewConfig(projectRoot)

			fs := ports.RealFileSystem{}
			sessionSvc := session.NewStorage(cfg.D3Dir, fs)
			featureSvc := feature.NewService(cfg.WorkspaceRoot, cfg.FeaturesDir, cfg.D3Dir, fs)
			ruleGenerator := rules.NewRuleGenerator()
			ruleSvc := rules.NewService(cfg.WorkspaceRoot, cfg.CursorRulesDir, ruleGenerator, fs)
			phaseSvc := phase.NewService(fs)

			cmdRunner.projectSvc = project.New(cfg.WorkspaceRoot, fs, sessionSvc, featureSvc, ruleSvc, phaseSvc)

			return cmdRunner.run(context.Background())
		},
	}
	return cmd
}

// run executes the logic to directly call ProjectService.EnterFeature.
func (c *FeatureEnterCommand) run(ctx context.Context) error {
	if c.projectSvc == nil {
		return fmt.Errorf("project service not initialized in FeatureEnterCommand")
	}

	// EnterFeature now returns *Result, error
	result, err := c.projectSvc.EnterFeature(ctx, c.featureName)
	if err != nil {
		return err // Return the error from the service call
	}

	// Print the formatted result from the project service
	fmt.Println(result.FormatCLI())

	return nil
}
