package rules

import (
	"sort"
	"strconv"

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

func UsableWorkshopClearingsBySuit(state game.GameState) map[game.Suit][]int {
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

func WorkshopIDRoutesForCost(cost game.CraftingCost, workshops map[game.Suit][]int) [][]int {
	if cost.Fox < 0 || cost.Rabbit < 0 || cost.Mouse < 0 || cost.Any < 0 {
		return nil
	}

	states := []workshopRouteState{
		{
			route:     []int{},
			remaining: cloneWorkshopMap(workshops),
		},
	}

	for _, requirement := range []struct {
		suit  game.Suit
		count int
	}{
		{suit: game.Fox, count: cost.Fox},
		{suit: game.Rabbit, count: cost.Rabbit},
		{suit: game.Mouse, count: cost.Mouse},
	} {
		nextStates := []workshopRouteState{}
		for _, state := range states {
			for _, selected := range workshopIDCombinations(state.remaining[requirement.suit], requirement.count) {
				nextRemaining := cloneWorkshopMap(state.remaining)
				nextRemaining[requirement.suit] = removeWorkshopIDs(nextRemaining[requirement.suit], selected)
				nextStates = append(nextStates, workshopRouteState{
					route:     append(append([]int(nil), state.route...), selected...),
					remaining: nextRemaining,
				})
			}
		}
		if len(nextStates) == 0 {
			return nil
		}
		states = nextStates
	}

	routes := [][]int{}
	for _, state := range states {
		availableAny := []int{}
		for _, suit := range []game.Suit{game.Fox, game.Rabbit, game.Mouse, game.Bird} {
			availableAny = append(availableAny, state.remaining[suit]...)
		}
		for _, selected := range workshopIDCombinations(availableAny, cost.Any) {
			routes = append(routes, append(append([]int(nil), state.route...), selected...))
		}
	}

	return uniqueWorkshopRoutes(routes)
}

type workshopRouteState struct {
	route     []int
	remaining map[game.Suit][]int
}

func cloneWorkshopMap(workshops map[game.Suit][]int) map[game.Suit][]int {
	cloned := map[game.Suit][]int{}
	for suit, ids := range workshops {
		cloned[suit] = append([]int(nil), ids...)
	}
	return cloned
}

func workshopIDCombinations(ids []int, count int) [][]int {
	if count < 0 || count > len(ids) {
		return nil
	}
	if count == 0 {
		return [][]int{{}}
	}

	combinations := [][]int{}
	var choose func(start int, selected []int)
	choose = func(start int, selected []int) {
		if len(selected) == count {
			combinations = append(combinations, append([]int(nil), selected...))
			return
		}
		needed := count - len(selected)
		for index := start; index <= len(ids)-needed; index++ {
			choose(index+1, append(selected, ids[index]))
		}
	}
	choose(0, nil)
	return combinations
}

func removeWorkshopIDs(ids []int, selected []int) []int {
	counts := map[int]int{}
	for _, id := range selected {
		counts[id]++
	}

	remaining := make([]int, 0, len(ids)-len(selected))
	for _, id := range ids {
		if counts[id] > 0 {
			counts[id]--
			continue
		}
		remaining = append(remaining, id)
	}
	return remaining
}

func uniqueWorkshopRoutes(routes [][]int) [][]int {
	seen := map[string]bool{}
	unique := make([][]int, 0, len(routes))
	for _, route := range routes {
		key := workshopRouteKey(route)
		if seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, route)
	}
	return unique
}

func workshopRouteKey(route []int) string {
	canonical := append([]int(nil), route...)
	sort.Ints(canonical)

	key := ""
	for _, id := range canonical {
		key += strconv.Itoa(id) + ","
	}
	return key
}

func WorkshopRouteMatches(route []int, legal []int) bool {
	if len(route) != len(legal) {
		return false
	}
	routeKey := append([]int(nil), route...)
	legalKey := append([]int(nil), legal...)
	sort.Ints(routeKey)
	sort.Ints(legalKey)

	for index := range routeKey {
		if routeKey[index] != legalKey[index] {
			return false
		}
	}
	return true
}

func WorkshopRouteIsLegal(route []int, legalRoutes [][]int) bool {
	for _, legal := range legalRoutes {
		if WorkshopRouteMatches(route, legal) {
			return true
		}
	}
	return false
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

	workshops := UsableWorkshopClearingsBySuit(state)

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

		routes := WorkshopIDRoutesForCost(card.CraftingCost, workshops)
		if len(routes) == 0 {
			continue
		}

		for _, route := range routes {
			actions = append(actions, craftActionsWithVagabondDamageChoices(state, game.Action{
				Type: game.ActionCraft,
				Craft: &game.CraftAction{
					Faction:               game.Marquise,
					CardID:                card.ID,
					UsedWorkshopClearings: append([]int(nil), route...),
				},
			}, card)...)
		}
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
