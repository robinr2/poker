# Manual Poker Game Exploration Notes

**Date:** 2024-11-26
**Purpose:** Document game actions and identify potential issues for e2e test creation

## Browser Setup
- Player1 (Alice): Instance `467adc21-6af6-4f6c-a563-562403bf0945`
- Player2 (Bob): Instance `034585e5-bc15-4f15-97b5-0d1f8d9e7cb0`
- Player3 (Charlie): Instance `a7465820-f168-4282-9104-13f5fab554b0`

## Legend
- **Street contributions**: Amount bet on current street (in front of player)
- **Stack**: Total chips remaining
- **Pot**: Total chips in the middle

---

## Issues Found

**Note:** "NS" prefix = New Session (second game). Numbers without prefix refer to first session.

| # | Description | Severity | Hand # (Session) | Notes |
|---|-------------|----------|------------------|-------|
| 1 | Showdown shows "Unknown Hand" when winning by fold | Low | 1, NS-1, NS-2, NS-3, NS-5 | Should just say "All folded" or not show hand ranking. Confirmed across both sessions. |
| 2 | Bet amounts shown are cumulative for hand, not per-street | Low | 1 | Unusual UX - typically shows current street bets |
| 3 | **HEADS-UP POSTFLOP ACTION ORDER WRONG** | **HIGH** | 3, NS-4, NS-5 | **CRITICAL:** In true heads-up (2 players only), BB acts first on ALL postflop streets (flop, turn, river). Button/SB should act first. Confirmed across flop/turn/river in multiple hands. |
| 4 | All-in button doesn't work | Medium | NS-6 | Clicking "All-in" button doesn't trigger action. Need to use Min/Pot + Raise workaround. |
| 5 | Showdown hand description lacks detail | Low | NS-6 | Shows "One Pair" but doesn't clarify kickers or full 5-card hand. Could be more descriptive. |
| 6 | **ACTION BUTTONS FAIL AFTER FIRST ACTION** | **CRITICAL** | All hands (3rd session) | **CRITICAL:** After first player action in a hand, subsequent action buttons (fold/call/check/bet/raise) stop working. No WebSocket messages reach server. Actions fail silently with no error. Makes game unplayable beyond first action. |
| 7 | Disconnect creates extra chips | Medium | 3rd session Hand 1 | When Bob disconnected mid-hand (lost 20 to pot), then rejoined, he got fresh 1000 stack instead of 980. Charlie kept his 1020 from winning. Total chips became 3020 instead of 3000. |

---

## Game Log

### Setup Phase

- Alice, Bob, Charlie all joined Table 1
- Initial stacks: Alice=1000, Bob=1000, Charlie=1000
- Seating: Seat 0=Alice, Seat 1=Bob, Seat 2=Charlie

---

### Hand 1

**Preflop - Initial Deal**
| Player | Position | Cards | Stack | Street Bet | 
|--------|----------|-------|-------|------------|
| Alice | D (Button) | 5♥ 3♠ | 1000 | 0 |
| Bob | SB | K♥ 3♥ | 990 | 10 |
| Charlie | BB | 7♦ 9♣ | 980 | 20 |

Pot: 30 | Street: Preflop | Action on: Alice
Actions available: Fold, Call 20, Min, Pot, All-in, Raise

**Action 1: Alice calls 20**
State after: Alice $20/stack 980, Bob $10/stack 990, Charlie $20/stack 980, Pot 50

**Action 2: Bob calls 10**
State after: All $20, all stacks 980, Pot 60

**Action 3: Charlie checks (BB option)**
Proceeds to Flop: T♥ K♣ K♠

**Flop**
Pot: 60 | Action on: Bob (SB - first to act postflop)
NOTE: Bet amounts still showing $20 each from preflop (cumulative display)

**Action 4: Bob bets 20 (min)**
State after: Bob $40/stack 960, others $20/stack 980, Pot 80

**Action 5: Charlie folds**

**Action 6: Alice folds**

**Result:** Bob wins 80 chips (all opponents folded)
**ISSUE:** Showdown overlay shows "Unknown Hand" - shouldn't show hand ranking for fold win

Final stacks: Alice=980, Bob=1040, Charlie=980

---

### Hand 2

**Preflop - Initial Deal**
| Player | Position | Cards | Stack | Street Bet | 
|--------|----------|-------|-------|------------|
| Alice | BB | 5♥ 2♦ | 960 | 20 |
| Bob | D (Button) | 2♠ 6♠ | 1040 | 0 |
| Charlie | SB | 3♣ K♠ | 970 | 10 |

