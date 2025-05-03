# d3 - Define, Design, Deliver!

<div align="center">
  <img src="https://img.shields.io/badge/status-alpha-orange" alt="Status: Alpha">
  <img src="https://img.shields.io/badge/language-go-blue" alt="Language: Go">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License: MIT">
</div>

<p align="center">
  <b>A structured workflow engine for AI-driven development within Cursor</b>
</p>

---

## 🧠 What is d3?

d3 is a CLI tool that orchestrates intentional, AI-driven development workflows within [Cursor](https://cursor.sh). By imposing a structured, phase-based process managed directly from your terminal, d3 acts as an agent control system that enhances your AI pair programming experience.

### Core Benefits

- **Structured AI Collaboration**: Move beyond chaotic, ad-hoc AI interactions to a predictable, phase-based workflow
- **Separation of Concerns**: Keep problem definition, solution planning, and implementation distinct and focused
- **Optimized AI Context**: Provide the right context to your AI agent at each phase for better outcomes
- **Consistent Documentation**: Automatically generate and maintain technical documentation as you build

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher (for building from source)
- [Cursor IDE](https://cursor.sh/)

### Installation

#### Download from GitHub Releases

1. Go to the [Releases page](https://github.com/imcclaskey/d3/releases) of this repository
2. Download the binary for your platform:
   - macOS/Intel: `d3-darwin-amd64`
   - macOS/Apple Silicon: `d3-darwin-arm64`
   - Linux/Intel: `d3-linux-amd64`
   - Linux/ARM: `d3-linux-arm64`
   - Windows: `d3-windows-amd64.exe`
3. Make the binary executable (Linux/macOS):
   ```bash
   chmod +x d3-darwin-arm64  # Example for Mac with Apple Silicon
   ```
4. Move the binary to a location in your PATH (Linux/macOS):
   ```bash
   mv d3-darwin-arm64 /usr/local/bin/d3
   ```

#### Building from Source

```bash
# Clone the repository
git clone https://github.com/imcclaskey/d3.git
cd d3

# Build and install
make install
```

### Basic Workflow

```bash
# Initialize d3 in your project
d3 init

# Create a new feature
d3 create my-feature

# Progress through development phases
d3 phase define     # Define the problem and requirements
d3 phase design     # Plan the technical implementation
d3 phase deliver    # Write the actual code

# Check your current status
d3 status
```

## 📋 Development Phases

d3 enforces a structured workflow through three main phases:

### 1. Define Phase

Focus on the problem space, requirements, and user needs. Answer the "what" and "why" without diving into implementation details. Document everything in `define.md`.

### 2. Design Phase

Translate ideas into a technical blueprint. Design architecture, component interactions, and implementation steps in `design.md`. This becomes the roadmap for implementation.

### 3. Deliver Phase

Generate code following the technical plan, with progress tracked in `deliver.json`. Focus exclusively on writing high-quality, maintainable code that aligns with the established plan.

## 🛠️ Commands

| Command | Description |
|---------|-------------|
| `d3 init [--clean]` | Initialize d3 in the current workspace |
| `d3 create <feature>` | Create a new feature and set it as the current context |
| `d3 enter <feature>` | Set the current feature context |
| `d3 leave` | Leave the current feature context |
| `d3 phase <phase>` | Set the current phase (setup, define, design, deliver) |
| `d3 status` | Show current d3 feature and phase context |
| `d3 refresh` | Ensure necessary d3 files and directories exist |
| `d3 version` | Display the current version of d3 |

## 📂 Project Structure

```
project/
├── .d3/                  # d3 configuration and feature documentation
│   ├── features/         # Feature-specific documentation
│   │   └── my-feature/   # Individual feature folder
│   │       ├── define.md      # Problem definition and requirements
│   │       ├── design.md      # Technical implementation plan
│   │       └── deliver.json   # Implementation progress tracking
│   ├── context.json      # Current feature and phase context
│   ├── project.md        # Project overview and business objectives
│   └── tech.md           # Technology stack documentation
└── .cursor/              # Cursor IDE configuration
    └── rules/            # Cursor rules
        └── d3/           # d3-specific rules for Cursor
            ├── core.gen.mdc     # Core rules for d3
            └── phase.gen.mdc    # Phase-specific rules
```

## 🔄 How It Works

1. **Feature Creation**: Each feature gets its own documentation directory with phase-specific files
2. **Context Setting**: d3 maintains your current feature and phase context in `context.json`
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