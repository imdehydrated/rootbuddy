package rules

import "github.com/imdehydrated/rootbuddy/game"

func UsableAllianceBasesBySuit(state game.GameState) map[game.Suit][]int {
	bases := map[game.Suit][]int{}
	for _, clearing := range allianceBaseClearings(state) {
		bases[clearing.Suit] = append(bases[clearing.Suit], clearing.ID)
	}

	return bases
}

func ValidAllianceCraftActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Alliance || state.CurrentPhase != game.Daylight {
		return nil
	}

	actions := []game.Action{}
	bases := UsableAllianceBasesBySuit(state)

	for _, card := range state.Alliance.CardsInHand {
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

		routes := WorkshopIDRoutesForCost(card.CraftingCost, bases)
		if len(routes) == 0 {
			continue
		}

		for _, route := range routes {
			actions = append(actions, craftActionsWithVagabondDamageChoices(state, game.Action{
				Type: game.ActionCraft,
				Craft: &game.CraftAction{
					Faction:               game.Alliance,
					CardID:                card.ID,
					UsedWorkshopClearings: append([]int(nil), route...),
				},
			}, card)...)
		}
	}

	return actions
}
