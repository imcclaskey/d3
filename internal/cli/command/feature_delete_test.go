package command

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	featuremocks "github.com/imcclaskey/d3/internal/core/feature/mocks"
)

// helper function to simulate stdin for tests
func mockStdin(t *testing.T, input string) func() {
	t.Helper()
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdin mock: %v", err)
	}
	os.Stdin = r
	go func() {
		defer w.Close()
		_, err := io.WriteString(w, input+"\n") // Add newline as ReadString expects it
		if err != nil {
			t.Logf("error writing to stdin pipe: %v", err) // Log instead of Fatal in goroutine
		}
	}()
	return func() {
		os.Stdin = origStdin
		r.Close() // Close read end as well
	}
}

// helper function to capture stdout for tests
func captureStdout(t *testing.T) (*os.File, *os.File, func()) {
	t.Helper()
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout capture: %v", err)
	}
	os.Stdout = w
	return r, w, func() {
		os.Stdout = origStdout
	}
}

func TestFeatureDeleteCommand_runLogic(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name                string
		featureNameArg      string
		userInput           string
		setupMockFeatureSvc func(mockSvc *featuremocks.MockFeatureServicer, featureName string)
		wantErr             bool
		wantOutputContains  string
	}{
		{
			name:           "successful deletion with y confirmation",
			featureNameArg: "my-feature-to-delete",
			userInput:      "y",
			setupMockFeatureSvc: func(mockSvc *featuremocks.MockFeatureServicer, featureName string) {
				mockSvc.EXPECT().DeleteFeature(gomock.Any(), featureName).Return(false, nil).Times(1)
			},
			wantErr:            false,
			wantOutputContains: "Feature 'my-feature-to-delete' deleted successfully.",
		},
		{
			name:           "successful deletion with yes confirmation",
			featureNameArg: "another-feature",
			userInput:      "yes",
			setupMockFeatureSvc: func(mockSvc *featuremocks.MockFeatureServicer, featureName string) {
				mockSvc.EXPECT().DeleteFeature(gomock.Any(), featureName).Return(false, nil).Times(1)
			},
			wantErr:            false,
			wantOutputContains: "Feature 'another-feature' deleted successfully.",
		},
		{
			name:           "deletion cancelled with n",
			featureNameArg: "safe-feature",
			userInput:      "n",
			setupMockFeatureSvc: func(mockSvc *featuremocks.MockFeatureServicer, featureName string) {
				// DeleteFeature should not be called
			},
			wantErr:            false,
			wantOutputContains: "Feature deletion cancelled.",
		},
		{
			name:           "deletion cancelled with empty input",
			featureNameArg: "empty-input-feature",
			userInput:      "",
			setupMockFeatureSvc: func(mockSvc *featuremocks.MockFeatureServicer, featureName string) {
				// DeleteFeature should not be called
			},
			wantErr:            false,
			wantOutputContains: "Feature deletion cancelled.",
		},
		{
			name:           "feature service returns error on delete",
			featureNameArg: "error-prone-feature",
			userInput:      "y",
			setupMockFeatureSvc: func(mockSvc *featuremocks.MockFeatureServicer, featureName string) {
				mockSvc.EXPECT().DeleteFeature(gomock.Any(), featureName).Return(false, fmt.Errorf("internal service error")).Times(1)
			},
			wantErr:            true,                                                            // Error should be returned by runLogic
			wantOutputContains: "Are you sure you want to delete feature 'error-prone-feature'", // Prompt still shown
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFeatureSvc := featuremocks.NewMockFeatureServicer(ctrl)

			if tt.setupMockFeatureSvc != nil {
				tt.setupMockFeatureSvc(mockFeatureSvc, tt.featureNameArg)
			}

			cmdRunner := &featureDeleteCmdRunner{
				featureName: tt.featureNameArg,
				featureSvc:  mockFeatureSvc,
			}

			// Mock stdin
			cleanupStdin := mockStdin(t, tt.userInput)
			defer cleanupStdin()

			// Capture stdout
			rStdout, wStdout, cleanupStdout := captureStdout(t)
			defer cleanupStdout()

			err := cmdRunner.runLogic(ctx)

			wStdout.Close() // Close writer to allow reader to finish
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, rStdout) // Use io.Copy, ignore bytes written for simplicity in test
			rStdout.Close()
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("featureDeleteCmdRunner.runLogic() error = %v, wantErr %v\nOutput:\n%s", err, tt.wantErr, output)
			}

			if !strings.Contains(output, tt.wantOutputContains) {
				t.Errorf("featureDeleteCmdRunner.runLogic() output = %q, want to contain %q", output, tt.wantOutputContains)
			}

			// If an error was expected, also check if the error message contains what we want (if applicable)
			// For this test, the main check is presence of error and stdout. Detailed error string check can be added if needed.
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.featureNameArg) { // Basic check: error mentions the feature
					// t.Logf("For feature '%s', error was: %v", tt.featureNameArg, err)
				}
			}
		})
	}
}
