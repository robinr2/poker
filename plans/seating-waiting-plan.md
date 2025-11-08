## Plan: Seating & Waiting System

This plan implements seat assignment, single-seat enforcement, and leave/disconnect handling for poker tables. Players can join tables to occupy seats, see which seats are occupied, and leave tables to return to the lobby. The system enforces one seat per player across all tables and clears seats immediately on disconnect. This builds on Features 1-2 to enable actual gameplay positioning before hand dealing begins in Feature 4.

**Phases: 4**

### **Phase 1: Backend Seat Assignment Logic**
- **Objective:** Create methods for assigning players to first available seat and clearing seats
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` (MODIFY) - Add AssignSeat(), ClearSeat(), GetSeatByToken() methods
  - `internal/server/table_test.go` (MODIFY) - Add tests for new seat methods
  - `internal/server/server.go` (MODIFY) - Add FindPlayerSeat() helper to search across all tables
  - `internal/server/server_test.go` (MODIFY) - Add integration tests for FindPlayerSeat
- **Tests to Write:**
  - `TestTableAssignSeat` - Assigns to first empty seat (0-5 sequential)
  - `TestTableAssignSeatWhenFull` - Returns error when all 6 seats occupied
  - `TestTableClearSeat` - Clears seat by token, sets Token to nil
  - `TestTableClearSeatNotFound` - Returns error when token not found
  - `TestTableGetSeatByToken` - Returns seat if player is seated at table
  - `TestTableConcurrentAssignments` - Multiple goroutines assign seats safely
  - `TestServerFindPlayerSeat` - Finds player across all 4 tables
  - `TestServerFindPlayerSeatNotFound` - Returns nil when player not seated
- **Steps:**
  1. Write test for Table.AssignSeat() - first empty seat gets token
  2. Run tests - should fail (method doesn't exist)
  3. Implement AssignSeat() - lock table, iterate seats 0-5, assign to first where Token == nil
  4. Write test for Table.AssignSeat() when table full
  5. Implement error return when no empty seats
  6. Write test for Table.ClearSeat() - sets Token to nil for matching seat
  7. Run tests - should fail
  8. Implement ClearSeat() - lock table, find seat by token, clear Token pointer
  9. Write test for Table.GetSeatByToken() - returns seat or nil
  10. Implement GetSeatByToken() - RLock, iterate seats, return match
  11. Write test for Server.FindPlayerSeat() - searches all tables
  12. Run tests - should fail
  13. Implement FindPlayerSeat() - RLock server, iterate tables calling GetSeatByToken
  14. Write concurrent assignment test with sync.WaitGroup
  15. Run all tests - should pass
  16. Run linter and format code

### **Phase 2: Single Seat Enforcement and Join Protocol**
- **Objective:** Add join_table message handler with validation: player not already seated, table not full
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` (MODIFY) - Add Status field to Seat struct ("empty", "waiting", "active")
  - `internal/server/handlers.go` (MODIFY) - Add JoinTablePayload, SeatAssignedPayload structs
  - `internal/server/handlers.go` (MODIFY) - Add Client.HandleJoinTable() method
  - `internal/server/handlers.go` (MODIFY) - Add Client.SendSeatAssigned() method
  - `internal/server/websocket.go` (MODIFY) - Add "join_table" case to message router
  - `internal/server/websocket_integration_test.go` (MODIFY) - Add join_table protocol tests
- **Tests to Write:**
  - `TestSeatStatusField` - Seat has Status field with valid values
  - `TestHandleJoinTableSuccess` - Player joins empty table, gets seat 0, Status "waiting"
  - `TestHandleJoinTableFull` - Returns error when table full
  - `TestHandleJoinTableAlreadySeated` - Returns error when player already seated elsewhere
  - `TestHandleJoinTableInvalidTableID` - Returns error for non-existent table
  - `TestSeatAssignedMessage` - Client receives seat_assigned with tableId, seatIndex, status
  - `TestJoinTableUpdatesSession` - Session.TableID and Session.SeatIndex updated
  - `TestJoinTableBroadcastsLobbyState` - All clients receive updated lobby_state after join
