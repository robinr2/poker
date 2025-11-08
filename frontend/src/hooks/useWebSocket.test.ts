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
        payload: JSON.stringify([
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
        ]),
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
        payload: JSON.stringify([
          {
            id: 'table-1',
            name: 'Table 1',
            seats_occupied: 0,
            max_seats: 6,
          },
        ]),
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
        payload: JSON.stringify([
          {
            id: 'table-1',
            name: 'Table 1',
            seats_occupied: 3,
            max_seats: 6,
          },
        ]),
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
        expect(result.current.tableState?.seats[0].playerName).toBe('NewPlayer');
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
  });
});
