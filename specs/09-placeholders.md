# Step 09 — Placeholder System for Type Definitions

## Goal

Type definitions can include placeholders in their `run:` command that are substituted with agent-specific values at start time. This lets a single type definition work for any agent name or variant without duplication.

---

## Placeholder syntax

Placeholders are written as `{{name}}` — double curly braces around a variable name. They appear anywhere in the `run:` string.

### Available placeholders

| Placeholder | Value |
|---|---|
| `{{name}}` | The agent's name as registered in `.agnt.yaml` |
| `{{variant}}` | The agent's variant, or empty string if none is set |

Using any other placeholder (e.g. `{{server_url}}`) is an error at start time.

### Example

```yaml
types:
  simple:
    run: "claude --agent {{name}}"
  elaborate:
    run: "docker run -e AGENT_NAME={{name}} -e VARIANT={{variant}} my-image"

agents:
  Bob:
    type: simple
    pane: "0.0"
  Alice:
    type: elaborate
    variant: go
    pane: "0.1"
```

Starting Bob sends: `claude --agent Bob`
Starting Alice sends: `docker run -e AGENT_NAME=Alice -e VARIANT=go my-image`

---

## Substitution rules

- Substitution is plain string replacement — no quoting or escaping is applied.
- `{{variant}}` with no variant set substitutes as empty string (e.g. `-e VARIANT=` ).
- An unknown placeholder in `run:` causes `agnt start` to exit with an error before sending any keys.
- A `run:` string with no placeholders is valid and used as-is.

---

## Changes to `agnt register`

To prevent broken substitutions, `agnt register` now rejects names and variants that contain spaces. This keeps placeholder substitution unambiguous without needing quoting logic.

- `agnt register "Bob Smith" simple` → error: name must not contain spaces
- `agnt register Bob simple --variant "java 17"` → error: variant must not contain spaces

---

## Acceptance Criteria

- `run: "echo {{name}}"` with agent `Bob` substitutes to `echo Bob`.
- `run: "echo {{variant}}"` with no variant set substitutes to `echo ` (empty string for variant).
- `run: "echo {{server_url}}"` causes an error at start time: unknown placeholder `{{server_url}}`.
- `run: "echo hello"` (no placeholders) is used as-is.
- `agnt register` rejects names containing spaces with an error message.
- `agnt register` rejects variants containing spaces with an error message.
