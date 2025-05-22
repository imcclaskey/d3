package command

import (
	"errors"
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

// InitCommand represents the init command implementation
type InitCommand struct {
	clean       bool
	refresh     bool
	customRules bool
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
			if cmdRunner.clean && cmdRunner.refresh {
				return errors.New("--clean and --refresh flags are mutually exclusive")
			}
			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not determine project root: %w", err)
			}
			cfg := NewConfig(projectRoot)

			fs := ports.RealFileSystem{}
			featureSvc := feature.NewService(cfg.WorkspaceRoot, cfg.FeaturesDir, cfg.D3Dir, fs)
			ruleGenerator := rules.NewRuleGenerator(cfg.WorkspaceRoot, fs)
			rulesSvc := rules.NewService(cfg.WorkspaceRoot, cfg.CursorRulesDir, ruleGenerator, fs)
			phaseSvc := phase.NewService(fs)
			fileOp := projectfiles.NewDefaultFileOperator()

			cmdRunner.projectSvc = project.New(cfg.WorkspaceRoot, fs, featureSvc, rulesSvc, phaseSvc, fileOp)

			return cmdRunner.run(cmdRunner.clean, cmdRunner.refresh, cmdRunner.customRules)
		},
	}
	cobraCmd.Flags().BoolVar(&cmdRunner.clean, "clean", false, "Perform a clean initialization (remove existing .d3 directory)")
	cobraCmd.Flags().BoolVar(&cmdRunner.refresh, "refresh", false, "Refresh an existing d3 environment, creating missing standard files/directories without data loss")
	cobraCmd.Flags().BoolVar(&cmdRunner.customRules, "custom-rules", false, "Create a directory for custom rule templates (.d3/rules/) and populate it with default templates")
	return cobraCmd
}

// run is the core logic.
func (c *InitCommand) run(clean bool, refresh bool, customRules bool) error {
	// If projectSvc is nil (e.g. not set by test or RunE), it would panic.
	// This implies RunE should always set it, or tests should always set it.
	if c.projectSvc == nil {
		return fmt.Errorf("project service not initialized in InitCommand")
	}
	// The ProjectService.Init method now handles the logic for clean and refresh.
	result, err := c.projectSvc.Init(clean, refresh, customRules) // Pass all flags
	if err != nil {
		return err
	}
	fmt.Println(result.FormatCLI())
	return nil
}
