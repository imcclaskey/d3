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

// Assuming executeCommand helper exists elsewhere in package command

func TestNewExitCommand(t *testing.T) {
	cmd := NewExitCommand()

	if cmd.Use != "exit" {
		t.Errorf("Expected Use to be 'exit', got '%s'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Expected Short description to be non-empty")
	}
	// Check Args - cobra.NoArgs is not directly verifiable by a field,
	// but we can test its behavior.

	// Test argument validation (requires root command context)
	rootCmd := &cobra.Command{Use: "d3"}
	rootCmd.AddCommand(cmd)

	_, err := executeCommand(rootCmd, "exit", "unexpected-arg") // Too many args
	if err == nil || !strings.Contains(err.Error(), "unknown command \"unexpected-arg\" for \"d3 exit\"") {
		// Cobra's default error message for unexpected args on a NoArgs command
		t.Errorf("Command did not return correct error for unexpected args, got: %v", err)
	}
}

func TestExitCommand_RunLogic(t *testing.T) {
	tests := []struct {
		name                string
		setupMockProjectSvc func(mockSvc *projectmocks.MockProjectService)
		wantErr             bool
		wantOutputContains  string
	}{
		{
			name: "successful exit",
			setupMockProjectSvc: func(mockSvc *projectmocks.MockProjectService) {
				mockSvc.EXPECT().ExitFeature(gomock.Any()).
					Return(project.NewResultWithRulesChanged("Exited feature 'old-feat'. No active feature. Cursor rules cleared."), nil).Times(1)
			},
			wantErr:            false,
			wantOutputContains: "Exited feature 'old-feat'. No active feature. Cursor rules cleared. Cursor rules have been updated.",
		},
		{
			name: "exit fails",
			setupMockProjectSvc: func(mockSvc *projectmocks.MockProjectService) {
				mockSvc.EXPECT().ExitFeature(gomock.Any()).
					Return(nil, fmt.Errorf("failed to save session during exit")).Times(1)
			},
			wantErr:            true,
			wantOutputContains: "failed to save session during exit",
		},
		{
			name: "exit when no feature active (no-op)",
			setupMockProjectSvc: func(mockSvc *projectmocks.MockProjectService) {
				mockSvc.EXPECT().ExitFeature(gomock.Any()).
					Return(project.NewResult("No active feature to exit."), nil).Times(1)
			},
			wantErr:            false,
			wantOutputContains: "No active feature to exit.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockProjectSvc := projectmocks.NewMockProjectService(ctrl)

			if tt.setupMockProjectSvc != nil {
				tt.setupMockProjectSvc(mockProjectSvc)
			}

			cmdInstance := &ExitCommand{
				projectSvc: mockProjectSvc, // Inject mock
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
				t.Errorf("ExitCommand.run() error = %v, wantErr %v\nOutput:\n%s", err, tt.wantErr, output)
				return
			}

			if !tt.wantErr {
				if !strings.Contains(output, tt.wantOutputContains) {
					t.Errorf("ExitCommand.run() output = %q, want to contain %q", output, tt.wantOutputContains)
				}
			} else if err != nil {
				if !strings.Contains(err.Error(), tt.wantOutputContains) {
					t.Errorf("ExitCommand.run() error string = %q, want to contain %q", err.Error(), tt.wantOutputContains)
				}
			}
		})
	}
}
