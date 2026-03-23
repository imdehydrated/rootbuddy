package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidOverworkActions(state game.GameState) []game.Action {
	return ValidMarquiseOverworkActions(state)
}

func ValidMarquiseOverworkActions(state game.GameState) []game.Action {
	actions := []game.Action{}
	if !marquiseIsDaylightActionStep(state) || marquiseActionLimitReached(state) {
		return actions
	}

	for _, clearing := range state.Map.Clearings {
		if !marquiseHasSawmill(clearing) {
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
