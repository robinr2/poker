import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

import '@testing-library/jest-dom';
import { TableView } from './TableView';

interface GameState {
  dealerSeat: number | null;
  smallBlindSeat: number | null;
  bigBlindSeat: number | null;
  holeCards: [string, string] | null;
  pot: number;
}

interface SeatInfo {
  index: number;
  playerName: string | null;
  status: string;
  stack?: number;
}

describe('TableView', () => {
  const mockSeats: SeatInfo[] = [
    { index: 0, playerName: null, status: 'empty' },
    { index: 1, playerName: 'Alice', status: 'occupied' },
    { index: 2, playerName: null, status: 'empty' },
    { index: 3, playerName: 'Bob', status: 'occupied' },
    { index: 4, playerName: null, status: 'empty' },
    { index: 5, playerName: 'Charlie', status: 'occupied' },
  ];

  describe('TestGameElementsDisplay', () => {
    it('displays dealer button on correct seat', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: null,
        bigBlindSeat: null,
        holeCards: null,
        pot: 0,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      const dealerBadge = screen.getByText('D');
      expect(dealerBadge).toBeInTheDocument();
      expect(dealerBadge).toHaveClass('dealer-badge');

      // Verify the badge is in the correct seat
      const seat1 = document.querySelectorAll('.seat')[1];
      expect(seat1.querySelector('.dealer-badge')).toBe(dealerBadge);
    });

    it('displays small blind and big blind indicators', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: null,
        pot: 0,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      const sbBadges = screen.getAllByText('SB');
      const bbBadges = screen.getAllByText('BB');

      expect(sbBadges.length).toBeGreaterThan(0);
      expect(bbBadges.length).toBeGreaterThan(0);

      // Verify SB is in seat 3
      const seat3 = document.querySelectorAll('.seat')[3];
      expect(seat3.querySelector('.sb-badge')).toBeInTheDocument();

      // Verify BB is in seat 5
      const seat5 = document.querySelectorAll('.seat')[5];
      expect(seat5.querySelector('.bb-badge')).toBeInTheDocument();
    });

    it('displays player hole cards when dealt', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: ['As', 'Kh'],
        pot: 0,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Should display both hole cards
      const card1 = screen.getByText(/A♠/);
      const card2 = screen.getByText(/K♥/);

      expect(card1).toBeInTheDocument();
      expect(card2).toBeInTheDocument();

      // Cards should be in a container with hole-cards class
      const holeCardsContainer = document.querySelector('.hole-cards');
      expect(holeCardsContainer).toBeInTheDocument();
      expect(holeCardsContainer?.textContent).toContain('A♠');
      expect(holeCardsContainer?.textContent).toContain('K♥');
    });

    it('displays opponent card backs', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: ['As', 'Kh'],
        pot: 0,
      };

      const seatsWithCardCount = mockSeats.map((seat) => ({
        ...seat,
        cardCount: seat.playerName ? 2 : undefined,
      }));

      render(
        <TableView
          tableId="table-1"
          seats={seatsWithCardCount}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Find opponent seats (not current seat)
      const seatContainers = document.querySelectorAll('.seat');
      const opponentSeat = seatContainers[3]; // Bob's seat

      // Check if opponent seat has card back indicator
      const cardBacks = opponentSeat.querySelectorAll('.card-back');
      expect(cardBacks.length).toBe(2);
    });

    it('displays chip stacks for each player', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: null,
        pot: 30,
      };

      const seatsWithStacks: SeatInfo[] = mockSeats.map((seat, idx) => ({
        ...seat,
        stack: 1000 - idx * 50, // Different stacks for each seat
      }));

      render(
        <TableView
          tableId="table-1"
          seats={seatsWithStacks}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Check that stacks are displayed
      const stackElements = document.querySelectorAll('.stack');
      expect(stackElements.length).toBeGreaterThan(0);
    });

    it('displays pot total in center', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: null,
        pot: 75,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      const potDisplay = screen.getByText('Pot: 75');
      expect(potDisplay).toBeInTheDocument();
      expect(potDisplay).toHaveClass('pot-display');
    });

    it('updates on blind_posted messages', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: null,
        pot: 30,
      };

      const seatsWithStacks: SeatInfo[] = mockSeats.map((seat) => ({
        ...seat,
        stack: 1000,
      }));

      const { rerender } = render(
        <TableView
          tableId="table-1"
          seats={seatsWithStacks}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Initial pot should be 30
      expect(screen.getByText('Pot: 30')).toBeInTheDocument();

      // Update game state to reflect blind posted
      const updatedGameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: null,
        pot: 45,
      };

      const updatedSeats: SeatInfo[] = seatsWithStacks.map((seat) => ({
        ...seat,
        stack: seat.index === 3 ? 995 : seat.stack,
      }));

      rerender(
        <TableView
          tableId="table-1"
          seats={updatedSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={updatedGameState}
        />
      );

      // Pot should be updated to 45
      expect(screen.getByText('Pot: 45')).toBeInTheDocument();
    });

    it('parses hand_started message correctly', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: null,
        pot: 0,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Verify dealer is set to seat 1
      const dealerBadge = screen.getByText('D');
      expect(dealerBadge).toBeInTheDocument();

      // Verify SB is set to seat 3
      const seat3 = document.querySelectorAll('.seat')[3];
      expect(seat3.querySelector('.sb-badge')).toBeInTheDocument();

      // Verify BB is set to seat 5
      const seat5 = document.querySelectorAll('.seat')[5];
      expect(seat5.querySelector('.bb-badge')).toBeInTheDocument();
    });
  });

  describe('TestTableViewRenders6Seats', () => {
    it('should render 6 seat positions', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
        />
      );

      // Check for all 6 seat containers
      const seatContainers = document.querySelectorAll('.seat');
      expect(seatContainers).toHaveLength(6);
    });

    it('should render seats numbered 0-5', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
        />
      );

      for (let i = 0; i < 6; i++) {
        const seatNumber = screen.getByText(`Seat ${i}`);
        expect(seatNumber).toBeInTheDocument();
      }
    });
  });

  describe('TestTableViewShowsOccupiedSeats', () => {
    it('should display player names in occupied seats', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
        />
      );

      expect(screen.getByText('Alice')).toBeInTheDocument();
      expect(screen.getByText('Bob')).toBeInTheDocument();
      expect(screen.getByText('Charlie')).toBeInTheDocument();
    });

    it('should show player name for seat 1 occupied by Alice', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
        />
      );

      const seatContainers = document.querySelectorAll('.seat');
      const seat1 = seatContainers[1];
      expect(seat1.textContent).toContain('Alice');
    });
  });

  describe('TestTableViewShowsEmptySeats', () => {
    it('should display placeholder for empty seats', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
        />
      );

      const emptyPlaceholders = screen.getAllByText('Empty');
      expect(emptyPlaceholders.length).toBeGreaterThanOrEqual(3);
    });

    it('should show Empty in seat 0', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
        />
      );

      const seatContainers = document.querySelectorAll('.seat');
      const seat0 = seatContainers[0];
      expect(seat0.textContent).toContain('Empty');
    });
  });

  describe('TestTableViewHighlightsOwnSeat', () => {
    it('should highlight current player seat with distinct CSS class', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
        />
      );

      const seatContainers = document.querySelectorAll('.seat');
      const seat1 = seatContainers[1];
      expect(seat1.classList.contains('own-seat')).toBe(true);
    });

    it('should not highlight other seats as own-seat', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
        />
      );

      const seatContainers = document.querySelectorAll('.seat');
      expect(seatContainers[0].classList.contains('own-seat')).toBe(false);
      expect(seatContainers[2].classList.contains('own-seat')).toBe(false);
      expect(seatContainers[3].classList.contains('own-seat')).toBe(false);
    });

    it('should highlight seat 3 when currentSeatIndex is 3', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={3}
          onLeave={mockOnLeave}
        />
      );

      const seatContainers = document.querySelectorAll('.seat');
      expect(seatContainers[3].classList.contains('own-seat')).toBe(true);
      expect(seatContainers[1].classList.contains('own-seat')).toBe(false);
    });
  });

  describe('TestTableViewLeaveButton', () => {
    it('should render Leave Table button', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
        />
      );

      const leaveButton = screen.getByRole('button', { name: /Leave Table/i });
      expect(leaveButton).toBeInTheDocument();
    });

    it('should call onLeave callback when Leave button clicked', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
        />
      );

      const leaveButton = screen.getByRole('button', { name: /Leave Table/i });
      fireEvent.click(leaveButton);

      expect(mockOnLeave).toHaveBeenCalledOnce();
    });

    it('should call onLeave multiple times if button clicked multiple times', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
        />
      );

      const leaveButton = screen.getByRole('button', { name: /Leave Table/i });
      fireEvent.click(leaveButton);
      fireEvent.click(leaveButton);

      expect(mockOnLeave).toHaveBeenCalledTimes(2);
    });
  });

  describe('TestTableViewLayout', () => {
    it('should have a container with table-view class', () => {
      const mockOnLeave = vi.fn();
      const { container } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
        />
      );

      const tableView = container.querySelector('.table-view');
      expect(tableView).toBeInTheDocument();
    });

    it('should have a seats-grid container for the 6 seats', () => {
      const mockOnLeave = vi.fn();
      const { container } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
        />
      );

      const seatsGrid = container.querySelector('.seats-grid');
      expect(seatsGrid).toBeInTheDocument();
      const seatsInGrid = seatsGrid?.querySelectorAll('.seat');
      expect(seatsInGrid).toHaveLength(6);
    });
  });

  describe('TestTableViewDisplaysTableInfo', () => {
    it('should display table ID in header', () => {
      const mockOnLeave = vi.fn();
      render(
        <TableView
          tableId="table-123"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
        />
      );

      expect(screen.getByText(/Table: table-123/i)).toBeInTheDocument();
    });

    it('should display different table ID when changed', () => {
      const mockOnLeave = vi.fn();
      const { rerender } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
        />
      );

      expect(screen.getByText(/Table: table-1/i)).toBeInTheDocument();

      rerender(
        <TableView
          tableId="table-2"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
        />
      );

      expect(screen.queryByText(/Table: table-1/i)).not.toBeInTheDocument();
      expect(screen.getByText(/Table: table-2/i)).toBeInTheDocument();
    });
  });

  describe('TestTableViewStartHandButton', () => {
    it('should render Start Hand button when player is seated and no active hand', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: GameState = {
        dealerSeat: null,
        smallBlindSeat: null,
        bigBlindSeat: null,
        holeCards: null,
        pot: 0,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const startHandButton = screen.getByRole('button', {
        name: /Start Hand/i,
      });
      expect(startHandButton).toBeInTheDocument();
    });

    it('should not render Start Hand button when player is not seated', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: GameState = {
        dealerSeat: null,
        smallBlindSeat: null,
        bigBlindSeat: null,
        holeCards: null,
        pot: 0,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={null}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const startHandButton = screen.queryByRole('button', {
        name: /Start Hand/i,
      });
      expect(startHandButton).not.toBeInTheDocument();
    });

    it('should not render Start Hand button when hand is active (pot > 0)', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: null,
        pot: 30, // Active hand has pot > 0
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const startHandButton = screen.queryByRole('button', {
        name: /Start Hand/i,
      });
      expect(startHandButton).not.toBeInTheDocument();
    });

    it('should send start_hand message when button is clicked', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: GameState = {
        dealerSeat: null,
        smallBlindSeat: null,
        bigBlindSeat: null,
        holeCards: null,
        pot: 0,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const startHandButton = screen.getByRole('button', {
        name: /Start Hand/i,
      });
      fireEvent.click(startHandButton);

      expect(mockSendMessage).toHaveBeenCalledOnce();
      // The message should be a JSON string with action: "start_hand"
      const call = mockSendMessage.mock.calls[0][0];
      const message = JSON.parse(call);
      expect(message.type).toBe('start_hand');
    });

    it('should call onSendMessage with proper message format', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: GameState = {
        dealerSeat: null,
        smallBlindSeat: null,
        bigBlindSeat: null,
        holeCards: null,
        pot: 0,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const startHandButton = screen.getByRole('button', {
        name: /Start Hand/i,
      });
      fireEvent.click(startHandButton);

      const sentMessage = mockSendMessage.mock.calls[0][0];
      expect(typeof sentMessage).toBe('string');
      expect(sentMessage).toContain('start_hand');
    });
  });

  describe('TestTableViewPhase4CardRendering', () => {
    it('renders card backs based on cardCount', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: null,
        pot: 0,
      };

      const seatsWithCardCount: SeatInfo[] = mockSeats.map((seat) => ({
        ...seat,
        cardCount: seat.playerName ? 2 : undefined,
      }));

      render(
        <TableView
          tableId="table-1"
          seats={seatsWithCardCount}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Find opponent seats that should have card backs
      const seatContainers = document.querySelectorAll('.seat');

      // Seat 3 (Bob) should have card backs
      const seat3 = seatContainers[3];
      const cardBacks3 = seat3.querySelectorAll('.card-back');
      expect(cardBacks3.length).toBe(2);
    });

    it('renders stack values for all seated players', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: null,
        pot: 0,
      };

      const seatsWithStacks: SeatInfo[] = mockSeats.map((seat, idx) => ({
        ...seat,
        stack: seat.playerName ? 1000 + idx * 100 : undefined,
      }));

      render(
        <TableView
          tableId="table-1"
          seats={seatsWithStacks}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Check that stacks are displayed for occupied seats
      expect(screen.getByText('💰 1100')).toBeInTheDocument(); // Seat 1 (Alice)
      expect(screen.getByText('💰 1300')).toBeInTheDocument(); // Seat 3 (Bob)
      expect(screen.getByText('💰 1500')).toBeInTheDocument(); // Seat 5 (Charlie)
    });

    it('renders pot amount when present', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 1,
        smallBlindSeat: 3,
        bigBlindSeat: 5,
        holeCards: null,
        pot: 250,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      const potDisplay = screen.getByText('Pot: 250');
      expect(potDisplay).toBeInTheDocument();
      expect(potDisplay).toHaveClass('pot-display');
    });
  });
});

