## Plan: Hand Start & Blinds

This plan implements the core game state for starting poker hands: tracking player chip stacks, rotating the dealer button, posting blinds (SB=10, BB=20), shuffling and dealing hole cards to active players, and broadcasting game state to clients while maintaining card privacy.

**Phases (6 phases)**

1. **Phase 1: Game State Structures**
    - **Objective:** Add Card type, Hand struct, and chip stacks to enable game state tracking
    - **Files/Functions to Modify/Create:**
        - `internal/server/table.go`: Add `Card` type, `Hand` struct, add `Stack int` field to `Seat`
        - `internal/server/table_test.go`: Test structures
    - **Tests to Write:**
        - `TestCardString` - verify card representation (e.g., "As", "Kh")
        - `TestNewDeck` - verify 52-card deck generation
        - `TestHandInitialization` - verify Hand struct fields
        - `TestSeatWithStack` - verify stack field on Seat
    - **Steps:**
        1. Write tests for Card type (rank, suit, String method)
        2. Implement Card type with 2-char string representation
        3. Write tests for deck generation (52 unique cards)
        4. Implement NewDeck() function
        5. Write tests for Hand struct initialization
        6. Add Hand struct with: DealerSeat, SmallBlindSeat, BigBlindSeat, Pot, Deck, HoleCards (map[int][]Card)
        7. Write tests for Seat with Stack field
        8. Add Stack int field to Seat struct (default 1000)
        9. Run all tests to confirm green

2. **Phase 2: Dealer Button & Blind Position Logic**
    - **Objective:** Implement dealer rotation and blind position calculation with heads-up rules
    - **Files/Functions to Modify/Create:**
        - `internal/server/table.go`: Add `NextDealer()`, `GetBlindPositions()`, add `DealerSeat *int` field to `Table`
        - `internal/server/table_test.go`: Test dealer rotation and blind logic
    - **Tests to Write:**
        - `TestNextDealerFirstHand` - first hand assigns seat 0
        - `TestNextDealerRotation` - rotates clockwise through active players
        - `TestNextDealerSkipsWaiting` - skips "waiting" status seats
        - `TestGetBlindPositionsNormal` - 3+ players: SB after dealer, BB after SB
        - `TestGetBlindPositionsHeadsUp` - 2 players: dealer IS small blind
        - `TestGetBlindPositionsInsufficientPlayers` - <2 active returns error
    - **Steps:**
        1. Write tests for NextDealer() with various table states
        2. Implement NextDealer() to rotate dealer button clockwise through "active" seats
        3. Write tests for GetBlindPositions() covering normal and heads-up cases
        4. Implement GetBlindPositions() with heads-up exception (button=SB)
        5. Run tests to confirm green

3. **Phase 3: Deck Shuffle & Hole Card Dealing**
    - **Objective:** Shuffle deck and deal 2 cards to each active player
    - **Files/Functions to Modify/Create:**
        - `internal/server/table.go`: Add `ShuffleDeck()`, `DealHoleCards()`
        - `internal/server/table_test.go`: Test shuffle and dealing
    - **Tests to Write:**
        - `TestShuffleDeck` - verify deck remains 52 cards after shuffle
        - `TestDealHoleCardsToActivePlayers` - only "active" seats get 2 cards
        - `TestDealHoleCardsSkipsWaiting` - "waiting" seats get no cards
        - `TestDealHoleCardsReducesDeck` - deck size decreases correctly
    - **Steps:**
        1. Write tests for ShuffleDeck() using crypto/rand
        2. Implement ShuffleDeck() with Fisher-Yates shuffle
        3. Write tests for DealHoleCards() dealing to active players only
        4. Implement DealHoleCards() to deal 2 cards per active seat
        5. Run tests to confirm green

