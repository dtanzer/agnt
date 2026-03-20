# Step 14 — `agnt attach --podman <container-name>`

## Goal

`agnt attach` tells the server that the agent running in the current tmux pane has started a podman container. The server records the container name so it can act on the agent later. `agnt server status` shows which agents are currently attached.

---

## CLI

`agnt attach --podman <container-name>`

The agent name is not passed explicitly — it is inferred from the current tmux pane by looking up which agent in the workspace config is registered to that pane.

1. Requires a tmux session (`$TMUX` must be set). If not, exits with: `agnt attach must be run inside a tmux session`
2. Discovers the server via the status file. If no server is running, exits with: `agnt attach requires a running server — run "agnt server start" first`
3. Determines the current pane index from tmux.
4. Looks up the workspace config to find the agent registered to that pane. If none found, exits with: `no agent registered to the current tmux pane`
5. Calls `POST /agents/{name}/attach` with the container name.
6. On success, prints: `Attached {name} (podman: {container-name})`
7. On error, prints the server's error message and exits non-zero.

Intended use in a start script, before `podman run`:

```bash
SUFFIX=$(tr -dc 'a-z0-9' < /dev/urandom | head -c 8)
CONTAINER_NAME="${AGENT_NAME}-${SUFFIX}"

agnt attach --podman "$CONTAINER_NAME"

podman run -it --rm \
  --name "$CONTAINER_NAME" \
  ...
```

---

## Server changes

### In-memory agent state

The server gains an in-memory map of attached agents, keyed by agent name. Each entry records the agent name and its podman container name. This state is not persisted — it is cleared when the server restarts.

### `POST /agents/{name}/attach`

Request body:
```json
{ "podman": "alice-x3k2m9f7" }
```

1. Looks up the agent name in the workspace config — 404 if not found.
2. Stores the entry in the in-memory map, replacing any existing entry for that name.
3. Logs the attach event: `attached "alice" (podman: alice-x3k2m9f7)`
4. Returns HTTP 200 with JSON:
```json
{ "name": "alice", "podman": "alice-x3k2m9f7" }
```

### `GET /health` — updated response

Adds an `agents` field listing currently attached agents:
```json
{
  "pid": 123,
  "port": 7717,
  "started": "...",
  "uptime": "6m25s",
  "agents": [
    { "name": "alice", "podman": "alice-x3k2m9f7" }
  ]
}
```

Empty array when no agents are attached.

### `agnt server status` — updated output

```
Server:   running
PID:      123
Address:  localhost:7717
Uptime:   6m25s
Agents:   alice (podman: alice-x3k2m9f7)
```

Multiple agents, one per line, indented to align:
```
Agents:   alice (podman: alice-x3k2m9f7)
          bob (podman: bob-9f2a1c3d)
```

No agents attached:
```
Agents:   none
```

---

## Acceptance Criteria

- `agnt attach --podman <name>` outside a tmux session prints the error and exits non-zero.
- `agnt attach --podman <name>` with no server running prints the error and exits non-zero.
- `agnt attach --podman <name>` from a pane with no registered agent prints the error and exits non-zero.
- `agnt attach --podman alice-x3k2m9f7` from alice's pane succeeds and prints `Attached alice (podman: alice-x3k2m9f7)`.
- `agnt server status` after attaching shows the agent name and container name.
- `agnt server status` with no attached agents shows `Agents:   none`.
- Attaching the same name twice replaces the previous entry.
- Restarting the server clears all attached agent state.
