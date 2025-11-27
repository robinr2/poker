---
description: Manually explore the poker game with 3 browser instances and document findings
---

You are a professional poker player QA tester. Your job is to manually explore the poker application using 3 browser instances and document any issues you find.

## Step 1: Start the Server

First, check if the server is running. If not, start it:
```bash
POKER_TEST_MODE=true go run ./cmd/server
```
The server runs on http://localhost:8080

## Step 2: Create 3 Browser Instances

Use the concurrent-browser MCP tools to create 3 browser instances:

For each browser (Alice, Bob, Charlie), call `concurrent-browser_browser_create_instance` with:
- browserType: "chromium"
- headless: false (so we can see the browsers)
- viewport: { width: 950, height: 1200 }
- metadata: { name: "Alice" } (or "Bob", "Charlie")

Save each instanceId for later use.

## Step 3: Navigate and Join Table

For each browser instance:

1. **Navigate** using `concurrent-browser_browser_navigate`:
   - url: "http://localhost:8080"
   - waitUntil: "load"

2. **Enter player name** using `concurrent-browser_browser_fill`:
   - selector: 'input[type="text"]'
   - value: "Alice" (or Bob, Charlie)

3. **Submit name** using `concurrent-browser_browser_click`:
   - selector: 'button[type="submit"]'

4. **Join table** using `concurrent-browser_browser_click`:
   - selector: 'button.join-button'

5. **Take a screenshot** using `concurrent-browser_browser_screenshot` to verify state

## Step 4: Start a Hand

Once all 3 players have joined:

1. **Start the hand** with one player using `concurrent-browser_browser_click`:
   - selector: 'button:has-text("Start Hand")'

2. **Take screenshots** of all browsers to see the dealt cards

## Step 5: Play the Game

The action bar appears when it's a player's turn. Use these selectors:

- **Check whose turn**: Look for `.action-bar` visibility
- **Fold**: `button:has-text("Fold")`
- **Call**: `button:has-text("Call")`
- **Check**: `button:has-text("Check")`
- **Min raise**: Click `button.preset-button:has-text("Min")` then `button.raise-button:has-text("Raise")`
- **Pot raise**: Click `button.preset-button:has-text("Pot")` then `button.raise-button:has-text("Raise")`
- **All-in**: Click `button.preset-button:has-text("All-in")` then `button.raise-button:has-text("Raise")`

After each action:
1. Take screenshots of all browsers using `concurrent-browser_browser_screenshot`
2. Get page content using `concurrent-browser_browser_get_markdown` to read the game state
3. Note any issues you observe

## What to Look For

After EACH action, observe the UI state and check for:
1. **Pot amount** - Is it correctly updated?
2. **Player stacks** - Do chip counts make sense?
3. **Turn indicator** - Is the correct player's turn shown?
4. **Cards displayed** - Are hole cards and board cards correct?
5. **Action buttons** - Do they work? Are they enabled/disabled correctly?
6. **Street progression** - Preflop -> Flop -> Turn -> River -> Showdown
7. **Showdown results** - Is the winner determined correctly?
8. **Hand descriptions** - Are they accurate? (e.g., "Pair of Aces" vs "Unknown Hand")

## Document Your Findings

Read @e2e/manual-exploration-notes.md to see the existing format, then append your findings.

### Issues Table Format (add to existing table)
```
| # | Description | Severity | Hand # | Notes |
|---|-------------|----------|--------|-------|
| X | Brief description | Low/Medium/High/CRITICAL | Hand X | Details |
```

### Game Log Format (append new hands)
```
### Hand X (Session Name)

**Preflop - Initial Deal**
| Player | Position | Cards | Stack | Street Bet |
|--------|----------|-------|-------|------------|
| Name | D/SB/BB | X suit Y suit | XXXX | XX |

Pot: XX | Street: Preflop | Action on: PlayerName

**Action X: Player does action**
State after: ...
```

## Professional Poker Player Perspective

When evaluating issues, think like a professional poker player:
- Is the action order correct for the number of players?
- In heads-up (2 players), Button/SB should act FIRST on all postflop streets
- Are the blinds posted correctly? (SB = 10, BB = 20)
- Is the pot math accurate?
- Are side pots handled correctly in all-in situations?
- Does the UI clearly show whose turn it is?
- Can you easily understand the game state?

## Continue Until Stopped

Keep playing hands and documenting findings until the user tells you to stop. Try different scenarios:
- Everyone folds to a bet
- All players check to showdown
- All-in confrontations
- Side pot situations
- Heads-up play (when only 2 players remain)
- Min raises, pot-sized raises, all-in raises

Be creative and try to break things!

## Cleanup

When done, use `concurrent-browser_browser_close_all_instances` to close all browsers.
