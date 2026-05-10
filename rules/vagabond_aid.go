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
		for _, card := range state.Vagabond.CardsInHand {
			if !matchesSuitOrBird(card, clearing.Suit) {
				continue
			}

			for _, itemIndex := range readyItemIndexes {
				actions = append(actions, aidAction(clearing.ID, targetFaction, card.ID, itemIndex, nil))
				for craftedItemIndex := range state.CraftedItems[targetFaction] {
					takeIndex := craftedItemIndex
					actions = append(actions, aidAction(clearing.ID, targetFaction, card.ID, itemIndex, &takeIndex))
				}
			}
		}
	}

	return actions
}

func aidAction(clearingID int, targetFaction game.Faction, cardID game.CardID, itemIndex int, takeItemIndex *int) game.Action {
	return game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: targetFaction,
			ClearingID:    clearingID,
			CardID:        cardID,
			ItemIndex:     itemIndex,
			TakeItemIndex: takeItemIndex,
		},
	}
}
