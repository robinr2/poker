import { renderHook, act, waitFor } from '@testing-library/react';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

import type { WebSocketService } from '../services/WebSocketService';

import { useWebSocket } from './useWebSocket';

// Mock instance that will be reused across tests
let mockServiceInstance: Partial<WebSocketService>;
let mockStatusCallbacks: ((status: string) => void)[] = [];
let mockMessageCallbacks: ((data: string) => void)[] = [];

// Mock the WebSocketService module before any imports
vi.mock('../services/WebSocketService', () => {
  return {
    WebSocketService: class MockWebSocketService {
      constructor() {
        return mockServiceInstance;
      }
    },
  };
});

// Helper function to create a fresh mock
function createMockService(): Partial<WebSocketService> {
  mockStatusCallbacks = [];
  mockMessageCallbacks = [];

  return {
    connect: vi.fn().mockResolvedValue(undefined),
    disconnect: vi.fn(),
    send: vi.fn(),
    onMessage: vi.fn((callback) => {
      mockMessageCallbacks.push(callback);
      return () => {
        const index = mockMessageCallbacks.indexOf(callback);
        if (index > -1) mockMessageCallbacks.splice(index, 1);
      };
    }),
    onStatusChange: vi.fn((callback) => {
      mockStatusCallbacks.push(callback);
      return () => {
        const index = mockStatusCallbacks.indexOf(callback);
        if (index > -1) mockStatusCallbacks.splice(index, 1);
      };
    }),
    getStatus: vi.fn().mockReturnValue('disconnected'),
  };
}

