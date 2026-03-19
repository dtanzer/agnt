# Plan for Implementing agnt

[x] Deliver a working `agnt` binary that a user can install and invoke. It doesn't manage agents yet, but it feels like a real tool: it has a consistent command structure, useful help text, and an `info` command that tells you what version you're running. - See specs/01-cli-skeleton.md

[x] `agnt new-workspace` — creates a fresh `.agnt.yaml` in the current directory, even if one already exists in a parent directory. - See specs/02-new-workspace.md

[x] `agnt register <name> <type> [--variant <variant>]` — registers the current tmux pane as a named agent, writing to `.agnt.yaml`. Searches parent directories for an existing config file; creates one in the current directory if none found. Errors if the agent name already exists. - See specs/03-register.md

[x] `agnt --workspace-config <file> <command>` — global flag that overrides the workspace config file lookup for any command. Errors if the specified file does not exist. - See specs/04-workspace-config-flag.md

[x] `agnt validate` — finds the workspace, reports where it is (or that none was found), checks the file is syntactically valid, and verifies that the registered tmux panes exist and look correct. - See specs/05-validate.md

[x] Migrate pane identification from tmux pane ID (`%42`) to window.pane index (`0.1`). - See specs/06-pane-index.md

[x] Test pane index stability with tmux-resurrect — run test cases from specs/07-pane-index-stability-tests.md on a real machine and update the plan based on results. - See specs/07-pane-index-stability-tests.md

[x] Fix `validate` bug: pane lookup searches all tmux sessions (`list-panes -a`), so a pane index that exists in *any* session is reported OK — even if the registered agent's pane is gone. Lookup must be scoped to the current tmux session only.

[x] Agent type definitions in workspace config — users define named types in `.agnt.yaml` under a `types:` key. Each type has a `run:` field with the command to execute (e.g. `claude --some-flag` or `docker run my-image`). `agnt validate` gains a check that each registered agent's type is defined in the config. - See specs/08-type-definitions.md

[x] Placeholder system for type definitions — the `run:` command can include placeholders (e.g. `{{name}}`, `{{variant}}`, `{{server_url}}`) that are substituted at start time. For bare-metal types, placeholders expand to CLI arguments appended to the command; for container types, they expand to `--env` flags passed to `docker run`. Exact placeholder syntax and available variables defined here.

[x] `agnt start <name>` — looks up the agent by name, finds its type definition, substitutes placeholders, and sends the resulting command as keystrokes to the registered tmux pane.

[ ] `agnt server` — starts a background HTTP daemon. `agnt server status` checks if it's running and prints basic info (pid, uptime, listening address). No agent state yet — just liveness.

[ ] `agnt attach <name>` — registers a running agent with the server. Captures the pane and enough info to kill it later: a local PID for bare-metal agents, a container ID for containerised agents (distinguished by type or an explicit flag). `agnt server status` lists attached agents.

[ ] `agnt kill <name>` — instructs the server to terminate the named agent using the info captured at attach time (SIGTERM for bare-metal, `docker stop` for containers).

## Later

- `agnt register` currently errors if the agent name already exists — may need a `--force` flag (or similar) to update an existing entry.
- Multiple agents sharing the same pane index is intentional (groups are alternative configurations for the same layout) — `validate` should not flag this as an error.
- `agnt remap` — interactively or automatically reassign agents to their current pane indices, for when the tmux layout has changed since the config was written. `validate` should suggest running this when it detects index mismatches.
- Global type definitions in a home-directory config (e.g. `~/.agnt.yaml`), merged with workspace config at runtime.
- `agnt start` auto-attaches with the server at startup, if one is running.
