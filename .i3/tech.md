# i3 Technology Stack

## Programming Languages
- **Go (Golang)** - Primary implementation language (Go 1.21)

## Frameworks and Libraries
- **Cobra** - CLI application framework for Go
  - Used for command-line interface structure and command management
  - Provides flags, arguments parsing, and help documentation
- **Testify** - Testing toolkit for Go
  - Used for unit testing with enhanced assertions

## Database Technologies
- No database currently in use, as the application focuses on file-based operations

## Infrastructure and Deployment
- **Local Development**: Command-line tool for local execution
- **Distribution**: Binary executable for macOS/Linux/Windows
- **Installation**: Local build or distribution via package managers (potential future)

## Development Tools and Workflows
- **Make** - Build automation
  - Used for standardizing build commands and processes
- **Git** - Version control system
  - Standard branching and merge workflow
- **Cursor** - Primary IDE with AI integration
  - Tool is designed specifically to enhance Cursor workflows

## Architecture Patterns and Practices
- **Command Pattern** - Used for implementing CLI commands
- **Phase-Based Architecture** - Separation of development into distinct phases:
  - Ideation
  - Instruction
  - Implementation
- **Directory Structure**:
  - `/bin` - Compiled binaries
  - `/i3` - Core application code
  - `/internal` - Internal packages:
    - `/command` - CLI command implementations
    - `/context` - Context management for AI interactions
    - `/rules` - Rule templates for different phases
    - `/validation` - Input validation logic
    - `/workspace` - Workspace management utilities
  - `/.i3` - Project-specific configuration and feature documentation

## Coding Standards and Practices
- Go idiomatic code style
- Unit testing for critical functionality
- Clear separation of concerns between packages
- Structured error handling
- Documentation with Markdown
