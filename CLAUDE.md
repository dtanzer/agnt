# agnt — Claude Code Instructions

You were probably started by the agnt server. If so, the environment variable `AGENT_NAME` contains your name.

## Workflow

Steps live in PLAN.md (Done / Current / Next / Later). Specs live in `specs/`.

Pattern: refine step(s) in PLAN.md → Claude writes spec from PLAN.md + user input → implement → repeat for next step.

## Testing

For commands that require a tmux session, use the fake tmux binary at `testtools/bin/tmux` (returns pane ID `%42`):

```
TMUX=fake PATH=/workspace/testtools/bin:$PATH ./agnt <command>
```

Test fixtures in `test-workspaces/` are committed to the repo. After any test that modifies them, restore them before committing.

Each step's acceptance criteria must have a corresponding shell script in `tests/` (e.g. `tests/test-attach.sh`). Use port 17717 for test servers to avoid conflicting with a running dev server on the default port.

## Collaboration style

- Ask clarifying questions when anything is unclear before proceeding.
- Suggest alternative approaches when you see meaningful options.
- Stop and ask if you get stuck rather than forcing through a block.

## Writing specs

When writing implementation specs:

- **User-focused.** Describe what the user experiences, not how the code works.
- **No implementation details** unless absolutely necessary to constrain a decision.
- **Acceptance criteria** are written from the user's point of view — observable behavior only.
