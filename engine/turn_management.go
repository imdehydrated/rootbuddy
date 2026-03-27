package engine

import "github.com/imdehydrated/rootbuddy/game"

var defaultTurnOrder = []game.Faction{
	game.Marquise,
	game.Eyrie,
	game.Alliance,
	game.Vagabond,
}

func effectiveTurnOrder(state game.GameState) []game.Faction {
	if len(state.TurnOrder) > 0 {
		return state.TurnOrder
	}

	return defaultTurnOrder
}

func nextFactionInTurnOrder(state game.GameState) game.Faction {
	order := effectiveTurnOrder(state)
	if len(order) == 0 {
		return state.FactionTurn
	}

	for i, faction := range order {
		if faction == state.FactionTurn {
			return order[(i+1)%len(order)]
		}
	}

	return order[0]
}

func resetTurnProgress(state *game.GameState) {
	state.TurnProgress = game.TurnProgress{}
}

func beginNextFactionTurn(state *game.GameState) {
	order := effectiveTurnOrder(*state)
	if len(state.TurnOrder) == 0 {
		state.TurnOrder = append([]game.Faction(nil), order...)
	}

	state.FactionTurn = nextFactionInTurnOrder(*state)
	state.CurrentPhase = game.Birdsong
	state.CurrentStep = game.StepBirdsong
	resetTurnProgress(state)
	checkBirdsongDominanceWin(state)
}
