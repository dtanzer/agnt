# agnt

`agnt` is a named-agent session manager for tmux. It lets you define a fixed layout of AI agents — each with a name, a type, and a permanent pane — and manage them by name from anywhere, including from inside containers.

Inspired by [Lada Kesseler's talk](https://www.youtube.com/watch?v=_LSK2bVf0Lc&t=8125).

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

A workspace is a `.agnt.yaml` file that describes your agents. `agnt` searches for it by walking up from the current directory toward `$HOME` — so you can run `agnt` commands from anywhere inside your project tree.

```yaml
agents:
  Alice:
    type: simple
    variant: java
    pane: "0:0"
  Bob:
    type: simple
    pane: "0:1"
```

Each agent has:
- **name** — stable identifier used in all commands
- **type** — determines which launch script is used
- **variant** *(optional)* — passed as a parameter to the launch script
- **pane** — tmux window:pane index (e.g. `0:1`)

### Pane layout

The recommended workflow is to create a fixed tmux split layout first, then register your agents into it. The pane indices (shown in tmux when you press `Ctrl+B q`) stay stable as long as your layout doesn't change. If it does, use `agnt remap` to reassign agents to their new positions.

### Agent groups

Groups are alternative agent configurations for the same pane layout. Multiple agents can share the same pane index — but only one runs there at a time:

```yaml
agents:
  Alice:   { type: simple, pane: "0:0" }   # Group 1
  Bob:     { type: simple, pane: "0:1" }   # Group 1
  Charlie: { type: simple, pane: "0:0" }   # Group 2
  Dave:    { type: simple, pane: "0:1" }   # Group 2
```

You run either Group 1 (Alice+Bob) or Group 2 (Charlie+Dave) in the same two panes. Stop one group before starting the other. The tool doesn't need to know about groups — you decide which agents to start and stop.

---

## Commands

### `agnt new-workspace`

Creates a new `.agnt.yaml` in the current directory.

```bash
agnt new-workspace
# Created workspace at /home/user/project/.agnt.yaml
```

Errors if a `.agnt.yaml` already exists in the current directory. Use `--workspace-config` to create a file at an arbitrary path.

---

### `agnt register <name> <type> [--variant <variant>]`

Registers the current tmux pane as a named agent. Must be run from inside a tmux pane.

```bash
agnt register Alice simple
agnt register Bob simple --variant java
```

- Searches parent directories for an existing `.agnt.yaml`; creates one in the current directory if none is found.
- Errors if an agent with that name already exists.

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
  Alice        pane 0:0  OK  [claude]
  Bob          pane 0:1  OK  [bash]

Summary: 2/2 checks passed
```

Exits non-zero if any check fails (missing workspace, invalid YAML, missing pane).

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

## Typical workflow

### First-time setup

```bash
# 1. Create your tmux session and split it into your desired layout
tmux new-session -s agents
# ... split panes with Ctrl+B % and Ctrl+B "

# 2. Create a workspace in your project directory
cd ~/project
agnt new-workspace

# 3. In each pane, register the agent
#    (run this from inside each pane)
agnt register Alice simple
agnt register Bob simple --variant java

# 4. Verify everything looks right
agnt validate
```

### After reopening tmux

If your pane indices have changed (e.g. new tmux session), run:

```bash
agnt validate        # see what's missing
agnt remap           # reassign agents to current panes (coming soon)
```

### Working with multiple workspaces

Use `--workspace-config` to work with a specific config file without changing your current directory:

```bash
agnt --workspace-config ~/teamA/.agnt.yaml validate
agnt --workspace-config ~/teamB/.agnt.yaml register Dave simple
```

---

## License

Copyright 2025 David Tanzer (business@davidtanzer.net)

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
