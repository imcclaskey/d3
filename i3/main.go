package main

import (
	"context"
	"fmt"
	"os"
	
	"github.com/spf13/cobra"
	
	"github.com/imcclaskey/i3/internal/command"
)

var (
	rootCmd = &cobra.Command{
		Use:   "i3",
		Short: "i3 Framework CLI",
		Long:  "Interactive Intelligent Interface (i3) Framework CLI",
	}
	
	// Command flags
	cleanFlag bool
)

func init() {
	// Initialize command and flags
	initCommands()
}

func initCommands() {
	// Create command
	createCmd := &cobra.Command{
		Use:   "create <feature>",
		Short: "Create a new feature and set it as the current context",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runCommand(command.NewCreate(args[0]))
		},
	}
	
	// Init command
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize i3 in the current workspace (clears context)",
		Run: func(cmd *cobra.Command, args []string) {
			runCommand(command.NewInit(cleanFlag))
		},
	}
	
	// Add flags to init command
	initCmd.Flags().BoolVar(&cleanFlag, "clean", false, "Perform a clean initialization (remove existing files)")
	
	// Enter command
	enterCmd := &cobra.Command{
		Use:   "enter <feature>",
		Short: "Set the current feature context",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runCommand(command.NewEnter(args[0]))
		},
	}
	
	// Leave command
	leaveCmd := &cobra.Command{
		Use:   "leave",
		Short: "Leave the current feature context",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runCommand(command.NewLeave())
		},
	}
	
	// Phase command
	phaseCmd := &cobra.Command{
		Use:   "phase <phase>",
		Short: "Set the current phase within the active feature context",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runCommand(command.NewPhase(args[0]))
		},
	}
	
	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show current i3 feature and phase context",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runCommand(command.NewStatus())
		},
	}
	
	// Refresh command
	refreshCmd := &cobra.Command{
		Use:   "refresh",
		Short: "Ensure necessary i3 files and directories exist",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// NewRefresh takes no arguments now
			runCommand(command.NewRefresh())
		},
	}
	
	// Add all commands to root
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(enterCmd)
	rootCmd.AddCommand(leaveCmd)
	rootCmd.AddCommand(phaseCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(refreshCmd)
}

// runCommand is a helper to reduce boilerplate in command Run functions
func runCommand(cmd command.Command) {
	workspaceRoot := getWorkspaceRoot()
	cfg := command.New(workspaceRoot)
	result, err := cmd.Run(context.Background(), cfg)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print success message (use Result.Message)
	if result.Message != "" {
		fmt.Println(result.Message)
	}
	// Optionally print warnings
	// for _, warning := range result.Warnings {
	// 	fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
	// }
	// Optionally handle/print result.Data
}

// getWorkspaceRoot gets the workspace root directory
func getWorkspaceRoot() string {
	// Always use the current directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	
	return cwd
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 