- **Steps:**
  1. Write test for Seat struct with Status field
  2. Run tests - should fail
  3. Add Status string field to Seat struct, update NewTable to init seats with Status="empty"
  4. Update AssignSeat to set Status="waiting" when assigning
  5. Update ClearSeat to set Status="empty" when clearing
  6. Write tests for JoinTablePayload and SeatAssignedPayload structs
  7. Run tests - should fail
  8. Define JoinTablePayload struct with TableID field
  9. Define SeatAssignedPayload struct with TableID, SeatIndex, Status, PlayerCount fields
  10. Write test for HandleJoinTable - success path
  11. Run tests - should fail
  12. Implement Client.HandleJoinTable() method:
      - Parse payload to get tableId
      - Check if player already seated via FindPlayerSeat (error: "already_seated")
      - Get table by ID (error: "invalid_table")
      - Call table.AssignSeat (error: "table_full")
      - Update session via sessionManager.UpdateSession
      - Send seat_assigned to client
      - Broadcast lobby_state to all clients
  13. Write test for SendSeatAssigned message format
  14. Implement Client.SendSeatAssigned() method
  15. Add "join_table" case in readPump message router calling HandleJoinTable
  16. Write integration tests for error cases (full, already seated, invalid table)
  17. Run all tests - should pass
  18. Run linter and format code

### **Phase 3: Leave Table and Disconnect Handling**
- **Objective:** Implement leave_table message and automatic seat clearing on disconnect
- **Files/Functions to Modify/Create:**
  - `internal/server/handlers.go` (MODIFY) - Add LeaveTablePayload, SeatClearedPayload structs
  - `internal/server/handlers.go` (MODIFY) - Add Client.HandleLeaveTable() method
  - `internal/server/handlers.go` (MODIFY) - Add Client.SendSeatCleared() method
  - `internal/server/server.go` (MODIFY) - Add HandleDisconnect() method
  - `internal/server/websocket.go` (MODIFY) - Add "leave_table" case to message router
  - `internal/server/websocket.go` (MODIFY) - Call HandleDisconnect in readPump defer block
  - `internal/server/websocket_integration_test.go` (MODIFY) - Add leave and disconnect tests
- **Tests to Write:**
  - `TestHandleLeaveTableSuccess` - Player leaves table, seat cleared, session updated
  - `TestHandleLeaveTableNotSeated` - Returns error when player not seated
  - `TestSeatClearedMessage` - Client receives seat_cleared confirmation
  - `TestLeaveTableBroadcastsLobbyState` - All clients receive updated lobby_state
  - `TestHandleDisconnectClearsSeat` - Disconnect clears seat if player was seated
  - `TestHandleDisconnectNoSeat` - Disconnect with no seat doesn't error
  - `TestDisconnectBroadcastsLobbyState` - Remaining clients receive updated lobby_state
- **Steps:**
  1. Write test for LeaveTablePayload (empty payload) and SeatClearedPayload structs
  2. Run tests - should fail
  3. Define LeaveTablePayload struct (empty or just marker)
  4. Define SeatClearedPayload struct (empty or success confirmation)
  5. Write test for HandleLeaveTable - success path
  6. Run tests - should fail
  7. Implement Client.HandleLeaveTable() method:
      - Find player's current seat via FindPlayerSeat
      - If not seated, return error "not_seated"
      - Get table reference
      - Call table.ClearSeat(token)
      - Update session TableID and SeatIndex to nil via UpdateSession
      - Send seat_cleared to client
      - Broadcast lobby_state to all clients
  8. Write test for SendSeatCleared message
  9. Implement Client.SendSeatCleared() method
  10. Add "leave_table" case in readPump message router
  11. Write test for Server.HandleDisconnect() method
  12. Run tests - should fail
  13. Implement Server.HandleDisconnect(token string):
      - Find player's seat via FindPlayerSeat
      - If seated, clear seat and update session to nil values
      - Broadcast lobby_state (or call via hub if needed)
  14. Modify readPump defer block to call server.HandleDisconnect(c.Token) before unregister
  15. Write integration test: connect, join, disconnect, verify seat cleared
  16. Run all tests - should pass
  17. Run linter and format code

