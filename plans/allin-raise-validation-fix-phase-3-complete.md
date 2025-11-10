## Phase 3 Complete: Backend - Verify Side Pot Logic (2-6 Players)

Successfully implemented and verified comprehensive side pot handling logic for 2-6 player games with multiple all-in scenarios. All 8 test cases cover edge cases including effective all-ins, equal stacks, single and multiple side pots, and whale overbetting scenarios.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- GetMaxOpponentCoverage() - NEW: Calculates maximum chips active opponents can cover for bet capping
- ProcessActionWithSeats() - NEW: Wrapper around ProcessAction() that caps bets for side pot support

**Tests created/changed:**
- TestSidePots_2P_EffectiveAllIn - NEW: One player bets more than other can cover
- TestSidePots_2P_BothAllIn - NEW: Both go all-in with equal stacks
- TestSidePots_3P_OneAllInCreatesSidePot - NEW: One short stack all-in creates side pot
- TestSidePots_3P_AllDifferentStacks - NEW: Three different stacks create multiple side pots
- TestSidePots_3P_ShortestWinsMainPotOnly - NEW: Shortest stack wins and gets main pot only
- TestSidePots_4P_MultipleAllIns - NEW: Multiple players all-in with different amounts
- TestSidePots_6P_WhaleExcessReturned - NEW: Whale bets 5000, others ~1000, excess returned
- TestSidePots_6P_MultipleSidePots - NEW: Complex scenario with multiple side pots

**Review Status:** APPROVED

**Git Commit Message:**
```
test: Add comprehensive side pot tests for multi-player all-in scenarios

- Add 8 new test cases covering 2-6 player side pot scenarios
- Test cases: 2P effective all-in, 2P both all-in, 3P one all-in creates side pot
- Test cases: 3P all different stacks, 3P shortest wins main pot only
- Test cases: 4P multiple all-ins, 6P whale excess returned, 6P multiple side pots
- Add GetMaxOpponentCoverage() to calculate bet caps for side pot support
- Add ProcessActionWithSeats() wrapper for side pot-aware bet processing
- Fix test pattern: Use raiseToAmount = stack + PlayerBets[i] for all-in raises
- Verify chip deduction after ProcessAction() calls
- All 314 backend tests passing
- All 244 frontend tests passing
```
