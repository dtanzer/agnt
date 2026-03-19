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

---

**Test 2 — Split a registered pane**
1. In one of the registered panes, split it: `Ctrl+B %`
2. Run: `agnt validate`

*Question: do the indices of the other 3 panes shift? Does the split pane's index change?*

---

**Test 3 — Close a registered pane**
1. Close one of the registered panes: `exit` or `Ctrl+D`
2. Run: `agnt validate`

*Question: do the indices of the remaining 3 panes shift?*

---

**Test 4 — Add a new window**
1. Open a new window: `Ctrl+B c`
2. Run: `agnt validate` (from the original window)

*Question: do the pane indices in window 0 change when a new window is added?*

---

**Test 5 — Reorder windows**
1. Swap the current window with the next: `Ctrl+B :swap-window -t 1`
2. Run: `agnt validate`

*Question: do pane indices within the moved window change?*

---

### Session restart

**Test 6 — Kill and recreate**
1. Kill the server: `tmux kill-server`
2. Start a new session and recreate the same 2×2 layout manually
3. Run: `agnt validate`

*Question: do you get the same indices (0.0–0.3) as before?*

---

### tmux-resurrect

**Test 7 — Save, kill, restore**
1. Save: `Ctrl+B Ctrl+S` (resurrect save)
2. Kill the server: `tmux kill-server`
3. Start a new tmux session and restore: `Ctrl+B Ctrl+R`
4. Run: `agnt validate`

*Question: are all 4 agents found with their original indices?*

---

**Test 8 — Reboot**
1. Reboot the machine (with resurrect configured to auto-restore, if set up)
2. After boot, run: `agnt validate`

*Question: same indices after a full reboot?*

---

### Config edge case

**Test 9 — Non-zero pane-base-index**
1. Check your `.tmux.conf` for `pane-base-index` — what is it set to (default is 0)?
2. If it's set to 1 (or another value), note what indices `agnt register` captures.

*Question: does agnt capture the actual index tmux uses, regardless of base-index setting?*

---

## What to report back

For each test: pass/fail, and any unexpected behaviour. Particularly interested in:
- Whether splits or closures cause index shifts
- Whether resurrect reliably restores indices
- Whether `pane-base-index` affects what agnt captures
