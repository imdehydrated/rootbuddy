package rules

import "github.com/imdehydrated/rootbuddy/game"

func vagabondDrawCount(state game.GameState) int {
	if state.Vagabond.InForest {
		return 0
	}

	coinCount := len(vagabondItemIndexes(state, game.ItemCoin, game.ItemReady, game.ItemExhausted))
	if coinCount > 3 {
		return 3
	}

	return coinCount
}

func ValidVagabondEveningActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Vagabond || state.CurrentPhase != game.Evening {
		return nil
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepEvening {
		return nil
	}

	return []game.Action{
		{
			Type: game.ActionEveningDraw,
			EveningDraw: &game.EveningDrawAction{
				Faction: game.Vagabond,
				Count:   vagabondDrawCount(state),
			},
		},
	}
}
