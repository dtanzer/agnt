# Step 06 — Pane Index Migration

## Goal

Replace tmux pane IDs (`%42`) with window:pane indices (`0.1`) as the way agents are located in the config. Indices are visible in the tmux UI (`Ctrl+B q`), stable within a session layout, and meaningful to users.

---

## Config format change

The `pane` field in `.agnt.yaml` changes from a tmux pane ID to a `window.pane` index:

```yaml
# Before
agents:
  Bob:
    type: simple
    pane: "%42"

# After
agents:
  Bob:
    type: simple
    pane: "0.1"
```

The `.` separator matches tmux's own target syntax (`window.pane`), so values can be passed directly to tmux commands without conversion.

---

## Changes to `register`

Instead of capturing `#{pane_id}`, captures `#{window_index}.#{pane_index}` for the current pane.

---

## Changes to `validate`

- Lists all current panes using `tmux list-panes -a -F '#{window_index}.#{pane_index}'`
- Checks each agent's pane field against that list
- Retrieves the running command with `tmux display-message -p -t <pane> '#{pane_current_command}'`

---

## Acceptance Criteria

- `agnt register Bob simple` inside tmux writes a pane value like `0.1` (not `%42`).
- `agnt validate` correctly identifies existing and missing panes by index.
- The running command is still shown per agent in `validate` output.
- The fake tmux in `testtools/bin/tmux` is updated to return index-style values.
- A test fixture using index-style pane values is added to `test-workspaces/`.
