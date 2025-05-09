// Package mcp implements Model Context Protocol server functionality for d3
package mcp

import (
	"path/filepath"

	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/phase"
	"github.com/imcclaskey/d3/internal/core/ports"
	"github.com/imcclaskey/d3/internal/core/projectfiles"
	"github.com/imcclaskey/d3/internal/core/rules"
	"github.com/imcclaskey/d3/internal/core/session"
	"github.com/imcclaskey/d3/internal/mcp/tools"
	"github.com/imcclaskey/d3/internal/project"
	"github.com/imcclaskey/d3/internal/version"
)

// NewServer creates a new MCP server for d3
func NewServer(workspaceRoot string) *server.MCPServer {
	// Initialize services
	fs := ports.RealFileSystem{}

	d3Dir := filepath.Join(workspaceRoot, ".d3")
	featuresDir := filepath.Join(d3Dir, "features")
	cursorRulesDir := filepath.Join(workspaceRoot, ".cursor", "rules")

	sessionSvc := session.NewStorage(d3Dir, fs)
	featureSvc := feature.NewService(workspaceRoot, featuresDir, d3Dir, fs)
	ruleGenerator := rules.NewRuleGenerator()
	rulesSvc := rules.NewService(workspaceRoot, cursorRulesDir, ruleGenerator, fs)
	phaseSvc := phase.NewService(fs)
	fileOp := projectfiles.NewDefaultFileOperator()

	// Initialize real project instance. It implements ProjectService.
	proj := project.New(workspaceRoot, fs, sessionSvc, featureSvc, rulesSvc, phaseSvc, fileOp)

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"d3 - Define, Design, Deliver!",
		version.Version,
		server.WithInstructions("d3 is a structured workflow engine for AI-driven development within Cursor"),
		server.WithToolCapabilities(true),
	)

	// Register tools, proj (a *project.Project) satisfies project.ProjectService.
	tools.RegisterTools(mcpServer, proj)

	return mcpServer
}

// ServeStdio starts the MCP server over stdio
func ServeStdio(s *server.MCPServer) error {
	return server.ServeStdio(s)
}
