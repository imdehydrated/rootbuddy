package rules

import (
	"sort"
	"strconv"
	"strings"

	"github.com/imdehydrated/rootbuddy/game"
)

func vagabondCurrentClearing(state game.GameState) (game.Clearing, bool) {
	if state.Vagabond.InForest || state.Vagabond.ClearingID == 0 {
		return game.Clearing{}, false
	}

	return findClearingByID(state.Map, state.Vagabond.ClearingID)
}

func findForestByID(m game.Map, id int) (game.Forest, bool) {
	for _, forest := range m.Forests {
		if forest.ID == id {
			return forest, true
		}
	}

	return game.Forest{}, false
}

func vagabondCurrentForest(state game.GameState) (game.Forest, bool) {
	if !state.Vagabond.InForest || state.Vagabond.ForestID == 0 {
		return game.Forest{}, false
	}

	return findForestByID(state.Map, state.Vagabond.ForestID)
}

func vagabondRelationshipLevel(state game.GameState, faction game.Faction) game.RelationshipLevel {
	if state.Vagabond.Relationships == nil {
		return game.RelIndifferent
	}

	relationship, ok := state.Vagabond.Relationships[faction]
	if !ok {
		return game.RelIndifferent
	}

	return relationship
}

func vagabondItemIndexes(state game.GameState, itemType game.ItemType, statuses ...game.ItemStatus) []int {
	indexes := []int{}
	statusSet := map[game.ItemStatus]bool{}
	for _, status := range statuses {
		statusSet[status] = true
	}

	for index, item := range state.Vagabond.Items {
		if item.Type != itemType {
			continue
		}
		if len(statusSet) > 0 && !statusSet[item.Status] {
			continue
		}
		indexes = append(indexes, index)
	}

	return indexes
}

func vagabondReadyItemIndexes(state game.GameState) []int {
	indexes := []int{}
	for index, item := range state.Vagabond.Items {
		if item.Status == game.ItemReady {
			indexes = append(indexes, index)
		}
	}

	return indexes
}

func vagabondExhaustedItemIndexes(state game.GameState) []int {
	indexes := []int{}
	for index, item := range state.Vagabond.Items {
		if item.Status == game.ItemExhausted {
			indexes = append(indexes, index)
		}
	}

	return indexes
}

func vagabondUndamagedItemIndexes(state game.GameState) []int {
	indexes := []int{}
	for index, item := range state.Vagabond.Items {
		if item.Status != game.ItemDamaged {
			indexes = append(indexes, index)
		}
	}

	return indexes
}

func chooseItemIndexSubsets(indexes []int, choose int) [][]int {
	if choose <= 0 || choose > len(indexes) {
		return nil
	}

	subsets := [][]int{}
	current := make([]int, 0, choose)

	var build func(start int)
	build = func(start int) {
		if len(current) == choose {
			subset := make([]int, len(current))
			copy(subset, current)
			subsets = append(subsets, subset)
			return
		}

		remaining := choose - len(current)
		maxStart := len(indexes) - remaining
		for i := start; i <= maxStart; i++ {
			current = append(current, indexes[i])
			build(i + 1)
			current = current[:len(current)-1]
		}
	}

	build(0)
	return subsets
}

func clearingHasFactionPieces(clearing game.Clearing, faction game.Faction) bool {
	if clearing.Warriors != nil && clearing.Warriors[faction] > 0 {
		return true
	}

	for _, building := range clearing.Buildings {
		if building.Faction == faction {
			return true
		}
	}

	for _, token := range clearing.Tokens {
		if token.Faction == faction {
			return true
		}
	}

	return faction == game.Marquise && clearing.Wood > 0
}

func vagabondFactionsInClearing(clearing game.Clearing) []game.Faction {
	factions := []game.Faction{}
	seen := map[game.Faction]bool{}
	for _, faction := range []game.Faction{game.Marquise, game.Eyrie, game.Alliance} {
		if clearingHasFactionPieces(clearing, faction) {
			factions = append(factions, faction)
			seen[faction] = true
		}
	}

	_ = seen
	return factions
}

func hostileFactionCountInClearing(state game.GameState, clearing game.Clearing) int {
	count := 0
	for _, faction := range vagabondFactionsInClearing(clearing) {
		if game.VagabondHostileTo(state, faction) {
			count++
		}
	}

	return count
}

func forestIDsAdjacentToClearing(m game.Map, clearingID int) []int {
	forestIDs := []int{}
	for _, forest := range m.Forests {
		for _, adjacentClearingID := range forest.AdjacentClearings {
			if adjacentClearingID == clearingID {
				forestIDs = append(forestIDs, forest.ID)
				break
			}
		}
	}

	return forestIDs
}

func questByID(quests []game.Quest, id game.QuestID) (game.Quest, bool) {
	for _, quest := range quests {
		if quest.ID == id {
			return quest, true
		}
	}

	return game.Quest{}, false
}

func readyItemIndexChoicesForTypes(state game.GameState, requiredItems []game.ItemType) [][]int {
	if len(requiredItems) == 0 {
		return nil
	}

	choices := [][]int{}
	current := make([]int, 0, len(requiredItems))
	used := map[int]bool{}
	seen := map[string]bool{}

	var build func(requiredIndex int)
	build = func(requiredIndex int) {
		if requiredIndex == len(requiredItems) {
			chosen := make([]int, len(current))
			copy(chosen, current)
			sort.Ints(chosen)

			keyParts := make([]string, len(chosen))
			for i, index := range chosen {
				keyParts[i] = strconv.Itoa(index)
			}
			key := strings.Join(keyParts, ",")
			if !seen[key] {
				choices = append(choices, chosen)
				seen[key] = true
			}
			return
		}

		for _, itemIndex := range vagabondItemIndexes(state, requiredItems[requiredIndex], game.ItemReady) {
			if used[itemIndex] {
				continue
			}

			used[itemIndex] = true
			current = append(current, itemIndex)
			build(requiredIndex + 1)
			current = current[:len(current)-1]
			delete(used, itemIndex)
		}
	}

	build(0)
	return choices
}
