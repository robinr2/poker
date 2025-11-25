import { test, expect, Browser, BrowserContext, Page } from '@playwright/test';
import { setDeck, buildDeck, DECK_SCENARIOS, resetTable } from './helpers/deterministic-helpers';
import {
  launchBrowser,
  positionWindow,
  WINDOW_WIDTH,
  WINDOW_HEIGHT,
} from './helpers/browser-helpers';

/**
 * Verify server is running with test mode enabled
 * Assumes server was started with: POKER_TEST_MODE=true ./server
 */
async function verifyServerReady() {
  console.log('Verifying server is ready...');

  try {
    const response = await fetch('http://localhost:8080/health');
    if (!response.ok) {
      throw new Error('Server health check failed');
    }
    console.log('Server is ready');
  } catch (error) {
    throw new Error(
      'Server not running. Start with: POKER_TEST_MODE=true ./server (or go run ./cmd/server)'
    );
  }
}

/**
 * Helper class to manage a poker player (browser context)
 */
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
    console.log(`${this.name}: Entering name...`);
    const nameInput = this.page.locator('input[type="text"]');
    await expect(nameInput).toBeVisible({ timeout: 10000 });
    await nameInput.fill(this.name);
    await this.page.locator('button[type="submit"]').click();
    await this.page.waitForTimeout(1500);
  }

  async joinTable() {
    console.log(`${this.name}: Joining table...`);
    const joinButton = this.page.locator('button.join-button').first();
    await expect(joinButton).toBeVisible({ timeout: 10000 });
    await joinButton.click();
    await this.page.waitForTimeout(2000);

    // Verify we're at the table
    const leaveButton = this.page.locator('button:has-text("Leave")');
    await expect(leaveButton).toBeVisible({ timeout: 10000 });
    console.log(`${this.name}: Successfully joined table`);
  }

  async startHand() {
    console.log(`${this.name}: Starting hand...`);
    const startButton = this.page.locator('button:has-text("Start Hand")');
    await expect(startButton).toBeVisible({ timeout: 10000 });
    await startButton.click();
    await this.page.waitForTimeout(2000);
  }

  async waitForTurn(timeout = 30000) {
    console.log(`${this.name}: Waiting for turn...`);
    const actionBar = this.page.locator('.action-bar');
    await expect(actionBar).toBeVisible({ timeout });
    console.log(`${this.name}: It's my turn!`);
  }

  async fold() {
    console.log(`${this.name}: Folding...`);
    await this.waitForTurn();
    const foldButton = this.page.locator('button:has-text("Fold")');
    await foldButton.click();
    await this.page.waitForTimeout(1500);
  }

  async check() {
    console.log(`${this.name}: Checking...`);
    await this.waitForTurn();
    const checkButton = this.page.locator('button:has-text("Check")');
    await expect(checkButton).toBeVisible({ timeout: 5000 });
    await checkButton.click();
    await this.page.waitForTimeout(1500);
  }

  async call() {
    console.log(`${this.name}: Calling...`);
    await this.waitForTurn();
    const callButton = this.page.locator('button:has-text("Call")');
    await expect(callButton).toBeVisible({ timeout: 5000 });
    await callButton.click();
    await this.page.waitForTimeout(1500);
  }

  async allIn() {
    console.log(`${this.name}: Going all-in...`);
    await this.waitForTurn();

    // Click All-in preset button to fill in all-in amount and submit
    const allInButton = this.page.locator('button.preset-button:has-text("All-in")');
    await expect(allInButton).toBeVisible({ timeout: 5000 });
    await allInButton.click();

    // Click raise button to submit
    const raiseButton = this.page.locator('button.raise-button:has-text("Raise")');
    await expect(raiseButton).toBeEnabled({ timeout: 5000 });
    await raiseButton.click();
    await this.page.waitForTimeout(1500);
  }

  async getShowdownResult(): Promise<string> {
    const showdownOverlay = this.page.locator('.showdown-overlay');
    await expect(showdownOverlay).toBeVisible({ timeout: 10000 });
    const text = await showdownOverlay.textContent();
    return text || '';
  }

  async close() {
    await this.context.close();
  }
}

