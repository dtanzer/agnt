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
