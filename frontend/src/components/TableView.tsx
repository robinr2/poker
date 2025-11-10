import { useState, useEffect } from 'react';

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
  handInProgress?: boolean;
  minRaise?: number;
  maxRaise?: number;
  playerBets?: Record<number, number>;
  boardCards?: string[];
  street?: string;
  showdown?: {
    winnerSeats: number[];
    winningHand: string;
    potAmount: number;
    amountsWon: Record<number, number>;
  };
  handComplete?: {
    message: string;
  };
}

interface TableViewProps {
  tableId: string;
  seats: SeatInfo[];
  currentSeatIndex: number | null;
  onLeave: () => void;
  gameState?: GameState;
  onSendMessage?: (message: string) => void;
  sendAction?: (action: string, amount?: number) => void;
  sendStartHand?: () => void;
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

// Helper function to get player names from seat indices
function getPlayerNamesFromSeats(
  seatIndices: number[],
  seats: SeatInfo[]
): string[] {
  return seatIndices
    .map((index) => seats[index]?.playerName || `Seat ${index}`)
    .filter((name) => name !== null && name !== undefined);
}

export function TableView({
  tableId,
  seats,
  currentSeatIndex,
  onLeave,
  gameState,
  onSendMessage,
  sendAction, // TODO: Use sendAction for player actions to enable optimistic updates
  sendStartHand,
}: TableViewProps) {
  // Suppress unused warning for sendAction - will be used in future optimistic updates
  void sendAction;

  const [raiseAmount, setRaiseAmount] = useState<string>('');
  const [showShowdown, setShowShowdown] = useState<boolean>(true);

  // Reset showShowdown when a new showdown appears
  useEffect(() => {
    if (gameState?.showdown) {
      setShowShowdown(true);
    }
  }, [gameState?.showdown]);

  console.log('[TableView] render with gameState:', gameState);
  console.log('[TableView] currentSeatIndex:', currentSeatIndex);
  console.log(
    '[TableView] seats:',
    seats.map((s) => ({
      index: s.index,
      name: s.playerName,
      cardCount: s.cardCount,
    }))
  );

  // Determine if Start Hand button should be visible
  // Show when: player is seated AND (no hand in progress OR hand complete)
  const isSeated = currentSeatIndex !== null && currentSeatIndex !== undefined;
  const hasHoleCards = gameState?.holeCards !== null && gameState?.holeCards !== undefined;
  const hasDealer = gameState?.dealerSeat !== null && gameState?.dealerSeat !== undefined;
  const handInProgress = hasHoleCards || hasDealer;
  const isHandComplete = gameState?.handComplete !== undefined;
  const showStartHandButton = isSeated && (!handInProgress || isHandComplete);

  const handleStartHand = () => {
    if (sendStartHand) {
      sendStartHand();
    } else if (onSendMessage) {
      // Fallback for backward compatibility
      const message = JSON.stringify({
        type: 'start_hand',
        payload: {},
      });
      onSendMessage(message);
    }
  };

  const handleAction = (action: string, amount?: number) => {
    if (onSendMessage && currentSeatIndex !== null) {
      const payload: Record<string, unknown> = {
        seatIndex: currentSeatIndex,
        action: action,
      };
      if (amount !== undefined) {
        payload.amount = amount;
      }
      const message = JSON.stringify({
        type: 'player_action',
        payload,
      });
      onSendMessage(message);
      // Clear raise input after any action
      setRaiseAmount('');
    }
  };

  // Raise button validation logic
  const isRaiseValid = gameState?.validActions?.includes('raise') || false;
  const raiseAmountNum = raiseAmount ? parseInt(raiseAmount, 10) : 0;
  const isRaiseAmountValid =
    raiseAmountNum >= (gameState?.minRaise ?? 0) &&
    raiseAmountNum <= (gameState?.maxRaise ?? 0);

  const handleMinRaise = () => {
    setRaiseAmount((gameState?.minRaise ?? 0).toString());
  };

  const handlePotRaise = () => {
    const potSized = (gameState?.callAmount ?? 0) + (gameState?.pot ?? 0);
    setRaiseAmount(potSized.toString());
  };

   const handleAllIn = () => {
     setRaiseAmount((gameState?.maxRaise ?? 0).toString());
   };

  return (
    <div className="table-view">
      <h1>Table: {tableId}</h1>
      <div className="table-container">
        <div className="seats-grid">
          {seats.map((seat) => (
            <div
              key={seat.index}
              className={`seat ${seat.index === currentSeatIndex ? 'own-seat' : ''} ${gameState?.currentActor === seat.index ? 'turn-active' : ''} ${gameState?.showdown?.winnerSeats.includes(seat.index) ? 'winner-seat' : ''} ${gameState?.foldedPlayers?.includes(seat.index) ? 'folded' : ''}`}
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

              {/* Bet Amount Display */}
              {seat.playerName &&
                gameState?.playerBets &&
                gameState.playerBets[seat.index] !== undefined &&
                gameState.playerBets[seat.index] > 0 && (
                  <div className="bet-amount">
                    <span className="money-icon">$</span>{' '}
                    {gameState.playerBets[seat.index]}
                  </div>
                )}

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
                seat.cardCount &&
                seat.cardCount > 0 && (
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
                  <span className="chip-icon">⬤</span>{' '}
                  {seat.stack !== undefined ? seat.stack : 'N/A'}
                </p>
              )}
            </div>
          ))}
        </div>

        {/* Game Info Section */}
        {gameState &&
          !showStartHandButton && (
            <div className="game-info">
              {gameState.street && (
                <div className="street-indicator">
                  {gameState.street.charAt(0).toUpperCase() +
                    gameState.street.slice(1)}
                </div>
              )}
              <div className="pot-display">Pot: {gameState.pot}</div>
            </div>
          )}

        {/* Board Cards Display - show during hand and after completion until new hand starts */}
        {gameState &&
          !showStartHandButton && (
          <div className="board-cards">
            {Array.from({ length: 5 }).map((_, idx) => {
              const card = gameState?.boardCards?.[idx];
              const hasAnyBoardCards = gameState?.boardCards && gameState.boardCards.length > 0;
              return (
                <div
                  key={idx}
                  className={`board-card ${card ? 'face-up' : hasAnyBoardCards ? 'empty' : 'card-back'} ${card && isRedSuit(card) ? 'red-suit' : card ? 'black-suit' : ''}`}
                >
                  {card ? formatCardDisplay(card) : hasAnyBoardCards ? '' : '🂠'}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Showdown Overlay */}
      {gameState?.showdown && showShowdown && (
        <div
          className="showdown-overlay"
          onClick={() => setShowShowdown(false)}
        >
          <div
            className="showdown-content"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              className="showdown-close-button"
              onClick={() => setShowShowdown(false)}
              aria-label="Close"
            >
              ×
            </button>
            <div className="winning-hand">{gameState.showdown.winningHand}</div>
            <div className="winners">
              Winners:{' '}
              {getPlayerNamesFromSeats(
                gameState.showdown.winnerSeats,
                seats
              ).join(', ')}
            </div>
            <div className="pot-amount">
              Pot: {gameState.showdown.potAmount}
            </div>
            {gameState.handComplete && (
              <div className="hand-complete-message">
                {gameState.handComplete.message}
              </div>
            )}
          </div>
        </div>
      )}

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

          {/* Raise Controls */}
          {isRaiseValid && (
            <div className="raise-controls">
              <input
                type="text"
                aria-label="Raise Amount"
                value={raiseAmount}
                onChange={(e) => setRaiseAmount(e.target.value)}
                placeholder="Raise amount"
                className="raise-input"
              />
              <button
                onClick={() => handleMinRaise()}
                className="preset-button"
              >
                Min
              </button>
              <button
                onClick={() => handlePotRaise()}
                className="preset-button"
              >
                Pot
              </button>
              <button onClick={() => handleAllIn()} className="preset-button">
                All-in
              </button>
              <button
                onClick={() => handleAction('raise', raiseAmountNum)}
                disabled={!isRaiseAmountValid || raiseAmount === ''}
                className="raise-button"
              >
                Raise
              </button>
            </div>
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
