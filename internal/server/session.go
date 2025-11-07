package server

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session represents a player session with identity and optional table placement
type Session struct {
	Token     string
	Name      string
	TableID   *string
	SeatIndex *int
	CreatedAt time.Time
}

// SessionManager manages player sessions with thread-safe operations
type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
	logger   *slog.Logger
}

// NewSessionManager creates and returns a new SessionManager instance
func NewSessionManager(logger *slog.Logger) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		logger:   logger,
	}
}

// nameValidationRegex matches names with 1-20 alphanumeric characters, spaces, dashes, or underscores
var nameValidationRegex = regexp.MustCompile(`^[a-zA-Z0-9 _-]{1,20}$`)

// validateName checks if a name meets requirements
func validateName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(trimmed) > 20 {
		return fmt.Errorf("name cannot exceed 20 characters")
	}
	if !nameValidationRegex.MatchString(trimmed) {
		return fmt.Errorf("name can only contain alphanumeric characters, spaces, dashes, and underscores")
	}
	return nil
}

// CreateSession creates a new session with a UUID token and validates the name
func (sm *SessionManager) CreateSession(name string) (*Session, error) {
	// Validate the name
	if err := validateName(name); err != nil {
		sm.logger.Warn("invalid name for session creation", "name", name, "error", err)
		return nil, err
	}

	// Trim whitespace from name
	trimmedName := strings.TrimSpace(name)

	// Generate UUID token
	token := uuid.New().String()

	// Create session
	session := &Session{
		Token:     token,
		Name:      trimmedName,
		TableID:   nil,
		SeatIndex: nil,
		CreatedAt: time.Now(),
	}

	// Store session in map (thread-safe)
	sm.mutex.Lock()
	sm.sessions[token] = session
	sm.mutex.Unlock()

	sm.logger.Info("session created", "token", token, "name", trimmedName)
	return session, nil
}

// GetSession retrieves a session by token
func (sm *SessionManager) GetSession(token string) (*Session, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, ok := sm.sessions[token]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", token)
	}

	return session, nil
}

// UpdateSession updates a session's table and seat information
func (sm *SessionManager) UpdateSession(token string, tableID *string, seatIndex *int) (*Session, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, ok := sm.sessions[token]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", token)
	}

	session.TableID = tableID
	session.SeatIndex = seatIndex

	sm.logger.Info("session updated", "token", token, "tableID", tableID, "seatIndex", seatIndex)
	return session, nil
}

// RemoveSession removes a session by token
func (sm *SessionManager) RemoveSession(token string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	_, ok := sm.sessions[token]
	if !ok {
		return fmt.Errorf("session not found: %s", token)
	}

	delete(sm.sessions, token)
	sm.logger.Info("session removed", "token", token)
	return nil
}
