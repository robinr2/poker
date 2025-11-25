/**
 * Deterministic testing helpers for poker e2e tests
 *
 * These helpers allow tests to set a predetermined deck order so that
 * tests can verify exact outcomes (who wins, etc.)
 *
 * IMPORTANT: The server must be running with POKER_TEST_MODE=true
 * for these helpers to work.
 */

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

/**
 * Card string format: Rank + Suit
 * Ranks: A, 2, 3, 4, 5, 6, 7, 8, 9, T, J, Q, K
 * Suits: s (spades), h (hearts), d (diamonds), c (clubs)
 * Examples: "As" = Ace of spades, "Kh" = King of hearts, "Td" = Ten of diamonds
 */
export type CardString = string;

/**
 * Set a predetermined deck for a table
 * The deck will be used on the next hand start, then cleared
 *
 * @param tableId - The table ID (e.g., "table-1")
 * @param deck - Array of 52 card strings in deal order
 * @returns Promise that resolves when deck is set
 * @throws Error if server is not in test mode or deck is invalid
 */
export async function setDeck(tableId: string, deck: CardString[]): Promise<void> {
  const response = await fetch(`${BASE_URL}/api/test/set-deck`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      tableId,
      deck,
    }),
  });

  const result = await response.json();

  if (!response.ok || !result.success) {
    throw new Error(`Failed to set deck: ${result.message}`);
  }
}

/**
 * Reset a table to its initial state
 * Clears any ongoing hand, test deck, and resets dealer position
 * Players keep their seats and stacks
 *
 * @param tableId - The table ID (e.g., "table-1")
 * @returns Promise that resolves when table is reset
 * @throws Error if server is not in test mode or table not found
 */
export async function resetTable(tableId: string): Promise<void> {
  const response = await fetch(`${BASE_URL}/api/test/reset-table`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      tableId,
    }),
  });

  const result = await response.json();

  if (!response.ok || !result.success) {
    throw new Error(`Failed to reset table: ${result.message}`);
  }
}

/**
 * All 52 cards in a standard deck
 */
const ALL_CARDS: CardString[] = [
  // Spades
  'As', '2s', '3s', '4s', '5s', '6s', '7s', '8s', '9s', 'Ts', 'Js', 'Qs', 'Ks',
  // Hearts
  'Ah', '2h', '3h', '4h', '5h', '6h', '7h', '8h', '9h', 'Th', 'Jh', 'Qh', 'Kh',
  // Diamonds
  'Ad', '2d', '3d', '4d', '5d', '6d', '7d', '8d', '9d', 'Td', 'Jd', 'Qd', 'Kd',
  // Clubs
  'Ac', '2c', '3c', '4c', '5c', '6c', '7c', '8c', '9c', 'Tc', 'Jc', 'Qc', 'Kc',
];

/**
 * Configuration for building a deterministic deck
 */
export interface DeckSetup {
  /** Hole cards per seat index. Key is seat number (0-5), value is [card1, card2] */
  holeCards: Record<number, [CardString, CardString]>;
  /** The three flop cards (optional - will use filler cards if not provided) */
  flop?: [CardString, CardString, CardString];
  /** The turn card (optional - will use filler card if not provided) */
  turn?: CardString;
  /** The river card (optional - will use filler card if not provided) */
  river?: CardString;
}

/**
 * Build a complete 52-card deck with specific cards in specific positions
 * Cards are dealt in order from the front of the array
 *
 * Deal order for a multi-player game:
 * - Cards are dealt to active seats in ascending seat order
 * - Each player gets 2 hole cards
 * - Then: burn card, 3 flop cards
 * - Then: burn card, 1 turn card
 * - Then: burn card, 1 river card
 *
 * @param setup - Object mapping positions to specific cards
 * @returns Complete 52-card deck array
 *
 * @example
 * // Create a deck where seat 0 gets AA, seat 1 gets KK,
 * // and the board runs out A-2-7-3-9
 * const deck = buildDeck({
 *   holeCards: {
 *     0: ['As', 'Ah'],  // Seat 0: Pocket Aces
 *     1: ['Ks', 'Kh'],  // Seat 1: Pocket Kings
 *   },
 *   flop: ['Ad', '2c', '7d'],  // Seat 0 hits set
 *   turn: '3h',
 *   river: '9s',
 * });
 */
