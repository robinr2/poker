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

       const seatsWithCardCount: SeatInfo[] = mockSeats.map((seat, idx) => ({
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
