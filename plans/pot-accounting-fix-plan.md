## Plan: Fix Pot Accounting and All-In Raise Validation

Fix critical pot accounting bug where chips are added to pot immediately instead of being swept from PlayerBets at street transitions. Also fix GetMaxRaise calculation and frontend all-in button to correctly handle raises when player has already posted blinds/bets.

**Phases: 5**

### Phase 1: Fix GetMaxRaise Calculation
- **Objective:** Correct GetMaxRaise to return total chips player can commit (current bet + remaining stack) instead of just remaining stack
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` - GetMaxRaise() function (line ~1190)
  - `internal/server/table_test.go` - Add/update tests for GetMaxRaise
- **Tests to Write:**
  - `TestGetMaxRaise_WithExistingBet` - Player with SB posted (10) and 990 stack should return 1000
  - `TestGetMaxRaise_WithoutBet` - Player with 1000 stack and no bet should return 1000
  - `TestGetMaxRaise_AfterPartialCall` - Player who called 50 with 1000 stack should return 1000
- **Steps:**
  1. Write failing test `TestGetMaxRaise_WithExistingBet` expecting maxRaise = playerBet + stack
  2. Write failing test `TestGetMaxRaise_WithoutBet` expecting maxRaise = stack when no bet
  3. Write failing test `TestGetMaxRaise_AfterPartialCall` for call scenarios
  4. Run tests to confirm failures
  5. Update GetMaxRaise implementation to return `t.PlayerBets[seatIndex] + t.PlayerStacks[seatIndex]`
  6. Run tests to confirm they pass
  7. Run existing table_test.go to identify any broken tests
  8. Fix any broken tests that relied on old maxRaise behavior

### Phase 2: Fix Pot Accounting - Remove Immediate Additions (Including Blinds)
- **Objective:** Stop adding chips directly to Pot during StartHand and ProcessAction/ProcessActionWithSeats; only update PlayerBets during active betting rounds including blind posting
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` - StartHand() (line ~780) - Remove `hand.Pot = sbPosted + bbPosted`
  - `internal/server/table.go` - ProcessAction() (lines ~1302-1375)
  - `internal/server/table.go` - ProcessActionWithSeats() (lines ~1399-1483)
  - `internal/server/table_test.go` - Update tests checking pot during betting rounds
- **Tests to Write:**
  - `TestStartHand_NoPotUpdate` - After blinds posted, Pot should be 0, PlayerBets should have blind amounts
  - `TestProcessAction_Call_NoPotUpdate` - Calling should update PlayerBets but NOT Pot
  - `TestProcessAction_Raise_NoPotUpdate` - Raising should update PlayerBets but NOT Pot
  - `TestProcessActionWithSeats_Call_NoPotUpdate` - Same for WithSeats variant
  - `TestPotRemainsZero_DuringBettingRound` - Pot stays 0 during entire preflop betting round
- **Steps:**
  1. Write failing test `TestStartHand_NoPotUpdate` checking Pot = 0 after blinds posted
  2. Write failing test `TestProcessAction_Call_NoPotUpdate` checking Pot unchanged after call
  3. Write failing test `TestProcessAction_Raise_NoPotUpdate` checking Pot unchanged after raise
  4. Write failing test `TestProcessActionWithSeats_Call_NoPotUpdate` for WithSeats variant
  5. Write failing test `TestPotRemainsZero_DuringBettingRound` for full betting round
  6. Run tests to confirm failures
  7. Remove `hand.Pot = sbPosted + bbPosted` from StartHand (line ~780)
  8. Remove `t.Pot += chipsToBet` from ProcessAction Call branch (line ~1304)
  9. Remove `t.Pot += chipsToBet` from ProcessAction Raise branch (line ~1364)
  10. Remove `t.Pot += chipsToBet` from ProcessActionWithSeats Call branch (line ~1419)
  11. Run new tests to confirm they pass
  12. Run full test suite to identify broken tests expecting old pot behavior

