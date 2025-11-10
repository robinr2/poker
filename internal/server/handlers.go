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

// ActionRequestPayload represents the payload for action_request messages
type ActionRequestPayload struct {
	SeatIndex    int      `json:"seatIndex"`
	ValidActions []string `json:"validActions"`
	CallAmount   int      `json:"callAmount"`
	CurrentBet   int      `json:"currentBet"`
	PlayerBet    int      `json:"playerBet"`
	Pot          int      `json:"pot"`
	MinRaise     int      `json:"minRaise"`
	MaxRaise     int      `json:"maxRaise"`
}

// PlayerActionPayload represents the payload for player_action messages
type PlayerActionPayload struct {
	SeatIndex int    `json:"seatIndex"`
	Action    string `json:"action"`
	Amount    *int   `json:"amount,omitempty"`
}

// ActionResultPayload represents the payload for action_result messages
type ActionResultPayload struct {
	SeatIndex   int    `json:"seatIndex"`
	Action      string `json:"action"`
	AmountActed int    `json:"amountActed"`
	NewStack    int    `json:"newStack"`
	Pot         int    `json:"pot"`
	NextActor   *int   `json:"nextActor,omitempty"`
	RoundOver   bool   `json:"roundOver,omitempty"`
	RoundWinner *int   `json:"roundWinner,omitempty"`
}

// BoardDealtPayload represents the payload for board_dealt messages
type BoardDealtPayload struct {
	BoardCards []Card `json:"boardCards"`
	Street     string `json:"street"`
}

// ShowdownResultPayload represents the result of a showdown
type ShowdownResultPayload struct {
	WinnerSeats []int       `json:"winnerSeats"` // Seat indices of winners
	WinningHand string      `json:"winningHand"` // Human-readable hand name
	PotAmount   int         `json:"potAmount"`   // Total pot size
	AmountsWon  map[int]int `json:"amountsWon"`  // Map of seat index to amount won
}

// HandCompletePayload represents hand completion
type HandCompletePayload struct {
	Message string `json:"message"` // Completion message
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

	// Marshal the lobby state to JSON
	payloadBytes, err := json.Marshal(lobbyState)
	if err != nil {
		return fmt.Errorf("failed to marshal lobby state: %w", err)
	}

	response := WebSocketMessage{
		Type:    "lobby_state",
		Payload: json.RawMessage(payloadBytes),
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
	Stack      *int    `json:"stack"`
	CardCount  *int    `json:"cardCount,omitempty"`
}

// TableStatePayload represents the payload for table_state messages
type TableStatePayload struct {
	TableId        string           `json:"tableId"`
	Seats          []TableStateSeat `json:"seats"`
	HandInProgress bool             `json:"handInProgress"`
	DealerSeat     *int             `json:"dealerSeat,omitempty"`
	SmallBlindSeat *int             `json:"smallBlindSeat,omitempty"`
	BigBlindSeat   *int             `json:"bigBlindSeat,omitempty"`
	Pot            *int             `json:"pot,omitempty"`
	HoleCards      map[int][]Card   `json:"holeCards,omitempty"`
}

// SendTableState sends a table_state message to a single client
// If the client is seated and a hand is active, includes their hole cards
// Always includes card counts for occupied seats so spectators can render card backs
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

	// Determine which seat (if any) the client is sitting at
	var clientSeatIndex *int
	session, err := server.sessionManager.GetSession(c.Token)
	if err == nil && session.SeatIndex != nil {
		clientSeatIndex = session.SeatIndex
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
			// Set stack for occupied seat
			stack := seat.Stack
			seats[i].Stack = &stack
		} else {
			seats[i].PlayerName = nil
			seats[i].Stack = nil
		}
	}

	// Get game state info when hand is active
	var dealerSeat *int
	var smallBlindSeat *int
	var bigBlindSeat *int
	var pot *int
	handInProgress := false

	if table.CurrentHand != nil {
		handInProgress = true
		dealerSeat = table.DealerSeat
		sbSeat := table.CurrentHand.SmallBlindSeat
		bbSeat := table.CurrentHand.BigBlindSeat
		potAmount := table.CurrentHand.Pot
		smallBlindSeat = &sbSeat
		bigBlindSeat = &bbSeat
		pot = &potAmount
	}

	// Populate card counts for all occupied seats during active hand
	if table.CurrentHand != nil {
		for i, seat := range table.Seats {
			if seat.Token != nil {
				// This seat has a player
				if cardList, hasCards := table.CurrentHand.HoleCards[i]; hasCards {
					cardCount := len(cardList)
					seats[i].CardCount = &cardCount
				}
			}
		}
	}

	table.mu.RUnlock()

	// Populate hole cards - only if client is seated and hand is active
	var holeCards map[int][]Card
	if clientSeatIndex != nil && table.CurrentHand != nil {
		// Client is seated - give them their own cards only (privacy)
		holeCards = make(map[int][]Card)
		table.mu.RLock()
		if clientCards, hasCards := table.CurrentHand.HoleCards[*clientSeatIndex]; hasCards {
			holeCards[*clientSeatIndex] = clientCards
		}
		table.mu.RUnlock()
	}

	// Create payload
	payloadObj := TableStatePayload{
		TableId:        tableID,
		Seats:          seats,
		HandInProgress: handInProgress,
		DealerSeat:     dealerSeat,
		SmallBlindSeat: smallBlindSeat,
		BigBlindSeat:   bigBlindSeat,
		Pot:            pot,
		HoleCards:      holeCards,
	}

	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// DEBUG: Log the payload being sent
	logger.Info("DEBUG table_state payload", "payload", string(payloadBytes))

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
// Personalizes the table_state for each client (hole cards only for their own seat, card counts for all occupied seats)
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

	// Send personalized table_state to each client
	for _, client := range clients {
		if excludeClient != nil && client == excludeClient {
			continue
		}

		// Build a personalized table_state for this client
		err := s.sendPersonalizedTableState(client, table)
		if err != nil {
			s.logger.Warn("failed to send personalized table_state", "error", err)
			continue
		}
	}

	return nil
}

