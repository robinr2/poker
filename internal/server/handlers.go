package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// TableInfo represents table information for the lobby view
type TableInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	SeatsOccupied int    `json:"seats_occupied"`
	MaxSeats      int    `json:"max_seats"`
}

// WebSocketMessage represents a generic WebSocket message structure
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SetNamePayload represents the payload for set_name messages
type SetNamePayload struct {
	Name string `json:"name"`
}

// SessionCreatedPayload represents the payload for session_created messages
type SessionCreatedPayload struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

// SessionRestoredPayload represents the payload for session_restored messages
type SessionRestoredPayload struct {
	Name      string  `json:"name"`
	TableID   *string `json:"tableID,omitempty"`
	SeatIndex *int    `json:"seatIndex,omitempty"`
}

// ErrorPayload represents the payload for error messages
type ErrorPayload struct {
	Message string `json:"message"`
}

// JoinTablePayload represents the payload for join_table messages
type JoinTablePayload struct {
	TableId string `json:"tableId"`
}

// SeatAssignedPayload represents the payload for seat_assigned messages
type SeatAssignedPayload struct {
	TableId   string `json:"tableId"`
	SeatIndex int    `json:"seatIndex"`
	Status    string `json:"status"`
}

// LeaveTablePayload represents the payload for leave_table messages (empty)
type LeaveTablePayload struct{}

// SeatClearedPayload represents the payload for seat_cleared messages (empty)
type SeatClearedPayload struct{}

// HandleSetName processes a set_name message and creates a session for the client
func (c *Client) HandleSetName(sm *SessionManager, server *Server, logger *slog.Logger, payload []byte) error {
	var setNamePayload SetNamePayload
	err := json.Unmarshal(payload, &setNamePayload)
	if err != nil {
		return fmt.Errorf("invalid set_name payload: %w", err)
	}

	// Create a new session
	session, err := sm.CreateSession(setNamePayload.Name)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Update client token
	c.Token = session.Token

	// Send session_created message
	payloadObj := SessionCreatedPayload{
		Token: session.Token,
		Name:  session.Name,
	}
	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	response := WebSocketMessage{
		Type:    "session_created",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	logger.Info("session created via websocket", "token", session.Token, "name", session.Name)

	c.send <- responseBytes

	// Send lobby_state after session_created
	c.SendLobbyState(server, logger)

	return nil
}

// SendSessionRestored sends a session_restored message to the client
func (c *Client) SendSessionRestored(session *Session, logger *slog.Logger) error {
	payloadObj := SessionRestoredPayload{
		Name:      session.Name,
		TableID:   session.TableID,
		SeatIndex: session.SeatIndex,
	}
	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	response := WebSocketMessage{
		Type:    "session_restored",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	logger.Info("session restored via websocket", "token", session.Token, "name", session.Name)

	c.send <- responseBytes
	return nil
}

// SendError sends an error message to the client
func (c *Client) SendError(message string, logger *slog.Logger) error {
	payloadObj := ErrorPayload{
		Message: message,
	}
	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	response := WebSocketMessage{
		Type:    "error",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal error response: %w", err)
	}

	logger.Info("error sent to client", "message", message)

	c.send <- responseBytes
	return nil
}

// SendLobbyState sends the current lobby state to the client
func (c *Client) SendLobbyState(server *Server, logger *slog.Logger) error {
	lobbyState := server.GetLobbyState()

	// First marshal the lobby state to JSON
	payloadBytes, err := json.Marshal(lobbyState)
	if err != nil {
		return fmt.Errorf("failed to marshal lobby state: %w", err)
	}

	// Then marshal it again as a JSON string (double-encode)
	payloadString, err := json.Marshal(string(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to marshal payload string: %w", err)
	}

	response := WebSocketMessage{
		Type:    "lobby_state",
		Payload: json.RawMessage(payloadString),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	logger.Info("lobby_state sent to client")

	c.send <- responseBytes
	return nil
}

// HealthCheckHandler returns an HTTP handler function for the health check endpoint.
func HealthCheckHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]string{
			"status": "ok",
		}

		json.NewEncoder(w).Encode(response)
	}
}

// GetLobbyState returns a slice of TableInfo for all tables in the server
// Thread-safe method using RLock on Server.mu
func (s *Server) GetLobbyState() []TableInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lobbyState := make([]TableInfo, 0, len(s.tables))
	for _, table := range s.tables {
		if table == nil {
			continue
		}
		tableInfo := TableInfo{
			ID:            table.ID,
			Name:          table.Name,
			MaxSeats:      table.MaxSeats,
			SeatsOccupied: table.GetOccupiedSeatCount(),
		}
		lobbyState = append(lobbyState, tableInfo)
	}
	return lobbyState
}

