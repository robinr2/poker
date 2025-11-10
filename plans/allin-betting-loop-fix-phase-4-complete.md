## Phase 4 Complete: Fix GetNextActiveSeat and Test Multi-Player

Successfully fixed `GetNextActiveSeat()` to skip all-in players (stack = 0) in action rotation and return `nil` when only all-in players remain. Added comprehensive multi-player tests and fixed regression tests.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- `GetNextActiveSeat()` (line 1089) - Added Stack > 0 check to skip all-in players

**Tests created/changed:**
- TestGetNextActiveSeat_AllInScenarios (7 subtests):
  - two_players_one_allin - Skip all-in, return nil (only 1 active)
  - two_players_both_allin - Both all-in returns nil
  - three_players_one_allin - Skip all-in in 3-player rotation
  - three_players_two_allin - Two all-in returns nil (only 1 active)
  - four_players_mixed_allin_folded - Skip all-in and folded
  - all_folded_except_allin - Return nil when only all-in remains
  - no_allin_normal_rotation - Control test
- TestAdvanceAction - Fixed regression by adding Stack initialization
- TestAdvanceAction_WithFoldedPlayers - Fixed regression by adding Stack initialization

**Review Status:** APPROVED

**Git Commit Message:**
```
fix: Skip all-in players in GetNextActiveSeat action rotation

- Add Stack > 0 check to skip players who are all-in (stack = 0)
- Return nil when only all-in players remain (no more actions needed)
- Add 7 comprehensive tests covering 2, 3, 4+ player scenarios
- Fix regression in TestAdvanceAction tests with Stack initialization
- All 352 backend tests pass with zero regressions
```
