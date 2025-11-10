## Phase 1 Complete: Fix Missing Showdown Trigger (Backend)

Added showdown trigger when betting completes on river with multiple active players. Previously, hands would stall after final river action; now showdown executes automatically to determine winner and settle pot.

**Files created/changed:**
- internal/server/handlers.go
- internal/server/handlers_test.go

**Functions created/changed:**
- `HandlePlayerAction()` - Added else branch to trigger showdown when river betting completes
- `TestHandleAction_RiverBettingCompleteTriggersShowdown()` - New test verifying showdown trigger
- `TestHandleAction_RiverNoShowdownIfNotComplete()` - New test verifying no premature showdown

**Tests created/changed:**
- TestHandleAction_RiverBettingCompleteTriggersShowdown - Verifies showdown is called when betting completes on river
- TestHandleAction_RiverNoShowdownIfNotComplete - Verifies showdown is NOT called if betting incomplete

**Review Status:** APPROVED

**Git Commit Message:**
```
fix: trigger showdown when river betting completes

- Add else branch in HandlePlayerAction() to call HandleShowdown() when betting completes on river
- Previously hands would stall after final river action with no winner determination
- Follows existing mutex unlock/lock pattern for safe broadcast handling
- Add two tests verifying showdown triggers correctly on river completion
```