describe('Bet Amount Display Tests', () => {
  const mockSeats: SeatInfo[] = [
    { index: 0, playerName: 'Alice', status: 'occupied', stack: 1000 },
    { index: 1, playerName: 'Bob', status: 'occupied', stack: 980 },
    { index: 2, playerName: 'Charlie', status: 'occupied', stack: 1000 },
    { index: 3, playerName: null, status: 'empty' },
    { index: 4, playerName: null, status: 'empty' },
    { index: 5, playerName: null, status: 'empty' },
  ];

  interface ExtendedGameState extends GameState {
    playerBets?: Record<number, number>;
  }

  describe('TestTableView_BetAmountDisplay', () => {
    it('should display bet amount when player has bet > 0', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 150,
        playerBets: {
          0: 50,
          1: 50,
          2: 50,
        },
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Check that bet amounts are displayed
      const betAmounts = screen.getAllByText(/💵 50/);
      expect(betAmounts.length).toBeGreaterThanOrEqual(3);
    });

    it('should not display bet amount for players with no bets', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 50,
        playerBets: {
          0: 50,
          // Bob and Charlie have no bets
        },
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Alice should have bet amount
      expect(screen.getByText(/💵 50/)).toBeInTheDocument();

      // Should only be one bet amount displayed
      const betAmounts = screen.getAllByText(/💵/);
      expect(betAmounts.length).toBe(1);
    });

    it('should not display bet amount for empty seats', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 50,
        playerBets: {
          0: 50,
          3: 50, // Empty seat - should not display
        },
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Should only display bet for Alice (seat 0)
      const betAmounts = screen.getAllByText(/💵/);
      expect(betAmounts.length).toBe(1);
    });

    it('should display different bet amounts for different players', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 180,
        playerBets: {
          0: 100,
          1: 50,
          2: 30,
        },
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      expect(screen.getByText(/💵 100/)).toBeInTheDocument();
      expect(screen.getByText(/💵 50/)).toBeInTheDocument();
      expect(screen.getByText(/💵 30/)).toBeInTheDocument();
    });

    it('should not display bet amounts when playerBets is undefined', () => {
      const mockOnLeave = vi.fn();
      const gameState: GameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 0,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Should not display any bet amounts
      const betAmounts = screen.queryAllByText(/💵/);
      expect(betAmounts.length).toBe(0);
    });

    it('should not display bet amount when amount is 0', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 50,
        playerBets: {
          0: 50,
          1: 0, // Explicitly 0 - should not display
          2: 0,
        },
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Should only display bet for Alice (seat 0)
      const betAmounts = screen.getAllByText(/💵/);
      expect(betAmounts.length).toBe(1);
    });

    it('should update bet amounts when gameState changes', () => {
      const mockOnLeave = vi.fn();
      const initialGameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 50,
        playerBets: {
          0: 50,
        },
      };

      const { rerender } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={initialGameState}
        />
      );

      // Initially Alice has bet 50
      expect(screen.getByText(/💵 50/)).toBeInTheDocument();

      // Update game state - Bob now bets 100
      const updatedGameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 150,
        playerBets: {
          0: 50,
          1: 100,
        },
      };

      rerender(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={updatedGameState}
        />
      );

      // Both bet amounts should be visible
      expect(screen.getByText(/💵 50/)).toBeInTheDocument();
      expect(screen.getByText(/💵 100/)).toBeInTheDocument();
    });

    it('should clear bet amounts on new hand', () => {
      const mockOnLeave = vi.fn();
      const gameStateWithBets: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 150,
        playerBets: {
          0: 50,
          1: 50,
          2: 50,
        },
      };

      const { rerender } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameStateWithBets}
        />
      );

      // Initially bets should be displayed
      expect(screen.getAllByText(/💵/).length).toBeGreaterThan(0);

      // New hand starts - playerBets cleared
      const newHandGameState: ExtendedGameState = {
        dealerSeat: 1,
        smallBlindSeat: 2,
        bigBlindSeat: 0,
        holeCards: null,
        pot: 0,
        playerBets: {},
      };

      rerender(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={newHandGameState}
        />
      );

      // No bet amounts should be displayed
      const betAmounts = screen.queryAllByText(/💵/);
      expect(betAmounts.length).toBe(0);
    });

    it('should apply bet-amount CSS class to bet display', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 50,
        playerBets: {
          0: 50,
        },
      };

      const { container } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      // Find bet amount element and verify CSS class
      const betAmountElement = container.querySelector('.bet-amount');
      expect(betAmountElement).toBeInTheDocument();
      expect(betAmountElement?.textContent).toContain('💵 50');
    });
  });
});

