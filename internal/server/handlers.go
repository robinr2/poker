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

// HandStartedPayload represents the payload for hand_started messages
type HandStartedPayload struct {
	DealerSeat     int `json:"dealerSeat"`
	SmallBlindSeat int `json:"smallBlindSeat"`
	BigBlindSeat   int `json:"bigBlindSeat"`
}

// BlindPostedPayload represents the payload for blind_posted messages
type BlindPostedPayload struct {
	SeatIndex int `json:"seatIndex"`
	Amount    int `json:"amount"`
	NewStack  int `json:"newStack"`
}

// CardsDealtPayload represents the payload for cards_dealt messages (with privacy-filtered hole cards)
type CardsDealtPayload struct {
	HoleCards map[int][]Card `json:"holeCards"`
}

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
// Note: Only clients NOT at a table receive lobby_state (clients at tables only receive table_state)
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

	// Send directly to all hub clients except excludeClient and clients at a table
	// (to avoid ordering issues with direct sends)
	s.hub.mu.RLock()
	for client := range s.hub.clients {
		if client != excludeClient {
			// Check if this client is at a table (skip if they are)
			session, err := s.sessionManager.GetSession(client.Token)
			if err != nil {
				s.logger.Warn("failed to get session for client", "token", client.Token, "error", err)
				continue
			}

			// Only send to clients NOT at a table
			if session.TableID == nil {
				select {
				case client.send <- responseBytes:
				default:
					s.logger.Warn("client send channel full, skipping message")
				}
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

	// Send table_state to the joining client
	err = c.SendTableState(server, table.ID, logger)
	if err != nil {
		logger.Warn("failed to send table_state to joining client", "error", err)
	}

	// Broadcast table_state to other players at the table (excluding the joining player)
	err = server.broadcastTableState(table.ID, c)
	if err != nil {
		logger.Warn("failed to broadcast table_state", "error", err)
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

	// Broadcast table_state to remaining players at the table BEFORE broadcasting lobby_state
	err = server.broadcastTableState(table.ID, nil)
	if err != nil {
		logger.Warn("failed to broadcast table_state after leave", "error", err)
	}

	// Send updated lobby_state to the client who left
	err = c.SendLobbyState(server, logger)
	if err != nil {
		logger.Warn("failed to send lobby state to leaving client", "error", err)
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

// TableStateSeat represents a single seat in the table_state message
type TableStateSeat struct {
	Index      int     `json:"index"`
	PlayerName *string `json:"playerName"`
	Status     string  `json:"status"`
}

// TableStatePayload represents the payload for table_state messages
type TableStatePayload struct {
	TableId string           `json:"tableId"`
	Seats   []TableStateSeat `json:"seats"`
}

// SendTableState sends a table_state message to a single client
func (c *Client) SendTableState(server *Server, tableID string, logger *slog.Logger) error {
	// Get the table
	var table *Table
	server.mu.RLock()
	for _, t := range server.tables {
		if t != nil && t.ID == tableID {
			table = t
			break
		}
	}
	server.mu.RUnlock()

	if table == nil {
		return fmt.Errorf("table not found: %s", tableID)
	}

	// Build seats array with player names
	seats := make([]TableStateSeat, 6)
	table.mu.RLock()
	for i, seat := range table.Seats {
		seats[i].Index = i
		seats[i].Status = seat.Status

		if seat.Token != nil {
			// Get player name by token
			playerName, err := server.sessionManager.GetPlayerName(*seat.Token)
			if err != nil {
				logger.Warn("failed to get player name", "token", *seat.Token, "error", err)
				seats[i].PlayerName = nil
			} else {
				seats[i].PlayerName = &playerName
			}
		} else {
			seats[i].PlayerName = nil
		}
	}
	table.mu.RUnlock()

	// Create payload
	payloadObj := TableStatePayload{
		TableId: tableID,
		Seats:   seats,
	}

	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	response := WebSocketMessage{
		Type:    "table_state",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	logger.Info("table_state sent to client", "tableId", tableID)

	c.send <- responseBytes
	return nil
}

// broadcastTableState sends the current table state to all clients at a specific table except the sender
func (s *Server) broadcastTableState(tableID string, excludeClient *Client) error {
	// Get all clients at the table
	clients := s.GetClientsAtTable(tableID)
	s.logger.Info("broadcastTableState", "tableID", tableID, "num_clients", len(clients), "excludeClient", excludeClient != nil)

	// Get the table
	var table *Table
	s.mu.RLock()
	for _, t := range s.tables {
		if t != nil && t.ID == tableID {
			table = t
			break
		}
	}
	s.mu.RUnlock()

	if table == nil {
		return fmt.Errorf("table not found: %s", tableID)
	}

	// Build seats array with player names
	seats := make([]TableStateSeat, 6)
	table.mu.RLock()
	for i, seat := range table.Seats {
		seats[i].Index = i
		seats[i].Status = seat.Status

		if seat.Token != nil {
			// Get player name by token
			playerName, err := s.sessionManager.GetPlayerName(*seat.Token)
			if err != nil {
				s.logger.Warn("failed to get player name", "token", *seat.Token, "error", err)
				seats[i].PlayerName = nil
			} else {
				seats[i].PlayerName = &playerName
			}
		} else {
			seats[i].PlayerName = nil
		}
	}
	table.mu.RUnlock()

	// Create payload
	payloadObj := TableStatePayload{
		TableId: tableID,
		Seats:   seats,
	}

	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	response := WebSocketMessage{
		Type:    "table_state",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Send to all clients at the table except the excludeClient
	for _, client := range clients {
		if excludeClient != nil && client == excludeClient {
			continue
		}
		select {
		case client.send <- responseBytes:
		default:
			s.logger.Warn("client send channel full, skipping table_state message")
		}
	}

	return nil
}

// broadcastHandStarted sends hand_started message to all clients at the table with dealer and blind info
func (s *Server) broadcastHandStarted(table *Table) error {
	// Get all clients at the table
	clients := s.GetClientsAtTable(table.ID)

	table.mu.RLock()
	hand := table.CurrentHand
	if hand == nil {
		table.mu.RUnlock()
		return fmt.Errorf("CurrentHand is nil")
	}
	dealerSeat := *table.DealerSeat
	sbSeat := hand.SmallBlindSeat
	bbSeat := hand.BigBlindSeat
	table.mu.RUnlock()

	// Create payload
	payloadObj := HandStartedPayload{
		DealerSeat:     dealerSeat,
		SmallBlindSeat: sbSeat,
		BigBlindSeat:   bbSeat,
	}

	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	response := WebSocketMessage{
		Type:    "hand_started",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Send to all clients at the table
	for _, client := range clients {
		select {
		case client.send <- responseBytes:
		default:
			s.logger.Warn("client send channel full, skipping hand_started message")
		}
	}

	return nil
}

// broadcastBlindPosted sends blind_posted message to all clients at the table
func (s *Server) broadcastBlindPosted(table *Table, seatNum int, amount int) error {
	// Get all clients at the table
	clients := s.GetClientsAtTable(table.ID)

	// Get the player's new stack
	table.mu.RLock()
	newStack := table.Seats[seatNum].Stack
	table.mu.RUnlock()

	// Create payload
	payloadObj := BlindPostedPayload{
		SeatIndex: seatNum,
		Amount:    amount,
		NewStack:  newStack,
	}

	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	response := WebSocketMessage{
		Type:    "blind_posted",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Send to all clients at the table
	for _, client := range clients {
		select {
		case client.send <- responseBytes:
		default:
			s.logger.Warn("client send channel full, skipping blind_posted message")
		}
	}

	return nil
}

// broadcastCardsDealt sends cards_dealt messages to all clients at the table with privacy filtering
// Each player receives only their own hole cards
func (s *Server) broadcastCardsDealt(table *Table) error {
	// Get all clients at the table
	clients := s.GetClientsAtTable(table.ID)

	table.mu.RLock()
	hand := table.CurrentHand
	if hand == nil {
		table.mu.RUnlock()
		return fmt.Errorf("CurrentHand is nil")
	}
	holeCards := hand.HoleCards
	table.mu.RUnlock()

	// Send personalized message to each player
	for _, client := range clients {
		// Find which seat this client is at
		session, err := s.sessionManager.GetSession(client.Token)
		if err != nil {
			s.logger.Warn("failed to get session for client", "token", client.Token, "error", err)
			continue
		}

		if session.SeatIndex == nil {
			s.logger.Warn("client has no seat assigned", "token", client.Token)
			continue
		}

		seatIndex := *session.SeatIndex

		// Filter hole cards to only show this player's cards
		filteredCards := filterHoleCardsForPlayer(holeCards, seatIndex)

		// Create payload
		payloadObj := CardsDealtPayload{
			HoleCards: filteredCards,
		}

		payloadBytes, err := json.Marshal(payloadObj)
		if err != nil {
			s.logger.Warn("failed to marshal cards_dealt payload", "error", err)
			continue
		}

		response := WebSocketMessage{
			Type:    "cards_dealt",
			Payload: json.RawMessage(payloadBytes),
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			s.logger.Warn("failed to marshal cards_dealt response", "error", err)
			continue
		}

		// Send to this client
		select {
		case client.send <- responseBytes:
		default:
			s.logger.Warn("client send channel full, skipping cards_dealt message")
		}
	}

	return nil
}
