## Phase 1 Complete: Fix Start Hand Button Visibility Logic

Successfully implemented the button visibility logic to show the "Start Hand" button at the correct times: when it's the first hand (pot === 0) or after the winner has been determined (handComplete present).

**Files created/changed:**
- `frontend/src/components/TableView.tsx`
- `frontend/src/components/TableView.test.tsx`

**Functions created/changed:**
- `TableView` component - Implemented `showStartHandButton` logic (line 118)
  - Shows button when: `isSeated && (isFirstHand || isHandComplete)`
  - `isFirstHand` checks: `!gameState || gameState.pot === 0 || gameState.pot === undefined`
  - `isHandComplete` checks: `gameState?.handComplete !== undefined`

**Tests created/changed:**
- `Start Hand button shows when pot is 0 and player is seated (first hand)` - Verifies button shows on first hand
- `Start Hand button shows when handComplete is present (after winner determined)` - Verifies button shows after showdown
- `Start Hand button hides during active hand play (pot > 0, no handComplete)` - Verifies button hidden during hand
- `Start Hand button shows even when showdown overlay is dismissed` - Verifies button persistence

**Why This Works:**

The visibility logic ensures:
1. On first hand: `pot === 0` → button visible ✓
2. During active hand: `pot > 0` AND `handComplete === undefined` → button hidden ✓
3. After showdown: `handComplete !== undefined` → button visible ✓

**Review Status:** APPROVED ✅

**Git Commit Message:**
```
fix: Fix Start Hand button visibility logic

- Button shows when pot is 0 (first hand of lobby)
- Button shows when handComplete is present (after winner determined)
- Button hides during active hand play (pot > 0, no handComplete)
- Add 4 comprehensive tests covering all scenarios
```
