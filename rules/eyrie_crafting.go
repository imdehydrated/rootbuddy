package rules

import "github.com/imdehydrated/rootbuddy/game"

func usableRoostClearingsBySuit(state game.GameState) map[game.Suit][]int {
	roosts := map[game.Suit][]int{}
	for _, clearing := range state.Map.Clearings {
		total := roostCountInClearing(clearing)
		if total <= 0 {
			continue
		}

		for i := 0; i < total; i++ {
			roosts[clearing.Suit] = append(roosts[clearing.Suit], clearing.ID)
		}
	}

	return roosts
}

func ValidEyrieCraftActions(state game.GameState) []game.Action {
	actions := []game.Action{}

	if state.FactionTurn != game.Eyrie {
		return actions
	}

	if state.CurrentPhase != game.Daylight {
		return actions
	}

	roosts := usableRoostClearingsBySuit(state)

	for _, card := range state.Eyrie.CardsInHand {
		if !isCraftable(card.Kind) {
			continue
		}
		if card.CraftingCost.Fox == 0 &&
			card.CraftingCost.Rabbit == 0 &&
			card.CraftingCost.Mouse == 0 &&
			card.CraftingCost.Any == 0 {
			continue
		}

		usedRoostIDs, ok := workshopIDsForCost(card.CraftingCost, roosts)
		if !ok {
			continue
		}

		actions = append(actions, game.Action{
			Type: game.ActionCraft,
			Craft: &game.CraftAction{
				Faction:               game.Eyrie,
				CardID:                card.ID,
				UsedWorkshopClearings: usedRoostIDs,
			},
		})
	}

	return actions
}
