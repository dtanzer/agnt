# Step 10 — `agnt start`

## Goal

`agnt start <name>` launches a registered agent by sending its resolved command as keystrokes to the correct tmux pane.

---

## Command

```
agnt start <name>
agnt --workspace-config <file> start <name>
```

---

## Behaviour

1. Find the workspace config (same lookup as other commands).
2. Look up `<name>` in `agents:`. Error if not found.
3. Look up the agent's type in `types:`. Error if not defined.
4. Resolve placeholders in `run:` using the agent's name and variant. Error if an unknown placeholder is found.
5. Check that the agent's pane exists in the current tmux session. Error if missing.
6. Send the resolved command as keystrokes to the pane, followed by Enter.
7. Print a confirmation line to stdout.

### Example output

```
Starting Bob in pane 0.0: echo Bob
```

---

## Acceptance Criteria

- `agnt start Bob` with a valid config sends `echo Bob` + Enter to the registered pane and prints the confirmation line.
- If `<name>` is not in the config, prints an error to stderr and exits non-zero.
- If the agent's type is not defined in `types:`, prints an error to stderr and exits non-zero.
- If the resolved `run:` command contains an unknown placeholder, prints an error to stderr and exits non-zero. No keystrokes are sent.
- If the agent's pane does not exist in the current tmux session, prints an error to stderr and exits non-zero. No keystrokes are sent.
- `agnt start` with no `<name>` argument prints a usage error to stderr and exits non-zero.
