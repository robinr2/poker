import { useWebSocket } from './hooks/useWebSocket'
import './App.css'

function App() {
  const { status, lastMessage } = useWebSocket('ws://localhost:8080/ws')

  const getStatusColor = (): string => {
    switch (status) {
      case 'connected':
        return '#10b981'
      case 'connecting':
        return '#f59e0b'
      case 'error':
        return '#ef4444'
      case 'disconnected':
      default:
        return '#6b7280'
    }
  }

  const getStatusText = (): string => {
    switch (status) {
      case 'connected':
        return 'Connected'
      case 'connecting':
        return 'Connecting...'
      case 'error':
        return 'Error'
      case 'disconnected':
      default:
        return 'Disconnected'
    }
  }

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
        </div>
      </header>

      <main className="app-main">
        <div className="status-card">
          <h2>WebSocket Status</h2>
          <p>Status: <strong>{status}</strong></p>
          {lastMessage && (
            <p>Last Message: <code>{lastMessage.substring(0, 100)}</code></p>
          )}
        </div>
      </main>
    </div>
  )
}

export default App
