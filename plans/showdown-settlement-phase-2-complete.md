## Phase 2 Complete: Showdown Trigger & Winner Detection

Successfully implemented showdown trigger logic and winner determination using the hand evaluator from Phase 1. The system now detects when showdown occurs (river betting completes or all but one player folds), evaluates all non-folded players' hands, and correctly identifies winner(s) including tie scenarios.

**Files created/changed:**
- internal/server/table.go (MODIFIED)
- internal/server/handlers.go (MODIFIED)
- internal/server/table_test.go (MODIFIED)
- internal/server/handlers_test.go (MODIFIED)

**Functions created/changed:**
- `Hand.DetermineWinner(seats []*Seat) (winners []int, winningRank *HandRank)` - Evaluates all non-folded players and returns winner seat indices
- `Table.HandleShowdown()` - Main showdown orchestration method (skeleton, will be extended in Phase 3)
- Modified `handlers.go` line ~1270 - Added showdown trigger when river betting completes
- Added early winner handling when all but one player folds

**Tests created/changed:**
- `TestDetermineWinner_SingleWinner_HighCard` - Single winner with high card
- `TestDetermineWinner_SingleWinner_Flush` - Single winner with flush
- `TestDetermineWinner_Tie_TwoPlayers` - Two-way tie with identical hands
- `TestDetermineWinner_Tie_ThreePlayers` - Three-way tie with identical hands
- `TestDetermineWinner_HeadsUp` - Heads-up showdown (2 players)
- `TestDetermineWinner_MultiWay_FourPlayers` - Four-way showdown
- `TestDetermineWinner_SkipsFoldedPlayers` - Folded players excluded from evaluation
- `TestHandleShowdown_TriggersOnRiverComplete` - HandleShowdown called on river completion
- `TestHandleShowdown_EarlyWinner_AllFold` - Single remaining player wins without evaluation
- `TestHandlerFlow_RiverToShowdown` - Full hand flow from deal to river to showdown
- `TestHandlerFlow_AllFoldBeforeShowdown` - All players fold, one wins early

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: Add showdown trigger and winner detection

- Implement DetermineWinner to evaluate non-folded players' hands
- Handle ties by returning multiple winners with identical hand ranks
- Add HandleShowdown orchestration method (skeleton for Phase 3)
- Trigger showdown when river betting completes
- Handle early winner case when all but one player folds
- Skip folded players from hand evaluation
- Add 11 tests covering single winner, ties, heads-up, multi-way scenarios
- All tests passing
```
