# RootBuddy

RootBuddy is a Go backend project that models the rules and game state for the board game Root.

The goal is to provide a deterministic rules engine and API for analyzing Root game states, generating legal actions, applying chosen actions, and supporting future frontend or tooling work. The project is being built around explicit game-state transitions, testable rule logic, and a clean separation between domain modeling, engine behavior, and HTTP transport.

## What It Does

- Models core Root domain concepts such as clearings, factions, actions, buildings, cards, and turn progress
- Generates legal actions for the Marquise de Cat
- Applies chosen actions to produce the next game state
- Resolves battles from user-entered die rolls
- Exposes the engine through a small JSON API

## Current Scope

The backend currently includes:

- Autumn map data
- Base deck catalog
- Rule generation for ruling, movement, battle, build, recruit, overwork, and crafting
- Engine orchestration for step-based action generation
- Action application for recruit, movement, build, craft, overwork, and resolved battle outcomes
- HTTP endpoints for health checks, valid actions, action application, and battle resolution

Planned next:

- Broader turn sequencing and action application coverage
- API refinement and frontend integration
- Additional factions and full rulebook coverage

## Architecture

The project is organized into a few clear layers:

- `game/`: shared domain types
- `mapdata/`: static board data
- `carddata/`: static card catalog data
- `rules/`: pure rule-generation functions
- `engine/`: action sequencing, state transitions, and battle resolution
- `server/`: HTTP handlers and request/response DTOs

The rules layer is intentionally pure: given a state, it returns legal actions without mutating anything. The engine layer is responsible for applying chosen actions and moving the game state forward.

## Running Locally

Requirements:

- Go 1.26+

Start the server:

```bash
go run .
```

The server listens on `http://localhost:8080`.

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

Example usage:

- send a `GameState` to `/api/actions/valid` to get the legal next actions
- send a `GameState` plus one chosen `Action` to `/api/actions/apply` to get the next state
- send a battle action plus die rolls to `/api/battles/resolve` to get a resolved battle action that can then be applied

## Testing

The codebase includes unit and integration-style tests for:

- domain modeling
- map and card data
- rules logic
- engine state transitions
- HTTP handlers

Run the test suite with:

```bash
go test ./...
```
