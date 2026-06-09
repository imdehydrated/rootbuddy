package rules

import "github.com/imdehydrated/rootbuddy/game"

func UsableRoostClearingsBySuit(state game.GameState) map[game.Suit][]int {
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

	roosts := UsableRoostClearingsBySuit(state)

	for _, card := range state.Eyrie.CardsInHand {
		if !isCraftable(card.Kind) {
			continue
		}
		if !itemCraftAvailable(state, card) {
			continue
		}
		if card.CraftingCost.Fox == 0 &&
			card.CraftingCost.Rabbit == 0 &&
			card.CraftingCost.Mouse == 0 &&
			card.CraftingCost.Any == 0 {
			continue
		}

		routes := WorkshopIDRoutesForCost(card.CraftingCost, roosts)
		if len(routes) == 0 {
			continue
		}

		for _, route := range routes {
			actions = append(actions, craftActionsWithVagabondDamageChoices(state, game.Action{
				Type: game.ActionCraft,
				Craft: &game.CraftAction{
					Faction:               game.Eyrie,
					CardID:                card.ID,
					UsedWorkshopClearings: append([]int(nil), route...),
				},
			}, card)...)
		}
	}

	return actions
}
