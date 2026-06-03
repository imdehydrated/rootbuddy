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
	return len(LegalAmbushCardIDs(state, faction, suit)) > 0 || (!tracksHandForFaction(state, faction) && state.OtherHandCounts[faction] > 0)
}

func LegalAmbushCardIDs(state game.GameState, faction game.Faction, suit game.Suit) []game.CardID {
	if tracksHandForFaction(state, faction) {
		cardIDs := []game.CardID{}
		for _, card := range factionHand(state, faction) {
			if card.Kind == game.AmbushCard && cardMatchesSuitOrBird(card, suit) {
				cardIDs = append(cardIDs, card.ID)
			}
		}
		return cardIDs
	}

	cardIDs := []game.CardID{}
	for _, hidden := range state.HiddenCards {
		if hidden.OwnerFaction != faction || hidden.Zone != game.HiddenCardZoneHand || hidden.KnownCardID <= 0 {
			continue
		}
		card, ok := CardByID(hidden.KnownCardID)
		if ok && card.Kind == game.AmbushCard && cardMatchesSuitOrBird(card, suit) {
			cardIDs = append(cardIDs, card.ID)
		}
	}
	return cardIDs
}

func resolveAmbushCardID(state game.GameState, faction game.Faction, suit game.Suit, requested game.CardID) (game.CardID, bool) {
	cardIDs := LegalAmbushCardIDs(state, faction, suit)
	if requested > 0 {
		for _, cardID := range cardIDs {
			if cardID == requested {
				return requested, true
			}
		}
		if !tracksHandForFaction(state, faction) && state.OtherHandCounts[faction] > 0 {
			card, ok := CardByID(requested)
			return requested, ok && card.Kind == game.AmbushCard && cardMatchesSuitOrBird(card, suit)
		}
		return 0, false
	}
	if len(cardIDs) == 1 {
		return cardIDs[0], true
	}
	if !tracksHandForFaction(state, faction) && state.OtherHandCounts[faction] > 0 && len(cardIDs) == 0 {
		return 0, true
	}
	return 0, false
}

func consumeAmbushCard(state *game.GameState, faction game.Faction, suit game.Suit, cardID game.CardID) bool {
	if !tracksHandForFaction(*state, faction) {
		if cardID > 0 {
			card, ok := CardByID(cardID)
			if !ok || card.Kind != game.AmbushCard || !cardMatchesSuitOrBird(card, suit) {
				return false
			}
		}
		if state.OtherHandCounts[faction] <= 0 {
			return false
		}
		decrementOtherHandCount(state, faction, 1)
		return true
	}

	for _, card := range factionHand(*state, faction) {
		if cardID > 0 && card.ID != cardID {
			continue
		}
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
		return requiredVagabondDamageCount(*state, hits)
	}

	index := findClearingIndex(state.Map, clearingID)
	if index == -1 {
		return 0
	}

	remaining := removeWarriorLosses(state, &state.Map.Clearings[index], faction, hits)
	return hits - remaining
}

func attackersRemainAfterAmbush(state game.GameState, faction game.Faction, clearingID int) bool {
	if faction == game.Vagabond {
		return true
	}

	return warriorCountInClearing(state, clearingID, faction) > 0
}
