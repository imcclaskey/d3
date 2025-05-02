// Package tools implements agentic tools for i3
package tools

import (
	"github.com/imcclaskey/i3/internal/core"
)

// ToolManager manages all i3 agentic tools
type ToolManager struct {
	services *core.Services
}

// NewToolManager creates a new tool manager
func NewToolManager(services *core.Services) *ToolManager {
	return &ToolManager{
		services: services,
	}
}
