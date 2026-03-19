# Step 08 — Agent Type Definitions

## Goal

Users can define what each agent type means in their workspace config. A type maps a name (like `simple` or `elaborate`) to a command that will eventually be used to launch an agent. `agnt validate` verifies that every registered agent has a known type.

---

## Config format

Types are defined in `.agnt.yaml` under a `types:` key. Each entry has a `run:` field containing the command to execute when starting an agent of that type.

```yaml
types:
  simple:
    run: "claude --dangerously-skip-permissions"
  elaborate:
    run: "docker run --rm my-claude-image"

agents:
  Bob:
    type: simple
    pane: "0.0"
  Alice:
    type: elaborate
    pane: "0.1"
```

- The `types:` section is optional. A config with no `types:` section is still valid.
- A type name can be any string. There are no reserved names.
- The `run:` value is treated as a raw command string for now (no placeholder substitution yet — that comes in the next step).

---

## Changes to `agnt validate`

Validate gains a new check after syntax: **type resolution** — for each registered agent, verify its type is defined in the `types:` section.

Example output with a missing type:

```
Workspace:  /home/user/project/.agnt.yaml
Syntax:     OK
Types (1):
  elaborate  OK
Agents (2):
  Alice        pane 0.0  OK  [claude]
  Bob          pane 0.1  MISSING TYPE (simple)

Summary: 2/3 checks passed, 1 failed
```

- The types section is listed between Syntax and Agents, showing each type that is referenced by at least one agent and whether it resolves.
- If the `types:` section is absent entirely and agents reference types, each agent is reported as `MISSING TYPE`.
- If the `types:` section is absent and no agents reference any type, no types block is shown and no check is run.

---

## Acceptance Criteria

- A `.agnt.yaml` with a `types:` section parses correctly; `agnt validate` reports OK for all agents whose type is defined.
- A `.agnt.yaml` with no `types:` section and no agents is still valid; `agnt validate` reports OK with no types block shown.
- If an agent's type is not defined in `types:`, `agnt validate` reports that agent as `MISSING TYPE (typename)` and exits non-zero.
- If an agent's type is defined, the type check passes regardless of what the `run:` value contains (no validation of the command string itself).
- Type names are case-sensitive: `Simple` and `simple` are different types.
- Defining a type that no agent uses is allowed and produces no warning.
- The `agnt validate` summary count includes the type resolution check in its passed/failed totals.
