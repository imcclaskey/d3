package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/phase"
	"github.com/imcclaskey/d3/internal/core/ports"
	"github.com/imcclaskey/d3/internal/core/projectfiles"
	"github.com/imcclaskey/d3/internal/core/rules"
	"github.com/imcclaskey/d3/internal/project"
)

// PhaseCommand represents the phase command implementation
type PhaseCommand struct{}

// NewPhaseCommand creates a new cobra command for the phase functionality
func NewPhaseCommand() *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "phase",
		Short: "Move between phases within a feature",
		Long:  "Move between phases (define, design, deliver) within a feature",
	}

	// Add move subcommand
	cobraCmd.AddCommand(NewPhaseMoveCommand())

	return cobraCmd
}

// PhaseMoveCommand represents the phase move command implementation
type PhaseMoveCommand struct {
	projectSvc project.ProjectService
}

// NewPhaseMoveCommand creates a new cobra command for the phase move functionality
func NewPhaseMoveCommand() *cobra.Command {
	cmdRunner := &PhaseMoveCommand{}
	cobraCmd := &cobra.Command{
		Use:   "move <phase>",
		Short: "Move to a different phase",
		Long:  "Move the current feature to a different phase (define, design, deliver)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			targetPhaseStr := strings.ToLower(args[0])

			// Convert string to phase enum
			var targetPhase phase.Phase
			switch targetPhaseStr {
			case "define":
				targetPhase = phase.Define
			case "design":
				targetPhase = phase.Design
			case "deliver":
				targetPhase = phase.Deliver
			default:
				return fmt.Errorf("invalid phase: %s (valid phases are: define, design, deliver)", targetPhaseStr)
			}

			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not determine project root: %w", err)
			}
			cfg := NewConfig(projectRoot)

			fs := ports.RealFileSystem{}
			featureSvc := feature.NewService(cfg.WorkspaceRoot, cfg.FeaturesDir, cfg.D3Dir, fs)
			ruleGenerator := rules.NewRuleGenerator()
			rulesSvc := rules.NewService(cfg.WorkspaceRoot, cfg.CursorRulesDir, ruleGenerator, fs)
			phaseSvc := phase.NewService(fs)
			fileOp := projectfiles.NewDefaultFileOperator()

			cmdRunner.projectSvc = project.New(cfg.WorkspaceRoot, fs, featureSvc, rulesSvc, phaseSvc, fileOp)

			return cmdRunner.run(targetPhase)
		},
	}

	return cobraCmd
}

// run is the core logic for phase move
func (c *PhaseMoveCommand) run(targetPhase phase.Phase) error {
	if c.projectSvc == nil {
		return fmt.Errorf("project service not initialized in PhaseMoveCommand")
	}

	ctx := context.Background()
	result, err := c.projectSvc.ChangePhase(ctx, targetPhase)
	if err != nil {
		if errors.Is(err, project.ErrNoActiveFeature) {
			return fmt.Errorf("no active feature. Run 'd3 feature enter <feature-name>' first")
		}
		return err
	}

	fmt.Println(result.FormatCLI())
	return nil
}
