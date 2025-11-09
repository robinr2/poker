## Phase 6 Complete: Raise UI Components

Successfully implemented raise UI components with input field and preset buttons. The implementation adds a raise amount input field (hidden when raise is unavailable), preset buttons for Min/Pot/All-in raises, and a Raise button with disabled state when amount is invalid. All components work with the backend raise bounds and provide an intuitive UX for players to select raise amounts.

**Files created/changed:**
- frontend/src/components/TableView.tsx
- frontend/src/components/TableView.test.tsx
- frontend/src/styles/TableView.css

**Components created/changed:**
- `TableView` component (added raise UI controls)
- `GameState` interface (confirmed minRaise and maxRaise fields)
- `handleAction()` function (updated to support optional amount parameter for raises)
- Raise preset handlers: `handleMinRaise()`, `handlePotRaise()`, `handleAllIn()`
- Raise validation logic: checks amount is within minRaise/maxRaise bounds

**UI Elements added:**
- `.raise-controls` container for raise inputs
- `.raise-input` number input field for custom amounts
- `.preset-button` buttons for Min, Pot, and All-in presets
- `.raise-button` main raise submission button with disabled state
- Validation: button disabled when amount < minRaise or amount > maxRaise or input empty

**Tests created/changed:**
- `TestTableView_RaiseButtonVisibility` (new)
- `TestTableView_RaiseButtonHiddenWhenUnavailable` (new)
- `TestTableView_MinPresetSetsAmount` (new)
- `TestTableView_PotPresetCalculatesAmount` (new)
- `TestTableView_AllInPresetSetsAmount` (new)
- `TestTableView_RaiseInputVisible` (new)
- `TestTableView_RaiseSendsActionWithAmount` (new)
- `TestTableView_RaiseDisabledBelowMin` (new)
- `TestTableView_RaiseDisabledAboveMax` (new)
- `TestTableView_RaiseEnabledWhenValid` (new)

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: Implement raise UI components with presets and validation

- Add raise amount input field visible only when raise is available
- Implement Min, Pot, and All-in preset buttons for quick raise selection
- Add Raise button with validation: disabled when amount invalid
- Integrate with backend raise bounds (minRaise, maxRaise)
- Update handleAction() to support optional amount parameter
- Add styling for raise controls with responsive design
- Add 10 comprehensive tests covering UI visibility, presets, and validation
```
