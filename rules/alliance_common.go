package rules

import "github.com/imdehydrated/rootbuddy/game"

var allianceSympathyTrack = []int{0, 1, 1, 1, 2, 2, 2, 3, 3, 4}

func matchesSuitOrBird(card game.Card, suit game.Suit) bool {
	return card.Suit == suit || card.Suit == game.Bird
}

func allianceSupporterCost(sympathyPlaced int) int {
	switch {
	case sympathyPlaced >= 5:
		return 3
	case sympathyPlaced >= 3:
		return 2
	default:
		return 1
	}
}

func allianceSympathyPoints(sympathyPlaced int) int {
	if sympathyPlaced < 0 || sympathyPlaced >= len(allianceSympathyTrack) {
		return 0
	}

	return allianceSympathyTrack[sympathyPlaced]
}

func hasAllianceSympathy(clearing game.Clearing) bool {
	for _, token := range clearing.Tokens {
		if token.Faction == game.Alliance && token.Type == game.TokenSympathy {
			return true
		}
	}

	return false
}

func hasKeepToken(clearing game.Clearing) bool {
	for _, token := range clearing.Tokens {
		if token.Faction == game.Marquise && token.Type == game.TokenKeep {
			return true
		}
	}

	return false
}

func adjacentToAllianceSympathy(clearing game.Clearing, board game.Map) bool {
	for _, adjacentID := range clearing.Adj {
		adjacent, ok := findClearingByID(board, adjacentID)
		if !ok {
			continue
		}

		if hasAllianceSympathy(adjacent) {
			return true
		}
	}

	return false
}

func allianceSupporterCardIDs(state game.GameState, suit game.Suit) []game.CardID {
	cardIDs := make([]game.CardID, 0, len(state.Alliance.Supporters))
	for _, card := range state.Alliance.Supporters {
		if matchesSuitOrBird(card, suit) {
			cardIDs = append(cardIDs, card.ID)
		}
	}

	return cardIDs
}

func supporterCardSubsets(cardIDs []game.CardID, choose int) [][]game.CardID {
	if choose <= 0 || choose > len(cardIDs) {
		return nil
	}

	subsets := [][]game.CardID{}
	current := make([]game.CardID, 0, choose)

	var build func(start int)
	build = func(start int) {
		if len(current) == choose {
			subset := make([]game.CardID, len(current))
			copy(subset, current)
			subsets = append(subsets, subset)
			return
		}

		remaining := choose - len(current)
		maxStart := len(cardIDs) - remaining
		for i := start; i <= maxStart; i++ {
			current = append(current, cardIDs[i])
			build(i + 1)
			current = current[:len(current)-1]
		}
	}

	build(0)
	return subsets
}

func allianceBaseClearings(state game.GameState) []game.Clearing {
	clearings := []game.Clearing{}
	for _, clearing := range state.Map.Clearings {
		for _, building := range clearing.Buildings {
			if building.Faction == game.Alliance && building.Type == game.Base {
				clearings = append(clearings, clearing)
				break
			}
		}
	}

	return clearings
}

func allianceHasBaseInSuit(state game.GameState, suit game.Suit) bool {
	for _, clearing := range allianceBaseClearings(state) {
		if clearing.Suit == suit {
			return true
		}
	}

	return false
}

func allianceHasAnyBase(state game.GameState) bool {
	return len(allianceBaseClearings(state)) > 0
}

func sympathyCountBySuit(board game.Map, suit game.Suit) int {
	count := 0
	for _, clearing := range board.Clearings {
		if clearing.Suit != suit || !hasAllianceSympathy(clearing) {
			continue
		}
		count++
	}

	return count
}
