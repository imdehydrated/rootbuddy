package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidAidActions(state game.GameState) []game.Action {
	if len(vagabondReadyItemIndexes(state)) == 0 || len(state.Vagabond.CardsInHand) == 0 {
		return nil
	}

	clearing, ok := vagabondCurrentClearing(state)
	if !ok {
		return nil
	}

	actions := []game.Action{}
	for _, targetFaction := range vagabondFactionsInClearing(clearing) {
		if vagabondRelationshipLevel(state, targetFaction) == game.RelHostile {
			continue
		}

		for _, card := range state.Vagabond.CardsInHand {
			actions = append(actions, game.Action{
				Type: game.ActionAid,
				Aid: &game.AidAction{
					Faction:       game.Vagabond,
					TargetFaction: targetFaction,
					ClearingID:    clearing.ID,
					CardID:        card.ID,
				},
			})
		}
	}

	return actions
}
