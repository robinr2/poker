## Phase 2 Complete: Fix Start Hand Button Visibility

Successfully fixed the bug where the Start Hand button remained visible during hand play. Removed the problematic `pot = 0` from the optimistic update, allowing the button to hide correctly based on the existing visibility logic.

**Files created/changed:**
- `frontend/src/hooks/useWebSocket.ts`
- `frontend/src/components/TableView.test.tsx`

**Functions created/changed:**
- `useWebSocket` - Removed `updated.pot = 0;` from `sendAction` optimistic update (previously line 614)

**Tests created/changed:**
- `Start Hand button hides immediately after clicking (optimistic)` - Verifies button stays hidden after clicking because pot is not reset to 0
- `Start Hand button remains hidden while pot is greater than 0 (active hand)` - Confirms button stays hidden during active hand
- `Start Hand button shows again only when handComplete is set` - Verifies button only reappears when hand completes

**Review Status:** APPROVED ✅ (after fixing minor linting issue)

**Git Commit Message:**
```
fix: Prevent Start Hand button from showing during active hand

- Remove pot=0 from start_hand optimistic update to fix visibility logic
- Button now correctly hides after clicking and stays hidden during hand
- Write 3 tests verifying button visibility in different states
- Fix linting issue (let → const) in test file
- All 227 frontend tests passing
```
