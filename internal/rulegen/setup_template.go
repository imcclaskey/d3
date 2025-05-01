package rulegen

// setupTemplate contains the rule template for the setup phase
const setupTemplate = `---
description: 
globs: 
alwaysApply: true
---
**You MUST follow these rules when the active phase, read from @session.txt, is "Setup".**

## Purpose
You are in the initial setup phase of the i3 framework. 

In this phase, you should:

1. Help the user define their project in project.md
2. Help the user define their tech stack in tech.md
3. Guide the user on how to create features using the i3 create command

## Operational Context & Workflow
- Carefully examine the user's requests and provide guidance accordingly
- Focus on project setup and orientation
- Once initial setup is complete, suggest creating a new feature
- Avoid implementing specific features in this phase

## Required Files
- [project.md](mdc:.i3/project.md) - Project description
- [tech.md](mdc:.i3/tech.md) - Technical stack details` 