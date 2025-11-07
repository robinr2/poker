# Poker Features Implementation Structure

## Overview
This document organizes the features from `features.md` into a logical implementation order with dependencies mapped. Each feature group will be implemented as a separate plan with multiple phases.

## Implementation Order & Dependencies

### 1. Identity & Rejoin (FOUNDATION - No dependencies)
**Priority:** Critical - Required for all other features
**Complexity:** Low
**Estimated Phases:** 3

- Phase 1.1: Backend token generation and session management
- Phase 1.2: Frontend name prompt and token storage
- Phase 1.3: Rejoin logic and session restoration

**Dependencies:** None
**Enables:** All other features

---

### 2. Lobby System (Depends on: Identity)
**Priority:** Critical - Entry point for game
**Complexity:** Medium
**Estimated Phases:** 4

- Phase 2.1: Backend table structure and preseeded tables
- Phase 2.2: Lobby state management and WebSocket events
- Phase 2.3: Frontend lobby UI (4 tables display)
- Phase 2.4: Real-time seat count updates

**Dependencies:** Identity & Rejoin
**Enables:** Seating & Waiting

---

### 3. Seating & Waiting (Depends on: Identity, Lobby)
**Priority:** Critical - Required for gameplay
**Complexity:** Medium
**Estimated Phases:** 4

- Phase 3.1: Backend seat assignment logic
- Phase 3.2: Single seat enforcement and waiting status
- Phase 3.3: Leave/disconnect seat clearing
- Phase 3.4: Frontend join/leave UI and seat display

**Dependencies:** Identity & Rejoin, Lobby System
**Enables:** Hand Start & Blinds

---

### 4. Hand Start & Blinds (Depends on: Seating)
**Priority:** Critical - Core game mechanic
**Complexity:** High
**Estimated Phases:** 5

- Phase 4.1: Dealer/button rotation logic
- Phase 4.2: Blind posting (SB/BB, HU rules)
- Phase 4.3: Deck shuffling and card dealing
- Phase 4.4: Frontend dealer/blind markers
- Phase 4.5: Hole card display (owner-only visibility)

**Dependencies:** Seating & Waiting
**Enables:** Preflop Actions

---

### 5. Preflop Actions - Call/Check/Fold Only (Depends on: Hand Start)
**Priority:** Critical - Basic betting
**Complexity:** High
**Estimated Phases:** 5

- Phase 5.1: Turn order logic (UTG, HU rules)
- Phase 5.2: Action validation (call/check/fold)
- Phase 5.3: Pot management and bet tracking
- Phase 5.4: Betting round closure logic
- Phase 5.5: Frontend action bar and turn indicator

**Dependencies:** Hand Start & Blinds
**Enables:** Raises, Postflop Streets

---

### 6. Raises - Basic NL (Depends on: Preflop Actions)
**Priority:** High - Complete betting system
**Complexity:** Medium
**Estimated Phases:** 4

- Phase 6.1: Min-raise computation and validation
- Phase 6.2: Current bet tracking per street
- Phase 6.3: All-in detection (reject side pot scenarios initially)
- Phase 6.4: Frontend raise UI (slider, presets)

**Dependencies:** Preflop Actions
**Enables:** Complete betting, Postflop Streets

---

### 7. Postflop Streets - Flop/Turn/River (Depends on: Preflop, Raises)
**Priority:** Critical - Complete hand progression
**Complexity:** High
**Estimated Phases:** 5

- Phase 7.1: Board card revelation logic (3/1/1)
- Phase 7.2: Street bet state reset
- Phase 7.3: Action rotation per street
- Phase 7.4: Frontend board display
- Phase 7.5: Street indicator and flow

**Dependencies:** Preflop Actions, Raises
**Enables:** Showdown & Settlement

---

### 8. Showdown & Settlement (Depends on: Postflop Streets)
**Priority:** Critical - Hand completion
**Complexity:** High
**Estimated Phases:** 6

- Phase 8.1: Hand evaluation library integration
- Phase 8.2: Winner determination logic
- Phase 8.3: Pot distribution
- Phase 8.4: Stack updates and dealer rotation
- Phase 8.5: Waiting seat promotion
- Phase 8.6: Frontend winner reveal and stack updates

