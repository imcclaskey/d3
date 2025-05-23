package tools

import (
	"context"
	"fmt"

	corephase "github.com/imcclaskey/d3/internal/core/phase"
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
func HandleMove(proj project.ProjectService) server.ToolHandlerFunc {
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
		targetPhase, err := parsePhaseString(targetPhaseStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid phase '%s': %v", targetPhaseStr, err)), nil
		}

		// Call project's ChangePhase function with the parsed phase
		result, err := proj.ChangePhase(ctx, targetPhase)
		if err != nil {
			// Specific error check based on common project errors
			if err == project.ErrNoActiveFeature {
				return mcp.NewToolResultError("Cannot move phase: no active feature"), nil
			} else if err == project.ErrNotInitialized {
				return mcp.NewToolResultError("Cannot move phase: project not initialized"), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("Failed to change phase: %v", err)), nil
		}

		// Return the formatted result
		return mcp.NewToolResultText(result.FormatMCP()), nil
	}
}

// parsePhaseString converts a string to a phase.Phase type.
func parsePhaseString(phaseStr string) (corephase.Phase, error) {
	switch phaseStr {
	case string(corephase.Define):
		return corephase.Define, nil
	case string(corephase.Design):
		return corephase.Design, nil
	case string(corephase.Deliver):
		return corephase.Deliver, nil
	default:
		return corephase.None, fmt.Errorf("invalid phase: %s", phaseStr)
	}
}
