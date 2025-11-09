## Phase 1 Complete: Add Failing Tests for Postflop Action Order

Successfully added unit tests demonstrating that GetFirstActor() should return different players based on the current street (preflop vs postflop). Tests initially failed (as expected), exposing the bug where the function didn't check h.Street.

**Files created/changed:**
- `internal/server/table_test.go` - Added 3 unit tests

**Functions created/changed:**
- None (Phase 1 added tests only, no production code changes)

**Tests created/changed:**
- `TestGetFirstActor_Postflop_MultiPlayer` (new) - Tests that SB should act first on flop in multi-player game
- `TestGetFirstActor_Postflop_HeadsUp` (new) - Tests that BB (non-dealer) should act first on flop in heads-up
- `TestGetFirstActor_Postflop_WithFoldedSB` (new) - Tests that BB acts first when SB has folded on flop

**Review Status:** APPROVED

All three tests failed initially (demonstrating the bug):
- Multi-player test: Expected seat 1 (SB), got seat 3 (UTG)
- Heads-up test: Expected seat 2 (BB), got seat 0 (dealer)
- Folded SB test: Expected seat 2 (BB), got seat 3 (UTG)

**Git Commit Message:**
```
test: Add failing tests for postflop action order bug

- Add test for multi-player postflop (SB should act first on flop)
- Add test for heads-up postflop (BB should act first on flop)
- Add test for folded SB edge case (BB acts first when SB folded)
- Tests demonstrate GetFirstActor doesn't check Street field
```
