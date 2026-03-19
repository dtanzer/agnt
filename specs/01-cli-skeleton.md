# Step 01 — CLI Skeleton

## Goal

Deliver a working `agnt` binary that a user can install and invoke. It doesn't manage agents yet, but it feels like a real tool: it has a consistent command structure, useful help text, and an `info` command that tells you what version you're running.

This step establishes the CLI shape that all future commands will slot into.

---

## Commands

### `agnt help`

Prints a summary of all available commands with a one-line description of each.
Also triggered by `agnt --help`, `agnt -h`, and `agnt` with no arguments.

### `agnt info`

Prints version and build information about the installed binary. Should include at minimum:

- Version number
- Build date / commit
- Go version used to build it
- The binary's own path on disk

---

## Acceptance Criteria

- Running `agnt` with no arguments prints help and exits cleanly (exit code 0).
- Running `agnt help` or `agnt --help` or `agnt -h` prints help and exits cleanly.
- Running `agnt info` prints version and build info and exits cleanly.
- Running `agnt <unknown-command>` prints a short error message and exits with a non-zero code.
- All output goes to stdout; errors go to stderr.
- The binary is named `agnt`.
