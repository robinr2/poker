## Phase 2 Complete: Backend - Add Comprehensive Integration Tests

Successfully added comprehensive integration tests that verify auto-kick works correctly in real showdown scenarios with pot distribution, multiple simultaneous bust-outs, and edge cases.

**Files created/changed:**
- `internal/server/table_test.go`

**Functions created/changed:**
- N/A (tests only)

**Tests created/changed:**
- `TestShowdown_AllInPlayerBustsOut` - Single all-in player loses and gets auto-kicked
- `TestShowdown_MultiplePlayersBustOut` - Multiple all-in losers with zero stacks all get kicked simultaneously
- `TestShowdown_WinnerWithStackNotKicked` - Winner with remaining stack is NOT kicked (revised to fix tie scenario)
- `TestShowdown_AllInWinnerNotKicked` - All-in winner who receives pot is NOT kicked (edge case)

**Review Status:** ✅ APPROVED

**Implementation Highlights:**
- ✅ All tests use specific hole cards for deterministic outcomes (no random card dealing)
- ✅ Clear, specific assertions for each scenario (winner/loser stacks, seat status, token clearing)
- ✅ Comprehensive edge case coverage (single bust, multiple bust, winners not kicked, all-in winners not kicked)
- ✅ Realistic betting scenarios simulated (blinds, all-in, pot distribution)
- ✅ Fixed Test 3 tie issue (changed board from A-K-Q-J-T to 9-8-7-5-2 for deterministic winner)
- ✅ All 277 backend tests passing (4 new integration tests + 3 existing unit tests from Phase 1)

**Test Results:**
- TestShowdown_AllInPlayerBustsOut ✅
- TestShowdown_MultiplePlayersBustOut ✅
- TestShowdown_WinnerWithStackNotKicked ✅ (revised to fix tie scenario)
- TestShowdown_AllInWinnerNotKicked ✅

**Git Commit Message:**
```
test: Add comprehensive integration tests for auto-kick feature

- Add TestShowdown_AllInPlayerBustsOut for single bust-out scenario
- Add TestShowdown_MultiplePlayersBustOut for simultaneous bust-outs
- Add TestShowdown_WinnerWithStackNotKicked to verify winners stay seated
- Add TestShowdown_AllInWinnerNotKicked for all-in winner edge case
- Use specific hole cards (AA, KK, 22, 33, 44, 55, KQ) for deterministic outcomes
- Set specific board cards to guarantee predictable hand rankings
- Verify busted seats have Token=nil and Status="empty"
- Verify winners and losers with remaining stacks stay seated
- All 277 backend tests passing
```
