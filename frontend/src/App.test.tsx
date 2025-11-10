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
        // Wait for at least 1 socket to be created with the token
        expect(createdMockSockets.length).toBeGreaterThanOrEqual(1);
      });

      // The socket should have the token in the URL
      const mockSocket = createdMockSockets[createdMockSockets.length - 1];
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

  describe('TestApp_InvalidToken_ShowsPrompt', () => {
    it('should clear token and show NamePrompt when server returns invalid token error', async () => {
      // Start with an invalid token in localStorage
      localStorage.setItem('poker_session_token', 'invalid-expired-token');

      render(<App />);

      // Initially, prompt should not be shown (token exists)
      expect(screen.queryByText('Enter Your Name')).not.toBeInTheDocument();

      // Wait for socket to be created
      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[0];

      // Simulate error message from server about invalid token
      const errorMessage = JSON.stringify({
        type: 'error',
        payload: { message: 'Invalid or expired token' },
      });

      mockSocket.simulateMessage(errorMessage);

      // Should clear the token from localStorage
      await waitFor(() => {
        const savedToken = localStorage.getItem('poker_session_token');
        expect(savedToken).toBeNull();
      });

      // Should show the name prompt
      await waitFor(() => {
        expect(screen.getByText('Enter Your Name')).toBeInTheDocument();
      });
    });

    it('should allow user to create new session after invalid token', async () => {
      // Start with an invalid token
      localStorage.setItem('poker_session_token', 'bad-token');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[0];

      // Simulate invalid token error
      mockSocket.simulateMessage(
        JSON.stringify({
          type: 'error',
          payload: { message: 'Invalid or expired token' },
        })
      );

      // Wait for prompt to show
      await waitFor(() => {
        expect(screen.getByText('Enter Your Name')).toBeInTheDocument();
      });

      // Enter a new name
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Bob' } });
      fireEvent.click(button);

      // Simulate session created
      mockSocket.simulateMessage(
        JSON.stringify({
          type: 'session_created',
          payload: { token: 'new-valid-token', name: 'Bob' },
        })
      );

      // Should save new token
      await waitFor(() => {
        const savedToken = localStorage.getItem('poker_session_token');
        expect(savedToken).toBe('new-valid-token');
      });

      // Should hide prompt
      await waitFor(() => {
        expect(screen.queryByText('Enter Your Name')).not.toBeInTheDocument();
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
        payload: { name: 'Dave', tableId: 'table-1', seatIndex: 2 },
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
        payload: { name: 'Eve', tableId: 'table-2', seatIndex: 1 },
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
        payload: { name: 'Frank', tableId: 'table-3', seatIndex: 0 },
      });

      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.queryByText('Enter Your Name')).not.toBeInTheDocument();
      });
    });
  });

  describe('TestApp_LobbyView_Integration', () => {
    it('should render LobbyView after session created', async () => {
      render(<App />);

      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Alice' } });

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      fireEvent.click(button);

      const mockSocket = createdMockSockets[0];

      const sessionMessage = JSON.stringify({
        type: 'session_created',
        payload: { token: 'token-lobby', name: 'Alice' },
      });

      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });
    });

    it('should display tables from lobby_state message', async () => {
      localStorage.setItem('poker_session_token', 'lobby-token');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      // Send session_restored first
      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'Alice' },
      });
      mockSocket.simulateMessage(sessionMessage);

      // Wait for lobby to render
      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });

      // Send lobby_state with tables
      const lobbyMessage = JSON.stringify({
        type: 'lobby_state',
        payload: [
          {
            id: 'table-1',
            name: 'Table 1',
            seats_occupied: 2,
            max_seats: 6,
          },
          {
            id: 'table-2',
            name: 'Table 2',
            seats_occupied: 4,
            max_seats: 6,
          },
        ],
      });

      mockSocket.simulateMessage(lobbyMessage);

      await waitFor(() => {
        expect(screen.getByText('Table 1')).toBeInTheDocument();
        expect(screen.getByText('Table 2')).toBeInTheDocument();
        expect(screen.getByText('2/6')).toBeInTheDocument();
        expect(screen.getByText('4/6')).toBeInTheDocument();
      });
    });

    it('should update table seat counts when lobby_state updates', async () => {
      localStorage.setItem('poker_session_token', 'lobby-update-token');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      // Send session_restored first
      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'Bob' },
      });
      mockSocket.simulateMessage(sessionMessage);

      // Wait for lobby to render
      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });

      // Send initial lobby_state
      const initialLobbyMessage = JSON.stringify({
        type: 'lobby_state',
        payload: [
          {
            id: 'table-1',
            name: 'Table 1',
            seats_occupied: 1,
            max_seats: 6,
          },
        ],
      });

      mockSocket.simulateMessage(initialLobbyMessage);

      await waitFor(() => {
        expect(screen.getByText('1/6')).toBeInTheDocument();
      });

      // Send updated lobby_state
      const updatedLobbyMessage = JSON.stringify({
        type: 'lobby_state',
        payload: [
          {
            id: 'table-1',
            name: 'Table 1',
            seats_occupied: 5,
            max_seats: 6,
          },
        ],
      });

      mockSocket.simulateMessage(updatedLobbyMessage);

      await waitFor(() => {
        expect(screen.getByText('5/6')).toBeInTheDocument();
      });
    });
  });

  describe('TestAppHandlesSeatAssignedMessage', () => {
    it('should switch to table view when receiving seat_assigned message', async () => {
      localStorage.setItem('poker_session_token', 'table-token');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      // Send session_restored first
      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'Alice' },
      });
      mockSocket.simulateMessage(sessionMessage);

      // Wait for lobby to render
      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });

      // Send seat_assigned message
      const seatMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 2,
        },
      });
      mockSocket.simulateMessage(seatMessage);

      // Wait for table view to render
      await waitFor(() => {
        expect(screen.queryByText('Lobby')).not.toBeInTheDocument();
        expect(screen.getByText(/Table: table-1/i)).toBeInTheDocument();
      });
    });

    it('should display table ID and seat info from seat_assigned', async () => {
      localStorage.setItem('poker_session_token', 'table-token-2');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'Bob' },
      });
      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });

      const seatMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-42',
          seatIndex: 3,
        },
      });
      mockSocket.simulateMessage(seatMessage);

      await waitFor(() => {
        expect(screen.getByText(/Table: table-42/i)).toBeInTheDocument();
      });
    });
  });

  describe('TestAppHandlesSeatClearedMessage', () => {
    it('should switch back to lobby view when receiving seat_cleared message', async () => {
      localStorage.setItem('poker_session_token', 'return-lobby-token');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'Charlie' },
      });
      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });

      // Send seat_assigned to go to table
      const seatMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 1,
        },
      });
      mockSocket.simulateMessage(seatMessage);

      await waitFor(() => {
        expect(screen.getByText(/Table: table-1/i)).toBeInTheDocument();
      });

      // Send seat_cleared to return to lobby
      const clearMessage = JSON.stringify({
        type: 'seat_cleared',
        payload: {},
      });
      mockSocket.simulateMessage(clearMessage);

      await waitFor(() => {
        expect(screen.queryByText(/Table:/i)).not.toBeInTheDocument();
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });
    });

    it('test_bust_out_flow_complete: should complete full bust-out flow from name prompt through showdown to lobby', async () => {
      render(<App />);

      // Step 1: Verify NamePrompt is shown
      expect(screen.getByText('Enter Your Name')).toBeInTheDocument();

      // Step 2: Enter name and submit
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'BustOutPlayer' } });

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      fireEvent.click(button);

      // Step 3: Receive session_created and verify lobby appears
      const mockSocket = createdMockSockets[0];

      const sessionMessage = JSON.stringify({
        type: 'session_created',
        payload: { token: 'bust-out-token', name: 'BustOutPlayer' },
      });

      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });

      // Step 4: Receive lobby_state with available table
      const lobbyMessage = JSON.stringify({
        type: 'lobby_state',
        payload: [
          {
            id: 'table-1',
            name: 'Table 1',
            seats_occupied: 1,
            max_seats: 6,
          },
        ],
      });

      mockSocket.simulateMessage(lobbyMessage);

      await waitFor(() => {
        expect(screen.getByText('Table 1')).toBeInTheDocument();
      });

      // Step 5: Join table by clicking Join button
      const joinButton = screen.getByRole('button', { name: 'Join' });
      fireEvent.click(joinButton);

      // Step 6: Receive seat_assigned and switch to table view
      const seatMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 1,
        },
      });

      mockSocket.simulateMessage(seatMessage);

      await waitFor(() => {
        expect(screen.getByText(/Table: table-1/i)).toBeInTheDocument();
      });

      // Step 7: Receive table_state with game in progress
      const tableStateMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            { index: 0, playerName: 'Alice', status: 'occupied', stack: 1500 },
            {
              index: 1,
              playerName: 'BustOutPlayer',
              status: 'occupied',
              stack: 100,
            },
          ],
          handInProgress: true,
          dealerSeat: 0,
          smallBlindSeat: 1,
          bigBlindSeat: 0,
          pot: 150,
        },
      });

      mockSocket.simulateMessage(tableStateMessage);

      await waitFor(() => {
        // Verify we can see other player and our stack
        expect(screen.getByText('Alice')).toBeInTheDocument();
        expect(screen.getByText(/100/)).toBeInTheDocument();
      });

      // Step 8: Simulate showdown where player busts out (their stack goes to 0)
      const showdownMessage = JSON.stringify({
        type: 'showdown_result',
        payload: {
          winnerSeats: [0],
          winningHand: 'Pair of Kings',
          potAmount: 250,
          amountsWon: {
            '0': 250,
          },
        },
      });

      mockSocket.simulateMessage(showdownMessage);

      await waitFor(() => {
        // Verify showdown results are displayed
        expect(screen.getByText(/Pair of Kings/i)).toBeInTheDocument();
      });

      // Step 9: Receive hand_complete message
      const handCompleteMessage = JSON.stringify({
        type: 'hand_complete',
        payload: {
          message: 'Hand complete! Alice wins the pot.',
        },
      });

      mockSocket.simulateMessage(handCompleteMessage);

      await waitFor(() => {
        expect(screen.getByText(/Hand complete/i)).toBeInTheDocument();
      });

      // Step 10: Receive seat_cleared message indicating player has busted out
      const seatClearedMessage = JSON.stringify({
        type: 'seat_cleared',
        payload: {},
      });

      mockSocket.simulateMessage(seatClearedMessage);

      // Step 11: Verify return to lobby view
      await waitFor(() => {
        expect(screen.queryByText(/Table:/i)).not.toBeInTheDocument();
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });

      // Step 12: Verify session persists (can rejoin other tables)
      const savedToken = localStorage.getItem('poker_session_token');
      expect(savedToken).toBe('bust-out-token');
    });
  });

  describe('TestAppJoinTableIntegration', () => {
    it('should send join_table message when joining a table from lobby', async () => {
      localStorage.setItem('poker_session_token', 'join-token');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      // Send session_restored
      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'Dave' },
      });
      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });

      // Send lobby_state with a table
      const lobbyMessage = JSON.stringify({
        type: 'lobby_state',
        payload: [
          {
            id: 'table-1',
            name: 'Table 1',
            seats_occupied: 2,
            max_seats: 6,
          },
        ],
      });
      mockSocket.simulateMessage(lobbyMessage);

      await waitFor(() => {
        expect(screen.getByText('Table 1')).toBeInTheDocument();
      });

      // Click Join button
      const joinButton = screen.getByRole('button', { name: 'Join' });
      fireEvent.click(joinButton);

      // Check that join_table message was sent
      await waitFor(() => {
        const sentMessages = mockSocket.messages;
        const joinMessage = sentMessages.find((msg) => {
          try {
            const parsed = JSON.parse(msg);
            return parsed.type === 'join_table';
          } catch {
            return false;
          }
        });
        expect(joinMessage).toBeDefined();
      });

      // Parse the join message to verify tableId
      const joinMessage = mockSocket.messages
        .map((msg) => {
          try {
            return JSON.parse(msg);
          } catch {
            return null;
          }
        })
        .find((msg) => msg?.type === 'join_table');

      expect(joinMessage?.payload?.tableId).toBe('table-1');
    });

    it('should complete full flow: lobby -> join -> table view -> leave -> lobby', async () => {
      localStorage.setItem('poker_session_token', 'flow-token');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      // Step 1: Restore session and show lobby
      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'Eve' },
      });
      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });

      // Step 2: Send lobby_state
      const lobbyMessage = JSON.stringify({
        type: 'lobby_state',
        payload: [
          {
            id: 'table-1',
            name: 'Table 1',
            seats_occupied: 1,
            max_seats: 6,
          },
        ],
      });
      mockSocket.simulateMessage(lobbyMessage);

      await waitFor(() => {
        expect(screen.getByText('Table 1')).toBeInTheDocument();
      });

      // Step 3: Click Join button
      const joinButton = screen.getByRole('button', { name: 'Join' });
      fireEvent.click(joinButton);

      // Step 4: Receive seat_assigned and switch to table
      const seatMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 2,
        },
      });
      mockSocket.simulateMessage(seatMessage);

      await waitFor(() => {
        expect(screen.getByText(/Table: table-1/i)).toBeInTheDocument();
      });

      // Step 5: Click Leave Table button
      const leaveButton = screen.getByRole('button', { name: /Leave Table/i });
      fireEvent.click(leaveButton);

      // Step 6: Receive seat_cleared and return to lobby
      const clearMessage = JSON.stringify({
        type: 'seat_cleared',
        payload: {},
      });
      mockSocket.simulateMessage(clearMessage);

      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
        expect(screen.queryByText(/Table:/i)).not.toBeInTheDocument();
      });
    });

    it('should update seat information when receiving table_state message', async () => {
      // Set up localStorage with token
      localStorage.setItem('sessionToken', 'test-token-123');

      render(<App />);

      // Step 1: Simulate session restored
      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: {
          name: 'TestPlayer',
        },
      });
      const mockSocket = createdMockSockets[createdMockSockets.length - 1];
      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.getByText(/TestPlayer/)).toBeInTheDocument();
      });

      // Step 2: Simulate seat assignment
      const seatMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 2,
        },
      });
      mockSocket.simulateMessage(seatMessage);

      await waitFor(() => {
        expect(screen.getByText(/Table: table-1/i)).toBeInTheDocument();
      });

      // Step 3: Simulate table_state update with other players
      const tableStateMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            {
              index: 0,
              playerName: 'Alice',
              status: 'occupied',
            },
            {
              index: 1,
              playerName: 'Bob',
              status: 'occupied',
            },
            {
              index: 2,
              playerName: 'TestPlayer',
              status: 'occupied',
            },
            {
              index: 3,
              playerName: null,
              status: 'empty',
            },
            {
              index: 4,
              playerName: null,
              status: 'empty',
            },
            {
              index: 5,
              playerName: null,
              status: 'empty',
            },
          ],
        },
      });
      mockSocket.simulateMessage(tableStateMessage);

      // Verify other players' names are displayed
      await waitFor(() => {
        expect(screen.getByText('Alice')).toBeInTheDocument();
        expect(screen.getByText('Bob')).toBeInTheDocument();
        expect(screen.getByText('TestPlayer')).toBeInTheDocument();
      });
    });

    it('should preserve stack information from table_state', async () => {
      localStorage.setItem('poker_session_token', 'stack-test-token');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      // Session restored
      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'StackPlayer' },
      });
      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });

      // Seat assigned
      const seatMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 1,
        },
      });
      mockSocket.simulateMessage(seatMessage);

      await waitFor(() => {
        expect(screen.getByText(/Table: table-1/i)).toBeInTheDocument();
      });

      // Table state with stack information
      const tableStateMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            {
              index: 0,
              playerName: 'Alice',
              status: 'occupied',
              stack: 1000,
            },
            {
              index: 1,
              playerName: 'StackPlayer',
              status: 'occupied',
              stack: 2000,
            },
            {
              index: 2,
              playerName: null,
              status: 'empty',
            },
            {
              index: 3,
              playerName: null,
              status: 'empty',
            },
            {
              index: 4,
              playerName: null,
              status: 'empty',
            },
            {
              index: 5,
              playerName: null,
              status: 'empty',
            },
          ],
        },
      });
      mockSocket.simulateMessage(tableStateMessage);

      // Verify stacks are displayed
      await waitFor(() => {
        expect(screen.getByText(/1000/)).toBeInTheDocument();
        expect(screen.getByText(/2000/)).toBeInTheDocument();
      });
    });

    it('should display dealer and blind badges when gameState is updated', async () => {
      localStorage.setItem('poker_session_token', 'dealer-test-token');

      render(<App />);

      await waitFor(() => {
        expect(createdMockSockets.length).toBeGreaterThan(0);
      });

      const mockSocket = createdMockSockets[createdMockSockets.length - 1];

      // Session restored
      const sessionMessage = JSON.stringify({
        type: 'session_restored',
        payload: { name: 'DealerPlayer' },
      });
      mockSocket.simulateMessage(sessionMessage);

      await waitFor(() => {
        expect(screen.getByText('Lobby')).toBeInTheDocument();
      });

      // Seat assigned
      const seatMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 2,
        },
      });
      mockSocket.simulateMessage(seatMessage);

      await waitFor(() => {
        expect(screen.getByText(/Table: table-1/i)).toBeInTheDocument();
      });

      // Table state
      const tableStateMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            {
              index: 0,
              playerName: 'Alice',
              status: 'occupied',
              stack: 1000,
            },
            {
              index: 1,
              playerName: 'Bob',
              status: 'occupied',
              stack: 1500,
            },
            {
              index: 2,
              playerName: 'DealerPlayer',
              status: 'occupied',
              stack: 2000,
            },
            {
              index: 3,
              playerName: null,
              status: 'empty',
            },
            {
              index: 4,
              playerName: null,
              status: 'empty',
            },
            {
              index: 5,
              playerName: null,
              status: 'empty',
            },
          ],
        },
      });
      mockSocket.simulateMessage(tableStateMessage);

      // Hand started message sets dealer, SB, BB
      const handStartedMessage = JSON.stringify({
        type: 'hand_started',
        payload: JSON.stringify({
          dealerSeat: 0,
          smallBlindSeat: 1,
          bigBlindSeat: 2,
        }),
      });
      mockSocket.simulateMessage(handStartedMessage);

      // Verify badges are displayed
      await waitFor(() => {
        // Check for dealer badge (D)
        const dealerBadges = screen.getAllByText('D');
        expect(dealerBadges.length).toBeGreaterThan(0);
        // Check for SB badge
        const sbBadges = screen.getAllByText('SB');
        expect(sbBadges.length).toBeGreaterThan(0);
        // Check for BB badge
        const bbBadges = screen.getAllByText('BB');
        expect(bbBadges.length).toBeGreaterThan(0);
      });
    });
  });
});
