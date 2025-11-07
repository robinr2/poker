## Plan: Lobby System with Real-Time Seat Counts

This plan implements the lobby system that displays 4 preseeded poker tables with real-time seat availability. Players will see all tables, their seat counts (occupied/total), and can identify which tables have open seats. This builds on the Identity & Rejoin system to enable table discovery before implementing join functionality.

**Phases: 4**

### **Phase 1: Backend Table Structure and Preseeding**
- **Objective:** Create Table and Seat data structures, preseed 4 tables (6-max) on server startup
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` (NEW) - Table, Seat structs
  - `internal/server/table_test.go` (NEW) - Unit tests
  - `internal/server/server.go` (MODIFY) - Add tables field to Server, preseed in NewServer()
- **Tests to Write:**
  - `TestNewTable` - Table creation with ID, name, 6 empty seats
  - `TestSeatStructure` - Seat fields (Index, Token, Status)
  - `TestGetOccupiedSeatCount` - Count seats with non-nil Token
  - `TestTableThreadSafety` - Concurrent seat access with RWMutex
  - `TestServerTablesPreseeded` - NewServer creates 4 tables with correct IDs/names
- **Steps:**
  1. Write tests for Seat struct (Index 0-5, Token *string, Status enum)
  2. Write tests for Table struct (ID, Name, MaxSeats=6, Seats [6]Seat, RWMutex)
  3. Run tests - should fail (types don't exist yet)
  4. Create Seat struct with Index, Token, Status fields
  5. Create Table struct with ID, Name, MaxSeats, Seats array, sync.RWMutex
  6. Implement NewTable(id, name string) constructor
  7. Implement Table.GetOccupiedSeatCount() method (counts non-nil Tokens)
  8. Add tables [4]*Table field to Server struct
  9. Modify NewServer() to preseed tables: "table-1", "table-2", "table-3", "table-4"
  10. Run tests - should pass
  11. Run linter and format code

### **Phase 2: Lobby State and WebSocket Events**
- **Objective:** Add lobby_state message type, broadcast lobby state on client connect and seat changes
- **Files/Functions to Modify/Create:**
  - `internal/server/handlers.go` (MODIFY) - Add GetLobbyState() method to Server
  - `internal/server/websocket.go` (MODIFY) - Send lobby_state on connection, add broadcast helper
  - `internal/server/handlers_test.go` (MODIFY) - Test GetLobbyState returns correct data
  - `internal/server/websocket_integration_test.go` (MODIFY) - Test lobby_state message sent on connect
- **Tests to Write:**
  - `TestGetLobbyState` - Returns array with 4 tables, correct seat counts
  - `TestGetLobbyStateWithOccupiedSeats` - Reflects occupied seats correctly
  - `TestWebSocketSendsLobbyStateOnConnect` - Client receives lobby_state after connection
  - `TestLobbyStateMessageFormat` - Verify JSON structure matches protocol
- **Steps:**
  1. Write test for Server.GetLobbyState() returning TableInfo array
  2. Run tests - should fail (method doesn't exist)
  3. Define TableInfo struct (ID, Name, SeatsOccupied, MaxSeats)
  4. Implement Server.GetLobbyState() - iterate tables, build TableInfo array
  5. Write test for lobby_state message sent on WebSocket connect
  6. Run tests - should fail (message not sent)
  7. Modify HandleWebSocket() or readPump() to send lobby_state after session established
  8. Implement helper: Server.broadcastLobbyState() using hub.broadcast
  9. Run tests - should pass
  10. Run linter and format code

### **Phase 3: Frontend Lobby UI Components**
- **Objective:** Create LobbyView component displaying 4 tables with seat counts and Join buttons
- **Files/Functions to Modify/Create:**
  - `frontend/src/components/LobbyView.tsx` (NEW) - Main lobby container
  - `frontend/src/components/LobbyView.test.tsx` (NEW) - Component tests
  - `frontend/src/components/TableCard.tsx` (NEW) - Individual table display
  - `frontend/src/components/TableCard.test.tsx` (NEW) - Component tests
  - `frontend/src/styles/LobbyView.css` (NEW) - Styling for 2x2 grid
  - `frontend/src/App.tsx` (MODIFY) - Add LobbyView below NamePrompt
- **Tests to Write:**
  - `TestLobbyViewRendersEmptyState` - Shows loading when no tables
  - `TestLobbyViewRendersFourTables` - Renders 4 TableCard components
  - `TestTableCardDisplaysInfo` - Shows table name, seat count (e.g., "3/6")
  - `TestTableCardJoinButton` - Renders "Join" button
  - `TestTableCardJoinButtonDisabled` - Button disabled when table full (6/6)
  - `TestTableCardJoinButtonEnabled` - Button enabled when seats available
- **Steps:**
  1. Write test for TableCard component with mock table data
  2. Run tests - should fail (component doesn't exist)
  3. Create TableCard.tsx with props: table (id, name, seatsOccupied, maxSeats), onJoin callback
  4. Render table name, seat count "X/Y", Join button (disabled if full)
  5. Write tests for LobbyView component rendering 4 tables
  6. Run tests - should fail (LobbyView doesn't exist)
  7. Create LobbyView.tsx with props: tables array, onJoinTable callback
  8. Map tables to TableCard components in 2x2 grid layout
  9. Create LobbyView.css with grid styling (2 columns, responsive)
  10. Modify App.tsx to render LobbyView (onJoinTable as no-op for now)
  11. Run tests - should pass
  12. Run linter and format code

### **Phase 4: Real-Time Lobby State Updates**
- **Objective:** Wire WebSocket lobby_state messages to React state, update UI in real-time
- **Files/Functions to Modify/Create:**
  - `frontend/src/services/WebSocketService.ts` (MODIFY) - Add lobby_state message handler
  - `frontend/src/hooks/useWebSocket.ts` (MODIFY) - Expose lobbyState in hook return
  - `frontend/src/App.tsx` (MODIFY) - Pass lobbyState from hook to LobbyView
  - `frontend/src/services/WebSocketService.test.ts` (MODIFY) - Test message parsing
  - `frontend/src/hooks/useWebSocket.test.ts` (MODIFY) - Test state updates
- **Tests to Write:**
  - `TestWebSocketServiceParsesLobbyState` - Parses lobby_state message correctly
  - `TestUseWebSocketHookExposesLobbyState` - Hook returns lobbyState array
  - `TestLobbyStateUpdatesOnMessage` - State updates when lobby_state received
  - `TestAppPassesLobbyStateToLobbyView` - App component wires state to LobbyView
- **Steps:**
  1. Write test for WebSocketService handling lobby_state message type
  2. Run tests - should fail (handler doesn't exist)
  3. Add lobby_state case to WebSocketService message router
  4. Store parsed table array in WebSocketService internal state
  5. Write test for useWebSocket hook exposing lobbyState
  6. Run tests - should fail (not exposed in hook)
  7. Add lobbyState to useWebSocket return value (array of TableInfo)
  8. Modify App.tsx to destructure lobbyState from useWebSocket()
  9. Pass lobbyState to LobbyView component
  10. Run tests - should pass
  11. Run integration test: Start server, connect client, verify lobby_state updates
  12. Run linter and format code

**Design Decisions (Approved):**
1. **Table Naming Convention**: Human-readable "Table 1", "Table 2", "Table 3", "Table 4" for better UX
2. **Seat Status**: Use nil/non-nil Token for Phase 1, add SeatStatus enum in Feature 3 when implementing join/reserve
3. **Loading State**: Display "Loading tables..." spinner until first lobby_state message received
4. **Table Full UI**: Keep disabled Join button for consistency, add visual distinction via CSS
5. **Broadcast Timing**: Immediate broadcast on any seat change, optimize batching later if needed
