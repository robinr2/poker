## Plan: Side Pot Distribution Fix (With Comprehensive Testing)

**TL;DR:** Fix the showdown payout bug where all-in players can win more than they contributed. Add `TotalContributions` tracking to the `Hand` struct to record cumulative chip contributions across all streets, then implement proper side pot calculation and distribution. Test exhaustively across 2-6 player scenarios with varying stack sizes, all-in patterns, and winner combinations.

**Phases: 6**

### 1. **Phase 1: Add TotalContributions Tracking Infrastructure**
   - **Objective:** Add a new field to track cumulative chip contributions per player across all streets, ensuring contributions are recorded but never cleared until hand completion.
   - **Files/Functions to Modify/Create:**
     - `internal/server/table.go`: Modify `Hand` struct to add `TotalContributions map[int]int` field
     - `internal/server/table.go`: Modify `StartHand()` to initialize `TotalContributions`
   - **Tests to Write:**
     - `TestTotalContributions_InitializedOnStartHand` - Verify TotalContributions initialized empty
     - `TestTotalContributions_TracksBlindPosting` - Verify blinds are tracked
     - `TestTotalContributions_PersistsAcrossStreets` - Verify not cleared on AdvanceStreet
   - **Steps:**
     1. Write failing tests that verify TotalContributions field exists and is initialized
     2. Add `TotalContributions map[int]int` field to Hand struct (line ~40)
     3. Initialize `TotalContributions = make(map[int]int)` in StartHand() where other Hand fields are initialized
     4. Run tests to confirm they pass

### 2. **Phase 2: Track Contributions During Betting**
   - **Objective:** Update all betting operations (blinds, calls, raises, all-ins) to record chip amounts in `TotalContributions` as they are wagered, ensuring accurate cumulative tracking.
   - **Files/Functions to Modify/Create:**
     - `internal/server/table.go`: Modify `StartHand()` to track blind contributions
     - `internal/server/table.go`: Modify `ProcessAction()` to track bet contributions (call and raise cases)
     - `internal/server/table.go`: Modify `ProcessActionWithSeats()` to track bet contributions
   - **Tests to Write:**
     - `TestTotalContributions_TracksCallAmount` - Verify calls are tracked
     - `TestTotalContributions_TracksRaiseAmount` - Verify raises are tracked (incremental, not absolute)
     - `TestTotalContributions_TracksAllInAmount` - Verify all-ins are tracked
     - `TestTotalContributions_AccumulatesAcrossStreets` - Verify contributions sum preflop→flop→turn→river
     - `TestTotalContributions_HandlesMultipleRaisesPerPlayer` - Player raises multiple times in same street
   - **Steps:**
     1. Write failing tests for each betting action contribution tracking
     2. In StartHand(), add `h.TotalContributions[sbSeat] = sb` after posting small blind
     3. In StartHand(), add `h.TotalContributions[bbSeat] = bb` after posting big blind
     4. In ProcessAction()/ProcessActionWithSeats() call action, calculate chips moved and add to TotalContributions
     5. In ProcessAction()/ProcessActionWithSeats() raise action, calculate incremental chips and add to TotalContributions
     6. Run tests to confirm tracking works correctly

