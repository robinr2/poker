import { useEffect, useState, useCallback, useRef } from 'react';

import type { TableInfo } from '../components/TableCard';
import {
  WebSocketService,
  type ConnectionStatus,
} from '../services/WebSocketService';

interface SeatAssignedPayload {
  tableId: string;
  seatIndex: number;
  status?: string;
}

interface SeatMessage {
  type: 'seat_assigned' | 'seat_cleared';
  payload?: SeatAssignedPayload | Record<string, unknown>;
}

interface TableSeat {
  index: number;
  playerName: string | null;
  status: string;
  stack?: number;
  cardCount?: number;
}

interface TableState {
  tableId: string;
  seats: TableSeat[];
}

interface ShowdownState {
  winnerSeats: number[];
  winningHand: string;
  potAmount: number;
  amountsWon: Record<number, number>;
}

interface HandCompleteState {
  message: string;
}

interface GameState {
  dealerSeat: number | null;
  smallBlindSeat: number | null;
  bigBlindSeat: number | null;
  holeCards: [string, string] | null;
  pot: number;
  currentActor: number | null;
  validActions: string[] | null;
  callAmount: number | null;
  foldedPlayers: number[];
  roundOver: boolean | null;
  handInProgress?: boolean;
  minRaise?: number;
  maxRaise?: number;
  playerBets: Record<number, number>; // Track each player's bet amount in current round
  boardCards?: string[];
  street?: string;
  showdown?: ShowdownState;
  handComplete?: HandCompleteState;
}

interface HandStartedPayload {
  dealerSeat: number;
  smallBlindSeat: number;
  bigBlindSeat: number;
}

interface BlindPostedPayload {
  seatIndex: number;
  amount: number;
  newStack: number;
}

interface Card {
  Rank: string;
  Suit: string;
}

interface CardsDealtPayload {
  holeCards: Record<number, Card[]>;
}

interface ActionRequestPayload {
  seatIndex: number;
  validActions: string[];
  callAmount: number;
  minRaise?: number;
  maxRaise?: number;
}

interface ActionResultPayload {
  seatIndex: number;
  action: string;
  amountActed: number;
  newStack: number;
  pot: number;
  nextActor: number | null;
  roundOver: boolean;
  validActions?: string[];
}

interface BoardDealtPayload {
  boardCards: Card[];
  street: string;
}

interface UseWebSocketReturn {
  status: ConnectionStatus;
  sendMessage: (message: string) => void;
  sendAction?: (action: string, amount?: number) => void;
  lastMessage: string | null;
  lobbyState: TableInfo[];
  lastSeatMessage: SeatMessage | null;
  tableState: TableState | null;
  gameState: GameState;
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
  const [lastSeatMessage, setLastSeatMessage] = useState<SeatMessage | null>(
    null
  );
  const [tableState, setTableState] = useState<TableState | null>(null);
  const [playerSeatIndex, setPlayerSeatIndex] = useState<number | null>(null);
  const [gameState, setGameState] = useState<GameState>({
    dealerSeat: null,
    smallBlindSeat: null,
    bigBlindSeat: null,
    holeCards: null,
    pot: 0,
    currentActor: null,
    validActions: null,
    callAmount: null,
    foldedPlayers: [],
    roundOver: null,
    playerBets: {},
  });
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