describe('useWebSocket', () => {
  beforeEach(() => {
    mockServiceInstance = createMockService();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with disconnected status', () => {
    const { result } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));

    expect(result.current.status).toBe('disconnected');
  });

  it('should connect on mount', async () => {
    renderHook(() => useWebSocket('ws://localhost:8080/ws'));

    await waitFor(() => {
      expect(mockServiceInstance.connect).toHaveBeenCalled();
    });
  });

  it('should disconnect on unmount', async () => {
    const { unmount } = renderHook(() =>
      useWebSocket('ws://localhost:8080/ws')
    );

    await waitFor(() => {
      expect(mockServiceInstance.connect).toHaveBeenCalled();
    });

    unmount();

    expect(mockServiceInstance.disconnect).toHaveBeenCalled();
  });

  it('should update status when connection status changes', async () => {
    mockServiceInstance.getStatus.mockReturnValue('connecting');

    const { result } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));

    // Simulate status change
    act(() => {
      mockStatusCallbacks.forEach((cb) => cb('connecting'));
    });

    await waitFor(() => {
      expect(result.current.status).toBe('connecting');
    });
  });

  it('should update lastMessage when a message is received', async () => {
    mockServiceInstance.getStatus.mockReturnValue('connected');

    const { result } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));

    const testMessage = JSON.stringify({ type: 'test', data: 'hello' });

    // Simulate message received
    act(() => {
      mockMessageCallbacks.forEach((cb) => cb(testMessage));
    });

    await waitFor(() => {
      expect(result.current.lastMessage).toBe(testMessage);
    });
  });

  it('should send message through service', async () => {
    mockServiceInstance.getStatus.mockReturnValue('connected');

    const { result } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));

    await waitFor(() => {
      expect(mockServiceInstance.connect).toHaveBeenCalled();
    });

    const testMessage = JSON.stringify({ type: 'test', data: 'hello' });
    result.current.sendMessage(testMessage);

    expect(mockServiceInstance.send).toHaveBeenCalledWith(testMessage);
  });

  it('should handle multiple status changes', async () => {
    mockServiceInstance.getStatus
      .mockReturnValueOnce('disconnected')
      .mockReturnValueOnce('connecting')
      .mockReturnValueOnce('connected');

    const { result } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));

    // Initial status
    expect(result.current.status).toBe('disconnected');

    // Simulate connection attempt
    act(() => {
      mockStatusCallbacks.forEach((cb) => cb('connecting'));
    });

    await waitFor(() => {
      expect(result.current.status).toBe('connecting');
    });

    // Simulate successful connection
    act(() => {
      mockStatusCallbacks.forEach((cb) => cb('connected'));
    });

    await waitFor(() => {
      expect(result.current.status).toBe('connected');
    });
  });

  it('should handle received messages correctly', async () => {
    mockServiceInstance.getStatus.mockReturnValue('connected');

    const { result } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));

    const messages = [
      JSON.stringify({ type: 'msg1', data: 'first' }),
      JSON.stringify({ type: 'msg2', data: 'second' }),
    ];

    // Simulate multiple messages
    act(() => {
      messages.forEach((msg) => {
        mockMessageCallbacks.forEach((cb) => cb(msg));
      });
    });

    await waitFor(() => {
      // Should have the last message
      expect(result.current.lastMessage).toBe(messages[messages.length - 1]);
    });
  });

  it('should provide sendMessage function that works when connected', async () => {
    mockServiceInstance.getStatus.mockReturnValue('connected');

    const { result } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));

    await waitFor(() => {
      expect(mockServiceInstance.connect).toHaveBeenCalled();
    });

    const testMessage = JSON.stringify({ type: 'action', data: 'test' });

    act(() => {
      result.current.sendMessage(testMessage);
    });

    expect(mockServiceInstance.send).toHaveBeenCalledWith(testMessage);
  });

  it('should unsubscribe from callbacks on unmount to prevent memory leaks', async () => {
    const { unmount } = renderHook(() =>
      useWebSocket('ws://localhost:8080/ws')
    );

    await waitFor(() => {
      expect(mockServiceInstance.connect).toHaveBeenCalled();
    });

    // Verify callbacks are registered
    expect(mockStatusCallbacks.length).toBe(1);
    expect(mockMessageCallbacks.length).toBe(1);

    // Unmount component
    unmount();

    // Verify callbacks are cleaned up
    expect(mockStatusCallbacks.length).toBe(0);
    expect(mockMessageCallbacks.length).toBe(0);
    expect(mockServiceInstance.disconnect).toHaveBeenCalled();
  });

  describe('TestUseWebSocketHook_LobbyState', () => {
    it('should expose lobbyState in hook return value', () => {
      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      expect(result.current.lobbyState).toBeDefined();
      expect(Array.isArray(result.current.lobbyState)).toBe(true);
      expect(result.current.lobbyState.length).toBe(0);
    });

    it('should update lobbyState when lobby_state message is received', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const lobbyStateMessage = JSON.stringify({
        type: 'lobby_state',
        payload: [
          {
            id: 'table-1',
            name: 'Table 1',
            seats_occupied: 0,
            max_seats: 6,
          },
          {
            id: 'table-2',
            name: 'Table 2',
            seats_occupied: 2,
            max_seats: 6,
          },
        ],
      });

      // Simulate message received
      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(lobbyStateMessage));
      });

      await waitFor(() => {
        expect(result.current.lobbyState.length).toBe(2);
        expect(result.current.lobbyState[0]).toEqual({
          id: 'table-1',
          name: 'Table 1',
          seatsOccupied: 0,
          maxSeats: 6,
        });
        expect(result.current.lobbyState[1]).toEqual({
          id: 'table-2',
          name: 'Table 2',
          seatsOccupied: 2,
          maxSeats: 6,
        });
      });
    });

    it('should handle lobby_state updates with different seat counts', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First message
      const firstMessage = JSON.stringify({
        type: 'lobby_state',
        payload: [
          {
            id: 'table-1',
            name: 'Table 1',
            seats_occupied: 0,
            max_seats: 6,
          },
        ],
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(firstMessage));
      });

      await waitFor(() => {
        expect(result.current.lobbyState[0].seatsOccupied).toBe(0);
      });

      // Second message with updated seat count
      const secondMessage = JSON.stringify({
        type: 'lobby_state',
        payload: [
          {
            id: 'table-1',
            name: 'Table 1',
            seats_occupied: 3,
            max_seats: 6,
          },
        ],
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(secondMessage));
      });

      await waitFor(() => {
        expect(result.current.lobbyState[0].seatsOccupied).toBe(3);
      });
    });

    it('should not update lobbyState for non-lobby_state messages', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const otherMessage = JSON.stringify({
        type: 'session_created',
        payload: JSON.stringify({ token: 'test', name: 'Alice' }),
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(otherMessage));
      });

      await waitFor(() => {
        // lobbyState should remain empty
        expect(result.current.lobbyState.length).toBe(0);
      });
    });
  });

  describe('TestUseWebSocketHook_TableState', () => {
    it('should expose tableState in hook return value', () => {
      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      expect(result.current.tableState).toBeDefined();
      expect(result.current.tableState).toEqual(null);
    });

    it('should update tableState when table_state message is received', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const tableStateMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            {
              index: 0,
              playerName: 'Player1',
              status: 'occupied',
              stack: 1000,
            },
            {
              index: 1,
              playerName: null,
              status: 'empty',
            },
          ],
        },
      });

      // Simulate message received
      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(tableStateMessage));
      });

      await waitFor(() => {
        expect(result.current.tableState).not.toBeNull();
        expect(result.current.tableState?.tableId).toBe('table-1');
        expect(result.current.tableState?.seats.length).toBe(2);
        expect(result.current.tableState?.seats[0]).toEqual({
          index: 0,
          playerName: 'Player1',
          status: 'occupied',
          stack: 1000,
        });
        expect(result.current.tableState?.seats[1]).toEqual({
          index: 1,
          playerName: null,
          status: 'empty',
        });
      });
    });

    it('should handle table_state updates with multiple players', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

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

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(tableStateMessage));
      });

      await waitFor(() => {
        expect(result.current.tableState?.seats.length).toBe(6);
        expect(result.current.tableState?.seats[0].playerName).toBe('Alice');
        expect(result.current.tableState?.seats[1].playerName).toBe('Bob');
        expect(result.current.tableState?.seats[2].playerName).toBeNull();
      });
    });

    it('should update tableState when seat becomes occupied', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First state - seat empty
      const firstMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            {
              index: 0,
              playerName: null,
              status: 'empty',
            },
          ],
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(firstMessage));
      });

      await waitFor(() => {
        expect(result.current.tableState?.seats[0].playerName).toBeNull();
      });

      // Second state - seat occupied
      const secondMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            {
              index: 0,
              playerName: 'NewPlayer',
              status: 'occupied',
            },
          ],
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(secondMessage));
      });

      await waitFor(() => {
        expect(result.current.tableState?.seats[0].playerName).toBe(
          'NewPlayer'
        );
        expect(result.current.tableState?.seats[0].status).toBe('occupied');
      });
    });

    it('should not update tableState for non-table_state messages', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const otherMessage = JSON.stringify({
        type: 'session_created',
        payload: { token: 'test', name: 'Alice' },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(otherMessage));
      });

      await waitFor(() => {
        // tableState should remain null
        expect(result.current.tableState).toBeNull();
      });
    });

    it('table_state updates stack values for seats', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const tableStateMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            {
              index: 0,
              playerName: 'Alice',
              status: 'occupied',
              stack: 5000,
            },
            {
              index: 1,
              playerName: 'Bob',
              status: 'occupied',
              stack: 3500,
            },
            {
              index: 2,
              playerName: null,
              status: 'empty',
            },
          ],
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(tableStateMessage));
      });

      await waitFor(() => {
        expect(result.current.tableState?.seats[0].stack).toBe(5000);
        expect(result.current.tableState?.seats[1].stack).toBe(3500);
        expect(result.current.tableState?.seats[2].stack).toBeUndefined();
      });
    });

    it('table_state updates game state when hand is active', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const tableStateMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            { index: 0, playerName: 'Alice', status: 'occupied', stack: 5000 },
            { index: 1, playerName: 'Bob', status: 'occupied', stack: 3500 },
          ],
          handInProgress: true,
          dealerSeat: 0,
          smallBlindSeat: 1,
          bigBlindSeat: 0,
          pot: 150,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(tableStateMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.dealerSeat).toBe(0);
        expect(result.current.gameState.smallBlindSeat).toBe(1);
        expect(result.current.gameState.bigBlindSeat).toBe(0);
        expect(result.current.gameState.pot).toBe(150);
      });
    });

    it('table_state sets hole cards from payload', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const tableStateMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            { index: 0, playerName: 'Alice', status: 'occupied', stack: 5000 },
          ],
          holeCards: {
            '0': [
              { Rank: 'A', Suit: 's' },
              { Rank: 'K', Suit: 'h' },
            ],
          },
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(tableStateMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.holeCards).toEqual(['As', 'Kh']);
      });
    });

    it('table_state updates card counts per seat', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const tableStateMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            {
              index: 0,
              playerName: 'Alice',
              status: 'occupied',
              stack: 5000,
              cardCount: 2,
            },
            {
              index: 1,
              playerName: 'Bob',
              status: 'occupied',
              stack: 3500,
              cardCount: 2,
            },
            {
              index: 2,
              playerName: null,
              status: 'empty',
            },
          ],
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(tableStateMessage));
      });

      await waitFor(() => {
        expect(result.current.tableState?.seats[0].cardCount).toBe(2);
        expect(result.current.tableState?.seats[1].cardCount).toBe(2);
        expect(result.current.tableState?.seats[2].cardCount).toBeUndefined();
      });
    });

    it('table_state preserves game state when fields absent', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First set initial game state
      const initialMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            { index: 0, playerName: 'Alice', status: 'occupied', stack: 5000 },
          ],
          dealerSeat: 0,
          pot: 100,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(initialMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.dealerSeat).toBe(0);
        expect(result.current.gameState.pot).toBe(100);
      });

      // Send second message without game state fields
      const secondMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            { index: 0, playerName: 'Alice', status: 'occupied', stack: 4950 },
          ],
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(secondMessage));
      });

      // Game state should be preserved
      await waitFor(() => {
        expect(result.current.gameState.dealerSeat).toBe(0);
        expect(result.current.gameState.pot).toBe(100);
      });
    });
  });
});

