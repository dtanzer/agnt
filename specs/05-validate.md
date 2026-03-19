# Step 05 — Validate

## Goal

`agnt validate` gives the user a quick health check of their workspace: does a config file exist, where is it, is it valid, and are the registered panes still alive?

---

## Command

```
agnt validate
agnt --workspace-config <file> validate
```

---

## Output

Prints one line per check, then a summary. Example:

```
Workspace:  /home/user/project/.agnt.yaml
Syntax:     OK
Agents (2):
  Bob       pane %1  OK  [claude]
  Alice     pane %2  OK  [bash]
  Charlie   pane %3  MISSING

Summary: 2 checks passed, 1 failed
```

- Each agent line shows the pane ID, whether it exists (OK / MISSING), and if it exists, what command is currently running in it (informational only — any command is acceptable).
- If showing the running command is not straightforward to implement, add a note to PLAN.md for later instead.

---

## Checks performed

1. **Workspace found** — a config file exists (via flag or directory walk). If not, print an error and exit non-zero immediately.
2. **Syntax valid** — the file parses as valid YAML with the expected structure.
3. **Per-agent pane check** — for each registered agent, verify the tmux pane ID still exists.

---

## Acceptance Criteria

- If no workspace is found (and no `--workspace-config` given), prints an error to stderr and exits non-zero.
- If `--workspace-config` points to a missing file, prints an error to stderr and exits non-zero.
- If the file exists but is invalid YAML, reports the syntax check as failed and exits non-zero.
- For each agent, reports OK if the pane exists, MISSING if not.
- If all checks pass, exits with code 0.
- If any check fails, exits with code 1.
- The running command in each pane is shown as informational (best effort — if unavailable, omit silently).
