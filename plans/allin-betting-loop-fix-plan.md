## Plan: Fix All-In Betting Round Infinite Loop

The bug occurs when players go all-in with unequal stacks - the betting round doesn't complete because it only checks if bets match, not if players can act. This causes an infinite action loop. The fix must work for 2-player, 3-player, and multi-player scenarios across all streets (preflop, flop, turn, river).

**Root Cause:** `IsBettingRoundComplete()` checks if `PlayerBet == CurrentBet` but doesn't account for all-in players (stack = 0) who can't match higher bets.

**Solution:** Check `stack == 0` to skip all-in players when determining betting round completion and action eligibility.

**Phases (5 phases):**

---

## **Phase 1: Write Failing Tests for IsBettingRoundComplete**

**Objective:** Create comprehensive tests that expose the bug in betting round completion logic for 2-player, 3-player, and multi-player scenarios

**Files/Functions to Modify/Create:**
- `internal/server/table_test.go` - Add `TestIsBettingRoundComplete_AllInScenarios` test group

**Tests to Write:**
- `TestIsBettingRoundComplete_TwoPlayerBothAllInUnequalStacks` - 2 players: SB has 900, BB has 1000, both all-in
- `TestIsBettingRoundComplete_TwoPlayerOneAllInOneMatched` - 2 players: One player all-in, other matched bet
- `TestIsBettingRoundComplete_ThreePlayerTwoAllInOneActive` - 3 players: Two all-in (different stacks), one active with matched bet
- `TestIsBettingRoundComplete_ThreePlayerAllDifferentStacks` - 3 players: All-in with 500, 700, 1000 stacks
- `TestIsBettingRoundComplete_MultiPlayerSomeAllInSomeFolded` - 5 players: 2 all-in, 2 folded, 1 active matched
- `TestIsBettingRoundComplete_AllPlayersAllIn` - All remaining players are all-in

**Steps:**
1. Write test: Two players, SB has 900 total, BB has 1000 total, both go all-in preflop
2. Set up hand state: `CurrentBet=1000`, `PlayerBets[SB]=900`, `PlayerBets[BB]=1000`, `seats[SB].Stack=0`, `seats[BB].Stack=0`
3. Call `IsBettingRoundComplete()` and assert it returns `true`
4. Run test - expect **FAIL** (currently returns `false` because 900 != 1000)
5. Write test: Three players with different all-in amounts (500, 700, 1000)
6. Write test: Multi-player with mix of all-in, folded, and active players
7. Run all new tests - expect **FAIL** for all-in scenarios

---

## **Phase 2: Fix IsBettingRoundComplete to Handle All-In Players**

**Objective:** Modify betting round completion logic to skip all-in players (stack = 0) when checking if bets are matched

**Files/Functions to Modify/Create:**
- `internal/server/table.go` - Modify `IsBettingRoundComplete()` (lines 1504-1571)

**Tests to Write:** None (tests already exist from Phase 1)

**Steps:**
1. Locate the loop in `IsBettingRoundComplete()` that checks if players matched current bet (around line 1552)
2. Add condition to skip all-in players: `if seats[seatNum].Stack == 0 { continue }`
3. Logic should be: Only non-folded AND non-all-in players need to match current bet
4. Run Phase 1 tests - expect **PASS** for all scenarios (2-player, 3-player, multi-player)
5. Run full `table_test.go` suite - expect all existing tests still **PASS**

---

## **Phase 3: Fix GetValidActions and Write Tests for All Streets**

**Objective:** Prevent all-in players (stack = 0) from receiving any action options, with tests covering preflop, flop, turn, and river

**Files/Functions to Modify/Create:**
- `internal/server/table.go` - Modify `GetValidActions()` (lines 1146-1186)
- `internal/server/table_test.go` - Add `TestGetValidActions_AllInPlayers` test group

**Tests to Write:**
- `TestGetValidActions_AllInPlayerZeroStackPreflop` - Preflop: stack=0 returns empty actions
- `TestGetValidActions_AllInPlayerZeroStackFlop` - Flop: stack=0 returns empty actions
- `TestGetValidActions_AllInPlayerZeroStackTurn` - Turn: stack=0 returns empty actions
- `TestGetValidActions_AllInPlayerZeroStackRiver` - River: stack=0 returns empty actions
- `TestGetValidActions_AllInPlayerWithCallAmount` - Even with call amount > 0, stack=0 returns empty
- `TestGetValidActions_AllInPlayerWithRaise` - Even with raise available, stack=0 returns empty

**Steps:**
1. Write failing tests for all streets (preflop, flop, turn, river) with stack=0
2. Assert each test expects `GetValidActions()` to return `[]` (empty slice)
3. Run tests - expect **FAIL**
4. Add check at start of `GetValidActions()`: `if playerStack == 0 { return []string{} }`
5. Run tests - expect **PASS** for all streets
6. Run full `table_test.go` suite - expect all tests **PASS**

