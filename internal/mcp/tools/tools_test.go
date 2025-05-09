package tools

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/imcclaskey/d3/internal/core/phase"
	"github.com/imcclaskey/d3/internal/project"
	"github.com/imcclaskey/d3/internal/testutil"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestHandleMove(t *testing.T) {
	tests := []struct {
		name           string
		toolNameForReq string
		params         map[string]interface{}
		setupMockProj  func(mockProj *project.MockProjectService)
		wantResultText string
		wantIsErrorSet bool
		wantHandlerErr bool
	}{
		{
			name:           "successful move",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{"to": "design"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().ChangePhase(gomock.Any(), phase.Design).
					Return(project.NewResult("Moved to design phase."), nil).Times(1)
			},
			wantResultText: "Moved to design phase.",
			wantIsErrorSet: false,
		},
		{
			name:           "missing 'to' parameter",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{},
			setupMockProj: func(mockProj *project.MockProjectService) {
				// No call to project service expected
			},
			wantResultText: "Target phase 'to' must be specified",
			wantIsErrorSet: true,
		},
		{
			name:           "invalid 'to' phase value",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{"to": "invalidPhase"},
			setupMockProj:  func(mockProj *project.MockProjectService) {},
			wantResultText: "Invalid phase 'invalidPhase': invalid phase: invalidPhase",
			wantIsErrorSet: true,
		},
		{
			name:           "project ChangePhase returns ErrNoActiveFeature",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{"to": "define"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().ChangePhase(gomock.Any(), phase.Define).
					Return(nil, project.ErrNoActiveFeature).Times(1)
			},
			wantResultText: "Cannot move phase: no active feature",
			wantIsErrorSet: true,
		},
		{
			name:           "project ChangePhase returns ErrNotInitialized",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{"to": "design"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().ChangePhase(gomock.Any(), phase.Design).
					Return(nil, project.ErrNotInitialized).Times(1)
			},
			wantResultText: "Cannot move phase: project not initialized",
			wantIsErrorSet: true,
		},
		{
			name:           "project ChangePhase returns other error",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{"to": "deliver"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().ChangePhase(gomock.Any(), phase.Deliver).
					Return(nil, fmt.Errorf("random failure")).Times(1)
			},
			wantResultText: "Failed to change phase: random failure",
			wantIsErrorSet: true,
		},
		{
			name:           "project service is nil (simulating internal error)",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{"to": "design"},
			setupMockProj:  nil,
			wantResultText: "Internal error: Project context is nil",
			wantIsErrorSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockProjSvc := project.NewMockProjectService(ctrl)

			var handler server.ToolHandlerFunc
			if tt.setupMockProj == nil {
				handler = HandleMove(nil)
			} else {
				tt.setupMockProj(mockProjSvc)
				handler = HandleMove(mockProjSvc)
			}

			request := testutil.NewTestCallToolRequest(tt.toolNameForReq, tt.params)
			result, err := handler(context.Background(), request)

			if (err != nil) != tt.wantHandlerErr {
				t.Errorf("HandleMove() handler error = %v, wantHandlerErr %v", err, tt.wantHandlerErr)
			}

			if result == nil {
				if !tt.wantHandlerErr {
					t.Fatal("HandleMove() result is nil when no handler error was expected")
				}
				return
			}

			if result.IsError != tt.wantIsErrorSet {
				t.Errorf("HandleMove() result.IsError = %v, wantIsErrorSet %v. Result: %+v", result.IsError, tt.wantIsErrorSet, result)
			}

			if len(result.Content) >= 1 {
				contentItem := result.Content[0]
				switch content := contentItem.(type) {
				case mcp.TextContent:
					if content.Text != tt.wantResultText {
						t.Errorf("HandleMove() result.Content[0].(TextContent).Text = %q, wantResultText %q. Result: %+v",
							content.Text, tt.wantResultText, result)
					}
				default:
					t.Errorf("HandleMove() result.Content[0] is not TextContent, got %T. Result: %+v",
						contentItem, result)
				}
			} else if tt.wantResultText != "" {
				t.Errorf("HandleMove() result.Content length = %d, want >= 1 (for text %q). Full result: %+v",
					len(result.Content), tt.wantResultText, result)
			}
		})
	}
}

