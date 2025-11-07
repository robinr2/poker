import { useEffect, useState, useCallback, useRef } from 'react';

import {
  WebSocketService,
  type ConnectionStatus,
} from '../services/WebSocketService';

interface UseWebSocketReturn {
  status: ConnectionStatus;
  sendMessage: (message: string) => void;
  lastMessage: string | null;
}

interface UseWebSocketOptions {
  onMessage?: (message: string) => void;
}

export function useWebSocket(
  url: string,
  token?: string,
  options?: UseWebSocketOptions
): UseWebSocketReturn {
  const [status, setStatus] = useState<ConnectionStatus>('disconnected');
  const [lastMessage, setLastMessage] = useState<string | null>(null);
  const serviceRef = useRef<WebSocketService | null>(null);
  const onMessageRef = useRef(options?.onMessage);

  // Update the ref when the callback changes
  useEffect(() => {
    onMessageRef.current = options?.onMessage;
  }, [options?.onMessage]);

  // Initialize and connect
  useEffect(() => {
    // Always create a new service instance when url or token changes
    // Cleanup from previous effect will disconnect the old service
    serviceRef.current = new WebSocketService(url, token);
    const service = serviceRef.current;

    // Set up status change listener and store unsubscribe function
    const unsubscribeStatus = service.onStatusChange((newStatus) => {
      setStatus(newStatus);
    });

    // Set up message listener and store unsubscribe function
    const unsubscribeMessage = service.onMessage((data) => {
      // Call the callback immediately if provided
      if (onMessageRef.current) {
        onMessageRef.current(data);
      }
      // Also update state for components that need it
      setLastMessage(data);
    });

    // Connect
    service.connect().catch(() => {
      // Connection will retry automatically with exponential backoff
    });

    // Cleanup on unmount or dependency change
    return () => {
      unsubscribeStatus();
      unsubscribeMessage();
      service.disconnect();
    };
  }, [url, token]);

  // Memoize sendMessage to prevent unnecessary re-renders
  const sendMessage = useCallback((message: string) => {
    if (serviceRef.current) {
      try {
        serviceRef.current.send(message);
      } catch (error) {
        console.error('Failed to send message:', error);
      }
    }
  }, []);

  return {
    status,
    sendMessage,
    lastMessage,
  };
}
