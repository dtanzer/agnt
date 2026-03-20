# Step 13 — `agnt start` delegates to the server

## Goal

`agnt start <name>` stops sending keystrokes directly and instead asks the server to do it. The server resolves the agent config, substitutes placeholders, and sends the keys. If no server is running, the command fails with a clear error.

---

## CLI changes

`agnt start <name>` now:

1. Discovers the server via the status file (same lookup as `agnt server status`).
2. If no status file is found, exits with an error: `agnt start requires a running server — run "agnt server start" first`
3. Calls `POST /agents/{name}/start` on the server.
4. On success, prints the line returned by the server (same format as before): `Starting {name} in pane {pane}: {command}`
5. On error, prints the error message from the server response and exits non-zero.

The `--workspace-config` flag continues to work for status file discovery (finding the server). The server uses its own workspace config for everything else.

---

## Server changes

### `POST /agents/{name}/start`

The server:

1. Reads the workspace config fresh from disk.
2. Looks up the agent by name — 404 if not found.
3. Looks up the agent's type — 500 if the type is not defined (config inconsistency).
4. Substitutes placeholders in the `run:` command.
5. Checks the pane exists in the current tmux session — 422 if not.
6. Sends the resolved command to the pane via `tmux send-keys`.
7. Returns HTTP 200 with JSON:

```json
{
  "name": "alice",
  "pane": "0.1",
  "command": "claude --some-flag --name alice"
}
```

Error responses use a JSON body: `{"error": "agent \"alice\" not found"}`.

---

## Acceptance Criteria

- `agnt start <name>` with no server running prints `agnt start requires a running server — run "agnt server start" first` and exits non-zero.
- `agnt start <name>` with a running server sends keys via the server and prints `Starting {name} in pane {pane}: {command}`.
- `agnt start <name>` for an unknown agent name prints the server's error and exits non-zero.
- `agnt start <name>` when the pane does not exist prints the server's error and exits non-zero.
- The old direct tmux path in `start.go` is removed.
- `POST /agents/{name}/start` returns 200 + JSON on success.
- `POST /agents/{name}/start` returns 404 JSON error for unknown agent.
- `POST /agents/{name}/start` returns 422 JSON error when the pane does not exist.