/**
 * Deterministic Poker Tests
 *
 * These tests use predetermined deck orders to verify exact game outcomes.
 * The server must be running with POKER_TEST_MODE=true for these tests to work.
 *
 * To run these tests:
 * 1. Start the server with: POKER_TEST_MODE=true go run ./cmd/server
 * 2. Run the tests: npx playwright test e2e/deterministic-winner.spec.ts
 */
test.describe('Deterministic Winner Tests', () => {
  let browser: Browser;
  let players: Player[] = [];
  const baseURL = 'http://localhost:8080';
  const TABLE_ID = 'table-1';

  test.beforeAll(async () => {
    browser = await launchBrowser();
  });

  test.beforeEach(async () => {
    // Verify server is running (don't restart - assume it's already running with POKER_TEST_MODE=true)
    await verifyServerReady();

    // Reset the table state before each test to ensure clean state
    await resetTable(TABLE_ID);

    // Clear any existing players
    for (const player of players) {
      await player.close().catch(() => {});
    }
    players = [];
  });

  test.afterAll(async () => {
    for (const player of players) {
      await player.close().catch(() => {});
    }
    if (browser) {
      await browser.close();
    }
  });

  test('Player with pocket Aces beats player with pocket Kings', async () => {
    console.log('\n=== DETERMINISTIC TEST: Pocket Aces vs Pocket Kings ===');

    // ==========================================
    // SETUP: Create 2 players
    // ==========================================
    console.log('\n=== SETUP: Creating 2 players ===');

    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();

      // Position window (only in headed mode)
      await positionWindow(context, page, i);

      const player = new Player(context, page, `Player${i + 1}`, i);
      players.push(player);
    }

    // ==========================================
    // Navigate and enter names
    // ==========================================
    console.log('\n=== Navigating players to app ===');
    for (const player of players) {
      await player.goto(baseURL);
    }

    console.log('\n=== Players entering names ===');
    for (const player of players) {
      await player.enterName();
    }

    // ==========================================
    // All players join the table
    // ==========================================
    console.log('\n=== Players joining table ===');
    for (const player of players) {
      await player.joinTable();
    }

    // Wait for all players to be fully seated
    await players[0].page.waitForTimeout(3000);

    // ==========================================
    // SET DETERMINISTIC DECK
    // ==========================================
    console.log('\n=== Setting deterministic deck ===');

    // Build a deck where:
    // - Seat 0 (Player1) gets Pocket Aces (As, Ah)
    // - Seat 1 (Player2) gets Pocket Kings (Ks, Kh)
    // - Board: Ad, 2c, 7d, 3h, 9s (Player1 makes trip Aces)
    const deck = buildDeck({
      holeCards: {
        0: ['As', 'Ah'], // Player1: Pocket Aces
        1: ['Ks', 'Kh'], // Player2: Pocket Kings
      },
      flop: ['Ad', '2c', '7d'], // Player1 hits set of Aces
      turn: '3h',
      river: '9s',
    });

    console.log('Deck prepared, calling setDeck API...');
    await setDeck(TABLE_ID, deck);
    console.log('Deck set successfully!');

    // ==========================================
    // START HAND
    // ==========================================
    console.log('\n=== STARTING HAND ===');
    await players[0].startHand();

    // Wait for hand to be dealt
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // PLAY: Go all-in (to reach showdown quickly)
    // ==========================================
    console.log('\n=== PLAYING HAND (all-in) ===');

    // Find who has the action and go all-in
    for (let i = 0; i < players.length; i++) {
      const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
      if (isMyTurn) {
        console.log(`${players[i].name} going all-in`);
        await players[i].allIn();
        break;
      }
    }

    await players[0].page.waitForTimeout(1500);

    // Second player calls
    for (let i = 0; i < players.length; i++) {
      const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
      if (isMyTurn) {
        console.log(`${players[i].name} calling all-in`);
        await players[i].call();
        break;
      }
    }

    // ==========================================
    // VERIFY SHOWDOWN RESULT
    // ==========================================
    console.log('\n=== VERIFYING SHOWDOWN ===');

    // Wait for showdown overlay
    await players[0].page.waitForTimeout(5000);

    const showdownResult = await players[0].getShowdownResult();
    console.log('Showdown result:', showdownResult);

    // Player1 (with Aces) should win
    // The exact text depends on your UI implementation
    expect(showdownResult).toContain('Player1');
    expect(showdownResult.toLowerCase()).toMatch(/three of a kind|trips|set/i);

    console.log('\n=== TEST PASSED: Player1 (Pocket Aces) won as expected ===');
  });

  test('Player with flush beats player with trips', async () => {
    console.log('\n=== DETERMINISTIC TEST: Flush vs Trips ===');

    // ==========================================
    // SETUP: Create 2 players
    // ==========================================
    console.log('\n=== SETUP: Creating 2 players ===');

    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();

      // Position window (only in headed mode)
      await positionWindow(context, page, i);

      const player = new Player(context, page, `FlushPlayer${i + 1}`, i);
      players.push(player);
    }

    // Navigate, enter names, join table
    for (const player of players) {
      await player.goto(baseURL);
    }
    for (const player of players) {
      await player.enterName();
    }
    for (const player of players) {
      await player.joinTable();
    }

    await players[0].page.waitForTimeout(3000);

    // ==========================================
    // SET DETERMINISTIC DECK - Flush wins over Trips
    // ==========================================
    console.log('\n=== Setting deterministic deck (Flush vs Trips) ===');

    const deck = buildDeck({
      holeCards: {
        0: ['As', 'Ac'], // Player1: Pocket Aces (will make trips)
        1: ['2h', '3h'], // Player2: Suited hearts (will make flush)
      },
      flop: ['Ad', '5h', '9h'], // Trips for P1, flush draw for P2
      turn: 'Kh', // P2 makes flush!
      river: '2c',
    });

    await setDeck(TABLE_ID, deck);
    console.log('Deck set successfully!');

    // Start hand and play all-in
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // Go all-in
    for (let i = 0; i < players.length; i++) {
      const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
      if (isMyTurn) {
        await players[i].allIn();
        break;
      }
    }

    await players[0].page.waitForTimeout(1500);

    for (let i = 0; i < players.length; i++) {
      const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
      if (isMyTurn) {
        await players[i].call();
        break;
      }
    }

    // Verify showdown
    await players[0].page.waitForTimeout(5000);

    const showdownResult = await players[0].getShowdownResult();
    console.log('Showdown result:', showdownResult);

    // Player2 (with flush) should win
    expect(showdownResult).toContain('FlushPlayer2');
    expect(showdownResult.toLowerCase()).toContain('flush');

    console.log('\n=== TEST PASSED: FlushPlayer2 (Flush) won as expected ===');
  });

  test('Three player game - highest hand wins', async () => {
    console.log('\n=== DETERMINISTIC TEST: Three Player Game ===');

    // ==========================================
    // SETUP: Create 3 players
    // ==========================================
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();

      // Position window (only in headed mode)
      await positionWindow(context, page, i);

      const player = new Player(context, page, `TrioPlayer${i + 1}`, i);
      players.push(player);
    }

    // Navigate, enter names, join table
    for (const player of players) {
      await player.goto(baseURL);
    }
    for (const player of players) {
      await player.enterName();
    }
    for (const player of players) {
      await player.joinTable();
    }

    await players[0].page.waitForTimeout(3000);

    // ==========================================
    // SET DETERMINISTIC DECK - Player 2 wins with Quads
    // ==========================================
    console.log('\n=== Setting deterministic deck (Quads wins) ===');

    // Use the pre-built scenario where seat 1 wins with quads
    await setDeck(TABLE_ID, DECK_SCENARIOS.threePlayerSeat1WinsQuads);
    console.log('Deck set successfully!');

    // Start hand
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // Everyone goes all-in
    for (let round = 0; round < 3; round++) {
      for (let i = 0; i < players.length; i++) {
        const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
        if (isMyTurn) {
          // First player goes all-in, others call
          if (round === 0) {
            await players[i].allIn();
          } else {
            await players[i].call();
          }
          break;
        }
      }
      await players[0].page.waitForTimeout(1500);
    }

    // Verify showdown
    await players[0].page.waitForTimeout(5000);

    const showdownResult = await players[0].getShowdownResult();
    console.log('Showdown result:', showdownResult);

    // Player 2 (TrioPlayer2, seat 1) should win with quads
    expect(showdownResult).toContain('TrioPlayer2');
    expect(showdownResult.toLowerCase()).toMatch(/four of a kind|quads/i);

    console.log('\n=== TEST PASSED: TrioPlayer2 (Quads) won as expected ===');
  });
});
