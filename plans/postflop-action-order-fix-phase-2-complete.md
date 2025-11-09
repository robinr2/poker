## Phase 2 Complete: Fix GetFirstActor to Handle Postflop Streets

Updated GetFirstActor() to check h.Street field and apply correct action order rules for preflop vs postflop streets. This fixes the bug where UTG continued to act first on the flop instead of SB.

**Files created/changed:**
- `internal/server/table.go` - Modified GetFirstActor method

**Functions created/changed:**
- `GetFirstActor()` - Added street branching logic to return correct first actor based on preflop vs postflop

**Tests created/changed:**
- None (using Phase 1 tests)

**Review Status:** APPROVED

All Phase 1 tests now pass:
- ✅ TestGetFirstActor_Postflop_MultiPlayer (SB acts first on flop)
- ✅ TestGetFirstActor_Postflop_HeadsUp (BB acts first on flop in heads-up)
- ✅ TestGetFirstActor_Postflop_WithFoldedSB (BB acts first when SB folded)
- ✅ All 5 existing preflop tests still pass (no regressions)
- ✅ All 247 backend tests passing

**Implementation Details:**
- Added street check: `if h.Street == "preflop"` to branch logic
- Preflop branch: Preserved existing logic (UTG in multi-player, dealer in heads-up)
- Postflop branch: Implemented correct poker rules
  - Multi-player: Returns SB, or next active player if SB folded
  - Heads-up: Returns BB (non-dealer player)
  - Properly checks h.FoldedPlayers map to skip folded players

**Git Commit Message:**
```
fix: Update GetFirstActor to handle postflop action order

- Add street branching to check h.Street field
- Preserve preflop logic (UTG/dealer acts first)
- Add postflop logic (SB/BB acts first on flop/turn/river)
- Handle folded players correctly using h.FoldedPlayers map
- Multi-player postflop: SB acts first, skip if folded
- Heads-up postflop: BB (non-dealer) acts first
- All 247 backend tests passing
```