export function buildDeck(setup: DeckSetup): CardString[] {
  const usedCards = new Set<CardString>();
  const deck: CardString[] = [];

  // Helper to add a card and track usage
  const addCard = (card: CardString) => {
    if (usedCards.has(card)) {
      throw new Error(`Duplicate card in deck setup: ${card}`);
    }
    usedCards.add(card);
    deck.push(card);
  };

  // Helper to add a filler card (any unused card)
  const addFiller = () => {
    const available = ALL_CARDS.find((c) => !usedCards.has(c));
    if (!available) {
      throw new Error('No cards left for filler');
    }
    usedCards.add(available);
    deck.push(available);
  };

  // Add hole cards in seat order (0-5)
  // Cards are dealt to active seats in ascending order
  const seatNumbers = Object.keys(setup.holeCards)
    .map(Number)
    .sort((a, b) => a - b);

  for (const seatNum of seatNumbers) {
    const [card1, card2] = setup.holeCards[seatNum];
    addCard(card1);
    addCard(card2);
  }

  // Add burn + flop
  addFiller(); // Burn card before flop
  if (setup.flop) {
    addCard(setup.flop[0]);
    addCard(setup.flop[1]);
    addCard(setup.flop[2]);
  } else {
    addFiller();
    addFiller();
    addFiller();
  }

  // Add burn + turn
  addFiller(); // Burn card before turn
  if (setup.turn) {
    addCard(setup.turn);
  } else {
    addFiller();
  }

  // Add burn + river
  addFiller(); // Burn card before river
  if (setup.river) {
    addCard(setup.river);
  } else {
    addFiller();
  }

  // Fill remaining deck with unused cards (in case game needs them)
  while (deck.length < 52) {
    addFiller();
  }

  return deck;
}

/**
 * Pre-built deck scenarios for common test cases
 */
export const DECK_SCENARIOS = {
  /**
   * Player at seat 0 wins with trip Aces vs seat 1's pair of Kings
   * Assumes seats 0 and 1 are the active players
   */
  seat0WinsTrips: buildDeck({
    holeCards: {
      0: ['As', 'Ah'], // Pocket Aces
      1: ['Ks', 'Kh'], // Pocket Kings
    },
    flop: ['Ad', '2c', '7d'], // Seat 0 hits set of Aces
    turn: '3h',
    river: '9s',
  }),

  /**
   * Player at seat 1 wins with a flush
   * Assumes seats 0 and 1 are the active players
   */
  seat1WinsFlush: buildDeck({
    holeCards: {
      0: ['As', 'Kc'], // Ace-King offsuit
      1: ['2h', '3h'], // Low hearts (will make flush)
    },
    flop: ['4h', '5h', '9h'], // Flush for seat 1
    turn: '2c',
    river: '8d',
  }),

  /**
   * Split pot scenario - both players have same straight
   * Board makes the straight, both players play the board
   */
  splitPot: buildDeck({
    holeCards: {
      0: ['2s', '3c'],
      1: ['2d', '4c'],
    },
    flop: ['Ts', 'Jh', 'Qd'],
    turn: 'Kc',
    river: 'Ac', // Both have A-high straight on board
  }),

  /**
   * Three-player scenario: seat 0 wins with full house
   * Assumes seats 0, 1, and 2 are active
   */
  threePlayerSeat0Wins: buildDeck({
    holeCards: {
      0: ['As', 'Ah'], // Pocket Aces
      1: ['Ks', 'Kh'], // Pocket Kings
      2: ['Qs', 'Qh'], // Pocket Queens
    },
    flop: ['Ad', 'Kd', '2c'], // Set for seat 0, set for seat 1
    turn: 'Kc', // Full house for seat 0 (AAA-KK), quads for seat 1 (but AAA-KK > KKKK)
    river: '3h',
  }),

  /**
   * Three-player scenario: seat 1 wins with quads
   */
  threePlayerSeat1WinsQuads: buildDeck({
    holeCards: {
      0: ['As', 'Ah'], // Pocket Aces
      1: ['Ks', 'Kh'], // Pocket Kings
      2: ['Qs', 'Qh'], // Pocket Queens
    },
    flop: ['Kd', 'Kc', '2c'], // Quads for seat 1
    turn: 'Ad', // Full house for seat 0 (AAA-KK)
    river: '3h',
  }),
};