// broadcastLobbyState sends the current lobby state to all connected clients
func (s *Server) broadcastLobbyState() error {
	lobbyState := s.GetLobbyState()

	// First marshal the lobby state to JSON
	payloadBytes, err := json.Marshal(lobbyState)
	if err != nil {
		return fmt.Errorf("failed to marshal lobby state: %w", err)
	}

	// Then marshal it again as a JSON string (double-encode)
	payloadString, err := json.Marshal(string(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to marshal payload string: %w", err)
	}

	response := WebSocketMessage{
		Type:    "lobby_state",
		Payload: json.RawMessage(payloadString),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Send to hub.broadcast channel (non-blocking)
	select {
	case s.hub.broadcast <- responseBytes:
	default:
		// Channel full, skip broadcast
	}

	return nil
}

// broadcastLobbyStateExcluding sends the current lobby state to all connected clients except one
// This is used to send the state to other players after one player makes a change
func (s *Server) broadcastLobbyStateExcluding(excludeClient *Client) error {
	lobbyState := s.GetLobbyState()

	// First marshal the lobby state to JSON
	payloadBytes, err := json.Marshal(lobbyState)
	if err != nil {
		return fmt.Errorf("failed to marshal lobby state: %w", err)
	}

	// Then marshal it again as a JSON string (double-encode)
	payloadString, err := json.Marshal(string(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to marshal payload string: %w", err)
	}

	response := WebSocketMessage{
		Type:    "lobby_state",
		Payload: json.RawMessage(payloadString),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Send directly to all hub clients except excludeClient (to avoid ordering issues with direct sends)
	s.hub.mu.RLock()
	for client := range s.hub.clients {
		if client != excludeClient {
			select {
			case client.send <- responseBytes:
			default:
				s.logger.Warn("client send channel full, skipping message")
			}
		}
	}
	s.hub.mu.RUnlock()

	return nil
}

// HandleJoinTable processes a join_table message and assigns the player to a table seat
func (c *Client) HandleJoinTable(sm *SessionManager, server *Server, logger *slog.Logger, payload []byte) error {
	var joinTablePayload JoinTablePayload
	err := json.Unmarshal(payload, &joinTablePayload)
	if err != nil {
		return fmt.Errorf("invalid join_table payload: %w", err)
	}

	// Verify session exists
	_, err = sm.GetSession(c.Token)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Check if player is already seated at another table
	playerSeat := server.FindPlayerSeat(&c.Token)
	if playerSeat != nil {
		return fmt.Errorf("already_seated")
	}

	// Get table by ID
	var table *Table
	server.mu.RLock()
	for _, t := range server.tables {
		if t != nil && t.ID == joinTablePayload.TableId {
			table = t
			break
		}
	}
	server.mu.RUnlock()

	if table == nil {
		return fmt.Errorf("invalid_table")
	}

	// Assign seat on the table
	seat, err := table.AssignSeat(&c.Token)
	if err != nil {
		return fmt.Errorf("table_full")
	}

	// Update session with table and seat info
	_, err = sm.UpdateSession(c.Token, &table.ID, &seat.Index)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Send seat_assigned message to client
	err = c.SendSeatAssigned(table.ID, seat.Index, seat.Status, logger)
	if err != nil {
		return fmt.Errorf("failed to send seat_assigned: %w", err)
	}

	// Broadcast lobby_state to other clients
	err = server.broadcastLobbyStateExcluding(c)
	if err != nil {
		logger.Warn("failed to broadcast lobby state", "error", err)
	}

	logger.Info("player joined table", "token", c.Token, "tableId", table.ID, "seatIndex", seat.Index)

	return nil
}

// SendSeatAssigned sends a seat_assigned message to the client
func (c *Client) SendSeatAssigned(tableID string, seatIndex int, status string, logger *slog.Logger) error {
	payloadObj := SeatAssignedPayload{
		TableId:   tableID,
		SeatIndex: seatIndex,
		Status:    status,
	}
	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	response := WebSocketMessage{
		Type:    "seat_assigned",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	logger.Info("seat_assigned sent to client", "tableId", tableID, "seatIndex", seatIndex)

	c.send <- responseBytes
	return nil
}

// HandleLeaveTable processes a leave_table message and removes player from their seat
func (c *Client) HandleLeaveTable(sm *SessionManager, server *Server, logger *slog.Logger, payload []byte) error {
	// Verify session exists
	session, err := sm.GetSession(c.Token)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Find player's current seat
	playerSeat := server.FindPlayerSeat(&c.Token)
	if playerSeat == nil {
		return fmt.Errorf("not_seated")
	}

	// Get table reference
	var table *Table
	server.mu.RLock()
	for _, t := range server.tables {
		if t != nil && t.ID == *session.TableID {
			table = t
			break
		}
	}
	server.mu.RUnlock()

	if table == nil {
		return fmt.Errorf("table not found")
	}

	// Clear the seat
	err = table.ClearSeat(&c.Token)
	if err != nil {
		return fmt.Errorf("failed to clear seat: %w", err)
	}

	// Update session to clear TableID and SeatIndex
	_, err = sm.UpdateSession(c.Token, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Send seat_cleared message to client
	err = c.SendSeatCleared(logger)
	if err != nil {
		return fmt.Errorf("failed to send seat_cleared: %w", err)
	}

	// Broadcast lobby_state to other clients
	err = server.broadcastLobbyStateExcluding(c)
	if err != nil {
		logger.Warn("failed to broadcast lobby state", "error", err)
	}

	logger.Info("player left table", "token", c.Token, "tableId", table.ID)

	return nil
}

// SendSeatCleared sends a seat_cleared message to the client
func (c *Client) SendSeatCleared(logger *slog.Logger) error {
	payloadObj := SeatClearedPayload{}
	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	response := WebSocketMessage{
		Type:    "seat_cleared",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	logger.Info("seat_cleared sent to client")

	c.send <- responseBytes
	return nil
}
