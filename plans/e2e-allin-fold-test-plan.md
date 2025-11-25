# E2E Test Plan: All-In and Fold Scenarios

## Overview

This plan outlines e2e tests to ensure all-in and fold mechanics work correctly in multi-player lobbies. These are critical gameplay scenarios that must work flawlessly.

## Current Coverage

| Scenario | File | Status |
|----------|------|--------|
| 2-player all-in to showdown | `deterministic-winner.spec.ts` | ✅ Done |
| 3-player all-in to showdown | `deterministic-winner.spec.ts` | ✅ Done |
| 3-player fold-to-win preflop | `three-player-fold-to-win.spec.ts` | ✅ Done |
| Full hand with 1 fold on flop | `three-player-complete-hand.spec.ts` | ✅ Done |

## Priority Tests to Add

### Priority 1: Critical All-In Scenarios

#### Test 1: Partial All-In (Short Stack)
**File:** `e2e/allin-short-stack.spec.ts`

**Scenario:**
- 4 players with different stack sizes: P1=1000, P2=1000, P3=300, P4=1000
- P3 (short stack) goes all-in for 300 preflop
- P1 calls 300, P2 folds, P4 calls 300
- Hand goes to showdown
- **Verify:** P3 can only win 900 (300 x 3 players)

**Why important:** Tests that all-in players with less than the bet amount are handled correctly.

---

#### Test 2: Multiple All-Ins with Different Stack Sizes (Side Pots)
**File:** `e2e/allin-side-pots.spec.ts`

**Scenario:**
- 4 players: P1=1000, P2=500, P3=200, P4=1000
- Preflop: P3 all-in 200, P2 all-in 500, P1 calls 500, P4 folds
- **Expected pots:**
  - Main pot: 600 (200 x 3 players)
  - Side pot: 600 (300 x 2 players - P1 and P2)
- P3 wins main pot, P1 wins side pot (use deterministic deck)
- **Verify:** Correct pot splits and chip distributions

**Why important:** Side pots are essential for fair multi-way all-in scenarios.

**Note:** This test depends on side pot implementation (see `future-todos.md`).

---

#### Test 3: All-In When Facing a Raise
**File:** `e2e/allin-facing-raise.spec.ts`

