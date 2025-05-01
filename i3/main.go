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
	forceFlag      bool
	featureFlag    string
)

func init() {
	// Initialize command and flags
	initCommands()
}

func initCommands() {
	// Create command
	createCmd := &cobra.Command{
		Use:   "create <feature>",
		Short: "Create a new feature",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get workspace root
			workspaceRoot := getWorkspaceRoot()
			
			// Create command config
			cfg := command.New(workspaceRoot)
			
			// Parse arguments
			featureName := args[0]
			
			// Create and execute command
			create := command.NewCreate(featureName)
			message, err := create.Run(context.Background(), cfg)
			
			// Handle result
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			
			// Print success message
			fmt.Println(message)
		},
	}
	
	// Init command
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize i3 in the current workspace",
		Run: func(cmd *cobra.Command, args []string) {
			// Get workspace root
			workspaceRoot := getWorkspaceRoot()
			
			// Create command config
			cfg := command.New(workspaceRoot)
			
			// Create and execute command
			init := command.NewInit(forceFlag)
			message, err := init.Run(context.Background(), cfg)
			
			// Handle result
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			
			// Print success message
			fmt.Println(message)
		},
	}
	
	// Add flags to init command
	initCmd.Flags().BoolVar(&forceFlag, "force", false, "Force re-initialization (overwrites existing files)")
	
	// Move command
	moveCmd := &cobra.Command{
		Use:   "move <phase> [feature]",
		Short: "Move to a different phase or feature",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get workspace root
			workspaceRoot := getWorkspaceRoot()
			
			// Create command config
			cfg := command.New(workspaceRoot)
			
			// Parse arguments
			phase := args[0]
			var feature string
			if len(args) > 1 {
				feature = args[1]
			}
			
			// Create and execute command
			move := command.NewMove(phase, feature, false)
			message, err := move.Run(context.Background(), cfg)
			
			// Handle result
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			
			// Print success message
			fmt.Println(message)
		},
	}
	
	// Exit command
	exitCmd := &cobra.Command{
		Use:   "exit",
		Short: "Exit the current session",
		Run: func(cmd *cobra.Command, args []string) {
			// Get workspace root
			workspaceRoot := getWorkspaceRoot()
			
			// Create command config
			cfg := command.New(workspaceRoot)
			
			// Create and execute command
			exit := command.NewExit()
			message, err := exit.Run(context.Background(), cfg)
			
			// Handle result
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			
			// Print success message
			fmt.Println(message)
		},
	}
	
	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show current i3 status",
		Run: func(cmd *cobra.Command, args []string) {
			// Get workspace root
			workspaceRoot := getWorkspaceRoot()
			
			// Create command config
			cfg := command.New(workspaceRoot)
			
			// Create and execute command
			status := command.NewStatus()
			message, err := status.Run(context.Background(), cfg)
			
			// Handle result
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			
			// Print success message
			fmt.Println(message)
		},
	}
	
	// Refresh command
	refreshCmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh rules for the current session",
		Run: func(cmd *cobra.Command, args []string) {
			// Get workspace root
			workspaceRoot := getWorkspaceRoot()
			
			// Create command config
			cfg := command.New(workspaceRoot)
			
			// Create and execute command
			refresh := command.NewRefresh(featureFlag)
			message, err := refresh.Run(context.Background(), cfg)
			
			// Handle result
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			
			// Print success message
			fmt.Println(message)
		},
	}
	
	// Add flags to refresh command
	refreshCmd.Flags().StringVar(&featureFlag, "feature", "", "Feature to refresh (defaults to current feature)")
	
	// Add all commands to root
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(moveCmd)
	rootCmd.AddCommand(exitCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(refreshCmd)
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