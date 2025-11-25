# E2E Test Scenarios

## Completed

### 1. Complete Hand to Showdown
**File:** `three-player-complete-hand.spec.ts`
- 3 players join and start a hand
- Preflop: UTG calls, SB calls, BB checks
- Flop: Check, Min Raise, Call, Fold
- Turn: Check, Check
- River: Check, Check
- Showdown: Winner determined and pot awarded
- **Tests:** Full betting flow, all streets, showdown mechanics

### 2. Player Fold to Win (No Showdown)
**File:** `three-player-fold-to-win.spec.ts`
- 3 players in hand
- Preflop: One player raises big
- Other two players fold
- Winner takes pot without showdown
- **Tests:** Fold-to-win logic, no showdown UI, pot awarded correctly

### 3. Deterministic Winner Tests
**File:** `deterministic-winner.spec.ts`
- Pocket Aces vs Pocket Kings heads-up
- Flush vs Three of a Kind
- 3-player showdown with Quads winner
- **Tests:** Deterministic deck, showdown hand ranking, pot distribution

### 4. Fold Scenarios (All Streets)
**File:** `fold-scenarios.spec.ts`
- 4-player sequential folds - raiser wins blinds
- All fold to Big Blind - BB wins SB uncontested
- 3-player fold on flop - bettor wins
- 2-player fold on turn - bettor wins
- 2-player fold on river - bettor wins
- **Tests:** Fold-to-win at every street, pot verification, stack accounting

### 5. All-In Scenarios
**File:** `allin-scenarios.spec.ts`
- 3-player all-in facing raise with fold and call
- 4-player all-in with some callers, some folders
- 2-player mutual all-in showdown (heads-up)
- **Tests:** All-in mechanics, showdown after all-in, pot distribution

### 6. Elimination and Continuation
**File:** `elimination-continuation.spec.ts`
- 3 players, one gets eliminated via all-in, remaining 2 continue playing
- 3 players play 3 consecutive hands with chip tracking
- 3 hands played after a player is eliminated
- Second player elimination (3 -> 2 -> 1 player) - game ends correctly
- Button rotation verified to skip eliminated player's seat
- **Tests:** Player elimination, continuation with fewer players, chip conservation, button rotation

### 7. Check-Raise Scenarios
**File:** `check-raise.spec.ts`
- 3-player check-raise on flop - action reopens after raise
- 2-player check-raise on turn
- 2-player check-raise on river with fold (bluff scenario)
- **Tests:** Check-raise action flow, action reopening after raise, fold to check-raise

### 8. Big Blind Option Tests
**File:** `bb-option.spec.ts`
- BB can check after all players limp (BB option)
- BB can raise after all players limp (BB option raise)
- BB has no option after someone raises (must call/fold/re-raise)
- **Tests:** BB option flag, action availability (check/raise) after limps, no check after raise

### 9. Min Raise Validation
**File:** `min-raise-validation.spec.ts`
- Min raise equals 2x BB preflop (first raise)
- Min re-raise equals previous raise + last raise increment
- Raise button disabled when amount below minimum
- Min raise resets to BB on new street
- All-in below min raise is allowed (special exception)
- **Tests:** Min raise calculation, preflop/postflop validation, all-in exception

### 10. Heads-Up Blind Structure
**File:** `heads-up-blinds.spec.ts`
- Dealer posts SB in heads-up (dealer = SB, other = BB)
- Dealer/SB acts first preflop in heads-up
- BB acts first on flop (postflop action order)
- Correct blinds when going from 3 to 2 players
- **Tests:** Heads-up blind posting, action order, transition from 3+ players

### 11. Player Disconnect/Rejoin
**File:** `player-rejoin.spec.ts`
- Session token preserves player identity (no re-login needed)
- Disconnecting loses table seat (player returns to lobby)
- Player can manually rejoin table after disconnect
- Remaining player continues after opponent disconnects
- **Tests:** Session persistence, seat loss on disconnect, manual rejoin flow

---

## To Implement

### 12. All-In with Side Pots (BLOCKED)
**File:** `three-player-allin-sidepot.spec.ts`
- Player 1 has 500 chips, Player 2 has 1000 chips, Player 3 has 200 chips
- Player 3 goes all-in preflop (200)
- Player 1 calls, Player 2 raises to 500
- Player 1 goes all-in (500 total)
- Player 2 calls
- Hand plays out to showdown with main pot + side pot
- **Tests:** All-in mechanics, side pot creation/distribution, short stack handling
- **Status:** Blocked until side pot feature is implemented

### 12. Player Disconnect/Rejoin Mid-Hand
**File:** `three-player-rejoin.spec.ts`
- 3 players in hand
- Player 2 closes browser mid-hand
- Player 2 reopens and rejoins with same name
- Verify they see correct game state and can continue playing
- **Tests:** Identity persistence, mid-game state sync, reconnection handling

### 13. Split Pot / Tie at Showdown
**File:** `split-pot.spec.ts`
- 2 players with identical hands (e.g., both have same straight from board)
- Pot splits evenly between winners
- Odd chip handling (101 chips split = 51/50)
- 3-way tie split verification
- **Tests:** Tie detection, even split, odd chip remainder handling
- **Backend:** `table.go:169`, `hand_evaluator.go`

### 14. Short Stack Posts Partial Blind
**File:** `short-stack-blind.spec.ts`
- Player has less chips than BB (e.g., 15 chips, BB is 20)
- Player posts all 15 as their blind (all-in)
- Player is still dealt in and can win main pot
- **Tests:** Partial blind posting, all-in on blind, pot eligibility

### 15. Multiple Streets with Raises
**File:** `multi-street-raises.spec.ts`
- Betting action with raises on flop, turn, AND river
- Pot accumulates correctly across all streets
- Street bets reset each street but pot grows
- **Tests:** Multi-street pot accumulation, bet reset per street