func TestHandleFeatureCreate(t *testing.T) {
	tests := []struct {
		name           string
		toolNameForReq string
		params         map[string]interface{}
		setupMockProj  func(mockProj *project.MockProjectService)
		wantResultText string
		wantIsErrorSet bool
		wantHandlerErr bool
	}{
		{
			name:           "successful feature creation",
			toolNameForReq: "d3_feature_create",
			params:         map[string]interface{}{"name": "test-feature"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().CreateFeature(gomock.Any(), "test-feature").
					Return(project.NewResult("Feature 'test-feature' created."), nil).Times(1)
			},
			wantResultText: "Feature 'test-feature' created.",
			wantIsErrorSet: false,
		},
		{
			name:           "feature creation with rules changed",
			toolNameForReq: "d3_feature_create",
			params:         map[string]interface{}{"name": "test-feature-rules"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().CreateFeature(gomock.Any(), "test-feature-rules").
					Return(project.NewResultWithRulesChanged("Feature created with rules"), nil).Times(1)
			},
			wantResultText: "Feature created with rules Cursor rules have changed. Stop your current behavior and await further instruction.",
			wantIsErrorSet: false,
		},
		{
			name:           "missing feature name",
			toolNameForReq: "d3_feature_create",
			params:         map[string]interface{}{},
			setupMockProj: func(mockProj *project.MockProjectService) {
				// No calls expected
			},
			wantResultText: "Feature name 'name' is required",
			wantIsErrorSet: true,
		},
		{
			name:           "empty feature name",
			toolNameForReq: "d3_feature_create",
			params:         map[string]interface{}{"name": ""},
			setupMockProj: func(mockProj *project.MockProjectService) {
				// No calls expected
			},
			wantResultText: "Feature name 'name' is required",
			wantIsErrorSet: true,
		},
		{
			name:           "project not initialized",
			toolNameForReq: "d3_feature_create",
			params:         map[string]interface{}{"name": "test-feature"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().CreateFeature(gomock.Any(), "test-feature").
					Return(nil, project.ErrNotInitialized).Times(1)
			},
			wantResultText: "Cannot create feature: project not initialized",
			wantIsErrorSet: true,
		},
		{
			name:           "other error from CreateFeature",
			toolNameForReq: "d3_feature_create",
			params:         map[string]interface{}{"name": "test-feature"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().CreateFeature(gomock.Any(), "test-feature").
					Return(nil, fmt.Errorf("feature creation failed")).Times(1)
			},
			wantResultText: "System error creating feature: feature creation failed",
			wantIsErrorSet: true,
		},
		{
			name:           "project service is nil",
			toolNameForReq: "d3_feature_create",
			params:         map[string]interface{}{"name": "test-feature"},
			setupMockProj:  nil,
			wantResultText: "Internal error: Project context is nil",
			wantIsErrorSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockProjSvc := project.NewMockProjectService(ctrl)

			var handler server.ToolHandlerFunc
			if tt.setupMockProj == nil {
				handler = HandleFeatureCreate(nil)
			} else {
				tt.setupMockProj(mockProjSvc)
				handler = HandleFeatureCreate(mockProjSvc)
			}

			request := testutil.NewTestCallToolRequest(tt.toolNameForReq, tt.params)
			result, err := handler(context.Background(), request)

			if (err != nil) != tt.wantHandlerErr {
				t.Errorf("HandleFeatureCreate() handler error = %v, wantHandlerErr %v", err, tt.wantHandlerErr)
			}

			if result == nil {
				if !tt.wantHandlerErr {
					t.Fatal("HandleFeatureCreate() result is nil when no handler error was expected")
				}
				return
			}

			if result.IsError != tt.wantIsErrorSet {
				t.Errorf("HandleFeatureCreate() result.IsError = %v, wantIsErrorSet %v. Result: %+v", result.IsError, tt.wantIsErrorSet, result)
			}

			if len(result.Content) >= 1 {
				contentItem := result.Content[0]
				switch content := contentItem.(type) {
				case mcp.TextContent:
					if content.Text != tt.wantResultText {
						t.Errorf("HandleFeatureCreate() result.Content[0].(TextContent).Text = %q, wantResultText %q. Result: %+v",
							content.Text, tt.wantResultText, result)
					}
				default:
					t.Errorf("HandleFeatureCreate() result.Content[0] is not TextContent, got %T. Result: %+v",
						contentItem, result)
				}
			} else if tt.wantResultText != "" {
				t.Errorf("HandleFeatureCreate() result.Content length = %d, want >= 1 (for text %q). Full result: %+v",
					len(result.Content), tt.wantResultText, result)
			}
		})
	}
}

