import { useEffect, useState, useCallback, useRef } from 'react';

import { LobbyView } from './components/LobbyView';
import { NamePrompt } from './components/NamePrompt';
import { useWebSocket } from './hooks/useWebSocket';
import { SessionService } from './services/SessionService';
import './App.css';

interface SessionMessage {
  type: string;
  payload?: Record<string, unknown>;
}

const WS_URL = 'ws://localhost:8080/ws';

function App() {
  const [playerName, setPlayerName] = useState<string | null>(null);
  const [showPrompt, setShowPrompt] = useState(false);
  const [initialToken] = useState<string | null>(() => SessionService.getToken());
  
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
      }
      // lobby_state is now handled by useWebSocket hook
    } catch (error) {
      console.error('Failed to parse message:', error);
    }
  }, []); // Empty deps - we use refs for current values

  const { status, sendMessage, lobbyState } = useWebSocket(
    WS_URL,
    initialToken || undefined,
    { onMessage: handleMessage }
  );

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
    // Phase 4 will implement actual join logic
    console.log('Join table:', tableId);
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

        {!showPrompt && <LobbyView tables={lobbyState} onJoinTable={handleJoinTable} />}
      </main>
    </div>
  );
}

export default App;
