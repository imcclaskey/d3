// Package tools implements agentic tools for d3
package tools

import (
	"github.com/imcclaskey/d3/internal/core"
)

// ToolManager manages all d3 agentic tools
type ToolManager struct {
	services *core.Services
}

// NewToolManager creates a new tool manager
func NewToolManager(services *core.Services) *ToolManager {
	return &ToolManager{
		services: services,
	}
}
