## Phase 1 Complete: Backend Seat Assignment Logic

Phase 1 successfully implements core backend methods for seat assignment, seat clearing, and player lookup across tables. All methods are thread-safe and return values by copy to prevent race conditions. The implementation includes comprehensive test coverage with 10 new tests, all passing with race detector enabled.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go
- internal/server/server.go
- internal/server/server_test.go

**Functions created/changed:**
- Table.AssignSeat(token *string) (Seat, error) - Assigns player to first empty seat (0-5)
- Table.ClearSeat(token *string) error - Clears seat by token
- Table.GetSeatByToken(token *string) (Seat, bool) - Finds seat by token
- Server.FindPlayerSeat(token *string) *Seat - Searches all tables for player

**Tests created/changed:**
- TestTableAssignSeat - Sequential seat assignment
- TestTableAssignSeatSequential - Verifies 0-5 order
- TestTableAssignSeatWhenFull - Error when table full
- TestTableClearSeat - Clears seat successfully
- TestTableClearSeatNotFound - Error when token not found
- TestTableGetSeatByToken - Returns seat when found
- TestTableGetSeatByTokenNotFound - Returns false when not found
- TestTableConcurrentAssignments - Thread safety with 6 goroutines
- TestServerFindPlayerSeat - Finds player across all tables
- TestServerFindPlayerSeatNotFound - Returns nil when not seated

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: implement backend seat assignment logic

- Add AssignSeat() for sequential seat assignment (0-5)
- Add ClearSeat() to remove player from seat
- Add GetSeatByToken() to find player's seat at table
- Add FindPlayerSeat() to search across all tables
- Return values by copy to prevent race conditions
- All methods thread-safe with RWMutex locks
- 10 new tests, race detector clean
```
