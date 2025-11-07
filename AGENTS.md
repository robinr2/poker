Identity & Rejoin
- Backend: Issue UUID; map token→name, optional seat/table; in-memory only.
- Frontend: One-time name prompt; store token; auto-attach; on load, reopen current table if seated.
Lobby (4 Fixed Tables, 6‑max)
- Backend: Preseed 4 identical tables; expose seat counts; push live updates.
- Frontend: Lobby lists 4 tables with seats used/open; “Join” disabled when full.
Seating & Waiting
- Backend: Immediate sit to first empty seat; enforce single seat per token; if join mid-hand, mark “waiting” (dealt next hand); leave seats frees immediately; seat clears on disconnect.
- Frontend: “Join Table” takes a seat instantly; show “Waiting for next hand” if seated mid-hand; Leave button returns to lobby.
Hand Start & Blinds
- Backend: When ≥2 active players and no hand running: set dealer/button, SB/BB (HU rule: button is SB), post blinds, shuffle, deal hole cards to active (not waiting) seats.
- Frontend: Dealer/button markers; blinds posted indicators; show hole cards to seat owner only.
Preflop (Call/Check/Fold Only)
- Backend: Turn order (UTG after BB; HU: BB acts first post-blinds); validate actions; maintain pot; close round when all matched or single player remains.
- Frontend: Action bar with only valid options; highlight actor; simple countdown ring.
Raises (Basic NL)
- Backend: Add min-raise computation and validation; track current bet per street; allow all-in amounts but initially reject actions that would require side pots (until “All-ins & Side Pots” is done).
- Frontend: Add Raise button with min/pot/all-in presets and slider; show “to call” amount; disable invalid amounts.
Postflop Streets (Flop/Turn/River)
- Backend: On betting round closure, reveal board cards (3/1/1); reset street bet state; rotate action properly; maintain pot.
- Frontend: Reveal board per street; street indicator; maintain action flow.
Showdown & Settlement
- Backend: Evaluate remaining live hands; distribute pot; update stacks; rotate dealer; promote waiting seats to active for next hand; if any stack reaches 0, kick the seat after settlement.
- Frontend: Reveal sequence; winner highlight; stack updates; brief “Next hand starting” banner.
Timers & Reconnect
- Backend: Server-authoritative per-turn deadline; auto-check/fold on expiry; reconnect snapshot (seat, street, pot, board, toAct, deadlines).
- Frontend: Countdown ring; auto-reconnect; request/apply snapshot on reload or desync.
All‑ins & Side Pots (Add When Base Is Stable)
- Backend: Permit all-ins without restriction; construct side pots by ascending all-in amounts; distribute each pot to eligible winners; handle partial raises not reopening action.
- Frontend: Show multiple pot totals; “All‑in” badge; aggregate winner display.
Bust‑Out Handling (No Rebuy)
- Backend: After settlement, if stack == 0 → remove from seat; seat becomes empty and broadcast update.
- Frontend: If you bust, show “You’re out. Rejoin from lobby to play again.”
Real‑Time Transport
- Backend: WebSocket for lobby seat counts and per-table events: snapshot, seat_update, state_change, action_request, action_result, board_update, pot_update, timer_tick.
- Frontend: Socket manager with auto-reconnect; event → store updates; snapshot on connect/reconnect.
Minimal UI
- Backend: Serve static SPA.
- Frontend: Two screens: Lobby (4 tables with Join/Full), Table (6 seats, names, stacks, dealer/blinds markers, board, pot, action bar, timer). No confirmations, light theme only.
Logging & Errors
- Backend: Structured logs; simple error codes (table_full, already_seated, mid_hand_wait, invalid_action, not_your_turn, action_timeout).
- Frontend: Map codes to concise toasts; retry affordances where safe.
Notes:
- Seats taken mid-hand do not receive cards until the next hand (poker-correct, keeps dealing simple).
- No queue/waitlist; Join is disabled when full; users wait in lobby until a seat frees.
- Equal starting stacks on every sit; no top-up, no rebuy; stacks diverge naturally across hands.
