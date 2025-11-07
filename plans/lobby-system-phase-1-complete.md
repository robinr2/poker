## Phase 1 Complete: Backend Table Structure and Preseeding

Successfully implemented backend table data structures and preseeded 4 poker tables (6-max) on server startup. The foundation provides thread-safe table management with proper initialization and seat tracking.

**Files created/changed:**
- internal/server/table.go (NEW)
- internal/server/table_test.go (NEW)
- internal/server/server.go (MODIFIED)

**Functions created/changed:**
- `NewTable(id, name string) *Table` - Constructor that initializes table with 6 empty seats
- `Table.GetOccupiedSeatCount() int` - Thread-safe method to count occupied seats using RLock
- `NewServer()` - Modified to preseed 4 tables: "table-1"/"Table 1" through "table-4"/"Table 4"

**Structs created:**
- `Seat` - Index (0-5), Token (*string, nil for empty)
- `Table` - ID, Name, MaxSeats (6), Seats [6]Seat, sync.RWMutex

**Tests created:**
- `TestNewTable` - Verifies table creation with correct ID, name, MaxSeats=6
- `TestSeatInitialization` - Verifies all 6 seats have Index 0-5 and nil Token
- `TestGetOccupiedSeatCount` - Verifies returns 0 for empty table
- `TestGetOccupiedSeatCountWithOccupiedSeats` - Verifies count with 3 and 6 occupied seats
- `TestTableThreadSafety` - Concurrent reads/writes with RWMutex (5 goroutines, 50 ops each)
- `TestServerTablesPreseeded` - Verifies NewServer creates 4 tables with correct IDs/names

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: Add backend table structure with 4 preseeded tables

- Create Seat struct with Index (0-5) and Token (*string for occupancy)
- Create Table struct with ID, Name, MaxSeats=6, fixed [6]Seat array, sync.RWMutex
- Implement NewTable() constructor initializing 6 empty seats
- Implement GetOccupiedSeatCount() with thread-safe RLock access
- Add tables [4]*Table field to Server struct
- Preseed 4 tables in NewServer(): "Table 1" through "Table 4"
- Add 6 comprehensive tests (initialization, counting, thread safety, preseeding)
```
