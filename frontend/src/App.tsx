import { useEffect, useState, useCallback, useRef } from 'react';

import { LobbyView } from './components/LobbyView';
import { NamePrompt } from './components/NamePrompt';
import { TableView } from './components/TableView';
import { useWebSocket } from './hooks/useWebSocket';
import { SessionService } from './services/SessionService';
import './App.css';
import './styles/TableView.css';

interface SessionMessage {
  type: string;
  payload?: Record<string, unknown>;
}

interface SeatInfo {
  index: number;
  playerName: string | null;
  status: string;
  stack?: number;
}

const WS_URL = 'ws://localhost:8080/ws';

function App() {
  const [playerName, setPlayerName] = useState<string | null>(null);
  const [showPrompt, setShowPrompt] = useState(false);
  const [initialToken] = useState<string | null>(() =>
    SessionService.getToken()
  );
  const [view, setView] = useState<'lobby' | 'table'>('lobby');
  const [currentTableId, setCurrentTableId] = useState<string | null>(null);
  const [currentSeatIndex, setCurrentSeatIndex] = useState<number | null>(null);
  const [seats, setSeats] = useState<SeatInfo[]>([]);

  // Use refs to avoid stale closures in the message handler
  const playerNameRef = useRef(playerName);
  const showPromptRef = useRef(showPrompt);

  useEffect(() => {
    playerNameRef.current = playerName;
    showPromptRef.current = showPrompt;
  }, [playerName, showPrompt]);

  useEffect(() => {
    setShowPrompt(!initialToken); // Show prompt if no token
  }, [initialToken]);

  // Handle incoming WebSocket messages - use useCallback with empty deps
  // since we use refs for current values
  const handleMessage = useCallback((rawMessage: string) => {
    try {
      const message: SessionMessage = JSON.parse(rawMessage);

      if (message.type === 'session_created' && message.payload) {
        const token = message.payload.token as string;
        const name = message.payload.name as string;
        SessionService.setToken(token);
        setPlayerName(name);
        setShowPrompt(false);
      } else if (message.type === 'session_restored' && message.payload) {
        const name = message.payload.name as string;
        setPlayerName(name);
        setShowPrompt(false);
      } else if (message.type === 'error' && message.payload) {
        // Handle error messages (e.g., invalid token)
        const errorMessage = message.payload.message as string;
        console.error('[App] Error from server:', errorMessage);
        
        // If token is invalid/expired, clear it and show name prompt
        if (errorMessage && errorMessage.toLowerCase().includes('token')) {
          SessionService.clearToken();
          setPlayerName(null);
          setShowPrompt(true);
        }
      }
      // lobby_state is now handled by useWebSocket hook
    } catch (error) {
      console.error('Failed to parse message:', error);
    }
  }, []); // Empty deps - we use refs for current values

  const {
    status,
    sendMessage,
    sendAction,
    sendStartHand,
    lobbyState,
    lastSeatMessage,
    tableState,
    gameState,
  } = useWebSocket(WS_URL, initialToken || undefined, {
    onMessage: handleMessage,
  });

  // Debug: Log gameState changes
  useEffect(() => {
    console.log('[App] gameState updated:', gameState);
  }, [gameState]);

  // Handle seat messages (seat_assigned and seat_cleared)
  useEffect(() => {
    if (lastSeatMessage) {
      if (lastSeatMessage.type === 'seat_assigned' && lastSeatMessage.payload) {
        const payload = lastSeatMessage.payload as {
          tableId: string;
          seatIndex: number;
          status?: string;
        };
        setCurrentTableId(payload.tableId);
        setCurrentSeatIndex(payload.seatIndex);
        // Initialize seats array with 6 empty seats
        const newSeats: SeatInfo[] = [];
        for (let i = 0; i < 6; i++) {
          newSeats.push({
            index: i,
            playerName: i === payload.seatIndex ? playerNameRef.current : null,
            status: i === payload.seatIndex ? 'occupied' : 'empty',
          });
        }
        setSeats(newSeats);
        setView('table');
      } else if (lastSeatMessage.type === 'seat_cleared') {
        setView('lobby');
        setCurrentTableId(null);
        setCurrentSeatIndex(null);
        setSeats([]);
      }
    }
  }, [lastSeatMessage]);

  // Handle table state updates
  useEffect(() => {
    if (tableState) {
      const updatedSeats: SeatInfo[] = tableState.seats.map((seat) => ({
        index: seat.index,
        playerName: seat.playerName,
        status: seat.status,
        stack: seat.stack,
        cardCount: seat.cardCount,
      }));
      setSeats(updatedSeats);
    }
  }, [tableState]);

  const handleNameSubmit = (name: string): void => {
    try {
      const message = JSON.stringify({
        type: 'set_name',
        payload: { name },
      });
      sendMessage(message);
    } catch (error) {
      console.error('Failed to send name:', error);
    }
  };

  const handleJoinTable = (tableId: string): void => {
    try {
      const message = JSON.stringify({
        type: 'join_table',
        payload: { tableId },
      });
      sendMessage(message);
    } catch (error) {
      console.error('Failed to send join_table message:', error);
    }
  };

  const handleLeaveTable = (): void => {
    try {
      const message = JSON.stringify({
        type: 'leave_table',
        payload: {},
      });
      sendMessage(message);
    } catch (error) {
      console.error('Failed to send leave_table message:', error);
    }
  };

  const getStatusColor = (): string => {
    switch (status) {
      case 'connected':
        return '#10b981';
      case 'connecting':
        return '#f59e0b';
      case 'error':
        return '#ef4444';
      case 'disconnected':
      default:
        return '#6b7280';
    }
  };

  const getStatusText = (): string => {
    switch (status) {
      case 'connected':
        return 'Connected';
      case 'connecting':
        return 'Connecting...';
      case 'error':
        return 'Error';
      case 'disconnected':
      default:
        return 'Disconnected';
    }
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>Poker</h1>
        <div className="connection-status">
          <div
            className="status-indicator"
            style={{ backgroundColor: getStatusColor() }}
          ></div>
          <span className="status-text">{getStatusText()}</span>
          {playerName && <span className="player-name">• {playerName}</span>}
        </div>
      </header>

      <main className="app-main">
        {showPrompt && <NamePrompt onSubmit={handleNameSubmit} />}

        {!showPrompt && view === 'lobby' && (
          <LobbyView tables={lobbyState} onJoinTable={handleJoinTable} />
        )}
        {!showPrompt && view === 'table' && currentTableId && (
          <TableView
            tableId={currentTableId}
            seats={seats}
            currentSeatIndex={currentSeatIndex}
            onLeave={handleLeaveTable}
            gameState={gameState}
            onSendMessage={sendMessage}
            sendAction={sendAction}
            sendStartHand={sendStartHand}
          />
        )}
      </main>
    </div>
  );
}

export default App;
