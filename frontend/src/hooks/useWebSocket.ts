import { useEffect, useState, useCallback, useRef } from 'react';

import type { TableInfo } from '../components/TableCard';
import {
  WebSocketService,
  type ConnectionStatus,
} from '../services/WebSocketService';

interface UseWebSocketReturn {
  status: ConnectionStatus;
  sendMessage: (message: string) => void;
  lastMessage: string | null;
  lobbyState: TableInfo[];
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
  const [lobbyState, setLobbyState] = useState<TableInfo[]>([]);
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

      // Parse and handle lobby_state messages
      try {
        const message = JSON.parse(data);
        if (message.type === 'lobby_state' && message.payload) {
          // Payload is an array of table objects with snake_case fields
          const tables = JSON.parse(message.payload);
          const convertedTables: TableInfo[] = tables.map((t: {
            id: string;
            name: string;
            seats_occupied: number;
            max_seats: number;
          }) => ({
            id: t.id,
            name: t.name,
            seatsOccupied: t.seats_occupied,
            maxSeats: t.max_seats,
          }));
          setLobbyState(convertedTables);
        }
      } catch {
        // Silently ignore parsing errors for non-JSON messages
      }
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
    lobbyState,
  };
}
