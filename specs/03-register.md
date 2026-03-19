# Step 02 — Register

## Goal

`agnt register` lets a user claim a tmux pane as a named agent, recording it in a shared config file. This is the first step toward a world where agents have stable identities.

---

## Command

```
agnt register <name> <type> [--variant <variant>]
```

- `<name>` — the agent's name (e.g. `Bob`)
- `<type>` — the agent type, which determines which launch script will be used (e.g. `simple`)
- `--variant <variant>` — optional sub-type passed as a parameter to the launch script (e.g. `java`)

### Examples

```
agnt register Bob simple
agnt register Bob simple --variant java
```

---

## Config file

Agents are stored in `.agnt.yaml`. Example:

```yaml
agents:
  Bob:
    type: simple
    variant: java
    pane: "%3"
```

- `pane` is the tmux pane ID of the pane where the command was run.
- `variant` is omitted from the file when not provided.

### Config file lookup

When `agnt register` runs, it searches for `.agnt.yaml` by walking up from the current directory toward `$HOME`. The first file found is used. If no file is found, a new `.agnt.yaml` is created in the current directory.

---

## Acceptance Criteria

- Running `agnt register Bob simple` outside tmux prints an error to stderr and exits non-zero.
- Running `agnt register Bob simple` inside tmux adds `Bob` to the config file with `type: simple` and the current pane ID, and exits cleanly.
- Running `agnt register Bob simple --variant java` adds `Bob` with `type: simple` and `variant: java`.
- If `Bob` already exists in the config file, prints an error to stderr and exits non-zero. The file is not modified.
- If no `.agnt.yaml` exists in any parent directory up to `$HOME`, a new one is created in the current directory.
- If a `.agnt.yaml` exists in a parent directory, it is updated (not a new file created in the current directory).
- Running `agnt register` with a missing `<name>` or `<type>` prints a usage error to stderr and exits non-zero.
