## Phases 3-5 Complete: Pot Accounting Fix Implementation

Successfully implemented the core pot accounting fix where chips stay in PlayerBets during betting rounds and are swept to Pot at street transitions or showdown, plus fixed frontend all-in button and updated all legacy tests.

**Files created/changed:**

Backend:
- internal/server/table.go (AdvanceStreet, HandleShowdown sweep logic)
- internal/server/table_test.go (5 new Phase 3 tests, 20+ Phase 5 test updates)

Frontend:
- frontend/src/components/TableView.tsx (All-in button fix, game info display fix)
- frontend/src/components/TableView.test.tsx (Test updates)

**Functions created/changed:**

Backend:
- `AdvanceStreet()` - Added pot sweep logic before clearing PlayerBets
- `HandleShowdown()` - Added pot sweep logic at beginning for early winner case
- `GetMaxRaise()` - Returns playerBet + playerStack (from Phase 1)
- Test fixtures - 20+ tests updated for new pot accounting model

Frontend:
- `handleAllIn()` - Uses maxRaise from gameState instead of playerStack
- Removed unused `playerStack` variable
- Fixed game info/board cards display condition to work with pot=0 during betting

**Tests created/changed:**

New Phase 3 Tests (5):
- `TestAdvanceStreet_SweepsBetsIntoPot_Preflop` - Verifies sweep on first advance
- `TestAdvanceStreet_ClearsPlayerBets` - Verifies PlayerBets cleared after sweep
- `TestAdvanceStreet_AccumulatesPot` - Verifies pot accumulates across streets
- `TestFullHandPotAccounting_PreflopToRiver` - End-to-end pot accounting validation
- `TestHandleShowdown_EarlyWinner_UnsweptBets` - Early winner with unswept bets

Updated Phase 5 Tests (20+):
- `TestStartHandSetsPot` - Expects Pot=0 after blinds
- `TestProcessAction_Call` - Verifies PlayerBets instead of Pot
- `TestHandlePlayerAction_AllFoldPreflop_EarlyWinner` - Updated pot expectations
- `TestHandlePlayerAction_BigBlindWinsWithCheck` - Updated pot expectations
- `TestHandlePlayerAction_CallShowdown` - Updated pot expectations
- `TestHandlePlayerAction_RaiseCallShowdown` - Updated pot expectations
- `TestHandlePlayerAction_AllFoldFlop_EarlyWinner` - Updated pot expectations
- `TestHandlePlayerAction_AllFoldTurn_EarlyWinner` - Updated pot expectations
- `TestHandlePlayerAction_AllFoldRiver_EarlyWinner` - Updated pot expectations
- `TestShowdown_WinnerWithStackNotKicked` - Updated pot expectations
- `TestShowdown_AllInPlayerBustsOut` - Fixed PlayerBets accounting
- `TestShowdown_MultiplePlayersBustOut` - Fixed PlayerBets accounting
- `TestShowdown_AllInWinnerNotKicked` - Fixed PlayerBets accounting
- `TestSidePots_2P_EffectiveAllIn` - Updated pot expectations
- `TestSidePots_2P_BothAllIn` - Updated pot expectations
- `TestSidePots_3P_OneAllInCreatesSidePot` - Updated pot expectations
- `TestSidePots_3P_AllDifferentStacks` - Updated pot expectations
- `TestSidePots_3P_ShortestWinsMainPotOnly` - Updated pot expectations
- `TestSidePots_4P_MultipleAllIns` - Updated pot expectations
- `TestSidePots_6P_WhaleExcessReturned` - Asserts on PlayerBets
- `TestSidePots_6P_MultipleSidePots` - Asserts on PlayerBets

**Review Status:** APPROVED

**Test Results:**
- Backend: 311 tests passing, 0 failures ✅
- Frontend: 247 tests passing, 0 failures ✅

**Git Commit Message:**
```
fix: Correct pot accounting to sweep PlayerBets at street transitions

- Add pot sweep logic in AdvanceStreet() before clearing PlayerBets map
- Add pot sweep in HandleShowdown() to handle early winner edge case
- Fix frontend all-in button to use maxRaise (playerBet + stack)
- Fix frontend game info/board display to show when hand active (not pot > 0)
- Update 20+ backend tests to expect new deferred pot accounting behavior
- Chips now stay in PlayerBets during betting, swept to Pot at street advance/showdown
- Fixes bug where early winners received 0 chips instead of full pot
- Fixes bug where community cards not visible during preflop with new pot accounting
```