describe('Phase 4: Board Card Display - WebSocket Event Handling', () => {
  describe('TestUseWebSocket_HandlesBoardDealtEvent', () => {
    it('should initialize boardCards as empty array', () => {
      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // Board cards should be part of gameState
      expect(result.current.gameState).toBeDefined();
    });

    it('should handle board_dealt message with flop cards', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const boardDealtMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
          ],
          street: 'flop',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(boardDealtMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.boardCards).toBeDefined();
        expect(result.current.gameState.boardCards?.length).toBe(3);
        expect(result.current.gameState.boardCards?.[0]).toBe('As');
        expect(result.current.gameState.boardCards?.[1]).toBe('Kh');
        expect(result.current.gameState.boardCards?.[2]).toBe('Qd');
      });
    });

    it('should handle board_dealt message with turn card', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First send flop
      const flopMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
          ],
          street: 'flop',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(flopMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.boardCards?.length).toBe(3);
      });

      // Then send turn
      const turnMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
            { Rank: 'J', Suit: 'c' },
          ],
          street: 'turn',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(turnMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.boardCards?.length).toBe(4);
        expect(result.current.gameState.boardCards?.[3]).toBe('Jc');
      });
    });

    it('should handle board_dealt message with river card', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const riverMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
            { Rank: 'J', Suit: 'c' },
            { Rank: 'T', Suit: 's' },
          ],
          street: 'river',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(riverMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.boardCards?.length).toBe(5);
        expect(result.current.gameState.boardCards?.[4]).toBe('Ts');
      });
    });

    it('should update street indicator when board_dealt is received', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const boardDealtMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
          ],
          street: 'flop',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(boardDealtMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.street).toBe('flop');
      });
    });

    it('should incrementally update board cards on subsequent deals', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // Flop
      const flopMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
          ],
          street: 'flop',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(flopMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.boardCards?.length).toBe(3);
        expect(result.current.gameState.street).toBe('flop');
      });

      // Turn
      const turnMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
            { Rank: 'J', Suit: 'c' },
          ],
          street: 'turn',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(turnMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.boardCards?.length).toBe(4);
        expect(result.current.gameState.street).toBe('turn');
      });

      // River
      const riverMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
            { Rank: 'J', Suit: 'c' },
            { Rank: 'T', Suit: 's' },
          ],
          street: 'river',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(riverMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.boardCards?.length).toBe(5);
        expect(result.current.gameState.street).toBe('river');
      });
    });

    it('should not update board cards for non-board_dealt messages', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const otherMessage = JSON.stringify({
        type: 'action_result',
        payload: {
          seatIndex: 1,
          action: 'call',
          amountActed: 20,
          newStack: 980,
          pot: 60,
          nextActor: 2,
          roundOver: false,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(otherMessage));
      });

      await waitFor(() => {
        // boardCards should remain undefined or empty
        expect(result.current.gameState.boardCards).toBeUndefined();
      });
    });
  });
});

