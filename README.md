# agnt

`agnt` is a host-side server and CLI tool for managing named AI agent sessions in tmux panes.
The server runs on the host (inside the tmux session it manages), accepts REST commands from
anywhere that can reach it — including Claude/opencode instances running in Docker or Podman
containers — and executes the actual tmux operations on the host.

The CLI is a thin REST client that locates the running server via a session file and issues
commands to it.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  Host (tmux session)                                │
│                                                     │
│  ┌──────────────┐   REST    ┌───────────────────┐   │
│  │  agnt serve  │◄──────────│  agnt CLI         │   │
│  │  (HTTP server│           │  (thin client)    │   │
│  │   port N)    │           └───────────────────┘   │
│  │              │                                   │
│  │              │──tmux send-keys──► pane A (claude)│
│  │              │──tmux send-keys──► pane B (claude)│
│  └──────────────┘                                   │
│         ▲                                           │
│         │ REST /agents/{name}/register              │
│  ┌──────┴─────────────────────────────────────────┐ │
│  │  Container / remote process                    │ │
│  │  (agent script calls back with PID)            │ │
│  └────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

**Components:**

- **`agnt serve`** — foreground REST server, runs in the tmux session it will manage.
  Logs spawns, stops, and errors to stdout. Does not log individual messages.
- **`agnt <command>`** — thin CLI client. Discovers server port from the session file and
  issues REST requests.
