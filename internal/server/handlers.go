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
