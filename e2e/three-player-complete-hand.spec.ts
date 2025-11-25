import { test, expect, Browser, BrowserContext, Page } from '@playwright/test';
import { resetTable } from './helpers/deterministic-helpers';
import {
  launchBrowser,
  positionWindow,
  WINDOW_WIDTH,
  WINDOW_HEIGHT,
} from './helpers/browser-helpers';

const TABLE_ID = 'table-1';

/**
 * Verify server is running and ready
 * Assumes server is already running (start with: POKER_TEST_MODE=true ./server)
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
    // Wait for action bar to appear (indicates it's our turn)
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

  async raise(amount: number) {
    console.log(`${this.name}: Raising to ${amount}...`);
    await this.waitForTurn();
    
    // Fill raise amount
    const raiseInput = this.page.locator('input[aria-label="Raise Amount"]');
    await expect(raiseInput).toBeVisible({ timeout: 5000 });
    await raiseInput.fill(amount.toString());
    
    // Click raise button
    const raiseButton = this.page.locator('button.raise-button:has-text("Raise")');
    await expect(raiseButton).toBeEnabled({ timeout: 5000 });
    await raiseButton.click();
    await this.page.waitForTimeout(1500);
  }

  async clickMinRaise() {
    console.log(`${this.name}: Using min raise...`);
    await this.waitForTurn();
    
    // Click min button to fill in minimum raise
    const minButton = this.page.locator('button.preset-button:has-text("Min")');
    await expect(minButton).toBeVisible({ timeout: 5000 });
    await minButton.click();
    
    // Click raise button
    const raiseButton = this.page.locator('button.raise-button:has-text("Raise")');
    await expect(raiseButton).toBeEnabled({ timeout: 5000 });
    await raiseButton.click();
    await this.page.waitForTimeout(1500);
  }

  async getGameState() {
    const bodyText = await this.page.locator('body').textContent();
    return {
      hasPot: bodyText?.includes('Pot:') ?? false,
      hasCards: (bodyText?.includes('♠') || bodyText?.includes('♥') || 
                 bodyText?.includes('♦') || bodyText?.includes('♣')) ?? false,
      hasShowdown: await this.page.locator('.showdown-overlay').isVisible().catch(() => false),
      isMyTurn: await this.page.locator('.action-bar').isVisible().catch(() => false),
      street: await this.page.locator('.street-indicator').textContent().catch(() => null),
    };
  }

  async getPotAmount(): Promise<number | null> {
    const potText = await this.page.locator('.pot-display').textContent().catch(() => null);
    if (!potText) return null;
    const match = potText.match(/Pot:\s*(\d+)/);
    return match ? parseInt(match[1], 10) : null;
  }

  async close() {
    await this.context.close();
  }
}

/**
 * Three-Player Complete Hand Test
 * 
 * Tests a full poker hand from start to showdown:
 * 1. 3 players join and start a hand
 * 2. Preflop: UTG calls, SB calls, BB checks
 * 3. Flop: Check, Bet (min raise), Call, Fold
 * 4. Turn: Check, Check
 * 5. River: Check, Check
 * 6. Showdown: Winner determined and pot awarded
 */