**Scenario:**
- 3 players: all start with 1000
- P1 raises to 100, P2 re-raises to 300
- P3 goes all-in for 1000
- P1 folds, P2 calls
- Showdown between P2 and P3
- **Verify:** Pot is correct (2000 from P2+P3 + 100 from P1's dead money = 2100)

**Why important:** Tests all-in as a response to aggression, with dead money accounting.

---

### Priority 2: Fold Scenarios

#### Test 4: Sequential Folds (Last Player Standing)
**File:** `e2e/fold-last-standing.spec.ts`

**Scenario:**
- 4 players
- Preflop: P1 raises, P2 folds, P3 folds, P4 folds
- **Verify:** P1 wins blinds without showdown, hand completes

**Why important:** Tests fold-to-win with more than 3 players.

---

#### Test 5: Fold on Each Street
**File:** `e2e/fold-each-street.spec.ts`

**Scenarios (3 sub-tests):**
1. **Fold on Flop:** 3 players see flop, P1 bets, P2 and P3 fold
2. **Fold on Turn:** 2 players see turn, P1 bets, P2 folds
3. **Fold on River:** 2 players see river, P1 bets, P2 folds

**Verify for each:** Winner gets pot, no showdown overlay, hand completes cleanly.

**Why important:** Tests fold-to-win at each street, not just preflop.

---

#### Test 6: All Fold to Big Blind
**File:** `e2e/fold-to-big-blind.spec.ts`

**Scenario:**
- 3 players
- Preflop: UTG folds, SB folds
- **Verify:** BB wins without posting additional chips, wins SB's blind

**Why important:** Edge case where BB wins uncontested.

---

### Priority 3: All-In + Fold Combinations

#### Test 7: All-In with Some Callers, Some Folders
**File:** `e2e/allin-mixed-response.spec.ts`

**Scenario:**
- 4 players with 1000 each
- Preflop: P1 goes all-in (1000)
- P2 calls, P3 folds, P4 calls
- Showdown between P1, P2, P4
- **Verify:** 
  - P3's folded chips stay in pot (blinds only)
  - Winner takes full pot (3000 + blinds)

**Why important:** Tests realistic multi-player all-in scenario.

---

#### Test 8: Multiple Hands with All-Ins (Chip Continuity)
**File:** `e2e/allin-chip-continuity.spec.ts`

**Scenario:**
- 3 players start with 1000 each
- Hand 1: P1 wins all-in, ends with ~2000 chips
- Hand 2: P2 goes all-in with remaining ~500, loses
- Hand 3: Verify P2 is either eliminated or has 0 chips
- **Verify:** Chip stacks persist correctly across multiple hands

**Why important:** Tests that all-in results persist and affect subsequent hands.

---

### Priority 4: Edge Cases

#### Test 9: All-In for Exact Amount
**File:** `e2e/allin-exact-call.spec.ts`

**Scenario:**
- 3 players
- P1 has 500, P2 has 500, P3 has 1000
- P1 raises to 500 (all-in), P2 calls 500 (all-in), P3 folds
- **Verify:** Both players are all-in, hand goes to showdown

**Why important:** Tests all-in when players have equal stacks.

---

#### Test 10: Big Blind All-In (Short Stack BB)
**File:** `e2e/allin-short-bb.spec.ts`

**Scenario:**
- 3 players: P1=1000, P2=1000, P3 (BB)=15 chips
- Blinds are 10/20
- P3 can only post 15 as BB (partial blind)
- P1 calls, P2 folds
- P3 is automatically all-in
- **Verify:** Hand plays correctly with partial blind and automatic all-in

**Why important:** Tests edge case of player with less than BB.

---

## Implementation Order

**Phase 1: Core Functionality (Must Have)**
1. Test 4: Sequential Folds (Last Player Standing)
2. Test 5: Fold on Each Street
3. Test 3: All-In When Facing a Raise

**Phase 2: Multi-Player Scenarios**
4. Test 7: All-In with Some Callers, Some Folders
5. Test 6: All Fold to Big Blind
6. Test 8: Multiple Hands with All-Ins

**Phase 3: Side Pots (After Implementation)**
7. Test 1: Partial All-In (Short Stack)
8. Test 2: Multiple All-Ins with Different Stack Sizes

**Phase 4: Edge Cases**
9. Test 9: All-In for Exact Amount
10. Test 10: Big Blind All-In (Short Stack BB)

---

## Test File Structure

```
e2e/
  helpers/
    browser-helpers.ts
    deterministic-helpers.ts
  # Existing tests
  deterministic-winner.spec.ts        # 3 tests
  three-player-complete-hand.spec.ts  # 1 test
  three-player-fold-to-win.spec.ts    # 1 test
  # New tests (Phase 1)
  fold-scenarios.spec.ts              # Tests 4, 5, 6
  allin-facing-raise.spec.ts          # Test 3
  # New tests (Phase 2)
  allin-mixed-response.spec.ts        # Test 7
  allin-chip-continuity.spec.ts       # Test 8
  # New tests (Phase 3 - requires side pot implementation)
  allin-side-pots.spec.ts             # Tests 1, 2
  # New tests (Phase 4)
  allin-edge-cases.spec.ts            # Tests 9, 10
```

---

## Prerequisites

1. **For all tests:** Server running with `POKER_TEST_MODE=true`
2. **For deterministic tests:** Use `setDeck()` helper to control outcomes
3. **For side pot tests:** Side pot implementation must be complete (see `future-todos.md`)

---

## Acceptance Criteria

Each test must verify:
1. **Correct pot amounts** at each stage
2. **Correct winner(s)** receive correct chip amounts
3. **UI state** reflects game state (showdown overlay, stack changes)
4. **Hand completion** - new hand can be started after
5. **No console errors** in browser

---

## Notes

- Side pot tests (Phase 3) are blocked until side pot feature is implemented
- All tests should use the `browser-helpers.ts` for headless/headed mode support
- Use `resetTable()` in `beforeEach` for clean state between tests
- Target: ~30-60 seconds per test in headless mode
