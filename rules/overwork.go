package rules

import "github.com/imdehydrated/rootbuddy/game"

func hasMarquiseSawmill(c game.Clearing) bool {
	for _, building := range c.Buildings {
		if building.Faction == game.Marquise && building.Type == game.Sawmill {
			return true
		}
	}
	return false
}

func ValidOverworkActions(state game.GameState) []game.Action {
	actions := []game.Action{}
	if state.FactionTurn != game.Marquise {
		return actions
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepDaylightActions {
		return actions
	}

	if state.CurrentPhase != game.Daylight {
		return actions
	}

	if state.TurnProgress.ActionsUsed >= 3+state.TurnProgress.BonusActions {
		return actions
	}

	for _, clearing := range state.Map.Clearings {
		if !hasMarquiseSawmill(clearing) {
			continue
		}

		for _, card := range state.Marquise.CardsInHand {
			if card.Suit != clearing.Suit && card.Suit != game.Bird {
				continue
			}
			actions = append(actions, game.Action{
				Type: game.ActionOverwork,
				Overwork: &game.OverworkAction{
					Faction:    game.Marquise,
					ClearingID: clearing.ID,
					CardID:     card.ID,
				},
			})
		}
	}

	return actions
}
