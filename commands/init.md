# Command: /init
> Initialize a local project with brain repo rules and instructions.

## Description
This command sets up the necessary symlinks in the current directory to enable "brain" functionality for project-specific tools like Cursor and GitHub Copilot.
It also generates a stack-aware project context file at `.brain/skill-context.md` and installs a git-native pre-commit Guardian hook.

## Execution
Execute the initialization script from your terminal:

```bash
bash ~/.brain/scripts/init.sh
```

## How to use
Type `/init` in any agent that supports slash commands, or run the bash snippet manually in your project root.
