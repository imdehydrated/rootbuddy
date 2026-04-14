# RootBuddy

RootBuddy is a Go backend and React frontend for tracking, validating, and playing through Root game states.

The project is built around deterministic state transitions, explicit rule logic, and a clean split between domain modeling, engine behavior, transport, and UI. It supports local play tools, assist workflows, and lobby-backed multiplayer with server authority and WebSocket updates.

## What It Does

- Models Root game state across the base factions, map, cards, items, quests, effects, and turn progress
- Generates legal actions and applies chosen actions for Marquise, Eyrie, Woodland Alliance, and Vagabond play
- Handles staged setup, battle context, multiplayer battle prompts, battle resolution, crafting, scoring, and dominance
- Exposes the engine through a JSON API and a browser frontend
- Supports lobby-backed multiplayer with join codes, faction claims, readiness, reconnect handling, action logs, and live redacted game updates

## Current Scope

The current implementation includes:

- Autumn map support
- Base deck, quest deck, item supply, and dominance card support
- Rule generation and action application for the main base-game faction workflows
- Local setup, load, play, and assist-mode flows through the API and frontend
- Multiplayer lobbies with WebSocket updates, reconnect handling, battle prompt coordination, and server-authoritative state
- A board-first frontend with setup, lobby, gameplay, assist, HUD, and correction-mode surfaces

Current focus:

- UI polish
- board interaction reliability
- responsive/layout work
- ongoing rules-compliance auditing

## Architecture

The project is organized into a few clear layers:

- `game/`: shared domain types
- `mapdata/`: static board data
- `carddata/`: static card and quest data
- `rules/`: pure legal-action generation
- `engine/`: deterministic state transitions and battle resolution
- `server/`: HTTP, multiplayer authority, persistence, redaction, battle sessions, and WebSocket fanout
- `frontend/`: React + TypeScript client for setup, board play, assist mode, lobby flow, and multiplayer UI

The rules layer is intentionally pure: given a state, it returns legal actions without mutating anything. The engine applies chosen actions and produces the next state. The server wraps that deterministic core with persistence, player perspective, multiplayer coordination, and transport.

## Running Locally

Requirements:

- Go 1.26+
- Node.js 20+

Start the backend:

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

## API

Current endpoints:

- `GET /health`
- `GET /api/health`
- `POST /actions/valid`
- `POST /api/actions/valid`
- `POST /actions/apply`
- `POST /api/actions/apply`
- `POST /battles/context`
- `POST /api/battles/context`
- `POST /battles/open`
- `POST /api/battles/open`
- `POST /battles/resolve`
- `POST /api/battles/resolve`
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

Typical flow:

- send a `GameState` to `/api/actions/valid` to get legal next actions
- send a `GameState` plus one chosen `Action` to `/api/actions/apply` to get the next state
- use `/api/battles/*` for battle previews, multiplayer prompt flow, and battle resolution
- use `/api/lobby/*` plus `/api/ws` for multiplayer lobby and live game updates

## Testing

Backend:

```bash
go test ./...
```

Frontend:

```bash
cd frontend
npm run test
npm run build
```
