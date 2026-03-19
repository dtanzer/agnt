# Step 11 — `agnt server`

## Goal

`agnt server start` runs a local HTTP server in the foreground that will later coordinate agent lifecycle. For now it just needs to start, stay running, and be discoverable by `agnt server status`.

---

## Commands

```
agnt server start [--port <port>]
agnt server status
```

---

## `agnt server start`

Starts an HTTP server listening on `localhost:<port>` (default port: 7717). Runs in the foreground — the user is expected to run it in a dedicated tmux pane.

- Errors and exits if not running inside a tmux session.
- Errors and exits if another server is already running (detected via the status file).
- On startup, writes a status file at a known location (see below).
- On shutdown (Ctrl-C / SIGTERM), removes the status file.
- Rejects connections from non-localhost addresses.
- Prints a startup message: `agnt server listening on localhost:7717`

### Status file

Written to the same directory as the workspace config, named `.agnt-server.yaml`. Contains:

```yaml
pid: 12345
port: 7717
started: "2026-03-19T15:00:00Z"
```

The status file is the only mechanism `agnt server status` and other commands use to discover the server.

---

## `agnt server status`

Reads the status file from the workspace directory. Checks whether the recorded PID is still running, then pings the server's `/health` endpoint to confirm it is responding.

Output when server is running and healthy:

```
Server:   running
PID:      12345
Address:  localhost:7717
Uptime:   4m32s
```

Output when no status file exists:

```
Server:   not running
```

Output when status file exists but PID is not running (stale file):

```
Server:   not running (stale status file)
```

Output when PID is alive but `/health` does not respond:

```
Server:   not healthy (process exists but not responding)
```

---

## Acceptance Criteria

- `agnt server start` outside tmux prints an error and exits non-zero.
- `agnt server start` inside tmux starts the server, writes the status file, and prints the startup message.
- `agnt server start` when a server is already running prints an error and exits non-zero.
- The server rejects (closes) connections from non-localhost addresses.
- Stopping the server (Ctrl-C / SIGTERM) removes the status file.
- `agnt server status` with a running, healthy server prints pid, address, and uptime.
- `agnt server status` with no status file prints `not running`.
- `agnt server status` with a stale status file (pid gone) prints `not running (stale status file)`.
- `agnt server status` when the process exists but `/health` does not respond prints `not healthy (process exists but not responding)`.
- The `/health` endpoint returns HTTP 200.
