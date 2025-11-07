package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

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
func (c *Client) HandleSetName(sm *SessionManager, logger *slog.Logger, payload []byte) error {
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
