package server

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestSessionManager_CreateSession tests that sessions are created with valid names
// and UUID tokens are generated
func TestSessionManager_CreateSession(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	tests := []struct {
		name      string
		inputName string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid single word",
			inputName: "Alice",
			wantErr:   false,
		},
		{
			name:      "valid multiple words with space",
			inputName: "Alice Cooper",
			wantErr:   false,
		},
		{
			name:      "valid with dashes",
			inputName: "Alice-Cooper",
			wantErr:   false,
		},
		{
			name:      "valid with underscores",
			inputName: "Alice_Cooper",
			wantErr:   false,
		},
		{
			name:      "valid with numbers",
			inputName: "Alice2024",
			wantErr:   false,
		},
		{
			name:      "valid mixed",
			inputName: "Alice_2024-A",
			wantErr:   false,
		},
		{
			name:      "valid single character",
			inputName: "A",
			wantErr:   false,
		},
		{
			name:      "valid 20 characters",
			inputName: "12345678901234567890",
			wantErr:   false,
		},
		{
			name:      "valid with leading/trailing spaces (should be trimmed)",
			inputName: "  Alice  ",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := sm.CreateSession(tt.inputName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if session == nil {
					t.Fatal("expected session, got nil")
				}
				if session.Token == "" {
					t.Fatal("expected token to be set")
				}
				if session.Name == "" {
					t.Fatal("expected name to be set")
				}
				if session.TableID != nil {
					t.Fatal("expected TableID to be nil for new session")
				}
				if session.SeatIndex != nil {
					t.Fatal("expected SeatIndex to be nil for new session")
				}
				if session.CreatedAt.IsZero() {
					t.Fatal("expected CreatedAt to be set")
				}
			}
		})
	}
}

// TestSessionManager_InvalidNames tests that invalid names are rejected
func TestSessionManager_InvalidNames(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	tests := []struct {
		name      string
		inputName string
		wantErr   bool
	}{
		{
			name:      "empty string",
			inputName: "",
			wantErr:   true,
		},
		{
			name:      "only whitespace",
			inputName: "   ",
			wantErr:   true,
		},
		{
			name:      "21 characters (too long)",
			inputName: "123456789012345678901",
			wantErr:   true,
		},
		{
			name:      "invalid special characters",
			inputName: "Alice@Bob",
			wantErr:   true,
		},
		{
			name:      "invalid special characters (comma)",
			inputName: "Alice,Bob",
			wantErr:   true,
		},
		{
			name:      "invalid special characters (exclamation)",
			inputName: "Alice!",
			wantErr:   true,
		},
		{
			name:      "invalid special characters (parenthesis)",
			inputName: "Alice(Bob)",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := sm.CreateSession(tt.inputName)
			if !tt.wantErr && err != nil {
				t.Errorf("CreateSession() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err == nil {
				t.Error("CreateSession() expected error, got nil")
			}
			if tt.wantErr && session != nil {
				t.Error("CreateSession() expected nil session on error")
			}
		})
	}
}

// TestSessionManager_GetSession tests session retrieval by token
func TestSessionManager_GetSession(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	// Create a session
	session, err := sm.CreateSession("Alice")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Get the session by token
	retrieved, err := sm.GetSession(session.Token)
	if err != nil {
		t.Errorf("GetSession failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected session, got nil")
	}
	if retrieved.Token != session.Token {
		t.Errorf("Token mismatch: got %s, want %s", retrieved.Token, session.Token)
	}
	if retrieved.Name != session.Name {
		t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, session.Name)
	}
}

// TestSessionManager_GetSession_NotFound tests that non-existent tokens return error
func TestSessionManager_GetSession_NotFound(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	_, err := sm.GetSession("nonexistent-token")
	if err == nil {
		t.Fatal("expected error for non-existent token, got nil")
	}
}

// TestSessionManager_UpdateSession tests updating session table and seat
func TestSessionManager_UpdateSession(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	// Create a session
	session, err := sm.CreateSession("Alice")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Update session with table and seat
	tableID := "table-123"
	seatIndex := 2
	updatedSession, err := sm.UpdateSession(session.Token, &tableID, &seatIndex)
	if err != nil {
		t.Errorf("UpdateSession failed: %v", err)
	}

	if updatedSession.TableID == nil || *updatedSession.TableID != tableID {
		t.Errorf("TableID mismatch: got %v, want %s", updatedSession.TableID, tableID)
	}
	if updatedSession.SeatIndex == nil || *updatedSession.SeatIndex != seatIndex {
		t.Errorf("SeatIndex mismatch: got %v, want %d", updatedSession.SeatIndex, seatIndex)
	}

	// Verify update persisted
	retrieved, err := sm.GetSession(session.Token)
	if err != nil {
		t.Errorf("GetSession failed: %v", err)
	}
	if retrieved.TableID == nil || *retrieved.TableID != tableID {
		t.Errorf("TableID mismatch after retrieval: got %v, want %s", retrieved.TableID, tableID)
	}
}

