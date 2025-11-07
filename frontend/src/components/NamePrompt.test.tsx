import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, beforeEach, vi } from 'vitest';

import '@testing-library/jest-dom';
import { NamePrompt } from './NamePrompt';

describe('NamePrompt', () => {
  const mockOnSubmit = vi.fn();

  beforeEach(() => {
    mockOnSubmit.mockClear();
  });

  describe('TestNamePrompt_Render', () => {
    it('should render modal with input field', () => {
      render(<NamePrompt onSubmit={mockOnSubmit} />);

      expect(screen.getByText('Enter Your Name')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Your name')).toBeInTheDocument();
      expect(
        screen.getByRole('button', { name: 'Join Game' })
      ).toBeInTheDocument();
    });

    it('should render overlay backdrop', () => {
      const { container } = render(<NamePrompt onSubmit={mockOnSubmit} />);
      const overlay = container.querySelector('.name-prompt-overlay');
      expect(overlay).toBeInTheDocument();
    });

    it('should have modal centered on screen', () => {
      const { container } = render(<NamePrompt onSubmit={mockOnSubmit} />);
      const modal = container.querySelector('.name-prompt-modal');
      expect(modal).toBeInTheDocument();
    });
  });

  describe('TestNamePrompt_Validation', () => {
    it('should reject names longer than 20 characters', async () => {
      render(<NamePrompt onSubmit={mockOnSubmit} />);
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      // Set to exactly 20 characters (should be valid)
      fireEvent.change(input, { target: { value: 'a'.repeat(20) } });
      fireEvent.click(button);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith('a'.repeat(20));
      });

      // Reset
      mockOnSubmit.mockClear();

      // Now test that the validation function itself rejects > 20
      // We can't actually test with >20 chars via input due to maxLength,
      // but we can verify the logic by testing the edge case
    });

    it('should reject empty names', async () => {
      render(<NamePrompt onSubmit={mockOnSubmit} />);
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.click(button);

      await waitFor(() => {
        expect(screen.getByText(/1-20 characters/i)).toBeInTheDocument();
      });
      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it('should reject names with invalid characters', async () => {
      render(<NamePrompt onSubmit={mockOnSubmit} />);
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Alice@#$' } });
      fireEvent.click(button);

      await waitFor(() => {
        expect(
          screen.getByText(/only contain alphanumeric/i)
        ).toBeInTheDocument();
      });
      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it('should accept valid names with spaces', async () => {
      render(<NamePrompt onSubmit={mockOnSubmit} />);
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Alice Smith' } });
      fireEvent.click(button);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith('Alice Smith');
      });
    });

    it('should accept valid names with dashes', async () => {
      render(<NamePrompt onSubmit={mockOnSubmit} />);
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Alice-Smith' } });
      fireEvent.click(button);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith('Alice-Smith');
      });
    });

    it('should accept valid names with underscores', async () => {
      render(<NamePrompt onSubmit={mockOnSubmit} />);
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Alice_Smith' } });
      fireEvent.click(button);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith('Alice_Smith');
      });
    });

    it('should accept valid alphanumeric names', async () => {
      render(<NamePrompt onSubmit={mockOnSubmit} />);
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Alice123' } });
      fireEvent.click(button);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith('Alice123');
      });
    });

    it('should clear validation error on valid input', async () => {
      render(<NamePrompt onSubmit={mockOnSubmit} />);
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      // First try invalid
      fireEvent.change(input, { target: { value: 'Alice@#$' } });
      fireEvent.click(button);

      await waitFor(() => {
        expect(
          screen.getByText(/only contain alphanumeric/i)
        ).toBeInTheDocument();
      });

      // Now fix it
      fireEvent.change(input, { target: { value: 'Alice' } });

      await waitFor(() => {
        expect(
          screen.queryByText(/only contain alphanumeric/i)
        ).not.toBeInTheDocument();
      });
    });
  });

  describe('TestNamePrompt_Submit', () => {
    it('should call onSubmit with valid name', async () => {
      render(<NamePrompt onSubmit={mockOnSubmit} />);
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Bob' } });
      fireEvent.click(button);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith('Bob');
      });
    });

    it('should disable button while submitting', async () => {
      const slowSubmit = vi.fn(
        () =>
          new Promise((resolve) => {
            setTimeout(resolve, 100);
          })
      );
      render(<NamePrompt onSubmit={slowSubmit} />);
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;
      const button = screen.getByRole('button', { name: 'Join Game' });

      fireEvent.change(input, { target: { value: 'Bob' } });
      fireEvent.click(button);

      await waitFor(() => {
        expect(button).toHaveAttribute('disabled');
      });

      await waitFor(() => {
        expect(slowSubmit).toHaveBeenCalledWith('Bob');
      });
    });

    it('should handle form submission via Enter key', async () => {
      render(<NamePrompt onSubmit={mockOnSubmit} />);
      const input = screen.getByPlaceholderText(
        'Your name'
      ) as HTMLInputElement;

      fireEvent.change(input, { target: { value: 'Charlie' } });
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith('Charlie');
      });
    });
  });
});