describe('Phase 5: Raise Protocol - Frontend Protocol and State', () => {
  describe('TestUseWebSocket_ParseRaiseProtocol', () => {
    it('parses minRaise and maxRaise from action_request', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const actionRequestMessage = JSON.stringify({
        type: 'action_request',
        payload: {
          seatIndex: 2,
          validActions: ['fold', 'call', 'raise'],
          callAmount: 20,
          minRaise: 40,
          maxRaise: 500,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(actionRequestMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.currentActor).toBe(2);
        expect(result.current.gameState.validActions).toEqual([
          'fold',
          'call',
          'raise',
        ]);
        expect(result.current.gameState.callAmount).toBe(20);
        expect(result.current.gameState.minRaise).toBe(40);
        expect(result.current.gameState.maxRaise).toBe(500);
      });
    });

    it('includes minRaise and maxRaise in gameState when raise is available', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const actionRequestMessage = JSON.stringify({
        type: 'action_request',
        payload: {
          seatIndex: 0,
          validActions: ['fold', 'raise'],
          callAmount: 0,
          minRaise: 40,
          maxRaise: 1000,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(actionRequestMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.minRaise).toBe(40);
        expect(result.current.gameState.maxRaise).toBe(1000);
      });
    });

    it('handles action_request without minRaise/maxRaise fields', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const actionRequestMessage = JSON.stringify({
        type: 'action_request',
        payload: {
          seatIndex: 1,
          validActions: ['fold', 'check'],
          callAmount: 0,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(actionRequestMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.currentActor).toBe(1);
        expect(result.current.gameState.validActions).toEqual([
          'fold',
          'check',
        ]);
        // minRaise and maxRaise should not be present or undefined
        expect(result.current.gameState.minRaise).toBeUndefined();
        expect(result.current.gameState.maxRaise).toBeUndefined();
      });
    });
  });

  describe('TestUseWebSocket_SendActionWithAmount', () => {
    it('sendAction includes amount for raise', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      await waitFor(() => {
        expect(mockServiceInstance.connect).toHaveBeenCalled();
      });

      // First send seat_assigned message to set player's seat
      const seatAssignedMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 2,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(seatAssignedMessage));
      });

      // Reset mock to clear previous calls
      vi.clearAllMocks();

      act(() => {
        result.current.sendAction?.('raise', 100);
      });

      expect(mockServiceInstance.send).toHaveBeenCalledWith(
        JSON.stringify({
          type: 'player_action',
          payload: {
            seatIndex: 2,
            action: 'raise',
            amount: 100,
          },
        })
      );
    });

    it('sendAction omits amount for fold', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      await waitFor(() => {
        expect(mockServiceInstance.connect).toHaveBeenCalled();
      });

      // First send seat_assigned message to set player's seat
      const seatAssignedMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 1,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(seatAssignedMessage));
      });

      // Reset mock to clear previous calls
      vi.clearAllMocks();

      act(() => {
        result.current.sendAction?.('fold');
      });

      expect(mockServiceInstance.send).toHaveBeenCalledWith(
        JSON.stringify({
          type: 'player_action',
          payload: {
            seatIndex: 1,
            action: 'fold',
          },
        })
      );
    });

    it('sendAction omits amount for check', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      await waitFor(() => {
        expect(mockServiceInstance.connect).toHaveBeenCalled();
      });

      // First send seat_assigned message to set player's seat
      const seatAssignedMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 0,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(seatAssignedMessage));
      });

      // Reset mock to clear previous calls
      vi.clearAllMocks();

      act(() => {
        result.current.sendAction?.('check');
      });

      expect(mockServiceInstance.send).toHaveBeenCalledWith(
        JSON.stringify({
          type: 'player_action',
          payload: {
            seatIndex: 0,
            action: 'check',
          },
        })
      );
    });

    it('sendAction omits amount for call', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      await waitFor(() => {
        expect(mockServiceInstance.connect).toHaveBeenCalled();
      });

      // First send seat_assigned message to set player's seat
      const seatAssignedMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 3,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(seatAssignedMessage));
      });

      // Reset mock to clear previous calls
      vi.clearAllMocks();

      act(() => {
        result.current.sendAction?.('call');
      });

      expect(mockServiceInstance.send).toHaveBeenCalledWith(
        JSON.stringify({
          type: 'player_action',
          payload: {
            seatIndex: 3,
            action: 'call',
          },
        })
      );
    });
  });
});

