import { test } from '@playwright/test';
import { launchBrowser, WINDOW_WIDTH, WINDOW_HEIGHT } from './helpers/browser-helpers';

/**
 * Simple diagnostic test to see what's actually loading on the page
 */
test('diagnostic - check what loads on page', async () => {
  const browser = await launchBrowser();
  
  const context = await browser.newContext({
    viewport: { width: WINDOW_WIDTH, height: WINDOW_HEIGHT },
  });
  
  const page = await context.newPage();
  
  console.log('Navigating to http://localhost:8080...');
  await page.goto('http://localhost:8080');
  
  // Wait for page to load
  await page.waitForLoadState('domcontentloaded');
  await page.waitForTimeout(5000);
  
  // Log the entire page HTML
  const html = await page.content();
  console.log('=== PAGE HTML ===');
  console.log(html);
  
  // Log visible text
  const bodyText = await page.locator('body').textContent();
  console.log('\n=== VISIBLE TEXT ===');
  console.log(bodyText);
  
  // Check for specific elements
  const hasNameInput = await page.locator('input[type="text"]').count();
  const hasLobby = await page.locator('.lobby-view').count();
  const hasNamePrompt = await page.locator('.name-prompt-overlay').count();
  
  console.log('\n=== ELEMENT COUNTS ===');
  console.log('Name inputs:', hasNameInput);
  console.log('Lobby views:', hasLobby);
  console.log('Name prompts:', hasNamePrompt);
  
  // Take a screenshot
  await page.screenshot({ path: 'diagnostic-screenshot.png', fullPage: true });
  console.log('\nScreenshot saved to diagnostic-screenshot.png');
  
  // Keep browser open for observation
  await page.waitForTimeout(10000);
  
  await context.close();
  await browser.close();
});