func TestHandleInit(t *testing.T) {
	tests := []struct {
		name           string
		toolNameForReq string
		params         map[string]interface{}
		// setupMockProj  func(mockProj *project.MockProjectService) // No longer needed as proj is not used
		wantResultText string
		wantIsErrorSet bool
		wantHandlerErr bool
	}{
		{
			name:           "returns guidance message to use CLI",
			toolNameForReq: "d3_init",
			params:         map[string]interface{}{}, // Params like 'clean' are now ignored by the handler
			// setupMockProj: func(mockProj *project.MockProjectService) {}, // No mock setup needed
			wantResultText: "To initialize d3 in your project, please run the `d3 init` command in your terminal. You can use flags like `--clean` or `--refresh` as needed. For example: `d3 init --refresh`",
			wantIsErrorSet: false,
			wantHandlerErr: false,
		},
		// Other previous test cases for different init scenarios are removed
		// as the tool now has a single behavior: guide to CLI.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ctrl := gomock.NewController(t) // No longer needed as no mocks are used
			// mockProjSvc := project.NewMockProjectService(ctrl) // ProjSvc is not used by the handler anymore

			// The handler no longer uses the project service, so passing nil is acceptable for the test.
			// If it were still used, we'd pass mockProjSvc.
			handler := HandleInit(nil) // proj project.ProjectService is not used

			request := testutil.NewTestCallToolRequest(tt.toolNameForReq, tt.params)
			result, err := handler(context.Background(), request)

			if (err != nil) != tt.wantHandlerErr {
				t.Errorf("HandleInit() handler error = %v, wantHandlerErr %v", err, tt.wantHandlerErr)
			}

			if result == nil {
				if !tt.wantHandlerErr {
					t.Fatal("HandleInit() result is nil when no handler error was expected")
				}
				return
			}

			if result.IsError != tt.wantIsErrorSet {
				t.Errorf("HandleInit() result.IsError = %v, wantIsErrorSet %v. Result: %+v", result.IsError, tt.wantIsErrorSet, result)
			}

			if len(result.Content) >= 1 {
				contentItem := result.Content[0]
				switch content := contentItem.(type) {
				case mcp.TextContent:
					if content.Text != tt.wantResultText {
						t.Errorf("HandleInit() result.Content[0].(TextContent).Text = %q, wantResultText %q. Result: %+v",
							content.Text, tt.wantResultText, result)
					}
				default:
					t.Errorf("HandleInit() result.Content[0] is not TextContent, got %T. Result: %+v",
						contentItem, result)
				}
			} else if tt.wantResultText != "" {
				t.Errorf("HandleInit() result.Content length = %d, want >= 1 (for text %q). Full result: %+v",
					len(result.Content), tt.wantResultText, result)
			}
		})
	}
}

// --- Tests for Feature Enter/Exit Handlers ---

