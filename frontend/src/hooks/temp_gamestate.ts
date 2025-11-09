interface GameState {
  dealerSeat: number | null;
  smallBlindSeat: number | null;
  bigBlindSeat: number | null;
  holeCards: [string, string] | null;
  pot: number;
  currentActor: number | null;
  validActions: string[] | null;
  callAmount: number | null;
  foldedPlayers: number[];
  roundOver: boolean | null;
  handInProgress?: boolean;
}
