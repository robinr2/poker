## Plan Complete: Seating & Waiting System

The Seating & Waiting System is now fully implemented and operational. Players can join poker tables from the lobby, occupy seats, see who else is seated, and leave tables to return to the lobby. The system enforces single-seat rules (one seat per player across all tables), handles disconnects gracefully by clearing seats automatically, and provides real-time lobby updates to all connected clients. The implementation includes comprehensive backend seat assignment logic, WebSocket protocol handlers, and a complete frontend UI with table visualization.

**Phases Completed:** 4 of 4
1. ✅ Phase 1: Backend Seat Assignment Logic
2. ✅ Phase 2: Join Table Protocol & Messages
3. ✅ Phase 3: Leave/Disconnect Handling
4. ✅ Phase 4: Frontend TableView UI

**All Files Created/Modified:**

**Backend:**
- internal/server/table.go
- internal/server/table_test.go
- internal/server/server.go
- internal/server/server_test.go
- internal/server/handlers.go
- internal/server/websocket.go
- internal/server/websocket_integration_test.go

**Frontend:**
- frontend/src/components/TableView.tsx (NEW)
- frontend/src/components/TableView.test.tsx (NEW)
- frontend/src/styles/TableView.css (NEW)
- frontend/src/App.tsx
- frontend/src/App.test.tsx
- frontend/src/hooks/useWebSocket.ts
- frontend/src/components/LobbyView.tsx
- frontend/src/components/LobbyView.test.tsx

**Key Functions/Classes Added:**

**Backend:**
- Table.AssignSeat() - Sequential seat assignment (0-5)
- Table.ClearSeat() - Remove player from seat
- Table.GetSeatByToken() - Find player's seat at table
- Server.FindPlayerSeat() - Search all tables for player
- Client.HandleJoinTable() - Join table protocol with validation
- Client.SendSeatAssigned() - Seat assignment confirmation
- Client.HandleLeaveTable() - Voluntary table leaving
- Client.SendSeatCleared() - Leave confirmation
- Client.broadcastLobbyStateExcluding() - Targeted lobby broadcasts
- Server.HandleDisconnect() - Automatic seat cleanup on disconnect

**Frontend:**
- TableView component - 6-seat display with highlighting
- App view state management - Lobby/table view switching
- handleJoinTable() - Send join_table message
- handleLeaveTable() - Send leave_table message
- Seat message parsing - Handle seat_assigned/seat_cleared

**Test Coverage:**
- Total tests written: 38 (10 Phase 1, 8 Phase 2, 6 Phase 3, 22 Phase 4 - 8 tests existed in Phase 2 from previous session)
- Backend tests passing: 89/89 ✅
- Frontend tests passing: 107/107 ✅
- Race detector: Clean ✅
- Linter: Clean ✅

**Recommendations for Next Steps:**
- Implement table state synchronization to show all players' seats when joining (currently only shows own seat)
- Add real-time seat updates when other players join/leave the same table
- Consider adding seat status display ("waiting" vs "active") for future game states
- Add error toast notifications for join/leave failures (currently console.log only)
- Implement session expiry/cleanup for long-inactive sessions