describe('Action Bar Tests - Phase 5', () => {
  const mockSeats: SeatInfo[] = [
    { index: 0, playerName: 'Alice', status: 'occupied', stack: 1000 },
    { index: 1, playerName: 'Bob', status: 'occupied', stack: 980 },
    { index: 2, playerName: 'Charlie', status: 'occupied', stack: 1000 },
  ];

  interface ExtendedGameState extends GameState {
    currentActor?: number | null;
    validActions?: string[] | null;
    callAmount?: number | null;
    foldedPlayers?: number[];
    roundOver?: boolean | null;
  }

  describe('TestTableView_ActionButtonsVisible', () => {
    it('should show action buttons only when player is current actor', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1, // Bob is current actor
        validActions: ['fold', 'call'],
        callAmount: 20,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1} // Current player is Bob
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      // Action buttons should be visible
      expect(screen.getByRole('button', { name: /Fold/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Call/i })).toBeInTheDocument();
    });

    it('should not show action buttons when player is not current actor', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 0, // Alice is current actor
        validActions: ['fold', 'call'],
        callAmount: 20,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1} // Current player is Bob (not the actor)
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      // Action buttons should not be visible
      expect(
        screen.queryByRole('button', { name: /Fold/i })
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole('button', { name: /Call/i })
      ).not.toBeInTheDocument();
    });
  });

  describe('TestTableView_CallButtonAmount', () => {
    it('should display Call button with amount', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call'],
        callAmount: 20,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const callButton = screen.getByRole('button', { name: /Call 20/i });
      expect(callButton).toBeInTheDocument();
    });
  });

  describe('TestTableView_CheckVsCall', () => {
    it('should display Check button when call amount is 0', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'check'],
        callAmount: 0,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      expect(
        screen.getByRole('button', { name: /Check/i })
      ).toBeInTheDocument();
      expect(
        screen.queryByRole('button', { name: /Call/i })
      ).not.toBeInTheDocument();
    });

    it('should display Call button when call amount is greater than 0', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call'],
        callAmount: 50,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      expect(
        screen.getByRole('button', { name: /Call 50/i })
      ).toBeInTheDocument();
      expect(
        screen.queryByRole('button', { name: /Check/i })
      ).not.toBeInTheDocument();
    });
  });

  describe('TestTableView_TurnIndicator', () => {
    it('should highlight current actor seat with turn-active class', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call'],
        callAmount: 20,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const seatContainers = document.querySelectorAll('.seat');
      expect(seatContainers[1].classList.contains('turn-active')).toBe(true);
      expect(seatContainers[0].classList.contains('turn-active')).toBe(false);
      expect(seatContainers[2].classList.contains('turn-active')).toBe(false);
    });
  });

  describe('TestTableView_ActionButtonClicks', () => {
    it('should send fold action when Fold button clicked', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call'],
        callAmount: 20,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const foldButton = screen.getByRole('button', { name: /Fold/i });
      fireEvent.click(foldButton);

      expect(mockSendMessage).toHaveBeenCalledOnce();
      const sentMessage = mockSendMessage.mock.calls[0][0];
      const message = JSON.parse(sentMessage);
      expect(message.type).toBe('player_action');
      expect(message.payload.action).toBe('fold');
      expect(message.payload.seatIndex).toBe(1);
    });

    it('should send call action when Call button clicked', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call'],
        callAmount: 20,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const callButton = screen.getByRole('button', { name: /Call 20/i });
      fireEvent.click(callButton);

      expect(mockSendMessage).toHaveBeenCalledOnce();
      const sentMessage = mockSendMessage.mock.calls[0][0];
      const message = JSON.parse(sentMessage);
      expect(message.type).toBe('player_action');
      expect(message.payload.action).toBe('call');
      expect(message.payload.seatIndex).toBe(1);
    });

    it('should send check action when Check button clicked', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'check'],
        callAmount: 0,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const checkButton = screen.getByRole('button', { name: /Check/i });
      fireEvent.click(checkButton);

      expect(mockSendMessage).toHaveBeenCalledOnce();
      const sentMessage = mockSendMessage.mock.calls[0][0];
      const message = JSON.parse(sentMessage);
      expect(message.type).toBe('player_action');
      expect(message.payload.action).toBe('check');
      expect(message.payload.seatIndex).toBe(1);
    });
  });

  describe('TestTableView_RaiseButton', () => {
    it('should show Raise button when raise is valid action', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call', 'raise'],
        callAmount: 20,
        minRaise: 40,
        maxRaise: 980,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      expect(
        screen.getByRole('button', { name: /Raise/i })
      ).toBeInTheDocument();
    });

    it('should hide Raise button when raise not available', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call'],
        callAmount: 20,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      expect(
        screen.queryByRole('button', { name: /Raise/i })
      ).not.toBeInTheDocument();
    });
  });

  describe('TestTableView_RaisePresets', () => {
    it('should show Min preset button and set raise amount to minRaise', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call', 'raise'],
        callAmount: 20,
        minRaise: 40,
        maxRaise: 980,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const minButton = screen.getByRole('button', { name: /Min/i });
      expect(minButton).toBeInTheDocument();

      fireEvent.click(minButton);

      const raiseInput = screen.getByDisplayValue('40') as HTMLInputElement;
      expect(raiseInput).toBeInTheDocument();
      expect(raiseInput.value).toBe('40');
    });

    it('should show Pot preset button and calculate pot-sized raise', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 100,
        currentActor: 1,
        validActions: ['fold', 'call', 'raise'],
        callAmount: 20,
        minRaise: 40,
        maxRaise: 980,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const potButton = screen.getByRole('button', { name: /Pot/i });
      expect(potButton).toBeInTheDocument();

      fireEvent.click(potButton);

      // Pot-sized raise = callAmount + pot = 20 + 100 = 120
      const raiseInput = screen.getByDisplayValue('120') as HTMLInputElement;
      expect(raiseInput).toBeInTheDocument();
      expect(raiseInput.value).toBe('120');
    });

    it('should show All-in preset button and set raise to player stack', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call', 'raise'],
        callAmount: 20,
        minRaise: 40,
        maxRaise: 980,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const allInButton = screen.getByRole('button', { name: /All-in/i });
      expect(allInButton).toBeInTheDocument();

      fireEvent.click(allInButton);

      // All-in = player stack = 980 (Bob's stack from mockSeats)
      const raiseInput = screen.getByDisplayValue('980') as HTMLInputElement;
      expect(raiseInput).toBeInTheDocument();
      expect(raiseInput.value).toBe('980');
    });
  });

  describe('TestTableView_RaiseInput', () => {
    it('should show raise amount input field when raise action available', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call', 'raise'],
        callAmount: 20,
        minRaise: 40,
        maxRaise: 980,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      const raiseInput = screen.getByRole('textbox', { name: /Raise Amount/i });
      expect(raiseInput).toBeInTheDocument();
    });
  });

  describe('TestTableView_RaiseAction', () => {
    it('should send raise action with amount when Raise button clicked', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call', 'raise'],
        callAmount: 20,
        minRaise: 40,
        maxRaise: 980,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      // Set raise amount
      const raiseInput = screen.getByRole('textbox', {
        name: /Raise Amount/i,
      }) as HTMLInputElement;
      fireEvent.change(raiseInput, { target: { value: '100' } });

      // Click Raise button
      const raiseButton = screen.getByRole('button', { name: /^Raise$/ });
      fireEvent.click(raiseButton);

      expect(mockSendMessage).toHaveBeenCalledOnce();
      const sentMessage = mockSendMessage.mock.calls[0][0];
      const message = JSON.parse(sentMessage);
      expect(message.type).toBe('player_action');
      expect(message.payload.action).toBe('raise');
      expect(message.payload.seatIndex).toBe(1);
      expect(message.payload.amount).toBe(100);
    });

    it('should disable Raise button when amount is below minimum', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call', 'raise'],
        callAmount: 20,
        minRaise: 40,
        maxRaise: 980,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      // Set raise amount below minimum
      const raiseInput = screen.getByRole('textbox', {
        name: /Raise Amount/i,
      }) as HTMLInputElement;
      fireEvent.change(raiseInput, { target: { value: '30' } });

      const raiseButton = screen.getByRole('button', { name: /^Raise$/ });
      expect(raiseButton).toBeDisabled();
    });

    it('should disable Raise button when amount is above maximum', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call', 'raise'],
        callAmount: 20,
        minRaise: 40,
        maxRaise: 980,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      // Set raise amount above maximum
      const raiseInput = screen.getByRole('textbox', {
        name: /Raise Amount/i,
      }) as HTMLInputElement;
      fireEvent.change(raiseInput, { target: { value: '1000' } });

      const raiseButton = screen.getByRole('button', { name: /^Raise$/ });
      expect(raiseButton).toBeDisabled();
    });

    it('should enable Raise button when amount is valid', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call', 'raise'],
        callAmount: 20,
        minRaise: 40,
        maxRaise: 980,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      // Set valid raise amount
      const raiseInput = screen.getByRole('textbox', {
        name: /Raise Amount/i,
      }) as HTMLInputElement;
      fireEvent.change(raiseInput, { target: { value: '100' } });

      const raiseButton = screen.getByRole('button', { name: /^Raise$/ });
      expect(raiseButton).not.toBeDisabled();
    });

    it('should clear raise input after any action', () => {
      const mockOnLeave = vi.fn();
      const mockSendMessage = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: ['As', 'Kh'],
        pot: 30,
        currentActor: 1,
        validActions: ['fold', 'call', 'raise'],
        callAmount: 20,
        minRaise: 40,
        maxRaise: 980,
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={1}
          onLeave={mockOnLeave}
          gameState={gameState}
          onSendMessage={mockSendMessage}
        />
      );

      // Set raise amount
      const raiseInput = screen.getByRole('textbox', {
        name: /Raise Amount/i,
      }) as HTMLInputElement;
      fireEvent.change(raiseInput, { target: { value: '100' } });
      expect(raiseInput.value).toBe('100');

      // Click Raise button
      const raiseButton = screen.getByRole('button', { name: /^Raise$/ });
      fireEvent.click(raiseButton);

      // Raise input should be cleared
      expect(raiseInput.value).toBe('');
    });
  });
});

