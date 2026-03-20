# Step 12 — `agnt server status` via health endpoint

## Goal

`agnt server status` should determine liveness solely by querying the `/health` endpoint — no PID checks. The health endpoint returns all the data the status command needs. This makes `agnt server status` usable from inside a container running with `--network host` and the workspace directory mounted.

---

## Changes

### `/health` endpoint

Returns JSON instead of plain text:

```json
{
  "pid": 12345,
  "port": 7717,
  "started": "2026-03-19T15:00:00Z",
  "uptime": "4m32s"
}
```

HTTP 200 on success (unchanged).

### `agnt server status`

Discovery: reads the status file to find the port, constructs `http://localhost:{port}/health`, and hits that endpoint. All displayed data comes from the JSON response — the status file is used only for URL discovery.

Drops the PID-running check entirely.

**Output — server running and healthy** (unchanged):

```
Server:   running
PID:      12345
Address:  localhost:7717
Uptime:   4m32s
```

**Output — no status file found:**

```
Server:   not running
```

**Output — status file exists but health endpoint not responding:**

```
Server:   not running
Checked:  http://localhost:7717/health
```

The `Checked:` line tells the user what URL was tried, which is useful for diagnosing misconfigured ports or a crashed server with a stale status file.

Note: the "stale status file (pid gone)" case from step 11 is gone — if the health endpoint doesn't respond, the server is not running regardless of what the status file says.

---

## Container use

Containers running agents must use `--network host` and mount the workspace directory so they can read the status file. With host networking, `localhost` inside the container resolves to the host, so the port in the status file is valid.

---

## Acceptance Criteria

- `agnt server status` with a running, healthy server prints pid, address, and uptime (data sourced from `/health` response).
- `agnt server status` with no status file prints `Server: not running` with no `Checked:` line.
- `agnt server status` when the status file exists but `/health` does not respond prints `Server: not running` followed by `Checked: http://localhost:{port}/health`.
- `GET /health` returns HTTP 200 with a JSON body containing `pid`, `port`, `started`, and `uptime`.
- No PID-running check anywhere in the status flow.
