import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useWebSocket } from './useWebSocket';
import type { WebSocketService } from '../services/WebSocketService';

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
    const { unmount } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));

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
      mockStatusCallbacks.forEach(cb => cb('connecting'));
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
      mockMessageCallbacks.forEach(cb => cb(testMessage));
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
      mockStatusCallbacks.forEach(cb => cb('connecting'));
    });

    await waitFor(() => {
      expect(result.current.status).toBe('connecting');
    });

    // Simulate successful connection
    act(() => {
      mockStatusCallbacks.forEach(cb => cb('connected'));
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
      messages.forEach(msg => {
        mockMessageCallbacks.forEach(cb => cb(msg));
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
     const { unmount } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));

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
 });