func TestHandleFeatureEnter(t *testing.T) {
	tests := []struct {
		name           string
		toolNameForReq string
		params         map[string]interface{}
		setupMockProj  func(mockProj *project.MockProjectService)
		wantResultText string // Expected text in the MCP result
		wantIsErrorSet bool   // Whether the MCP result IsError should be true
		wantHandlerErr bool   // Whether the handler function itself should return an error
	}{
		{
			name:           "successful feature enter",
			toolNameForReq: "d3_feature_enter",
			params:         map[string]interface{}{"feature_name": "existing-feature"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().EnterFeature(gomock.Any(), "existing-feature").
					Return(project.NewResultWithRulesChanged("Entered feature 'existing-feature' in phase 'design'."), nil).Times(1)
			},
			wantResultText: "Entered feature 'existing-feature' in phase 'design'. Cursor rules have changed. Stop your current behavior and await further instruction.",
			wantIsErrorSet: false,
		},
		{
			name:           "missing feature_name parameter",
			toolNameForReq: "d3_feature_enter",
			params:         map[string]interface{}{}, // Missing 'feature_name'
			setupMockProj:  func(mockProj *project.MockProjectService) {},
			wantResultText: "Feature name 'feature_name' is required",
			wantIsErrorSet: true,
		},
		{
			name:           "empty feature_name parameter",
			toolNameForReq: "d3_feature_enter",
			params:         map[string]interface{}{"feature_name": ""},
			setupMockProj:  func(mockProj *project.MockProjectService) {},
			wantResultText: "Feature name 'feature_name' is required",
			wantIsErrorSet: true,
		},
		{
			name:           "project not initialized",
			toolNameForReq: "d3_feature_enter",
			params:         map[string]interface{}{"feature_name": "some-feature"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().EnterFeature(gomock.Any(), "some-feature").
					Return(nil, project.ErrNotInitialized).Times(1)
			},
			wantResultText: "Cannot enter feature: project not initialized",
			wantIsErrorSet: true,
		},
		{
			name:           "other error from EnterFeature",
			toolNameForReq: "d3_feature_enter",
			params:         map[string]interface{}{"feature_name": "error-feature"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().EnterFeature(gomock.Any(), "error-feature").
					Return(nil, fmt.Errorf("get phase failed")).Times(1)
			},
			wantResultText: "System error entering feature: get phase failed",
			wantIsErrorSet: true,
		},
		{
			name:           "project service is nil",
			toolNameForReq: "d3_feature_enter",
			params:         map[string]interface{}{"feature_name": "any-feature"},
			setupMockProj:  nil,
			wantResultText: "Internal error: Project context is nil",
			wantIsErrorSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockProjSvc := project.NewMockProjectService(ctrl)

			var handler server.ToolHandlerFunc
			if tt.setupMockProj == nil {
				handler = HandleFeatureEnter(nil)
			} else {
				tt.setupMockProj(mockProjSvc)
				handler = HandleFeatureEnter(mockProjSvc)
			}

			request := testutil.NewTestCallToolRequest(tt.toolNameForReq, tt.params)
			result, err := handler(context.Background(), request)

			if (err != nil) != tt.wantHandlerErr {
				t.Errorf("HandleFeatureEnter() handler error = %v, wantHandlerErr %v", err, tt.wantHandlerErr)
			}

			if result == nil {
				if !tt.wantHandlerErr {
					t.Fatal("HandleFeatureEnter() result is nil when no handler error was expected")
				}
				return
			}

			if result.IsError != tt.wantIsErrorSet {
				t.Errorf("HandleFeatureEnter() result.IsError = %v, wantIsErrorSet %v. Result: %+v", result.IsError, tt.wantIsErrorSet, result)
			}

			if len(result.Content) >= 1 {
				contentItem := result.Content[0]
				switch content := contentItem.(type) {
				case mcp.TextContent:
					if content.Text != tt.wantResultText {
						t.Errorf("HandleFeatureEnter() result text = %q, want %q", content.Text, tt.wantResultText)
					}
				default:
					t.Errorf("HandleFeatureEnter() result content is not TextContent, got %T", contentItem)
				}
			} else if tt.wantResultText != "" {
				t.Errorf("HandleFeatureEnter() result has no content, want text %q", tt.wantResultText)
			}
		})
	}
}

