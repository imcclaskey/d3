package command

import (
	"github.com/spf13/cobra"
)

// NewFeatureCommand creates a new cobra command for feature-related operations.
func NewFeatureCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feature",
		Short: "Manage features (create, enter, exit, etc.)",
		Long:  `Provides subcommands to manage the lifecycle and context of features within the d3 project.`,
	}
	return cmd
}
