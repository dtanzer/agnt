# Plan for Implementing agnt

## Done

[x] Deliver a working `agnt` binary that a user can install and invoke. It doesn't manage agents yet, but it feels like a real tool: it has a consistent command structure, useful help text, and an `info` command that tells you what version you're running. - See specs/01-cli-skeleton.md

[x] `agnt new-workspace` — creates a fresh `.agnt.yaml` in the current directory, even if one already exists in a parent directory. - See specs/02-new-workspace.md

[x] `agnt register <name> <type> [--variant <variant>]` — registers the current tmux pane as a named agent, writing to `.agnt.yaml`. Searches parent directories for an existing config file; creates one in the current directory if none found. Errors if the agent name already exists. - See specs/03-register.md

## Current

## Next

[ ] `agnt validate` — finds the workspace, reports where it is (or that none was found), checks the file is syntactically valid, and verifies that the registered tmux panes exist and look correct.

## Later

- `agnt register` currently errors if the agent name already exists — may need a `--force` flag (or similar) to update an existing entry.
- Implement the rest of the functionality, to be refined.