describe('Phase 6: Showdown & Settlement - WebSocket Event Handling', () => {
  describe('TestUseWebSocket_ShowdownResultEvent', () => {
    it('should handle showdown_result event with single winner', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const showdownMessage = JSON.stringify({
        type: 'showdown_result',
        payload: {
          winnerSeats: [1],
          winningHand: 'Pair of Aces',
          potAmount: 300,
          amountsWon: {
            '1': 300,
          },
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(showdownMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.showdown).toBeDefined();
        expect(result.current.gameState.showdown?.winnerSeats).toEqual([1]);
        expect(result.current.gameState.showdown?.winningHand).toBe(
          'Pair of Aces'
        );
        expect(result.current.gameState.showdown?.potAmount).toBe(300);
        expect(result.current.gameState.showdown?.amountsWon).toEqual({
          1: 300,
        });
      });
    });

    it('should handle showdown_result event with multiple winners (split pot)', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const showdownMessage = JSON.stringify({
        type: 'showdown_result',
        payload: {
          winnerSeats: [0, 2],
          winningHand: 'Pair of Kings',
          potAmount: 400,
          amountsWon: {
            '0': 200,
            '2': 200,
          },
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(showdownMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.showdown?.winnerSeats).toEqual([0, 2]);
        expect(result.current.gameState.showdown?.amountsWon).toEqual({
          0: 200,
          2: 200,
        });
      });
    });

    it('should convert amountsWon string keys to numbers', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const showdownMessage = JSON.stringify({
        type: 'showdown_result',
        payload: {
          winnerSeats: [3],
          winningHand: 'Flush',
          potAmount: 500,
          amountsWon: {
            '3': 500,
          },
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(showdownMessage));
      });

      await waitFor(() => {
        const amountsWon = result.current.gameState.showdown?.amountsWon;
        expect(amountsWon).toBeDefined();
        // Verify keys are numbers, not strings
        const keys = Object.keys(amountsWon || {});
        expect(keys.length).toBeGreaterThan(0);
        expect(amountsWon?.[3]).toBe(500);
      });
    });
  });

  describe('TestUseWebSocket_HandCompleteEvent', () => {
    it('should handle hand_complete event', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const handCompleteMessage = JSON.stringify({
        type: 'hand_complete',
        payload: {
          message: 'Hand complete! Winner(s) collected the pot.',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(handCompleteMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.handComplete).toBeDefined();
        expect(result.current.gameState.handComplete?.message).toBe(
          'Hand complete! Winner(s) collected the pot.'
        );
      });
    });

    it('should handle hand_complete event after showdown', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First send showdown
      const showdownMessage = JSON.stringify({
        type: 'showdown_result',
        payload: {
          winnerSeats: [1],
          winningHand: 'Pair of Aces',
          potAmount: 300,
          amountsWon: { '1': 300 },
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(showdownMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.showdown).toBeDefined();
      });

      // Then send hand_complete
      const handCompleteMessage = JSON.stringify({
        type: 'hand_complete',
        payload: {
          message: 'Hand complete!',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(handCompleteMessage));
      });

      // Should have both showdown and handComplete set
      await waitFor(() => {
        expect(result.current.gameState.handComplete).toBeDefined();
        expect(result.current.gameState.showdown).toBeDefined();
      });
    });
  });
});

