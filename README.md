# RootBuddy

RootBuddy is a Go backend and React frontend for tracking, validating, and playing through game states for the board game Root.

The project is built around deterministic state transitions, explicit rule logic, and a clean split between domain modeling, engine behavior, transport, and UI. It supports local play tools, assist tooling for observed turns, and lobby-backed multiplayer with server authority.

## What It Does

- Models Root game state across the base factions, map, cards, items, quests, effects, and turn progress
- Generates legal actions and applies chosen actions for Marquise, Eyrie, Woodland Alliance, and Vagabond play
- Handles setup flow, battle resolution, crafting, scoring, dominance, and faction-specific turn logic
- Exposes the engine through a JSON API and a browser frontend
- Supports lobby-backed multiplayer with websocket updates, reconnects, battle prompts, and an action log

## Current Scope

The current implementation includes:

- Autumn map support
- Base deck, quest deck, item supply, and dominance card support
- Rule generation and action application for the main base-game faction workflows
- Local setup, load, and play flows through the API and frontend
- Assist tooling for observed turns and public-action tracking
- Multiplayer lobbies with join codes, readiness, faction claims, reconnect handling, and live game updates
- Server-authoritative multiplayer validation for turns, battle responses, and hidden-information checks

Planned next:

- Continued rule coverage and rules polish
- More frontend polish and quality-of-life improvements
- Broader multiplayer and persistence refinement

## Architecture

The project is organized into a few clear layers:

- `game/`: shared domain types
- `mapdata/`: static board data
- `carddata/`: static card catalog data
- `rules/`: pure rule-generation functions
- `engine/`: action sequencing, state transitions, and battle resolution
- `server/`: HTTP handlers and request/response DTOs
- `frontend/`: React + TypeScript client for setup, board state, action flow, assist mode, and multiplayer

The rules layer is intentionally pure: given a state, it returns legal actions without mutating anything. The engine layer is responsible for applying chosen actions and moving the game state forward.

## Running Locally

Requirements:

- Go 1.26+
- Node.js 20+

Start the server:

```bash
go run .
```

The server listens on `http://localhost:8080`.

Start the frontend:

```bash
cd frontend
npm install
npm run dev
```

The frontend runs on `http://localhost:5173`.

Build the frontend:

```bash
cd frontend
npm run build
```

## API

Current endpoints:

- `GET /health`
- `GET /api/health`
- `POST /actions/valid`
- `POST /api/actions/valid`
- `POST /actions/apply`
- `POST /api/actions/apply`
- `POST /battles/resolve`
- `POST /api/battles/resolve`
- `POST /battles/context`
- `POST /api/battles/context`
- `POST /battles/open`
- `POST /api/battles/open`
- `GET /battles/session`
- `GET /api/battles/session`
- `POST /battles/respond`
- `POST /api/battles/respond`
- `POST /game/setup`
- `POST /api/game/setup`
- `POST /game/load`
- `POST /api/game/load`
- `GET /game/log`
- `GET /api/game/log`
- `POST /lobby/create`
- `POST /api/lobby/create`
- `POST /lobby/join`
- `POST /api/lobby/join`
- `GET /lobby/state`
- `GET /api/lobby/state`
- `POST /lobby/claim-faction`
- `POST /api/lobby/claim-faction`
- `POST /lobby/ready`
- `POST /api/lobby/ready`
- `POST /lobby/start`
- `POST /api/lobby/start`
- `POST /lobby/leave`
- `POST /api/lobby/leave`
- `GET /ws`
- `GET /api/ws`

Example usage:

- send a `GameState` to `/api/actions/valid` to get the legal next actions
- send a `GameState` plus one chosen `Action` to `/api/actions/apply` to get the next state
- use `/api/game/setup` to create a new game and `/api/game/load` to load a saved one
- use the `/api/lobby/*` endpoints plus `/api/ws` for multiplayer lobby and game sessions
- use the `/api/battles/*` endpoints for battle previews, multiplayer response flow, and battle resolution

## Testing

The codebase includes unit and integration-style tests for:

- domain modeling
- map and card data
- rules logic
- engine state transitions
- HTTP handlers

The frontend lives under `frontend/` as a separate React + TypeScript app.

Run the test suite with:

```bash
go test ./...
```

Run the frontend build/typecheck with:

```bash
cd frontend
npm run build
```
