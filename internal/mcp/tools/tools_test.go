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

func TestHandleCreate(t *testing.T) {
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
			toolNameForReq: "d3_create_feature",
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
			toolNameForReq: "d3_create_feature",
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
			toolNameForReq: "d3_create_feature",
			params:         map[string]interface{}{},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				// No calls expected
			},
			wantResultText: "Feature name 'name' is required",
			wantIsErrorSet: true,
		},
		{
			name:           "empty feature name",
			toolNameForReq: "d3_create_feature",
			params:         map[string]interface{}{"name": ""},
			setupMockProj: func(mockProj *projectmocks.MockProjectService) {
				// No calls expected
			},
			wantResultText: "Feature name 'name' is required",
			wantIsErrorSet: true,
		},
		{
			name:           "project not initialized",
			toolNameForReq: "d3_create_feature",
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
			toolNameForReq: "d3_create_feature",
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
			toolNameForReq: "d3_create_feature",
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
				handler = HandleCreate(nil)
			} else {
				tt.setupMockProj(mockProjSvc)
				handler = HandleCreate(mockProjSvc)
			}

			request := testutil.NewTestCallToolRequest(tt.toolNameForReq, tt.params)
			result, err := handler(context.Background(), request)

			if (err != nil) != tt.wantHandlerErr {
				t.Errorf("HandleCreate() handler error = %v, wantHandlerErr %v", err, tt.wantHandlerErr)
			}

			if result == nil {
				if !tt.wantHandlerErr {
					t.Fatal("HandleCreate() result is nil when no handler error was expected")
				}
				return
			}

			if result.IsError != tt.wantIsErrorSet {
				t.Errorf("HandleCreate() result.IsError = %v, wantIsErrorSet %v. Result: %+v", result.IsError, tt.wantIsErrorSet, result)
			}

			if len(result.Content) >= 1 {
				contentItem := result.Content[0]
				switch content := contentItem.(type) {
				case mcp.TextContent:
					if content.Text != tt.wantResultText {
						t.Errorf("HandleCreate() result.Content[0].(TextContent).Text = %q, wantResultText %q. Result: %+v",
							content.Text, tt.wantResultText, result)
					}
				default:
					t.Errorf("HandleCreate() result.Content[0] is not TextContent, got %T. Result: %+v",
						contentItem, result)
				}
			} else if tt.wantResultText != "" {
				t.Errorf("HandleCreate() result.Content length = %d, want >= 1 (for text %q). Full result: %+v",
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