func TestHandleFeatureExit(t *testing.T) {
	tests := []struct {
		name           string
		toolNameForReq string
		params         map[string]interface{} // Should be empty for exit
		setupMockProj  func(mockProj *project.MockProjectService)
		wantResultText string // Expected text in the MCP result
		wantIsErrorSet bool   // Whether the MCP result IsError should be true
		wantHandlerErr bool   // Whether the handler function itself should return an error
	}{
		{
			name:           "successful feature exit",
			toolNameForReq: "d3_feature_exit",
			params:         nil, // No parameters
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().ExitFeature(gomock.Any()).
					Return(project.NewResultWithRulesChanged("Exited feature 'old-feature'. No active feature. Cursor rules cleared."), nil).Times(1)
			},
			// Note: The FormatMCP adds the specific MCP suffix
			wantResultText: "Exited feature 'old-feature'. No active feature. Cursor rules cleared. Cursor rules have changed. Stop your current behavior and await further instruction.",
			wantIsErrorSet: false,
		},
		{
			name:           "exit when no active feature (no-op success)",
			toolNameForReq: "d3_feature_exit",
			params:         nil,
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().ExitFeature(gomock.Any()).
					Return(project.NewResult("No active feature to exit."), nil).Times(1)
			},
			wantResultText: "No active feature to exit.", // FormatMCP doesn't add suffix if RulesChanged=false
			wantIsErrorSet: false,
		},
		{
			name:           "project not initialized",
			toolNameForReq: "d3_feature_exit",
			params:         nil,
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().ExitFeature(gomock.Any()).
					Return(nil, project.ErrNotInitialized).Times(1)
			},
			wantResultText: "Cannot exit feature: project not initialized",
			wantIsErrorSet: true,
		},
		{
			name:           "other error from ExitFeature",
			toolNameForReq: "d3_feature_exit",
			params:         nil,
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().ExitFeature(gomock.Any()).
					Return(nil, fmt.Errorf("session save failed during exit")).Times(1)
			},
			wantResultText: "System error exiting feature: session save failed during exit",
			wantIsErrorSet: true,
		},
		{
			name:           "project service is nil",
			toolNameForReq: "d3_feature_exit",
			params:         nil,
			setupMockProj:  nil,
			wantResultText: "Internal error: Project context is nil",
			wantIsErrorSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockProjSvc := project.NewMockProjectService(ctrl)

			var handler server.ToolHandlerFunc
			if tt.setupMockProj == nil {
				handler = HandleFeatureExit(nil)
			} else {
				tt.setupMockProj(mockProjSvc)
				handler = HandleFeatureExit(mockProjSvc)
			}

			request := testutil.NewTestCallToolRequest(tt.toolNameForReq, tt.params) // params are nil
			result, err := handler(context.Background(), request)

			if (err != nil) != tt.wantHandlerErr {
				t.Errorf("HandleFeatureExit() handler error = %v, wantHandlerErr %v", err, tt.wantHandlerErr)
			}

			if result == nil {
				if !tt.wantHandlerErr {
					t.Fatal("HandleFeatureExit() result is nil when no handler error was expected")
				}
				return
			}

			if result.IsError != tt.wantIsErrorSet {
				t.Errorf("HandleFeatureExit() result.IsError = %v, wantIsErrorSet %v. Result: %+v", result.IsError, tt.wantIsErrorSet, result)
			}

			if len(result.Content) >= 1 {
				contentItem := result.Content[0]
				switch content := contentItem.(type) {
				case mcp.TextContent:
					if content.Text != tt.wantResultText {
						t.Errorf("HandleFeatureExit() result text = %q, want %q", content.Text, tt.wantResultText)
					}
				default:
					t.Errorf("HandleFeatureExit() result content is not TextContent, got %T", contentItem)
				}
			} else if tt.wantResultText != "" {
				t.Errorf("HandleFeatureExit() result has no content, want text %q", tt.wantResultText)
			}
		})
	}
}

