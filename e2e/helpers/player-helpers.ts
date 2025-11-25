import { Browser, BrowserContext, Page } from '@playwright/test';

/**
 * Configuration for positioning browser windows on screen
 * This allows us to see all players at once without overlap
 */
const WINDOW_WIDTH = 800;
const WINDOW_HEIGHT = 900;
const SCREEN_PADDING = 20;

/**
 * Helper class to manage multiple poker players (browser contexts)
 * Each player gets their own browser context to simulate different users
 * with separate localStorage
 */
export class PokerPlayer {
  public context: BrowserContext;
  public page: Page;
  public name: string;
  public position: number;

  constructor(context: BrowserContext, page: Page, name: string, position: number) {
    this.context = context;
    this.page = page;
    this.name = name;
    this.position = position;
  }

  /**
   * Navigate to the poker application
   */
  async goto(url: string) {
    await this.page.goto(url);
  }

  /**
   * Enter player name and submit
   */
  async enterName() {
    // Wait for the name input to be visible
    await this.page.waitForSelector('input[type="text"]', { timeout: 5000 });
    
    // Fill in the name
    await this.page.fill('input[type="text"]', this.name);
    
    // Submit the form (look for submit button or press Enter)
    const submitButton = this.page.locator('button[type="submit"]');
    if (await submitButton.count() > 0) {
      await submitButton.click();
    } else {
      // If no submit button, try pressing Enter
      await this.page.press('input[type="text"]', 'Enter');
    }
    
    // Wait for navigation/state change
    await this.page.waitForTimeout(1000);
  }

  /**
   * Join a poker table by ID
   */
  async joinTable(tableId: string) {
    // Look for table join input or button
    const tableInput = this.page.locator('input[placeholder*="table" i], input[placeholder*="Table" i]');
    
    if (await tableInput.count() > 0) {
      await tableInput.fill(tableId);
      
      // Find and click join button
      const joinButton = this.page.locator('button:has-text("Join"), button:has-text("join")');
      await joinButton.click();
    } else {
      // Alternative: might be a direct link or different UI
      const tableLink = this.page.locator(`text=${tableId}`);
      if (await tableLink.count() > 0) {
        await tableLink.click();
      }
    }
    
    // Wait for table to load
    await this.page.waitForTimeout(1500);
  }

  /**
   * Start a new hand (usually available to all seated players)
   */
  async startHand() {
    const startButton = this.page.locator('button:has-text("Start"), button:has-text("start")');
    await startButton.click();
    await this.page.waitForTimeout(1000);
  }

  /**
   * Perform a poker action (check, call, raise, fold)
   */
  async performAction(action: 'check' | 'call' | 'raise' | 'fold', amount?: number) {
    const actionButton = this.page.locator(`button:has-text("${action}"), button:has-text("${action.charAt(0).toUpperCase() + action.slice(1)}")`);
    
    if (action === 'raise' && amount !== undefined) {
      // Fill in raise amount first
      const raiseInput = this.page.locator('input[type="number"]');
      await raiseInput.fill(amount.toString());
    }
    
    await actionButton.click();
    await this.page.waitForTimeout(800);
  }

  /**
   * Get current player info from the page
   */
  async getPlayerInfo() {
    // This will depend on your UI structure
    const stackText = await this.page.locator('[data-testid="player-stack"], .player-stack, .stack').textContent().catch(() => null);
    const positionText = await this.page.locator('[data-testid="player-position"], .player-position').textContent().catch(() => null);
    
    return {
      stack: stackText,
      position: positionText,
    };
  }

  /**
   * Close this player's context
   */
  async close() {
    await this.context.close();
  }
}

/**
 * Create multiple poker players with properly positioned browser windows
 */
export async function createPlayers(browser: Browser, baseURL: string, count: number): Promise<PokerPlayer[]> {
  const players: PokerPlayer[] = [];
  
  for (let i = 0; i < count; i++) {
    // Calculate window position to tile them nicely
    const row = Math.floor(i / 3); // 3 windows per row
    const col = i % 3;
    
    const x = SCREEN_PADDING + col * (WINDOW_WIDTH + SCREEN_PADDING);
    const y = SCREEN_PADDING + row * (WINDOW_HEIGHT + SCREEN_PADDING);
    
    // Create new browser context (isolated localStorage)
    const context = await browser.newContext({
      viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      screen: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
    });
    
    // Create new page in this context
    const page = await context.newPage();
    
    // Set window position (this works in headed mode)
    try {
      await page.evaluate(({ x, y, width, height }) => {
        window.moveTo(x, y);
        window.resizeTo(width, height);
      }, { x, y, width: WINDOW_WIDTH, height: WINDOW_HEIGHT });
    } catch (e) {
      // Window positioning might not work in all environments
      console.log(`Could not position window for player ${i + 1}`);
    }
    
    const playerName = `Player${i + 1}`;
    const player = new PokerPlayer(context, page, playerName, i);
    players.push(player);
  }
  
  return players;
}

/**
 * Close all players
 */
export async function closeAllPlayers(players: PokerPlayer[]) {
  for (const player of players) {
    await player.close();
  }
}

/**
 * Helper to wait and observe (useful for debugging)
 */
export async function pauseToObserve(duration: number = 3000) {
  await new Promise(resolve => setTimeout(resolve, duration));
}
