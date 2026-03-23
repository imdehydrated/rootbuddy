package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidVagabondCraftActions(state game.GameState) []game.Action {
	clearing, ok := vagabondCurrentClearing(state)
	if !ok {
		return nil
	}

	hammerCount := len(vagabondItemIndexes(state, game.ItemHammer, game.ItemReady))
	if hammerCount == 0 {
		return nil
	}

	workshops := map[game.Suit][]int{}
	for i := 0; i < hammerCount; i++ {
		workshops[clearing.Suit] = append(workshops[clearing.Suit], clearing.ID)
	}

	actions := []game.Action{}
	for _, card := range state.Vagabond.CardsInHand {
		if !isCraftable(card.Kind) {
			continue
		}
		if card.CraftingCost.Fox == 0 &&
			card.CraftingCost.Rabbit == 0 &&
			card.CraftingCost.Mouse == 0 &&
			card.CraftingCost.Any == 0 {
			continue
		}

		usedHammerSlots, ok := workshopIDsForCost(card.CraftingCost, workshops)
		if !ok {
			continue
		}

		actions = append(actions, game.Action{
			Type: game.ActionCraft,
			Craft: &game.CraftAction{
				Faction:               game.Vagabond,
				CardID:                card.ID,
				UsedWorkshopClearings: usedHammerSlots,
			},
		})
	}

	return actions
}