// sendPersonalizedTableState sends a personalized table_state to a specific client
// The client sees their own hole cards (if seated) and card counts for all occupied seats
func (s *Server) sendPersonalizedTableState(client *Client, table *Table) error {
	// Determine which seat (if any) the client is sitting at
	var clientSeatIndex *int
	session, err := s.sessionManager.GetSession(client.Token)
	if err == nil && session.SeatIndex != nil {
		clientSeatIndex = session.SeatIndex
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
			// Set stack for occupied seat
			stack := seat.Stack
			seats[i].Stack = &stack
		} else {
			seats[i].PlayerName = nil
			seats[i].Stack = nil
		}
	}

	// Get game state info when hand is active
	var dealerSeat *int
	var smallBlindSeat *int
	var bigBlindSeat *int
	var pot *int
	handInProgress := false

	if table.CurrentHand != nil {
		handInProgress = true
		dealerSeat = table.DealerSeat
		sbSeat := table.CurrentHand.SmallBlindSeat
		bbSeat := table.CurrentHand.BigBlindSeat
		potAmount := table.CurrentHand.Pot
		smallBlindSeat = &sbSeat
		bigBlindSeat = &bbSeat
		pot = &potAmount
	}

	// Populate card counts for all occupied seats during active hand
	if table.CurrentHand != nil {
		for i, seat := range table.Seats {
			if seat.Token != nil {
				// This seat has a player
				if cardList, hasCards := table.CurrentHand.HoleCards[i]; hasCards {
					cardCount := len(cardList)
					seats[i].CardCount = &cardCount
				}
			}
		}
	}

	table.mu.RUnlock()

	// Populate hole cards - only if client is seated and hand is active
	var holeCards map[int][]Card
	if clientSeatIndex != nil && table.CurrentHand != nil {
		// Client is seated - give them their own cards only (privacy)
		holeCards = make(map[int][]Card)
		table.mu.RLock()
		if clientCards, hasCards := table.CurrentHand.HoleCards[*clientSeatIndex]; hasCards {
			holeCards[*clientSeatIndex] = clientCards
		}
		table.mu.RUnlock()
	}

	// Create payload
	payloadObj := TableStatePayload{
		TableId:        table.ID,
		Seats:          seats,
		HandInProgress: handInProgress,
		DealerSeat:     dealerSeat,
		SmallBlindSeat: smallBlindSeat,
		BigBlindSeat:   bigBlindSeat,
		Pot:            pot,
		HoleCards:      holeCards,
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

	// Send to this client (safely handle closed channel)
	defer func() {
		if r := recover(); r != nil {
			s.logger.Debug("recovered from send on closed channel", "error", r)
		}
	}()

	select {
	case client.send <- responseBytes:
	default:
		s.logger.Warn("client send channel full, skipping table_state message")
	}

	return nil
}

