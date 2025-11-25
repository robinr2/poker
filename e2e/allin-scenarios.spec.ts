import { test, expect, Browser, BrowserContext, Page } from '@playwright/test';
import { resetTable, setDeck, buildDeck } from './helpers/deterministic-helpers';
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
   * Get showdown result text
   */
  async getShowdownResult(): Promise<string> {
    const showdownOverlay = this.page.locator('.showdown-overlay');
    await expect(showdownOverlay).toBeVisible({ timeout: 15000 });
    const text = await showdownOverlay.textContent();
    return text || '';
  }

  /**
   * Check if Start Hand button is visible (indicates hand is complete)
   */
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
// TEST SUITE: All-In Scenarios
// ==========================================

test.describe('All-In Scenarios', () => {
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
  // Test 1: All-In When Facing a Raise
  // P1 raises to 100, P2 re-raises to 300, P3 goes all-in (1000)
  // P1 folds, P2 calls -> Showdown
  // ==========================================
  test('3-player all-in facing raise with fold and call', async () => {
    console.log('\n=== TEST: All-In Facing Raise ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `AllIn_P${i + 1}`, i);
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

    // Set deterministic deck so we know who wins
    // P3 (seat 2) will have pocket Aces and win
    console.log('\n--- Setting Deterministic Deck ---');
    const deck = buildDeck({
      holeCards: {
        0: ['2s', '7c'], // P1: Garbage
        1: ['Ks', 'Kh'], // P2: Pocket Kings
        2: ['As', 'Ah'], // P3: Pocket Aces (winner)
      },
      flop: ['3d', '5c', '9h'],
      turn: '2d',
      river: 'Jc',
    });
    await setDeck(TABLE_ID, deck);

    // Record initial stacks
    console.log('\n--- Initial Stacks ---');
    for (const player of players) {
      await player.verifyStack(STARTING_STACK);
    }

    // Start hand
    console.log('\n--- Starting Hand ---');
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // Verify pot has blinds
    await players[0].verifyPot(SMALL_BLIND + BIG_BLIND, 'Pot after blinds');

    // === PREFLOP ACTION ===
    console.log('\n--- Preflop Action ---');

    // First actor raises to 100
    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    const firstRaiser = actor!;
    console.log(`${firstRaiser.name}: Raises to 100`);
    await firstRaiser.raise(100);
    await players[0].page.waitForTimeout(1000);

    // Verify pot after first raise
    // Pot = blinds (30) + raise (100) = 130
    await players[0].verifyPot(130, 'Pot after first raise to 100');

    // Second actor re-raises to 300
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    const reRaiser = actor!;
    console.log(`${reRaiser.name}: Re-raises to 300`);
    await reRaiser.raise(300);
    await players[0].page.waitForTimeout(1000);

    // Verify pot after re-raise
    // Pot should include blinds + raise + re-raise amount (may vary based on blind positions)
    const potAfterReRaise = await players[0].getPot();
    console.log(`Pot after re-raise to 300: ${potAfterReRaise}`);
    expect(potAfterReRaise).toBeGreaterThanOrEqual(400);

    // Third actor goes all-in (1000)
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    const allInPlayer = actor!;
    console.log(`${allInPlayer.name}: Goes ALL-IN (1000)`);
    await allInPlayer.clickAllIn();
    await players[0].page.waitForTimeout(1000);

    // Verify pot after all-in
    // Need to account for what allInPlayer already put in
    // If allInPlayer was BB (20), they add 980 more = 1000 total
    // Pot should be: previous (430) + allInPlayer's contribution
    const potAfterAllIn = await players[0].getPot();
    console.log(`Pot after all-in: ${potAfterAllIn}`);
    expect(potAfterAllIn).toBeGreaterThanOrEqual(1000);

    // First raiser folds
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: Folds (losing their 100 raise)`);
    await actor!.fold();
    await players[0].page.waitForTimeout(1000);

    // Re-raiser calls the all-in
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: Calls all-in`);
    await actor!.call();
    await players[0].page.waitForTimeout(3000);

    // === VERIFY SHOWDOWN ===
    console.log('\n--- Verifying Showdown ---');
    
    // Wait for showdown
    const hasShowdown = await players[0].hasShowdown();
    expect(hasShowdown).toBe(true);

    const showdownResult = await players[0].getShowdownResult();
    console.log(`Showdown result: ${showdownResult}`);

    // AllIn_P3 should win (has pocket Aces)
    // Check the showdown mentions the winner
    expect(showdownResult).toContain('AllIn_P3');

    // === VERIFY FINAL STACKS ===
    console.log('\n--- Final Stacks ---');
    
    // Calculate expected stacks:
    // - First raiser folded after putting in 100: 1000 - 100 = 900
    // - Re-raiser called all-in (1000) and lost: 1000 - 1000 = 0
    // - All-in player won: 1000 + 100 (dead money) + 1000 (call) = 2100
    //   But wait, we need to track what each player put in:
    //   - First raiser: 100 (then folded) = dead money
    //   - Re-raiser: 1000 total (called the all-in)
    //   - All-in player: 1000 (all-in)
    //   Total pot: 100 + 1000 + 1000 = 2100
    //   Winner gets 2100, so their stack = 0 + 2100 = 2100
    //   (Actually winner started with 1000, put in 1000, won 2100, so 1000 - 1000 + 2100 = 2100)

    // Wait for stacks to update
    await players[0].page.waitForTimeout(2000);

    // Check that all-in winner has chips
    const allInPlayerStack = await allInPlayer.getStack();
    console.log(`All-in winner (${allInPlayer.name}) stack: ${allInPlayerStack}`);
    
    // Winner should have more than starting stack
    expect(allInPlayerStack).toBeGreaterThan(STARTING_STACK);

    // Hand should be complete
    const canStart = await players[0].canStartNewHand();
    expect(canStart).toBe(true);

    console.log('\n=== TEST PASSED: All-In Facing Raise ===\n');
  });

  // ==========================================
  // Test 2: All-In with Mixed Response (4 players)
  // P1 goes all-in, P2 calls, P3 folds, P4 calls -> Showdown
  // ==========================================
  test('4-player all-in with some callers, some folders', async () => {
    console.log('\n=== TEST: 4-Player All-In Mixed Response ===\n');

    // Create 4 players
    for (let i = 0; i < 4; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `Mix_P${i + 1}`, i);
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

    // Set deterministic deck - P2 wins with flush
    console.log('\n--- Setting Deterministic Deck ---');
    const deck = buildDeck({
      holeCards: {
        0: ['Ks', 'Kh'], // P1: Pocket Kings
        1: ['Ah', '2h'], // P2: Ace-high hearts (will make flush)
        2: ['7c', '8c'], // P3: Will fold
        3: ['Qs', 'Qh'], // P4: Pocket Queens
      },
      flop: ['5h', '9h', 'Th'], // P2 has flush
      turn: '3d',
      river: '4c',
    });
    await setDeck(TABLE_ID, deck);

    // Start hand
    console.log('\n--- Starting Hand ---');
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // Verify initial pot
    await players[0].verifyPot(SMALL_BLIND + BIG_BLIND, 'Pot after blinds');

    // === PREFLOP ACTION ===
    console.log('\n--- Preflop: All-In and Responses ---');

    // First actor goes all-in
    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    const allInPlayer = actor!;
    console.log(`${allInPlayer.name}: Goes ALL-IN`);
    await allInPlayer.clickAllIn();
    await players[0].page.waitForTimeout(1000);

    // Second player calls
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    const caller1 = actor!;
    console.log(`${caller1.name}: Calls all-in`);
    await caller1.call();
    await players[0].page.waitForTimeout(1000);

    // Third player folds
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    const folder = actor!;
    console.log(`${folder.name}: Folds`);
    await folder.fold();
    await players[0].page.waitForTimeout(1000);

    // Fourth player calls
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    const caller2 = actor!;
    console.log(`${caller2.name}: Calls all-in`);
    await caller2.call();
    await players[0].page.waitForTimeout(3000);

    // === VERIFY SHOWDOWN ===
    console.log('\n--- Verifying Showdown ---');
    
    const hasShowdown = await players[0].hasShowdown();
    expect(hasShowdown).toBe(true);

    const showdownResult = await players[0].getShowdownResult();
    console.log(`Showdown result: ${showdownResult}`);

    // Verify there's a winner declared (the specific winner depends on seat positions)
    expect(showdownResult.toLowerCase()).toMatch(/winners?:/i);

    // === VERIFY FINAL STACKS ===
    console.log('\n--- Final Stacks ---');
    
    await players[0].page.waitForTimeout(2000);

    // The winner should have all the chips from 3 players (3 x 1000 = 3000)
    // Plus their own 1000 that comes back = total 3000 chips
    // Actually: winner started with 1000, others each put in 1000
    // Folder only lost their blind
    
    let totalChips = 0;
    for (const player of players) {
      const stack = await player.getStack();
      console.log(`${player.name}: ${stack}`);
      totalChips += stack;
    }
    
    // Total chips should equal 4 x starting stack
    console.log(`Total chips: ${totalChips} (expected: ${4 * STARTING_STACK})`);
    expect(totalChips).toBe(4 * STARTING_STACK);

    // Hand should be complete
    const canStart = await players[0].canStartNewHand();
    expect(canStart).toBe(true);

    console.log('\n=== TEST PASSED: 4-Player All-In Mixed Response ===\n');
  });

  // ==========================================
  // Test 3: Heads-up All-In (2 players both all-in)
  // ==========================================
  test('2-player mutual all-in showdown', async () => {
    console.log('\n=== TEST: Heads-up All-In ===\n');

    // Create 2 players
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `HU_P${i + 1}`, i);
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

    // Set deterministic deck - P1 wins with straight
    console.log('\n--- Setting Deterministic Deck ---');
    const deck = buildDeck({
      holeCards: {
        0: ['Ts', 'Jh'], // P1: TJ (will make straight)
        1: ['As', '2c'], // P2: A2 (pair of aces)
      },
      flop: ['8d', '9c', 'Qh'], // P1 has open-ended
      turn: '7s', // P1 makes straight
      river: 'Ac', // P2 makes pair of aces (but straight beats pair)
    });
    await setDeck(TABLE_ID, deck);

    // Record initial stacks
    console.log('\n--- Initial Stacks ---');
    for (const player of players) {
      await player.verifyStack(STARTING_STACK);
    }

    // Start hand
    console.log('\n--- Starting Hand ---');
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // In heads-up, SB is dealer and acts first preflop
    // Verify pot has blinds
    await players[0].verifyPot(SMALL_BLIND + BIG_BLIND, 'Pot after blinds');

    // === PREFLOP ACTION ===
    console.log('\n--- Preflop: Both All-In ---');

    // First actor (SB/Dealer) goes all-in
    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: Goes ALL-IN`);
    await actor!.clickAllIn();
    await players[0].page.waitForTimeout(1000);

    // Verify pot increased
    const potAfterFirstAllIn = await players[0].getPot();
    console.log(`Pot after first all-in: ${potAfterFirstAllIn}`);

    // Second player calls
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: Calls all-in`);
    await actor!.call();
    await players[0].page.waitForTimeout(3000);

    // === VERIFY SHOWDOWN ===
    console.log('\n--- Verifying Showdown ---');
    
    const hasShowdown = await players[0].hasShowdown();
    expect(hasShowdown).toBe(true);

    const showdownResult = await players[0].getShowdownResult();
    console.log(`Showdown result: ${showdownResult}`);

    // P1 should win with straight
    expect(showdownResult).toContain('HU_P1');
    expect(showdownResult.toLowerCase()).toContain('straight');

    // === VERIFY FINAL STACKS ===
    console.log('\n--- Final Stacks ---');
    
    await players[0].page.waitForTimeout(2000);

    // Winner gets all 2000 chips
    const p1Stack = await players[0].getStack();
    const p2Stack = await players[1].getStack();
    
    console.log(`HU_P1: ${p1Stack}`);
    console.log(`HU_P2: ${p2Stack}`);
    
    // Total should still be 2000
    expect(p1Stack + p2Stack).toBe(2 * STARTING_STACK);
    
    // Winner should have 2000, loser should have 0
    expect(p1Stack).toBe(2 * STARTING_STACK);
    expect(p2Stack).toBe(0);

    // Hand should be complete - we already verified showdown above
    // Note: When one player busts out, we may not be able to start a new hand
    // (need 2+ players), so we just verify the hand completed via showdown
    const canStart = await players[0].canStartNewHand();
    console.log(`Showdown visible: ${hasShowdown}, Can start new hand: ${canStart}`);
    
    // The hand should have completed (showdown happened, which we already verified above)
    // We don't require start hand button since one player is busted

    console.log('\n=== TEST PASSED: Heads-up All-In ===\n');
  });
});
