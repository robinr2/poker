import { test, chromium } from '@playwright/test';

/**
 * Single client test to check if hand auto-starts
 */
test('single player - check for auto-start', async () => {
  const browser = await chromium.launch({
    headless: false,
    slowMo: 1000,
  });
  
  const context = await browser.newContext({
    viewport: { width: 1200, height: 900 },
  });
  
  const page = await context.newPage();
  
  console.log('=== Single Player Test ===');
  console.log('1. Navigating to app...');
  await page.goto('http://localhost:8080');
  await page.waitForSelector('.app', { timeout: 10000 });
  
  console.log('2. Entering name...');
  const nameInput = page.locator('input[type="text"]');
  await nameInput.fill('SinglePlayer');
  await page.locator('button[type="submit"]').click();
  await page.waitForTimeout(2000);
  
  console.log('3. In lobby, checking tables...');
  const lobbyText = await page.locator('body').textContent();
  console.log('Lobby text:', lobbyText?.substring(0, 200));
  
  console.log('4. Joining first table...');
  const joinButton = page.locator('button.join-button').first();
  await joinButton.click();
  await page.waitForTimeout(3000);
  
  console.log('5. Checking table state after join...');
  const tableText = await page.locator('body').textContent();
  console.log('Table text:', tableText?.substring(0, 300));
  
  // Check for cards, pot, start button
  const hasCards = tableText?.includes('♠') || tableText?.includes('♥') || tableText?.includes('♦') || tableText?.includes('♣');
  const hasPot = tableText?.includes('Pot:');
  const hasStartButton = await page.locator('button:has-text("Start Hand")').count() > 0;
  
  console.log('\n=== Results ===');
  console.log('Cards dealt:', hasCards);
  console.log('Pot exists:', hasPot);
  console.log('Start Hand button visible:', hasStartButton);
  
  if (hasCards || hasPot) {
    console.log('⚠ Hand AUTO-STARTED with single player!');
  } else if (hasStartButton) {
    console.log('✓ No auto-start, Start Hand button is available');
  } else {
    console.log('? No auto-start, but also no Start Hand button');
  }
  
  // Wait for observation
  console.log('\nPausing for 15 seconds for observation...');
  await page.waitForTimeout(15000);
  
  await context.close();
  await browser.close();
});
