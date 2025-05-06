package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/phase"
	"github.com/imcclaskey/d3/internal/core/ports"
	"github.com/imcclaskey/d3/internal/core/rules"
	"github.com/imcclaskey/d3/internal/core/session"
	"github.com/imcclaskey/d3/internal/project"
)

// InitCommand represents the init command implementation
type InitCommand struct {
	clean bool
	// Use an unexported field for the project service dependency, allowing tests to set it.
	// Production code will set it with the real instance.
	projectSvc project.ProjectService
}

// NewInitCommand creates a new cobra command for the init functionality
func NewInitCommand() *cobra.Command {
	cmdRunner := &InitCommand{}
	cobraCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize d3 in the current workspace",
		Long:  "Initialize d3 in the current workspace and create base project files",
		Args:  cobra.NoArgs,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not determine project root: %w", err)
			}
			cfg := NewConfig(projectRoot)

			// Create real services and real project instance here for production path
			fs := ports.RealFileSystem{}
			sessionSvc := session.NewStorage(cfg.D3Dir, fs)
			featureSvc := feature.NewService(cfg.WorkspaceRoot, cfg.FeaturesDir, cfg.D3Dir, fs)
			ruleGenerator := rules.NewRuleGenerator()
			rulesSvc := rules.NewService(cfg.WorkspaceRoot, cfg.CursorRulesDir, ruleGenerator, fs)
			phaseSvc := phase.NewService(fs)

			// Assign the real project instance to the command runner for production
			cmdRunner.projectSvc = project.New(cfg.WorkspaceRoot, fs, sessionSvc, featureSvc, rulesSvc, phaseSvc)

			return cmdRunner.run(cmdRunner.clean)
		},
	}
	cobraCmd.Flags().BoolVar(&cmdRunner.clean, "clean", false, "Perform a clean initialization (remove existing files)")
	return cobraCmd
}

// run is the core logic, now using the projectSvc field.
func (c *InitCommand) run(clean bool) error {
	// If projectSvc is nil (e.g. not set by test or RunE), it would panic.
	// This implies RunE should always set it, or tests should always set it.
	if c.projectSvc == nil {
		return fmt.Errorf("project service not initialized in InitCommand")
	}
	result, err := c.projectSvc.Init(clean)
	if err != nil {
		return err
	}
	fmt.Println(result.FormatCLI())
	return nil
}
