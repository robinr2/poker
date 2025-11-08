import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

import '@testing-library/jest-dom';
import { TableView } from './TableView';

describe('TableView', () => {
  const mockSeats = [
    { index: 0, playerName: null, status: 'empty' },
    { index: 1, playerName: 'Alice', status: 'occupied' },
    { index: 2, playerName: null, status: 'empty' },
    { index: 3, playerName: 'Bob', status: 'occupied' },
    { index: 4, playerName: null, status: 'empty' },
    { index: 5, playerName: 'Charlie', status: 'occupied' },
  ];

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
});
