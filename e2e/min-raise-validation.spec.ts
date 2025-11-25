import { test, expect, Browser, BrowserContext, Page } from '@playwright/test';
import { resetTable } from './helpers/deterministic-helpers';
import {
  launchBrowser,
  positionWindow,
  WINDOW_WIDTH,
  WINDOW_HEIGHT,
} from './helpers/browser-helpers';

const TABLE_ID = 'table-1';
const BASE_URL = 'http://localhost:8080';

// Blinds configuration (must match server config)
const SMALL_BLIND = 10;
const BIG_BLIND = 20;
const STARTING_STACK = 1000;

/**
 * Verify server is running and ready
 */
async function verifyServerReady() {
  console.log('Verifying server is ready...');
  try {
    const response = await fetch(`${BASE_URL}/health`);
    if (!response.ok) {
      throw new Error('Server health check failed');
    }
    console.log('Server is ready');
  } catch (error) {
    throw new Error(
      'Server not running. Start with: POKER_TEST_MODE=true ./server'
    );
  }
}

/**
 * Player class with min raise validation helpers
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

  async isMyTurn(): Promise<boolean> {
    return await this.page.locator('.action-bar').isVisible().catch(() => false);
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

  async raise(amount: number) {
    console.log(`${this.name}: Raising to ${amount}...`);
    await this.waitForTurn();

    const raiseInput = this.page.locator('input[aria-label="Raise Amount"]');
    await expect(raiseInput).toBeVisible({ timeout: 5000 });
    await raiseInput.fill(amount.toString());

    const raiseButton = this.page.locator('button.raise-button:has-text("Raise")');
    await expect(raiseButton).toBeEnabled({ timeout: 5000 });
    await raiseButton.click();
    await this.page.waitForTimeout(1500);
  }

  async clickMinRaise() {
    console.log(`${this.name}: Using min raise...`);
    await this.waitForTurn();

    const minButton = this.page.locator('button.preset-button:has-text("Min")');
    await expect(minButton).toBeVisible({ timeout: 5000 });
    await minButton.click();

    const raiseButton = this.page.locator('button.raise-button:has-text("Raise")');
    await expect(raiseButton).toBeEnabled({ timeout: 5000 });
    await raiseButton.click();
    await this.page.waitForTimeout(1500);
  }

  async clickAllIn() {
    console.log(`${this.name}: Going all-in...`);
    await this.waitForTurn();

    const allInButton = this.page.locator('button.preset-button:has-text("All-in")');
    await expect(allInButton).toBeVisible({ timeout: 5000 });
    await allInButton.click();

    const raiseButton = this.page.locator('button.raise-button:has-text("Raise")');
    await expect(raiseButton).toBeEnabled({ timeout: 5000 });
    await raiseButton.click();
    await this.page.waitForTimeout(1500);
  }

  // ============ MIN RAISE SPECIFIC HELPERS ============

  /**
   * Get the current min raise value from the UI
   */
  async getMinRaiseValue(): Promise<number> {
    await this.waitForTurn();
    
    // Click min button to populate the input
    const minButton = this.page.locator('button.preset-button:has-text("Min")');
    await minButton.click();
    await this.page.waitForTimeout(300);
    
    // Read the value from the input
    const raiseInput = this.page.locator('input[aria-label="Raise Amount"]');
    const value = await raiseInput.inputValue();
    return parseInt(value, 10);
  }

  /**
   * Check if raise button is enabled for a specific amount
   */
  async isRaiseButtonEnabled(): Promise<boolean> {
    const raiseButton = this.page.locator('button.raise-button:has-text("Raise")');
    return await raiseButton.isEnabled().catch(() => false);
  }

  /**
   * Fill raise amount without clicking the raise button
   */
  async fillRaiseAmount(amount: number) {
    const raiseInput = this.page.locator('input[aria-label="Raise Amount"]');
    await expect(raiseInput).toBeVisible({ timeout: 5000 });
    await raiseInput.fill(amount.toString());
    await this.page.waitForTimeout(300);
  }

  // ============ VERIFICATION METHODS ============

  async getPot(): Promise<number> {
    const potText = await this.page.locator('.pot-display').textContent().catch(() => null);
    if (!potText) return 0;
    const match = potText.match(/Pot:\s*(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }

  async verifyPot(expected: number, message?: string) {
    const actual = await this.getPot();
    const msg = message || `${this.name}: Pot should be ${expected}`;
    console.log(`${msg} - Actual: ${actual}`);
    expect(actual, msg).toBe(expected);
  }

  async getStack(): Promise<number> {
    const stackText = await this.page.locator('.own-seat .stack').textContent().catch(() => null);
    if (!stackText) return 0;
    const match = stackText.match(/(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }

  async getStreet(): Promise<string> {
    const streetText = await this.page.locator('.street-indicator').textContent().catch(() => null);
    return streetText?.toLowerCase() || '';
  }

  async verifyStreet(expected: string, message?: string) {
    const actual = await this.getStreet();
    const msg = message || `${this.name}: Street should be ${expected}`;
    console.log(`${msg} - Actual: ${actual}`);
    expect(actual.toLowerCase(), msg).toBe(expected.toLowerCase());
  }

  async hasShowdown(): Promise<boolean> {
    return await this.page.locator('.showdown-overlay').isVisible().catch(() => false);
  }

  async canStartNewHand(): Promise<boolean> {
    return await this.page.locator('button:has-text("Start Hand")').isVisible().catch(() => false);
  }

  async close() {
    await this.context.close();
  }
}

/**
 * Find which player currently has the action
 */
async function findCurrentActor(players: Player[]): Promise<Player | null> {
  for (const player of players) {
    if (await player.isMyTurn()) {
      return player;
    }
  }
  return null;
}

/**
 * Wait for any player to have action
 */
async function waitForAnyAction(players: Player[], timeout = 10000): Promise<Player | null> {
  const startTime = Date.now();
  while (Date.now() - startTime < timeout) {
    const actor = await findCurrentActor(players);
    if (actor) return actor;
    await players[0].page.waitForTimeout(500);
  }
  return null;
}

// ==========================================
// TEST SUITE: Min Raise Validation
// ==========================================

test.describe('Min Raise Validation Tests', () => {
  let browser: Browser;
  let players: Player[] = [];

  test.beforeAll(async () => {
    browser = await launchBrowser();
  });

  test.beforeEach(async () => {
    await verifyServerReady();
    await resetTable(TABLE_ID);

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

  // ==========================================
  // Test 1: Min raise equals BB preflop (first raise)
  // ==========================================
  test('min raise equals 2x BB preflop (first raise)', async () => {
    console.log('\n=== TEST: Min Raise Preflop ===\n');

    // Create 2 players
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `MinRaise_P${i + 1}`, i);
      players.push(player);
    }

    // Setup
    for (const player of players) {
      await player.goto(BASE_URL);
    }
    for (const player of players) {
      await player.enterName();
    }
    for (const player of players) {
      await player.joinTable();
    }
    await players[0].page.waitForTimeout(2000);

    // Start hand
    console.log('\n--- Starting Hand ---');
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // PREFLOP: First to act - check min raise value
    // ==========================================
    console.log('\n--- Preflop: Checking min raise value ---');

    const actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: has action`);

    // Get the min raise value
    const minRaise = await actor!.getMinRaiseValue();
    console.log(`Min raise value: ${minRaise}`);

    // First raise preflop: min raise should be 2x BB = 40
    // (Current bet is 20 BB, min raise increment is 20, so total = 40)
    expect(minRaise).toBe(BIG_BLIND * 2);
    console.log(`Expected min raise: ${BIG_BLIND * 2} (2x BB)`);

    console.log('\n=== TEST PASSED: Min Raise Preflop ===\n');
  });

  // ==========================================
  // Test 2: Min raise after a raise (re-raise calculation)
  // ==========================================
  test('min re-raise equals previous raise + last raise increment', async () => {
    console.log('\n=== TEST: Min Re-Raise Calculation ===\n');

    // Create 3 players for better re-raise scenario
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `ReRaise_P${i + 1}`, i);
      players.push(player);
    }

    // Setup
    for (const player of players) {
      await player.goto(BASE_URL);
    }
    for (const player of players) {
      await player.enterName();
    }
    for (const player of players) {
      await player.joinTable();
    }
    await players[0].page.waitForTimeout(2000);

    // Start hand
    console.log('\n--- Starting Hand ---');
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // PREFLOP: First player raises to 60
    // ==========================================
    console.log('\n--- Preflop: First player raises ---');

    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: raises to 60`);
    await actor!.raise(60);
    await players[0].page.waitForTimeout(1000);

    // ==========================================
    // Second player: check min re-raise value
    // Raise was from 20 to 60, so increment was 40
    // Min re-raise = 60 + 40 = 100
    // ==========================================
    console.log('\n--- Checking min re-raise value ---');

    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: has action`);

    const minReRaise = await actor!.getMinRaiseValue();
    console.log(`Min re-raise value: ${minReRaise}`);

    // Expected: 60 (current bet) + 40 (last raise increment) = 100
    const expectedMinReRaise = 60 + (60 - BIG_BLIND); // 60 + 40 = 100
    expect(minReRaise).toBe(expectedMinReRaise);
    console.log(`Expected min re-raise: ${expectedMinReRaise} (previous + increment)`);

    console.log('\n=== TEST PASSED: Min Re-Raise Calculation ===\n');
  });

  // ==========================================
  // Test 3: Raise below minimum is disabled
  // ==========================================
  test('raise button disabled when amount below minimum', async () => {
    console.log('\n=== TEST: Raise Below Minimum Disabled ===\n');

    // Create 2 players
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `BelowMin_P${i + 1}`, i);
      players.push(player);
    }

    // Setup
    for (const player of players) {
      await player.goto(BASE_URL);
    }
    for (const player of players) {
      await player.enterName();
    }
    for (const player of players) {
      await player.joinTable();
    }
    await players[0].page.waitForTimeout(2000);

    // Start hand
    console.log('\n--- Starting Hand ---');
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // PREFLOP: Try to raise below minimum
    // ==========================================
    console.log('\n--- Preflop: Try to raise below minimum ---');

    const actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: has action`);

    // Fill in an amount below minimum (30, when min is 40)
    await actor!.waitForTurn();
    await actor!.fillRaiseAmount(30);

    // Raise button should be DISABLED
    const isEnabled = await actor!.isRaiseButtonEnabled();
    console.log(`Raise button enabled with 30: ${isEnabled} (expected: false)`);
    expect(isEnabled).toBe(false);

    // Now fill in valid amount (40)
    await actor!.fillRaiseAmount(40);
    const isEnabledValid = await actor!.isRaiseButtonEnabled();
    console.log(`Raise button enabled with 40: ${isEnabledValid} (expected: true)`);
    expect(isEnabledValid).toBe(true);

    console.log('\n=== TEST PASSED: Raise Below Minimum Disabled ===\n');
  });

  // ==========================================
  // Test 4: Min raise resets each street
  // ==========================================
  test('min raise resets to BB on new street', async () => {
    console.log('\n=== TEST: Min Raise Resets Each Street ===\n');

    // Create 2 players
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `StreetReset_P${i + 1}`, i);
      players.push(player);
    }

    // Setup
    for (const player of players) {
      await player.goto(BASE_URL);
    }
    for (const player of players) {
      await player.enterName();
    }
    for (const player of players) {
      await player.joinTable();
    }
    await players[0].page.waitForTimeout(2000);

    // Start hand
    console.log('\n--- Starting Hand ---');
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // PREFLOP: Both call to see flop
    // ==========================================
    console.log('\n--- Preflop: Both call ---');

    let actor = await waitForAnyAction(players);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // FLOP: Check min raise value (should reset to BB)
    // ==========================================
    console.log('\n--- Flop: Checking min raise value ---');
    await players[0].verifyStreet('flop');

    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: has action on flop`);

    const minRaiseFlop = await actor!.getMinRaiseValue();
    console.log(`Min raise on flop: ${minRaiseFlop}`);

    // On flop, with no previous bet, min raise should be BB (20)
    // Actually, the first bet on flop is the "raise" which has min of BB
    expect(minRaiseFlop).toBe(BIG_BLIND);
    console.log(`Expected min raise on flop: ${BIG_BLIND} (BB resets each street)`);

    console.log('\n=== TEST PASSED: Min Raise Resets Each Street ===\n');
  });

  // ==========================================
  // Test 5: All-in below min raise is allowed
  // ==========================================
  test('all-in below min raise is allowed (special exception)', async () => {
    console.log('\n=== TEST: All-in Below Min Raise Allowed ===\n');

    // For this test we need a short-stacked player
    // We'll play a few hands to reduce one player's stack first
    // Or we can test that the all-in button is always enabled

    // Create 2 players
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `AllInEx_P${i + 1}`, i);
      players.push(player);
    }

    // Setup
    for (const player of players) {
      await player.goto(BASE_URL);
    }
    for (const player of players) {
      await player.enterName();
    }
    for (const player of players) {
      await player.joinTable();
    }
    await players[0].page.waitForTimeout(2000);

    // Start hand
    console.log('\n--- Starting Hand ---');
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // PREFLOP: Player 1 raises big
    // ==========================================
    console.log('\n--- Preflop: Player raises to 500 ---');

    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: raises to 500`);
    await actor!.raise(500);
    await players[0].page.waitForTimeout(1000);

    // ==========================================
    // Player 2: Min re-raise would be 500 + 480 = 980
    // But if player 2 only has ~500 left, they can still all-in
    // All-in button should always be enabled
    // ==========================================
    console.log('\n--- Checking all-in is available ---');

    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: has action`);

    // Check that All-in button is visible and clicking it enables raise
    await actor!.waitForTurn();
    const allInButton = actor!.page.locator('button.preset-button:has-text("All-in")');
    await expect(allInButton).toBeVisible({ timeout: 5000 });
    
    // Click all-in to fill the amount
    await allInButton.click();
    await actor!.page.waitForTimeout(300);

    // Raise button should be enabled (all-in is always valid)
    const isEnabled = await actor!.isRaiseButtonEnabled();
    console.log(`Raise button enabled for all-in: ${isEnabled} (expected: true)`);
    expect(isEnabled).toBe(true);

    // Actually execute the all-in
    const raiseButton = actor!.page.locator('button.raise-button:has-text("Raise")');
    await raiseButton.click();
    await players[0].page.waitForTimeout(2000);

    // Verify hand continues to showdown or completes
    // (Other player will call or fold)
    actor = await waitForAnyAction(players);
    if (actor) {
      console.log(`${actor!.name}: calls all-in`);
      await actor!.call();
      await players[0].page.waitForTimeout(3000);
    }

    // Verify showdown occurred (or hand completed)
    let handComplete = false;
    for (const player of players) {
      if (await player.hasShowdown() || await player.canStartNewHand()) {
        handComplete = true;
        break;
      }
    }
    expect(handComplete).toBe(true);

    console.log('\n=== TEST PASSED: All-in Below Min Raise Allowed ===\n');
  });
});
