package tools

import (
	"context"
	"fmt"

	"github.com/imcclaskey/d3/internal/core/session"
	"github.com/imcclaskey/d3/internal/project"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MoveTool defines the d3_phase_move tool
var MoveTool = mcp.NewTool("d3_phase_move",
	mcp.WithDescription("Move to a different phase in the current feature"),
	mcp.WithString("to",
		mcp.Required(),
		mcp.Description("Target phase to move to (define, design, deliver)"),
	),
)

// HandleMove returns a handler for the d3_phase_move tool
func HandleMove(proj *project.Project) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract phase parameter
		targetPhaseStr, ok := request.Params.Arguments["to"].(string)
		if !ok {
			return mcp.NewToolResultError("Target phase 'to' must be specified"), nil
		}

		if proj == nil {
			return mcp.NewToolResultError("Internal error: Project context is nil"), nil
		}

		// Parse the phase string to a Phase enum
		targetPhase, err := session.ParsePhase(targetPhaseStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid phase '%s': %v", targetPhaseStr, err)), nil
		}

		// Call project's ChangePhase function with the parsed phase
		result, err := proj.ChangePhase(ctx, targetPhase)
		if err != nil {
			if err.Error() == "no active feature" {
				return mcp.NewToolResultError("Cannot move phase: no active feature"), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("Failed to change phase: %v", err)), nil
		}

		// Return the formatted result
		return mcp.NewToolResultText(result.FormatMCP()), nil
	}
}
