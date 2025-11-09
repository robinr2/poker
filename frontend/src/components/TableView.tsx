interface SeatInfo {
  index: number;
  playerName: string | null;
  status: string;
  stack?: number;
  cardCount?: number;
}

interface GameState {
  dealerSeat: number | null;
  smallBlindSeat: number | null;
  bigBlindSeat: number | null;
  holeCards: [string, string] | null;
  pot: number;
  currentActor?: number | null;
  validActions?: string[] | null;
  callAmount?: number | null;
  foldedPlayers?: number[];
  roundOver?: boolean | null;
}

interface TableViewProps {
  tableId: string;
  seats: SeatInfo[];
  currentSeatIndex: number | null;
  onLeave: () => void;
  gameState?: GameState;
  onSendMessage?: (message: string) => void;
}

// Helper function to convert card string format to display format
// e.g., "As" -> "A♠", "Kh" -> "K♥"
function formatCardDisplay(card: string): string {
  if (card.length !== 2) return card;

  const rank = card[0];
  const suit = card[1];
  const suitSymbols: Record<string, string> = {
    s: '♠',
    h: '♥',
    d: '♦',
    c: '♣',
  };

  return rank + (suitSymbols[suit] || suit);
}

// Helper function to check if a card is a red suit (hearts or diamonds)
function isRedSuit(card: string): boolean {
  if (card.length !== 2) return false;
  const suit = card[1];
  return suit === 'h' || suit === 'd';
}

export function TableView({
  tableId,
  seats,
  currentSeatIndex,
  onLeave,
  gameState,
  onSendMessage,
}: TableViewProps) {
  console.log('[TableView] render with gameState:', gameState);
  console.log('[TableView] currentSeatIndex:', currentSeatIndex);
  console.log('[TableView] seats:', seats.map(s => ({ 
    index: s.index, 
    name: s.playerName, 
    cardCount: s.cardCount 
  })));
  
  // Determine if Start Hand button should be visible
  // Show only when: player is seated AND no active hand (pot === 0 or undefined)
  const isSeated = currentSeatIndex !== null && currentSeatIndex !== undefined;
  const hasNoActiveHand = !gameState || gameState.pot === 0 || gameState.pot === undefined;
  const showStartHandButton = isSeated && hasNoActiveHand;

  const handleStartHand = () => {
    if (onSendMessage) {
      const message = JSON.stringify({
        type: 'start_hand',
        payload: {},
      });
      onSendMessage(message);
    }
  };

  const handleAction = (action: string) => {
    if (onSendMessage && currentSeatIndex !== null) {
      const message = JSON.stringify({
        type: 'player_action',
        payload: {
          seatIndex: currentSeatIndex,
          action: action,
        },
      });
      onSendMessage(message);
    }
  };

  return (
    <div className="table-view">
      <h1>Table: {tableId}</h1>
      <div className="table-container">
        <div className="seats-grid">
           {seats.map((seat) => (
             <div
               key={seat.index}
               className={`seat ${seat.index === currentSeatIndex ? 'own-seat' : ''} ${gameState?.currentActor === seat.index ? 'turn-active' : ''}`}
             >
              {/* Dealer Button */}
              {gameState?.dealerSeat === seat.index && (
                <span className="dealer-badge">D</span>
              )}

              {/* Blind Indicators */}
              {gameState?.smallBlindSeat === seat.index && (
                <span className="sb-badge">SB</span>
              )}
              {gameState?.bigBlindSeat === seat.index && (
                <span className="bb-badge">BB</span>
              )}

              <p className="seat-number">Seat {seat.index}</p>
              <p className="seat-player">{seat.playerName || 'Empty'}</p>

              {/* Hole Cards Display */}
              {seat.index === currentSeatIndex && gameState?.holeCards && (
                <div className="hole-cards">
                  {gameState.holeCards.map((card, idx) => (
                    <span 
                      key={idx} 
                      className={`card face-up ${isRedSuit(card) ? 'red-suit' : 'black-suit'}`}
                    >
                      {formatCardDisplay(card)}
                    </span>
                  ))}
                </div>
              )}

               {/* Card Backs for Opponents */}
               {seat.index !== currentSeatIndex &&
                 seat.playerName &&
                 seat.cardCount && seat.cardCount > 0 && (
                   <div className="opponent-cards">
                     {Array.from({ length: seat.cardCount }).map((_, idx) => (
                       <span key={idx} className="card card-back">
                         🂠
                       </span>
                     ))}
                   </div>
                 )}

              {/* Chip Stack Display */}
              {seat.playerName && (
                <p className="stack">
                  💰 {seat.stack !== undefined ? seat.stack : 'N/A'}
                </p>
              )}
            </div>
          ))}
        </div>

        {/* Pot Display in Center */}
         {gameState && <div className="pot-display">Pot: {gameState.pot}</div>}
       </div>

       {/* Action Bar */}
       {gameState?.currentActor === currentSeatIndex && (
         <div className="action-bar">
           <button onClick={() => handleAction('fold')}>Fold</button>
           {gameState.callAmount === 0 ? (
             <button onClick={() => handleAction('check')}>Check</button>
           ) : (
             <button onClick={() => handleAction('call')}>
               Call {gameState.callAmount}
             </button>
           )}
         </div>
       )}

       <div className="button-group">
        {showStartHandButton && (
          <button onClick={handleStartHand} className="start-hand-button">
            Start Hand
          </button>
        )}
        <button onClick={onLeave} className="leave-button">
          Leave Table
        </button>
      </div>
    </div>
  );
}
