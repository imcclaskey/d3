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
		setupMockProjectSvc func(mockSvc *project.MockProjectService)
		wantErr             bool
		wantOutputContains  string
	}{
		{
			name:      "successful init, no clean flag",
			cleanFlag: false,
			setupMockProjectSvc: func(mockSvc *project.MockProjectService) {
				mockSvc.EXPECT().Init(false).Return(project.NewResultWithRulesChanged("Project initialized successfully."), nil).Times(1)
			},
			wantErr:            false,
			wantOutputContains: "Project initialized successfully. Cursor rules have been updated.",
		},
		{
			name:      "successful init, with clean flag",
			cleanFlag: true,
			setupMockProjectSvc: func(mockSvc *project.MockProjectService) {
				mockSvc.EXPECT().Init(true).Return(project.NewResultWithRulesChanged("Project initialized successfully."), nil).Times(1)
			},
			wantErr:            false,
			wantOutputContains: "Project initialized successfully. Cursor rules have been updated.",
		},
		{
			name:      "init fails in projectSvc.Init",
			cleanFlag: false,
			setupMockProjectSvc: func(mockSvc *project.MockProjectService) {
				mockSvc.EXPECT().Init(false).Return(nil, fmt.Errorf("project init failed")).Times(1)
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
				clean:      tt.cleanFlag,
				projectSvc: mockProjectSvc,
			}

			var outBuf bytes.Buffer
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := cmdInstance.run(tt.cleanFlag)

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
