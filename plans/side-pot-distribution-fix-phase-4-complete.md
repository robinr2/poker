## Phase 4 Complete: Rewrite DistributePot() to Use Side Pot Calculation

Rewrote the `DistributePot()` function to use the `CalculateSidePots()` algorithm, enabling proper side pot distribution with multiple winners per pot. The implementation correctly handles short stacks, big stacks winning multiple pots, odd chip distribution, and edge cases.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- `DistributePot(winners []int) map[int]int` (rewritten) - now uses CalculateSidePots for proper side pot distribution

**Tests created/changed:**
- `TestDistributePot_Phase4_SingleWinner_TakesAll` (new)
- `TestDistributePot_Phase4_TwoWinners_EqualStacks_SplitPot` (new)
- `TestDistributePot_Phase4_ThreeWinners_EqualStacks_SplitThreeWay` (new)
- `TestDistributePot_Phase4_OneShortStack_WinnerIsShortStack` (new)
- `TestDistributePot_Phase4_OneShortStack_WinnerIsBigStack` (new)
- `TestDistributePot_Phase4_TwoShortStacks_MultipleWinners` (new)
- `TestDistributePot_Phase4_ComplexSidePots_SingleWinner` (new)
- `TestDistributePot_Phase4_OddChipDistribution` (new)
- `TestDistributePot_Phase4_OddChip_RemainderGoesToFirst` (new)
- `TestDistributePot_Phase4_ZeroPot` (new)
- `TestDistributePot_Phase4_AllInScenario_MultipleWinners` (new)
- `TestDistributePot_Phase4_NoWinnersForSidePot` (new)
- `TestDistributePot_Phase4_PartialRefund` (new)

**Review Status:** APPROVED

**Git Commit Message:**
feat: rewrite DistributePot to use side pot calculation

- Rewrite DistributePot to use CalculateSidePots algorithm
- Filter winners by pot eligibility (short stacks only win main pot)
- Handle odd chip distribution (remainder to first eligible winner)
- Support multiple winners per pot with correct chip splits
- Add 13 comprehensive tests covering side pot scenarios and edge cases
