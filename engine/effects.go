package engine

import "github.com/imdehydrated/rootbuddy/game"

func hasPersistentEffect(state game.GameState, faction game.Faction, effectID string) bool {
	for _, cardID := range state.PersistentEffects[faction] {
		card, ok := CardByID(cardID)
		if ok && card.EffectID == effectID {
			return true
		}
	}

	return false
}

func persistentEffectCardID(state game.GameState, faction game.Faction, effectID string) (game.CardID, bool) {
	for _, cardID := range state.PersistentEffects[faction] {
		card, ok := CardByID(cardID)
		if ok && card.EffectID == effectID {
			return cardID, true
		}
	}

	return 0, false
}

func consumePersistentEffect(state *game.GameState, faction game.Faction, effectID string) bool {
	for _, cardID := range state.PersistentEffects[faction] {
		card, ok := CardByID(cardID)
		if !ok || card.EffectID != effectID {
			continue
		}
		if !removePersistentEffect(state, faction, cardID) {
			return false
		}
		DiscardCard(state, cardID)
		return true
	}

	return false
}

func persistentEffectUsedThisTurn(state game.GameState, effectID string) bool {
	for _, usedEffectID := range state.TurnProgress.UsedPersistentEffectIDs {
		if usedEffectID == effectID {
			return true
		}
	}

	return false
}

func markPersistentEffectUsed(state *game.GameState, effectID string) {
	if effectID == "" || persistentEffectUsedThisTurn(*state, effectID) {
		return
	}

	state.TurnProgress.UsedPersistentEffectIDs = append(state.TurnProgress.UsedPersistentEffectIDs, effectID)
}

func effectStartsPhase(effectID string) bool {
	switch effectID {
	case "better_burrow_bank", "royal_claim", "stand_and_deliver", "command_warren", "cobbler", "codebreakers":
		return true
	default:
		return false
	}
}

func otherFactionsInTurnOrder(state game.GameState, faction game.Faction) []game.Faction {
	others := []game.Faction{}
	for _, other := range effectiveTurnOrder(state) {
		if other == faction {
			continue
		}
		others = append(others, other)
	}

	return others
}

func factionsWithWarriors(state game.GameState, faction game.Faction) []int {
	clearingIDs := []int{}
	for _, clearing := range state.Map.Clearings {
		if clearing.Warriors[faction] > 0 {
			clearingIDs = append(clearingIDs, clearing.ID)
		}
	}

	return clearingIDs
}

func addPersistentEffect(state *game.GameState, faction game.Faction, cardID game.CardID) {
	if cardID <= 0 {
		return
	}
	if state.PersistentEffects == nil {
		state.PersistentEffects = map[game.Faction][]game.CardID{}
	}
	state.PersistentEffects[faction] = append(state.PersistentEffects[faction], cardID)
}

func removePersistentEffect(state *game.GameState, faction game.Faction, cardID game.CardID) bool {
	cardIDs := state.PersistentEffects[faction]
	for index, existing := range cardIDs {
		if existing != cardID {
			continue
		}

		state.PersistentEffects[faction] = append(cardIDs[:index], cardIDs[index+1:]...)
		if len(state.PersistentEffects[faction]) == 0 {
			delete(state.PersistentEffects, faction)
		}
		return true
	}

	return false
}

func favorSuit(effectID string) (game.Suit, bool) {
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

func removeAllFactionPiecesFromClearing(state *game.GameState, clearing *game.Clearing, faction game.Faction) (int, int, int, int) {
	removedWarriors := 0
	if clearing.Warriors != nil {
		removedWarriors = clearing.Warriors[faction]
		clearing.Warriors[faction] = 0
		returnRemovedWarriorsToSupply(state, clearing, faction, removedWarriors)
	}

	buildingLosses := len(clearing.Buildings)
	beforeBuildings := len(clearing.Buildings)
	removeBuildingLosses(state, clearing, faction, buildingLosses)
	removedBuildings := beforeBuildings - len(clearing.Buildings)

	tokenLosses := len(clearing.Tokens) + clearing.Wood
	_, removedTokens, removedSympathy := removeTokenLosses(state, clearing, faction, tokenLosses)

	return removedWarriors, removedBuildings, removedTokens, removedSympathy
}

func resolveFavorCard(state *game.GameState, faction game.Faction, card game.Card) {
	suit, ok := favorSuit(card.EffectID)
	if !ok {
		return
	}

	for index := range state.Map.Clearings {
		clearing := &state.Map.Clearings[index]
		if clearing.Suit != suit {
			continue
		}

		removedBuildings := 0
		removedTokens := 0
		removedSympathy := 0
		for _, target := range []game.Faction{game.Marquise, game.Eyrie, game.Alliance} {
			if !game.AreEnemies(*state, faction, target) {
				continue
			}

			_, targetBuildings, targetTokens, targetSympathy := removeAllFactionPiecesFromClearing(state, clearing, target)
			removedBuildings += targetBuildings
			removedTokens += targetTokens
			removedSympathy += targetSympathy
		}

		addVictoryPoints(state, faction, removedBuildings+removedTokens)
		if removedSympathy > 0 && faction != game.Alliance {
			transferOutrageCard(state, faction, clearing.Suit)
		}
		if faction != game.Vagabond &&
			game.AreEnemies(*state, faction, game.Vagabond) &&
			!state.Vagabond.InForest &&
			state.Vagabond.ClearingID == clearing.ID {
			damageVagabondItems(state, 3)
		}
	}
}

func resolveCraftedCard(state *game.GameState, faction game.Faction, card game.Card) {
	points := card.VP
	if faction == game.Eyrie && card.CraftedItem != nil && state.Eyrie.Leader != game.LeaderBuilder {
		points = 1
	}
	addVictoryPoints(state, faction, points)

	switch card.Kind {
	case game.PersistentEffectCard:
		addPersistentEffect(state, faction, card.ID)
	case game.OneTimeEffectCard:
		if _, ok := favorSuit(card.EffectID); ok {
			resolveFavorCard(state, faction, card)
		}
		DiscardCard(state, card.ID)
	default:
		DiscardCard(state, card.ID)
	}
}
