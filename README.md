# i3 - Ideate, Instruct, Implement!

<div align="center">
  <img src="https://img.shields.io/badge/status-alpha-orange" alt="Status: Alpha">
  <img src="https://img.shields.io/badge/language-go-blue" alt="Language: Go">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License: MIT">
</div>

<p align="center">
  <b>A structured workflow engine for AI-driven development within Cursor</b>
</p>

---

## 🧠 What is i3?

i3 is a CLI tool that orchestrates intentional, AI-driven development workflows within [Cursor](https://cursor.sh). By imposing a structured, phase-based process managed directly from your terminal, i3 acts as an agent control system that enhances your AI pair programming experience.

### Core Benefits

- **Structured AI Collaboration**: Move beyond chaotic, ad-hoc AI interactions to a predictable, phase-based workflow
- **Separation of Concerns**: Keep problem definition, solution planning, and implementation distinct and focused
- **Optimized AI Context**: Provide the right context to your AI agent at each phase for better outcomes
- **Consistent Documentation**: Automatically generate and maintain technical documentation as you build

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- [Cursor IDE](https://cursor.sh/)

### Installation

```bash
# Clone the repository
git clone https://github.com/imcclaskey/i3.git
cd i3

# Build and install
make install
```

### Basic Workflow

```bash
# Initialize i3 in your project
i3 init

# Create a new feature
i3 create my-feature

# Progress through development phases
i3 phase ideation     # Define the problem and requirements
i3 phase instruction  # Plan the technical implementation
i3 phase implementation  # Write the actual code

# Check your current status
i3 status
```

## 📋 Development Phases

i3 enforces a structured workflow through three main phases:

### 1. Ideation Phase

Focus on the problem space, requirements, and user needs. Answer the "what" and "why" without diving into implementation details. Document everything in `ideation.md`.

### 2. Instruction Phase

Translate ideas into a technical blueprint. Design architecture, component interactions, and implementation steps in `instruction.md`. This becomes the roadmap for implementation.

### 3. Implementation Phase

Generate code following the technical plan, with progress tracked in `implementation.json`. Focus exclusively on writing high-quality, maintainable code that aligns with the established plan.

## 🛠️ Commands

| Command | Description |
|---------|-------------|
| `i3 init [--clean]` | Initialize i3 in the current workspace |
| `i3 create <feature>` | Create a new feature and set it as the current context |
| `i3 enter <feature>` | Set the current feature context |
| `i3 leave` | Leave the current feature context |
| `i3 phase <phase>` | Set the current phase (setup, ideation, instruction, implementation) |
| `i3 status` | Show current i3 feature and phase context |
| `i3 refresh` | Ensure necessary i3 files and directories exist |

## 📂 Project Structure

```
project/
├── .i3/                  # i3 configuration and feature documentation
│   ├── features/         # Feature-specific documentation
│   │   └── my-feature/   # Individual feature folder
│   │       ├── ideation.md      # Problem definition and requirements
│   │       ├── instruction.md   # Technical implementation plan
│   │       └── implementation.json  # Implementation progress tracking
│   ├── context.json      # Current feature and phase context
│   ├── project.md        # Project overview and business objectives
│   └── tech.md           # Technology stack documentation
└── .cursor/              # Cursor IDE configuration
    └── rules/            # Cursor rules
        └── i3/           # i3-specific rules for Cursor
            ├── core.gen.mdc     # Core rules for i3
            └── phase.gen.mdc    # Phase-specific rules
```

## 🔄 How It Works

1. **Feature Creation**: Each feature gets its own documentation directory with phase-specific files
2. **Context Setting**: i3 maintains your current feature and phase context in `context.json`
3. **Rule Generation**: Phase-appropriate rule files are generated for Cursor based on your context
4. **AI Guidance**: Cursor's AI assistant uses these rules to provide phase-appropriate guidance
5. **Documentation Tracking**: Progress and decisions are documented throughout the development lifecycle

## 🤝 Contributing

Contributions are welcome! Feel free to submit issues or pull requests.

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add some amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

<p align="center">
  Built with ❤️ for better AI collaboration
</p>