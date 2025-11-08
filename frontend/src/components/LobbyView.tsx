import type { TableInfo } from './TableCard';
import { TableCard } from './TableCard';
import '../styles/LobbyView.css';

interface LobbyViewProps {
  tables: TableInfo[];
  onJoinTable: (tableId: string) => void;
}

export function LobbyView({ tables, onJoinTable }: LobbyViewProps) {
  return (
    <div className="lobby-view">
      <h1>Lobby</h1>
      <div className="lobby-grid">
        {tables.map((table) => (
          <TableCard key={table.id} table={table} onJoin={onJoinTable} />
        ))}
      </div>
    </div>
  );
}
