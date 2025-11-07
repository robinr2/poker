import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import "@testing-library/jest-dom";
import { LobbyView } from "./LobbyView";
import { TableInfo } from "./TableCard";

describe("LobbyView", () => {
  const mockTables: TableInfo[] = [
    { id: "table-1", name: "Table 1", seatsOccupied: 2, maxSeats: 6 },
    { id: "table-2", name: "Table 2", seatsOccupied: 4, maxSeats: 6 },
    { id: "table-3", name: "Table 3", seatsOccupied: 0, maxSeats: 6 },
    { id: "table-4", name: "Table 4", seatsOccupied: 6, maxSeats: 6 },
  ];

  describe("rendering", () => {
    it("renders the lobby title", () => {
      render(<LobbyView tables={mockTables} onJoinTable={vi.fn()} />);
      expect(screen.getByText("Lobby")).toBeInTheDocument();
    });

    it("renders all four tables", () => {
      render(<LobbyView tables={mockTables} onJoinTable={vi.fn()} />);
      expect(screen.getByText("Table 1")).toBeInTheDocument();
      expect(screen.getByText("Table 2")).toBeInTheDocument();
      expect(screen.getByText("Table 3")).toBeInTheDocument();
      expect(screen.getByText("Table 4")).toBeInTheDocument();
    });

    it("renders tables in a grid container", () => {
      const { container } = render(
        <LobbyView tables={mockTables} onJoinTable={vi.fn()} />
      );
      const grid = container.querySelector(".lobby-grid");
      expect(grid).toBeInTheDocument();
      expect(grid?.children).toHaveLength(4);
    });

    it("renders empty lobby when no tables provided", () => {
      render(<LobbyView tables={[]} onJoinTable={vi.fn()} />);
      expect(screen.getByText("Lobby")).toBeInTheDocument();
      const grid = screen
        .getByText("Lobby")
        .parentElement?.querySelector(".lobby-grid");
      expect(grid?.children).toHaveLength(0);
    });
  });

  describe("table interaction", () => {
    it("passes onJoinTable callback to TableCard components", () => {
      const mockOnJoin = vi.fn();
      render(<LobbyView tables={mockTables} onJoinTable={mockOnJoin} />);

      // Click the Join button on Table 3 (which is empty and available)
      const joinButtons = screen.getAllByText("Join");
      const availableButton = joinButtons.find(
        (btn) => !btn.hasAttribute("disabled")
      );
      availableButton?.click();

      // Should have called the callback with a table ID
      expect(mockOnJoin).toHaveBeenCalledWith(expect.any(String));
    });
  });
});
