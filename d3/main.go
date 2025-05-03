package main

import (
	"fmt"
	"os"

	"github.com/imcclaskey/d3/internal/cli"
)

func main() {
	// Create new CLI instance
	app := cli.NewCLI()

	// Initialize commands
	app.InitCommands()

	// Execute the CLI
	if err := app.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
