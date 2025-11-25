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

  // ============ ACTION AVAILABILITY CHECKS ============

  /**
   * Check if Check button is available
   */
  async canCheck(): Promise<boolean> {
    const checkButton = this.page.locator('button:has-text("Check")');
    return await checkButton.isVisible().catch(() => false);
  }

  /**
   * Check if Raise button/input is available
   */
  async canRaise(): Promise<boolean> {
    const raiseInput = this.page.locator('input[aria-label="Raise Amount"]');
    return await raiseInput.isVisible().catch(() => false);
  }

  /**
   * Check if Call button is available
   */
  async canCall(): Promise<boolean> {
    const callButton = this.page.locator('button:has-text("Call")');
    return await callButton.isVisible().catch(() => false);
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
// TEST SUITE: Big Blind Option Tests
// ==========================================

test.describe('Big Blind Option Tests', () => {
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
  // Test 1: BB Option - Check after limps
  // ==========================================
  test('BB can check after all players limp (BB option)', async () => {
    console.log('\n=== TEST: BB Option - Check After Limps ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `BBOpt_P${i + 1}`, i);
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
    // PREFLOP: UTG calls, SB calls, BB has option
    // ==========================================
    console.log('\n--- Preflop: Players limp ---');

    // UTG calls
    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`UTG ${actor!.name}: calls (limp)`);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    // SB calls (completes)
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`SB ${actor!.name}: calls (completes)`);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    // ==========================================
    // KEY TEST: BB should have option to CHECK
    // ==========================================
    console.log('\n--- KEY TEST: BB has option to check ---');
    const bbPlayer = await waitForAnyAction(players);
    expect(bbPlayer).not.toBeNull();
    console.log(`BB ${bbPlayer!.name}: has action`);

    // Verify BB can CHECK (not just call or fold)
    const canCheck = await bbPlayer!.canCheck();
    const canRaise = await bbPlayer!.canRaise();
    
    console.log(`BB can CHECK: ${canCheck} (expected: true)`);
    console.log(`BB can RAISE: ${canRaise} (expected: true)`);
    
    expect(canCheck).toBe(true);
    expect(canRaise).toBe(true);

    // BB exercises option to CHECK
    console.log(`${bbPlayer!.name}: exercises BB option - CHECKS`);
    await bbPlayer!.check();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // Verify: We should now be on the FLOP
    // ==========================================
    console.log('\n--- Verifying Flop ---');
    await players[0].verifyStreet('flop');
    
    // Pot should be 60 (3 players x 20)
    await players[0].verifyPot(60, 'Pot on flop after BB checks');

    console.log('\n=== TEST PASSED: BB Option - Check After Limps ===\n');
  });

  // ==========================================
  // Test 2: BB Option - Raise after limps
  // ==========================================
  test('BB can raise after all players limp (BB option raise)', async () => {
    console.log('\n=== TEST: BB Option - Raise After Limps ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `BBRaise_P${i + 1}`, i);
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
    // PREFLOP: UTG calls, SB calls, BB raises
    // ==========================================
    console.log('\n--- Preflop: Players limp ---');

    // UTG calls
    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`UTG ${actor!.name}: calls (limp)`);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    // SB calls
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`SB ${actor!.name}: calls (completes)`);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    // ==========================================
    // KEY TEST: BB raises using their option
    // ==========================================
    console.log('\n--- KEY TEST: BB raises using option ---');
    const bbPlayer = await waitForAnyAction(players);
    expect(bbPlayer).not.toBeNull();
    console.log(`BB ${bbPlayer!.name}: has action`);

    // Verify BB can raise
    const canRaise = await bbPlayer!.canRaise();
    expect(canRaise).toBe(true);

    // BB raises to 60 (min-raise from 20 is 40, so raise to 40 or more)
    console.log(`${bbPlayer!.name}: exercises BB option - RAISES to 60`);
    await bbPlayer!.raise(60);
    await players[0].page.waitForTimeout(1000);

    // ==========================================
    // Other players must now respond to the raise
    // ==========================================
    console.log('\n--- Other players respond to BB raise ---');

    // UTG must respond
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    expect(actor!.name).not.toBe(bbPlayer!.name); // Should not be BB
    console.log(`${actor!.name}: faces BB raise - folds`);
    await actor!.fold();
    await players[0].page.waitForTimeout(1000);

    // SB must respond
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: faces BB raise - calls`);
    await actor!.call();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // Verify: Should be on flop now
    // ==========================================
    console.log('\n--- Verifying Flop ---');
    await players[0].verifyStreet('flop');

    // Pot should be: UTG(20) + SB(60) + BB(60) = 140
    // Actually UTG folded after putting in 20, so:
    // UTG(20) + SB(60) + BB(60) = 140
    await players[0].verifyPot(140, 'Pot on flop after BB raise');

    console.log('\n=== TEST PASSED: BB Option - Raise After Limps ===\n');
  });

  // ==========================================
  // Test 3: BB Option - No option after raise
  // ==========================================
  test('BB has no option after someone raises (must call/fold/re-raise)', async () => {
    console.log('\n=== TEST: BB No Option After Raise ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `BBNoOpt_P${i + 1}`, i);
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
    // PREFLOP: UTG raises, SB folds, BB must respond (no check option)
    // ==========================================
    console.log('\n--- Preflop: UTG raises ---');

    // UTG raises
    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`UTG ${actor!.name}: raises to 60`);
    await actor!.raise(60);
    await players[0].page.waitForTimeout(1000);

    // SB folds
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`SB ${actor!.name}: folds`);
    await actor!.fold();
    await players[0].page.waitForTimeout(1000);

    // ==========================================
    // KEY TEST: BB should NOT be able to check (must call/fold/raise)
    // ==========================================
    console.log('\n--- KEY TEST: BB cannot check after raise ---');
    const bbPlayer = await waitForAnyAction(players);
    expect(bbPlayer).not.toBeNull();
    console.log(`BB ${bbPlayer!.name}: has action`);

    // Verify BB canNOT check
    const canCheck = await bbPlayer!.canCheck();
    const canCall = await bbPlayer!.canCall();
    const canRaise = await bbPlayer!.canRaise();

    console.log(`BB can CHECK: ${canCheck} (expected: false)`);
    console.log(`BB can CALL: ${canCall} (expected: true)`);
    console.log(`BB can RAISE: ${canRaise} (expected: true)`);

    expect(canCheck).toBe(false);
    expect(canCall).toBe(true);
    expect(canRaise).toBe(true);

    // BB calls
    console.log(`${bbPlayer!.name}: calls`);
    await bbPlayer!.call();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // Verify: Should be on flop (heads-up after SB folded)
    // ==========================================
    console.log('\n--- Verifying Flop ---');
    await players[0].verifyStreet('flop');

    // Pot should be: SB(10) + BB(60) + UTG(60) = 130
    await players[0].verifyPot(130, 'Pot on flop heads-up');

    console.log('\n=== TEST PASSED: BB No Option After Raise ===\n');
  });
});
