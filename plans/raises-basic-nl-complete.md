## Plan Complete: Raises (Basic NL)

The Raises (Basic NL) feature is now fully implemented and operational. Players can now make raise actions with proper validation, min/max raise computation, side pot prevention, and a complete UI with preset buttons. The implementation includes backend logic for raise validation and amount tracking, WebSocket protocol updates to handle raise amounts, and a complete frontend UI with raise presets (Min/Pot/All-in) and real-time validation.

**Phases Completed:** 6 of 6
1. ✅ Phase 1: Min-Raise Computation and Validation
2. ✅ Phase 2: Max-Raise and Side Pot Prevention
3. ✅ Phase 3: Raise Action Processing
4. ✅ Phase 4: Handler Protocol Updates
5. ✅ Phase 5: Frontend Protocol and State
6. ✅ Phase 6: Raise UI Components

**All Files Created/Modified:**

**Backend:**
- internal/server/table.go
- internal/server/table_test.go
- internal/server/handlers.go
- internal/server/handlers_test.go
- internal/server/server.go

**Frontend:**
- frontend/src/components/TableView.tsx
- frontend/src/components/TableView.test.tsx
- frontend/src/styles/TableView.css
- frontend/src/hooks/useWebSocket.ts
- frontend/src/hooks/useWebSocket.test.ts

**Key Functions/Classes Added:**

**Backend:**
- Hand.GetMinRaise() - Calculate minimum valid raise amount
- Hand.GetMaxRaise() - Calculate maximum raise (limits to prevent side pots)
- Hand.ValidateRaise() - Validate raise amounts with clear error messages
- Hand.ProcessAction() - Extended to handle raise actions with amounts
- Hand.GetValidActions() - Extended to include "raise" action
- Server.BroadcastActionRequest() - Extended to include MinRaise/MaxRaise in payload
- Server.HandlePlayerAction() - Extended to accept optional raise amount
- Client.HandlePlayerActionMessage() - Extended to extract and pass raise amount

**Frontend:**
- TableView raise controls - Input field + preset buttons (Min/Pot/All-in)
- handleMinRaise() - Set raise to minimum amount
- handlePotRaise() - Set raise to call amount + pot (pot-sized raise)
- handleAllIn() - Set raise to player's stack
- useWebSocket sendAction() - Helper to send player actions with optional amount
- GameState extended with minRaise and maxRaise fields

**Test Coverage:**
- Total tests written: 36 (10 Phase 1, 9 Phase 2, 8 Phase 3, 3 Phase 4, 4 Phase 5, 10 Phase 6)
- Backend tests passing: 49+/49+ ✅
- Frontend tests passing: 167/167 ✅
- All tests verify raise logic, validation, UI components, and integration

**Recommendations for Next Steps:**
- Implement post-flop streets (flop, turn, river) with proper betting round transitions
- Add hand evaluation and showdown logic
- Implement pot calculations for split pots and side pots
- Add re-raise functionality (multiple raises in same round)
- Consider adding time-based action timeouts (bet timer)
- Add action history/log display for better game tracking
