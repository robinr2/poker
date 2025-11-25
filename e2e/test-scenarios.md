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

---

## To Implement

### 2. Player Fold to Win (No Showdown)
**File:** `three-player-fold-to-win.spec.ts`
- 3 players in hand
- Preflop: One player raises big
- Other two players fold
- Winner takes pot without showdown
- **Tests:** Fold-to-win logic, no showdown UI, pot awarded correctly

### 3. Multiple Consecutive Hands
**File:** `three-player-multiple-hands.spec.ts`
- 3 players play 3 consecutive hands
- Verify dealer button rotates correctly
- Verify blinds are posted from correct positions each hand
- Verify chip stacks update correctly between hands
- **Tests:** Button rotation, blind posting, state reset between hands

### 4. Check-Raise Scenario
**File:** `three-player-check-raise.spec.ts`
- 3 players see flop
- Player 1 checks, Player 2 bets, Player 3 folds
- Player 1 check-raises
- Player 2 calls
- Play to showdown
- **Tests:** Check-raise action flow, action reopening after raise

### 5. All-In and Side Pot Scenario
**File:** `three-player-allin-sidepot.spec.ts`
- Player 1 has 500 chips, Player 2 has 1000 chips, Player 3 has 200 chips
- Player 3 goes all-in preflop (200)
- Player 1 calls, Player 2 raises to 500
- Player 1 goes all-in (500 total)
- Player 2 calls
- Hand plays out to showdown with main pot + side pot
- **Tests:** All-in mechanics, side pot creation/distribution, short stack handling

### 6. Player Disconnect/Rejoin Mid-Hand
**File:** `three-player-rejoin.spec.ts`
- 3 players in hand
- Player 2 closes browser mid-hand
- Player 2 reopens and rejoins with same name
- Verify they see correct game state and can continue playing
- **Tests:** Identity persistence, mid-game state sync, reconnection handling

### 7. Heads-Up After Elimination
**File:** `three-player-to-headsup.spec.ts`
- Start with 3 players
- One player loses all chips and is eliminated
- Verify game continues heads-up with 2 players
- Play another hand heads-up
- **Tests:** Player elimination, seat management, heads-up blind structure
