# agnt v2 — North Star

`agnt` makes it easy to run and coordinate multiple named AI agents (claude, opencode, etc.) across tmux panes, from anywhere — including from inside containers.

**The core idea:** A lightweight server runs on the host inside a tmux session. It owns the tmux operations. Everything else — agents in panes, agents in containers, other machines — talks to it over REST.

**What it gives you:**

- **Named agents with fixed identities.** You define agents in a config file. Each has a name, a type (which determines how it's launched), and a pane to live in. The name is stable; the process behind it comes and goes.

- **Lifecycle management.** Spawn an agent (server sends the launch script to the right pane), stop it gracefully (polite message → Ctrl+D → escalating signals), send it a message. The server handles all the tmux mechanics.

- **Pane arbitration.** If a pane is occupied, starting a new agent there stops the old one first automatically.

- **Container-aware stopping.** When an agent runs in a Docker or Podman container, stopping it means stopping the container — not trying to signal a PID that doesn't exist on the host.

- **Agent self-registration.** The server knows an agent is truly ready only when the agent's launch script calls back to register. This is the readiness signal.

- **A thin CLI.** `agnt spawn`, `agnt stop`, `agnt send`, `agnt list` — the CLI finds the running server automatically and just issues REST calls.

**In one sentence:** agnt is a named-agent session manager — it sits on the host, owns the tmux panes, and lets you spawn, stop, and message AI agents by name regardless of where the caller lives.

---

## Pane layouts and agent groups

Agents are assigned to tmux panes by window:pane index (e.g. `0:0`, `0:1`). These indices are stable within a tmux session layout and visible in the tmux UI, making it easy to reason about which agent lives where.

**Groups** are alternative agent configurations for the same layout. Multiple agents can share the same pane index — but only one runs there at a time. This lets you define several groups in a single config file and switch between them:

```
Group 1:  Alice → 0:0   Bob → 0:1
Group 2:  Charlie → 0:0   Dave → 0:1
```

You'd run Group 1 (Alice+Bob) or Group 2 (Charlie+Dave) in the same two panes, stopping one group before starting the other. The tool doesn't need to know about groups — it manages individual agents and you decide which ones to start and stop.

If your tmux layout changes (new session, panes reordered), `agnt remap` lets you re-assign agents to their new pane indices without editing the config by hand.

---

## Future ideas

These are not planned for now but worth keeping in mind:

- **Group commands** — `agnt start-group <name>`, `agnt stop-group <name>` to act on all agents sharing a pane or tagged with a group label in the config.
- **Watch mode** — a live dashboard showing agent status across all panes, refreshing automatically.
- **Remote agents** — agents running on other machines joining the same agnt server over the network.
- **Config generation** — `agnt init` that interactively assigns the current tmux panes to agent names and writes the config in one step.
