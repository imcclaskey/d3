package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/phase"
	"github.com/imcclaskey/d3/internal/core/ports"
	"github.com/imcclaskey/d3/internal/core/projectfiles"
	"github.com/imcclaskey/d3/internal/core/rules"
	"github.com/imcclaskey/d3/internal/core/session"
	"github.com/imcclaskey/d3/internal/project"
)

// FeatureCreateCommand holds dependencies for the feature create command.
type FeatureCreateCommand struct {
	featureName string
	projectSvc  project.ProjectService
}

// NewFeatureCreateCommand creates a new cobra command for creating features.
// This is intended to be a subcommand of 'feature'.
func NewFeatureCreateCommand() *cobra.Command {
	cmdRunner := &FeatureCreateCommand{}
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new feature and set it as the current context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdRunner.featureName = args[0]

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
			fileOp := projectfiles.NewDefaultFileOperator()

			cmdRunner.projectSvc = project.New(cfg.WorkspaceRoot, fs, sessionSvc, featureSvc, rulesSvc, phaseSvc, fileOp)

			return cmdRunner.run(context.Background())
		},
	}
	return cmd
}

// run is the core logic for feature creation.
func (c *FeatureCreateCommand) run(ctx context.Context) error {
	if c.projectSvc == nil {
		return fmt.Errorf("project service not initialized in FeatureCreateCommand")
	}
	result, err := c.projectSvc.CreateFeature(ctx, c.featureName)
	if err != nil {
		return err
	}
	fmt.Println(result.FormatCLI())
	return nil
}
