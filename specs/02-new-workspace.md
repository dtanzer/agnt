# Step 02 — New Workspace

## Goal

`agnt new-workspace` gives a user an explicit way to create a workspace config file in the current directory. Unlike `register` (which creates one as a side effect), this is a deliberate act: "this directory is an agnt workspace."

---

## Command

```
agnt new-workspace
```

No arguments.

---

## Config file

Creates `.agnt.yaml` in the current directory with an empty agents map:

```yaml
agents: {}
```

---

## Acceptance Criteria

- Running `agnt new-workspace` in a directory with no `.agnt.yaml` creates `.agnt.yaml` there and prints a confirmation message to stdout.
- Running `agnt new-workspace` in a directory that already has `.agnt.yaml` prints an error to stderr and exits non-zero. The existing file is not modified.
- Running `agnt new-workspace` when a `.agnt.yaml` exists only in a parent directory still creates a new `.agnt.yaml` in the current directory (does not modify the parent's file).
- The created file is valid YAML that `agnt validate` (future) will accept.
