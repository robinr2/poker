## Plan: Fix Postflop Action Order

Fix the bug where action order after the flop is incorrect. Currently, the player who acts first preflop (UTG) continues to act first postflop, but the correct poker rules dictate that the small blind should act first on all postflop streets (flop, turn, river).

**Phases: 3**

### **Phase 1: Add Failing Tests for Postflop Action Order**
- **Objective:** Write tests demonstrating that GetFirstActor() should return different players based on the current street
- **Files/Functions to Modify/Create:**
  - `internal/server/table_test.go` - Add unit tests for GetFirstActor on different streets
- **Tests to Write:**
  - `TestGetFirstActor_Postflop_MultiPlayer` - On flop/turn/river, SB should be first actor
  - `TestGetFirstActor_Postflop_HeadsUp` - On flop/turn/river, BB (non-dealer) should be first actor
  - `TestGetFirstActor_Postflop_WithFoldedSB` - If SB folded, BB should be first actor postflop
- **Steps:**
  1. Write test for multi-player postflop (4 players, SB should act first on flop)
  2. Write test for heads-up postflop (BB/non-dealer should act first on flop)
  3. Write test for edge case: SB folded preflop, BB should act first on flop
  4. Run tests - they should fail showing UTG/dealer acting first instead
  5. Run linter and formatter

### **Phase 2: Fix GetFirstActor to Check Street**
- **Objective:** Update GetFirstActor() to branch on h.Street and return correct first actor for preflop vs postflop
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` - Modify `GetFirstActor()` method
- **Tests to Write:**
  - None (using Phase 1 tests)
- **Steps:**
  1. Add street check at the top of GetFirstActor()
  2. Keep existing logic for "preflop" street (no changes)
  3. Add postflop branch for "flop", "turn", "river" streets
  4. Postflop multi-player: find first active (non-folded) player starting from SB
  5. Postflop heads-up: return BB (non-dealer player)
  6. Run Phase 1 tests - they should now pass
  7. Run all tests to ensure no regressions
  8. Run linter and formatter

### **Phase 3: Integration Testing**
- **Objective:** Verify fix works end-to-end when hands advance from preflop to flop with correct action order
- **Files/Functions to Modify/Create:**
  - `internal/server/table_test.go` - Add integration test
  - `internal/server/handlers_test.go` - Add WebSocket test (if needed)
- **Tests to Write:**
  - `TestHandFlow_ActionOrderChangesPostflop` - Full hand: verify UTG acts first preflop, SB acts first on flop
  - `TestHandFlow_ActionOrderHeadsUpPostflop` - Heads-up hand: dealer acts first preflop, BB acts first on flop
  - `TestHandFlow_ActionOrderWithFolds` - Player folds preflop, action order adjusts correctly on flop
- **Steps:**
  1. Write integration test: 4 players preflop → verify UTG acts first → advance to flop → verify SB acts first
  2. Write integration test: heads-up preflop → dealer first → flop → BB first
  3. Write integration test with folds: UTG folds preflop → flop starts with next active player after SB
  4. Run all tests (should all pass)
  5. Run linter and formatter

---

**Implementation Notes:**

1. **Street Branching**: GetFirstActor should check `h.Street` and branch:
   - If "preflop": use existing logic (UTG in multi-player, dealer in heads-up)
   - If "flop"/"turn"/"river": use postflop logic (SB in multi-player, BB in heads-up)

2. **Postflop Multi-Player Logic**:
   - Find SB seat in active players list
   - Return first active (non-folded) player starting from SB position
   - If SB folded, return next active player (usually BB)

3. **Postflop Heads-Up Logic**:
   - Return the non-dealer player (BB)
   - This is the opposite of preflop where dealer acts first

4. **Folded Players**:
   - Skip folded players when determining first actor
   - Use existing `h.FoldedPlayers` map to check
