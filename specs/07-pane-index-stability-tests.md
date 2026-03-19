# Step 07 — Pane Index Stability Tests

## Goal

Verify how stable tmux window.pane indices are across common tmux operations and tmux-resurrect save/restore cycles. Results will inform whether `agnt remap` is needed often or rarely, and whether any edge cases need handling in the tool.

This is not an implementation task — it's a conversation between user and Claude. The user runs each test on their machine and reports back; Claude interprets the results and updates the plan accordingly.

---

## Setup (run once before all tests)

1. Create a tmux session with a 2×2 layout:
   ```
   tmux new-session -s agnt-test
   # Split into 4 panes: Ctrl+B %, Ctrl+B ", etc.
   ```
2. Create a workspace and register one agent per pane:
   ```
   agnt new-workspace
   agnt register TopLeft simple      # in pane 0.0
   agnt register TopRight simple     # in pane 0.1
   agnt register BottomLeft simple   # in pane 0.2
   agnt register BottomRight simple  # in pane 0.3
   ```
3. Confirm baseline: `agnt validate` — all 4 should pass.

Record the pane indices from the config as your baseline.

---

## Tests

### Within a session

**Test 1 — Detach and reattach**
1. Detach: `Ctrl+B d`
2. Reattach: `tmux attach -t agnt-test`
3. Run: `agnt validate`

*Question: do all 4 agents still show OK?*

**Result: PASS.** Indices are stable across detach/reattach.

---

**Test 2 — Split a registered pane**
1. In one of the registered panes, split it: `Ctrl+B %`
2. Run: `agnt validate`

*Question: do the indices of the other 3 panes shift? Does the split pane's index change?*

**Result: indices shift; validate gives false OK.** Splitting a pane re-indexes all panes after the split point. The config still holds the original indices which now resolve to different panes — validate reports OK because those index slots still exist, but the mapping is wrong. Closing the newly created pane restores the original indices (tested; may not always hold for more complex combinations).

*Decision: document this behaviour. Users must not split registered panes without running `remap` afterward.*

---

**Test 3 — Close a registered pane**
1. Close one of the registered panes: `exit` or `Ctrl+D`
2. Run: `agnt validate`

*Question: do the indices of the remaining 3 panes shift?*

**Result: BUG FOUND.** `validate` searches panes across all sessions (`tmux list-panes -a`), so if a matching pane index exists in *any* session, it reports OK even though the registered pane is gone. Example: closing pane 0.2 in `agnt-test` still passed validation because `dragon-trainer 0.2` existed in another session.

*Decision: `validate` must scope pane lookup to the current tmux session only.*

---

**Test 4 — Add a new window**
1. Open a new window: `Ctrl+B c`
2. Run: `agnt validate` (from the original window)

*Question: do the pane indices in window 0 change when a new window is added?*

**Result: PASS.** Adding a new window does not affect pane indices in existing windows.

---

**Test 5 — Reorder windows**
1. Swap the current window with the next: `Ctrl+B :swap-window -t 1`
2. Run: `agnt validate`

*Question: do pane indices within the moved window change?*

**Result: indices shift; validate gives false OK.** Window reordering changes the window component of every affected agent's index (e.g. `0.0` becomes `1.0`). Validate reports OK only if those index slots exist in some window of the session.

*Decision: document that users must run `remap` after reordering windows.*

---

### Session restart

**Test 6 — Kill and recreate**
1. Kill the server: `tmux kill-server`
2. Start a new session and recreate the same 2×2 layout manually
3. Run: `agnt validate`

*Question: do you get the same indices (0.0–0.3) as before?*

**Result: indices may differ; validate gives false OK.** After kill+recreate, manually recreating the layout can produce different pane index assignments. Validate reported OK because the index *numbers* in the config still existed — but the physical panes behind those indices were different agents (order had changed).

*Decision: kill+recreate is a full remap situation. Document this and have `validate` suggest `remap` when it detects a session age mismatch or when the user reports layout changes.*

---

### tmux-resurrect

**Test 7 — Save, kill, restore**
1. Save: `Ctrl+B Ctrl+S` (resurrect save)
2. Kill the server: `tmux kill-server`
3. Start a new tmux session and restore: `Ctrl+B Ctrl+R`
4. Run: `agnt validate`

*Question: are all 4 agents found with their original indices?*

**Result: PASS.** tmux-resurrect reliably restores pane indices — visually confirmed the same panes are in the same positions and validate reports all agents OK. No remap needed after a resurrect cycle.

---

### Config edge case

**Test 8 — Non-zero pane-base-index**
1. Check your `.tmux.conf` for `pane-base-index` — what is it set to (default is 0)?
2. If it's set to 1 (or another value), note what indices `agnt register` captures.

*Question: does agnt capture the actual index tmux uses, regardless of base-index setting?*

**Result: PASS.** With `pane-base-index 1`, a single-pane window registered as `0.1` — agnt captured the actual tmux index. Validate also reported `pane 0.1  OK`. agnt is transparent to this setting; whatever index tmux assigns is what gets stored and looked up.

---

## Summary of findings

| Test | Result | Action needed |
|------|--------|---------------|
| 1 — Detach/reattach | STABLE | None |
| 2 — Split a pane | Indices shift; validate blind | Document; users must remap after splits |
| 3 — Close a pane | **BUG**: validate searches all sessions | Fix: scope pane lookup to current session |
| 4 — Add a window | STABLE | None |
| 5 — Reorder windows | Indices shift; validate blind | Document; users must remap after reorder |
| 6 — Kill and recreate | Indices may differ; validate blind | Document; remap required after recreate |
| 7 — tmux-resurrect | STABLE | None — resurrect reliably restores indices |
| 8 — pane-base-index | PASS | None — agnt transparently uses whatever index tmux assigns |

## What to report back

For each remaining test: pass/fail, and any unexpected behaviour. Particularly interested in:
- Whether resurrect reliably restores indices
- Whether `pane-base-index` affects what agnt captures
