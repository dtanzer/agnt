# agnt v2 — Specification

## Overview

`agnt` is a host-side server and CLI tool for managing named AI agent sessions in tmux panes.
The server runs on the host (inside the tmux session it manages), accepts REST commands from
anywhere that can reach it — including Claude/opencode instances running in Docker or Podman
containers — and executes the actual tmux operations on the host.

The CLI is a thin REST client that locates the running server via a session file and issues
commands to it.

---

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

---

## File Layout

### Workspace config — `.agnt.toml`

Searched upward from CWD, stopping at `$HOME`. The directory where it is found is the
**workspace root**. All other files are relative to the workspace root.

### Global config — `$HOME/.agnt/config.toml`

Defines agent types available system-wide. Workspace config overrides global config when both
define the same agent type name.

### Session file — `.agnt-session.json`

Written to the workspace root by `agnt serve` on startup. Records the server port and the
current state of all running agents. Deleted on clean server shutdown.

The CLI reads this file to discover the server port. If the file is absent, the CLI exits with
a clear error ("No agnt server session found — run `agnt serve` in your workspace.").

---

## Configuration

### Workspace config — `.agnt.toml`

```toml
# Agent definitions — each agent has a fixed identity and a pane assignment.
# Multiple agents may share the same pane; only one can run at a time in a pane.
# Starting an agent whose pane is occupied stops the current occupant first.

[[agents]]
name    = "Orchestrator"
type    = "elaborate"
subtype = "nodejs"
pane    = "top-left"      # pane identifier: TBD during implementation

[[agents]]
name    = "Researcher"
type    = "simple"
subtype = "nodejs"
pane    = "bottom-left"

[[agents]]
name    = "Tester"
type    = "simple"
subtype = "java"
pane    = "bottom-left"   # shares pane with Researcher

# Agent type definitions — override any matching entry in the global config.

[[agent-types]]
name   = "elaborate"
script = "start-elaborate-agent.sh"

[[agent-types]]
name   = "simple"
script = "start-simple-agent.sh"
```

### Global config — `$HOME/.agnt/config.toml`

```toml
[[agent-types]]
name   = "elaborate"
script = "/usr/local/bin/start-elaborate-agent.sh"

[[agent-types]]
name   = "simple"
script = "/usr/local/bin/start-simple-agent.sh"
```

### Config resolution rules

1. Agent definitions come from the workspace config only (no global agents).
2. Agent type definitions: workspace config entries take precedence over global entries
   with the same `name`.
3. If an agent references a type not defined anywhere, `spawn` fails with a clear error.

---

## Agent Scripts

When spawning, the server constructs a shell command and sends it to the target pane via
`tmux send-keys`. The command sets environment variables and invokes the configured script:

```
AGNT_NAME="Researcher" \
AGNT_SUBTYPE="nodejs" \
AGNT_SERVER_PORT="8080" \
AGNT_WORKSPACE_DIR="/home/user/myproject" \
AGNT_PANE="bottom-left" \
start-simple-agent.sh
```

Environment variables passed to every script:

| Variable           | Description                                           |
|--------------------|-------------------------------------------------------|
| `AGNT_NAME`        | Agent name as defined in config                       |
| `AGNT_SUBTYPE`     | Agent subtype as defined in config                    |
| `AGNT_SERVER_PORT` | Port the agnt server is listening on                  |
| `AGNT_WORKSPACE_DIR` | Absolute path to the workspace root                 |
| `AGNT_PANE`        | Pane identifier the agent is running in               |

**Script contract:** The script is expected to start the AI process (claude, opencode, or
other) and then register with the server by calling:

```
POST http://localhost:$AGNT_SERVER_PORT/agents/$AGNT_NAME/register
{"pid": <pid>}
```

If the AI process runs inside a Docker or Podman container, the in-container PID is not
reachable from the host via OS signals. In that case the script should instead register a
container reference, and the server will use runtime commands (`docker`/`podman`) to stop it:

```
POST http://localhost:$AGNT_SERVER_PORT/agents/$AGNT_NAME/register
{
  "pid": <host-pid-of-runtime-process-or-null>,
  "container": {
    "runtime": "docker",   // or "podman"
    "id": "<container-id-or-name>"
  }
}
```

`pid` and `container` are both optional, but at least one must be present. The server uses
whichever is provided to manage the process lifecycle (see Stop Flow).

The server waits up to 60 seconds for the registration callback before timing out and marking
the spawn as failed. This callback is also the mechanism by which the server knows the agent
is actually ready.

Example — AI runs directly on host:

