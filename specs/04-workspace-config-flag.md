# Step 04 — Global --workspace-config Flag

## Goal

Give users a way to explicitly point `agnt` at a config file, bypassing the automatic directory-walk lookup. Useful for testing, scripting, and working with multiple workspaces.

---

## Flag

```
agnt --workspace-config <file> <command> [command args]
```

- Must appear before the subcommand.
- Applies to all commands that read or write the config file.

### Examples

```
agnt --workspace-config ~/projects/myteam/.agnt.yaml register Bob simple
agnt --workspace-config /tmp/test.yaml validate
agnt --workspace-config ./other.yaml new-workspace
```

---

## Behaviour

- If `--workspace-config` is given, the specified file is used directly — no directory walk is performed.
- If the file does not exist, `agnt` prints an error to stderr and exits non-zero. It does not create the file.
- Exception: `agnt --workspace-config <file> new-workspace` creates the file at the specified path (consistent with `new-workspace`'s purpose), but still errors if the file already exists.
- Without `--workspace-config`, behaviour is unchanged.

---

## Acceptance Criteria

- `agnt --workspace-config missing.yaml validate` prints an error and exits non-zero.
- `agnt --workspace-config existing.yaml register Bob simple` uses `existing.yaml` instead of searching parent directories.
- `agnt --workspace-config new.yaml new-workspace` creates `new.yaml` if it doesn't exist.
- `agnt --workspace-config existing.yaml new-workspace` errors if `existing.yaml` already exists.
- All existing commands (`register`, `validate`, future commands) respect the flag.
- Without the flag, all existing lookup behaviour is unchanged.
