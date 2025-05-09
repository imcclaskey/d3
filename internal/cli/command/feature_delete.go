package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/ports"
)

// featureDeleteCmdRunner holds the dependencies and logic for the feature delete command.
type featureDeleteCmdRunner struct {
	featureName string
	featureSvc  feature.FeatureServicer // Use the interface type
}

// NewFeatureDeleteCommand creates a new cobra command for deleting features.
func NewFeatureDeleteCommand() *cobra.Command {
	// cmdRunner instance is created here but its fields (featureSvc, featureName)
	// will be populated within RunE before calling its runLogic method.
	cmdRunner := &featureDeleteCmdRunner{}

	cmd := &cobra.Command{
		Use:   "delete [feature-name]",
		Short: "Delete a feature",
		Long:  `Delete a feature by removing its directory and all associated content. This action is permanent and cannot be undone.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdRunner.featureName = args[0]

			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not determine workspace root: %w", err)
			}
			cfg := NewConfig(projectRoot)
			fs := ports.RealFileSystem{}

			// Initialize the real service for actual command execution
			cmdRunner.featureSvc = feature.NewService(cfg.WorkspaceRoot, cfg.FeaturesDir, cfg.D3Dir, fs)

			return cmdRunner.runLogic(context.Background())
		},
	}
	return cmd
}

// runLogic contains the core logic for deleting a feature.
// This method is called by RunE and can be called directly in tests with a mock service.
func (c *featureDeleteCmdRunner) runLogic(ctx context.Context) error {
	// Confirmation prompt
	// For actual CLI execution, os.Stdin will be used.
	// For testing, os.Stdin can be mocked.
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Are you sure you want to delete feature '%s'? This action cannot be undone. [y/N]: ", c.featureName)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		fmt.Println("Feature deletion cancelled.")
		return nil
	}

	if c.featureSvc == nil {
		// This should ideally not happen if RunE populates it correctly
		return fmt.Errorf("feature service not initialized in featureDeleteCmdRunner")
	}

	activeContextCleared, err := c.featureSvc.DeleteFeature(ctx, c.featureName)
	if err != nil {
		// Error is simply returned to cobra, which will print it.
		return fmt.Errorf("failed to delete feature '%s': %w", c.featureName, err)
	}

	// If we reach here, deletion was successful.
	fmt.Printf("Feature '%s' deleted successfully.\n", c.featureName)
	if activeContextCleared {
		fmt.Println("The active feature context has been cleared.")
	}

	return nil
}