describe('Phase 1: Clear Street Label on Hand Start', () => {
  describe('TestUseWebSocket_StreetClearingOnStartHand', () => {
    it('should clear street field when start_hand action is sent', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First set player's seat
      const seatAssignedMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 1,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(seatAssignedMessage));
      });

      // Set up initial board cards with street = 'river'
      const boardDealtMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
            { Rank: 'J', Suit: 'c' },
            { Rank: 'T', Suit: 's' },
          ],
          street: 'river',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(boardDealtMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.street).toBe('river');
      });

      // Now send start_hand action
      act(() => {
        result.current.sendStartHand();
      });

      // After sending start_hand, street should be cleared (undefined)
      expect(result.current.gameState.street).toBeUndefined();
    });

    it('hand_started message clears street field', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First set up board cards with street = 'turn'
      const boardDealtMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
            { Rank: 'J', Suit: 'c' },
          ],
          street: 'turn',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(boardDealtMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.street).toBe('turn');
      });

      // Now receive hand_started message
      const handStartedMessage = JSON.stringify({
        type: 'hand_started',
        payload: {
          dealerSeat: 0,
          smallBlindSeat: 1,
          bigBlindSeat: 2,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(handStartedMessage));
      });

      // After hand_started, street should be cleared (undefined)
      expect(result.current.gameState.street).toBeUndefined();
    });

    it('street remains undefined until board_dealt is received', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // Initially, street should be undefined
      expect(result.current.gameState.street).toBeUndefined();

      // Receive hand_started message
      const handStartedMessage = JSON.stringify({
        type: 'hand_started',
        payload: {
          dealerSeat: 0,
          smallBlindSeat: 1,
          bigBlindSeat: 2,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(handStartedMessage));
      });

      // Street should still be undefined
      expect(result.current.gameState.street).toBeUndefined();

      // Now receive board_dealt message
      const boardDealtMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
          ],
          street: 'flop',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(boardDealtMessage));
      });

      // Now street should be set to 'flop'
      await waitFor(() => {
        expect(result.current.gameState.street).toBe('flop');
      });
    });
  });
});

describe('Phase 3: Remove Auto-Clear and Show Start Hand Button After Showdown', () => {
  describe('TestUseWebSocket_ShowdownPersists', () => {
    it('should not auto-clear showdown state after timeout', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First send showdown_result
      const showdownMessage = JSON.stringify({
        type: 'showdown_result',
        payload: {
          winnerSeats: [1],
          winningHand: 'Pair of Aces',
          potAmount: 300,
          amountsWon: { '1': 300 },
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(showdownMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.showdown).toBeDefined();
      });

      // Then send hand_complete
      const handCompleteMessage = JSON.stringify({
        type: 'hand_complete',
        payload: {
          message: 'Hand complete!',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(handCompleteMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.handComplete).toBeDefined();
        expect(result.current.gameState.showdown).toBeDefined();
      });

      // Wait 2 seconds (half of the old timeout)
      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Showdown should still be there (not auto-cleared)
      expect(result.current.gameState.showdown).toBeDefined();
      expect(result.current.gameState.handComplete).toBeDefined();
    }, 10000); // Increase test timeout to 10 seconds

    it('should clear showdown and handComplete when start_hand action is sent', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First set player's seat
      const seatAssignedMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 1,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(seatAssignedMessage));
      });

      // First set up showdown state
      const showdownMessage = JSON.stringify({
        type: 'showdown_result',
        payload: {
          winnerSeats: [1],
          winningHand: 'Pair of Aces',
          potAmount: 300,
          amountsWon: { '1': 300 },
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(showdownMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.showdown).toBeDefined();
      });

      // Then send hand_complete
      const handCompleteMessage = JSON.stringify({
        type: 'hand_complete',
        payload: {
          message: 'Hand complete!',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(handCompleteMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.handComplete).toBeDefined();
      });

      // Now send start_hand action
      act(() => {
        result.current.sendStartHand();
      });

      // After sending start_hand, local state should be cleared
      // (the server would send table_state to reset officially, but locally we clear immediately)
      expect(result.current.gameState.showdown).toBeUndefined();
      expect(result.current.gameState.handComplete).toBeUndefined();

      // Verify the message was sent to the server
      expect(mockServiceInstance.send).toHaveBeenCalledWith(
        JSON.stringify({
          type: 'start_hand',
          payload: {},
        })
      );
    });
  });

  describe('TestUseWebSocket_BoardCardsClearingOnStartHand', () => {
    it('should clear boardCards array when start_hand action is sent', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First set player's seat
      const seatAssignedMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 0,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(seatAssignedMessage));
      });

      // Set up initial board cards
      const boardDealtMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
          ],
          street: 'flop',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(boardDealtMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.boardCards).toBeDefined();
        expect(result.current.gameState.boardCards?.length).toBe(3);
      });

      // Now send start_hand action
      act(() => {
        result.current.sendStartHand();
      });

      // After sending start_hand, boardCards should be cleared to empty array
      expect(result.current.gameState.boardCards).toBeDefined();
      expect(result.current.gameState.boardCards).toEqual([]);
    });

    it('should have boardCards as empty array (not undefined) after start_hand action', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First set player's seat
      const seatAssignedMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 2,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(seatAssignedMessage));
      });

      // Set up initial board cards
      const boardDealtMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: '2', Suit: 's' },
            { Rank: '3', Suit: 'h' },
            { Rank: '4', Suit: 'd' },
            { Rank: '5', Suit: 'c' },
            { Rank: '6', Suit: 's' },
          ],
          street: 'river',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(boardDealtMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.boardCards?.length).toBe(5);
      });

      // Now send start_hand action
      act(() => {
        result.current.sendStartHand();
      });

      // boardCards should be an empty array, not undefined
      expect(result.current.gameState.boardCards).toStrictEqual([]);
      expect(Array.isArray(result.current.gameState.boardCards)).toBe(true);
    });
  });

  describe('TestUseWebSocket_StartHandPotReset', () => {
    it('sendAction resets pot to 0 when start_hand is sent', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First set player's seat
      const seatAssignedMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 1,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(seatAssignedMessage));
      });

      // Set up initial game state with pot > 0
      const tableStateMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            { index: 0, playerName: 'Alice', status: 'occupied', stack: 5000 },
            { index: 1, playerName: 'Bob', status: 'occupied', stack: 3500 },
          ],
          pot: 150,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(tableStateMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.pot).toBe(150);
      });

       // Now send start_hand action
       act(() => {
         result.current.sendStartHand();
       });

       // After sending start_hand, pot should remain at its previous value (150)
       // The optimistic update no longer sets pot = 0
       expect(result.current.gameState.pot).toBe(150);
     });

    it('start_hand optimistic update clears all hand state (pot, boardCards, showdown, handComplete)', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First set player's seat
      const seatAssignedMessage = JSON.stringify({
        type: 'seat_assigned',
        payload: {
          tableId: 'table-1',
          seatIndex: 0,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(seatAssignedMessage));
      });

      // Set up initial game state with multiple fields
      const tableStateMessage = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            { index: 0, playerName: 'Alice', status: 'occupied', stack: 5000 },
          ],
          pot: 300,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(tableStateMessage));
      });

      // Set board cards
      const boardDealtMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
          ],
          street: 'flop',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(boardDealtMessage));
      });

      // Set showdown state
      const showdownMessage = JSON.stringify({
        type: 'showdown_result',
        payload: {
          winnerSeats: [0],
          winningHand: 'Pair of Aces',
          potAmount: 300,
          amountsWon: { '0': 300 },
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(showdownMessage));
      });

      // Set hand complete state
      const handCompleteMessage = JSON.stringify({
        type: 'hand_complete',
        payload: {
          message: 'Hand complete!',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(handCompleteMessage));
      });

      // Verify all state is set
      await waitFor(() => {
        expect(result.current.gameState.pot).toBe(300);
        expect(result.current.gameState.boardCards?.length).toBe(3);
        expect(result.current.gameState.showdown).toBeDefined();
        expect(result.current.gameState.handComplete).toBeDefined();
      });

       // Now send start_hand action
       act(() => {
         result.current.sendStartHand();
       });

       // After sending start_hand, pot should remain at its previous value (300)
       // The optimistic update no longer sets pot = 0
       // But boardCards, showdown, and handComplete should be cleared
       expect(result.current.gameState.pot).toBe(300);
       expect(result.current.gameState.boardCards).toEqual([]);
       expect(result.current.gameState.showdown).toBeUndefined();
       expect(result.current.gameState.handComplete).toBeUndefined();
     });
  });
});

