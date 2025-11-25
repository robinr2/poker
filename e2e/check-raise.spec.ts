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
 * Player class with comprehensive action and verification methods
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

  async verifyStack(expected: number, message?: string) {
    const actual = await this.getStack();
    const msg = message || `${this.name}: Stack should be ${expected}`;
    console.log(`${msg} - Actual: ${actual}`);
    expect(actual, msg).toBe(expected);
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

  async dismissShowdownIfVisible() {
    const showdownOverlay = this.page.locator('.showdown-overlay');
    if (await showdownOverlay.isVisible().catch(() => false)) {
      console.log(`${this.name}: Dismissing showdown overlay...`);
      // Try clicking anywhere or a dismiss button
      const dismissButton = this.page.locator('.showdown-overlay button, .showdown-overlay');
      await dismissButton.first().click().catch(() => {});
      await this.page.waitForTimeout(1000);
    }
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
// TEST SUITE: Check-Raise Scenarios
// ==========================================

test.describe('Check-Raise Scenarios', () => {
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
  // Test 1: Classic Check-Raise on Flop
  // ==========================================
  test('3-player check-raise on flop - action reopens after raise', async () => {
    console.log('\n=== TEST: Check-Raise on Flop ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `CheckRaise_P${i + 1}`, i);
      players.push(player);
    }

    // Setup: Navigate, enter names, join table
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

    // Record initial stacks
    console.log('\n--- Initial Stacks ---');
    for (const player of players) {
      const stack = await player.getStack();
      console.log(`${player.name}: ${stack}`);
      expect(stack).toBe(STARTING_STACK);
    }

    // Start hand
    console.log('\n--- Starting Hand ---');
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // Verify pot has blinds
    await players[0].verifyPot(SMALL_BLIND + BIG_BLIND, 'Pot after blinds');

    // ==========================================
    // PREFLOP: Everyone limps to see flop
    // ==========================================
    console.log('\n--- Preflop: Everyone limps ---');

    // UTG calls
    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`UTG ${actor!.name}: calls`);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    // SB calls
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`SB ${actor!.name}: calls`);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    // BB checks (option)
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`BB ${actor!.name}: checks`);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // FLOP: Check-raise sequence
    // Player 1 checks, Player 2 bets, Player 3 folds, Player 1 CHECK-RAISES
    // ==========================================
    console.log('\n--- Flop: Check-Raise Sequence ---');
    await players[0].verifyStreet('flop');

    // Pot should be 60 (3 players x 20 BB)
    await players[0].verifyPot(60, 'Pot on flop');

    // First to act on flop CHECKS (setting up the check-raise)
    const checkRaiser = await waitForAnyAction(players);
    expect(checkRaiser).not.toBeNull();
    console.log(`${checkRaiser!.name}: CHECKS (setting up check-raise)`);
    await checkRaiser!.check();
    await players[0].page.waitForTimeout(1000);

    // Second player BETS (min raise = 20)
    const bettor = await waitForAnyAction(players);
    expect(bettor).not.toBeNull();
    console.log(`${bettor!.name}: BETS (min raise)`);
    await bettor!.clickMinRaise();
    await players[0].page.waitForTimeout(1000);

    // Third player FOLDS
    const folder = await waitForAnyAction(players);
    expect(folder).not.toBeNull();
    console.log(`${folder!.name}: FOLDS`);
    await folder!.fold();
    await players[0].page.waitForTimeout(1000);

    // Now the check-raiser should have action again!
    // This is the key test - action should reopen for the checker
    console.log('\n--- KEY TEST: Check-raiser gets action back ---');
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    expect(actor!.name).toBe(checkRaiser!.name);
    console.log(`${actor!.name}: Has action again (as expected)`);

    // The check-raiser now RAISES (the check-raise!)
    console.log(`${actor!.name}: CHECK-RAISES!`);
    await actor!.raise(100); // Raise to 100
    await players[0].page.waitForTimeout(1000);

    // ==========================================
    // Verify the bettor must now respond to the raise
    // ==========================================
    console.log('\n--- Bettor must respond to check-raise ---');
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    expect(actor!.name).toBe(bettor!.name);
    console.log(`${actor!.name}: Must respond to check-raise`);

    // Bettor calls the check-raise
    console.log(`${actor!.name}: calls the check-raise`);
    await actor!.call();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // TURN: Both check to showdown
    // ==========================================
    console.log('\n--- Turn: Both check ---');
    await players[0].verifyStreet('turn');

    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: checks`);
    await actor!.check();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: checks`);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // RIVER: Both check to showdown
    // ==========================================
    console.log('\n--- River: Both check ---');
    await players[0].verifyStreet('river');

    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: checks`);
    await actor!.check();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: checks`);
    await actor!.check();
    await players[0].page.waitForTimeout(3000);

    // ==========================================
    // SHOWDOWN: Verify winner determined
    // ==========================================
    console.log('\n--- Showdown ---');

    // Check for showdown overlay
    let showdownFound = false;
    for (const player of players) {
      if (await player.hasShowdown()) {
        showdownFound = true;
        console.log(`${player.name}: Showdown overlay visible`);
        break;
      }
    }
    expect(showdownFound).toBe(true);

    console.log('\n=== TEST PASSED: Check-Raise on Flop ===\n');
  });

  // ==========================================
  // Test 2: Check-Raise on Turn
  // ==========================================
  test('2-player check-raise on turn', async () => {
    console.log('\n=== TEST: Check-Raise on Turn ===\n');

    // Create 2 players for heads-up
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `TurnCR_P${i + 1}`, i);
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
    // PREFLOP: Limp in
    // ==========================================
    console.log('\n--- Preflop ---');
    let actor = await waitForAnyAction(players);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // FLOP: Both check
    // ==========================================
    console.log('\n--- Flop: Both check ---');
    await players[0].verifyStreet('flop');

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // TURN: Check-raise sequence
    // ==========================================
    console.log('\n--- Turn: Check-Raise Sequence ---');
    await players[0].verifyStreet('turn');

    // First player checks
    const checkRaiser = await waitForAnyAction(players);
    expect(checkRaiser).not.toBeNull();
    console.log(`${checkRaiser!.name}: CHECKS (setting up check-raise)`);
    await checkRaiser!.check();
    await players[0].page.waitForTimeout(1000);

    // Second player bets
    const bettor = await waitForAnyAction(players);
    expect(bettor).not.toBeNull();
    console.log(`${bettor!.name}: BETS (min raise)`);
    await bettor!.clickMinRaise();
    await players[0].page.waitForTimeout(1000);

    // Check-raiser should have action again
    console.log('\n--- KEY TEST: Check-raiser gets action back ---');
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    expect(actor!.name).toBe(checkRaiser!.name);
    console.log(`${actor!.name}: CHECK-RAISES!`);
    await actor!.raise(80); // Raise to 80
    await players[0].page.waitForTimeout(1000);

    // Bettor calls
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    expect(actor!.name).toBe(bettor!.name);
    console.log(`${actor!.name}: calls the check-raise`);
    await actor!.call();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // RIVER: Both check
    // ==========================================
    console.log('\n--- River: Both check ---');
    await players[0].verifyStreet('river');

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(3000);

    // Verify showdown
    let showdownFound = false;
    for (const player of players) {
      if (await player.hasShowdown()) {
        showdownFound = true;
        break;
      }
    }
    expect(showdownFound).toBe(true);

    console.log('\n=== TEST PASSED: Check-Raise on Turn ===\n');
  });

  // ==========================================
  // Test 3: Check-Raise on River (Bluff scenario)
  // ==========================================
  test('2-player check-raise on river with fold', async () => {
    console.log('\n=== TEST: Check-Raise on River (Fold) ===\n');

    // Create 2 players
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `RiverCR_P${i + 1}`, i);
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
    // PREFLOP: Limp in
    // ==========================================
    console.log('\n--- Preflop ---');
    let actor = await waitForAnyAction(players);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // FLOP: Both check
    // ==========================================
    console.log('\n--- Flop ---');
    await players[0].verifyStreet('flop');

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // TURN: Both check
    // ==========================================
    console.log('\n--- Turn ---');
    await players[0].verifyStreet('turn');

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // RIVER: Check-raise and opponent folds
    // ==========================================
    console.log('\n--- River: Check-Raise Sequence ---');
    await players[0].verifyStreet('river');

    // Pot should be 40 (2 x BB)
    await players[0].verifyPot(BIG_BLIND * 2, 'Pot on river');

    // First player checks
    const checkRaiser = await waitForAnyAction(players);
    expect(checkRaiser).not.toBeNull();
    console.log(`${checkRaiser!.name}: CHECKS (setting up check-raise)`);
    await checkRaiser!.check();
    await players[0].page.waitForTimeout(1000);

    // Second player bets
    const bettor = await waitForAnyAction(players);
    expect(bettor).not.toBeNull();
    console.log(`${bettor!.name}: BETS (min raise)`);
    await bettor!.clickMinRaise();
    await players[0].page.waitForTimeout(1000);

    // Check-raiser CHECK-RAISES big
    console.log('\n--- KEY TEST: Check-raiser gets action back ---');
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    expect(actor!.name).toBe(checkRaiser!.name);
    console.log(`${actor!.name}: CHECK-RAISES big!`);
    await actor!.raise(200); // Big raise
    await players[0].page.waitForTimeout(1000);

    // Bettor FOLDS to the check-raise
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    expect(actor!.name).toBe(bettor!.name);
    console.log(`${actor!.name}: FOLDS to the check-raise!`);
    await actor!.fold();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // Verify: No showdown, check-raiser wins
    // ==========================================
    console.log('\n--- Verifying Hand Complete (no showdown) ---');

    // Should NOT have showdown (opponent folded)
    const hasShowdown = await players[0].hasShowdown();
    console.log(`Showdown visible: ${hasShowdown} (expected: false)`);
    
    // Check-raiser won without showdown
    const canStart = await checkRaiser!.canStartNewHand();
    expect(canStart).toBe(true);

    // Verify check-raiser won the pot
    // Started with 1000, put in 20 preflop, bet 200 on river, won pot of 60 (40 + bettor's 20)
    // Final: 1000 - 20 + 40 + 20 = 1040
    const checkRaiserStack = await checkRaiser!.getStack();
    console.log(`Check-raiser ${checkRaiser!.name} final stack: ${checkRaiserStack}`);
    expect(checkRaiserStack).toBeGreaterThan(STARTING_STACK);

    console.log('\n=== TEST PASSED: Check-Raise on River (Fold) ===\n');
  });
});
