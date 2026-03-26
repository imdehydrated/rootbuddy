package engine

import "github.com/imdehydrated/rootbuddy/game"

func factionHand(state game.GameState, faction game.Faction) []game.Card {
	switch faction {
	case game.Marquise:
		return state.Marquise.CardsInHand
	case game.Eyrie:
		return state.Eyrie.CardsInHand
	case game.Alliance:
		return state.Alliance.CardsInHand
	case game.Vagabond:
		return state.Vagabond.CardsInHand
	default:
		return nil
	}
}

func clearingSuit(state game.GameState, clearingID int) game.Suit {
	index := findClearingIndex(state.Map, clearingID)
	if index == -1 {
		return game.Bird
	}

	return state.Map.Clearings[index].Suit
}

func canFactionPlayAmbush(state game.GameState, faction game.Faction, suit game.Suit) bool {
	if tracksHandForFaction(state, faction) {
		for _, card := range factionHand(state, faction) {
			if card.Kind == game.AmbushCard && cardMatchesSuitOrBird(card, suit) {
				return true
			}
		}
		return false
	}

	return state.OtherHandCounts[faction] > 0
}

func consumeAmbushCard(state *game.GameState, faction game.Faction, suit game.Suit) bool {
	if !tracksHandForFaction(*state, faction) {
		if state.OtherHandCounts[faction] <= 0 {
			return false
		}
		decrementOtherHandCount(state, faction, 1)
		return true
	}

	for _, card := range factionHand(*state, faction) {
		if card.Kind != game.AmbushCard || !cardMatchesSuitOrBird(card, suit) {
			continue
		}

		if _, ok := removeCardFromFactionHand(state, faction, card.ID); !ok {
			return false
		}
		DiscardCard(state, card.ID)
		return true
	}

	return false
}

func applyHypotheticalAmbushHits(state *game.GameState, faction game.Faction, clearingID int, hits int) int {
	if hits <= 0 {
		return 0
	}

	if faction == game.Vagabond {
		return damageVagabondItems(state, hits)
	}

	index := findClearingIndex(state.Map, clearingID)
	if index == -1 {
		return 0
	}

	remaining := removeWarriorLosses(&state.Map.Clearings[index], faction, hits)
	return hits - remaining
}

func attackersRemainAfterAmbush(state game.GameState, faction game.Faction, clearingID int) bool {
	if faction == game.Vagabond {
		return true
	}

	return warriorCountInClearing(state, clearingID, faction) > 0
}
