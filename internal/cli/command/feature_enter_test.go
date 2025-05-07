package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"

	// Import project service and mocks

	"github.com/imcclaskey/d3/internal/project"
	projectmocks "github.com/imcclaskey/d3/internal/project/mocks"
)

func TestNewFeatureEnterCommand(t *testing.T) {
	cmd := NewFeatureEnterCommand()

	if cmd.Use != "enter <name>" {
		t.Errorf("Expected Use to be 'enter <name>', got '%s'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Expected Short description to be non-empty")
	}

	// Test argument validation requires the correct command hierarchy
	rootCmd := &cobra.Command{Use: "d3"}
	featureCmd := NewFeatureCommand() // Assuming NewFeatureCommand() is accessible
	featureCmd.AddCommand(cmd)        // Add enter to feature
	rootCmd.AddCommand(featureCmd)    // Add feature to root

	_, err := executeCommand(rootCmd, "feature", "enter") // No args
	if err == nil || !strings.Contains(err.Error(), "accepts 1 arg(s), received 0") {
		t.Errorf("Command did not return correct error for missing args, got: %v", err)
	}

	_, err = executeCommand(rootCmd, "feature", "enter", "feat1", "feat2") // Too many args
	if err == nil || !strings.Contains(err.Error(), "accepts 1 arg(s), received 2") {
		t.Errorf("Command did not return correct error for too many args, got: %v", err)
	}
}

func TestFeatureEnterCommand_RunLogic(t *testing.T) {
	featureName := "my-test-feature"

	tests := []struct {
		name                string
		setupMockProjectSvc func(mockSvc *projectmocks.MockProjectService)
		wantErr             bool
		wantOutputContains  string
	}{
		{
			name: "successful enter",
			setupMockProjectSvc: func(mockSvc *projectmocks.MockProjectService) {
				// Expect EnterFeature to be called and return a *Result
				mockSvc.EXPECT().EnterFeature(gomock.Any(), featureName).
					Return(project.NewResultWithRulesChanged("Entered feature 'my-test-feature' in phase 'design'."), nil).Times(1)
			},
			wantErr: false,
			// Expect the CLI formatted output from the Result
			wantOutputContains: "Entered feature 'my-test-feature' in phase 'design'. Cursor rules have been updated.",
		},
		{
			name: "enter feature fails",
			setupMockProjectSvc: func(mockSvc *projectmocks.MockProjectService) {
				mockSvc.EXPECT().EnterFeature(gomock.Any(), featureName).
					Return(nil, fmt.Errorf("feature not found or invalid")).Times(1)
			},
			wantErr:            true,
			wantOutputContains: "feature not found or invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockProjectSvc := projectmocks.NewMockProjectService(ctrl)

			if tt.setupMockProjectSvc != nil {
				tt.setupMockProjectSvc(mockProjectSvc)
			}

			cmdInstance := &FeatureEnterCommand{
				featureName: featureName,
				projectSvc:  mockProjectSvc, // Inject mock
			}

			// Capture stdout
			originalStdout := os.Stdout
			rPipe, wPipe, _ := os.Pipe()
			os.Stdout = wPipe

			err := cmdInstance.run(context.Background())

			wPipe.Close()
			os.Stdout = originalStdout
			stdoutBuf := new(bytes.Buffer)
			stdoutBuf.ReadFrom(rPipe)
			output := stdoutBuf.String()
			rPipe.Close()

			if (err != nil) != tt.wantErr {
				t.Errorf("FeatureEnterCommand.run() error = %v, wantErr %v\nOutput:\n%s", err, tt.wantErr, output)
				return
			}

			if !tt.wantErr {
				if !strings.Contains(output, tt.wantOutputContains) {
					t.Errorf("FeatureEnterCommand.run() output = %q, want to contain %q", output, tt.wantOutputContains)
				}
			} else if err != nil {
				if !strings.Contains(err.Error(), tt.wantOutputContains) {
					t.Errorf("FeatureEnterCommand.run() error string = %q, want to contain %q", err.Error(), tt.wantOutputContains)
				}
			}
		})
	}
}
