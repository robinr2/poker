import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for poker application testing
 * - Headed mode for visual debugging
 * - Extended timeouts to observe test execution
 * - Multiple browser contexts for simulating different players
 */
export default defineConfig({
  testDir: './e2e',
  
  // Run tests in serial to avoid port conflicts
  fullyParallel: false,
  workers: 1,
  
  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,
  
  // No retries - we want to see failures clearly
  retries: 0,
  
  // Reporter to use
  reporter: [
    ['html'],
    ['list']
  ],
  
  // Shared settings for all projects
  use: {
    // Base URL for the application
    baseURL: 'http://localhost:8080',
    
    // Collect trace on failure
    trace: 'on-first-retry',
    
    // Screenshots on failure
    screenshot: 'only-on-failure',
    
    // Video on failure
    video: 'retain-on-failure',
    
    // Slow down operations so we can see what's happening
    actionTimeout: 10000,
    navigationTimeout: 10000,
  },
  
  // Global timeout for each test - 5 minutes should be plenty
  timeout: 300000,
  
  // Expect timeout
  expect: {
    timeout: 10000,
  },
  
  // Configure projects for different browsers
  projects: [
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        // Run in headed mode so we can watch the tests
        headless: false,
        // Slow down by 500ms per action to make it visible
        slowMo: 500,
        // Viewport sized for dual 1920x1080 screens (950px width per window)
        viewport: { width: 950, height: 1200 },
      },
    },
  ],
  
  // Don't start the dev server - assume it's already running
  // You can uncomment this if you want Playwright to start the server
  // webServer: {
  //   command: 'cd frontend && npm run dev',
  //   url: 'http://localhost:5173',
  //   reuseExistingServer: !process.env.CI,
  //   timeout: 120000,
  // },
});
