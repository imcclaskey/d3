# d3 - Define, Design, Deliver!

<div align="center">
  <img src="https://img.shields.io/badge/status-alpha-orange" alt="Status: Alpha">
  <img src="https://img.shields.io/badge/language-go-blue" alt="Language: Go">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License: MIT">
  <img src="https://github.com/imcclaskey/d3/workflows/CI/badge.svg" alt="CI Status">
</div>

<p align="center">
  <b>A structured workflow engine for AI-driven development within Cursor</b>
</p>

---

## 🧠 What is d3?

d3 helps you control and guide AI assistants through a structured development process. Think of it as a framework that brings intentionality to your AI collaboration, steering conversations through three distinct phases:

- **Define**: Clarify what you're building and why
- **Design**: Plan how to build it
- **Deliver**: Implement the solution

By integrating with the Model Context Protocol (MCP) in tools like [Cursor](https://cursor.sh), d3 enables your AI assistant to stay "on rails" within each development phase, producing more focused and relevant results at each step.

### Core Benefits

- **Guided AI Conversations**: Transform chaotic AI interactions into predictable, phase-based workflows where the AI knows exactly what type of assistance to provide.
- **Context Management**: Keep your AI assistant focused on the right level of abstraction—requirements, architecture, or implementation—based on your current phase.
- **Thought Before Action**: Encourage proper planning and design before jumping into code, resulting in more maintainable solutions.
- **Living Documentation**: Automatically capture requirements, design decisions, and implementation progress as you collaborate with your AI assistant.

## 🚀 Quick Start

### Installation

#### Homebrew (macOS/Linux)

```bash
# Add the tap and install
brew tap imcclaskey/tap
brew install d3
```

#### Download from GitHub Releases

Go to the [Releases page](https://github.com/imcclaskey/d3/releases) and download the binary for your platform:
- macOS: `d3-darwin-amd64` (Intel) or `d3-darwin-arm64` (Apple Silicon)
- Linux: `d3-linux-amd64` (Intel) or `d3-linux-arm64` (ARM)
- Windows: `d3-windows-amd64.exe`

### Basic Workflow

1. **Initialize d3**:

    ```bash
    # Initialize d3 in your project directory
    d3 init
    ```

2. **Interact via MCP Client (e.g., Cursor)**:
    - Create a Feature: Ask the AI to "create a new d3 feature named 'my-feature'"
    - Move Through Phases: Ask the AI to "move to the design phase" or "move to the deliver phase"
    - Exit Feature: When done, ask the AI to "exit the current feature"

3. **Develop within Phases**:
    - **Define**: Work with the AI to populate `problem.md` with requirements
    - **Design**: Collaborate with the AI on the technical plan in `plan.md`
    - **Deliver**: Generate code with the AI, tracking progress in `progress.yaml`

## 📋 Development Phases

d3 enforces a structured workflow through three main phases:

### 1. Define Phase

Focus on the problem space, requirements, and user needs. Answer the "what" and "why" without diving into implementation details. Document everything in `problem.md`.

### 2. Design Phase

Translate ideas into a technical blueprint. Design architecture, component interactions, and implementation steps in `plan.md`. This becomes the roadmap for implementation.

### 3. Deliver Phase

Generate code following the technical plan, with progress tracked in `progress.yaml`. Focus exclusively on writing high-quality, maintainable code that aligns with the established plan.

## 🔧 Custom Workflow Templates

d3 allows you to customize the workflow templates used in each phase:

1. **Initialize with custom templates**:
   ```bash
   d3 init --custom-rules
   ```
   This creates a `.d3/rules/` directory with default templates you can modify.

2. **Customize your templates** in the `.d3/rules/` directory:
   - Edit templates to match your team's workflow
   - Modify AI guidance for each phase
   - Customize documentation structures

3. **Automatic application**: d3 will use your custom templates for all phase transitions and project initializations, ensuring your workflow is consistent across your project.

## 🛠️ Commands & MCP Tools

### CLI Commands

| Command                    | Description                                                 |
|----------------------------|-------------------------------------------------------------|
| `d3 init [--custom-rules]` | Initialize d3 project. Use `--custom-rules` to create editable template files |
| `d3 feature create <name>` | Create a new feature and set it as the current context      |
| `d3 feature enter <name>`  | Enter a feature context, resuming its last known phase      |
| `d3 phase move <phase>`    | Move to a different phase (define, design, deliver)         |
| `d3 exit`                  | Exit the current feature context                            |
| `d3 feature delete <name>` | Delete a feature and its associated content                 |
| `d3 serve`                 | Start the d3 MCP server for AI interaction                  |
| `d3 version`               | Display the current version of d3                           |

### MCP Tool Functions (Used via AI Assistant)

| MCP Function          | Description                                          |
|-----------------------|------------------------------------------------------|
| `d3_feature_create`   | Create a new feature and set it as current context   |
| `d3_feature_enter`    | Enter a feature context, resuming its last phase     |
| `d3_feature_exit`     | Exit the current feature context                     |
| `d3_feature_delete`   | Delete a feature and its associated content          |
| `d3_phase_move`       | Move to a different phase (define, design, deliver)  |

## 📂 Project Structure

```text
project/
├── .d3/                  # d3 configuration and feature documentation
│   ├── features/         # Feature-specific documentation
│   │   └── my-feature/   # Individual feature folder
│   │       ├── define/        # Define Phase artifacts
│   │       │   └── problem.md   # Problem definition and requirements
│   │       ├── design/        # Design Phase artifacts
│   │       │   └── plan.md      # Technical implementation plan
│   │       ├── deliver/       # Deliver Phase artifacts
│   │       │   └── progress.yaml# Implementation progress tracking
│   │       └── .phase        # Stores the current phase for this feature
│   ├── rules/            # Custom workflow templates (when using --custom-rules)
│   └── .feature           # Current active feature name (if any)
├── .cursor/              # Cursor IDE configuration
│   └── rules/            # Client-side rules
│       └── d3/           # d3-specific rules
│           ├── core.gen.mdc     # Core rules for d3
│           └── phase.gen.mdc    # Phase-specific rules (generated by d3)
└── .gitignore           # Will include entries for proper d3 file handling
```

<p align="center">
  Built with ❤️ for better AI collaboration
</p>