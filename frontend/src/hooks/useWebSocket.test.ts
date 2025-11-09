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
        expect(result.current.gameState.validActions).toEqual(['fold', 'call', 'raise']);
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
        expect(result.current.gameState.validActions).toEqual(['fold', 'check']);
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
        expect(result.current.gameState.validActions).toEqual(['fold', 'check']);
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
            { index: 2, playerName: 'Charlie', status: 'occupied', stack: 1000 },
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
          amount: 20,
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
          amount: 0,
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
          amount: 20,
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