// broadcastHandStarted sends hand_started message to all clients at the table with dealer and blind info
func (s *Server) broadcastHandStarted(table *Table) error {
	// Get all clients at the table
	clients := s.GetClientsAtTable(table.ID)
	s.logger.Info("broadcasting hand_started", "tableID", table.ID, "num_clients", len(clients))

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

	s.logger.Info("hand_started details", "dealerSeat", dealerSeat, "sbSeat", sbSeat, "bbSeat", bbSeat)

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
	sentCount := 0
	for _, client := range clients {
		select {
		case client.send <- responseBytes:
			sentCount++
		default:
			s.logger.Warn("client send channel full, skipping hand_started message")
		}
	}

	s.logger.Info("hand_started broadcast complete", "tableID", table.ID, "sentCount", sentCount)
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

// broadcastBoardDealt sends board_dealt message to all clients at the table with community cards
// street parameter should be "flop", "turn", or "river"
func (s *Server) broadcastBoardDealt(table *Table, street string) error {
	// Get all clients at the table
	clients := s.GetClientsAtTable(table.ID)

	table.mu.RLock()
	hand := table.CurrentHand
	if hand == nil {
		table.mu.RUnlock()
		return fmt.Errorf("CurrentHand is nil")
	}
	boardCards := hand.BoardCards
	table.mu.RUnlock()

	// Create payload with board cards and street indicator
	payloadObj := BoardDealtPayload{
		BoardCards: boardCards,
		Street:     street,
	}

	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to marshal board_dealt payload: %w", err)
	}

	response := WebSocketMessage{
		Type:    "board_dealt",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal board_dealt response: %w", err)
	}

	// Send to all clients at the table
	for _, client := range clients {
		select {
		case client.send <- responseBytes:
		default:
			s.logger.Warn("client send channel full, skipping board_dealt message")
		}
	}

	return nil
}

// handRankToString converts a numeric rank to human-readable hand name
func handRankToString(rank int) string {
	switch rank {
	case 8:
		return "Straight Flush"
	case 7:
		return "Four of a Kind"
	case 6:
		return "Full House"
	case 5:
		return "Flush"
	case 4:
		return "Straight"
	case 3:
		return "Three of a Kind"
	case 2:
		return "Two Pair"
	case 1:
		return "One Pair"
	case 0:
		return "High Card"
	default:
		return "Unknown Hand"
	}
}

// broadcastShowdown sends showdown results to all players at the table
func (s *Server) broadcastShowdown(table *Table, winners []int, rank *HandRank, amountsWon map[int]int) {
	clients := s.GetClientsAtTable(table.ID)
	s.logger.Info("broadcasting showdown_result", "tableID", table.ID, "num_clients", len(clients))

	winningHandName := "Unknown Hand"
	if rank != nil {
		winningHandName = handRankToString(rank.Rank)
	}

	// Get the pot amount from the distribution
	potAmount := 0
	for _, amount := range amountsWon {
		potAmount += amount
	}

	payload := ShowdownResultPayload{
		WinnerSeats: winners,
		WinningHand: winningHandName,
		PotAmount:   potAmount,
		AmountsWon:  amountsWon,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("failed to marshal showdown result", "error", err)
		return
	}

	response := WebSocketMessage{
		Type:    "showdown_result",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		s.logger.Error("failed to marshal showdown message", "error", err)
		return
	}

	sentCount := 0
	for _, client := range clients {
		select {
		case client.send <- responseBytes:
			sentCount++
		default:
			s.logger.Warn("failed to send showdown to client - buffer full")
		}
	}

	s.logger.Info("showdown_result broadcast complete", "tableID", table.ID, "sentCount", sentCount)
}

// broadcastHandComplete sends hand completion message to all players at the table
func (s *Server) broadcastHandComplete(table *Table) {
	clients := s.GetClientsAtTable(table.ID)
	s.logger.Info("broadcasting hand_complete", "tableID", table.ID, "num_clients", len(clients))

	payload := HandCompletePayload{
		Message: "Hand complete. Click 'Start Hand' to begin next hand.",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("failed to marshal hand complete", "error", err)
		return
	}

	response := WebSocketMessage{
		Type:    "hand_complete",
		Payload: json.RawMessage(payloadBytes),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		s.logger.Error("failed to marshal hand complete message", "error", err)
		return
	}

	sentCount := 0
	for _, client := range clients {
		select {
		case client.send <- responseBytes:
			sentCount++
		default:
			s.logger.Warn("failed to send hand_complete to client - buffer full")
		}
	}

	s.logger.Info("hand_complete broadcast complete", "tableID", table.ID, "sentCount", sentCount)
}

