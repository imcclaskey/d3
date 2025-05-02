// Package common provides common utilities for i3
package common

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetWorkspaceRoot gets the workspace root directory
// It returns an error if the current working directory cannot be determined.
func GetWorkspaceRoot() (string, error) {
	// First check for environment variable
	if envPath := os.Getenv("I3_WORKSPACE_PATH"); envPath != "" {
		return filepath.Clean(envPath), nil
	}

	// Get the current directory
	cwd, err := os.Getwd()
	if err != nil {
		// Return error instead of falling back
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Clean the path to ensure it's in canonical form with no extra slashes
	workspaceRoot := filepath.Clean(cwd)

	// Ensure we never return root directory
	if workspaceRoot == "/" {
		fmt.Fprintf(os.Stderr, "Warning: Root directory detected as workspace root, using safer alternative\n")

		// Since we already checked for environment variable above, try home directory
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			// If home dir fails, use temp dir as last resort
			tmpDir := os.TempDir()
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v. Using temp dir: %s\n", homeErr, tmpDir)
			return filepath.Join(tmpDir, "i3-workspace"), nil // Return path, no error
		}
		return filepath.Join(homeDir, "i3-workspace"), nil // Return path, no error
	}

	return workspaceRoot, nil
}

// ProjectPaths contains the paths for standard i3 project files
type ProjectPaths struct {
	WorkspaceRoot  string
	I3Dir          string
	FeaturesDir    string
	CursorRulesDir string
}

// GetProjectPaths returns the standard paths for i3 project files
func GetProjectPaths() (ProjectPaths, error) {
	workspaceRoot, err := GetWorkspaceRoot()
	if err != nil {
		return ProjectPaths{}, fmt.Errorf("failed to get workspace root: %w", err) // Propagate error
	}

	i3Dir := filepath.Join(workspaceRoot, ".i3")
	featuresDir := filepath.Join(i3Dir, "features")
	cursorRulesDir := filepath.Join(workspaceRoot, ".cursor", "rules")

	return ProjectPaths{
		WorkspaceRoot:  workspaceRoot,
		I3Dir:          i3Dir,
		FeaturesDir:    featuresDir,
		CursorRulesDir: cursorRulesDir,
	}, nil
}

// ProjectFiles defines the standard project file templates
var ProjectFiles = map[string]string{
	"project.md": `# Project Overview

## Problem Statement

## Goals

## Scope

## Timeline
`,
	"tech.md": `# Technical Overview

## Technology Stack

## Architecture

## Key Dependencies

## Development Environment
`,
}

// EnsureProjectFiles ensures the base project files exist
// Returns a list of newly created files
func EnsureProjectFiles(clean bool) ([]string, error) {
	paths, err := GetProjectPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to get project paths: %w", err) // Handle error from GetProjectPaths
	}
	var newlyCreated []string

	// Create each base file
	for filename, content := range ProjectFiles {
		filePath := filepath.Join(paths.I3Dir, filename)

		// Skip if file exists and we're not doing a clean init
		if !clean {
			if _, err := os.Stat(filePath); err == nil {
				continue
			}
		}

		// Create directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		// Write file
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", filename, err)
		}

		newlyCreated = append(newlyCreated, filename)
	}

	return newlyCreated, nil
}
