package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/robinr2/poker/internal/server"
)

func main() {
	// Create a logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create session manager
	sm := server.NewSessionManager(logger)

	fmt.Println("=== Session Manager Manual Test ===\n")

	// Test 1: Create valid session
	fmt.Println("1. Creating session for 'Alice'...")
	session1, err := sm.CreateSession("Alice")
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Session created!\n")
		fmt.Printf("   Token: %s\n", session1.Token)
		fmt.Printf("   Name: %s\n", session1.Name)
		fmt.Printf("   Created: %s\n\n", session1.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Test 2: Create another session
	fmt.Println("2. Creating session for 'Bob Smith'...")
	session2, err := sm.CreateSession("Bob Smith")
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Session created!\n")
		fmt.Printf("   Token: %s\n", session2.Token)
		fmt.Printf("   Name: %s\n\n", session2.Name)
	}

	// Test 3: Retrieve session by token
	fmt.Println("3. Retrieving Alice's session by token...")
	retrieved, err := sm.GetSession(session1.Token)
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Session retrieved!\n")
		fmt.Printf("   Name: %s\n", retrieved.Name)
		fmt.Printf("   TableID: %v\n", retrieved.TableID)
		fmt.Printf("   SeatIndex: %v\n\n", retrieved.SeatIndex)
	}

	// Test 4: Update session with table and seat
	fmt.Println("4. Updating Alice's session (Table 1, Seat 3)...")
	tableID := "table-1"
	seatIndex := 3
	updated, err := sm.UpdateSession(session1.Token, &tableID, &seatIndex)
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Session updated!\n")
		fmt.Printf("   TableID: %s\n", *updated.TableID)
		fmt.Printf("   SeatIndex: %d\n\n", *updated.SeatIndex)
	}

	// Test 5: Try invalid name
	fmt.Println("5. Testing invalid name 'Alice@Bob'...")
	_, err = sm.CreateSession("Alice@Bob")
	if err != nil {
		fmt.Printf("   ✅ Correctly rejected: %v\n\n", err)
	} else {
		fmt.Printf("   ❌ Should have been rejected!\n\n")
	}

	// Test 6: Try empty name
	fmt.Println("6. Testing empty name...")
	_, err = sm.CreateSession("")
	if err != nil {
		fmt.Printf("   ✅ Correctly rejected: %v\n\n", err)
	} else {
		fmt.Printf("   ❌ Should have been rejected!\n\n")
	}

	// Test 7: Try too long name
	fmt.Println("7. Testing name that's too long (21 characters)...")
	_, err = sm.CreateSession("123456789012345678901")
	if err != nil {
		fmt.Printf("   ✅ Correctly rejected: %v\n\n", err)
	} else {
		fmt.Printf("   ❌ Should have been rejected!\n\n")
	}

	// Test 8: Try whitespace trimming
	fmt.Println("8. Testing name with whitespace '  Charlie  '...")
	session3, err := sm.CreateSession("  Charlie  ")
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Session created with trimmed name: '%s'\n\n", session3.Name)
	}

	// Test 9: Remove session
	fmt.Println("9. Removing Bob's session...")
	err = sm.RemoveSession(session2.Token)
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Session removed!\n")
		// Try to retrieve removed session
		_, err = sm.GetSession(session2.Token)
		if err != nil {
			fmt.Printf("   ✅ Confirmed: Session no longer exists\n\n")
		} else {
			fmt.Printf("   ❌ Session still exists!\n\n")
		}
	}

	// Test 10: Try to get non-existent session
	fmt.Println("10. Testing retrieval of non-existent token...")
	_, err = sm.GetSession("fake-token-12345")
	if err != nil {
		fmt.Printf("   ✅ Correctly returned error: %v\n\n", err)
	} else {
		fmt.Printf("   ❌ Should have returned error!\n\n")
	}

	fmt.Println("=== All Manual Tests Complete ===")
}