### 3. **Phase 3: Implement Side Pot Calculation Algorithm**
   - **Objective:** Create a helper function that takes contribution data and calculates multiple side pots with eligible winners for each pot, handling all player count scenarios (2-6 players).
   - **Files/Functions to Modify/Create:**
     - `internal/server/table.go`: Create new struct `SidePot` with fields: `Amount int`, `EligibleSeats []int`
     - `internal/server/table.go`: Create new function `CalculateSidePots(contributions map[int]int, foldedPlayers map[int]bool) []SidePot`
   - **Tests to Write:**
     - **2-Player Tests:**
       - `TestCalculateSidePots_2P_EqualStacks` - Both contribute same → one pot
       - `TestCalculateSidePots_2P_OneShortStack` - One all-in with less → main pot only
     - **3-Player Tests:**
       - `TestCalculateSidePots_3P_OneShortStack` - One all-in, two continue → main + side pot
       - `TestCalculateSidePots_3P_TwoShortStacks` - Two different all-ins → main + 2 side pots
       - `TestCalculateSidePots_3P_OneFolded` - One folds after contributing → chips in pot, not eligible
     - **4-Player Tests:**
       - `TestCalculateSidePots_4P_MultipleAllIns` - Multiple all-in levels → multiple side pots
       - `TestCalculateSidePots_4P_TwoFolded` - Two fold, two all-in different amounts
     - **5-Player Tests:**
       - `TestCalculateSidePots_5P_LadderAllIns` - 5 different stack sizes all-in (100, 200, 300, 400, 500)
     - **6-Player Tests:**
       - `TestCalculateSidePots_6P_ComplexScenario` - Mix of folds, equal stacks, and different all-ins
       - `TestCalculateSidePots_6P_OneWhale` - One huge stack vs 5 short stacks
     - **Edge Cases:**
       - `TestCalculateSidePots_AllEqualContributions` - No side pots needed
       - `TestCalculateSidePots_AllFoldedExceptOne` - Only one player remains
       - `TestCalculateSidePots_ZeroContributions` - Some players contribute 0 (fold preflop)
   - **Steps:**
     1. Write failing tests for all scenarios above
     2. Define `SidePot` struct with `Amount int` and `EligibleSeats []int` fields
     3. Implement `CalculateSidePots()` algorithm:
        - Sort players by contribution amount (ascending)
        - Build pot levels based on contribution tiers
        - Assign eligible seats to each pot (players who contributed at that level or more)
     4. Handle edge cases: folded players contribute but aren't eligible, zero contributions, single player
     5. Run all tests to confirm algorithm produces correct pot structures

### 4. **Phase 4: Rewrite DistributePot with Side Pot Distribution**
   - **Objective:** Replace simple pot split logic with side pot calculation and distribution, ensuring each pot is awarded only to eligible winners with comprehensive testing across all player counts.
   - **Files/Functions to Modify/Create:**
     - `internal/server/table.go`: Modify `DistributePot()` signature to accept `contributions map[int]int` and `foldedPlayers map[int]bool`
     - `internal/server/table.go`: Rewrite `DistributePot()` logic to use `CalculateSidePots()` and distribute each pot to eligible winners
   - **Tests to Write:**
     - **2-Player Payouts:**
       - `TestDistributePot_2P_ShortStackWins` - Short stack wins, gets 2×their contribution
       - `TestDistributePot_2P_BigStackWins` - Big stack wins, gets full pot
       - `TestDistributePot_2P_Split` - Tie, both win, split pot
     - **3-Player Payouts:**
       - `TestDistributePot_3P_ShortestWins` - Shortest stack wins main pot only
       - `TestDistributePot_3P_MiddleStackWins` - Middle stack wins main + first side pot
       - `TestDistributePot_3P_BiggestWins` - Biggest stack wins all pots
       - `TestDistributePot_3P_TwoWinnersSplitMain` - Two winners tie, split main pot only
       - `TestDistributePot_3P_ShortAndBigTie` - Short and big stack tie, short gets main, both split side pot
     - **4-Player Payouts:**
       - `TestDistributePot_4P_SecondShortestWins` - Wins main + side pot 1, not side pot 2
       - `TestDistributePot_4P_MultipleWinnersDifferentPots` - Winners eligible for different pot combinations
     - **5-Player Payouts:**
       - `TestDistributePot_5P_LadderWithMiddleWinner` - Middle stack wins their eligible pots
     - **6-Player Payouts:**
       - `TestDistributePot_6P_ComplexMultiWaySplit` - 3 winners, different eligibilities
       - `TestDistributePot_6P_WhaleWinsEverything` - Biggest stack wins all side pots
     - **Edge Cases:**
       - `TestDistributePot_OddChips_DistributedCorrectly` - Remainder goes to first winner
       - `TestDistributePot_AllFoldedExceptOne` - One player wins by default (no showdown needed)
   - **Steps:**
     1. Write failing tests for all distribution scenarios above
     2. Change `DistributePot()` signature to accept `contributions` and `foldedPlayers` parameters
     3. Call `CalculateSidePots()` to get pot structures
     4. For each side pot, filter winners to only those in `EligibleSeats`
     5. Distribute each pot among its eligible winners (split evenly, handle odd chips)
     6. Run all tests to confirm correct payouts in all scenarios

