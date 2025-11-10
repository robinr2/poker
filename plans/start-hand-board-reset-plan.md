## Plan: Fix Start Hand Button and Board Cards Reset

This plan fixes two bugs: (1) the "Start Hand" button showing at the wrong times, and (2) board cards persisting from the previous hand until the flop.

**Phases: 3**

1. **Phase 1: Fix Start Hand Button Visibility Logic**
    - **Objective:** Button shows when: (a) first hand of lobby, or (b) after winner determined (handComplete present)
    - **Files/Functions to Modify/Create:** 
      - `frontend/src/components/TableView.tsx` (lines 112-118) - `showStartHandButton` logic
      - `frontend/src/components/TableView.test.tsx` - update tests for button visibility
    - **Tests to Write:** 
      - Test: "Start Hand button shows when pot is 0 and player is seated (first hand)"
      - Test: "Start Hand button shows when handComplete is present (after winner determined)"
      - Test: "Start Hand button hides during active hand play (pot > 0, no handComplete)"
      - Test: "Start Hand button shows even when showdown overlay is dismissed"
    - **Steps:**
        1. Write tests for correct button visibility behavior (TDD - red)
        2. Run tests to confirm they fail
        3. Update `showStartHandButton` logic to: `isSeated && (hasNoActiveHand || gameState?.handComplete !== undefined)`
        4. Run tests to confirm they pass
        5. Run linter and fix any issues

2. **Phase 2: Clear Board Cards on New Hand Start (Optimistic Update)**
    - **Objective:** Clear boardCards array values when start_hand action is sent, keeping empty placeholders visible
    - **Files/Functions to Modify/Create:** 
      - `frontend/src/hooks/useWebSocket.ts` (lines 605-613) - `sendAction` function for `start_hand` action
      - `frontend/src/hooks/useWebSocket.test.ts` - update tests for boardCards clearing
    - **Tests to Write:** 
      - Test: "sendAction clears boardCards array when start_hand is sent"
      - Test: "boardCards is empty array (not undefined) after start_hand action"
    - **Steps:**
        1. Write tests for boardCards clearing on start_hand action (TDD - red)
        2. Run tests to confirm they fail
        3. Add `updated.boardCards = []` to the `start_hand` action handler in sendAction
        4. Run tests to confirm they pass
        5. Run linter and fix any issues

3. **Phase 3: Clear Board Cards on Hand Started Message (Backend Confirmation)**
    - **Objective:** Ensure boardCards array is cleared when backend confirms new hand started
    - **Files/Functions to Modify/Create:** 
      - `frontend/src/hooks/useWebSocket.ts` (lines 288-314) - `hand_started` message handler
      - `frontend/src/hooks/useWebSocket.test.ts` - update tests for boardCards clearing
    - **Tests to Write:** 
      - Test: "hand_started message sets boardCards to empty array"
      - Test: "hand_started preserves dealer/blind seats and other state"
    - **Steps:**
        1. Write tests for boardCards clearing on hand_started message (TDD - red)
        2. Run tests to confirm they fail
        3. Add `boardCards: []` to the setGameState call in hand_started handler
        4. Run tests to confirm they pass
        5. Run linter and fix any issues

**Open Questions: 0**
