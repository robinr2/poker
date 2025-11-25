# E2E Testing Guide

This document covers the end-to-end testing setup for the poker application, including lessons learned, common pitfalls, and best practices.

## Prerequisites

### Server Setup

The server must be running with test mode enabled:

```bash
POKER_TEST_MODE=true go run ./cmd/server
```

Or build and run:

```bash
go build -o ./server ./cmd/server
POKER_TEST_MODE=true ./server
```

Test mode enables special API endpoints used by e2e tests:
- `POST /api/test/set-deck` - Set a predetermined deck for deterministic testing
- `POST /api/test/reset-table` - Reset a table to initial state

### Running Tests

```bash
# Run all e2e tests
npx playwright test e2e/

# Run a specific test file
npx playwright test e2e/deterministic-winner.spec.ts

# Run with visible browser (non-headless)
npx playwright test e2e/ --headed

# Run with line reporter for cleaner output
npx playwright test e2e/ --reporter=line
```

## Test Architecture

### Test Structure

Each test file typically follows this pattern:

```typescript
test.describe('Test Suite Name', () => {
  let browser: Browser;
  let players: Player[] = [];
  const TABLE_ID = 'table-1';

  test.beforeAll(async () => {
    // Launch browser once for all tests
    browser = await chromium.launch({ headless: false, slowMo: 300 });
  });

  test.beforeEach(async () => {
    // Verify server is running
    await verifyServerReady();
    
    // Reset table state for clean test
    await resetTable(TABLE_ID);
    
    // Clean up any existing player contexts
    for (const player of players) {
      await player.close().catch(() => {});
    }
    players = [];
  });

  test.afterAll(async () => {
    // Cleanup
    for (const player of players) {
      await player.close().catch(() => {});
    }
    await browser?.close();
  });

  test('test case', async () => {
    // Test implementation
  });
});
```

### Player Helper Class

Tests use a `Player` class to manage browser contexts and common actions:

```typescript
class Player {
  constructor(
    public context: BrowserContext,
    public page: Page,
    public name: string,
    public index: number
  ) {}

  async enterName() { /* ... */ }
  async joinTable() { /* ... */ }
  async startHand() { /* ... */ }
  async fold() { /* ... */ }
  async call() { /* ... */ }
  async check() { /* ... */ }
  async allIn() { /* ... */ }
  // etc.
}
```

## Lessons Learned & Caveats

### 1. Server Process Management

**Problem:** Tests were failing with "405 Method Not Allowed" even though the endpoint was registered.

**Root Cause:** An old server process (without the new routes) was still running on port 8080.

**Solution:** Always ensure you kill old server processes before starting a new one:

```bash
# Kill any existing server
pkill -f "poker" 2>/dev/null
lsof -i :8080  # Check if port is still in use

# Then start fresh
POKER_TEST_MODE=true ./server
```

**Best Practice:** The test's `verifyServerReady()` function only checks if a server is running, not if it has the correct routes. If you add new test endpoints, rebuild and restart the server.

### 2. Test Isolation - Table State

**Problem:** Tests were interfering with each other. A hand started in one test would still be active in the next test, causing "Start Hand" button to not appear.

**Root Cause:** The server maintains state between tests since we don't restart it.

**Solution:** Added a `resetTable` API endpoint and helper that:
- Clears the current hand
- Resets all seats to empty with initial stacks
- Clears dealer position and test deck

```typescript
// In beforeEach
await resetTable(TABLE_ID);
```

**Best Practice:** Always call `resetTable()` in `beforeEach` to ensure each test starts with a clean slate.

### 3. Test Isolation - Browser Contexts

**Problem:** Player sessions from previous tests could persist.

**Solution:** Each player uses a separate `BrowserContext` (not just a new page). Close all contexts in `beforeEach`:

```typescript
for (const player of players) {
  await player.close().catch(() => {});  // catch errors from already-closed contexts
}
players = [];
```

### 4. Timing and Race Conditions

**Problem:** Actions happening too fast, elements not ready, or WebSocket messages not yet processed.

**Solutions:**
- Use `slowMo: 300` in browser launch options for visual debugging
- Add explicit waits after actions: `await page.waitForTimeout(1500)`
- Wait for specific elements rather than fixed timeouts when possible:
  ```typescript
  await expect(startButton).toBeVisible({ timeout: 10000 });
  ```

**Best Practice:** Prefer element-based waits over fixed timeouts, but use small fixed waits after actions to let WebSocket messages propagate.

### 5. Finding the Current Actor

**Problem:** In poker, different players act at different times. Tests need to find who should act.

**Solution:** Check which player's action bar is visible:

```typescript
async function findCurrentActor(players: Player[]): Promise<Player | null> {
  for (const player of players) {
    const isMyTurn = await player.page.locator('.action-bar').isVisible().catch(() => false);
    if (isMyTurn) {
      return player;
    }
  }
  return null;
}
```

### 6. Deterministic Testing

**Problem:** Poker involves random card dealing, making it hard to verify specific outcomes.

**Solution:** Use the deterministic deck helpers:

```typescript
import { setDeck, buildDeck, resetTable } from './helpers/deterministic-helpers';

// Build a deck with specific hole cards and board
const deck = buildDeck({
  holeCards: {
    0: ['As', 'Ah'],  // Seat 0: Pocket Aces
    1: ['Ks', 'Kh'],  // Seat 1: Pocket Kings
  },
  flop: ['Ad', '2c', '7d'],
  turn: '3h',
  river: '9s',
});

// Set the deck BEFORE starting the hand
await setDeck('table-1', deck);

// Now start the hand - cards will be dealt in predetermined order
await player.startHand();
```

**Important:** The deck must be set BEFORE `startHand()` is called. The deck is consumed on hand start and cleared.

### 7. Card String Format

Cards are represented as 2-character strings: `Rank + Suit`

- **Ranks:** A, 2, 3, 4, 5, 6, 7, 8, 9, T, J, Q, K
- **Suits:** s (spades), h (hearts), d (diamonds), c (clubs)

Examples: `As` = Ace of spades, `Kh` = King of hearts, `Td` = Ten of diamonds

### 8. Window Positioning for Debugging

When running tests visually, position browser windows side-by-side:

```typescript
const positions = [
  { x: 0, y: 40 },      // Player 1 - left
  { x: 950, y: 40 },    // Player 2 - middle  
  { x: 1900, y: 40 },   // Player 3 - right
];

// Use CDP to position windows
const client = await context.newCDPSession(page);
const { windowId } = await client.send('Browser.getWindowForTarget');
await client.send('Browser.setWindowBounds', {
  windowId,
  bounds: { left: positions[i].x, top: positions[i].y, width: 950, height: 1200 },
});
```

## Common Test Patterns

### Pattern: Wait for Turn Then Act

```typescript
async waitForTurn(timeout = 30000) {
  const actionBar = this.page.locator('.action-bar');
  await expect(actionBar).toBeVisible({ timeout });
}

async fold() {
  await this.waitForTurn();
  const foldButton = this.page.locator('button:has-text("Fold")');
  await foldButton.click();
  await this.page.waitForTimeout(1500);
}
```

### Pattern: Verify Showdown Result

```typescript
async getShowdownResult(): Promise<string> {
  const showdownOverlay = this.page.locator('.showdown-overlay');
  await expect(showdownOverlay).toBeVisible({ timeout: 10000 });
  return await showdownOverlay.textContent() || '';
}

// In test
const result = await player.getShowdownResult();
expect(result).toContain('Player1');
expect(result.toLowerCase()).toContain('three of a kind');
```

### Pattern: Check Stack Changes

```typescript
// Record initial stacks
const initialStacks = await Promise.all(players.map(p => p.getStack()));

// ... play hand ...

// Verify stack changes
const finalStacks = await Promise.all(players.map(p => p.getStack()));
expect(finalStacks[0]).toBeGreaterThan(initialStacks[0]!);  // Winner gained chips
```

## Debugging Tips

1. **Use `--headed` mode** to see what's happening in the browser

2. **Add console.log statements** - they appear in test output:
   ```typescript
   console.log(`${this.name}: Folding...`);
   ```

3. **Take screenshots on failure:**
   ```typescript
   await page.screenshot({ path: 'debug-screenshot.png' });
   ```

4. **Check server logs** for backend issues:
   ```bash
   tail -f /tmp/server.log
   ```

5. **Increase timeouts** when debugging:
   ```typescript
   test.setTimeout(120000);  // 2 minutes
   ```

## File Structure

```
e2e/
  helpers/
    deterministic-helpers.ts   # setDeck, resetTable, buildDeck utilities
  deterministic-winner.spec.ts # Tests with predetermined decks
  three-player-complete-hand.spec.ts
  three-player-fold-to-win.spec.ts
  smoke-test.spec.ts           # Basic connectivity tests
  diagnostic.spec.ts           # Debug/diagnostic tests
  README.md                    # This file
```

## Adding New Tests

1. Create a new `.spec.ts` file in `e2e/`
2. Import helpers: `import { resetTable } from './helpers/deterministic-helpers'`
3. Use the standard test structure with `beforeEach` calling `resetTable()`
4. For deterministic tests, call `setDeck()` before `startHand()`
5. Run your test in isolation first: `npx playwright test e2e/your-test.spec.ts`
6. Then run the full suite to ensure no interference