---

## **Phase 4: Fix GetNextActiveSeat and Test Multi-Player Scenarios**

**Objective:** Exclude all-in players from action rotation in 2-player, 3-player, and multi-player scenarios

**Files/Functions to Modify/Create:**
- `internal/server/table.go` - Modify `GetNextActiveSeat()` (lines 1079-1125)
- `internal/server/table_test.go` - Add `TestGetNextActiveSeat_AllInPlayers` test group

**Tests to Write:**
- `TestGetNextActiveSeat_TwoPlayerOneAllIn` - 2 players: Skip all-in player, return active player
- `TestGetNextActiveSeat_TwoPlayerBothAllIn` - 2 players: Both all-in, return `nil`
- `TestGetNextActiveSeat_ThreePlayerOneAllIn` - 3 players: Skip all-in player in rotation
- `TestGetNextActiveSeat_ThreePlayerTwoAllInOneActive` - 3 players: Skip 2 all-in, return 1 active
- `TestGetNextActiveSeat_MultiPlayerMixedStates` - 5 players: Mix of active, folded, all-in
- `TestGetNextActiveSeat_OnlyAllInPlayersRemaining` - All remaining players all-in, return `nil`

**Steps:**
1. Write failing test: 3 players, one all-in (stack=0), verify all-in player is skipped in rotation
2. Write test: All remaining players all-in, assert function returns `nil`
3. Run tests - expect **FAIL**
4. Modify condition in `GetNextActiveSeat()`: Add `&& seats[i].Stack > 0` to active player check
5. Updated condition: `if seats[i].Status == "active" && !h.FoldedPlayers[i] && seats[i].Stack > 0`
6. Run tests - expect **PASS** for all multi-player scenarios
7. Run full `table_test.go` suite - expect all tests **PASS**

---

## **Phase 5: Integration Tests for Full Scenarios Across All Streets**

**Objective:** End-to-end tests for exact bug scenario from issue #6, plus multi-player and multi-street scenarios

**Files/Functions to Modify/Create:**
- `internal/server/handlers_test.go` - Add `TestHandlePlayerAction_AllInBettingLoop` test group

**Tests to Write:**
- `TestHandlePlayerAction_TwoPlayerBothAllInPreflop` - Exact bug scenario from issue #6
- `TestHandlePlayerAction_TwoPlayerAllInFlop` - Both all-in on flop street
- `TestHandlePlayerAction_TwoPlayerAllInTurn` - Both all-in on turn street
- `TestHandlePlayerAction_TwoPlayerAllInRiver` - Both all-in on river street
- `TestHandlePlayerAction_ThreePlayerTwoAllInPreflop` - 3 players: 2 all-in preflop, verify 3rd doesn't get re-prompted
- `TestHandlePlayerAction_MultiPlayerSequentialAllIns` - 5 players: Sequential all-ins with different stacks

**Steps:**
1. Write integration test for exact bug from issue #6:
   - Setup: 2 players, SB posts 10, BB posts 20
   - SB calls 10 (total bet 20)
   - BB all-in 1000 (stack goes to 0)
   - SB calls with only 900 available (stack goes to 0)
   - Assert: No more `action_request` messages sent
   - Assert: Betting round completes, advances to flop
2. Write integration test for 3-player preflop all-in:
   - Setup: 3 players (Button, SB, BB)
   - BB all-in 500, Button calls 500, SB has 300 and goes all-in
   - Button and BB both have more chips
   - Assert: Button gets prompted again to call additional 200 (correct)
   - Button calls, betting round completes
3. Write tests for flop, turn, river all-in scenarios
4. Write test for multi-player sequential all-ins (5 players with stacks 100, 200, 300, 500, 1000)
5. Run all integration tests - expect **PASS** (all unit fixes in place)
6. Run full backend test suite (`make test-backend`) - expect all 314+ tests **PASS**
7. Run lint (`make lint`) - expect **PASS**

---

## **Open Questions - ANSWERED**

1. **Should GetNextActiveSeat return nil or skip to next street when all remaining players are all-in?**
   - **Answer:** Return `nil` (Option A) - cleaner separation of concerns, let caller advance to next street

2. **Should we add defensive check in HandlePlayerAction to reject actions from all-in players?**
   - **Answer:** No (Option B) - rely on `GetValidActions()` returning empty array as single source of truth

3. **Do we need to handle all-in players in other streets (flop, turn, river)?**
   - **Answer:** The fixes apply to all streets automatically since they're in generic functions
   - **Action:** Add explicit tests for each street (preflop, flop, turn, river) to verify