// TestSessionManager_UpdateSession_ClearValues tests clearing table and seat
func TestSessionManager_UpdateSession_ClearValues(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	// Create a session
	session, err := sm.CreateSession("Alice")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Update with table and seat
	tableID := "table-123"
	seatIndex := 2
	_, err = sm.UpdateSession(session.Token, &tableID, &seatIndex)
	if err != nil {
		t.Errorf("UpdateSession failed: %v", err)
	}

	// Clear values by passing nil
	updatedSession, err := sm.UpdateSession(session.Token, nil, nil)
	if err != nil {
		t.Errorf("UpdateSession failed: %v", err)
	}

	if updatedSession.TableID != nil {
		t.Errorf("TableID should be nil, got %v", updatedSession.TableID)
	}
	if updatedSession.SeatIndex != nil {
		t.Errorf("SeatIndex should be nil, got %v", updatedSession.SeatIndex)
	}
}

// TestSessionManager_RemoveSession tests session removal
func TestSessionManager_RemoveSession(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	// Create a session
	session, err := sm.CreateSession("Alice")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Remove session
	err = sm.RemoveSession(session.Token)
	if err != nil {
		t.Errorf("RemoveSession failed: %v", err)
	}

	// Verify it's gone
	_, err = sm.GetSession(session.Token)
	if err == nil {
		t.Fatal("expected error after removal, got nil")
	}
}

// TestSessionManager_RemoveSession_NotFound tests removing non-existent session
func TestSessionManager_RemoveSession_NotFound(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	err := sm.RemoveSession("nonexistent-token")
	if err == nil {
		t.Fatal("expected error for non-existent token, got nil")
	}
}

// TestSessionManager_ConcurrentAccess tests thread safety with concurrent operations
func TestSessionManager_ConcurrentAccess(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	numGoroutines := 100
	operationsPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Track successful operations
	var successfulCreates atomic.Int32
	var successfulGets atomic.Int32
	var successfulUpdates atomic.Int32
	var successfulRemoves atomic.Int32

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Create session
				sessionName := fmt.Sprintf("User_%d_%d", id, j)
				session, err := sm.CreateSession(sessionName)
				if err == nil && session != nil {
					successfulCreates.Add(1)

					// Get session
					if retrieved, err := sm.GetSession(session.Token); err == nil && retrieved != nil {
						successfulGets.Add(1)

						// Update session
						tableID := fmt.Sprintf("table_%d", id)
						seatIndex := j % 8
						if updated, err := sm.UpdateSession(session.Token, &tableID, &seatIndex); err == nil && updated != nil {
							successfulUpdates.Add(1)
						}
					}

					// Remove session
					if err := sm.RemoveSession(session.Token); err == nil {
						successfulRemoves.Add(1)
					}
				}
			}
		}(i)
	}

	wg.Wait()

	expectedOps := int32(numGoroutines * operationsPerGoroutine)
	if successfulCreates.Load() != expectedOps {
		t.Errorf("CreateSession operations: got %d, want %d", successfulCreates.Load(), expectedOps)
	}
	if successfulGets.Load() != expectedOps {
		t.Errorf("GetSession operations: got %d, want %d", successfulGets.Load(), expectedOps)
	}
	if successfulUpdates.Load() != expectedOps {
		t.Errorf("UpdateSession operations: got %d, want %d", successfulUpdates.Load(), expectedOps)
	}
	if successfulRemoves.Load() != expectedOps {
		t.Errorf("RemoveSession operations: got %d, want %d", successfulRemoves.Load(), expectedOps)
	}
}

// TestSessionManager_TokenUniqueness tests that each session gets a unique token
func TestSessionManager_TokenUniqueness(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	numSessions := 100
	tokens := make(map[string]bool)

	for i := 0; i < numSessions; i++ {
		sessionName := fmt.Sprintf("User_%d", i)
		session, err := sm.CreateSession(sessionName)
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		if tokens[session.Token] {
			t.Errorf("Duplicate token generated: %s", session.Token)
		}
		tokens[session.Token] = true
	}

	if len(tokens) != numSessions {
		t.Errorf("Expected %d unique tokens, got %d", numSessions, len(tokens))
	}
}

// TestSessionManager_CreatedAtTimestamp tests that CreatedAt is properly set
func TestSessionManager_CreatedAtTimestamp(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	before := time.Now()
	session, err := sm.CreateSession("Alice")
	after := time.Now()

	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if session.CreatedAt.Before(before) {
		t.Error("CreatedAt is before the test start time")
	}
	if session.CreatedAt.After(after) {
		t.Error("CreatedAt is after the test end time")
	}
}

// TestSessionManager_NameTrimming tests that names are trimmed before validation
func TestSessionManager_NameTrimming(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	session, err := sm.CreateSession("  Alice  ")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if session.Name != "Alice" {
		t.Errorf("Name should be trimmed, got %s", session.Name)
	}
}