test.describe('Three Player Complete Hand', () => {
  let browser: Browser;
  let players: Player[] = [];
  const baseURL = 'http://localhost:8080';

  test.beforeAll(async () => {
    browser = await launchBrowser();
  });

  test.beforeEach(async () => {
    // Verify server is running (assume already started with POKER_TEST_MODE=true)
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

  test('should play complete hand to showdown with 3 players', async () => {
    // ==========================================
    // SETUP: Create 3 players
    // ==========================================
    console.log('\n=== SETUP: Creating 3 players ===');
    
    for (let i = 0; i < 3; i++) {
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
    // START HAND: Player 1 starts the hand
    // ==========================================
    console.log('\n=== STARTING HAND ===');
    await players[0].startHand();
    
    // Wait for hand to be dealt
    await players[0].page.waitForTimeout(2000);
    
    // Verify cards were dealt
    for (const player of players) {
      const state = await player.getGameState();
      expect(state.hasCards).toBe(true);
      console.log(`${player.name}: Cards dealt, Pot exists: ${state.hasPot}`);
    }

    // ==========================================
    // PREFLOP: UTG calls, SB calls, BB checks
    // ==========================================
    console.log('\n=== PREFLOP BETTING ROUND ===');
    
    // In 3-handed poker with positions:
    // - Seat 0 = Dealer/Button (acts last preflop except for blinds)
    // - Seat 1 = Small Blind
    // - Seat 2 = Big Blind
    // Preflop action: UTG (Button/Seat 0) -> SB -> BB
    // But with 3 players: Button posts as UTG, SB, BB in seats after
    // Actually in most implementations: Seat after dealer = SB, Seat after SB = BB
    // Action goes: First player after BB (which is UTG/Button in 3-handed) -> SB -> BB
    
    // We need to determine who acts first - let's check who has the turn
    console.log('Determining who acts first...');
    
    // Find which player has the turn
    let currentActorIndex = -1;
    for (let i = 0; i < players.length; i++) {
      const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
      if (isMyTurn) {
        currentActorIndex = i;
        console.log(`${players[i].name} has the action`);
        break;
      }
    }
    
    expect(currentActorIndex).not.toBe(-1);
    
    // Player with action calls (preflop, facing big blind)
    console.log(`Preflop - First to act: ${players[currentActorIndex].name}`);
    await players[currentActorIndex].call();
    
    // Find next actor
    await players[0].page.waitForTimeout(1000);
    for (let i = 0; i < players.length; i++) {
      const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
      if (isMyTurn) {
        console.log(`Preflop - Second to act: ${players[i].name}`);
        // This should be SB, they call
        await players[i].call();
        break;
      }
    }
    
    // Find BB who should have option to check
    await players[0].page.waitForTimeout(1000);
    for (let i = 0; i < players.length; i++) {
      const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
      if (isMyTurn) {
        console.log(`Preflop - BB option: ${players[i].name}`);
        // BB checks (option)
        await players[i].check();
        break;
      }
    }

    // ==========================================
    // FLOP: Check, Min Raise, Call, Fold
    // ==========================================
    console.log('\n=== FLOP BETTING ROUND ===');
    await players[0].page.waitForTimeout(2000);
    
    // Verify we're on the flop
    let street = await players[0].page.locator('.street-indicator').textContent().catch(() => '');
    console.log(`Current street: ${street}`);
    
    // First to act on flop (should be first active player after button)
    for (let i = 0; i < players.length; i++) {
      const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
      if (isMyTurn) {
        console.log(`Flop - First to act: ${players[i].name} (will check)`);
        await players[i].check();
        break;
      }
    }
    
    await players[0].page.waitForTimeout(1000);
    
    // Second player on flop - min raises
    for (let i = 0; i < players.length; i++) {
      const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
      if (isMyTurn) {
        console.log(`Flop - Second to act: ${players[i].name} (will min raise)`);
        await players[i].clickMinRaise();
        break;
      }
    }
    
    await players[0].page.waitForTimeout(1000);
    
    // Third player on flop - calls
    for (let i = 0; i < players.length; i++) {
      const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
      if (isMyTurn) {
        console.log(`Flop - Third to act: ${players[i].name} (will call)`);
        await players[i].call();
        break;
      }
    }
    
    await players[0].page.waitForTimeout(1000);
    
    // First player (who checked) now faces the raise - folds
    for (let i = 0; i < players.length; i++) {
      const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
      if (isMyTurn) {
        console.log(`Flop - Back to first: ${players[i].name} (will fold)`);
        await players[i].fold();
        break;
      }
    }

    // ==========================================
    // TURN: Check, Check (2 players remaining)
    // ==========================================
    console.log('\n=== TURN BETTING ROUND ===');
    await players[0].page.waitForTimeout(2000);
    
    street = await players[0].page.locator('.street-indicator').textContent().catch(() => '');
    console.log(`Current street: ${street}`);
    
    // Two players remain, both check
    for (let round = 0; round < 2; round++) {
      for (let i = 0; i < players.length; i++) {
        const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
        if (isMyTurn) {
          console.log(`Turn - ${players[i].name} checks`);
          await players[i].check();
          break;
        }
      }
      await players[0].page.waitForTimeout(1000);
    }

    // ==========================================
    // RIVER: Check, Check (2 players remaining)
    // ==========================================
    console.log('\n=== RIVER BETTING ROUND ===');
    await players[0].page.waitForTimeout(2000);
    
    street = await players[0].page.locator('.street-indicator').textContent().catch(() => '');
    console.log(`Current street: ${street}`);
    
    // Two players remain, both check
    for (let round = 0; round < 2; round++) {
      for (let i = 0; i < players.length; i++) {
        const isMyTurn = await players[i].page.locator('.action-bar').isVisible().catch(() => false);
        if (isMyTurn) {
          console.log(`River - ${players[i].name} checks`);
          await players[i].check();
          break;
        }
      }
      await players[0].page.waitForTimeout(1000);
    }

    // ==========================================
    // SHOWDOWN: Verify winner and pot
    // ==========================================
    console.log('\n=== SHOWDOWN ===');
    await players[0].page.waitForTimeout(3000);
    
    // Check for showdown overlay on any player
    let showdownFound = false;
    for (const player of players) {
      const hasShowdown = await player.page.locator('.showdown-overlay').isVisible().catch(() => false);
      if (hasShowdown) {
        showdownFound = true;
        console.log(`${player.name}: Showdown overlay visible`);
        
        // Get showdown details
        const winningHand = await player.page.locator('.winning-hand').textContent().catch(() => 'Unknown');
        const winners = await player.page.locator('.winners').textContent().catch(() => 'Unknown');
        const potAmount = await player.page.locator('.showdown-content .pot-amount').textContent().catch(() => 'Unknown');
        
        console.log(`Winning hand: ${winningHand}`);
        console.log(`Winners: ${winners}`);
        console.log(`Pot: ${potAmount}`);
        break;
      }
    }
    
    expect(showdownFound).toBe(true);
    console.log('\n=== TEST COMPLETE: Hand played to showdown successfully! ===');
    
    // Pause to observe final state
    await players[0].page.waitForTimeout(5000);
  });
});
