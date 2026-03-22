package rules

import "github.com/imdehydrated/rootbuddy/game"

func isCraftable(kind game.CardKind) bool {
	return kind == game.ItemCard ||
		kind == game.PersistentEffectCard ||
		kind == game.OneTimeEffectCard
}

func usedWorkshopCount(id int, used []int) int {
	count := 0
	for _, usedID := range used {
		if usedID == id {
			count++
		}
	}
	return count
}

func marquiseWorkshopCount(c game.Clearing) int {
	count := 0
	for _, building := range c.Buildings {
		if building.Faction == game.Marquise && building.Type == game.Workshop {
			count++
		}
	}
	return count
}

func usableWorkshopClearingsBySuit(state game.GameState) map[game.Suit][]int {
	workshops := map[game.Suit][]int{}
	for _, clearing := range state.Map.Clearings {
		total := marquiseWorkshopCount(clearing)
		if total == 0 {
			continue
		}

		used := usedWorkshopCount(clearing.ID, state.TurnProgress.UsedWorkshopClearings)
		available := total - used
		if available <= 0 {
			continue
		}

		for i := 0; i < available; i++ {
			workshops[clearing.Suit] = append(workshops[clearing.Suit], clearing.ID)
		}
	}

	return workshops
}

func workshopIDsForCost(cost game.CraftingCost, workshops map[game.Suit][]int) ([]int, bool) {
	chosen := []int{}
	remaining := map[game.Suit][]int{}

	for suit, ids := range workshops {
		copied := make([]int, len(ids))
		copy(copied, ids)
		remaining[suit] = copied
	}

	claimFromSuit := func(suit game.Suit) bool {
		available := remaining[suit]
		if len(available) == 0 {
			return false
		}

		id := available[0]
		remaining[suit] = available[1:]
		chosen = append(chosen, id)
		return true
	}

	claimAny := func() bool {
		for _, suit := range []game.Suit{game.Fox, game.Rabbit, game.Mouse, game.Bird} {
			available := remaining[suit]
			if len(available) == 0 {
				continue
			}

			id := available[0]
			remaining[suit] = available[1:]
			chosen = append(chosen, id)
			return true
		}
		return false
	}

	for i := 0; i < cost.Fox; i++ {
		if !claimFromSuit(game.Fox) {
			return nil, false
		}
	}
	for i := 0; i < cost.Rabbit; i++ {
		if !claimFromSuit(game.Rabbit) {
			return nil, false
		}
	}
	for i := 0; i < cost.Mouse; i++ {
		if !claimFromSuit(game.Mouse) {
			return nil, false
		}
	}
	for i := 0; i < cost.Any; i++ {
		if !claimAny() {
			return nil, false
		}
	}

	return chosen, true
}

func ValidCraftActions(state game.GameState) []game.Action {
	actions := []game.Action{}

	if state.FactionTurn != game.Marquise {
		return actions
	}

	if state.CurrentPhase != game.Daylight {
		return actions
	}

	workshops := usableWorkshopClearingsBySuit(state)

	for _, card := range state.Marquise.CardsInHand {
		if !isCraftable(card.Kind) {
			continue
		}
		if card.CraftingCost.Fox == 0 &&
			card.CraftingCost.Rabbit == 0 &&
			card.CraftingCost.Mouse == 0 &&
			card.CraftingCost.Any == 0 {
			continue
		}

		usedWorkshopIDs, ok := workshopIDsForCost(card.CraftingCost, workshops)
		if !ok {
			continue
		}

		actions = append(actions, game.Action{
			Type: game.ActionCraft,
			Craft: &game.CraftAction{
				Faction:               game.Marquise,
				CardID:                card.ID,
				UsedWorkshopClearings: usedWorkshopIDs,
			},
		})
	}

	return actions
}
