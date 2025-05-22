package command

import (
	"bytes"
	// "context" // No longer needed for these tests
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/imcclaskey/d3/internal/project" // For project.Result and error types
	"github.com/spf13/cobra"
)

// executeCommand is a helper to execute cobra commands and capture output/error
func executeCommand(cmd *cobra.Command, args ...string) (output string, err error) {
	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return b.String(), err
}

func TestInitCommand_RunLogic(t *testing.T) {
	tests := []struct {
		name                string
		cleanFlag           bool
		refreshFlag         bool
		customRulesFlag     bool
		setupMockProjectSvc func(mockSvc *project.MockProjectService)
		wantErr             bool
		wantOutputContains  string
	}{
		{
			name:            "successful init, no clean flag",
			cleanFlag:       false,
			refreshFlag:     false,
			customRulesFlag: false,
			setupMockProjectSvc: func(mockSvc *project.MockProjectService) {
				mockSvc.EXPECT().Init(false, false, false).Return(project.NewResultWithRulesChanged("Project initialized successfully."), nil).Times(1)
			},
			wantErr:            false,
			wantOutputContains: "Project initialized successfully. Cursor rules have been updated.",
		},
		{
			name:            "successful init, with clean flag",
			cleanFlag:       true,
			refreshFlag:     false,
			customRulesFlag: false,
			setupMockProjectSvc: func(mockSvc *project.MockProjectService) {
				mockSvc.EXPECT().Init(true, false, false).Return(project.NewResultWithRulesChanged("Project initialized successfully."), nil).Times(1)
			},
			wantErr:            false,
			wantOutputContains: "Project initialized successfully. Cursor rules have been updated.",
		},
		{
			name:            "successful init, with refresh flag",
			cleanFlag:       false,
			refreshFlag:     true,
			customRulesFlag: false,
			setupMockProjectSvc: func(mockSvc *project.MockProjectService) {
				mockSvc.EXPECT().Init(false, true, false).Return(project.NewResultWithRulesChanged("Project refreshed successfully."), nil).Times(1)
			},
			wantErr:            false,
			wantOutputContains: "Project refreshed successfully. Cursor rules have been updated.",
		},
		{
			name:            "successful init, with custom-rules flag",
			cleanFlag:       false,
			refreshFlag:     false,
			customRulesFlag: true,
			setupMockProjectSvc: func(mockSvc *project.MockProjectService) {
				mockSvc.EXPECT().Init(false, false, true).Return(project.NewResultWithRulesChanged("Project initialized successfully. Custom rules directory created and populated with default templates."), nil).Times(1)
			},
			wantErr:            false,
			wantOutputContains: "Project initialized successfully. Custom rules directory created and populated with default templates. Cursor rules have been updated.",
		},
		{
			name:            "init fails in projectSvc.Init",
			cleanFlag:       false,
			refreshFlag:     false,
			customRulesFlag: false,
			setupMockProjectSvc: func(mockSvc *project.MockProjectService) {
				mockSvc.EXPECT().Init(false, false, false).Return(nil, fmt.Errorf("project init failed")).Times(1)
			},
			wantErr:            true,
			wantOutputContains: "project init failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockProjectSvc := project.NewMockProjectService(ctrl)

			if tt.setupMockProjectSvc != nil {
				tt.setupMockProjectSvc(mockProjectSvc)
			}

			cmdInstance := &InitCommand{
				clean:       tt.cleanFlag,
				refresh:     tt.refreshFlag,
				customRules: tt.customRulesFlag,
				projectSvc:  mockProjectSvc,
			}

			var outBuf bytes.Buffer
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := cmdInstance.run(tt.cleanFlag, tt.refreshFlag, tt.customRulesFlag)

			w.Close()
			os.Stdout = originalStdout
			output := outBuf.String()

			capturedOutputBytes := make(chan []byte)
			go func() {
				buf := new(bytes.Buffer)
				buf.ReadFrom(r)
				capturedOutputBytes <- buf.Bytes()
			}()
			w.Close()
			os.Stdout = originalStdout
			output = string(<-capturedOutputBytes)
			r.Close()

			if (err != nil) != tt.wantErr {
				t.Errorf("InitCommand.run() error = %v, wantErr %v\nOutput:\n%s", err, tt.wantErr, output)
				return
			}

			if !tt.wantErr {
				if !strings.Contains(output, tt.wantOutputContains) {
					t.Errorf("InitCommand.run() output = %q, want to contain %q", output, tt.wantOutputContains)
				}
			} else if err != nil {
				if !strings.Contains(err.Error(), tt.wantOutputContains) {
					t.Errorf("InitCommand.run() error string = %q, want to contain %q", err.Error(), tt.wantOutputContains)
				}
			}
		})
	}
}

// TestInitCommand_FlagsMutualExclusivity tests that --clean and --refresh are mutually exclusive.
func TestInitCommand_FlagsMutualExclusivity(t *testing.T) {
	cmd := NewInitCommand() // This sets up the cobra command with its RunE

	// We need to simulate the RunE part of NewInitCommand as it contains the check.
	// To do this effectively without calling the full project service, we can temporarily
	// assign a dummy projectSvc to a temporary cmdRunner inside this test, or
	// directly test the RunE logic if cobra allows easy invocation of it.

	// For simplicity, let's use executeCommand which will trigger RunE.
	_, err := executeCommand(cmd, "--clean", "--refresh")
	if err == nil {
		t.Errorf("Expected an error when both --clean and --refresh are provided, but got nil")
	} else {
		expectedErr := "--clean and --refresh flags are mutually exclusive"
		if err.Error() != expectedErr {
			t.Errorf("Expected error message %q, got %q", expectedErr, err.Error())
		}
	}

	// Test that it works with only --clean
	// This will fail if projectSvc is not mocked, as it tries to run the full init.
	// This test is primarily for the flag parsing logic, not the full command run.
	// To fully test this, NewInitCommand would need to allow injecting a mock for RunE testing.
	// For now, this specific test focuses on the RunE check for mutual exclusivity.
}