describe('Action Request and Result Handlers', () => {
  describe('TestUseWebSocket_ActionRequest', () => {
    it('should update game state when action_request message is received', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const actionRequestMessage = JSON.stringify({
        type: 'action_request',
        payload: {
          seatIndex: 2,
          validActions: ['fold', 'call'],
          callAmount: 20,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(actionRequestMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.currentActor).toBe(2);
        expect(result.current.gameState.validActions).toEqual(['fold', 'call']);
        expect(result.current.gameState.callAmount).toBe(20);
      });
    });

    it('should handle action_request with check available', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const actionRequestMessage = JSON.stringify({
        type: 'action_request',
        payload: {
          seatIndex: 1,
          validActions: ['fold', 'check'],
          callAmount: 0,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(actionRequestMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.currentActor).toBe(1);
        expect(result.current.gameState.validActions).toEqual([
          'fold',
          'check',
        ]);
        expect(result.current.gameState.callAmount).toBe(0);
      });
    });
  });

  describe('TestUseWebSocket_ActionResult', () => {
    it('should update game state when action_result message is received', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // Set initial table state
      const initialTableState = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            { index: 0, playerName: 'Alice', status: 'occupied', stack: 1000 },
            { index: 1, playerName: 'Bob', status: 'occupied', stack: 980 },
            {
              index: 2,
              playerName: 'Charlie',
              status: 'occupied',
              stack: 1000,
            },
          ],
          pot: 30,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(initialTableState));
      });

      await waitFor(() => {
        expect(result.current.tableState?.seats[1].stack).toBe(980);
        expect(result.current.gameState.pot).toBe(30);
      });

      // Player 1 calls 20
      const actionResultMessage = JSON.stringify({
        type: 'action_result',
        payload: {
          seatIndex: 1,
          action: 'call',
          amountActed: 20,
          newStack: 960,
          pot: 50,
          nextActor: 2,
          roundOver: false,
          validActions: ['fold', 'call'],
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(actionResultMessage));
      });

      await waitFor(() => {
        expect(result.current.tableState?.seats[1].stack).toBe(960);
        expect(result.current.gameState.pot).toBe(50);
        expect(result.current.gameState.currentActor).toBe(2);
        expect(result.current.gameState.validActions).toEqual(['fold', 'call']);
      });
    });

    it('should mark player as folded when fold action result received', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // Set initial table state
      const initialTableState = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            { index: 0, playerName: 'Alice', status: 'occupied', stack: 1000 },
            { index: 1, playerName: 'Bob', status: 'occupied', stack: 1000 },
          ],
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(initialTableState));
      });

      // Player 0 folds
      const actionResultMessage = JSON.stringify({
        type: 'action_result',
        payload: {
          seatIndex: 0,
          action: 'fold',
          amountActed: 0,
          newStack: 1000,
          pot: 30,
          nextActor: 1,
          roundOver: false,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(actionResultMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.foldedPlayers).toContain(0);
        expect(result.current.gameState.pot).toBe(30);
        expect(result.current.gameState.currentActor).toBe(1);
      });
    });

    it('should handle betting round completion when roundOver is true', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      const actionResultMessage = JSON.stringify({
        type: 'action_result',
        payload: {
          seatIndex: 2,
          action: 'call',
          amountActed: 20,
          newStack: 980,
          pot: 60,
          nextActor: null,
          roundOver: true,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(actionResultMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.pot).toBe(60);
        expect(result.current.gameState.currentActor).toBeNull();
        expect(result.current.gameState.roundOver).toBe(true);
      });
    });
  });
});