4. **Phase 4: Hand Start Orchestration**
    - **Objective:** Coordinate hand start: assign dealer, post blinds, deal cards, initialize pot
    - **Files/Functions to Modify/Create:**
        - `internal/server/table.go`: Add `StartHand()`, `CanStartHand()`
        - `internal/server/table_test.go`: Test hand start logic
    - **Tests to Write:**
        - `TestCanStartHandRequiresTwoActive` - false if <2 active players
        - `TestCanStartHandRequiresNoActiveHand` - false if hand already running
        - `TestStartHandInitializesDealer` - sets dealer position
        - `TestStartHandPostsBlinds` - deducts SB(10) and BB(20) from stacks
        - `TestStartHandDealsCards` - each active player has 2 cards
        - `TestStartHandSetsPot` - pot = SB + BB = 30
    - **Steps:**
        1. Write tests for CanStartHand() validation logic
        2. Implement CanStartHand() checking for ≥2 active and no active hand
        3. Write tests for StartHand() covering full hand initialization
        4. Implement StartHand(): call NextDealer, GetBlindPositions, post blinds, ShuffleDeck, DealHoleCards
        5. Run tests to confirm green

5. **Phase 5: WebSocket Protocol Extension**
    - **Objective:** Broadcast hand start events with card privacy (only own hole cards visible)
    - **Files/Functions to Modify/Create:**
        - `internal/server/handlers.go`: Add message types `hand_started`, `cards_dealt`, `blind_posted`
        - `internal/server/websocket.go`: Modify broadcast to handle per-player card filtering
        - `internal/server/handlers_test.go`: Test new message handlers
        - `internal/server/websocket_test.go`: Test card privacy in broadcasts
    - **Tests to Write:**
        - `TestBroadcastHandStarted` - all players receive dealer/blind positions
        - `TestBroadcastCardsDealtPrivacy` - each player only sees own hole cards
        - `TestBroadcastBlindPosted` - all players see blind deductions and pot
        - `TestBroadcastGameState` - stacks, pot, dealer button serialized correctly
    - **Steps:**
        1. Write tests for hand_started message broadcast
        2. Implement hand_started message with dealer, SB, BB seats and stacks
        3. Write tests for cards_dealt with privacy (own cards only)
        4. Implement per-player filtering in broadcast for hole cards
        5. Write tests for blind_posted message
        6. Implement blind_posted broadcast with updated stacks and pot
        7. Run all backend tests to confirm green

6. **Phase 6: Frontend Game Display**
    - **Objective:** Display dealer button, blind indicators, hole cards, chip stacks, and pot
    - **Files/Functions to Modify/Create:**
        - `frontend/src/components/TableView.tsx`: Add game state display elements
        - `frontend/src/components/TableView.test.tsx`: Test game element rendering
        - `frontend/src/styles/TableView.css`: Style dealer button, blinds, cards, stacks
        - `frontend/src/hooks/useWebSocket.ts`: Parse `hand_started`, `cards_dealt`, `blind_posted` messages
    - **Tests to Write:**
        - `TestTableViewDisplaysDealerButton` - dealer marker shown on correct seat
        - `TestTableViewDisplaysBlindIndicators` - SB/BB labels on correct seats
        - `TestTableViewDisplaysHoleCards` - player's own cards shown
        - `TestTableViewDisplaysStacks` - chip counts shown per seat
        - `TestTableViewDisplaysPot` - pot total shown in center
        - `TestUseWebSocketParsesHandStarted` - hook updates state on hand_started
    - **Steps:**
        1. Write tests for dealer button rendering in TableView
        2. Add dealer button indicator (e.g., "D" marker) to seat display
        3. Write tests for blind indicators (SB/BB labels)
        4. Add blind labels to appropriate seats
        5. Write tests for hole card rendering (2 cards for player's seat)
        6. Add hole card display for player's own seat
        7. Write tests for stack display per seat
        8. Add stack count display to each seat
        9. Write tests for pot display
        10. Add pot total display in table center
        11. Write tests for useWebSocket parsing new message types
        12. Implement message parsing in useWebSocket hook
        13. Run all frontend tests to confirm green
        14. Run all backend tests to confirm no regressions

**Implementation Decisions (Approved)**
1. **Hand Start Trigger:** Automatic when ≥2 active players (no manual trigger)
2. **Short Stacks:** Allow all-in for any amount if stack < BB (full handling in future betting phase)
3. **Opponent Cards:** Show face-down card backs for visual clarity
4. **First Dealer:** Start at seat 0 on first hand, then rotate clockwise
5. **Manual Testing:** No "Start Hand" button needed at this stage
