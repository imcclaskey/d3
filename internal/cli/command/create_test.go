package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/imcclaskey/d3/internal/project"
)

// executeCommand helper can still be used if we want to test the full Cobra command,
// but for unit testing the logic inside 'run', we call it directly.

func TestFeatureCreateCommand_RunLogic(t *testing.T) {
	tests := []struct {
		name                string
		featureNameArg      string // Argument to the command
		setupMockProjectSvc func(mockSvc *project.MockProjectService, featureName string)
		wantErr             bool
		wantOutputContains  string
	}{
		{
			name:           "successful feature creation",
			featureNameArg: "my-new-feature",
			setupMockProjectSvc: func(mockSvc *project.MockProjectService, featureName string) {
				mockSvc.EXPECT().CreateFeature(gomock.Any(), featureName).Return(project.NewResultWithRulesChanged("Feature '"+featureName+"' created and set as the current context."), nil).Times(1)
			},
			wantErr:            false,
			wantOutputContains: "Feature 'my-new-feature' created and set as the current context. Cursor rules have been updated.",
		},
		{
			name:           "project not initialized - projectSvc.CreateFeature returns error",
			featureNameArg: "another-feature",
			setupMockProjectSvc: func(mockSvc *project.MockProjectService, featureName string) {
				mockSvc.EXPECT().CreateFeature(gomock.Any(), featureName).Return(nil, project.ErrNotInitialized).Times(1)
			},
			wantErr:            true,
			wantOutputContains: project.ErrNotInitialized.Error(),
		},
		{
			name:           "feature creation fails with generic error",
			featureNameArg: "fail-feature",
			setupMockProjectSvc: func(mockSvc *project.MockProjectService, featureName string) {
				mockSvc.EXPECT().CreateFeature(gomock.Any(), featureName).Return(nil, fmt.Errorf("internal create error")).Times(1)
			},
			wantErr:            true,
			wantOutputContains: "internal create error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockProjectSvc := project.NewMockProjectService(ctrl)

			if tt.setupMockProjectSvc != nil {
				tt.setupMockProjectSvc(mockProjectSvc, tt.featureNameArg)
			}

			cmdInstance := &FeatureCreateCommand{
				featureName: tt.featureNameArg, // Set the featureName that RunE would set
				projectSvc:  mockProjectSvc,    // Inject the mock service
			}

			originalStdout := os.Stdout
			rPipe, wPipe, _ := os.Pipe()
			os.Stdout = wPipe

			err := cmdInstance.run(context.Background())

			wPipe.Close()              // Close writer to signal EOF to reader goroutine
			os.Stdout = originalStdout // Restore stdout early

			capturedOutputBytes := make(chan []byte)
			go func() {
				buf := new(bytes.Buffer)
				buf.ReadFrom(rPipe) // Read from the read-end of the pipe
				capturedOutputBytes <- buf.Bytes()
				rPipe.Close() // Close reader in the goroutine after reading
			}()
			output := string(<-capturedOutputBytes)
			// rPipe.Close() // Already closed in goroutine

			if (err != nil) != tt.wantErr {
				t.Errorf("FeatureCreateCommand.run() error = %v, wantErr %v\nOutput:\n%s", err, tt.wantErr, output)
				return // Return if error presence is not as expected
			}

			if !tt.wantErr { // Only check stdout if no error was wanted
				if !strings.Contains(output, tt.wantOutputContains) {
					t.Errorf("FeatureCreateCommand.run() output = %q, want to contain %q", output, tt.wantOutputContains)
				}
			} else if err != nil { // If an error was wanted and we got one, check its content
				if !strings.Contains(err.Error(), tt.wantOutputContains) {
					t.Errorf("FeatureCreateCommand.run() error string = %q, want to contain %q", err.Error(), tt.wantOutputContains)
				}
			}
		})
	}
}
