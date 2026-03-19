# Plan for Implementing agnt

## Done

[x] Deliver a working `agnt` binary that a user can install and invoke. It doesn't manage agents yet, but it feels like a real tool: it has a consistent command structure, useful help text, and an `info` command that tells you what version you're running. - See specs/01-cli-skeleton.md

[x] `agnt new-workspace` — creates a fresh `.agnt.yaml` in the current directory, even if one already exists in a parent directory. - See specs/02-new-workspace.md

[x] `agnt register <name> <type> [--variant <variant>]` — registers the current tmux pane as a named agent, writing to `.agnt.yaml`. Searches parent directories for an existing config file; creates one in the current directory if none found. Errors if the agent name already exists. - See specs/03-register.md

[x] `agnt --workspace-config <file> <command>` — global flag that overrides the workspace config file lookup for any command. Errors if the specified file does not exist. - See specs/04-workspace-config-flag.md

[x] `agnt validate` — finds the workspace, reports where it is (or that none was found), checks the file is syntactically valid, and verifies that the registered tmux panes exist and look correct. - See specs/05-validate.md

## Current

## Next

[ ] Migrate pane identification from tmux pane ID (`%42`) to window:pane index (`0:1`). Indices are stable within a layout, visible in the tmux UI (`Ctrl+B q`), and more meaningful to users. Affects `register`, `validate`, and the config file format. `register` should capture the index at registration time. `validate` should check by index.

[ ] `agnt remap` — interactively or automatically reassign agents to their current pane indices, for when the tmux layout has changed since the config was written. `validate` should suggest running this when it detects index mismatches.

## Later

- `agnt register` currently errors if the agent name already exists — may need a `--force` flag (or similar) to update an existing entry.
- Multiple agents can share a pane (user-managed groups) — this is intentional and should not be flagged as an error by `validate`.
- Implement the rest of the functionality, to be refined.
