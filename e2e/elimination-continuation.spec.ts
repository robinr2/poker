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
    // Dismiss showdown overlay if it's blocking
    await this.dismissShowdownIfVisible();
    const startButton = this.page.locator('button:has-text("Start Hand")');
    await expect(startButton).toBeVisible({ timeout: 10000 });
    await startButton.click();
    await this.page.waitForTimeout(2000);
  }

  async canStartNewHand(): Promise<boolean> {
    return await this.page.locator('button:has-text("Start Hand")').isVisible().catch(() => false);
  }

  async waitForStartHandButton(timeout = 15000) {
    console.log(`${this.name}: Waiting for Start Hand button...`);
    const startButton = this.page.locator('button:has-text("Start Hand")');
    await expect(startButton).toBeVisible({ timeout });
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

  async hasShowdown(): Promise<boolean> {
    return await this.page.locator('.showdown-overlay').isVisible().catch(() => false);
  }

  async getShowdownResult(): Promise<string> {
    const showdownOverlay = this.page.locator('.showdown-overlay');
    await expect(showdownOverlay).toBeVisible({ timeout: 15000 });
    const text = await showdownOverlay.textContent();
    return text || '';
  }

  async waitForShowdownToClear(timeout = 10000) {
    console.log(`${this.name}: Waiting for showdown to clear...`);
    const startTime = Date.now();
    while (Date.now() - startTime < timeout) {
      const hasShowdown = await this.hasShowdown();
      if (!hasShowdown) {
        console.log(`${this.name}: Showdown cleared`);
        return;
      }
      await this.page.waitForTimeout(500);
    }
    // If showdown didn't clear, try clicking on it to dismiss
    console.log(`${this.name}: Showdown still visible, attempting to dismiss...`);
    const showdownOverlay = this.page.locator('.showdown-overlay');
    await showdownOverlay.click().catch(() => {});
    await this.page.waitForTimeout(1000);
  }

  async dismissShowdownIfVisible() {
    const hasShowdown = await this.hasShowdown();
    if (hasShowdown) {
      console.log(`${this.name}: Dismissing showdown overlay...`);
      // Click the close button first if it exists
      const closeButton = this.page.locator('.showdown-close-button');
      const hasCloseButton = await closeButton.isVisible().catch(() => false);
      if (hasCloseButton) {
        await closeButton.click();
      } else {
        // Fall back to clicking overlay itself
        const showdownOverlay = this.page.locator('.showdown-overlay');
        await showdownOverlay.click({ position: { x: 10, y: 10 } }).catch(() => {});
      }
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

/**
 * Play a simple hand where everyone checks to showdown
 */
async function playCheckdownHand(players: Player[], handNumber: number) {
  console.log(`\n--- Hand ${handNumber}: Check-down ---`);

  // Wait for action and check through all streets
  let actionCount = 0;
  const maxActions = 20; // Safety limit

  while (actionCount < maxActions) {
    const actor = await waitForAnyAction(players, 5000);
    if (!actor) {
      // No action available - might be showdown or hand over
      break;
    }

    // Try to check, if not possible try to call (for BB facing limps)
    const checkButton = actor.page.locator('button:has-text("Check")');
    const isCheckVisible = await checkButton.isVisible().catch(() => false);

    if (isCheckVisible) {
      await actor.check();
    } else {
      // If we can't check, we need to call or fold
      const callButton = actor.page.locator('button:has-text("Call")');
      const isCallVisible = await callButton.isVisible().catch(() => false);
      if (isCallVisible) {
        await actor.call();
      } else {
        // No check or call available - unusual state
        console.log(`${actor.name}: No check or call available, breaking`);
        break;
      }
    }

    actionCount++;
    await actor.page.waitForTimeout(500);
  }

  // Wait for showdown or hand completion
  await players[0].page.waitForTimeout(2000);
}

// ==========================================
// TEST SUITE: Elimination and Continuation
// ==========================================

test.describe('Elimination and Continuation', () => {
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
  // Test 1: 3 players, one gets eliminated, remaining 2 play another hand
  // ==========================================
  test('3-player game continues after one player is eliminated', async () => {
    test.setTimeout(180000); // 3 minutes for this longer test
    console.log('\n=== TEST: Continuation After Elimination ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `Elim_P${i + 1}`, i);
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

    // === HAND 1: P3 loses all chips to P1 ===
    console.log('\n========== HAND 1: Elimination Hand ==========');

    // Set deck so P1 wins with better hand
    const deck1 = buildDeck({
      holeCards: {
        0: ['As', 'Ah'], // P1: Pocket Aces (winner)
        1: ['7c', '2d'], // P2: Garbage (will fold)
        2: ['Ks', 'Kh'], // P3: Pocket Kings (loser, goes all-in)
      },
      flop: ['3d', '5c', '9h'],
      turn: '4s',
      river: 'Jc',
    });
    await setDeck(TABLE_ID, deck1);

    // Record initial stacks
    console.log('\n--- Initial Stacks ---');
    for (const player of players) {
      const stack = await player.getStack();
      console.log(`${player.name}: ${stack}`);
    }

    // Start hand 1
    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // Preflop: P1 goes all-in, P2 folds, P3 calls
    console.log('\n--- Hand 1 Preflop ---');

    // First actor goes all-in (should be UTG, which depends on button position)
    let actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: Goes ALL-IN`);
    await actor!.clickAllIn();
    await players[0].page.waitForTimeout(1000);

    // Second actor - fold or call based on who it is
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();

    // We want to orchestrate P3 losing to P1
    // Whoever is next should respond appropriately
    // If it's the player with garbage (P2), they fold
    // If it's a player with good cards, they call

    // For simplicity: next player folds
    console.log(`${actor!.name}: Folds`);
    await actor!.fold();
    await players[0].page.waitForTimeout(1000);

    // Third actor calls (this should be the player who will lose)
    actor = await waitForAnyAction(players);
    expect(actor).not.toBeNull();
    console.log(`${actor!.name}: Calls all-in`);
    await actor!.call();
    await players[0].page.waitForTimeout(3000);

    // Wait for showdown
    console.log('\n--- Hand 1 Showdown ---');
    const hasShowdown1 = await players[0].hasShowdown();
    expect(hasShowdown1).toBe(true);

    const result1 = await players[0].getShowdownResult();
    console.log(`Hand 1 result: ${result1}`);

    // Wait for showdown to process
    await players[0].page.waitForTimeout(3000);

    // Check stacks after hand 1
    console.log('\n--- Stacks After Hand 1 ---');
    let eliminatedPlayer: Player | null = null;
    let remainingPlayers: Player[] = [];

    for (const player of players) {
      const stack = await player.getStack();
      console.log(`${player.name}: ${stack}`);
      if (stack === 0) {
        eliminatedPlayer = player;
      } else {
        remainingPlayers.push(player);
      }
    }

    // Verify someone got eliminated
    expect(eliminatedPlayer).not.toBeNull();
    console.log(`\nEliminated player: ${eliminatedPlayer!.name}`);
    expect(remainingPlayers.length).toBe(2);
    console.log(`Remaining players: ${remainingPlayers.map(p => p.name).join(', ')}`);

    // Verify total chips are conserved
    let totalChips = 0;
    for (const player of players) {
      totalChips += await player.getStack();
    }
    console.log(`Total chips: ${totalChips} (expected: ${3 * STARTING_STACK})`);
    expect(totalChips).toBe(3 * STARTING_STACK);

    // === HAND 2: Remaining 2 players play another hand ===
    console.log('\n========== HAND 2: Continuation Hand ==========');

    // Wait for showdown to clear and Start Hand button to appear
    await players[0].waitForShowdownToClear();
    await players[0].page.waitForTimeout(2000);

    // Check if we can start a new hand
    // One of the remaining players should be able to start
    let canStart = false;
    let starterPlayer: Player | null = null;

    for (const player of remainingPlayers) {
      const canPlayerStart = await player.canStartNewHand();
      console.log(`${player.name} can start new hand: ${canPlayerStart}`);
      if (canPlayerStart) {
        canStart = true;
        starterPlayer = player;
        break;
      }
    }

    // Also check if eliminated player sees start button (they shouldn't be able to play)
    const eliminatedCanStart = await eliminatedPlayer!.canStartNewHand();
    console.log(`${eliminatedPlayer!.name} (eliminated) sees Start Hand: ${eliminatedCanStart}`);

    // Key assertion: remaining players should be able to start a new hand
    expect(canStart).toBe(true);
    expect(starterPlayer).not.toBeNull();

    // Set deck for hand 2
    const deck2 = buildDeck({
      holeCards: {
        0: ['Qh', 'Qd'], // Will go to whoever is in seat 0
        1: ['Jc', 'Js'], // Will go to whoever is in seat 1
        2: ['2c', '3d'], // Will go to whoever is in seat 2 (might be empty)
      },
      flop: ['4d', '5c', '9h'],
      turn: '8s',
      river: 'Tc',
    });
    await setDeck(TABLE_ID, deck2);

    // Start hand 2
    console.log(`\n${starterPlayer!.name}: Starting hand 2...`);
    await starterPlayer!.startHand();
    await players[0].page.waitForTimeout(2000);

    // Record stacks before hand 2
    console.log('\n--- Stacks Before Hand 2 ---');
    const stacksBefore: Map<string, number> = new Map();
    for (const player of remainingPlayers) {
      const stack = await player.getStack();
      stacksBefore.set(player.name, stack);
      console.log(`${player.name}: ${stack}`);
    }

    // Verify pot has blinds (from 2 remaining players)
    const pot2 = await remainingPlayers[0].getPot();
    console.log(`Pot after blinds (hand 2): ${pot2}`);
    expect(pot2).toBe(SMALL_BLIND + BIG_BLIND);

    // Play hand 2 - simple check-down or fold scenario
    console.log('\n--- Hand 2 Action ---');

    // Get first actor and have them raise small
    actor = await waitForAnyAction(remainingPlayers);
    if (actor) {
      // Just call/check to keep it simple
      const checkButton = actor.page.locator('button:has-text("Check")');
      const isCheckVisible = await checkButton.isVisible().catch(() => false);

      if (isCheckVisible) {
        await actor.check();
      } else {
        await actor.call();
      }
      await actor.page.waitForTimeout(1000);
    }

    // Continue action until hand completes
    let actionsRemaining = 15;
    while (actionsRemaining > 0) {
      actor = await waitForAnyAction(remainingPlayers, 3000);
      if (!actor) break;

      const checkButton = actor.page.locator('button:has-text("Check")');
      const isCheckVisible = await checkButton.isVisible().catch(() => false);

      if (isCheckVisible) {
        await actor.check();
      } else {
        const callButton = actor.page.locator('button:has-text("Call")');
        const isCallVisible = await callButton.isVisible().catch(() => false);
        if (isCallVisible) {
          await actor.call();
        } else {
          break;
        }
      }

      actionsRemaining--;
      await actor.page.waitForTimeout(500);
    }

    // Wait for hand 2 to complete
    await remainingPlayers[0].page.waitForTimeout(3000);

    // Check for showdown or hand completion
    const hasShowdown2 = await remainingPlayers[0].hasShowdown();
    const canStartHand3 = await remainingPlayers[0].canStartNewHand();
    console.log(`Hand 2 - Showdown visible: ${hasShowdown2}, Can start hand 3: ${canStartHand3}`);

    // Hand 2 should have completed (either showdown or someone folded)
    expect(hasShowdown2 || canStartHand3).toBe(true);

    // Verify stacks after hand 2
    console.log('\n--- Stacks After Hand 2 ---');
    let totalChipsAfter = 0;
    for (const player of players) {
      const stack = await player.getStack();
      console.log(`${player.name}: ${stack}`);
      totalChipsAfter += stack;
    }

    // Total chips should still be conserved
    console.log(`Total chips after hand 2: ${totalChipsAfter} (expected: ${3 * STARTING_STACK})`);
    expect(totalChipsAfter).toBe(3 * STARTING_STACK);

    // Eliminated player should still have 0
    const eliminatedStack = await eliminatedPlayer!.getStack();
    expect(eliminatedStack).toBe(0);

    console.log('\n=== TEST PASSED: Continuation After Elimination ===\n');
  });

  // ==========================================
  // Test 2: Multiple hands with chip tracking (no elimination)
  // ==========================================
  test('3-player plays 3 consecutive hands with chip tracking', async () => {
    test.setTimeout(240000); // 4 minutes for this longer test
    console.log('\n=== TEST: Multiple Consecutive Hands ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `Multi_P${i + 1}`, i);
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

    // Track stacks across hands
    const stackHistory: number[][] = [];

    // Record initial stacks
    console.log('\n--- Initial Stacks ---');
    const initialStacks: number[] = [];
    for (const player of players) {
      const stack = await player.getStack();
      console.log(`${player.name}: ${stack}`);
      initialStacks.push(stack);
    }
    stackHistory.push([...initialStacks]);

    // Play 3 hands
    for (let handNum = 1; handNum <= 3; handNum++) {
      console.log(`\n========== HAND ${handNum} ==========`);

      // Set a simple deck (doesn't matter who wins for this test)
      const deck = buildDeck({
        holeCards: {
          0: ['As', 'Kh'],
          1: ['Qc', 'Jd'],
          2: ['Ts', '9h'],
        },
        flop: ['2d', '5c', '8h'],
        turn: '4s',
        river: '7c',
      });
      await setDeck(TABLE_ID, deck);

      // Start hand
      await players[0].startHand();
      await players[0].page.waitForTimeout(2000);

      // Play the hand (everyone checks to showdown)
      await playCheckdownHand(players, handNum);

      // Wait for showdown
      const hasShowdown = await players[0].hasShowdown();
      if (hasShowdown) {
        const result = await players[0].getShowdownResult();
        console.log(`Hand ${handNum} result: ${result.substring(0, 100)}...`);
      }

      // Wait for hand to fully complete
      await players[0].page.waitForTimeout(3000);
      await players[0].waitForShowdownToClear();

      // Record stacks after this hand
      console.log(`\n--- Stacks After Hand ${handNum} ---`);
      const currentStacks: number[] = [];
      for (const player of players) {
        const stack = await player.getStack();
        console.log(`${player.name}: ${stack}`);
        currentStacks.push(stack);
      }
      stackHistory.push([...currentStacks]);

      // Verify chip conservation
      const totalChips = currentStacks.reduce((a, b) => a + b, 0);
      console.log(`Total chips: ${totalChips} (expected: ${3 * STARTING_STACK})`);
      expect(totalChips).toBe(3 * STARTING_STACK);

      // Wait for Start Hand button if not the last hand
      if (handNum < 3) {
        await players[0].waitForStartHandButton();
      }
    }

    // Verify we completed all 3 hands
    console.log('\n--- Stack History ---');
    for (let i = 0; i < stackHistory.length; i++) {
      const label = i === 0 ? 'Initial' : `After Hand ${i}`;
      console.log(`${label}: ${stackHistory[i].join(', ')}`);
    }

    // Chips should have moved around (unlikely all hands were exact ties)
    // At minimum, blinds were posted and collected each hand

    console.log('\n=== TEST PASSED: Multiple Consecutive Hands ===\n');
  });

  // ==========================================
  // Test 3: Play multiple hands AFTER elimination to check for issues
  // ==========================================
  test('play 3 hands after player elimination', async () => {
    test.setTimeout(300000); // 5 minutes for this longer test
    console.log('\n=== TEST: Multiple Hands After Elimination ===\n');

    // Create 3 players
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({
        viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
      });
      const page = await context.newPage();
      await positionWindow(context, page, i);
      const player = new Player(context, page, `Post_P${i + 1}`, i);
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

    // === HAND 1: Eliminate P3 ===
    console.log('\n========== HAND 1: Elimination Hand ==========');

    const deck1 = buildDeck({
      holeCards: {
        0: ['As', 'Ah'], // P1: Pocket Aces (winner)
        1: ['7c', '2d'], // P2: Garbage (will fold)
        2: ['Ks', 'Kh'], // P3: Pocket Kings (loser)
      },
      flop: ['3d', '5c', '9h'],
      turn: '4s',
      river: 'Jc',
    });
    await setDeck(TABLE_ID, deck1);

    await players[0].startHand();
    await players[0].page.waitForTimeout(2000);

    // P1 goes all-in, P2 folds, P3 calls
    let actor = await waitForAnyAction(players);
    await actor!.clickAllIn();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.fold();
    await players[0].page.waitForTimeout(1000);

    actor = await waitForAnyAction(players);
    await actor!.call();
    await players[0].page.waitForTimeout(3000);

    // Verify elimination
    const p3Stack = await players[2].getStack();
    console.log(`P3 stack after elimination hand: ${p3Stack}`);
    expect(p3Stack).toBe(0);

    // Wait for showdown
    await players[0].page.waitForTimeout(3000);

    // Get remaining players (those with chips)
    const remainingPlayers = [players[0], players[1]];

    // === HAND 2: First post-elimination hand ===
    console.log('\n========== HAND 2: First Post-Elimination ==========');

    const deck2 = buildDeck({
      holeCards: {
        0: ['Qh', 'Qd'],
        1: ['Jc', 'Js'],
        2: ['2c', '3d'],
      },
      flop: ['4d', '6c', '8h'],
      turn: '9s',
      river: 'Tc',
    });
    await setDeck(TABLE_ID, deck2);

    await remainingPlayers[0].startHand();
    await remainingPlayers[0].page.waitForTimeout(2000);

    console.log('Hand 2 started, playing check-down...');
    await playCheckdownHand(remainingPlayers, 2);
    await remainingPlayers[0].page.waitForTimeout(3000);

    // Verify stacks
    console.log('\n--- Stacks After Hand 2 ---');
    for (const player of players) {
      const stack = await player.getStack();
      console.log(`${player.name}: ${stack}`);
    }

    // === HAND 3: Second post-elimination hand ===
    console.log('\n========== HAND 3: Second Post-Elimination ==========');

    const deck3 = buildDeck({
      holeCards: {
        0: ['Ah', 'Kd'],
        1: ['Tc', 'Ts'],
        2: ['5c', '6d'],
      },
      flop: ['2d', '3c', '7h'],
      turn: 'Qs',
      river: '8c',
    });
    await setDeck(TABLE_ID, deck3);

    await remainingPlayers[0].startHand();
    await remainingPlayers[0].page.waitForTimeout(2000);

    console.log('Hand 3 started, playing check-down...');
    await playCheckdownHand(remainingPlayers, 3);
    await remainingPlayers[0].page.waitForTimeout(3000);

    // Verify stacks
    console.log('\n--- Stacks After Hand 3 ---');
    for (const player of players) {
      const stack = await player.getStack();
      console.log(`${player.name}: ${stack}`);
    }

    // === HAND 4: Third post-elimination hand ===
    console.log('\n========== HAND 4: Third Post-Elimination ==========');

    const deck4 = buildDeck({
      holeCards: {
        0: ['9h', '9d'],
        1: ['Ac', 'Qh'],
        2: ['4c', '5d'],
      },
      flop: ['2s', '3h', 'Kd'],
      turn: 'Js',
      river: '6h',
    });
    await setDeck(TABLE_ID, deck4);

    await remainingPlayers[0].startHand();
    await remainingPlayers[0].page.waitForTimeout(2000);

    console.log('Hand 4 started, playing check-down...');
    await playCheckdownHand(remainingPlayers, 4);
    await remainingPlayers[0].page.waitForTimeout(3000);

    // Final verification
    console.log('\n--- Final Stacks ---');
    let totalChips = 0;
    for (const player of players) {
      const stack = await player.getStack();
      console.log(`${player.name}: ${stack}`);
      totalChips += stack;
    }

    console.log(`Total chips: ${totalChips} (expected: ${3 * STARTING_STACK})`);
    expect(totalChips).toBe(3 * STARTING_STACK);

    // Eliminated player should still have 0
    const finalP3Stack = await players[2].getStack();
    expect(finalP3Stack).toBe(0);

    console.log('\n=== TEST PASSED: Multiple Hands After Elimination ===\n');
  });
});
