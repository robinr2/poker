## Plan: Fix All-In Raise Validation for Multi-Player Games

Fix raise validation logic to allow players to bet their full stack regardless of opponent stack sizes, and ensure correct UI display for 2-6 player games.

**Phases: 3**

### **Phase 1: Fix Raise Validation Logic - Multi-Player Support**

- **Objective:** Remove opponent stack check from raise validation, allow players to always bet their full stack in 2-6 player games
- **Files/Functions to Modify/Create:**
  - `frontend/src/components/TableView.tsx`:
    - Find and remove validation that checks `raiseAmount > opponentStack`
    - Update max raise calculation to only use `playerRemainingStack`
    - Remove any logic that grays out raise/all-in buttons based on opponent stacks
  - `frontend/src/components/TableView.test.tsx`:
    - Add comprehensive multi-player test coverage
- **Tests to Write:**
  - **2 Players:**
    - `test_2p_sb_allin_990_when_bb_has_980` - Core bug: SB can go all-in even when BB has less
    - `test_2p_both_allin_equal_stacks` - Both players go all-in preflop
    - `test_2p_short_stack_can_allin` - Short stack can go all-in regardless of big stack
  - **3 Players:**
    - `test_3p_one_short_stack_can_allin` - One player with 500, others with 1000
    - `test_3p_multiple_different_stacks` - All three have different stacks
    - `test_3p_whale_can_overbet_all` - One player with 5000 can bet full stack
  - **4 Players:**
    - `test_4p_multiple_allins_same_hand` - Multiple players go all-in sequentially
    - `test_4p_shortest_stack_all_can_bet_full` - All players can bet full stack regardless of shortest
  - **5 Players:**
    - `test_5p_multiple_callers_with_different_stacks` - Various stack sizes, all can act correctly
  - **6 Players:**
    - `test_6p_whale_overbets_everyone` - One huge stack overbets all others
- **Steps:**
  1. Write tests first (TDD): Create test cases for 2-6 player scenarios where players with larger stacks can go all-in
  2. Run tests to see them fail (current validation blocks this)
  3. Read `TableView.tsx` to find raise validation logic
  4. Remove opponent stack checks from validation
  5. Update max raise to only check player's own stack
  6. Run tests to confirm they pass
  7. Run full frontend test suite to verify no regressions

### **Phase 2: Fix UI Display - Multi-Player Calculations**

- **Objective:** Ensure "To Call", "Green $", and preset buttons display correctly in 2-6 player games
- **Files/Functions to Modify/Create:**
  - `frontend/src/components/TableView.tsx`:
    - Update `toCall` calculation to cap at player's remaining stack
    - Verify `greenDollar` (contribution display) updates correctly for all players
    - Ensure `currentBet` tracks highest contribution across all players
    - Fix Min/Pot/All-In button calculations for multi-player scenarios
- **Tests to Write:**
  - `test_call_amount_capped_at_remaining_stack_multiplayer` - When bet > stack, call = stack
  - `test_green_dollar_updates_for_all_players` - Shows correct contribution for each player
  - `test_current_bet_tracks_highest_contribution` - Current bet = max(all contributions)
  - `test_min_raise_correct_after_multiple_raises` - Min raise = current bet + last raise size
  - `test_pot_size_includes_all_contributions` - Pot display includes all player contributions
  - `test_call_button_shows_correct_amount_2p` - "Call 10" for SB vs BB
  - `test_call_button_shows_correct_amount_3p` - Correct amounts with 3 players
  - `test_allin_button_always_shows_remaining_stack` - All-in = remaining stack for all players
- **Steps:**
  1. Write tests for UI display calculations in multi-player scenarios
  2. Run tests to see which calculations are incorrect
  3. Fix toCall calculation: `Math.min(currentBet - playerContribution, playerStack)`
  4. Verify greenDollar shows `playerContributionThisStreet`
  5. Fix Min/Pot/All-In button values
  6. Run tests to confirm they pass
  7. Run full frontend test suite

### **Phase 3: Backend - Verify Side Pot Logic (2-6 Players)**

- **Objective:** Ensure backend correctly creates and distributes side pots when players with different stacks go all-in
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go`:
    - Review pot calculation logic
    - Verify side pot creation for multiple all-ins
    - Check excess bet handling (when player bets more than others can cover)
  - `internal/server/table_test.go`:
    - Add comprehensive side pot tests
- **Tests to Write:**
  - **2 Players:**
    - `TestSidePots_2P_EffectiveAllIn` - One player bets more than other can cover
    - `TestSidePots_2P_BothAllIn` - Both go all-in with equal stacks
  - **3 Players:**
    - `TestSidePots_3P_OneAllInCreatesSidePot` - One short stack all-in, creates side pot
    - `TestSidePots_3P_AllDifferentStacks` - Stacks of 2000, 1000, 500 - multiple side pots
    - `TestSidePots_3P_ShortestWinsMainPotOnly` - Shortest stack wins, gets main pot only
  - **4+ Players:**
    - `TestSidePots_4P_MultipleAllIns` - Multiple players all-in with different amounts
    - `TestSidePots_6P_WhaleExcessReturned` - Whale bets 5000, others ~1000, excess returned
    - `TestSidePots_6P_MultipleSidePots` - Complex scenario with 3-4 side pots
- **Steps:**
  1. Write backend tests for side pot scenarios (2-6 players)
  2. Run tests to see current behavior
  3. If tests fail, review and fix pot calculation logic in `table.go`
  4. Ensure side pots created correctly based on stack sizes
  5. Verify excess bets returned to players
  6. Verify showdown distributes pots correctly
  7. Run full backend test suite to ensure no regressions

---

## Open Questions

1. **Should we cap call display at player's stack or show full amount?**
   - **RESOLVED:** Show capped amount (what player can actually put in)
   - Example: Bet is 1000, player has 980 → Show "Call 980"

2. **How to handle "Pot" button when player can't cover pot-sized raise?**
   - **RESOLVED:** Pot button should calculate pot raise, but if result > player stack, it becomes same as all-in
   - Example: Pot raise = 500, player has 300 → Pot button puts in 300 (all-in)

3. **Should we display side pot information to users?**
   - **DECISION NEEDED:** Currently not showing, should we add this?
   - Could show "Main Pot: 1500, Side Pot: 800" or just total "Pot: 2300"

4. **Backend: Do we already have side pot logic?**
   - **TO INVESTIGATE:** Need to check if side pot handling exists and is correct