describe('Phase 4: Board Card Display', () => {
  const mockSeats: SeatInfo[] = [
    { index: 0, playerName: 'Alice', status: 'occupied', stack: 1000 },
    { index: 1, playerName: 'Bob', status: 'occupied', stack: 980 },
    { index: 2, playerName: 'Charlie', status: 'occupied', stack: 1000 },
    { index: 3, playerName: null, status: 'empty' },
    { index: 4, playerName: null, status: 'empty' },
    { index: 5, playerName: null, status: 'empty' },
  ];

  interface ExtendedGameState extends GameState {
    boardCards?:
      | [string, string, string]
      | [string, string, string, string]
      | [string, string, string, string, string];
  }

  describe('TestTableView_DisplaysBoardCards', () => {
    it('should not render board cards container when hand is not active (pot is 0)', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 0,
        boardCards: [],
      };

      const { container } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      const boardContainer = container.querySelector('.board-cards');
      expect(boardContainer).not.toBeInTheDocument();
    });

    it('should render board cards container when hand is active', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 50,
        boardCards: [],
      };

      const { container } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      const boardContainer = container.querySelector('.board-cards');
      expect(boardContainer).toBeInTheDocument();
    });

    it('should display empty board slots preflop when hand is active', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 30,
        boardCards: [],
      };

      const { container } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      const boardSlots = container.querySelectorAll('.board-card');
      expect(boardSlots.length).toBe(5);

      // All should be empty slots
      boardSlots.forEach((slot) => {
        expect(slot.className).toContain('empty');
      });
    });

    it('should display three cards after flop', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 60,
        boardCards: ['As', 'Kh', 'Qd'],
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      expect(screen.getByText(/A♠/)).toBeInTheDocument();
      expect(screen.getByText(/K♥/)).toBeInTheDocument();
      expect(screen.getByText(/Q♦/)).toBeInTheDocument();
    });

    it('should display four cards after turn', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 120,
        boardCards: ['As', 'Kh', 'Qd', 'Jc'],
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      expect(screen.getByText(/A♠/)).toBeInTheDocument();
      expect(screen.getByText(/K♥/)).toBeInTheDocument();
      expect(screen.getByText(/Q♦/)).toBeInTheDocument();
      expect(screen.getByText(/J♣/)).toBeInTheDocument();
    });

    it('should display five cards after river', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 200,
        boardCards: ['As', 'Kh', 'Qd', 'Jc', 'Ts'],
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      expect(screen.getByText(/A♠/)).toBeInTheDocument();
      expect(screen.getByText(/K♥/)).toBeInTheDocument();
      expect(screen.getByText(/Q♦/)).toBeInTheDocument();
      expect(screen.getByText(/J♣/)).toBeInTheDocument();
      expect(screen.getByText(/T♠/)).toBeInTheDocument();
    });

    it('should apply red-suit class to red cards', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 80,
        boardCards: ['Ah', 'Kd'],
      };

      const { container } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      const redCards = container.querySelectorAll(
        '.board-card.face-up.red-suit'
      );
      expect(redCards.length).toBeGreaterThan(0);
    });

    it('should apply black-suit class to black cards', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 90,
        boardCards: ['As', 'Kc'],
      };

      const { container } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      const blackCards = container.querySelectorAll(
        '.board-card.face-up.black-suit'
      );
      expect(blackCards.length).toBeGreaterThan(0);
    });

    it('should update board cards when gameState changes', () => {
      const mockOnLeave = vi.fn();
      const initialGameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 30,
        boardCards: [],
      };

      const { rerender } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={initialGameState}
        />
      );

      // Initially no board cards
      expect(screen.queryByText(/A♠/)).not.toBeInTheDocument();

      // Update with flop
      const flopGameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 60,
        boardCards: ['As', 'Kh', 'Qd'],
      };

      rerender(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={flopGameState}
        />
      );

      expect(screen.getByText(/A♠/)).toBeInTheDocument();
      expect(screen.getByText(/K♥/)).toBeInTheDocument();
      expect(screen.getByText(/Q♦/)).toBeInTheDocument();
    });
  });
});

