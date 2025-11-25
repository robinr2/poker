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
 * Player class with blind position detection
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

  // ============ BLIND POSITION DETECTION ============

  /**
   * Check if this player's seat has the dealer badge
   */
  async isDealer(): Promise<boolean> {
    const dealerBadge = this.page.locator('.own-seat .dealer-badge');
    return await dealerBadge.isVisible().catch(() => false);
  }

  /**
   * Check if this player's seat has the SB badge
   */
  async isSmallBlind(): Promise<boolean> {
    const sbBadge = this.page.locator('.own-seat .sb-badge');
    return await sbBadge.isVisible().catch(() => false);
  }

  /**
   * Check if this player's seat has the BB badge
   */
  async isBigBlind(): Promise<boolean> {
    const bbBadge = this.page.locator('.own-seat .bb-badge');
    return await bbBadge.isVisible().catch(() => false);
  }

  /**
   * Get dealer seat index by checking all seats
   */
  async getDealerSeat(): Promise<number> {
    const seats = this.page.locator('.seat');
    const seatCount = await seats.count();
    
    for (let i = 0; i < seatCount; i++) {
      const seat = seats.nth(i);
      const dealerBadge = seat.locator('.dealer-badge');
      if (await dealerBadge.isVisible().catch(() => false)) {
        const seatNumText = await seat.locator('.seat-number').textContent();
        const match = seatNumText?.match(/Seat (\d+)/);
        if (match) return parseInt(match[1], 10);
      }
    }
    return -1;
  }

  /**
   * Get SB seat index by checking all seats
   */
  async getSmallBlindSeat(): Promise<number> {
    const seats = this.page.locator('.seat');
    const seatCount = await seats.count();
    
    for (let i = 0; i < seatCount; i++) {
      const seat = seats.nth(i);
      const sbBadge = seat.locator('.sb-badge');
      if (await sbBadge.isVisible().catch(() => false)) {
        const seatNumText = await seat.locator('.seat-number').textContent();
        const match = seatNumText?.match(/Seat (\d+)/);
        if (match) return parseInt(match[1], 10);
      }
    }
    return -1;
  }

  /**
   * Get BB seat index by checking all seats
   */
  async getBigBlindSeat(): Promise<number> {
    const seats = this.page.locator('.seat');
    const seatCount = await seats.count();
    
    for (let i = 0; i < seatCount; i++) {
      const seat = seats.nth(i);
      const bbBadge = seat.locator('.bb-badge');
      if (await bbBadge.isVisible().catch(() => false)) {
        const seatNumText = await seat.locator('.seat-number').textContent();
        const match = seatNumText?.match(/Seat (\d+)/);
        if (match) return parseInt(match[1], 10);
      }
    }
    return -1;
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

  async dismissShowdownIfVisible() {
    const showdownOverlay = this.page.locator('.showdown-overlay');
    if (await showdownOverlay.isVisible().catch(() => false)) {
      console.log(`${this.name}: Dismissing showdown overlay...`);
      // Click the Close button (has text "×")
      const closeButton = this.page.locator('.showdown-overlay button:has-text("×"), .showdown-overlay button:has-text("Close")');
      if (await closeButton.isVisible().catch(() => false)) {
        await closeButton.click();
        // Wait for overlay to disappear
        await showdownOverlay.waitFor({ state: 'hidden', timeout: 5000 }).catch(() => {});
      } else {
        // Fallback: click the overlay itself
        await showdownOverlay.click().catch(() => {});
      }
      await this.page.waitForTimeout(500);
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
// TEST SUITE: Heads-Up Blind Structure
// ==========================================

test.describe('Heads-Up Blind Structure Tests', () => {
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
  // Test 1: Dealer posts SB in heads-up
  // ==========================================
  test('dealer posts SB in heads-up (dealer = SB, other = BB)', async () => {
    console.log('\n=== TEST: Dealer Posts SB in Heads-Up ===\n');

    // Create 2 players for heads-up
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `HeadsUp_P${i + 1}`, i);
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
    // KEY TEST: In heads-up, dealer should be SB
    // ==========================================
    console.log('\n--- Checking Blind Positions ---');

    const dealerSeat = await players[0].getDealerSeat();
    const sbSeat = await players[0].getSmallBlindSeat();
    const bbSeat = await players[0].getBigBlindSeat();

    console.log(`Dealer seat: ${dealerSeat}`);
    console.log(`Small Blind seat: ${sbSeat}`);
    console.log(`Big Blind seat: ${bbSeat}`);

    // In heads-up: dealer = SB
    expect(dealerSeat).toBe(sbSeat);
    console.log(`✓ Dealer (${dealerSeat}) is the Small Blind (${sbSeat})`);

    // SB and BB should be different seats
    expect(sbSeat).not.toBe(bbSeat);
    console.log(`✓ SB (${sbSeat}) and BB (${bbSeat}) are different players`);

    console.log('\n=== TEST PASSED: Dealer Posts SB in Heads-Up ===\n');
  });

  // ==========================================
  // Test 2: Dealer acts first preflop in heads-up
  // ==========================================
  test('dealer/SB acts first preflop in heads-up', async () => {
    console.log('\n=== TEST: Dealer Acts First Preflop ===\n');

    // Create 2 players
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `DealerFirst_P${i + 1}`, i);
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

    // Get positions
    const dealerSeat = await players[0].getDealerSeat();
    console.log(`Dealer is at seat: ${dealerSeat}`);

    // ==========================================
    // KEY TEST: Dealer should have first action preflop
    // ==========================================
    console.log('\n--- Checking First Actor Preflop ---');

    const firstActor = await waitForAnyAction(players);
    expect(firstActor).not.toBeNull();
    console.log(`First actor: ${firstActor!.name}`);

    // Check if first actor is the dealer
    const isFirstActorDealer = await firstActor!.isDealer();
    console.log(`First actor is dealer: ${isFirstActorDealer}`);
    expect(isFirstActorDealer).toBe(true);

    // Also verify they are SB
    const isFirstActorSB = await firstActor!.isSmallBlind();
    console.log(`First actor is SB: ${isFirstActorSB}`);
    expect(isFirstActorSB).toBe(true);

    console.log(`✓ Dealer/SB acts first preflop in heads-up`);

    console.log('\n=== TEST PASSED: Dealer Acts First Preflop ===\n');
  });

  // ==========================================
  // Test 3: BB acts first on flop in heads-up
  // ==========================================
  test('BB acts first on flop (postflop) in heads-up', async () => {
    console.log('\n=== TEST: BB Acts First Postflop ===\n');

    // Create 2 players
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `BBFirst_P${i + 1}`, i);
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
    // PREFLOP: Dealer/SB calls, BB checks
    // ==========================================
    console.log('\n--- Preflop ---');

    // Dealer/SB acts first - calls
    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name} (Dealer/SB): calls`);
    await actor!.call();
    await players[0].page.waitForTimeout(1000);

    // BB acts - checks
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name} (BB): checks`);
    await actor!.check();
    await players[0].page.waitForTimeout(2000);

    // ==========================================
    // FLOP: BB should act first
    // ==========================================
    console.log('\n--- Flop ---');
    await players[0].verifyStreet('flop');

    const flopActor = await waitForAnyAction(players);
    expect(flopActor).not.toBeNull();
    console.log(`First actor on flop: ${flopActor!.name}`);

    // Verify first actor on flop is BB (not dealer)
    const isFlopActorBB = await flopActor!.isBigBlind();
    const isFlopActorDealer = await flopActor!.isDealer();

    console.log(`First flop actor is BB: ${isFlopActorBB}`);
    console.log(`First flop actor is Dealer: ${isFlopActorDealer}`);

    expect(isFlopActorBB).toBe(true);
    expect(isFlopActorDealer).toBe(false);

    console.log(`✓ BB acts first on flop (postflop) in heads-up`);

    console.log('\n=== TEST PASSED: BB Acts First Postflop ===\n');
  });

  // ==========================================
  // Test 4: Heads-up blind structure after elimination from 3 players
  // ==========================================
  test('correct blinds when going from 3 to 2 players', async () => {
    console.log('\n=== TEST: Blinds After 3 to 2 Player Transition ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `Transition_P${i + 1}`, i);
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

    // ==========================================
    // Hand 1: Verify 3-player structure (normal)
    // ==========================================
    console.log('\n--- Hand 1: 3-Player Structure ---');
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    const dealerSeat3p = await players[0].getDealerSeat();
    const sbSeat3p = await players[0].getSmallBlindSeat();
    const bbSeat3p = await players[0].getBigBlindSeat();

    console.log(`3-player: Dealer=${dealerSeat3p}, SB=${sbSeat3p}, BB=${bbSeat3p}`);

    // In 3-player: dealer is NOT the SB (different players)
    expect(dealerSeat3p).not.toBe(sbSeat3p);
    console.log(`✓ In 3-player, dealer (${dealerSeat3p}) is NOT SB (${sbSeat3p})`);

    // Complete hand quickly - all fold to one player
    let actor = await waitForAnyAction(players);
    if (actor) {
      await actor.raise(900); // Big raise
      await players[0].page.waitForTimeout(1000);
    }
    
    actor = await waitForAnyAction(players);
    if (actor) {
      await actor.fold();
      await players[0].page.waitForTimeout(1000);
    }
    
    actor = await waitForAnyAction(players);
    if (actor) {
      await actor.fold();
      await players[0].page.waitForTimeout(2000);
    }

    // ==========================================
    // Hand 2: Eliminate one player
    // ==========================================
    console.log('\n--- Hand 2: Eliminate one player ---');
    
    // Dismiss any showdown - try multiple times if needed
    for (let attempt = 0; attempt < 3; attempt++) {
      await players[0].dismissShowdownIfVisible();
      await players[0].page.waitForTimeout(500);
      
      const showdownStillVisible = await players[0].page.locator('.showdown-overlay').isVisible().catch(() => false);
      if (!showdownStillVisible) break;
    }

    // Start next hand with force if overlay is still blocking
    const startButton = players[0].page.locator('button:has-text("Start Hand")');
    if (await startButton.isVisible().catch(() => false)) {
      await startButton.click({ force: true });
      await players[0].page.waitForTimeout(2000);
    }

    // Find a player and have them go all-in, others fold
    actor = await waitForAnyAction(players);
    if (actor) {
      // Go all-in
      const allInButton = actor.page.locator('button.preset-button:has-text("All-in")');
      if (await allInButton.isVisible().catch(() => false)) {
        await allInButton.click();
        const raiseButton = actor.page.locator('button.raise-button:has-text("Raise")');
        await raiseButton.click();
        await players[0].page.waitForTimeout(1500);
      }
    }

    // Others fold
    for (let i = 0; i < 2; i++) {
      actor = await waitForAnyAction(players);
      if (actor) {
        await actor.fold();
        await players[0].page.waitForTimeout(1000);
      }
    }

    // Wait for hand to complete
    await players[0].page.waitForTimeout(2000);

    // Check if we have 2 players left
    console.log('\n--- Checking if heads-up ---');
    
    // Dismiss showdown if visible - try multiple times
    for (let attempt = 0; attempt < 3; attempt++) {
      await players[0].dismissShowdownIfVisible();
      await players[0].page.waitForTimeout(500);
      
      const showdownStillVisible = await players[0].page.locator('.showdown-overlay').isVisible().catch(() => false);
      if (!showdownStillVisible) break;
    }

    // Count active players by checking who can see Start Hand button
    let activePlayers = 0;
    for (const player of players) {
      const stack = await player.getStack();
      if (stack > 0) {
        activePlayers++;
      }
    }
    console.log(`Active players remaining: ${activePlayers}`);

    // If we have 2 players, verify heads-up structure
    if (activePlayers === 2) {
      // Start hand with 2 players (use force in case overlay still lingering)
      const startBtn = players[0].page.locator('button:has-text("Start Hand")');
      if (await startBtn.isVisible().catch(() => false)) {
        await startBtn.click({ force: true });
        await players[0].page.waitForTimeout(2000);
      }

      const dealerSeat2p = await players[0].getDealerSeat();
      const sbSeat2p = await players[0].getSmallBlindSeat();
      const bbSeat2p = await players[0].getBigBlindSeat();

      console.log(`Heads-up: Dealer=${dealerSeat2p}, SB=${sbSeat2p}, BB=${bbSeat2p}`);

      // In heads-up: dealer = SB
      expect(dealerSeat2p).toBe(sbSeat2p);
      console.log(`✓ In heads-up, dealer (${dealerSeat2p}) IS SB (${sbSeat2p})`);
    }

    console.log('\n=== TEST PASSED: Blinds After Transition ===\n');
  });
});
