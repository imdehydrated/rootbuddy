package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidAidActions(state game.GameState) []game.Action {
	readyItemIndexes := vagabondReadyItemIndexes(state)
	if len(readyItemIndexes) == 0 || len(state.Vagabond.CardsInHand) == 0 {
		return nil
	}

	clearing, ok := vagabondCurrentClearing(state)
	if !ok {
		return nil
	}

	actions := []game.Action{}
	for _, targetFaction := range vagabondFactionsInClearing(clearing) {
		if game.VagabondHostileTo(state, targetFaction) {
			continue
		}

		for _, card := range state.Vagabond.CardsInHand {
			if !matchesSuitOrBird(card, clearing.Suit) {
				continue
			}

			for _, itemIndex := range readyItemIndexes {
				actions = append(actions, game.Action{
					Type: game.ActionAid,
					Aid: &game.AidAction{
						Faction:       game.Vagabond,
						TargetFaction: targetFaction,
						ClearingID:    clearing.ID,
						CardID:        card.ID,
						ItemIndex:     itemIndex,
					},
				})
			}
		}
	}

	return actions
}
