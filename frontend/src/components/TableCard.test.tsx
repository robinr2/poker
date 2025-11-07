import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

import '@testing-library/jest-dom';
import { TableCard } from './TableCard';

interface TableInfo {
  id: string;
  name: string;
  seatsOccupied: number;
  maxSeats: number;
}

describe('TableCard', () => {
  const mockTable: TableInfo = {
    id: 'table-1',
    name: 'Table 1',
    seatsOccupied: 3,
    maxSeats: 6,
  };

  describe('TestTableCardDisplaysInfo', () => {
    it('should display table name', () => {
      const mockOnJoin = vi.fn();
      render(<TableCard table={mockTable} onJoin={mockOnJoin} />);

      expect(screen.getByText('Table 1')).toBeInTheDocument();
    });

    it('should display seat count in X/Y format', () => {
      const mockOnJoin = vi.fn();
      render(<TableCard table={mockTable} onJoin={mockOnJoin} />);

      expect(screen.getByText('3/6')).toBeInTheDocument();
    });

    it('should display seats label', () => {
      const mockOnJoin = vi.fn();
      render(<TableCard table={mockTable} onJoin={mockOnJoin} />);

      expect(screen.getByText(/seats/i)).toBeInTheDocument();
    });
  });

  describe('TestTableCardJoinButton', () => {
    it('should render Join button', () => {
      const mockOnJoin = vi.fn();
      render(<TableCard table={mockTable} onJoin={mockOnJoin} />);

      const button = screen.getByRole('button', { name: /join/i });
      expect(button).toBeInTheDocument();
    });

    it('should call onJoin with table id when Join clicked', () => {
      const mockOnJoin = vi.fn();
      render(<TableCard table={mockTable} onJoin={mockOnJoin} />);

      const button = screen.getByRole('button', { name: /join/i });
      fireEvent.click(button);

      expect(mockOnJoin).toHaveBeenCalledTimes(1);
      expect(mockOnJoin).toHaveBeenCalledWith('table-1');
    });
  });

  describe('TestTableCardJoinButtonEnabled', () => {
    it('should enable Join button when seats available', () => {
      const mockOnJoin = vi.fn();
      const tableWithSeats: TableInfo = {
        id: 'table-2',
        name: 'Table 2',
        seatsOccupied: 0,
        maxSeats: 6,
      };
      render(<TableCard table={tableWithSeats} onJoin={mockOnJoin} />);

      const button = screen.getByRole('button', { name: /join/i });
      expect(button).not.toBeDisabled();
    });

    it('should enable Join button when one seat available', () => {
      const mockOnJoin = vi.fn();
      const tableWithSeats: TableInfo = {
        id: 'table-3',
        name: 'Table 3',
        seatsOccupied: 5,
        maxSeats: 6,
      };
      render(<TableCard table={tableWithSeats} onJoin={mockOnJoin} />);

      const button = screen.getByRole('button', { name: /join/i });
      expect(button).not.toBeDisabled();
    });
  });

  describe('TestTableCardJoinButtonDisabled', () => {
    it('should disable Join button when table full', () => {
      const mockOnJoin = vi.fn();
      const fullTable: TableInfo = {
        id: 'table-4',
        name: 'Table 4',
        seatsOccupied: 6,
        maxSeats: 6,
      };
      render(<TableCard table={fullTable} onJoin={mockOnJoin} />);

      const button = screen.getByRole('button', { name: /join/i });
      expect(button).toBeDisabled();
    });

    it('should not call onJoin when disabled button clicked', () => {
      const mockOnJoin = vi.fn();
      const fullTable: TableInfo = {
        id: 'table-5',
        name: 'Table 5',
        seatsOccupied: 6,
        maxSeats: 6,
      };
      render(<TableCard table={fullTable} onJoin={mockOnJoin} />);

      const button = screen.getByRole('button', { name: /join/i });
      fireEvent.click(button);

      expect(mockOnJoin).not.toHaveBeenCalled();
    });
  });
});