// HandleStartHand processes a start_hand message to manually trigger hand start (temporary testing feature)
func (c *Client) HandleStartHand(sm *SessionManager, server *Server, logger *slog.Logger, payload []byte) error {
	// Verify session exists
	session, err := sm.GetSession(c.Token)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Verify player is seated at a table
	if session.TableID == nil || session.SeatIndex == nil {
		return fmt.Errorf("not_seated")
	}

	// Get the table reference
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

	// Start the hand (this will handle all broadcasting internally)
	err = table.StartHand()
	if err != nil {
		logger.Warn("failed to start hand", "error", err)
		return fmt.Errorf("failed to start hand: %w", err)
	}

	logger.Info("hand started via temporary testing button", "token", c.Token, "tableId", *session.TableID)

	return nil
}

// HandlePlayerAction processes a player action (fold, check, call, raise) during a hand
// For raise actions, amount should be provided as variadic parameter
func (server *Server) HandlePlayerAction(sm *SessionManager, client *Client, seatIndex int, action string, amount ...int) error {
	// Get the session for the client
	session, err := sm.GetSession(client.Token)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Verify player is seated at a table
	if session.TableID == nil || session.SeatIndex == nil {
		return fmt.Errorf("player not seated")
	}

	// Verify seat index matches the session
	if *session.SeatIndex != seatIndex {
		return fmt.Errorf("seat index mismatch: client at seat %d, action for seat %d", *session.SeatIndex, seatIndex)
	}

	// Get the table reference
	server.mu.RLock()
	var table *Table
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

	// Verify action is valid
	table.mu.Lock()
	defer table.mu.Unlock()

	// Check that a hand is in progress
	if table.CurrentHand == nil {
		return fmt.Errorf("no hand in progress")
	}

	// Verify it's the player's turn
	if table.CurrentHand.CurrentActor == nil || *table.CurrentHand.CurrentActor != seatIndex {
		return fmt.Errorf("not current actor: current actor is %v, player at seat %d", table.CurrentHand.CurrentActor, seatIndex)
	}

	// Get valid actions for this player
	validActions := table.CurrentHand.GetValidActions(seatIndex, table.Seats[seatIndex].Stack, table.Seats)
	isValid := false
	for _, va := range validActions {
		if va == action {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid action '%s' for seat %d: valid actions are %v", action, seatIndex, validActions)
	}

	// Process the action - pass amount if provided
	var amountActed int
	if action == "raise" {
		// Raise requires an amount
		if len(amount) == 0 {
			return fmt.Errorf("raise action requires amount parameter")
		}
		amountActed, err = table.CurrentHand.ProcessAction(seatIndex, action, table.Seats[seatIndex].Stack, amount[0])
	} else {
		// Other actions don't use amount
		amountActed, err = table.CurrentHand.ProcessAction(seatIndex, action, table.Seats[seatIndex].Stack)
	}
	if err != nil {
		return fmt.Errorf("failed to process action: %w", err)
	}

	// Update the player's stack after action (subtract chips moved)
	table.Seats[seatIndex].Stack -= amountActed
	newStack := table.Seats[seatIndex].Stack

	// Check if betting round is complete
	if table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		// Betting round is over - broadcast with no next actor
		// Temporarily unlock to broadcast
		table.mu.Unlock()
		err = server.BroadcastActionResult(
			table.ID, seatIndex, action, amountActed, newStack, table.CurrentHand.Pot,
			nil, true, nil,
		)
		table.mu.Lock()
		if err != nil {
			server.logger.Warn("failed to broadcast action_result", "error", err)
		}

		// Check if only one player remains (all others folded) - early winner
		nonFoldedCount := 0
		for i := 0; i < 6; i++ {
			if table.Seats[i].Status == "active" && !table.CurrentHand.FoldedPlayers[i] {
				nonFoldedCount++
			}
		}

		// If only one player remains, award pot immediately (early winner)
		if nonFoldedCount <= 1 {
			table.mu.Unlock()
			table.HandleShowdown()
			table.mu.Lock()
			return nil
		}

		// Multiple players remain - check if we're not on the river - advance to next street
		currentStreet := table.CurrentHand.Street
		if currentStreet != "river" {
			// Unlock before calling AdvanceToNextStreetWithBroadcast (long-running operation with broadcasting)
			table.mu.Unlock()
			err = table.AdvanceToNextStreetWithBroadcast()
			if err != nil {
				server.logger.Warn("failed to advance to next street", "error", err)
			}
			table.mu.Lock()

			// After advancing street, set first actor for the new street and request their action
			// Determine who acts first on the new street
			if table.CurrentHand != nil {
				firstActor := table.CurrentHand.GetFirstActor(table.Seats)
				table.CurrentHand.CurrentActor = &firstActor

				// Get valid actions and call amount for the first actor of the new street
				nextValidActions := table.CurrentHand.GetValidActions(firstActor, table.Seats[firstActor].Stack, table.Seats)
				nextCallAmount := table.CurrentHand.GetCallAmount(firstActor)

				// Unlock to broadcast action_request for the new street
				table.mu.Unlock()
				err = server.BroadcastActionRequest(
					table.ID, firstActor, nextValidActions, nextCallAmount,
					table.CurrentHand.CurrentBet, table.CurrentHand.Pot,
				)
				table.mu.Lock()
				if err != nil {
					server.logger.Warn("failed to broadcast action_request for new street", "error", err)
				}
			}
		} else {
			// We're on the river and betting is complete - trigger showdown
			table.mu.Unlock()
			table.HandleShowdown()
			table.mu.Lock()
		}

		return nil
	}

	// Advance to next actor
	nextActor, err := table.CurrentHand.AdvanceAction(table.Seats)
	if err != nil {
		return fmt.Errorf("failed to advance action: %w", err)
	}

	if nextActor == nil {
		// Only one player left (all others folded) - award pot immediately
		table.mu.Unlock()
		err = server.BroadcastActionResult(
			table.ID, seatIndex, action, amountActed, newStack, table.CurrentHand.Pot,
			nil, true, nil,
		)
		table.mu.Lock()
		if err != nil {
			server.logger.Warn("failed to broadcast action_result", "error", err)
		}

		// Immediately award pot to remaining player (early winner)
		// HandleShowdown already has correct logic for this (table.go lines 180-211)
		table.mu.Unlock()
		table.HandleShowdown()
		table.mu.Lock()

		return nil
	}

	// Update CurrentActor to the next player
	table.CurrentHand.CurrentActor = nextActor

	// Get valid actions and call amount for the next actor
	nextValidActions := table.CurrentHand.GetValidActions(*nextActor, table.Seats[*nextActor].Stack, table.Seats)
	nextCallAmount := table.CurrentHand.GetCallAmount(*nextActor)

	// Broadcast the action result with the next actor
	// Temporarily unlock to broadcast
	table.mu.Unlock()
	err = server.BroadcastActionResult(
		table.ID, seatIndex, action, amountActed, newStack, table.CurrentHand.Pot,
		nextActor, false, nil,
	)
	if err != nil {
		server.logger.Warn("failed to broadcast action_result", "error", err)
	}

	// Send action_request to the next actor with updated call amount
	err = server.BroadcastActionRequest(
		table.ID, *nextActor, nextValidActions, nextCallAmount,
		table.CurrentHand.CurrentBet, table.CurrentHand.Pot,
	)
	table.mu.Lock()
	if err != nil {
		server.logger.Warn("failed to broadcast action_request for next actor", "error", err)
	}

	return nil
}

// HandlePlayerActionMessage processes a player_action message from the WebSocket
// This is the entry point that extracts the action from the payload and calls HandlePlayerAction
func (c *Client) HandlePlayerActionMessage(sm *SessionManager, server *Server, logger *slog.Logger, payload []byte) error {
	var actionPayload PlayerActionPayload
	err := json.Unmarshal(payload, &actionPayload)
	if err != nil {
		return fmt.Errorf("invalid player_action payload: %w", err)
	}

	// Call the main handler, passing amount if present
	if actionPayload.Amount != nil {
		err = server.HandlePlayerAction(sm, c, actionPayload.SeatIndex, actionPayload.Action, *actionPayload.Amount)
	} else {
		err = server.HandlePlayerAction(sm, c, actionPayload.SeatIndex, actionPayload.Action)
	}
	if err != nil {
		return err
	}

	return nil
}