      // Parse and handle messages
      try {
        const message = JSON.parse(data);
        if (message.type === 'lobby_state' && message.payload) {
          // Parse payload directly - backend sends it as an array, not a string
          const tables = message.payload as {
            id: string;
            name: string;
            seats_occupied: number;
            max_seats: number;
          }[];
          const convertedTables: TableInfo[] = tables.map((t) => ({
            id: t.id,
            name: t.name,
            seatsOccupied: t.seats_occupied,
            maxSeats: t.max_seats,
          }));
          setLobbyState(convertedTables);
        } else if (
          message.type === 'seat_assigned' ||
          message.type === 'seat_cleared'
        ) {
          // Store the seat message for the app to handle
          setLastSeatMessage({
            type: message.type,
            payload: message.payload,
          });

          // If seat_assigned, store the player's seat index
          if (message.type === 'seat_assigned' && message.payload) {
            const payload = message.payload as SeatAssignedPayload;
            setPlayerSeatIndex(payload.seatIndex);
          } else if (message.type === 'seat_cleared') {
            // Clear seat index when seat is cleared
            setPlayerSeatIndex(null);
          }
        } else if (message.type === 'table_state' && message.payload) {
          // Handle table_state message with seat information
          const payload = message.payload as {
            tableId: string;
            seats: Array<{
              index: number;
              playerName: string | null;
              status: string;
              stack?: number;
              cardCount?: number;
            }>;
            handInProgress?: boolean;
            dealerSeat?: number;
            smallBlindSeat?: number;
            bigBlindSeat?: number;
            pot?: number;
            holeCards?: { [seatIndex: string]: Card[] };
          };

          // Update table state with all seat information
          console.log(
            '[DEBUG] table_state payload:',
            JSON.stringify(payload, null, 2)
          );
          setTableState({
            tableId: payload.tableId,
            seats: payload.seats,
          });

          // Update game state with new fields if present
          if (
            payload.dealerSeat !== undefined ||
            payload.smallBlindSeat !== undefined ||
            payload.bigBlindSeat !== undefined ||
            payload.pot !== undefined ||
            payload.holeCards !== undefined
          ) {
            setGameState((prev) => {
              const updated = { ...prev };

              // Update game state positions if present
              if (payload.dealerSeat !== undefined) {
                updated.dealerSeat = payload.dealerSeat;
              }
              if (payload.smallBlindSeat !== undefined) {
                updated.smallBlindSeat = payload.smallBlindSeat;
              }
              if (payload.bigBlindSeat !== undefined) {
                updated.bigBlindSeat = payload.bigBlindSeat;
              }
              if (payload.pot !== undefined) {
                updated.pot = payload.pot;
              }

              // Update hole cards if present
              if (payload.holeCards) {
                const holeCardsEntries = Object.entries(payload.holeCards);
                if (holeCardsEntries.length > 0) {
                  const [, cards] = holeCardsEntries[0];
                  if (cards && cards.length === 2) {
                    const cardStrings: [string, string] = [
                      cards[0].Rank + cards[0].Suit,
                      cards[1].Rank + cards[1].Suit,
                    ];
                    updated.holeCards = cardStrings;
                  }
                }
              }

              return updated;
            });
          }
        } else if (message.type === 'hand_started' && message.payload) {
          // Handle hand_started message
          // Payload can be either a string (needs parsing) or already an object
          console.log(
            '[useWebSocket] hand_started received, payload:',
            message.payload
          );
          const payload =
            typeof message.payload === 'string'
              ? (JSON.parse(message.payload) as HandStartedPayload)
              : (message.payload as HandStartedPayload);
          console.log('[useWebSocket] hand_started payload:', payload);
          setGameState((prev) => ({
            ...prev,
            dealerSeat: payload.dealerSeat,
            smallBlindSeat: payload.smallBlindSeat,
            bigBlindSeat: payload.bigBlindSeat,
            playerBets: {}, // Clear player bets for new hand
          }));
          console.log(
            '[useWebSocket] gameState updated with dealer:',
            payload.dealerSeat,
            'SB:',
            payload.smallBlindSeat,
            'BB:',
            payload.bigBlindSeat
          );
        } else if (message.type === 'blind_posted' && message.payload) {
          // Handle blind_posted message - update pot and player bets
          // Payload can be either a string (needs parsing) or already an object
          console.log(
            '[useWebSocket] blind_posted received, payload:',
            message.payload
          );
          const payload =
            typeof message.payload === 'string'
              ? (JSON.parse(message.payload) as BlindPostedPayload)
              : (message.payload as BlindPostedPayload);
          console.log('[useWebSocket] blind_posted payload:', payload);
          setGameState((prev) => ({
            ...prev,
            pot: prev.pot + payload.amount,
            playerBets: {
              ...prev.playerBets,
              [payload.seatIndex]:
                (prev.playerBets[payload.seatIndex] || 0) + payload.amount,
            },
          }));
          console.log('[useWebSocket] gameState updated, pot:', payload.amount);
          // Also update table state with new stack
          setTableState((prev) => {
            if (!prev) return prev;
            return {
              ...prev,
              seats: prev.seats.map((seat) =>
                seat.index === payload.seatIndex
                  ? { ...seat, stack: payload.newStack }
                  : seat
              ),
            };
          });
        } else if (message.type === 'cards_dealt' && message.payload) {
          // Handle cards_dealt message
          // Payload can be either a string (needs parsing) or already an object
          console.log(
            '[useWebSocket] cards_dealt received, payload:',
            message.payload
          );
          const payload =
            typeof message.payload === 'string'
              ? (JSON.parse(message.payload) as CardsDealtPayload)
              : (message.payload as CardsDealtPayload);
          console.log('[useWebSocket] cards_dealt payload:', payload);
          // Find the current player's hole cards
          // The backend sends personalized cards, so we need to find which seat index has cards
          const holeCardsEntries = Object.entries(payload.holeCards);
          if (holeCardsEntries.length > 0) {
            const [, cards] = holeCardsEntries[0];
            if (cards && cards.length === 2) {
              // Convert Card objects {Rank, Suit} to string format "As", "Kh"
              const cardStrings: [string, string] = [
                cards[0].Rank + cards[0].Suit,
                cards[1].Rank + cards[1].Suit,
              ];
              setGameState((prev) => ({
                ...prev,
                holeCards: cardStrings,
              }));
              console.log(
                '[useWebSocket] gameState updated with hole cards:',
                cardStrings
              );
            }
          }
        } else if (message.type === 'action_request' && message.payload) {
          // Handle action_request message
          console.log(
            '[useWebSocket] action_request received, payload:',
            message.payload
          );
          const payload =
            typeof message.payload === 'string'
              ? (JSON.parse(message.payload) as ActionRequestPayload)
              : (message.payload as ActionRequestPayload);
          console.log('[useWebSocket] action_request payload:', payload);
          setGameState((prev) => {
            const updated = { ...prev };
            updated.currentActor = payload.seatIndex;
            updated.validActions = payload.validActions;
            updated.callAmount = payload.callAmount;

            // Include minRaise and maxRaise if present in payload
            if (payload.minRaise !== undefined) {
              updated.minRaise = payload.minRaise;
            }
            if (payload.maxRaise !== undefined) {
              updated.maxRaise = payload.maxRaise;
            }

            return updated;
          });
        } else if (message.type === 'action_result' && message.payload) {
          // Handle action_result message
          console.log(
            '[useWebSocket] action_result received, payload:',
            message.payload
          );
          const payload =
            typeof message.payload === 'string'
              ? (JSON.parse(message.payload) as ActionResultPayload)
              : (message.payload as ActionResultPayload);
          console.log('[useWebSocket] action_result payload:', payload);

          // Update table state with new stack
          setTableState((prev) => {
            if (!prev) return prev;
            return {
              ...prev,
              seats: prev.seats.map((seat) =>
                seat.index === payload.seatIndex
                  ? { ...seat, stack: payload.newStack }
                  : seat
              ),
            };
          });

          // Update game state
          setGameState((prev) => {
            const updated = { ...prev };
            updated.pot = payload.pot;
            updated.currentActor = payload.nextActor;
            updated.roundOver = payload.roundOver;

            // Update player bets based on action
            if (payload.action === 'call' || payload.action === 'raise') {
              // Add the amount acted to the player's current bet
              updated.playerBets = {
                ...prev.playerBets,
                [payload.seatIndex]:
                  (prev.playerBets[payload.seatIndex] || 0) +
                  payload.amountActed,
              };
            } else if (payload.action === 'check') {
              // Check doesn't add to bets, but ensure player has entry
              if (!(payload.seatIndex in prev.playerBets)) {
                updated.playerBets = {
                  ...prev.playerBets,
                  [payload.seatIndex]: 0,
                };
              }
            }
            // Note: fold doesn't update playerBets, keeps their existing bet visible

            // If player folded, add to folded players list
            if (payload.action === 'fold') {
              updated.foldedPlayers = [
                ...prev.foldedPlayers,
                payload.seatIndex,
              ];
            }

            // Update valid actions if provided
            if (payload.validActions) {
              updated.validActions = payload.validActions;
            }

            return updated;
          });
        } else if (message.type === 'board_dealt' && message.payload) {
           // Handle board_dealt message
           console.log(
             '[useWebSocket] board_dealt received, payload:',
             message.payload
           );
           const payload =
             typeof message.payload === 'string'
               ? (JSON.parse(message.payload) as BoardDealtPayload)
               : (message.payload as BoardDealtPayload);
           console.log('[useWebSocket] board_dealt payload:', payload);

           // Convert Card objects {Rank, Suit} to string format "As", "Kh", etc.
           const boardCardStrings = payload.boardCards.map(
             (card) => card.Rank + card.Suit
           );

           setGameState((prev) => ({
             ...prev,
             boardCards: boardCardStrings,
             street: payload.street,
           }));
           console.log(
             '[useWebSocket] gameState updated with boardCards:',
             boardCardStrings,
             'street:',
             payload.street
           );
         } else if (message.type === 'showdown_result' && message.payload) {
           // Handle showdown_result message
           console.log(
             '[useWebSocket] showdown_result received, payload:',
             message.payload
           );
           const payload =
             typeof message.payload === 'string'
               ? (JSON.parse(message.payload) as Record<string, unknown>)
               : (message.payload as Record<string, unknown>);
           console.log('[useWebSocket] showdown_result payload:', payload);

           // Convert amountsWon string keys to numbers
           const amountsWonRaw = payload.amountsWon as Record<string, number>;
           const amountsWon: Record<number, number> = {};
           if (amountsWonRaw) {
             Object.entries(amountsWonRaw).forEach(([key, value]) => {
               amountsWon[parseInt(key, 10)] = value;
             });
           }

           setGameState((prev) => ({
             ...prev,
             showdown: {
               winnerSeats: (payload.winnerSeats as number[]) || [],
               winningHand: (payload.winningHand as string) || '',
               potAmount: (payload.potAmount as number) || 0,
               amountsWon,
             },
           }));
           console.log('[useWebSocket] gameState updated with showdown');
         } else if (message.type === 'hand_complete' && message.payload) {
           // Handle hand_complete message
           console.log(
             '[useWebSocket] hand_complete received, payload:',
             message.payload
           );
           const payload =
             typeof message.payload === 'string'
               ? (JSON.parse(message.payload) as Record<string, unknown>)
               : (message.payload as Record<string, unknown>);
           console.log('[useWebSocket] hand_complete payload:', payload);

            setGameState((prev) => ({
              ...prev,
              handComplete: {
                message: (payload.message as string) || '',
              },
            }));
            console.log('[useWebSocket] gameState updated with handComplete');
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

  // Memoize sendAction to send player actions with optional amount for raises
  const sendAction = useCallback(
    (action: string, amount?: number) => {
      if (serviceRef.current && playerSeatIndex !== null) {
        try {
          const payload: Record<string, unknown> = {
            seatIndex: playerSeatIndex,
            action: action,
          };

          // Include amount only for raise actions
          if (amount !== undefined) {
            payload.amount = amount;
          }

          const message = JSON.stringify({
            type: 'player_action',
            payload: payload,
          });
          serviceRef.current.send(message);

          // Clear showdown and handComplete state when starting a new hand
          if (action === 'start_hand') {
            setGameState((prev) => {
              const updated = { ...prev };
              delete updated.showdown;
              delete updated.handComplete;
              return updated;
            });
          }
        } catch (error) {
          console.error('Failed to send action:', error);
        }
      }
    },
    [playerSeatIndex]
  );

  return {
    status,
    sendMessage,
    sendAction,
    lastMessage,
    lobbyState,
    lastSeatMessage,
    tableState,
    gameState,
  };
}
