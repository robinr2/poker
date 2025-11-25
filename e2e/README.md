# E2E Testing Guide

This document covers the end-to-end testing setup for the poker application. Read this before starting any e2e testing work.

## Quick Start

### 1. Start the Server (in test mode)

```bash
# Build and run with test mode enabled
go build -o ./server ./cmd/server
POKER_TEST_MODE=true ./server

# Or run directly
POKER_TEST_MODE=true go run ./cmd/server
```

Test mode enables special API endpoints:
- `POST /api/test/set-deck` - Set a predetermined deck for deterministic testing
- `POST /api/test/reset-table` - Reset a table to initial state between tests

### 2. Run Tests

```bash
# Run all e2e tests (headless, fast - good for CI)
npx playwright test e2e/

# Run with visible browser (headed mode - good for debugging)
npx playwright test e2e/ --headed

# Run a specific test file
npx playwright test e2e/deterministic-winner.spec.ts

# Run with cleaner output
npx playwright test e2e/ --reporter=line
```

### 3. Test Results (Last Run)

All 9 tests pass in headless mode (~5 minutes total):
- `deterministic-winner.spec.ts` (3 tests) - Pocket Aces vs Kings, Flush vs Trips, Quads wins
- `three-player-complete-hand.spec.ts` - Full hand to showdown
- `three-player-fold-to-win.spec.ts` - Fold-to-win scenario
- `smoke-test.spec.ts` - Basic 3-player connectivity
- `quick-deterministic-test.spec.ts` - Quick deck injection verify
- `single-player-test.spec.ts` - Single player table join
- `diagnostic.spec.ts` - Page load diagnostic

## File Structure

```
e2e/
  helpers/
    browser-helpers.ts        # launchBrowser(), positionWindow() - USE THESE
    deterministic-helpers.ts  # setDeck(), resetTable(), buildDeck()
  deterministic-winner.spec.ts
  three-player-complete-hand.spec.ts
  three-player-fold-to-win.spec.ts
  smoke-test.spec.ts
  single-player-test.spec.ts
  quick-deterministic-test.spec.ts
  diagnostic.spec.ts
  README.md
```

## Writing Tests - Use the Helpers!

### Browser Helpers (IMPORTANT)

All tests should use the browser helpers for consistent headless/headed behavior:

```typescript
import { 
  launchBrowser, 
  positionWindow, 
  WINDOW_WIDTH, 
  WINDOW_HEIGHT 
} from './helpers/browser-helpers';

test.describe('My Test Suite', () => {
  let browser: Browser;

  test.beforeAll(async () => {
    // This automatically handles headless vs headed mode
    browser = await launchBrowser();
  });

  test('my test', async () => {
    const context = await browser.newContext({
      viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
    });
    const page = await context.newPage();
    
    // Position window (no-op in headless mode, positions in headed mode)
    await positionWindow(context, page, 0);  // 0 = first position (left)
  });
});
```

### Deterministic Helpers

For tests that need specific card outcomes:

```typescript
import { setDeck, buildDeck, resetTable } from './helpers/deterministic-helpers';

// Always reset table in beforeEach for test isolation
test.beforeEach(async () => {
  await resetTable('table-1');
});

test('specific hand outcome', async () => {
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

  // Set deck BEFORE starting the hand
  await setDeck('table-1', deck);
  
  // Now start the hand
  await player.startHand();
});
```

### Standard Test Structure

```typescript
import { test, expect, Browser, BrowserContext, Page } from '@playwright/test';
import { resetTable } from './helpers/deterministic-helpers';
import { launchBrowser, positionWindow, WINDOW_WIDTH, WINDOW_HEIGHT } from './helpers/browser-helpers';

const TABLE_ID = 'table-1';

test.describe('My Test Suite', () => {
  let browser: Browser;
  let players: Player[] = [];

  test.beforeAll(async () => {
    browser = await launchBrowser();
  });

  test.beforeEach(async () => {
    // Reset table state for clean test
    await resetTable(TABLE_ID);
    
    // Clean up player contexts from previous test
    for (const player of players) {
      await player.close().catch(() => {});
    }
    players = [];
  });

  test.afterAll(async () => {
    for (const player of players) {
      await player.close().catch(() => {});
    }
    await browser?.close();
  });

  test('my test case', async () => {
    // Create players
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      
      const player = new Player(context, page, `Player${i + 1}`, i);
      players.push(player);
    }
    
    // ... test logic
  });
});
```

### Player Helper Class Pattern

Most tests use a Player class for common actions:

