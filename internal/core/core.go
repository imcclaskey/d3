// Package core provides access to all d3 core services
package core

import (
	"path/filepath"

	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/files"
	"github.com/imcclaskey/d3/internal/core/session"
)

// Services provides access to all d3 core services
type Services struct {
	Feature *feature.Service
	Files   *files.Service
	Session *session.Manager
}

// NewServices creates a new Services instance with all core services
func NewServices(workspaceRoot string) *Services {
	d3Dir := filepath.Join(workspaceRoot, ".d3")
	featuresDir := filepath.Join(d3Dir, "features")

	return &Services{
		Feature: feature.NewService(workspaceRoot, featuresDir),
		Files:   files.NewService(workspaceRoot),
		Session: session.NewManager(workspaceRoot),
	}
}
