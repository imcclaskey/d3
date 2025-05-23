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
	"github.com/imcclaskey/d3/internal/project"
)

// FeatureEnterCommand holds dependencies for the feature enter command.
type FeatureEnterCommand struct {
	featureName string
	projectSvc  project.ProjectService
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

			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not determine workspace root: %w", err)
			}
			cfg := NewConfig(projectRoot)

			fs := ports.RealFileSystem{}
			featureSvc := feature.NewService(cfg.WorkspaceRoot, cfg.FeaturesDir, cfg.D3Dir, fs)
			phaseSvc := phase.NewService(fs)
			ruleGenerator := rules.NewRuleGenerator(cfg.WorkspaceRoot, fs)
			rulesSvc := rules.NewService(cfg.WorkspaceRoot, cfg.CursorRulesDir, ruleGenerator, fs)
			fileOp := projectfiles.NewDefaultFileOperator()

			cmdRunner.projectSvc = project.New(cfg.WorkspaceRoot, fs, featureSvc, rulesSvc, phaseSvc, fileOp)

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

	result, err := c.projectSvc.EnterFeature(ctx, c.featureName)
	if err != nil {
		return err
	}

	fmt.Println(result.FormatCLI())

	return nil
}
