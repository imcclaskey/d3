package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/d3/internal/project"
)

// FeatureCreateTool defines the d3_feature_create tool
var FeatureCreateTool = mcp.NewTool("d3_feature_create",
	mcp.WithDescription("Create a new feature and set it as the current context"),
	mcp.WithString("name",
		mcp.Required(),
		mcp.Description("Name of the feature to create"),
	),
)

// FeatureEnterTool defines the d3_feature_enter tool
var FeatureEnterTool = mcp.NewTool("d3_feature_enter",
	mcp.WithDescription("Enter a feature context, resuming its last known phase."),
	mcp.WithString("feature_name",
		mcp.Required(),
		mcp.Description("Name of the feature to enter"),
	),
)

// FeatureExitTool defines the d3_feature_exit tool
var FeatureExitTool = mcp.NewTool("d3_feature_exit",
	mcp.WithDescription("Exit the current feature context, clearing active feature state."),
	// No parameters required
)

// HandleFeatureCreate returns a handler for the d3_feature_create tool
// It now accepts project.ProjectService interface for testability.
func HandleFeatureCreate(proj project.ProjectService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract feature name
		featureName, ok := request.Params.Arguments["name"].(string)
		if !ok || featureName == "" {
			return mcp.NewToolResultError("Feature name 'name' is required"), nil
		}

		// Check if project is valid
		if proj == nil {
			return mcp.NewToolResultError("Internal error: Project context is nil"), nil
		}

		// Call the project method to create the feature
		result, err := proj.CreateFeature(ctx, featureName)
		if err != nil {
			if err == project.ErrNotInitialized {
				return mcp.NewToolResultError("Cannot create feature: project not initialized"), nil
			}
			// For other errors, return them as system errors to provide more detail to the MCP client if needed.
			return mcp.NewToolResultError(fmt.Sprintf("System error creating feature: %v", err)), nil
		}

		// Return the formatted result
		return mcp.NewToolResultText(result.FormatMCP()), nil
	}
}

// HandleFeatureEnter returns a handler for the d3_feature_enter tool
func HandleFeatureEnter(proj project.ProjectService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract feature name
		featureName, ok := request.Params.Arguments["feature_name"].(string)
		if !ok || featureName == "" {
			return mcp.NewToolResultError("Feature name 'feature_name' is required"), nil
		}

		if proj == nil {
			return mcp.NewToolResultError("Internal error: Project context is nil"), nil
		}

		// Call the project method to enter the feature
		// EnterFeature now returns *Result, error
		result, err := proj.EnterFeature(ctx, featureName)
		if err != nil {
			// Handle specific known errors if necessary, otherwise return generic error
			if err == project.ErrNotInitialized {
				return mcp.NewToolResultError("Cannot enter feature: project not initialized"), nil
			}
			// Pass through the error message from EnterFeature
			return mcp.NewToolResultError(fmt.Sprintf("System error entering feature: %v", err)), nil
		}

		// Format success result using Result.FormatMCP()
		// result is now *project.Result, which has FormatMCP method
		return mcp.NewToolResultText(result.FormatMCP()), nil
	}
}

// HandleFeatureExit returns a handler for the d3_feature_exit tool
func HandleFeatureExit(proj project.ProjectService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if proj == nil {
			return mcp.NewToolResultError("Internal error: Project context is nil"), nil
		}

		// Call the project method to exit the feature
		result, err := proj.ExitFeature(ctx)
		if err != nil {
			// Handle specific known errors if necessary
			if err == project.ErrNotInitialized {
				return mcp.NewToolResultError("Cannot exit feature: project not initialized"), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("System error exiting feature: %v", err)), nil
		}

		// ExitFeature returns a Result struct which has a FormatMCP method
		return mcp.NewToolResultText(result.FormatMCP()), nil
	}
}
