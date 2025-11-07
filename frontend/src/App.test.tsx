import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, beforeEach, afterEach } from 'vitest';

import '@testing-library/jest-dom';
import App from './App';

describe('App', () => {
  let createdMockSockets: MockWebSocket[] = [];
  let originalWebSocket: typeof WebSocket;

  // WebSocket constants
  const CONNECTING = 0;
  const OPEN = 1;
  const CLOSED = 3;

  // Mock WebSocket implementation
  class MockWebSocket {
    url: string;
    onopen: ((event: Event) => void) | null = null;
    onclose: ((event: CloseEvent) => void) | null = null;
    onerror: ((event: Event) => void) | null = null;
    onmessage: ((event: MessageEvent) => void) | null = null;
    readyState: number = CONNECTING;
    messages: string[] = [];

    constructor(url: string) {
      this.url = url;
      createdMockSockets.push(this);
      // Auto-open the connection immediately in tests
      setImmediate(() => {
        this.simulateOpen();
      });
    }

    send(data: string) {
      this.messages.push(data);
    }

    close() {
      this.readyState = CLOSED;
      this.onclose?.(new CloseEvent('close'));
    }

    simulateOpen() {
      this.readyState = OPEN;
      this.onopen?.(new Event('open'));
    }

    simulateMessage(data: string) {
      this.onmessage?.(new MessageEvent('message', { data }));
    }

    simulateError() {
      this.onerror?.(new Event('error'));
    }
  }

  beforeEach(() => {
    createdMockSockets = [];
    localStorage.clear();
    originalWebSocket = global.WebSocket as typeof WebSocket;
    global.WebSocket = MockWebSocket as typeof WebSocket;
  });

  afterEach(() => {
    global.WebSocket = originalWebSocket;
    localStorage.clear();
  });

  describe('TestApp_NoToken_ShowsPrompt', () => {
    it('should show NamePrompt when no token in localStorage', () => {
      render(<App />);

      expect(screen.getByText('Enter Your Name')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Your name')).toBeInTheDocument();
    });

    it('should hide other app content when showing prompt', () => {
      render(<App />);

      expect(screen.getByText('Enter Your Name')).toBeInTheDocument();
      // App main content should be covered by overlay
      const overlay = document.querySelector('.name-prompt-overlay');
      expect(overlay).toBeInTheDocument();
    });
  });

  describe('TestApp_ValidToken_AutoConnects', () => {
    it('should not show NamePrompt when token exists in localStorage', () => {
      localStorage.setItem('poker_session_token', 'valid-token-123');

      render(<App />);

      expect(screen.queryByText('Enter Your Name')).not.toBeInTheDocument();
    });

    it('should include token in WebSocket URL when connecting with saved token', async () => {
      localStorage.setItem('poker_session_token', 'saved-token-xyz');

      render(<App />);

      await waitFor(() => {
        // Wait for at least 2 sockets (initial undefined, then with token)
        expect(createdMockSockets.length).toBeGreaterThanOrEqual(2);
      });

      // The second socket should have the token
      const mockSocket = createdMockSockets[1];
      expect(mockSocket.url).toContain('saved-token-xyz');
    });

    it('should auto-connect to WebSocket with saved token', async () => {
      localStorage.setItem('poker_session_token', 'auto-token');

      render(<App />);

      await waitFor(() => {
        expect(screen.getByText('Connected')).toBeInTheDocument();
      });
    });
  });

  describe('TestApp_SessionCreated_SavesToken', () => {
    it('should save token to localStorage on session_created', async () => {
      render(<App />);

      // Enter a name and submit
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Alice' } });

      // Wait for socket to connect first
      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      fireEvent.click(button);

      // Get the mock socket
      const mockSocket = createdMockSockets[0];

      // Simulate session_created response
      const sessionMessage = JSON.stringify({
        type: 'session_created',
        payload: { token: 'new-session-token', name: 'Alice' },
      });

      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        const savedToken = localStorage.getItem('poker_session_token');
        expect(savedToken).toBe('new-session-token');
      });
    });

    it('should hide NamePrompt after session created', async () => {
      render(<App />);

      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Bob' } });

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      fireEvent.click(button);

      const mockSocket = createdMockSockets[0];

      const sessionMessage = JSON.stringify({
        type: 'session_created',
        payload: { token: 'token-456', name: 'Bob' },
      });

      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.queryByText('Enter Your Name')).not.toBeInTheDocument();
      });
    });

    it('should show player name after session created', async () => {
      render(<App />);

      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Charlie' } });

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      fireEvent.click(button);

      const mockSocket = createdMockSockets[0];

      const sessionMessage = JSON.stringify({
        type: 'session_created',
        payload: { token: 'token-789', name: 'Charlie' },
      });

      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        // Check in the header player name area
        const playerNameElement = document.querySelector('.player-name');
        expect(playerNameElement?.textContent).toContain('Charlie');
      });
    });
  });

  describe('TestApp_SessionRestored_WithTable', () => {
    it('should display welcome message on session restored', async () => {
      localStorage.setItem('poker_session_token', 'restored-token');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      // Use the latest socket (after token update triggers reconnection)
      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'Dave', tableID: 'table-1', seatIndex: 2 },
      });

      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        // Check in the header player name area
        const playerNameElement = document.querySelector('.player-name');
        expect(playerNameElement?.textContent).toContain('Dave');
      });
    });

    it('should show player name from restored session', async () => {
      localStorage.setItem('poker_session_token', 'restored-token-2');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'Eve', tableID: 'table-2', seatIndex: 1 },
      });

      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        // Check in the header player name area
        const playerNameElement = document.querySelector('.player-name');
        expect(playerNameElement).toBeInTheDocument();
        expect(playerNameElement?.textContent).toContain('Eve');
      });
    });

    it('should not show NamePrompt on session restored', async () => {
      localStorage.setItem('poker_session_token', 'restored-token-3');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      // Use the latest socket (after token update triggers reconnection)
      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'Frank', tableID: 'table-3', seatIndex: 0 },
      });

      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.queryByText('Enter Your Name')).not.toBeInTheDocument();
      });
    });
  });
});