```typescript
class Player {
  constructor(
    public context: BrowserContext,
    public page: Page,
    public name: string,
    public index: number
  ) {}

  async goto(url: string) {
    await this.page.goto(url, { waitUntil: 'domcontentloaded' });
    await this.page.waitForSelector('.app', { timeout: 15000 });
  }

  async enterName() {
    const nameInput = this.page.locator('input[type="text"]');
    await expect(nameInput).toBeVisible({ timeout: 10000 });
    await nameInput.fill(this.name);
    await this.page.locator('button[type="submit"]').click();
    await this.page.waitForTimeout(1500);
  }

  async joinTable() {
    const joinButton = this.page.locator('button.join-button').first();
    await expect(joinButton).toBeVisible({ timeout: 10000 });
    await joinButton.click();
    await this.page.waitForTimeout(2000);
  }

  async startHand() {
    const startButton = this.page.locator('button:has-text("Start Hand")');
    await expect(startButton).toBeVisible({ timeout: 10000 });
    await startButton.click();
    await this.page.waitForTimeout(2000);
  }

  async waitForTurn(timeout = 30000) {
    const actionBar = this.page.locator('.action-bar');
    await expect(actionBar).toBeVisible({ timeout });
  }

  async fold() {
    await this.waitForTurn();
    await this.page.locator('button:has-text("Fold")').click();
    await this.page.waitForTimeout(1500);
  }

  async call() {
    await this.waitForTurn();
    await this.page.locator('button:has-text("Call")').click();
    await this.page.waitForTimeout(1500);
  }

  async check() {
    await this.waitForTurn();
    await this.page.locator('button:has-text("Check")').click();
    await this.page.waitForTimeout(1500);
  }

  async allIn() {
    await this.waitForTurn();
    await this.page.locator('button.preset-button:has-text("All-in")').click();
    await this.page.locator('button.raise-button:has-text("Raise")').click();
    await this.page.waitForTimeout(1500);
  }

  async close() {
    await this.context.close();
  }
}
```

## Common Patterns

### Finding Current Actor

In poker, different players act at different times:

```typescript
async function findCurrentActor(players: Player[]): Promise<Player | null> {
  for (const player of players) {
    const isMyTurn = await player.page.locator('.action-bar').isVisible().catch(() => false);
    if (isMyTurn) return player;
  }
  return null;
}

// Usage
const actor = await findCurrentActor(players);
if (actor) await actor.call();
```

### Verify Showdown Result

```typescript
const showdownOverlay = player.page.locator('.showdown-overlay');
await expect(showdownOverlay).toBeVisible({ timeout: 10000 });
const result = await showdownOverlay.textContent();

expect(result).toContain('Player1');
expect(result.toLowerCase()).toMatch(/three of a kind|trips/i);
```

### Check Stack Changes

```typescript
// Record initial stacks
const initialStacks: number[] = [];
for (const player of players) {
  const stack = await player.getStack();
  initialStacks.push(stack ?? 0);
}

// ... play hand ...

// Verify winner gained chips
const finalStack = await winner.getStack();
expect(finalStack).toBeGreaterThan(initialStacks[winner.index]);
```

## Card String Format

Cards are 2-character strings: `Rank + Suit`

- **Ranks:** A, 2, 3, 4, 5, 6, 7, 8, 9, T, J, Q, K
- **Suits:** s (spades), h (hearts), d (diamonds), c (clubs)

Examples: `As` = Ace of spades, `Kh` = King of hearts, `Td` = Ten of diamonds

## Troubleshooting

### Tests Fail with "405 Method Not Allowed"

**Cause:** Old server process running without new routes.

**Fix:**
```bash
pkill -f "poker" 2>/dev/null
lsof -i :8080  # Verify port is free
POKER_TEST_MODE=true ./server
```

### Tests Interfere with Each Other

**Cause:** Server state persists between tests.

**Fix:** Always call `resetTable()` in `beforeEach`:
```typescript
test.beforeEach(async () => {
  await resetTable(TABLE_ID);
});
```

### "Start Hand" Button Not Appearing

**Cause:** Previous hand still active, or not enough players.

**Fix:** 
1. Ensure `resetTable()` is called
2. Ensure at least 2 players have joined

### Timing Issues / Race Conditions

**Fix:** 
- Use element-based waits: `await expect(button).toBeVisible({ timeout: 10000 })`
- Add small waits after actions for WebSocket propagation: `await page.waitForTimeout(1500)`
- Run in headed mode to see what's happening: `--headed`

### Window Positioning Not Working

**Cause:** Window positioning only works in headed mode.

**Note:** `positionWindow()` is automatically a no-op in headless mode, so this is expected.

## Debugging Tips

1. **Run headed:** `npx playwright test e2e/my-test.spec.ts --headed`

2. **Add console.log:** Shows in test output
   ```typescript
   console.log(`${this.name}: Folding...`);
   ```

3. **Take screenshots:**
   ```typescript
   await page.screenshot({ path: 'debug.png' });
   ```

4. **Check server logs:**
   ```bash
   # Server logs go to stdout, or check server.log if redirected
   ```

5. **Increase timeout:**
   ```typescript
   test.setTimeout(120000);  // 2 minutes
   ```

## Adding New Tests

1. Create `e2e/my-new-test.spec.ts`
2. Import helpers:
   ```typescript
   import { resetTable } from './helpers/deterministic-helpers';
   import { launchBrowser, positionWindow, WINDOW_WIDTH, WINDOW_HEIGHT } from './helpers/browser-helpers';
   ```
3. Use `launchBrowser()` in `beforeAll`
4. Call `resetTable()` in `beforeEach`
5. For deterministic tests, call `setDeck()` before `startHand()`
6. Test in isolation first: `npx playwright test e2e/my-new-test.spec.ts`
7. Then run full suite: `npx playwright test e2e/`
