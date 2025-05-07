package tools

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/imcclaskey/d3/internal/core/session"
	"github.com/imcclaskey/d3/internal/project"
	projectmocks "github.com/imcclaskey/d3/internal/project/mocks"
	"github.com/imcclaskey/d3/internal/testutil"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestHandleMove(t *testing.T) {
	tests := []struct {
		name           string
		toolNameForReq string
		params         map[string]interface{}
		setupMockProj  func(mockProj *projectmocks.MockProjectService)
		wantResultText string
		wantIsErrorSet bool
		wantHandlerErr bool
	}{
		{
			name:           "successful move",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{"to": "design"},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				mockProj.EXPECT().ChangePhase(gomock.Any(), session.Design).
					Return(project.NewResult("Moved to design phase."), nil).Times(1)
			},
			wantResultText: "Moved to design phase.",
			wantIsErrorSet: false,
		},
		{
			name:           "missing 'to' parameter",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				// No call to project service expected
			},
			wantResultText: "Target phase 'to' must be specified",
			wantIsErrorSet: true,
		},
		{
			name:           "invalid 'to' phase value",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{"to": "invalidPhase"},
			setupMockProj:  func(mockProj *projectmocks.MockProjectService) {},
			wantResultText: "Invalid phase 'invalidPhase': invalid phase: invalidPhase",
			wantIsErrorSet: true,
		},
		{
			name:           "project ChangePhase returns ErrNoActiveFeature",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{"to": "define"},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				mockProj.EXPECT().ChangePhase(gomock.Any(), session.Define).
					Return(nil, project.ErrNoActiveFeature).Times(1)
			},
			wantResultText: "Cannot move phase: no active feature",
			wantIsErrorSet: true,
		},
		{
			name:           "project ChangePhase returns ErrNotInitialized",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{"to": "design"},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				mockProj.EXPECT().ChangePhase(gomock.Any(), session.Design).
					Return(nil, project.ErrNotInitialized).Times(1)
			},
			wantResultText: "Cannot move phase: project not initialized",
			wantIsErrorSet: true,
		},
		{
			name:           "project ChangePhase returns other error",
			toolNameForReq: "d3_phase_move",
			params:         map[string]interface{}{"to": "deliver"},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				mockProj.EXPECT().ChangePhase(gomock.Any(), session.Deliver).
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
			mockProjSvc := projectmocks.NewMockProjectService(ctrl)

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
		setupMockProj  func(mockProj *projectmocks.MockProjectService)
		wantResultText string
		wantIsErrorSet bool
		wantHandlerErr bool
	}{
		{
			name:           "successful feature creation",
			toolNameForReq: "d3_feature_create",
			params:         map[string]interface{}{"name": "test-feature"},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
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
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
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
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				// No calls expected
			},
			wantResultText: "Feature name 'name' is required",
			wantIsErrorSet: true,
		},
		{
			name:           "empty feature name",
			toolNameForReq: "d3_feature_create",
			params:         map[string]interface{}{"name": ""},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				// No calls expected
			},
			wantResultText: "Feature name 'name' is required",
			wantIsErrorSet: true,
		},
		{
			name:           "project not initialized",
			toolNameForReq: "d3_feature_create",
			params:         map[string]interface{}{"name": "test-feature"},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
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
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
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
			mockProjSvc := projectmocks.NewMockProjectService(ctrl)

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
		setupMockProj  func(mockProj *projectmocks.MockProjectService)
		wantResultText string
		wantIsErrorSet bool
		wantHandlerErr bool
	}{
		{
			name:           "successful init without clean flag",
			toolNameForReq: "d3_init",
			params:         map[string]interface{}{},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				mockProj.EXPECT().Init(false).
					Return(project.NewResult("Project initialized successfully."), nil).Times(1)
			},
			wantResultText: "Project initialized successfully.",
			wantIsErrorSet: false,
		},
		{
			name:           "successful init with clean flag",
			toolNameForReq: "d3_init",
			params:         map[string]interface{}{"clean": true},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				mockProj.EXPECT().Init(true).
					Return(project.NewResult("Project initialized with clean option."), nil).Times(1)
			},
			wantResultText: "Project initialized with clean option.",
			wantIsErrorSet: false,
		},
		{
			name:           "init with rules changed",
			toolNameForReq: "d3_init",
			params:         map[string]interface{}{"clean": false},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				mockProj.EXPECT().Init(false).
					Return(project.NewResultWithRulesChanged("Project initialized with rules."), nil).Times(1)
			},
			wantResultText: "Project initialized with rules. Cursor rules have changed. Stop your current behavior and await further instruction.",
			wantIsErrorSet: false,
		},
		{
			name:           "error during initialization",
			toolNameForReq: "d3_init",
			params:         map[string]interface{}{},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				mockProj.EXPECT().Init(false).
					Return(nil, fmt.Errorf("initialization failed")).Times(1)
			},
			wantResultText: "System error initializing project: initialization failed",
			wantIsErrorSet: true,
		},
		{
			name:           "project service is nil",
			toolNameForReq: "d3_init",
			params:         map[string]interface{}{},
			setupMockProj:  nil,
			wantResultText: "Internal error: Project context is nil",
			wantIsErrorSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockProjSvc := projectmocks.NewMockProjectService(ctrl)

			var handler server.ToolHandlerFunc
			if tt.setupMockProj == nil {
				handler = HandleInit(nil)
			} else {
				tt.setupMockProj(mockProjSvc)
				handler = HandleInit(mockProjSvc)
			}

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
		setupMockProj  func(mockProj *projectmocks.MockProjectService)
		wantResultText string // Expected text in the MCP result
		wantIsErrorSet bool   // Whether the MCP result IsError should be true
		wantHandlerErr bool   // Whether the handler function itself should return an error
	}{
		{
			name:           "successful feature enter",
			toolNameForReq: "d3_feature_enter",
			params:         map[string]interface{}{"feature_name": "existing-feature"},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
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
			setupMockProj:  func(mockProj *projectmocks.MockProjectService) {},
			wantResultText: "Feature name 'feature_name' is required",
			wantIsErrorSet: true,
		},
		{
			name:           "empty feature_name parameter",
			toolNameForReq: "d3_feature_enter",
			params:         map[string]interface{}{"feature_name": ""},
			setupMockProj:  func(mockProj *projectmocks.MockProjectService) {},
			wantResultText: "Feature name 'feature_name' is required",
			wantIsErrorSet: true,
		},
		{
			name:           "project not initialized",
			toolNameForReq: "d3_feature_enter",
			params:         map[string]interface{}{"feature_name": "some-feature"},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
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
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
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
			mockProjSvc := projectmocks.NewMockProjectService(ctrl)

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
		setupMockProj  func(mockProj *projectmocks.MockProjectService)
		wantResultText string // Expected text in the MCP result
		wantIsErrorSet bool   // Whether the MCP result IsError should be true
		wantHandlerErr bool   // Whether the handler function itself should return an error
	}{
		{
			name:           "successful feature exit",
			toolNameForReq: "d3_feature_exit",
			params:         nil, // No parameters
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
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
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
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
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
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
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
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
			mockProjSvc := projectmocks.NewMockProjectService(ctrl)

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