describe('Phase 3: Clear Board Cards on Hand Started Message (Backend Confirmation)', () => {
  describe('TestUseWebSocket_HandStartedBoardCardsClear', () => {
    it('hand_started message sets boardCards to empty array', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First set up board cards (from a previous hand)
      const boardDealtMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
          ],
          street: 'flop',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(boardDealtMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.boardCards?.length).toBe(3);
      });

      // Now send hand_started message
      const handStartedMessage = JSON.stringify({
        type: 'hand_started',
        payload: {
          dealerSeat: 0,
          smallBlindSeat: 1,
          bigBlindSeat: 2,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(handStartedMessage));
      });

      await waitFor(() => {
        // boardCards should be cleared to empty array
        expect(result.current.gameState.boardCards).toBeDefined();
        expect(result.current.gameState.boardCards).toEqual([]);
      });
    });

    it('hand_started preserves dealer/blind seats and other state', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // First set up initial game state with some values
      const initialTableState = JSON.stringify({
        type: 'table_state',
        payload: {
          tableId: 'table-1',
          seats: [
            { index: 0, playerName: 'Alice', status: 'occupied', stack: 5000 },
            { index: 1, playerName: 'Bob', status: 'occupied', stack: 3500 },
          ],
          pot: 150,
          dealerSeat: 0,
          smallBlindSeat: 1,
          bigBlindSeat: 0,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(initialTableState));
      });

      await waitFor(() => {
        expect(result.current.gameState.dealerSeat).toBe(0);
        expect(result.current.gameState.smallBlindSeat).toBe(1);
        expect(result.current.gameState.bigBlindSeat).toBe(0);
        expect(result.current.gameState.pot).toBe(150);
      });

      // Set board cards
      const boardDealtMessage = JSON.stringify({
        type: 'board_dealt',
        payload: {
          boardCards: [
            { Rank: 'A', Suit: 's' },
            { Rank: 'K', Suit: 'h' },
            { Rank: 'Q', Suit: 'd' },
          ],
          street: 'flop',
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(boardDealtMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.boardCards?.length).toBe(3);
      });

      // Now send hand_started message with new dealer positions
      const handStartedMessage = JSON.stringify({
        type: 'hand_started',
        payload: {
          dealerSeat: 1,
          smallBlindSeat: 2,
          bigBlindSeat: 0,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(handStartedMessage));
      });

      await waitFor(() => {
        // Dealer/blind seats should be updated
        expect(result.current.gameState.dealerSeat).toBe(1);
        expect(result.current.gameState.smallBlindSeat).toBe(2);
        expect(result.current.gameState.bigBlindSeat).toBe(0);
        // boardCards should be cleared
        expect(result.current.gameState.boardCards).toEqual([]);
        // pot should be preserved (not affected by hand_started)
        expect(result.current.gameState.pot).toBe(150);
      });
    });

    it('hand_started clears foldedPlayers from previous hand', async () => {
      mockServiceInstance.getStatus.mockReturnValue('connected');

      const { result } = renderHook(() =>
        useWebSocket('ws://localhost:8080/ws')
      );

      // Set up initial game state with some folded players
      const actionResultMessage = JSON.stringify({
        type: 'action_result',
        payload: {
          seatIndex: 1,
          action: 'fold',
          amountActed: 0,
          newStack: 1000,
          pot: 30,
          nextActor: 2,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(actionResultMessage));
      });

      await waitFor(() => {
        expect(result.current.gameState.foldedPlayers).toContain(1);
      });

      // Now send hand_started message for new hand
      const handStartedMessage = JSON.stringify({
        type: 'hand_started',
        payload: {
          dealerSeat: 1,
          smallBlindSeat: 2,
          bigBlindSeat: 0,
        },
      });

      act(() => {
        mockMessageCallbacks.forEach((cb) => cb(handStartedMessage));
      });

      await waitFor(() => {
        // foldedPlayers should be cleared for new hand
        expect(result.current.gameState.foldedPlayers).toEqual([]);
        // Dealer/blind seats should be updated
        expect(result.current.gameState.dealerSeat).toBe(1);
        expect(result.current.gameState.smallBlindSeat).toBe(2);
        expect(result.current.gameState.bigBlindSeat).toBe(0);
      });
    });
  });
});
