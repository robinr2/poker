## Plan Complete: Pot Accounting Fix + Frontend Display Fixes

Successfully completed the pot accounting fix across backend and frontend, plus resolved three critical display bugs.

**Phases Completed:** 5 of 5
1. ✅ Phase 1: Backend pot sweep tests
2. ✅ Phase 2: Backend pot sweep implementation
3. ✅ Phase 3: Comprehensive backend test updates
4. ✅ Phase 4: Frontend all-in button fix
5. ✅ Phase 5: Frontend community cards display fix + card back placeholders

**All Files Created/Modified:**
- internal/server/table.go
- internal/server/table_test.go
- frontend/src/components/TableView.tsx
- frontend/src/components/TableView.test.tsx

**Key Changes:**

**Backend - Pot Accounting:**
- Modified `AdvanceStreet()` to sweep `PlayerBets` into `Pot` at street transitions
- Modified `HandleShowdown()` to sweep `PlayerBets` into `Pot` before winner calculation (both early winner and showdown paths)
- Fixed early winner bug where all-fold scenarios returned 0 chips (now correctly sweeps PlayerBets)
- Fixed `GetMaxRaise()` to return total commitment (bet + stack) instead of just stack
- Updated `StartHand()` to initialize PlayerBets with blinds, Pot stays at 0 until first street advance
- Updated 20+ tests to expect new pot accounting (Pot=0 during betting, sweeps at transitions)

**Frontend - All-In Button:**
- Fixed all-in button to use `gameState?.maxRaise` instead of calculating from stack alone
- This ensures correct all-in amounts that include both current bet and remaining stack
- Removed unused `playerStack` variable that was causing TypeScript warnings

**Frontend - Start Hand Button Logic:**
- Fixed start hand button visibility condition from `pot === 0` to proper hand-in-progress detection
- Now checks for `holeCards` or `dealerSeat` to determine if hand is active
- This fixes the issue where pot=0 during preflop made button appear incorrectly

**Frontend - Community Cards Display:**
- Fixed game info and board cards visibility condition from `!showStartHandButton` to `handInProgress`
- This allows board cards to remain visible after showdown/winner determination until new hand starts
- Players can now review the final board state and verify hand rankings after the pot is awarded
- Added logic to show 5 card back placeholders (🂠) during preflop when no board cards exist yet
- Placeholders transition to actual cards as flop/turn/river are dealt
- Empty slots shown for remaining cards after partial board is dealt
- Board cards and game info visibility now tied to hand state (holeCards or dealerSeat present)
- Game info (street indicator and pot) also persists until new hand starts

**Frontend - Test Updates:**
- Updated 3 board card display tests to match new behavior
- Tests now properly distinguish between "no hand active" vs "preflop with card backs"
- All tests now pass with new pot accounting and display logic

**Test Coverage:**
- Backend: 311 tests passing, 0 failures ✅
- Frontend: 247 tests passing, 0 failures ✅
- Build: Successful ✅

**Key Functions/Classes Modified:**

Backend:
- `Table.AdvanceStreet()` - Pot sweep logic added
- `Table.HandleShowdown()` - Pot sweep logic added (both paths)
- `Table.GetMaxRaise()` - Returns bet + stack instead of just stack
- `Table.StartHand()` - Initializes PlayerBets with blinds, Pot=0
- `Hand.ProcessAction()` - Early winner fix

Frontend:
- `TableView` component - Start hand button logic, all-in button fix, board visibility conditions using handInProgress, board placeholder logic
- Board card and game info display now based on hand state rather than button visibility

**Visual Improvements:**

Before: Community cards and game info hidden during preflop (pot=0), no visual feedback
After: 
- Preflop: 5 card back placeholders (🂠 🂠 🂠 🂠 🂠)
- Flop: 3 face-up cards + 2 empty slots (A♠ K♥ Q♦ _ _)
- Turn: 4 face-up cards + 1 empty slot (A♠ K♥ Q♦ J♣ _)
- River: 5 face-up cards (A♠ K♥ Q♦ J♣ T♠)

**Recommendations for Next Steps:**

1. **Pot Display Enhancement**: Consider showing total pot + current round bets separately in UI (e.g., "Pot: 150 + 75 in bets")
2. **Animation**: Add visual transitions when PlayerBets sweep into Pot at street changes
3. **Side Pot Display**: When multiple players all-in, clearly show main pot vs side pots
4. **Card Placeholder Styling**: Consider using a CSS class for card backs for easier theming
5. **Performance**: Profile table state updates on large tables (6+ players, many actions)
6. **Player Action Feedback**: Add visual feedback when chips move from player to bet area

**Technical Details:**

The pot accounting fix implements a "delayed sweep" pattern:
- During a betting round: `PlayerBets[i]` holds each player's total bet, `Pot` remains at previous street's value
- At street transition: `AdvanceStreet()` sweeps all `PlayerBets` into `Pot`, resets `PlayerBets`
- At showdown: `HandleShowdown()` performs final sweep before calculating winners
- This ensures accurate pot distribution while maintaining clean bet tracking

The frontend display fixes address timing issues where the old `pot > 0` condition failed with the new pot accounting, since pot stays at 0 during preflop betting until the flop is dealt. The new hand-in-progress detection uses presence of hole cards or dealer position to determine if a hand is active, which works correctly across all game states.