```bash
#!/usr/bin/env bash
cd "$AGNT_WORKSPACE_DIR"
claude &
AGENT_PID=$!
curl -sf -X POST \
  "http://localhost:${AGNT_SERVER_PORT}/agents/${AGNT_NAME}/register" \
  -H "Content-Type: application/json" \
  -d "{\"pid\": ${AGENT_PID}}"
wait "$AGENT_PID"
```

Example — AI runs in a Docker container:

```bash
#!/usr/bin/env bash
cd "$AGNT_WORKSPACE_DIR"
CONTAINER_ID=$(docker run -d --rm myimage claude)
curl -sf -X POST \
  "http://localhost:${AGNT_SERVER_PORT}/agents/${AGNT_NAME}/register" \
  -H "Content-Type: application/json" \
  -d "{\"container\": {\"runtime\": \"docker\", \"id\": \"${CONTAINER_ID}\"}}"
docker wait "$CONTAINER_ID"
```

> **Note:** If a script does not call back (e.g. third-party script), the server will attempt
> to detect the agent PID by inspecting the pane's child processes as a fallback, but this is
> best-effort and may be unreliable.

---

## Session File Format

`.agnt-session.json` in the workspace root:

```json
{
  "serverPort": 8080,
  "tmuxSession": "main",
  "startedAt": "2026-03-19T10:00:00Z",
  "agents": {
    "Orchestrator": {
      "pid": 12345,
      "pane": "top-left",
      "status": "running",
      "spawnedAt": "2026-03-19T10:01:00Z"
    },
    "Researcher": {
      "pid": null,
      "pane": "bottom-left",
      "status": "stopped",
      "spawnedAt": null
    }
  }
}
```

Status values: `running`, `stopped`, `stopping` (stop in progress), `spawning` (waiting for
register callback).

---

## REST API

Base URL: `http://localhost:<port>`

All request and response bodies are JSON. All endpoints return appropriate HTTP status codes.

### Health

```
GET /health
→ 200 {"status": "ok", "uptime": "42s"}
```

### Agent list

```
GET /agents
→ 200 [
    {"name": "Orchestrator", "type": "elaborate", "subtype": "nodejs",
     "pane": "top-left", "pid": 12345, "status": "running", "spawnedAt": "..."},
    {"name": "Researcher", "type": "simple", "subtype": "nodejs",
     "pane": "bottom-left", "pid": null, "status": "stopped", "spawnedAt": null},
    ...
  ]
```

### Single agent

```
GET /agents/{name}
→ 200  (same shape as one element of the list above)
→ 404  {"error": "agent 'X' not defined in config"}
```

### Spawn

```
POST /agents/{name}/spawn
→ 202  {"message": "Spawning Researcher — waiting for register callback"}
→ 404  {"error": "agent 'X' not defined in config"}
→ 409  {"error": "agent 'X' is already running"}
→ 409  {"error": "pane 'bottom-left' is occupied by 'Researcher' — stopping first"}
         (this case proceeds: stop Researcher, then spawn)
```

Spawn is asynchronous. The 202 is returned immediately; status transitions to `running` once
the register callback arrives. Clients can poll `GET /agents/{name}` for status.

If the target pane is occupied, the server stops the occupant inline before spawning (no
separate client action needed). The response reflects this.

### Stop

```
POST /agents/{name}/stop
→ 202  {"message": "Stopping Researcher"}
→ 404  {"error": "agent 'X' not defined in config"}
→ 409  {"error": "agent 'X' is not running"}
```

Stop is asynchronous (the full stop sequence takes up to ~45s). Clients poll status.

### Send

```
POST /agents/{name}/send
{"message": "Please review the auth module.", "from": "Orchestrator"}

→ 200  {"message": "Sent"}
→ 404  {"error": "agent 'X' not defined in config"}
→ 409  {"error": "agent 'X' is not running"}
```

### Register (called by agent scripts, not by users)

```
POST /agents/{name}/register
{"pid": 12345}

→ 200  {"message": "Registered"}
→ 404  {"error": "agent 'X' not defined in config"}
→ 409  {"error": "agent 'X' is not in spawning state"}
```

---

## CLI Commands

All commands search upward from CWD for the session file to find the server port.
`agnt serve` is the only command that does not need a session file.

```
agnt serve              Start the server (foreground). Reads workspace config.
                        Writes session file on startup, deletes it on clean exit.

agnt spawn <name>       Spawn a configured agent. Prints status updates.

agnt stop <name>        Initiate graceful stop. Prints status updates.

agnt send <name> <msg>  Send a message to a running agent.
                        Sender name is taken from AGNT_NAME env var if set,
                        otherwise "user".

agnt list               List all configured agents with name, type, subtype,
                        pane, pid, and status.

agnt status             Show server info: port, tmux session, uptime,
                        workspace root, session file path.
```

