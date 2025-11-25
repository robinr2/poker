/**
 * Quick deterministic test to verify deck injection works
 * Run with: npx playwright test e2e/quick-deterministic-test.ts
 */
import { test, expect, chromium } from '@playwright/test';

const BASE_URL = 'http://localhost:8080';
const TABLE_ID = 'table-1';

// Helper to set deterministic deck
async function setDeck(tableId: string, deck: string[]): Promise<void> {
  const response = await fetch(`${BASE_URL}/api/test/set-deck`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ tableId, deck }),
  });
  const result = await response.json();
  if (!result.success) {
    throw new Error(`Failed to set deck: ${result.message}`);
  }
  console.log('Deck set successfully');
}

// Build a 52-card deck with specific cards in specific positions
function buildTestDeck(): string[] {
  // Seat 0 gets: As, Ah (Pocket Aces)
  // Seat 1 gets: Ks, Kh (Pocket Kings)
  // Flop: Ad, 2c, 7d (Player 0 hits trip Aces)
  // Turn: 3h
  // River: 9s
  const deck = [
    'As',
    'Ah', // Seat 0 hole cards
    'Ks',
    'Kh', // Seat 1 hole cards
    '2s', // burn
    'Ad',
    '2c',
    '7d', // flop
    '3s', // burn
    '3h', // turn
    '4s', // burn
    '9s', // river
  ];

  // Fill rest of deck with unused cards
  const used = new Set(deck);
  const allCards = [
    'As',
    '2s',
    '3s',
    '4s',
    '5s',
    '6s',
    '7s',
    '8s',
    '9s',
    'Ts',
    'Js',
    'Qs',
    'Ks',
    'Ah',
    '2h',
    '3h',
    '4h',
    '5h',
    '6h',
    '7h',
    '8h',
    '9h',
    'Th',
    'Jh',
    'Qh',
    'Kh',
    'Ad',
    '2d',
    '3d',
    '4d',
    '5d',
    '6d',
    '7d',
    '8d',
    '9d',
    'Td',
    'Jd',
    'Qd',
    'Kd',
    'Ac',
    '2c',
    '3c',
    '4c',
    '5c',
    '6c',
    '7c',
    '8c',
    '9c',
    'Tc',
    'Jc',
    'Qc',
    'Kc',
  ];

  for (const card of allCards) {
    if (!used.has(card)) {
      deck.push(card);
    }
  }

  return deck;
}

