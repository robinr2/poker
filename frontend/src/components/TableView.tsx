interface SeatInfo {
  index: number;
  playerName: string | null;
  status: string;
}

interface TableViewProps {
  tableId: string;
  seats: SeatInfo[];
  currentSeatIndex: number | null;
  onLeave: () => void;
}

export function TableView({
  tableId,
  seats,
  currentSeatIndex,
  onLeave,
}: TableViewProps) {
  return (
    <div className="table-view">
      <h1>Table: {tableId}</h1>
      <div className="seats-grid">
        {seats.map((seat) => (
          <div
            key={seat.index}
            className={`seat ${seat.index === currentSeatIndex ? 'own-seat' : ''}`}
          >
            <p className="seat-number">Seat {seat.index}</p>
            <p className="seat-player">
              {seat.playerName || 'Empty'}
            </p>
          </div>
        ))}
      </div>
      <button onClick={onLeave} className="leave-button">
        Leave Table
      </button>
    </div>
  );
}
