## Plan: Start Hand State Reset Hotfix

This plan fixes two bugs discovered after Phase 3: (1) street label persisting after starting a new hand, and (2) Start Hand button remaining visible during hand play. Both bugs stem from incomplete state clearing in the optimistic update.

**Phases: 2**

1. **Phase 1: Clear Street Label on Hand Start**
    - **Objective:** Clear the street field when starting a new hand so "River" doesn't persist
    - **Files/Functions to Modify/Create:** 
      - `frontend/src/hooks/useWebSocket.ts` (lines 607-615) - Add `delete updated.street` to optimistic update
      - `frontend/src/hooks/useWebSocket.ts` (lines 300-306) - Add `street: undefined` to `hand_started` handler
      - `frontend/src/hooks/useWebSocket.test.ts` - Add tests for street clearing
    - **Tests to Write:** 
      - Test: "sendAction clears street when start_hand is sent"
      - Test: "hand_started message clears street field"
      - Test: "street remains undefined until board_dealt is received"
    - **Steps:**
        1. Write 3 tests for street clearing behavior (TDD - red)
        2. Run tests to confirm they fail
        3. Add `delete updated.street;` to optimistic update (line 607-615)
        4. Add `street: undefined` to `hand_started` handler (line 300-306)
        5. Run tests to confirm they pass
        6. Run linter and fix any issues

2. **Phase 2: Fix Start Hand Button Visibility**
    - **Objective:** Remove pot=0 from optimistic update so button hides correctly during hand
    - **Files/Functions to Modify/Create:** 
      - `frontend/src/hooks/useWebSocket.ts` (lines 607-615) - Remove `updated.pot = 0` line
      - `frontend/src/components/TableView.test.tsx` - Add/update tests for button visibility
    - **Tests to Write:** 
      - Test: "Start Hand button hides immediately after clicking (optimistic)"
      - Test: "Start Hand button remains hidden while pot is undefined"
      - Test: "Start Hand button shows again only when handComplete is set"
    - **Steps:**
        1. Write 3 tests for button visibility after start_hand action (TDD - red)
        2. Run tests to confirm they fail
        3. Remove `updated.pot = 0;` line from optimistic update (line 613)
        4. Run tests to confirm they pass
        5. Verify board cards still hide correctly (they should - boardCards=[] handles this)
        6. Run linter and fix any issues
        7. Run full frontend test suite to ensure no regressions

**Open Questions: 0**