func TestHandleFeatureDelete(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name           string
		params         map[string]interface{}
		setupMockProj  func(mockProj *project.MockProjectService)
		wantResultText string
		wantIsErrorSet bool
		wantHandlerErr bool // if the handler itself returns an error, not just result.IsError
	}{
		{
			name:   "successful feature deletion",
			params: map[string]interface{}{"feature_name": "test-feature-to-delete", "confirm": true},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().DeleteFeature(ctx, "test-feature-to-delete").
					Return(project.NewResult("Feature 'test-feature-to-delete' deleted successfully."), nil).Times(1)
			},
			wantResultText: "Feature 'test-feature-to-delete' deleted successfully.",
			wantIsErrorSet: false,
		},
		{
			name:   "successful deletion of active feature with rules changed",
			params: map[string]interface{}{"feature_name": "active-feature-deleted", "confirm": true},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().DeleteFeature(ctx, "active-feature-deleted").
					Return(project.NewResultWithRulesChanged("Feature 'active-feature-deleted' deleted successfully. Active feature context has been cleared."), nil).Times(1)
			},
			wantResultText: "Feature 'active-feature-deleted' deleted successfully. Active feature context has been cleared. Cursor rules have changed. Stop your current behavior and await further instruction.",
			wantIsErrorSet: false,
		},
		{
			name:   "deletion requires confirmation (confirm missing)",
			params: map[string]interface{}{"feature_name": "test-feature-confirm-missing"},
			setupMockProj: func(mockProj *project.MockProjectService) {
				// DeleteFeature should not be called
			},
			wantResultText: "Are you sure you want to delete feature 'test-feature-confirm-missing'? This action cannot be undone. Please call again with confirm=true.",
			wantIsErrorSet: true,
		},
		{
			name:   "deletion requires confirmation (confirm false)",
			params: map[string]interface{}{"feature_name": "test-feature-confirm-false", "confirm": false},
			setupMockProj: func(mockProj *project.MockProjectService) {
				// DeleteFeature should not be called
			},
			wantResultText: "Are you sure you want to delete feature 'test-feature-confirm-false'? This action cannot be undone. Please call again with confirm=true.",
			wantIsErrorSet: true,
		},
		{
			name:           "missing feature_name parameter",
			params:         map[string]interface{}{"confirm": true},
			setupMockProj:  func(mockProj *project.MockProjectService) { /* No call expected */ },
			wantResultText: "Feature name 'feature_name' is required",
			wantIsErrorSet: true,
		},
		{
			name:           "empty feature_name parameter",
			params:         map[string]interface{}{"feature_name": "", "confirm": true},
			setupMockProj:  func(mockProj *project.MockProjectService) { /* No call expected */ },
			wantResultText: "Feature name 'feature_name' is required",
			wantIsErrorSet: true,
		},
		{
			name:   "project not initialized",
			params: map[string]interface{}{"feature_name": "any-feature", "confirm": true},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().DeleteFeature(ctx, "any-feature").
					Return(nil, project.ErrNotInitialized).Times(1)
			},
			wantResultText: "Cannot delete feature: project not initialized",
			wantIsErrorSet: true,
		},
		{
			name:   "project service returns error (e.g., feature not found)",
			params: map[string]interface{}{"feature_name": "non-existent-feature", "confirm": true},
			setupMockProj: func(mockProj *project.MockProjectService) {
				mockProj.EXPECT().DeleteFeature(ctx, "non-existent-feature").
					Return(nil, fmt.Errorf("feature '%s' not found", "non-existent-feature")).Times(1)
			},
			wantResultText: "System error deleting feature 'non-existent-feature': feature 'non-existent-feature' not found",
			wantIsErrorSet: true,
		},
		{
			name:           "project service is nil",
			params:         map[string]interface{}{"feature_name": "any-feature", "confirm": true},
			setupMockProj:  nil,
			wantResultText: "Internal error: Project context is nil",
			wantIsErrorSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockProjSvc := project.NewMockProjectService(ctrl)

			var handler server.ToolHandlerFunc
			if tt.setupMockProj == nil {
				handler = HandleFeatureDelete(nil)
			} else {
				tt.setupMockProj(mockProjSvc)
				handler = HandleFeatureDelete(mockProjSvc)
			}

			request := testutil.NewTestCallToolRequest("d3_feature_delete", tt.params)
			result, err := handler(ctx, request)

			if (err != nil) != tt.wantHandlerErr {
				t.Errorf("HandleFeatureDelete() handler error = %v, wantHandlerErr %v", err, tt.wantHandlerErr)
			}

			if result == nil {
				if !tt.wantHandlerErr {
					t.Fatal("HandleFeatureDelete() result is nil when no handler error was expected")
				}
				return
			}

			if result.IsError != tt.wantIsErrorSet {
				t.Errorf("HandleFeatureDelete() result.IsError = %v, wantIsErrorSet %v. Result: %+v", result.IsError, tt.wantIsErrorSet, result)
			}

			if len(result.Content) == 1 {
				contentItem := result.Content[0]
				if textContent, ok := contentItem.(mcp.TextContent); ok {
					if textContent.Text != tt.wantResultText {
						t.Errorf("HandleFeatureDelete() result text = %q, want %q", textContent.Text, tt.wantResultText)
					}
				} else {
					t.Errorf("HandleFeatureDelete() result content is not mcp.TextContent, got %T", contentItem)
				}
			} else if tt.wantResultText != "" {
				t.Errorf("HandleFeatureDelete() result has %d content items, want 1 for text %q. Full result: %+v", len(result.Content), tt.wantResultText, result)
			}
		})
	}
}
