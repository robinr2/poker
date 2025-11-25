/**
 * Browser launch helpers for e2e tests
 *
 * Provides consistent browser launching that respects headed/headless mode.
 *
 * Usage:
 * - Headless (default, fast): npx playwright test e2e/
 * - Headed (visual):          npx playwright test e2e/ --headed
 */

import { chromium, Browser, BrowserContext, Page, LaunchOptions } from '@playwright/test';

/**
 * Check if tests are running in headed mode
 * Playwright sets this when --headed flag is used
 */
export function isHeadedMode(): boolean {
  // Check multiple ways to detect headed mode:
  // 1. HEADED env var explicitly set
  // 2. PWDEBUG for debug mode
  // 3. Check process.argv for --headed flag
  const hasHeadedFlag = process.argv.includes('--headed');
  return process.env.HEADED === 'true' || process.env.PWDEBUG === '1' || hasHeadedFlag;
}

/**
 * Get browser launch options based on current mode
 *
 * - Headed mode: visible browser with slowMo for debugging
 * - Headless mode: invisible browser, no slowMo for speed
 */
export function getBrowserLaunchOptions(): LaunchOptions {
  const headed = isHeadedMode();

  return {
    headless: !headed,
    slowMo: headed ? 300 : 0,
    args: headed ? ['--window-size=950,1200'] : [],
  };
}

/**
 * Launch a browser with appropriate settings for the current mode
 *
 * @returns Promise<Browser> - Launched browser instance
 *
 * @example
 * const browser = await launchBrowser();
 * const context = await browser.newContext();
 * const page = await context.newPage();
 */
export async function launchBrowser(): Promise<Browser> {
  const options = getBrowserLaunchOptions();
  return chromium.launch(options);
}

/**
 * Window positions for multi-player tests
 * Positions windows side-by-side on a wide monitor setup
 */
export const WINDOW_POSITIONS = [
  { x: 0, y: 40 }, // Player 1 - left
  { x: 950, y: 40 }, // Player 2 - middle
  { x: 1900, y: 40 }, // Player 3 - right
  { x: 0, y: 700 }, // Player 4 - bottom left
  { x: 950, y: 700 }, // Player 5 - bottom middle
  { x: 1900, y: 700 }, // Player 6 - bottom right
];

export const WINDOW_WIDTH = 950;
export const WINDOW_HEIGHT = 1200;

/**
 * Position a browser window (only works in headed mode)
 * In headless mode, this is a no-op
 *
 * @param context - Browser context
 * @param page - Page to position
 * @param index - Player index (0-5) to determine position
 */
export async function positionWindow(
  context: BrowserContext,
  page: Page,
  index: number
): Promise<void> {
  // Window positioning only works in headed mode
  if (!isHeadedMode()) {
    return;
  }

  const position = WINDOW_POSITIONS[index] || WINDOW_POSITIONS[0];

  try {
    const client = await context.newCDPSession(page);
    const { windowId } = await client.send('Browser.getWindowForTarget');
    await client.send('Browser.setWindowBounds', {
      windowId,
      bounds: {
        left: position.x,
        top: position.y,
        width: WINDOW_WIDTH,
        height: WINDOW_HEIGHT,
      },
    });
  } catch {
    // Ignore errors - CDP might not be available in some configurations
  }
}
