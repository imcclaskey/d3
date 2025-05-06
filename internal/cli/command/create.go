package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/ports"
	"github.com/imcclaskey/d3/internal/core/rules"
	"github.com/imcclaskey/d3/internal/core/session"
	"github.com/imcclaskey/d3/internal/project"
)

// CreateCommand represents the create command implementation
type CreateCommand struct {
	featureName string
	projectSvc  project.ProjectService // Unexported field for dependency
}

// NewCreateCommand creates a new cobra command for the create functionality
func NewCreateCommand() *cobra.Command {
	cmdRunner := &CreateCommand{}
	cmd := &cobra.Command{
		Use:   "create <feature>",
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

			cmdRunner.projectSvc = project.New(cfg.WorkspaceRoot, fs, sessionSvc, featureSvc, rulesSvc)

			return cmdRunner.run(context.Background())
		},
	}
	return cmd
}

// run is the core logic, using the projectSvc field.
func (c *CreateCommand) run(ctx context.Context) error {
	if c.projectSvc == nil {
		return fmt.Errorf("project service not initialized in CreateCommand")
	}
	result, err := c.projectSvc.CreateFeature(ctx, c.featureName)
	if err != nil {
		return err // Error message from project service is likely sufficient
	}
	fmt.Println(result.FormatCLI())
	return nil
}