### **Phase 4: Frontend Join/Leave UI and Seat Display**
- **Objective:** Create table view showing 6 seats, wire join/leave buttons, handle seat_assigned messages
- **Files/Functions to Modify/Create:**
  - `frontend/src/components/TableView.tsx` (NEW) - 6-seat table display component
  - `frontend/src/components/TableView.test.tsx` (NEW) - TableView component tests
  - `frontend/src/styles/TableView.css` (NEW) - Styling for seat layout
  - `frontend/src/components/LobbyView.tsx` (MODIFY) - Wire onJoinTable to send join_table message
  - `frontend/src/App.tsx` (MODIFY) - Add view state, handle seat_assigned/seat_cleared messages
  - `frontend/src/App.test.tsx` (MODIFY) - Add tests for view switching and seat messages
  - `frontend/src/hooks/useWebSocket.ts` (MODIFY) - Parse seat_assigned/seat_cleared messages
- **Tests to Write:**
  - `TestTableViewRenders6Seats` - Displays 6 seat positions (0-5)
  - `TestTableViewShowsOccupiedSeats` - Displays player names in occupied seats
  - `TestTableViewShowsEmptySeats` - Shows placeholder for empty seats
  - `TestTableViewHighlightsOwnSeat` - Player's own seat has different style
  - `TestTableViewLeaveButton` - Renders Leave button that calls onLeave callback
  - `TestLobbyViewJoinButtonSendsMessage` - Join button sends join_table WebSocket message
  - `TestAppHandlesSeatAssignedMessage` - Switches to table view on seat_assigned
  - `TestAppHandlesSeatClearedMessage` - Switches to lobby view on seat_cleared
  - `TestAppJoinTableIntegration` - Full flow: click join, receive seat_assigned, show table
- **Steps:**
  1. Write test for TableView component with mock seat data
  2. Run tests - should fail (component doesn't exist)
  3. Create TableView.tsx with props: tableId, seats array (6 elements), currentSeatIndex, onLeave callback
  4. Render 6 seat positions in circular/grid layout
  5. Show player name if seat.token exists, "Empty" if null
  6. Highlight current player's seat with CSS class
  7. Add Leave button at top/bottom calling onLeave
  8. Create TableView.css with seat layout styling (circular or grid)
  9. Write test for LobbyView Join button sending join_table message
  10. Run tests - should fail
  11. Modify LobbyView's onJoinTable to call sendMessage with join_table payload
  12. Write test for useWebSocket parsing seat_assigned message
  13. Run tests - should fail
  14. Add seat_assigned and seat_cleared cases to useWebSocket message handler
  15. Store seat info in hook state (or pass through onMessage callback to App)
  16. Write test for App.tsx handling seat_assigned - switches to table view
  17. Run tests - should fail
  18. Add view state to App: "lobby" | "table"
  19. Add currentTableId and currentSeatIndex state
  20. Handle seat_assigned message: set view="table", store tableId/seatIndex
  21. Handle seat_cleared message: set view="lobby", clear tableId/seatIndex
  22. Render LobbyView when view="lobby", TableView when view="table"
  23. Pass sendMessage to TableView's onLeave to send leave_table message
  24. Write integration test: full join and leave flow
  25. Run all tests - should pass
  26. Run linter and format code

**Open Questions:**
1. **Seat Status Display**: Should Phase 4 show seat status ("waiting" vs "active") visually, or just show occupied vs empty? Recommendation: Just occupied/empty for now, status becomes relevant in Feature 4 when dealing cards.
2. **TableView Layout**: Circular (like real poker table) or simple grid? Recommendation: Simple 2-row grid for Phase 4 (3 seats top, 3 seats bottom), can enhance layout later.
3. **Multiple Browser Tabs**: If player has multiple tabs open with same token and joins a table, all tabs should show table view. Is this desired? Recommendation: Yes, all connections with same token see same state.
4. **Session Persistence**: Sessions remain in SessionManager after disconnect, player can reconnect and rejoin. Should we add session expiry/cleanup? Recommendation: Defer to later feature, keep sessions indefinitely for now.
5. **Error Toasts**: Should frontend show error messages via toasts/alerts, or just console.log? Recommendation: Console.log for Phase 4, proper error UI in Feature 12 (Logging & Error Handling).