### 5. **Phase 5: Integrate Side Pot Distribution into HandleShowdown**
   - **Objective:** Update the showdown flow to pass contribution data to DistributePot and verify integration works end-to-end with full hand simulations.
   - **Files/Functions to Modify/Create:**
     - `internal/server/table.go`: Modify `HandleShowdown()` to pass `TotalContributions` and `FoldedPlayers` to `DistributePot()`
   - **Tests to Write:**
     - **2-Player Integration:**
       - `TestHandleShowdown_2P_AllInScenario` - Full hand with all-in, verify stacks
     - **3-Player Integration:**
       - `TestHandleShowdown_3P_OneSidePot` - Full hand with side pot creation and distribution
       - `TestHandleShowdown_3P_ShortStackWins` - Verify short stack gets correct amount
     - **4-Player Integration:**
       - `TestHandleShowdown_4P_MultipleSidePots` - Full hand with 3 all-ins at different levels
     - **5-Player Integration:**
       - `TestHandleShowdown_5P_MixedFoldsAndAllIns` - Some fold, some all-in, verify all stacks
     - **6-Player Integration:**
       - `TestHandleShowdown_6P_ComplexScenario` - Full 6-player hand with multiple side pots
     - **Multi-Street Integration:**
       - `TestHandleShowdown_AllInPreflopThenFlop` - Players all-in on different streets
       - `TestHandleShowdown_BettingAcrossAllStreets` - Contributions accumulate preflop→river
   - **Steps:**
     1. Write failing integration tests for complete showdown scenarios
     2. In `HandleShowdown()`, ensure final PlayerBets are added to TotalContributions before sweeping
     3. Update `DistributePot()` call to pass `t.CurrentHand.TotalContributions` and `t.CurrentHand.FoldedPlayers`
     4. Run integration tests to verify end-to-end flow
     5. Run all existing table tests to ensure no regressions

### 6. **Phase 6: Validate with Existing Side Pot Tests and Add Payout Verification**
   - **Objective:** Extend existing side pot tests to validate correct payouts at showdown, ensuring the system works correctly with real game scenarios already in the test suite.
   - **Files/Functions to Modify/Create:**
     - `internal/server/table_test.go`: Enhance `TestSidePots_2P_EffectiveAllIn` with showdown payout validation
     - `internal/server/table_test.go`: Enhance `TestSidePots_3P_ShortestWinsMainPotOnly` with payout validation
     - `internal/server/table_test.go`: Enhance `TestSidePots_4P_MultipleAllIns` with payout validation
     - `internal/server/table_test.go`: Enhance `TestSidePots_6P_MultipleSidePots` with payout validation
   - **Tests to Write:**
     - Extend 4-5 existing side pot tests to include:
       - Deal board cards
       - Set hole cards to determine winner
       - Call `HandleShowdown()`
       - Assert exact final stack amounts for each player
       - Assert pot is empty after distribution
   - **Steps:**
     1. Identify existing side pot tests to enhance (TestSidePots_2P_*, TestSidePots_3P_*, etc.)
     2. For each test, add showdown logic after betting completes
     3. Set specific hole cards and board to control winner
     4. Call HandleShowdown() or trigger showdown through game flow
     5. Add assertions for exact final stack amounts
     6. Run all tests to confirm side pot distribution works with existing test infrastructure
     7. Run full test suite (`make test`) to ensure no regressions anywhere

**Open Questions:**
1. **Should we track contributions for folded players?** Yes - they contributed chips that go into the pot(s), but they are not eligible to win.
2. **How to handle odd chips in split pots?** Award remainder to the first winner in position order (standard poker rule).
3. **What if a player contributes across multiple streets with different amounts?** Sum all contributions - TotalContributions accumulates across all streets.
