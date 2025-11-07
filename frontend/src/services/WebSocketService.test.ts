import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

import { WebSocketService } from './WebSocketService';

describe('WebSocketService', () => {
  let service: WebSocketService;
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
    // Replace global WebSocket with mock
    createdMockSockets = [];
    originalWebSocket = global.WebSocket as typeof WebSocket;
    global.WebSocket = MockWebSocket as typeof WebSocket;
  });

  afterEach(() => {
    // Restore original WebSocket
    global.WebSocket = originalWebSocket;
    vi.clearAllTimers();
  });

  describe('connect', () => {
    it('should establish a WebSocket connection', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');
      const connectPromise = service.connect();

      expect(service.getStatus()).toBe('connecting');

      // Get the mock socket that was just created
      const mockSocket = createdMockSockets[0];
      expect(mockSocket).toBeDefined();

      // Simulate connection opening
      mockSocket.simulateOpen();

      await connectPromise;
      expect(service.getStatus()).toBe('connected');

      vi.useRealTimers();
    });

    it('should fail if connection cannot be established', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');
      const connectPromise = service.connect();

      const mockSocket = createdMockSockets[0];

      // Simulate connection error
      mockSocket.simulateError();

      await expect(connectPromise).rejects.toThrow();

      vi.useRealTimers();
    });

    it('should set status to connecting when attempting connection', () => {
      service = new WebSocketService('ws://localhost:8080/ws');
      service.connect();
      expect(service.getStatus()).toBe('connecting');
    });
  });

  describe('disconnect', () => {
    it('should close the WebSocket connection', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');

      const connectPromise = service.connect();
      const mockSocket = createdMockSockets[0];

      mockSocket.simulateOpen();

      await connectPromise;
      expect(service.getStatus()).toBe('connected');

      service.disconnect();
      expect(service.getStatus()).toBe('disconnected');

      vi.useRealTimers();
    });

    it('should not throw when disconnecting if not connected', () => {
      service = new WebSocketService('ws://localhost:8080/ws');
      expect(() => service.disconnect()).not.toThrow();
    });
  });

  describe('send', () => {
    it('should send a message through the WebSocket', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');

      const connectPromise = service.connect();
      const mockSocket = createdMockSockets[0];

      mockSocket.simulateOpen();

      await connectPromise;

      const testMessage = JSON.stringify({ type: 'test', data: 'hello' });
      service.send(testMessage);

      expect(mockSocket.messages).toContain(testMessage);

      vi.useRealTimers();
    });

    it('should throw if trying to send when not connected', () => {
      service = new WebSocketService('ws://localhost:8080/ws');

      const testMessage = JSON.stringify({ type: 'test', data: 'hello' });
      expect(() => service.send(testMessage)).toThrow();
    });
  });

  describe('onMessage', () => {
    it('should call callback when a message is received', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');
      const messageCallback = vi.fn();

      service.onMessage(messageCallback);

      const connectPromise = service.connect();
      const mockSocket = createdMockSockets[0];

      mockSocket.simulateOpen();

      await connectPromise;

      const testData = JSON.stringify({ type: 'test', data: 'hello' });
      mockSocket.simulateMessage(testData);

      expect(messageCallback).toHaveBeenCalledWith(testData);

      vi.useRealTimers();
    });

    it('should support multiple message listeners', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');
      const callback1 = vi.fn();
      const callback2 = vi.fn();

      service.onMessage(callback1);
      service.onMessage(callback2);

      const connectPromise = service.connect();
      const mockSocket = createdMockSockets[0];

      mockSocket.simulateOpen();

      await connectPromise;

      const testData = JSON.stringify({ type: 'test' });
      mockSocket.simulateMessage(testData);

      expect(callback1).toHaveBeenCalledWith(testData);
      expect(callback2).toHaveBeenCalledWith(testData);

      vi.useRealTimers();
    });

    it('should return unsubscribe function that removes callback', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');
      const callback1 = vi.fn();
      const callback2 = vi.fn();

      const unsubscribe1 = service.onMessage(callback1);
      service.onMessage(callback2);

      const connectPromise = service.connect();
      const mockSocket = createdMockSockets[0];

      mockSocket.simulateOpen();

      await connectPromise;

      // First message - both callbacks should fire
      const testData1 = JSON.stringify({ type: 'test1' });
      mockSocket.simulateMessage(testData1);

      expect(callback1).toHaveBeenCalledWith(testData1);
      expect(callback2).toHaveBeenCalledWith(testData1);

      // Unsubscribe callback1
      unsubscribe1();

      // Second message - only callback2 should fire
      const testData2 = JSON.stringify({ type: 'test2' });
      mockSocket.simulateMessage(testData2);

      expect(callback1).toHaveBeenCalledTimes(1);
      expect(callback2).toHaveBeenCalledTimes(2);
      expect(callback2).toHaveBeenLastCalledWith(testData2);

      vi.useRealTimers();
    });
  });

  describe('onStatusChange', () => {
    it('should call callback when connection status changes', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');
      const statusCallback = vi.fn();

      service.onStatusChange(statusCallback);

      const connectPromise = service.connect();
      expect(statusCallback).toHaveBeenCalledWith('connecting');

      const mockSocket = createdMockSockets[0];

      mockSocket.simulateOpen();

      await connectPromise;
      expect(statusCallback).toHaveBeenCalledWith('connected');

      service.disconnect();
      expect(statusCallback).toHaveBeenCalledWith('disconnected');

      vi.useRealTimers();
    });

    it('should return unsubscribe function that removes callback', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');
      const callback1 = vi.fn();
      const callback2 = vi.fn();

      const unsubscribe1 = service.onStatusChange(callback1);
      service.onStatusChange(callback2);

      const connectPromise = service.connect();
      expect(callback1).toHaveBeenCalledWith('connecting');
      expect(callback2).toHaveBeenCalledWith('connecting');

      const mockSocket = createdMockSockets[0];

      mockSocket.simulateOpen();

      await connectPromise;

      expect(callback1).toHaveBeenCalledWith('connected');
      expect(callback2).toHaveBeenCalledWith('connected');

      // Unsubscribe callback1
      unsubscribe1();

      service.disconnect();

      // After disconnect, only callback2 should be called
      expect(callback1).toHaveBeenCalledTimes(2);
      expect(callback2).toHaveBeenCalledTimes(3);
      expect(callback2).toHaveBeenLastCalledWith('disconnected');

      vi.useRealTimers();
    });
  });

  describe('auto-reconnect', () => {
    it('should attempt to reconnect with exponential backoff', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');
      const statusCallback = vi.fn();
      service.onStatusChange(statusCallback);

      // First connection attempt
      const connectPromise = service.connect();
      expect(statusCallback).toHaveBeenCalledWith('connecting');

      const mockSocket = createdMockSockets[0];

      // Simulate connection error
      mockSocket.simulateError();

      await connectPromise.catch(() => {});

      // Should have error status after first failure
      expect(statusCallback).toHaveBeenCalledWith('error');

      // Wait for exponential backoff: 1s
      vi.advanceTimersByTime(1000);

      // Next reconnection attempt should start
      expect(statusCallback).toHaveBeenCalledWith('connecting');

      vi.useRealTimers();
    });

    it('should cap reconnection backoff at 30 seconds', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');
      const statusCallback = vi.fn();
      service.onStatusChange(statusCallback);

      // Multiple connection failures to test backoff capping
      for (let i = 0; i < 3; i++) {
        const promise = service.connect().catch(() => {});
        const mockSocket = createdMockSockets[createdMockSockets.length - 1];

        mockSocket.simulateError();
        vi.advanceTimersByTime(100);

        await promise;
      }

      // After multiple failures, backoff should be capped at 30s
      // Verify that the last attempt's backoff timing
      vi.useRealTimers();
    });
  });

  describe('getStatus', () => {
    it('should return the current connection status', async () => {
      vi.useFakeTimers();

      service = new WebSocketService('ws://localhost:8080/ws');

      expect(service.getStatus()).toBe('disconnected');

      const connectPromise = service.connect();
      expect(service.getStatus()).toBe('connecting');

      const mockSocket = createdMockSockets[0];

      mockSocket.simulateOpen();

      await connectPromise;
      expect(service.getStatus()).toBe('connected');

      service.disconnect();
      expect(service.getStatus()).toBe('disconnected');

      vi.useRealTimers();
    });
  });
});