---

## Spawn Flow (server-side)

1. Receive `POST /agents/{name}/spawn`.
2. Look up agent in config → get type, subtype, pane. Return 404 if not found.
3. If agent status is `running` or `spawning` → return 409.
4. Check if another agent is `running` or `spawning` in the same pane.
   If so, run full stop flow for that agent (inline, blocking).
5. Look up agent type → get script name (workspace overrides global). Error if not found.
6. Set agent status to `spawning` in session file.
7. Build shell command with env vars (see Agent Scripts section).
8. Send command to pane via `tmux send-keys -t <pane-id> "<command>" Enter`.
9. Return 202 to caller.
10. Wait up to 60s for `POST /agents/{name}/register` callback.
    - On callback: record PID, set status to `running`, update session file, log event.
    - On timeout: set status to `stopped`, log failure.

---

## Stop Flow (server-side)

The stop sequence has two variants depending on what was registered: a host PID, a container
reference, or both. Steps 1–5 and 11 are identical in both cases.

1. Receive `POST /agents/{name}/stop`.
2. Look up agent. Return 404 if not in config, 409 if not running.
3. Get PID and/or container reference from session state.
4. Set status to `stopping` in session file. Return 202 to caller.
5. Send message via tmux send-keys:
   `"Please wrap up your current task and exit."` + Enter.

**If a container reference is registered:**

6. Poll container status every 2s for up to 30s. If container has exited → done (step 11).
7. Send Ctrl+D to pane via `tmux send-keys -t <pane-id> C-d`.
8. Poll container status every 2s for up to 5s. If exited → done (step 11).
9. Run `docker stop <id>` / `podman stop <id>` (sends SIGTERM to container's main process,
   waits its own timeout, then SIGKILLs). Default docker/podman stop timeout applies.
10. If container still running: `docker kill <id>` / `podman kill <id>`.

**If only a host PID is registered:**

6. Poll PID every 2s for up to 30s. If process has exited → done (step 11).
7. Send Ctrl+D to pane via `tmux send-keys -t <pane-id> C-d`.
8. Poll PID every 2s for up to 5s. If exited → done (step 11).
9. Send SIGTERM to PID.
10. Poll PID every 2s for up to 5s. If still alive → send SIGKILL.

11. Set status to `stopped`, clear PID/container in session file. Log event.
    Pane remains open (shell stays alive after the AI process exits).

---

## Send Flow (server-side)

1. Receive `POST /agents/{name}/send`.
2. Look up agent. Return 404 if not in config, 409 if not running.
3. Get pane ID for agent.
4. Format message: `#[<from>] <message>`.
5. Send via `tmux send-keys -t <pane-id> C-u` (clear current input line).
6. Send via `tmux send-keys -t <pane-id> "<formatted-message>" Enter`.

---

## Server Startup

1. Search upward from CWD for `.agnt.toml`. Exit with error if not found.
2. Load and validate workspace config. Merge with global config for agent types.
3. Check that `$TMUX` is set (must be run inside a tmux session). Record tmux session name.
4. Bind to configured port (default: 8080). If port is in use, try next available or exit
   with error. (Port configurability: `--port` flag or `server.port` in config — TBD.)
5. Write session file to workspace root.
6. Log: `agnt server started on port <N> (workspace: <path>, tmux: <session>)`.
7. On SIGINT/SIGTERM: attempt graceful stop of all running agents, delete session file, exit.

---

## Open Items / Deferred

- **Pane identifiers**: The format of the `pane` field in config (e.g. `"top-left"`, a tmux
  pane title, a position index) is TBD. Pane titles can change unexpectedly even with
  `automatic-rename off`; a robust persistent identification mechanism may be needed.
  Decision deferred to implementation.

- **Port configuration**: Whether the port is fixed, random, or configurable via flag/config
  is TBD. Suggest `--port` flag with default 8080 for initial implementation.

- **Config file format**: TOML is specified here; open to YAML if the Go TOML library proves
  awkward.

- **Graceful stop timeouts**: 30s / 5s / 5s are initial defaults. Should be configurable in
  the server config.

- **Container / remote support**: The server runs on the host. Agents in containers reach it
  via `host.docker.internal:<port>` (Docker) or equivalent. No special server changes needed.
  Nice-to-have: running the server itself in a privileged container with tmux socket access.

- **Wrap-up message**: The exact text sent to the agent on stop is a starting point.
  Experimentation needed.