### Phase 3: Fix Pot Accounting - Sweep PlayerBets at AdvanceStreet
- **Objective:** When advancing to next street (or showdown), sweep all PlayerBets into Pot before clearing PlayerBets map. This includes the first sweep from preflop to flop which will collect the blinds.
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` - AdvanceStreet() function (line ~1593)
  - `internal/server/table_test.go` - Add/update tests for pot sweep logic
- **Tests to Write:**
  - `TestAdvanceStreet_SweepsBetsIntoPot_Preflop` - After preflop betting, advancing to flop should sweep blinds + all bets into Pot
  - `TestAdvanceStreet_ClearsPlayerBets` - PlayerBets should be empty after sweep
  - `TestAdvanceStreet_AccumulatesPot` - Pot should accumulate across multiple streets (flop → turn → river)
  - `TestFullHandPotAccounting` - End-to-end test: blinds → preflop → flop → turn → river → showdown pot
- **Steps:**
  1. Write failing test `TestAdvanceStreet_SweepsBetsIntoPot_Preflop` checking Pot = sum of all PlayerBets after first street advance
  2. Write failing test `TestAdvanceStreet_ClearsPlayerBets` checking PlayerBets empty after advance
  3. Write failing test `TestAdvanceStreet_AccumulatesPot` checking pot accumulates correctly across multiple streets
  4. Write failing test `TestFullHandPotAccounting` for complete hand lifecycle
  5. Run tests to confirm failures
  6. Add sweep logic in AdvanceStreet before clearing PlayerBets:
     ```go
     // Sweep all PlayerBets into Pot
     for _, bet := range h.PlayerBets {
         h.Pot += bet
     }
     ```
  7. Run new tests to confirm they pass
  8. Run full test suite to identify any remaining broken tests

### Phase 4: Fix Frontend All-In Button
- **Objective:** Update handleAllIn to use maxRaise from gameState instead of playerStack
- **Files/Functions to Modify/Create:**
  - `frontend/src/components/TableView.tsx` - handleAllIn() function (line ~177)
  - `frontend/src/components/TableView.test.tsx` - Add/update tests for all-in behavior
- **Tests to Write:**
  - `test_handleAllIn_usesMaxRaise` - All-in button should set raiseAmount to maxRaise
  - `test_handleAllIn_withPostedBlind` - Player with posted SB should go all-in for total commitment
  - `test_handleAllIn_disabled_whenNoMaxRaise` - Button disabled when maxRaise unavailable
- **Steps:**
  1. Write failing test `test_handleAllIn_usesMaxRaise` expecting raiseAmount = maxRaise
  2. Write failing test `test_handleAllIn_withPostedBlind` simulating SB posted scenario
  3. Write failing test `test_handleAllIn_disabled_whenNoMaxRaise` for edge case
  4. Run tests to confirm failures
  5. Update handleAllIn to use `gameState?.maxRaise` instead of `playerStack`
  6. Run tests to confirm they pass
  7. Run full frontend test suite

### Phase 5: Fix All Affected Backend Tests
- **Objective:** Update all existing tests that make assertions on Pot values to reflect new pot accounting model
- **Files/Functions to Modify/Create:**
  - `internal/server/table_test.go` - Update all tests with pot assertions
  - Potentially other test files if they check pot values
- **Tests to Write:**
  - No new tests - only updating existing test expectations
- **Steps:**
  1. Run full backend test suite: `go test ./internal/server/...`
  2. Identify all failing tests related to pot accounting
  3. For each failing test, update expectations:
     - During betting rounds: expect Pot unchanged, PlayerBets updated
     - After AdvanceStreet: expect Pot increased by PlayerBets, PlayerBets cleared
     - At showdown: expect Pot contains all accumulated chips
  4. Update test assertions one by one
  5. Re-run tests after each batch of fixes
  6. Continue until all tests pass
  7. Run integration tests to ensure end-to-end flow works correctly

**Decisions Made:**
1. **Side pots**: Deferred to future work (tracked in `plans/future-todos.md`)
2. **Blind accounting**: Blinds will NOT be added to Pot immediately. Like all other bets, they remain in PlayerBets until the first AdvanceStreet (preflop → flop) sweeps them into the pot.
3. **WebSocket payloads**: No changes needed. Backend already sends `Pot` value directly from `table.CurrentHand.Pot` via `SendTableState()` in handlers.go, so frontend will automatically receive correct pot values.