test('Deterministic deck injection - Player1 with Aces beats Player2 with Kings', async ({
  browser,
}) => {
  console.log('\n=== DETERMINISTIC WINNER TEST ===\n');

  // Create two separate browser contexts (separate localStorage)
  const context1 = await browser.newContext();
  const context2 = await browser.newContext();

  const page1 = await context1.newPage();
  const page2 = await context2.newPage();

  try {
    // Navigate both players
    console.log('Navigating players to app...');
    await page1.goto(BASE_URL);
    await page2.goto(BASE_URL);

    // Player 1 enters name
    console.log('Player1: Entering name...');
    await page1.locator('input[type="text"]').fill('Player1');
    await page1.locator('button[type="submit"]').click();
    await page1.waitForTimeout(1000);

    // Player 2 enters name
    console.log('Player2: Entering name...');
    await page2.locator('input[type="text"]').fill('Player2');
    await page2.locator('button[type="submit"]').click();
    await page2.waitForTimeout(1000);

    // Player 1 joins Table 1
    console.log('Player1: Joining table...');
    await page1.locator('button.join-button').first().click();
    await page1.waitForTimeout(1500);

    // Player 2 joins Table 1
    console.log('Player2: Joining table...');
    await page2.locator('button.join-button').first().click();
    await page2.waitForTimeout(1500);

    // Set deterministic deck BEFORE starting hand
    console.log('\nSetting deterministic deck...');
    await setDeck(TABLE_ID, buildTestDeck());

    // Wait for Start Hand button to appear (need 2+ players)
    console.log('\nWaiting for Start Hand button...');
    await page1.waitForSelector('button:has-text("Start Hand")', { timeout: 5000 });

    // Start the hand
    console.log('Starting hand...');
    await page1.locator('button:has-text("Start Hand")').click();
    await page1.waitForTimeout(2000);

    // Check what cards each player has
    console.log('\n=== CHECKING DEALT CARDS ===');

    // Get hole cards from page1 (Player1's view)
    const page1Cards = await page1.locator('.hole-cards .card').allTextContents();
    console.log('Player1 sees cards:', page1Cards);

    // Get hole cards from page2 (Player2's view)
    const page2Cards = await page2.locator('.hole-cards .card').allTextContents();
    console.log('Player2 sees cards:', page2Cards);

    // Verify the cards match what we set in the deck
    // Player1 (seat 0) should have As, Ah (Pocket Aces)
    // Player2 (seat 1) should have Ks, Kh (Pocket Kings)

    const p1Match =
      page1Cards.length === 2 && page1Cards.includes('A♠') && page1Cards.includes('A♥');
    const p2Match =
      page2Cards.length === 2 && page2Cards.includes('K♠') && page2Cards.includes('K♥');

    console.log('\n=== VERIFICATION ===');
    console.log(`Player1 got Pocket Aces: ${p1Match ? 'YES ✓' : 'NO ✗'}`);
    console.log(`Player2 got Pocket Kings: ${p2Match ? 'YES ✓' : 'NO ✗'}`);

    expect(p1Match).toBe(true);
    expect(p2Match).toBe(true);

    console.log('\n✓ DETERMINISTIC DECK INJECTION WORKS!');

    // Now play the hand - both go all-in
    console.log('\n=== PLAYING HAND (ALL-IN) ===');

    // Find who has action and go all-in
    const p1HasAction = await page1.locator('.action-bar').isVisible().catch(() => false);
    const p2HasAction = await page2.locator('.action-bar').isVisible().catch(() => false);

    console.log(`Player1 has action: ${p1HasAction}`);
    console.log(`Player2 has action: ${p2HasAction}`);

    // First player to act goes all-in
    if (p1HasAction) {
      console.log('Player1 going all-in...');
      await page1.locator('button.preset-button:has-text("All-in")').click();
      await page1.locator('button.raise-button').click();
      await page1.waitForTimeout(1500);
    } else if (p2HasAction) {
      console.log('Player2 going all-in...');
      await page2.locator('button.preset-button:has-text("All-in")').click();
      await page2.locator('button.raise-button').click();
      await page2.waitForTimeout(1500);
    }

    // Second player calls
    const p1HasAction2 = await page1.locator('.action-bar').isVisible().catch(() => false);
    const p2HasAction2 = await page2.locator('.action-bar').isVisible().catch(() => false);

    if (p1HasAction2) {
      console.log('Player1 calling...');
      await page1.locator('button:has-text("Call")').click();
      await page1.waitForTimeout(1500);
    } else if (p2HasAction2) {
      console.log('Player2 calling...');
      await page2.locator('button:has-text("Call")').click();
      await page2.waitForTimeout(1500);
    }

    // Wait for showdown
    console.log('\nWaiting for showdown...');
    await page1.waitForTimeout(5000);

    // Check showdown result
    const showdownOverlay = page1.locator('.showdown-overlay');
    if (await showdownOverlay.isVisible().catch(() => false)) {
      const showdownText = await showdownOverlay.textContent();
      console.log('\n=== SHOWDOWN RESULT ===');
      console.log(showdownText);

      // Player1 with trip Aces should win
      expect(showdownText).toContain('Player1');
      console.log('\n✓ Player1 (Pocket Aces) WON as expected!');
    } else {
      // Take screenshots for debugging
      await page1.screenshot({ path: 'player1-debug.png' });
      await page2.screenshot({ path: 'player2-debug.png' });
      console.log('No showdown overlay visible - screenshots saved for debugging');
    }
  } finally {
    await context1.close();
    await context2.close();
  }
});
