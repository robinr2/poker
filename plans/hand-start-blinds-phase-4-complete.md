## Phase 4 Complete: Hand Start Orchestration

Successfully implemented the StartHand() orchestration function that coordinates all hand initialization elements including dealer assignment, blind posting with all-in handling, deck shuffling, and hole card dealing.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- `Table.CurrentHand` (field added)
- `Table.CanStartHand()` - validates prerequisites for starting a hand
- `Table.StartHand()` - orchestrates complete hand initialization sequence
- `Table.assignDealerLocked()` - helper for dealer assignment
- `Table.getBlindPositionsLocked()` - helper for blind position calculation

**Tests created/changed:**
- `TestTableCanStartHand` - validates starting conditions
- `TestTableStartHandNoPlayers` - error case: insufficient players
- `TestTableStartHandActiveHand` - error case: hand already in progress
- `TestTableStartHandSuccess` - happy path with 3 players
- `TestTableStartHandDealerInitialization` - first hand dealer assignment
- `TestTableStartHandDealerRotation` - dealer rotates across multiple hands
- `TestTableStartHandBlindPosting` - verifies blind amounts and pot
- `TestTableStartHandAllInBlinds` - handles short stacks posting blinds

**Review Status:** APPROVED

**Git Commit Message:**
feat: Add hand start orchestration with blind posting

- Implement StartHand() to coordinate full hand initialization sequence
- Add CurrentHand field to Table for active hand state tracking
- Implement CanStartHand() validation for starting prerequisites
- Post blinds (SB=10, BB=20) with automatic all-in handling for short stacks
- Integrate dealer rotation, blind positions, deck shuffle, and card dealing
- Add 8 comprehensive tests covering validation, rotation, and edge cases
