package rules

import (
	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

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

func itemCraftAvailable(state game.GameState, card game.Card) bool {
	if card.CraftedItem == nil {
		return !persistentEffectAlreadyCrafted(state, card)
	}

	if state.ItemSupply == nil {
		return !persistentEffectAlreadyCrafted(state, card)
	}

	return state.ItemSupply[*card.CraftedItem] > 0 && !persistentEffectAlreadyCrafted(state, card)
}

func persistentEffectAlreadyCrafted(state game.GameState, card game.Card) bool {
	if card.Kind != game.PersistentEffectCard || card.EffectID == "" {
		return false
	}

	for _, cardID := range state.PersistentEffects[state.FactionTurn] {
		if craftedCard, ok := persistentCardByID(cardID); ok && craftedCard.EffectID == card.EffectID {
			return true
		}
	}

	return false
}

func persistentCardByID(cardID game.CardID) (game.Card, bool) {
	for _, card := range carddata.BaseDeck() {
		if card.ID == cardID {
			return card, true
		}
	}

	return game.Card{}, false
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
		if !itemCraftAvailable(state, card) {
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

		actions = append(actions, craftActionsWithVagabondDamageChoices(state, game.Action{
			Type: game.ActionCraft,
			Craft: &game.CraftAction{
				Faction:               game.Marquise,
				CardID:                card.ID,
				UsedWorkshopClearings: usedWorkshopIDs,
			},
		}, card)...)
	}

	return actions
}

func craftActionsWithVagabondDamageChoices(state game.GameState, action game.Action, card game.Card) []game.Action {
	if action.Craft == nil {
		return []game.Action{action}
	}

	hits := craftVagabondDamageHits(state, action.Craft.Faction, card)
	choices := vagabondDamageIndexChoices(state, hits)
	actions := make([]game.Action, 0, len(choices))
	for _, damagedItemIndexes := range choices {
		next := action
		next.Craft = &game.CraftAction{
			Faction:                    action.Craft.Faction,
			CardID:                     action.Craft.CardID,
			UsedWorkshopClearings:      append([]int(nil), action.Craft.UsedWorkshopClearings...),
			DamagedVagabondItemIndexes: damagedItemIndexes,
		}
		actions = append(actions, next)
	}
	return actions
}

func craftVagabondDamageHits(state game.GameState, faction game.Faction, card game.Card) int {
	suit, ok := craftFavorSuit(card.EffectID)
	if !ok || faction == game.Vagabond || state.Vagabond.InForest || state.Vagabond.ClearingID == 0 || !game.AreEnemies(state, faction, game.Vagabond) {
		return 0
	}

	clearing, ok := findClearingByID(state.Map, state.Vagabond.ClearingID)
	if !ok || clearing.Suit != suit {
		return 0
	}

	return 3
}

func craftFavorSuit(effectID string) (game.Suit, bool) {
	switch effectID {
	case "favor_foxes":
		return game.Fox, true
	case "favor_rabbits":
		return game.Rabbit, true
	case "favor_mice":
		return game.Mouse, true
	default:
		return 0, false
	}
}