Pot: 30 | Street: Preflop | Action on: Bob

**Action 1: Bob goes ALL-IN for 1040**
State after: Alice $20/stack 960, Bob $1040/stack 0, Charlie $10/stack 970, Pot 1070

**Action 2: Charlie folds**

**Action 3: Alice calls 960 (all-in)**
Alice puts in remaining 960 chips (total 980 in pot from Alice)

**Showdown**
Board: 8♦ 4♠ 3♦ K♥ K♦
- Bob: 2♠ 6♠ - One Pair (Kings on board)
- Alice: 5♥ 2♦ - One Pair (Kings on board) - but lower kicker

**Result:** Bob wins 2030 (total pot)
**Note:** Alice had 980 total, Bob had 1040. Side pot situation:
- Main pot: 980 * 2 + 10 (Charlie's SB) = 1970
- Bob's excess 60 should be returned or be in a side pot
- Actual pot shown: 2030 (appears correct as 980+1040+10=2030)

Final stacks: Alice=0 (ELIMINATED), Bob=2030, Charlie=970
Total chips: 3000 ✓

---

### Hand 3

**Heads-Up: Alice eliminated, only Bob and Charlie remain**

**Preflop - Initial Deal**
| Player | Position | Cards | Stack | Street Bet | 
|--------|----------|-------|-------|------------|
| Bob | BB | Q♥ K♥ | 2010 | 20 |
| Charlie | D+SB | 6♥ A♥ | 960 | 10 |

Pot: 30 | Street: Preflop | Action on: Charlie (correct - button acts first preflop in heads-up)

**Action 1: Charlie calls 10**
State after: Both $20, Bob stack 2010, Charlie stack 950, Pot 40

**Action 2: Bob raises to 40 (min raise)**
State after: Bob $40/stack 1990, Charlie $20/stack 950, Pot 60

**Action 3: Charlie calls 20**
Proceeds to Flop: Q♠ 9♥ 7♠

**Flop**
Pot: 80 | Board: Q♠ 9♥ 7♠

**BUG FOUND:** Bob (BB) has action first postflop, but in heads-up Charlie (Button) should act first!
This is a significant rules violation for heads-up play.

---

# NEW GAME SESSION STARTED (2024-11-26 - Continued Testing)

**Note:** Closed previous game and started fresh to continue testing with visible browser windows (900x1200).

### Setup Phase
- Alice, Bob, Charlie all joined Table 1
- Initial stacks: Alice=1000, Bob=1000, Charlie=1000
- Seating: Seat 0=Alice, Seat 1=Bob, Seat 2=Charlie

---

### Hand 1 (New Session)

**Preflop - Initial Deal**
| Player | Position | Cards | Stack | Street Bet | 
|--------|----------|-------|-------|------------|
| Alice | BB | J♣ 9♥ | 980 | 20 |
| Bob | D (Button) | 7♦ 5♣ | 1000 | 0 |
| Charlie | SB | A♥ Q♠ | 990 | 10 |

Pot: 30 | Street: Preflop | Action on: Bob

**Action 1: Bob calls 20**
State after: Alice $20/stack 980, Bob $20/stack 980, Charlie $10/stack 990, Pot 50

**Action 2: Charlie raises to 40 (min raise)**
State after: Alice $20/stack 980, Bob $20/stack 980, Charlie $40/stack 960, Pot 80

**Action 3: Alice folds**

**Action 4: Bob folds**

**Result:** Charlie wins 80 chips (all opponents folded)
**ISSUE CONFIRMED:** Showdown overlay shows "Unknown Hand" - same issue as before

Final stacks: Alice=980, Bob=980, Charlie=1040

---

### Hand 2 (New Session)

**Preflop - Initial Deal**
| Player | Position | Cards | Stack | Street Bet | 
|--------|----------|-------|-------|------------|
| Alice | SB | Q♦ 8♣ | 970 | 10 |
| Bob | BB | A♠ 6♥ | 960 | 20 |
| Charlie | D (Button) | 7♥ 4♣ | 1040 | 0 |

Pot: 30 | Street: Preflop | Action on: Charlie

**Action 1: Charlie raises to 40 (min raise)**
State after: Alice $10/stack 970, Bob $20/stack 960, Charlie $40/stack 1000, Pot 70

**Action 2: Alice calls 30**
State after: Alice $40/stack 940, Bob $20/stack 960, Charlie $40/stack 1000, Pot 100

**Action 3: Bob raises to 60 (min re-raise)**
State after: Alice $40/stack 940, Bob $60/stack 920, Charlie $40/stack 1000, Pot 140

**Action 4: Charlie folds**

**Action 5: Alice calls 20**
Proceeds to Flop: 9♦ J♣ 7♣
NOTE: Charlie folded preflop, now heads-up Alice vs Bob postflop

**Flop**
Pot: 160 | Board: 9♦ J♣ 7♣
Action on: Alice (SB - correct, SB acts first postflop in multi-way)

**Action 6: Alice checks**

**Action 7: Bob bets 20 (min)**
State after: Alice $60/stack 920, Bob $80/stack 900, Charlie $40 (folded)/stack 1000, Pot 180

**Action 8: Alice raises to 260 (pot-sized check-raise)**
State after: Alice $260/stack 720, Bob $80/stack 900, Pot 380

**Action 9: Bob folds**

**Result:** Alice wins 380 chips (Bob folded to check-raise)
**ISSUE CONFIRMED AGAIN:** Showdown overlay shows "Unknown Hand" - same issue

Final stacks: Alice=1100, Bob=900, Charlie=1000

**NOTE:** Check-raise functionality working correctly!

---

### Hand 3 (New Session)

**Preflop - Initial Deal**
| Player | Position | Cards | Stack | Street Bet | 
|--------|----------|-------|-------|------------|
| Alice | D (Button) | A♥ 5♠ | 1100 | 0 |
| Bob | SB | 9♠ 6♣ | 890 | 10 |
| Charlie | BB | Q♣ 2♥ | 980 | 20 |

Pot: 30 | Street: Preflop | Action on: Alice

**Action 1: Alice folds**
NOTE: Testing heads-up action order with Bob vs Charlie

**Action 2: Bob calls 10**
State after: Alice (folded), Bob $20/stack 880, Charlie $20/stack 980, Pot 40

**Action 3: Charlie raises to 40 (min raise, BB option)**
State after: Bob $20/stack 880, Charlie $40/stack 960, Pot 60

**Action 4: Bob calls 20**
Proceeds to Flop: 5♦ 9♣ 9♦

**Flop**
Pot: 80 | Board: 5♦ 9♣ 9♦
Action on: Bob (SB - acts first postflop)
Bob has: 9♠ 6♣ (TRIP NINES!)
Charlie has: Q♣ 2♥ (nothing)

**NOTE:** In this 3-way-becomes-2-way scenario, SB acts first postflop (standard multi-way rule). This is CORRECT.
**IMPORTANT:** Need to test TRUE heads-up (only 2 players from start) to verify Button/SB acts first in that scenario.

**Action 5: Bob bets 80 (pot-sized)**
State after: Bob $120/stack 780, Charlie $40/stack 960, Pot 160

**Action 6: Charlie folds**

**Result:** Bob wins 160 chips (Charlie folded)
**ISSUE CONFIRMED AGAIN:** Showdown overlay shows "Unknown Hand" - same issue

Final stacks: Alice=1100, Bob=940, Charlie=960

---

### Hand 4 (New Session - TRUE HEADS-UP TEST)

**Setup:** Charlie left the table. Now testing TRUE heads-up (only 2 players from start of hand).

**Preflop - Initial Deal**
| Player | Position | Cards | Stack | Street Bet | 
|--------|----------|-------|-------|------------|
| Alice | BB | T♣ 2♠ | 1080 | 20 |
| Bob | D + SB | 7♦ 5♣ | 930 | 10 |

Pot: 30 | Street: Preflop | Action on: Bob (correct - Button acts first preflop in heads-up)

**Action 1: Bob calls 10**
State after: Alice $20/stack 1080, Bob $20/stack 920, Pot 40

**Action 2: Alice checks (BB option)**
Proceeds to Flop: 9♥ K♣ 3♥

**Flop**
Pot: 40 | Board: 9♥ K♣ 3♥
Action on: Alice (BB) - **BUG! Button/SB should act first in heads-up postflop!**

**BUG CONFIRMED:** Alice (BB) has action first on flop. In TRUE heads-up, Button/SB should act first on ALL postflop streets.

**Action 3: Alice checks**

**Action 4: Bob checks**
Proceeds to Turn: 6♦

**Turn**
Pot: 40 | Board: 9♥ K♣ 3♥ 6♦
Action on: Alice (BB) - **BUG PERSISTS on turn!**

**Action 5: Alice checks**

**Action 6: Bob checks**
Proceeds to River: J♥

**River**
Pot: 40 | Board: 9♥ K♣ 3♥ 6♦ J♥
Action on: Alice (BB) - **BUG PERSISTS on river!**

**Action 7: Alice checks**

**Action 8: Bob checks**

**Showdown**
- Alice: T♣ 2♠ - High Card (Ten)
- Bob: 7♦ 5♣ - High Card (Seven)

**Result:** Alice wins 40 chips
**GOOD:** Showdown correctly shows "High Card" (not "Unknown Hand") because both players showed down

Final stacks: Alice=1120, Bob=920

**CRITICAL BUG CONFIRMED:** In true heads-up play, the BB acts first on ALL postflop streets (flop, turn, river). 
**EXPECTED BEHAVIOR:** Button/SB should act first on ALL postflop streets in heads-up.
**SEVERITY:** HIGH - This is a fundamental poker rules violation for heads-up play.

---

### Hand 5 (New Session - Heads-Up Continued)

**Preflop - Initial Deal**
| Player | Position | Cards | Stack | Street Bet | 
|--------|----------|-------|-------|------------|
| Alice | D + SB | A♥ 6♣ | 1110 | 10 |
| Bob | BB | 7♦ 5♣ | 920 | 20 |

Pot: 30 | Street: Preflop | Action on: Alice

**Action 1: Alice raises to 40 (min raise)**
State after: Alice $40/stack 1080, Bob $20/stack 920, Pot 60

**Action 2: Bob calls 20**
Proceeds to Flop: T♠ 8♥ 7♦

**Flop**
Pot: 80 | Board: T♠ 8♥ 7♦
Action on: Bob (BB) - **BUG: Button/SB should act first in heads-up**

**Action 3: Bob checks**

**Action 4: Alice checks**
Proceeds to Turn: 5♠

**Turn**
Pot: 80 | Board: T♠ 8♥ 7♦ 5♠
Action on: Bob (BB) - **BUG PERSISTS**

**Action 5: Bob bets 20 (min)**
State after: Alice $40/stack 1080, Bob $60/stack 860, Pot 100

**Action 6: Alice folds**

**Result:** Bob wins 100 chips
**ISSUE CONFIRMED:** "Unknown Hand" showdown display bug

Final stacks: Alice=1080, Bob=960

---

### Hand 6 (New Session - 3-Player Return)

**Setup:** Charlie rejoined the table

**Preflop - Initial Deal**
| Player | Position | Cards | Stack | Street Bet | 
|--------|----------|-------|-------|------------|
| Alice | BB | 5♣ 2♥ | 1060 | 20 |
| Bob | D (Button) | K♣ 7♣ | 960 | 0 |
| Charlie | SB | A♣ 3♣ | 990 | 10 |

Pot: 30 | Street: Preflop | Action on: Bob

**Action 1: Bob raises to 50 (pot-sized)**
State after: Alice $20/stack 1060, Bob $50/stack 910, Charlie $10/stack 990, Pot 80

**Action 2: Charlie calls 40**
State after: Alice $20/stack 1060, Bob $50/stack 910, Charlie $50/stack 950, Pot 120

**Action 3: Alice folds**
Proceeds to Flop: 4♣ 9♦ 4♠

**Flop** (Heads-up: Bob vs Charlie)
Pot: 120 | Board: 4♣ 9♦ 4♠
Action on: Charlie (SB - correct for 3-way-becomes-2-way)

**Action 4: Charlie bets 120 (pot-sized)**
State after: Bob $50/stack 910, Charlie $170/stack 830, Pot 240

**Action 5: Bob calls 120**
Proceeds to Turn: 5♠

**Turn**
Pot: 360 | Board: 4♣ 9♦ 4♠ 5♠
Action on: Charlie (SB)

**Action 6: Charlie checks**

**Action 7: Bob checks**
Proceeds to River: 6♣

**River**
Pot: 360 | Board: 4♣ 9♦ 4♠ 5♠ 6♣
Action on: Charlie (SB)
Charlie has: A♣ 3♣ (Ace-high, four clubs total including board)
Bob has: K♣ 7♣ (King-high, four clubs total including board)
Board has 2 clubs (4♣, 6♣) - neither player has flush (need 5 clubs)

**Action 8: Charlie bets 360 (pot-sized)**
State after: Bob $170/stack 790, Charlie $530/stack 470, Pot 720

**Action 9: Bob calls 360**

**Showdown**
Board: 4♣ 9♦ 4♠ 5♠ 6♣
- Charlie: A♣ 3♣ - One Pair (4s) with A-high kicker
- Bob: K♣ 7♣ - One Pair (4s) with K-high kicker

**Result:** Charlie wins 1080 chips (correct - Ace kicker beats King kicker)
**NEW ISSUE:** Showdown displays "One Pair" but doesn't clarify the kickers or full hand properly

Final stacks: Alice=1060, Bob=430, Charlie=1550

**NOTE:** All-in button appears to not work - clicking it doesn't trigger an all-in action. Need to use Min/Pot + Raise to bet.

---



---

## Third Session (Table 2) - CRITICAL BUG DISCOVERED

**Date:** 2025-11-26 16:40-16:45 CET
**Browser Setup:**
- Alice: Instance `06e6a062-7d6f-49b9-abaa-e5c14d47d78d`
- Bob: Instance `4f3304cc-a5be-405b-8015-662356e35098`
- Charlie: Instance `cf089114-af6e-44ba-bde1-d20f5e9e064f`

### Setup
All three players joined Table 2 with fresh 1000 chip stacks.

---

### Hand 1 (INCOMPLETE - CRITICAL BUG)

**Initial State:**
- Alice (BTN/D): 1000 chips, A♣4♣
- Bob (SB): 1000 chips (posted 10), T♠5♦
- Charlie (BB): 1000 chips (posted 20), 6♣7♥
- Pot: 30

**Preflop Actions:**
1. Alice folds (button clicked successfully)
2. Bob calls 10 to complete SB (button clicked successfully)  
3. Charlie checks BB (button clicked successfully)

**Flop:** Q♥ 2♦ 8♦
- Pot: 40
- Bob (SB) first to act

**CRITICAL BUG DISCOVERED:**
- Bob attempted to click "Pot" button - **NO RESPONSE**
- Bob attempted to click "Min" button - **NO RESPONSE**
- Server logs show hand_started but NO subsequent action messages
- WebSocket messages for player actions are NOT reaching the server
- **This makes the game completely unplayable after the first few actions**

**Additional Discovery - Disconnect Bug:**
- Bob's browser was refreshed (simulating disconnect)
- Hand ended immediately, Charlie won with "Unknown Hand" (Bug #1)
- Charlie won 40 chip pot, stack became 1020
- Bob rejoined table with FRESH 1000 stack (should have been 980)
- **Result:** Total chips = 3020 instead of 3000 (20 chips created)

---

### Hand 2 (INCOMPLETE - SAME CRITICAL BUG)

**Initial State:**
- Alice (SB): 1000 chips (posted 10), 7♥K♦
- Bob (BB): 1000 chips (posted 20), 3♠J♦
- Charlie (BTN/D): 1020 chips, T♠K♥
- Pot: 30

**Attempted Actions:**
- Charlie (button) tried to click "Pot" - **NO RESPONSE**
- Same critical bug - action buttons fail silently
- Server receives hand_started but no player action messages

---

## CRITICAL BLOCKER - BUG #6

**Summary:** After a hand starts and players take their first few actions, all subsequent action buttons stop working. The clicks register on the frontend (buttons can be clicked) but NO WebSocket messages reach the server. Server logs show only `hand_started` messages but no `player_action` or similar messages.

**Impact:** GAME BREAKING - Makes poker unplayable beyond 2-3 actions

**Reproducibility:** 100% - Occurs in every hand in this session after initial actions

**Hypothesis:** 
1. WebSocket connection may be failing/disconnecting after hand starts
2. Frontend may not be properly sending action messages
3. Server may not be properly handling action messages after certain state
4. Possible race condition or state desync between client/server

**Required Fix:** This must be fixed before any further gameplay testing can continue. All-in scenarios, side pot testing, and other gameplay features cannot be tested until basic actions work reliably.

---

## Testing Status

**Completed:**
- ✅ Basic gameplay flow (first session)
- ✅ Fold scenarios
- ✅ Betting and raising
- ✅ Heads-up postflop play
- ✅ Showdown mechanics
- ✅ Winner determination

**Blocked by Bug #6:**
- ❌ All-in scenarios
- ❌ Side pot calculations
- ❌ Multi-way all-in with different stack sizes
- ❌ Chip accounting verification
- ❌ Any extended gameplay testing

**Bugs Found:** 7 total (1 CRITICAL, 1 HIGH, 2 Medium, 3 Low)