**Dependencies:** Postflop Streets
**Enables:** Complete game loop, Bust-Out Handling

---

### 9. Timers & Reconnect (Depends on: Basic Game Loop)
**Priority:** High - Production readiness
**Complexity:** Medium
**Estimated Phases:** 4

- Phase 9.1: Server-authoritative turn timer
- Phase 9.2: Auto-check/fold on expiry
- Phase 9.3: Reconnect snapshot generation
- Phase 9.4: Frontend countdown and reconnect UI

**Dependencies:** Showdown & Settlement (complete game loop)
**Enables:** Better UX, production stability

---

### 10. All-ins & Side Pots (Depends on: Complete Base Game)
**Priority:** Medium - Advanced feature
**Complexity:** Very High
**Estimated Phases:** 6

- Phase 10.1: Unrestricted all-in handling
- Phase 10.2: Side pot construction (ascending amounts)
- Phase 10.3: Eligibility tracking per pot
- Phase 10.4: Multi-pot distribution
- Phase 10.5: Partial raise action rules
- Phase 10.6: Frontend multiple pot display

**Dependencies:** Complete base game (through Showdown)
**Enables:** Full poker rules compliance

---

### 11. Bust-Out Handling (Depends on: Showdown & Settlement)
**Priority:** Medium - Game hygiene
**Complexity:** Low
**Estimated Phases:** 2

- Phase 11.1: Backend zero-stack seat clearing
- Phase 11.2: Frontend bust-out message and lobby redirect

**Dependencies:** Showdown & Settlement
**Enables:** Clean table management

---

### 12. Logging & Error Handling (Continuous - All Features)
**Priority:** High - Production quality
**Complexity:** Low-Medium
**Estimated Phases:** 3

- Phase 12.1: Backend error code definitions
- Phase 12.2: Structured logging patterns
- Phase 12.3: Frontend error mapping and toasts

**Dependencies:** Implemented alongside each feature
**Enables:** Debugging, user feedback

---

## Already Implemented (Bootstrap)
✅ Real-Time Transport (WebSocket infrastructure)
✅ Minimal UI (SPA serving, basic structure)
✅ Basic WebSocket connection management

## Implementation Strategy

### Critical Path (MVP - Playable Game)
1. Identity & Rejoin
2. Lobby System
3. Seating & Waiting
4. Hand Start & Blinds
5. Preflop Actions (Call/Check/Fold)
6. Raises (Basic NL)
7. Postflop Streets
8. Showdown & Settlement

**After MVP:** Timers, All-ins & Side Pots, Bust-Out, Enhanced Logging

### Total Estimated Phases: 48 phases across 12 features

### Execution Plan
1. Start with Identity & Rejoin (Feature 1)
2. Progress sequentially through dependency chain
3. Each feature gets its own plan document
4. Use subagents for implementation and review
5. Commit after each phase completion
6. Test full game loop after Showdown & Settlement

## Feature Complexity Matrix

| Feature | Backend | Frontend | Testing | Total |
|---------|---------|----------|---------|-------|
| Identity & Rejoin | Low | Low | Low | Low |
| Lobby System | Medium | Medium | Medium | Medium |
| Seating & Waiting | Medium | Low | Medium | Medium |
| Hand Start & Blinds | High | Medium | High | High |
| Preflop Actions | High | Medium | High | High |
| Raises | Medium | Medium | Medium | Medium |
| Postflop Streets | High | Medium | High | High |
| Showdown & Settlement | Very High | Medium | Very High | Very High |
| Timers & Reconnect | Medium | Low | Medium | Medium |
| All-ins & Side Pots | Very High | Medium | Very High | Very High |
| Bust-Out Handling | Low | Low | Low | Low |
| Logging & Errors | Medium | Low | Low | Low-Medium |

## Notes
- Each feature will follow strict TDD principles
- WebSocket protocol will be extended incrementally
- All features maintain in-memory state (no database)
- Focus on correctness over performance initially
- UI remains minimal throughout (light theme, no confirmations)
