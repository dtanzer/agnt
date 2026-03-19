# agnt — Claude Code Instructions

## Workflow

Steps live in PLAN.md (Done / Current / Next / Later). Specs live in `specs/`.

Pattern: refine step(s) in PLAN.md → Claude writes spec from PLAN.md + user input → implement → repeat for next step.

## Collaboration style

- Ask clarifying questions when anything is unclear before proceeding.
- Suggest alternative approaches when you see meaningful options.
- Stop and ask if you get stuck rather than forcing through a block.

## Writing specs

When writing implementation specs:

- **User-focused.** Describe what the user experiences, not how the code works.
- **No implementation details** unless absolutely necessary to constrain a decision.
- **Acceptance criteria** are written from the user's point of view — observable behavior only.
