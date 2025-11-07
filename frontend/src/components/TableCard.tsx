export interface TableInfo {
  id: string;
  name: string;
  seatsOccupied: number;
  maxSeats: number;
}

interface TableCardProps {
  table: TableInfo;
  onJoin: (tableId: string) => void;
}

export function TableCard({ table, onJoin }: TableCardProps) {
  const isFull = table.seatsOccupied >= table.maxSeats;

  const handleJoinClick = () => {
    if (!isFull) {
      onJoin(table.id);
    }
  };

  return (
    <div className="table-card">
      <h3>{table.name}</h3>
      <p className="seat-count">
        {table.seatsOccupied}/{table.maxSeats}
      </p>
      <p className="seats-label">seats</p>
      <button
        onClick={handleJoinClick}
        disabled={isFull}
        className="join-button"
      >
        Join
      </button>
    </div>
  );
}
