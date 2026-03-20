# agnt

`agnt` is a host-side server and CLI tool for managing named AI agent sessions in tmux panes.
The server runs on the host (inside the tmux session it manages), accepts REST commands from
anywhere that can reach it — including Claude/opencode instances running in Docker or Podman
containers — and executes the actual tmux operations on the host.

The CLI is a thin REST client that locates the running server via a session file and issues
commands to it.

Inspired by [Lada Kesseler's talk](https://www.youtube.com/watch?v=_LSK2bVf0Lc&t=8125).

---

## Typical workflow

### First-time setup

```bash
# 1. Create your tmux session and split it into your desired layout
tmux new-session -s agents
# ... split panes with Ctrl+B % and Ctrl+B "

# 2. Create a workspace in your project directory
cd ~/project
agnt new-workspace

# 3. Edit .agnt.yaml to define your agent types
#    (add a types: section with run: commands)

# 4. In each pane, register the agent
#    (run this from inside each pane)
agnt register Alice claude
agnt register Bob docker-claude --variant go

# 5. Start the server in a dedicated pane
agnt server start

# 6. Start your agents
agnt start Alice
agnt start Bob

# 7. Verify everything looks right
agnt validate
```

### After reopening tmux

If your pane indices have changed (e.g. new tmux session), run:

```bash
agnt validate        # see what's missing
agnt remap           # reassign agents to current panes (coming soon)
```

If you use [tmux-resurrect](https://github.com/tmux-plugins/tmux-resurrect), your pane layout —
including pane indices — is reliably restored after a system restart. No remap needed.

### Working with multiple workspaces

Use `--workspace-config` to work with a specific config file without changing your current
directory:

```bash
agnt --workspace-config ~/teamA/.agnt.yaml validate
agnt --workspace-config ~/teamB/.agnt.yaml server status
```

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  Host (tmux session)                                │
│                                                     │
│  ┌──────────────┐   REST    ┌───────────────────┐   │
│  │  agnt server │◄──────────│  agnt CLI         │   │
│  │  start       │           │  (thin client)    │   │
│  │  (port 7717) │           └───────────────────┘   │
│  │              │                                   │
│  │              │──tmux send-keys──► pane A (claude)│
│  │              │──tmux send-keys──► pane B (claude)│
│  └──────────────┘                                   │
└─────────────────────────────────────────────────────┘
```

**Components:**

- **`agnt server start`** — foreground REST server, runs in the tmux session it will manage.
  Logs spawns and errors to stdout. Must be started before using `agnt start`.
- **`agnt <command>`** — thin CLI client. Discovers the server port from a session file next to
  `.agnt.yaml` and issues REST requests.

---

## Installation

```bash
git clone <repo>
cd agnt
make install
```

Or build locally:

```bash
make build
# produces ./agnt
```

**Requirements:** Go 1.21+, tmux

---

## Concepts

### Workspace

A workspace is a `.agnt.yaml` file that describes your agent types and registered agents. `agnt`
searches for it by walking up from the current directory toward `$HOME` — so you can run `agnt`
commands from anywhere inside your project tree.

```yaml
types:
  claude:
    run: "claude --agent {{name}}"
  docker-claude:
    run: "docker run --rm -e AGENT={{name}} -e VARIANT={{variant}} my-claude-image"

agents:
  Alice:
    type: claude
    pane: "0.0"
  Bob:
    type: docker-claude
    variant: go
    pane: "0.1"
```

Each agent has:
- **name** — stable identifier used in all commands
- **type** — must match a key in the `types:` section
- **variant** *(optional)* — passed to the type's run command via `{{variant}}`
- **pane** — tmux window.pane index (e.g. `0.1`)

### Agent types

Types are defined in the `types:` section of `.agnt.yaml`. Each type has a `run:` field with
the command to send to the pane when `agnt start` is called.

```yaml
types:
  simple:
    run: "claude"
  named:
    run: "claude --agent {{name}}"
```

#### Placeholders

The `run:` command can include placeholders that are substituted with agent-specific values at
start time:

| Placeholder   | Value |
|---------------|-------|
| `{{name}}`    | The agent's name as registered in `.agnt.yaml` |
| `{{variant}}` | The agent's variant, or empty string if none is set |

Substitution is plain string replacement. Using any unknown placeholder is an error at start
time. A `run:` string with no placeholders is used as-is.

Agent names and variants must not contain spaces (enforced by `agnt register`).

### Pane layout

The recommended workflow is to create a fixed tmux split layout first, then register your agents
into it. The pane indices (shown in tmux when you press `Ctrl+B q`) stay stable as long as your
layout doesn't change.

### Agent groups

Groups are alternative agent configurations for the same pane layout. Multiple agents can share
the same pane index — but only one runs there at a time:

```yaml
agents:
  Alice:   { type: claude, pane: "0.0" }   # Group 1
  Bob:     { type: claude, pane: "0.1" }   # Group 1
  Charlie: { type: claude, pane: "0.0" }   # Group 2
  Dave:    { type: claude, pane: "0.1" }   # Group 2
```

You run either Group 1 (Alice+Bob) or Group 2 (Charlie+Dave) in the same two panes. Stop one
group before starting the other. The tool doesn't need to know about groups — you decide which
agents to start and stop.

---

## Commands

### `agnt server start [--port <port>]`

Starts the REST server in the foreground. Must be run inside a tmux session. Logs to stdout.

```bash
agnt server start
# agnt server listening on localhost:7717
```

- Default port is `7717`. Override with `--port`.
- Writes a `.agnt-server.yaml` session file next to `.agnt.yaml` on start; removes it on exit.
- Errors if a server is already running for this workspace.

Run this in a dedicated tmux pane before using `agnt start`.

---

### `agnt server status`

Checks whether the server is running and prints basic info.

```bash
agnt server status
# Server:   running
# PID:      12345
# Address:  localhost:7717
# Uptime:   3m42s
```

Queries the server's health endpoint directly — works from inside a container if the server is
reachable (e.g. with `--network host` and the workspace directory mounted).

---

### `agnt start <name>`

Starts the named agent by sending its configured command as keystrokes to its registered tmux pane.

```bash
agnt start Alice
# Starting Alice in pane 0.0: claude --agent Alice
```

Requires a running server (`agnt server start`). Errors if no server is found, the agent name is
unknown, its pane doesn't exist, or its type is not defined.

---

### `agnt new-workspace`

Creates a new `.agnt.yaml` in the current directory.

```bash
agnt new-workspace
# Created workspace at /home/user/project/.agnt.yaml
```

Errors if a `.agnt.yaml` already exists in the current directory. Use `--workspace-config` to
create a file at an arbitrary path.

---

### `agnt register <name> <type> [--variant <variant>]`

Registers the current tmux pane as a named agent. Must be run from inside a tmux pane.

```bash
agnt register Alice claude
agnt register Bob docker-claude --variant go
```

- Searches parent directories for an existing `.agnt.yaml`; creates one in the current directory
  if none is found.
- Errors if an agent with that name already exists.
- Errors if the name or variant contains spaces.

---

### `agnt validate`

Checks the workspace config and reports the status of each agent's pane.

```bash
agnt validate
```

```
Workspace:  /home/user/project/.agnt.yaml
Syntax:     OK
Agents (2):
  Alice        pane 0.0  OK  [claude]
  Bob          pane 0.1  OK  [bash]

Summary: 2/2 checks passed
```

Checks that:
- The workspace file exists and is valid YAML
- Each agent's type is defined in `types:`
- Each agent's pane exists in the current tmux session

Exits non-zero if any check fails.

---

### `agnt info`

Prints version and build information.

```bash
agnt info
# Version:    1.0.0
# Commit:     abc1234
# Built:      2025-01-01T00:00:00Z
# Go version: go1.24.1
# Binary:     /usr/local/bin/agnt
```

---

### Global flag: `--workspace-config <file>`

Override the workspace config file lookup for any command:

```bash
agnt --workspace-config ~/other-project/.agnt.yaml validate
agnt --workspace-config /tmp/test.yaml new-workspace
```

Errors if the specified file does not exist (except with `new-workspace`, which creates it).

---

## Troubleshooting

### When do pane indices change?

Pane indices are stable in most situations, but a few operations will shift them and leave your
config pointing at the wrong panes:

| Operation | Effect on indices | What to do |
|-----------|------------------|------------|
| Detach and reattach | Stable | Nothing |
| Add a new window | Stable | Nothing |
| tmux-resurrect save/restore | Stable | Nothing |
| **Split a registered pane** | All panes after the split point are re-indexed | Run `agnt remap` |
| **Reorder windows** | Window numbers of affected agents change | Run `agnt remap` |
| **Kill server and recreate manually** | Indices likely differ | Run `agnt remap` |
| Close a registered pane | Pane is gone | Remove or remap the agent |

`agnt validate` reports OK as long as the index *number* exists somewhere in the current
session — it does not verify that the right process is running there.

### pane-base-index

If your `.tmux.conf` sets `pane-base-index` to a non-zero value (e.g. `1`), `agnt` captures and
uses whatever index tmux assigns — so a single-pane window at base-index 1 registers as `0.1`.
This works transparently; just be aware that your indices won't start at `0`.

---

## License

Copyright 2025 David Tanzer (business@davidtanzer.net)

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
