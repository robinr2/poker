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
 * Player class with comprehensive verification methods
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

  // ============ VERIFICATION METHODS ============

  /**
   * Get the current pot amount displayed
   */
  async getPot(): Promise<number> {
    const potText = await this.page.locator('.pot-display').textContent().catch(() => null);
    if (!potText) return 0;
    const match = potText.match(/Pot:\s*(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }

  /**
   * Verify pot amount equals expected value
   */
  async verifyPot(expected: number, message?: string) {
    const actual = await this.getPot();
    const msg = message || `${this.name}: Pot should be ${expected}`;
    console.log(`${msg} - Actual: ${actual}`);
    expect(actual, msg).toBe(expected);
  }

  /**
   * Get this player's current stack
   */
  async getStack(): Promise<number> {
    const stackText = await this.page.locator('.own-seat .stack').textContent().catch(() => null);
    if (!stackText) return 0;
    const match = stackText.match(/(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }

  /**
   * Get a specific seat's stack by seat index
   */
  async getSeatStack(seatIndex: number): Promise<number> {
    const seats = this.page.locator('.seat');
    const seatCount = await seats.count();
    
    for (let i = 0; i < seatCount; i++) {
      const seat = seats.nth(i);
      const seatNumText = await seat.locator('.seat-number').textContent();
      if (seatNumText?.includes(`Seat ${seatIndex}`)) {
        const stackText = await seat.locator('.stack').textContent().catch(() => null);
        if (!stackText) return 0;
        const match = stackText.match(/(\d+)/);
        return match ? parseInt(match[1], 10) : 0;
      }
    }
    return 0;
  }

  /**
   * Verify this player's stack equals expected value
   */
  async verifyStack(expected: number, message?: string) {
    const actual = await this.getStack();
    const msg = message || `${this.name}: Stack should be ${expected}`;
    console.log(`${msg} - Actual: ${actual}`);
    expect(actual, msg).toBe(expected);
  }

  /**
   * Get this player's current street bet (amount in front of player)
   */
  async getStreetBet(): Promise<number> {
    const betText = await this.page.locator('.own-seat .bet-amount').textContent().catch(() => null);
    if (!betText) return 0;
    const match = betText.match(/\$?\s*(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }

  /**
   * Get a specific seat's street bet by seat index
   */
  async getSeatStreetBet(seatIndex: number): Promise<number> {
    const seats = this.page.locator('.seat');
    const seatCount = await seats.count();
    
    for (let i = 0; i < seatCount; i++) {
      const seat = seats.nth(i);
      const seatNumText = await seat.locator('.seat-number').textContent();
      if (seatNumText?.includes(`Seat ${seatIndex}`)) {
        const betText = await seat.locator('.bet-amount').textContent().catch(() => null);
        if (!betText) return 0;
        const match = betText.match(/\$?\s*(\d+)/);
        return match ? parseInt(match[1], 10) : 0;
      }
    }
    return 0;
  }

  /**
   * Verify this player's street bet equals expected value
   */
  async verifyStreetBet(expected: number, message?: string) {
    const actual = await this.getStreetBet();
    const msg = message || `${this.name}: Street bet should be ${expected}`;
    console.log(`${msg} - Actual: ${actual}`);
    expect(actual, msg).toBe(expected);
  }

  /**
   * Get current street name
   */
  async getStreet(): Promise<string> {
    const streetText = await this.page.locator('.street-indicator').textContent().catch(() => null);
    return streetText?.toLowerCase() || '';
  }

  /**
   * Verify current street
   */
  async verifyStreet(expected: string, message?: string) {
    const actual = await this.getStreet();
    const msg = message || `${this.name}: Street should be ${expected}`;
    console.log(`${msg} - Actual: ${actual}`);
    expect(actual.toLowerCase(), msg).toBe(expected.toLowerCase());
  }

  /**
   * Check if showdown overlay is visible
   */
  async hasShowdown(): Promise<boolean> {
    return await this.page.locator('.showdown-overlay').isVisible().catch(() => false);
  }

  /**
   * Check if Start Hand button is visible (indicates hand is complete)
   */
  async canStartNewHand(): Promise<boolean> {
    return await this.page.locator('button:has-text("Start Hand")').isVisible().catch(() => false);
  }

  /**
   * Check if player has cards
   */
  async hasCards(): Promise<boolean> {
    const bodyText = await this.page.locator('body').textContent();
    return (bodyText?.includes('♠') || bodyText?.includes('♥') ||
            bodyText?.includes('♦') || bodyText?.includes('♣')) ?? false;
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
// TEST SUITE: Fold Scenarios
// ==========================================

test.describe('Fold Scenarios', () => {
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
  // Test 1: Sequential Folds (4 players, all fold to raiser)
  // ==========================================
  test('4-player sequential folds - raiser wins blinds', async () => {
    console.log('\n=== TEST: 4-Player Sequential Folds ===\n');

    // Create 4 players
    for (let i = 0; i < 4; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `Player${i + 1}`, i);
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

    // Find first actor and raise
    console.log('\n--- Preflop Action ---');
    const firstActor = await waitForAnyAction(players);
    expect(firstActor).not.toBeNull();
    console.log(`First to act: ${firstActor!.name}`);
    
    // First actor raises to 100
    await firstActor!.raise(100);
    
    // Verify pot increased
    await players[0].page.waitForTimeout(500);
    const potAfterRaise = await players[0].getPot();
    console.log(`Pot after raise: ${potAfterRaise}`);
    
    // Others fold sequentially
    for (let i = 0; i < 3; i++) {
      const actor = await waitForAnyAction(players);
      if (!actor) break;
      console.log(`${actor.name} folding...`);
      await actor.fold();
      await players[0].page.waitForTimeout(500);
    }

    // Wait for hand to complete
    await players[0].page.waitForTimeout(2000);

    // Verify hand is complete
    console.log('\n--- Verifying Hand Complete ---');
    const canStart = await players[0].canStartNewHand();
    expect(canStart).toBe(true);

    // Verify final stacks
    console.log('\n--- Final Stacks ---');
    // The raiser should have won the blinds (30 chips = SB + BB)
    const raiserStack = await firstActor!.getStack();
    console.log(`Raiser ${firstActor!.name}: ${raiserStack} (expected: ${STARTING_STACK + SMALL_BLIND + BIG_BLIND})`);
    
    // Raiser wins blinds (they didn't have to put in more since everyone folded)
    // Initial 1000 + SB(10) + BB(20) = 1030
    expect(raiserStack).toBe(STARTING_STACK + SMALL_BLIND + BIG_BLIND);

    console.log('\n=== TEST PASSED: 4-Player Sequential Folds ===\n');
  });

  // ==========================================
  // Test 2: All fold to Big Blind
  // ==========================================
  test('All fold to Big Blind - BB wins SB uncontested', async () => {
    console.log('\n=== TEST: All Fold to Big Blind ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `FoldBB_P${i + 1}`, i);
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

    // Verify pot has blinds
    await players[0].verifyPot(SMALL_BLIND + BIG_BLIND, 'Pot after blinds');

    // UTG (first to act) folds
    console.log('\n--- Preflop: Everyone folds to BB ---');
    const utg = await waitForAnyAction(players);
    expect(utg).not.toBeNull();
    console.log(`UTG: ${utg!.name} folds`);
    await utg!.fold();

    await players[0].page.waitForTimeout(1000);

    // SB folds
    const sb = await waitForAnyAction(players);
    expect(sb).not.toBeNull();
    console.log(`SB: ${sb!.name} folds`);
    await sb!.fold();

    // Wait for hand to complete
    await players[0].page.waitForTimeout(2000);

    // Verify hand is complete
    console.log('\n--- Verifying Hand Complete ---');
    const canStart = await players[0].canStartNewHand();
    expect(canStart).toBe(true);

    // Find who was BB (the one who didn't fold and won)
    console.log('\n--- Final Stacks ---');
    let bbPlayer: Player | null = null;
    let sbPlayer: Player | null = null;
    
    for (const player of players) {
      const stack = await player.getStack();
      console.log(`${player.name}: ${stack}`);
      
      if (stack === STARTING_STACK + SMALL_BLIND) {
        // BB won the SB
        bbPlayer = player;
      } else if (stack === STARTING_STACK - SMALL_BLIND) {
        // SB lost their blind
        sbPlayer = player;
      }
    }

    expect(bbPlayer).not.toBeNull();
    console.log(`\nBB (${bbPlayer!.name}) won ${SMALL_BLIND} chips (the SB)`);

    console.log('\n=== TEST PASSED: All Fold to Big Blind ===\n');
  });

  // ==========================================
  // Test 3: Fold on Flop
  // ==========================================
  test('3-player fold on flop - bettor wins', async () => {
    console.log('\n=== TEST: Fold on Flop ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `FlopFold_P${i + 1}`, i);
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

    // Preflop: Everyone calls to see flop
    console.log('\n--- Preflop: Everyone calls ---');
    
    // UTG calls
    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: calls`);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    // SB calls
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: calls`);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    // BB checks
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: checks`);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // Verify we're on the flop
    await players[0].verifyStreet('flop');
    
    // Pot should be 60 (3 players x 20 BB)
    await players[0].verifyPot(60, 'Pot on flop');

    // Flop: First player bets, others fold
    console.log('\n--- Flop: Bet and folds ---');
    
    // First to act on flop bets (min raise)
    const flopBettor = await waitForAnyAction(players);
    expect(flopBettor).not.toBeNull();
    console.log(`${flopBettor!.name}: bets (min raise)`);
    await flopBettor!.clickMinRaise();
    await players[0].page.waitForTimeout(1000);

    // Record bettor's stack after betting
    const bettorStackAfterBet = await flopBettor!.getStack();
    console.log(`Bettor stack after bet: ${bettorStackAfterBet}`);

    // Others fold
    for (let i = 0; i < 2; i++) {
      actor = await waitForAnyAction(players);
      if (!actor) break;
      console.log(`${actor.name}: folds`);
      await actor.fold();
      await players[0].page.waitForTimeout(1000);
    }

    // Wait for hand to complete
    await players[0].page.waitForTimeout(2000);

    // Verify hand is complete
    console.log('\n--- Verifying Hand Complete ---');
    const canStart = await players[0].canStartNewHand();
    expect(canStart).toBe(true);

    // Verify bettor won the pot
    console.log('\n--- Final Stacks ---');
    const bettorFinalStack = await flopBettor!.getStack();
    console.log(`Bettor ${flopBettor!.name} final stack: ${bettorFinalStack}`);
    
    // Bettor should have won: started with 1000, put in 20 preflop + bet on flop, won pot of 60 + bet
    // Since they won without showdown, they get their bet back + pot
    expect(bettorFinalStack).toBeGreaterThan(STARTING_STACK - BIG_BLIND);

    console.log('\n=== TEST PASSED: Fold on Flop ===\n');
  });

  // ==========================================
  // Test 4: Fold on Turn
  // ==========================================
  test('2-player fold on turn - bettor wins', async () => {
    console.log('\n=== TEST: Fold on Turn ===\n');

    // Create 3 players (one will fold on flop to get heads-up)
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `TurnFold_P${i + 1}`, i);
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

    // Preflop: UTG calls, SB calls, BB checks
    console.log('\n--- Preflop ---');
    let actor = await waitForAnyAction(players);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // Flop: Check, Check, Fold (to get to turn heads-up)
    console.log('\n--- Flop: One player folds ---');
    await players[0].verifyStreet('flop');

    actor = await waitForAnyAction(players);
    console.log(`${actor!.name}: checks`);
    await actor!.check();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    console.log(`${actor!.name}: bets (min)`);
    await actor!.clickMinRaise();
    await players[0].page.waitForTimeout(1000);

    // Track who bet on flop
    const flopBettor = actor;

    actor = await waitForAnyAction(players);
    console.log(`${actor!.name}: folds`);
    await actor!.fold();
    await players[0].page.waitForTimeout(1000);

    // First checker calls
    actor = await waitForAnyAction(players);
    console.log(`${actor!.name}: calls`);
    await actor!.call();
    await players[0].page.waitForTimeout(2000);

    // Turn: One player bets, other folds
    console.log('\n--- Turn: Bet and fold ---');
    await players[0].verifyStreet('turn');

    const potOnTurn = await players[0].getPot();
    console.log(`Pot on turn: ${potOnTurn}`);

    actor = await waitForAnyAction(players);
    const turnBettor = actor;
    console.log(`${actor!.name}: bets (min)`);
    await actor!.clickMinRaise();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    console.log(`${actor!.name}: folds`);
    await actor!.fold();
    await players[0].page.waitForTimeout(2000);

    // Verify hand is complete
    console.log('\n--- Verifying Hand Complete ---');
    const canStart = await players[0].canStartNewHand();
    expect(canStart).toBe(true);

    // Verify bettor won
    console.log('\n--- Final Stacks ---');
    const bettorFinalStack = await turnBettor!.getStack();
    console.log(`Turn bettor ${turnBettor!.name} final stack: ${bettorFinalStack}`);
    expect(bettorFinalStack).toBeGreaterThan(STARTING_STACK - BIG_BLIND * 3); // Account for multiple bets

    console.log('\n=== TEST PASSED: Fold on Turn ===\n');
  });

  // ==========================================
  // Test 5: Fold on River
  // ==========================================
  test('2-player fold on river - bettor wins', async () => {
    console.log('\n=== TEST: Fold on River ===\n');

    // Create 2 players for heads-up
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `RiverFold_P${i + 1}`, i);
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

    // Preflop: Call, Check
    console.log('\n--- Preflop ---');
    let actor = await waitForAnyAction(players);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // Flop: Check, Check
    console.log('\n--- Flop: Both check ---');
    await players[0].verifyStreet('flop');

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // Turn: Check, Check
    console.log('\n--- Turn: Both check ---');
    await players[0].verifyStreet('turn');

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // River: Bet, Fold
    console.log('\n--- River: Bet and fold ---');
    await players[0].verifyStreet('river');

    const potOnRiver = await players[0].getPot();
    console.log(`Pot on river: ${potOnRiver}`);
    expect(potOnRiver).toBe(BIG_BLIND * 2); // 40 chips (2 x BB)

    actor = await waitForAnyAction(players);
    const riverBettor = actor;
    console.log(`${actor!.name}: bets (min)`);
    await actor!.clickMinRaise();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    console.log(`${actor!.name}: folds`);
    await actor!.fold();
    await players[0].page.waitForTimeout(2000);

    // Verify hand is complete
    console.log('\n--- Verifying Hand Complete ---');
    const canStart = await players[0].canStartNewHand();
    expect(canStart).toBe(true);

    // Verify river bettor won
    console.log('\n--- Final Stacks ---');
    const bettorFinalStack = await riverBettor!.getStack();
    console.log(`River bettor ${riverBettor!.name} final stack: ${bettorFinalStack}`);
    
    // Bettor put in 20 preflop + min bet on river, won pot of 40
    // Should have: 1000 - 20 + 40 = 1020
    expect(bettorFinalStack).toBe(STARTING_STACK + BIG_BLIND);

    console.log('\n=== TEST PASSED: Fold on River ===\n');
  });
});
