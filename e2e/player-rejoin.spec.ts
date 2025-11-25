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
 * Player class for disconnect tests
 */
class Player {
  public sessionToken: string | null = null;

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
    
    this.sessionToken = await this.page.evaluate(() => {
      return localStorage.getItem('poker_session_token');
    });
    console.log(`${this.name}: Got session token: ${this.sessionToken?.substring(0, 8)}...`);
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
    await this.dismissShowdownIfVisible();
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

  async leaveTable() {
    console.log(`${this.name}: Leaving table...`);
    const leaveButton = this.page.locator('button:has-text("Leave Table"), button:has-text("Leave")');
    await expect(leaveButton).toBeVisible({ timeout: 5000 });
    await leaveButton.click();
    await this.page.waitForTimeout(1500);
  }

  async getPot(): Promise<number> {
    const potText = await this.page.locator('.pot-display').textContent().catch(() => null);
    if (!potText) return 0;
    const match = potText.match(/Pot:\s*(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }

  async getStack(): Promise<number> {
    const stackText = await this.page.locator('.own-seat .stack').textContent().catch(() => null);
    if (!stackText) return 0;
    const match = stackText.match(/(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
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
      // Try clicking the overlay background or close button
      const closeButton = this.page.locator('.showdown-overlay button:has-text("×"), .showdown-overlay button:has-text("Close")');
      if (await closeButton.isVisible().catch(() => false)) {
        await closeButton.click();
      } else {
        // Click the overlay itself
        await showdownOverlay.click({ position: { x: 10, y: 10 } }).catch(() => {});
      }
      await showdownOverlay.waitFor({ state: 'hidden', timeout: 5000 }).catch(() => {});
      await this.page.waitForTimeout(500);
    }
  }

  async isAtTable(): Promise<boolean> {
    const leaveButton = this.page.locator('button:has-text("Leave Table"), button:has-text("Leave")');
    return await leaveButton.isVisible().catch(() => false);
  }

  async isInLobby(): Promise<boolean> {
    const joinButton = this.page.locator('button.join-button').first();
    return await joinButton.isVisible().catch(() => false);
  }

  async close() {
    await this.context.close();
  }
}

/**
 * Find which player's turn it is
 */
async function findCurrentActor(players: Player[]): Promise<Player | null> {
  for (const player of players) {
    if (await player.isMyTurn()) {
      return player;
    }
  }
  return null;
}

// ==========================================
// TEST SUITE: Player Disconnect Mid-Hand
// 
// Expected behavior:
// - When a player disconnects/leaves during a hand, they are treated as folded
// - The remaining player(s) win the pot
// - The next hand can start without the disconnected player
// ==========================================

test.describe('Player Disconnect Mid-Hand Tests', () => {
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
  // Test 1: Player disconnects in 2-player game
  // Expected: Treated as fold, opponent wins pot immediately
  // ==========================================
  test('2-player disconnect is treated as fold - opponent wins', async () => {
    console.log('\n=== TEST: 2-Player Disconnect = Fold ===\n');

    // Create 2 players
    for (let i = 0; i < 2; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `Disconnect_P${i + 1}`, i);
      players.push(player);
    }

    // Setup both players
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
    await players[0].startHand();
    console.log('Hand started');

    // Find who acts first (dealer/SB in heads-up)
    const firstActor = await findCurrentActor(players);
    expect(firstActor).not.toBeNull();
    console.log(`First actor: ${firstActor!.name}`);

    // The other player is the one who will remain
    const remainingPlayer = players.find(p => p !== firstActor)!;
    
    // Record remaining player's stack before disconnect
    const stackBefore = await remainingPlayer.getStack();
    console.log(`Remaining player stack before: ${stackBefore}`);

    // First actor disconnects (simulates close tab)
    console.log(`${firstActor!.name} disconnecting...`);
    await firstActor!.context.close();
    
    // Wait for server to process disconnect
    await remainingPlayer.page.waitForTimeout(3000);

    // Remaining player should see showdown (early winner) or be able to start new hand
    const hasShowdown = await remainingPlayer.hasShowdown();
    console.log(`Has showdown: ${hasShowdown}`);

    if (hasShowdown) {
      // Dismiss showdown
      await remainingPlayer.dismissShowdownIfVisible();
      await remainingPlayer.page.waitForTimeout(1000);
    }

    // Remaining player should have won the pot (blinds)
    const stackAfter = await remainingPlayer.getStack();
    console.log(`Remaining player stack after: ${stackAfter}`);

    // They should have gained the blinds (SB + BB = 30) minus their own blind
    expect(stackAfter).toBeGreaterThan(stackBefore);
    console.log(`✓ Remaining player gained chips: ${stackAfter - stackBefore}`);

    // Should be able to see table (but can't start hand alone)
    expect(await remainingPlayer.isAtTable()).toBe(true);
    console.log('✓ Remaining player still at table');

    console.log('\n=== TEST PASSED ===\n');
  });

  // ==========================================
  // Test 2: 3-player game - one disconnects, other two continue through all streets
  // Expected: Disconnected player is skipped, remaining two play through flop/turn/river to showdown
  // ==========================================
  test('3-player game continues after one player disconnects - full hand completion', async () => {
    console.log('\n=== TEST: 3-Player Disconnect - Full Hand Continuation ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `Three_P${i + 1}`, i);
      players.push(player);
    }

    // Setup all players
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
    await players[0].startHand();
    console.log('Hand started with 3 players');

    // Find first actor (UTG in 3-player game)
    const firstActor = await findCurrentActor(players);
    expect(firstActor).not.toBeNull();
    console.log(`First actor (UTG): ${firstActor!.name}`);

    // First actor calls (UTG limps in)
    await firstActor!.call();
    console.log(`${firstActor!.name} called (limped)`);
    await players[0].page.waitForTimeout(1500);

    // Find next actor (should be SB or next player)
    const secondActor = await findCurrentActor(players);
    expect(secondActor).not.toBeNull();
    console.log(`Second actor: ${secondActor!.name}`);

    // Second actor will disconnect mid-hand
    const disconnectingPlayer = secondActor!;
    const remainingPlayers = players.filter(p => p !== disconnectingPlayer);
    
    // Record stacks before disconnect
    const stacksBefore: Record<string, number> = {};
    for (const p of remainingPlayers) {
      stacksBefore[p.name] = await p.getStack();
    }
    console.log('Stacks before disconnect:', stacksBefore);

    // Disconnect the second actor mid-hand
    console.log(`\n>>> ${disconnectingPlayer.name} DISCONNECTING mid-hand <<<\n`);
    await disconnectingPlayer.context.close();
    
    // Wait for server to process disconnect and advance action
    await remainingPlayers[0].page.waitForTimeout(3000);

    // The remaining two players should continue - find who has action now
    let currentActor = await findCurrentActor(remainingPlayers);
    
    // The game should continue with remaining players
    // Complete preflop betting
    console.log('\n--- PREFLOP (after disconnect) ---');
    while (currentActor) {
      const actionBar = currentActor.page.locator('.action-bar');
      const checkBtn = currentActor.page.locator('button:has-text("Check")');
      const callBtn = currentActor.page.locator('button:has-text("Call")');
      
      if (await checkBtn.isVisible().catch(() => false)) {
        console.log(`${currentActor.name}: checking`);
        await checkBtn.click();
      } else if (await callBtn.isVisible().catch(() => false)) {
        console.log(`${currentActor.name}: calling`);
        await callBtn.click();
      } else {
        console.log(`${currentActor.name}: no valid action found, breaking`);
        break;
      }
      
      await remainingPlayers[0].page.waitForTimeout(1500);
      currentActor = await findCurrentActor(remainingPlayers);
    }

    // Check if we're on the flop (3 community cards should be visible)
    console.log('\n--- FLOP ---');
    await remainingPlayers[0].page.waitForTimeout(1000);
    
    // Play through flop
    currentActor = await findCurrentActor(remainingPlayers);
    while (currentActor) {
      const checkBtn = currentActor.page.locator('button:has-text("Check")');
      const callBtn = currentActor.page.locator('button:has-text("Call")');
      
      if (await checkBtn.isVisible().catch(() => false)) {
        console.log(`${currentActor.name}: checking on flop`);
        await checkBtn.click();
      } else if (await callBtn.isVisible().catch(() => false)) {
        console.log(`${currentActor.name}: calling on flop`);
        await callBtn.click();
      } else {
        break;
      }
      
      await remainingPlayers[0].page.waitForTimeout(1500);
      currentActor = await findCurrentActor(remainingPlayers);
    }

    // Play through turn
    console.log('\n--- TURN ---');
    await remainingPlayers[0].page.waitForTimeout(1000);
    currentActor = await findCurrentActor(remainingPlayers);
    while (currentActor) {
      const checkBtn = currentActor.page.locator('button:has-text("Check")');
      const callBtn = currentActor.page.locator('button:has-text("Call")');
      
      if (await checkBtn.isVisible().catch(() => false)) {
        console.log(`${currentActor.name}: checking on turn`);
        await checkBtn.click();
      } else if (await callBtn.isVisible().catch(() => false)) {
        console.log(`${currentActor.name}: calling on turn`);
        await callBtn.click();
      } else {
        break;
      }
      
      await remainingPlayers[0].page.waitForTimeout(1500);
      currentActor = await findCurrentActor(remainingPlayers);
    }

    // Play through river
    console.log('\n--- RIVER ---');
    await remainingPlayers[0].page.waitForTimeout(1000);
    currentActor = await findCurrentActor(remainingPlayers);
    while (currentActor) {
      const checkBtn = currentActor.page.locator('button:has-text("Check")');
      const callBtn = currentActor.page.locator('button:has-text("Call")');
      
      if (await checkBtn.isVisible().catch(() => false)) {
        console.log(`${currentActor.name}: checking on river`);
        await checkBtn.click();
      } else if (await callBtn.isVisible().catch(() => false)) {
        console.log(`${currentActor.name}: calling on river`);
        await callBtn.click();
      } else {
        break;
      }
      
      await remainingPlayers[0].page.waitForTimeout(1500);
      currentActor = await findCurrentActor(remainingPlayers);
    }

    // Wait for showdown
    console.log('\n--- SHOWDOWN ---');
    await remainingPlayers[0].page.waitForTimeout(2000);

    // At least one player should see the showdown overlay
    let showdownSeen = false;
    for (const player of remainingPlayers) {
      if (await player.hasShowdown()) {
        showdownSeen = true;
        console.log(`✓ ${player.name} sees showdown overlay`);
        break;
      }
    }
    
    // If no showdown overlay, check if hand completed (Start Hand button visible)
    if (!showdownSeen) {
      const canStartNew = await remainingPlayers[0].canStartNewHand();
      if (canStartNew) {
        console.log('✓ Hand completed (Start Hand button visible)');
        showdownSeen = true;
      }
    }
    
    expect(showdownSeen).toBe(true);
    console.log('✓ Hand reached showdown/completion with 2 remaining players');

    // Dismiss showdown overlays
    for (const player of remainingPlayers) {
      await player.dismissShowdownIfVisible();
    }
    await remainingPlayers[0].page.waitForTimeout(1000);

    // Both remaining players should still be at the table
    for (const player of remainingPlayers) {
      expect(await player.isAtTable()).toBe(true);
    }
    console.log('✓ Both remaining players still at table');

    // Verify they can start a new hand
    const canStartNewHand = await remainingPlayers[0].canStartNewHand();
    expect(canStartNewHand).toBe(true);
    console.log('✓ Can start new hand with remaining players');

    console.log('\n=== TEST PASSED ===\n');
  });
});
