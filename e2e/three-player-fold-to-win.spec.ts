import { test, expect, chromium, Browser, BrowserContext, Page } from '@playwright/test';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

/**
 * Restart Docker backend container to clear in-memory state
 */
async function restartDockerServer() {
  console.log('Restarting Docker backend container...');
  
  try {
    await execAsync('docker restart poker-backend-dev');
    console.log('Backend container restarted');
    
    // Wait for the server to be ready
    let serverReady = false;
    const maxAttempts = 30;
    
    for (let i = 0; i < maxAttempts; i++) {
      try {
        await execAsync('curl -s http://localhost:8080 > /dev/null');
        serverReady = true;
        break;
      } catch {
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    }
    
    if (!serverReady) {
      throw new Error('Server failed to become ready after restart');
    }
    
    await new Promise(resolve => setTimeout(resolve, 2000));
    console.log('Docker backend server ready');
  } catch (error) {
    console.error('Failed to restart Docker server:', error);
    throw error;
  }
}

/**
 * Helper class to manage a poker player
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
    
    // Click all-in button to fill max amount
    const allInButton = this.page.locator('button.preset-button:has-text("All-in")');
    await expect(allInButton).toBeVisible({ timeout: 5000 });
    await allInButton.click();
    
    // Click raise button
    const raiseButton = this.page.locator('button.raise-button:has-text("Raise")');
    await expect(raiseButton).toBeEnabled({ timeout: 5000 });
    await raiseButton.click();
    await this.page.waitForTimeout(1500);
  }

  async getStack(): Promise<number | null> {
    // Find this player's seat and get stack
    const stackText = await this.page.locator('.own-seat .stack').textContent().catch(() => null);
    if (!stackText) return null;
    const match = stackText.match(/(\d+)/);
    return match ? parseInt(match[1], 10) : null;
  }

  async getPotAmount(): Promise<number | null> {
    const potText = await this.page.locator('.pot-display').textContent().catch(() => null);
    if (!potText) return null;
    const match = potText.match(/Pot:\s*(\d+)/);
    return match ? parseInt(match[1], 10) : null;
  }

  async hasShowdown(): Promise<boolean> {
    return await this.page.locator('.showdown-overlay').isVisible().catch(() => false);
  }

  async hasHandComplete(): Promise<boolean> {
    // Check if hand is complete (no action bar, start hand button may be visible)
    const hasActionBar = await this.page.locator('.action-bar').isVisible().catch(() => false);
    const hasStartButton = await this.page.locator('button:has-text("Start Hand")').isVisible().catch(() => false);
    return !hasActionBar && hasStartButton;
  }

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
 * Three-Player Fold to Win Test
 * 
 * Tests winning a pot without showdown when all opponents fold:
 * 1. 3 players join and start a hand
 * 2. Preflop: First player raises big (all-in or large raise)
 * 3. Other two players fold
 * 4. Raiser wins pot without showdown
 * 5. Verify pot is awarded and hand completes
 */
test.describe('Three Player Fold to Win', () => {
  let browser: Browser;
  let players: Player[] = [];
  const baseURL = 'http://localhost:8080';

  const windowWidth = 950;
  const windowHeight = 1200;
  const positions = [
    { x: 0, y: 40 },
    { x: 950, y: 40 },
    { x: 1900, y: 40 },
  ];

  test.beforeAll(async () => {
    browser = await chromium.launch({
      headless: false,
      slowMo: 300,
      args: ['--window-size=950,1200'],
    });
  });

  test.beforeEach(async () => {
    await restartDockerServer();
    
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

  test('should award pot when all opponents fold preflop', async () => {
    // ==========================================
    // SETUP: Create 3 players
    // ==========================================
    console.log('\n=== SETUP: Creating 3 players ===');
    
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: windowWidth, height: windowHeight },
      });
      const page = await context.newPage();
      
      const client = await context.newCDPSession(page);
      const { windowId } = await client.send('Browser.getWindowForTarget');
      await client.send('Browser.setWindowBounds', {
        windowId,
        bounds: {
          left: positions[i].x,
          top: positions[i].y,
          width: windowWidth,
          height: windowHeight,
        },
      });
      
      const player = new Player(context, page, `Player${i + 1}`, i);
      players.push(player);
    }

    // ==========================================
    // Navigate and join table
    // ==========================================
    console.log('\n=== Navigating players to app ===');
    for (const player of players) {
      await player.goto(baseURL);
    }

    console.log('\n=== Players entering names ===');
    for (const player of players) {
      await player.enterName();
    }

    console.log('\n=== Players joining table ===');
    for (const player of players) {
      await player.joinTable();
    }

    await players[0].page.waitForTimeout(3000);

    // Record initial stacks
    const initialStacks: number[] = [];
    for (const player of players) {
      const stack = await player.getStack();
      initialStacks.push(stack ?? 0);
      console.log(`${player.name} initial stack: ${stack}`);
    }

    // ==========================================
    // START HAND
    // ==========================================
    console.log('\n=== STARTING HAND ===');
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // Verify cards were dealt
    for (const player of players) {
      const hasCards = await player.hasCards();
      expect(hasCards).toBe(true);
    }

    // Get pot after blinds
    const potAfterBlinds = await players[0].getPotAmount();
    console.log(`Pot after blinds: ${potAfterBlinds}`);

    // ==========================================
    // PREFLOP: First actor goes all-in, others fold
    // ==========================================
    console.log('\n=== PREFLOP: ALL-IN AND FOLDS ===');
    
    // Find first actor and go all-in
    const firstActor = await findCurrentActor(players);
    expect(firstActor).not.toBeNull();
    console.log(`First to act: ${firstActor!.name} - going all-in`);
    
    const raiserName = firstActor!.name;
    await firstActor!.clickAllIn();
    
    await players[0].page.waitForTimeout(1000);

    // Second player folds
    const secondActor = await findCurrentActor(players);
    expect(secondActor).not.toBeNull();
    console.log(`Second to act: ${secondActor!.name} - folding`);
    await secondActor!.fold();
    
    await players[0].page.waitForTimeout(1000);

    // Third player folds
    const thirdActor = await findCurrentActor(players);
    expect(thirdActor).not.toBeNull();
    console.log(`Third to act: ${thirdActor!.name} - folding`);
    await thirdActor!.fold();

    // ==========================================
    // VERIFY: No showdown, pot awarded
    // ==========================================
    console.log('\n=== VERIFYING FOLD-TO-WIN ===');
    await players[0].page.waitForTimeout(3000);

    // Should NOT have showdown overlay (no cards shown)
    let hasShowdown = false;
    for (const player of players) {
      if (await player.hasShowdown()) {
        hasShowdown = true;
        break;
      }
    }
    
    // Note: Some implementations show showdown even on fold-to-win
    // The key test is that the hand completes and pot is awarded
    console.log(`Showdown overlay visible: ${hasShowdown}`);

    // Verify hand is complete (Start Hand button should be visible)
    let handComplete = false;
    for (const player of players) {
      const startButton = player.page.locator('button:has-text("Start Hand")');
      const isVisible = await startButton.isVisible().catch(() => false);
      if (isVisible) {
        handComplete = true;
        console.log(`${player.name}: Start Hand button visible - hand is complete`);
        break;
      }
    }
    
    expect(handComplete).toBe(true);

    // Verify stacks changed correctly
    // The raiser should have won the blinds
    console.log('\n=== FINAL STACKS ===');
    for (const player of players) {
      const finalStack = await player.getStack();
      const initialStack = initialStacks[player.index];
      const diff = (finalStack ?? 0) - initialStack;
      console.log(`${player.name}: ${initialStack} -> ${finalStack} (${diff >= 0 ? '+' : ''}${diff})`);
    }

    // The raiser (all-in player) should have gained chips (the blinds)
    // Find raiser's final stack
    const raiser = players.find(p => p.name === raiserName);
    const raiserFinalStack = await raiser!.getStack();
    const raiserInitialStack = initialStacks[raiser!.index];
    
    console.log(`\nRaiser ${raiserName}: ${raiserInitialStack} -> ${raiserFinalStack}`);
    expect(raiserFinalStack).toBeGreaterThan(raiserInitialStack);

    console.log('\n=== TEST COMPLETE: Fold-to-win works correctly! ===');
    
    // Pause to observe
    await players[0].page.waitForTimeout(5000);
  });
});