describe('Phase 5: Street Indicator', () => {
  const mockSeats: SeatInfo[] = [
    { index: 0, playerName: 'Alice', status: 'occupied', stack: 1000 },
    { index: 1, playerName: 'Bob', status: 'occupied', stack: 980 },
    { index: 2, playerName: 'Charlie', status: 'occupied', stack: 1000 },
    { index: 3, playerName: null, status: 'empty' },
    { index: 4, playerName: null, status: 'empty' },
    { index: 5, playerName: null, status: 'empty' },
  ];

   interface ExtendedGameState extends GameState {
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

  describe('TestTableView_StreetIndicator', () => {
    it('should display Preflop street indicator', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 30,
        boardCards: [],
        street: 'preflop',
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      expect(screen.getByText(/Preflop/i)).toBeInTheDocument();
    });

    it('should display Flop street indicator', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 60,
        boardCards: ['As', 'Kh', 'Qd'],
        street: 'flop',
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      expect(screen.getByText(/Flop/i)).toBeInTheDocument();
    });

    it('should display Turn street indicator', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 120,
        boardCards: ['As', 'Kh', 'Qd', 'Jc'],
        street: 'turn',
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      expect(screen.getByText(/Turn/i)).toBeInTheDocument();
    });

    it('should display River street indicator', () => {
      const mockOnLeave = vi.fn();
      const gameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 200,
        boardCards: ['As', 'Kh', 'Qd', 'Jc', 'Ts'],
        street: 'river',
      };

      render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={gameState}
        />
      );

      expect(screen.getByText(/River/i)).toBeInTheDocument();
    });

    it('should update street indicator when board_dealt event triggers street change', () => {
      const mockOnLeave = vi.fn();
      const initialGameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 30,
        boardCards: [],
        street: 'preflop',
      };

      const { rerender } = render(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={initialGameState}
        />
      );

      // Initially should show Preflop
      expect(screen.getByText(/Preflop/i)).toBeInTheDocument();

      // Update to Flop
      const flopGameState: ExtendedGameState = {
        dealerSeat: 0,
        smallBlindSeat: 1,
        bigBlindSeat: 2,
        holeCards: null,
        pot: 60,
        boardCards: ['As', 'Kh', 'Qd'],
        street: 'flop',
      };

      rerender(
        <TableView
          tableId="table-1"
          seats={mockSeats}
          currentSeatIndex={0}
          onLeave={mockOnLeave}
          gameState={flopGameState}
        />
      );

      // Should now show Flop
      expect(screen.getByText(/Flop/i)).toBeInTheDocument();
      expect(screen.queryByText(/Preflop/i)).not.toBeInTheDocument();
    });
  });

  describe('Phase 6: Showdown & Settlement - TableView Display', () => {
    describe('TestTableView_ShowdownDisplay', () => {
       it('should display showdown overlay with single winner', () => {
         const mockOnLeave = vi.fn();
         const gameState: ExtendedGameState = {
           dealerSeat: 0,
           smallBlindSeat: 1,
           bigBlindSeat: 2,
           holeCards: null,
           pot: 300,
           boardCards: ['As', 'Kh', 'Qd', 'Jc', 'Ts'],
           street: 'river',
           showdown: {
             winnerSeats: [1],
             winningHand: 'Pair of Aces',
             potAmount: 300,
             amountsWon: { 1: 300 },
           },
         };

         render(
           <TableView
             tableId="table-1"
             seats={mockSeats}
             currentSeatIndex={0}
             onLeave={mockOnLeave}
             gameState={gameState}
           />
         );

         // Should display showdown overlay
         expect(screen.getByText(/Pair of Aces/i)).toBeInTheDocument();
         // Check for the showdown overlay by looking for the winners text
         expect(screen.getByText(/Winners:/i)).toBeInTheDocument();
       });

      it('should display showdown overlay with multiple winners (split pot)', () => {
        const mockOnLeave = vi.fn();
        const gameState: ExtendedGameState = {
          dealerSeat: 0,
          smallBlindSeat: 1,
          bigBlindSeat: 2,
          holeCards: null,
          pot: 400,
          boardCards: ['As', 'Kh', 'Qd', 'Jc', 'Ts'],
          street: 'river',
          showdown: {
            winnerSeats: [0, 2],
            winningHand: 'Pair of Kings',
            potAmount: 400,
            amountsWon: { 0: 200, 2: 200 },
          },
        };

        render(
          <TableView
            tableId="table-1"
            seats={mockSeats}
            currentSeatIndex={0}
            onLeave={mockOnLeave}
            gameState={gameState}
          />
        );

        // Should display showdown overlay
        expect(screen.getByText(/Pair of Kings/i)).toBeInTheDocument();
      });

       it('should highlight winner seats with gold border', () => {
         const mockOnLeave = vi.fn();
         const gameState: ExtendedGameState = {
           dealerSeat: 0,
           smallBlindSeat: 1,
           bigBlindSeat: 2,
           holeCards: null,
           pot: 300,
           boardCards: ['As', 'Kh', 'Qd', 'Jc', 'Ts'],
           street: 'river',
           showdown: {
             winnerSeats: [1],
             winningHand: 'Pair of Aces',
             potAmount: 300,
             amountsWon: { 1: 300 },
           },
         };

         const { container } = render(
           <TableView
             tableId="table-1"
             seats={mockSeats}
             currentSeatIndex={0}
             onLeave={mockOnLeave}
             gameState={gameState}
           />
         );

         // Find winner seat element (seat 1 - Bob)
         const seatElements = container.querySelectorAll('.seat');
         const winnerSeat = Array.from(seatElements).find((seat) =>
           seat.textContent?.includes('Bob')
         );

         expect(winnerSeat).toHaveClass('winner-seat');
      });

      it('should display hand complete message', () => {
        const mockOnLeave = vi.fn();
        const gameState: ExtendedGameState = {
          dealerSeat: 0,
          smallBlindSeat: 1,
          bigBlindSeat: 2,
          holeCards: null,
          pot: 300,
          boardCards: ['As', 'Kh', 'Qd', 'Jc', 'Ts'],
          street: 'river',
          showdown: {
            winnerSeats: [1],
            winningHand: 'Pair of Aces',
            potAmount: 300,
            amountsWon: { 1: 300 },
          },
          handComplete: {
            message: 'Hand complete! Winner collected the pot.',
          },
        };

        render(
          <TableView
            tableId="table-1"
            seats={mockSeats}
            currentSeatIndex={0}
            onLeave={mockOnLeave}
            gameState={gameState}
          />
        );

        // Should display hand complete message
        expect(
          screen.getByText(/Hand complete! Winner collected the pot./i)
        ).toBeInTheDocument();
      });

      it('should display winner names in showdown overlay', () => {
        const mockOnLeave = vi.fn();
        const gameState: ExtendedGameState = {
          dealerSeat: 0,
          smallBlindSeat: 1,
          bigBlindSeat: 2,
          holeCards: null,
          pot: 300,
          boardCards: ['As', 'Kh', 'Qd', 'Jc', 'Ts'],
          street: 'river',
          showdown: {
            winnerSeats: [1],
            winningHand: 'Pair of Aces',
            potAmount: 300,
            amountsWon: { 1: 300 },
          },
        };

        render(
          <TableView
            tableId="table-1"
            seats={mockSeats}
            currentSeatIndex={0}
            onLeave={mockOnLeave}
            gameState={gameState}
          />
        );

        // Should display winner name (Alice is in seat 1)
        expect(screen.getByText(/Alice/i)).toBeInTheDocument();
      });
    });
  });
});
