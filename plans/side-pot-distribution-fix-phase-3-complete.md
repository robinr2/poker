## Phase 3 Complete: Side Pot Calculation Algorithm

Implemented the core side pot calculation algorithm that converts contribution data into proper side pots using a layer-based approach. The algorithm correctly handles 2-6 player scenarios, multiple all-ins at different levels, folded players, and edge cases.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- `SidePot` struct (new) - holds pot amount and eligible seats
- `CalculateSidePots(contributions map[int]int, foldedPlayers map[int]bool) []SidePot` (new) - converts contributions to side pots

**Tests created/changed:**
- `TestCalculateSidePots2PlayerEqualStacks` (new)
- `TestCalculateSidePots2PlayerOneShortStack` (new)
- `TestCalculateSidePots3PlayerOneShortStack` (new)
- `TestCalculateSidePots3PlayerTwoShortStacks` (new)
- `TestCalculateSidePots3PlayerOnePlayerFolds` (new)
- `TestCalculateSidePots4PlayerMultipleAllIns` (new)
- `TestCalculateSidePots4PlayerWithFolds` (new)
- `TestCalculateSidePots4PlayerComplexScenario` (new)
- `TestCalculateSidePots5PlayerLadderAllIns` (new)
- `TestCalculateSidePots5PlayerWithMultipleFolds` (new)
- `TestCalculateSidePots6PlayerComplexMultiWay` (new)
- `TestCalculateSidePots6PlayerWhaleScenario` (new)
- `TestCalculateSidePotsAllEqualContributions` (new)
- `TestCalculateSidePotsAllFoldedExceptOne` (new)
- `TestCalculateSidePotsZeroContributions` (new)

**Review Status:** APPROVED

**Git Commit Message:**
feat: implement side pot calculation algorithm

- Add SidePot struct with Amount and EligibleSeats fields
- Implement CalculateSidePots function using layer-based pot building
- Handle 2-6 player scenarios with multiple all-ins at different levels
- Properly exclude folded players from eligibility while counting contributions
- Add 15 comprehensive tests covering realistic scenarios and edge cases
