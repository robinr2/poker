## Phase 1 Complete: Add Stack Field to TableStateSeat Struct

Successfully added the Stack field to TableStateSeat struct and ensured it's properly populated in both SendTableState and broadcastTableState functions.

**Files created/changed:**
- internal/server/handlers.go
- internal/server/handlers_test.go

**Functions created/changed:**
- `TableStateSeat` struct - Added `Stack *int` field
- `SendTableState` function - Populates Stack field from seat data
- `broadcastTableState` function - Populates Stack field from seat data (bug fix)

**Tests created/changed:**
- `TestTableStateSeatIncludesStack` - Verifies Stack field in table_state payload
- `TestTableStateSerializationWithStacks` - Verifies JSON serialization with stacks
- `TestBroadcastTableStateIncludesStack` - Verifies Stack in broadcast scenarios

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: add stack field to table_state message

- Add Stack field to TableStateSeat struct with proper JSON serialization
- Populate stack values in SendTableState for direct sends
- Fix broadcastTableState to include stack values in broadcasts
- Add comprehensive tests for stack inclusion and serialization
```
