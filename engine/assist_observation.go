package engine

import "github.com/imdehydrated/rootbuddy/game"

func canUseObservedHiddenCards(state game.GameState, faction game.Faction) bool {
	return state.GameMode == game.GameModeAssist && !tracksHandForFaction(state, faction)
}

func spendFactionHandCard(state *game.GameState, faction game.Faction, cardID game.CardID) (game.Card, bool) {
	if card, ok := removeCardFromFactionHand(state, faction, cardID); ok {
		return card, true
	}

	if !canUseObservedHiddenCards(*state, faction) {
		return game.Card{}, false
	}

	consumeHiddenCards(state, faction, game.HiddenCardZoneHand, 1)
	card, ok := CardByID(cardID)
	return card, ok
}

func spendAllianceSupporters(state *game.GameState, cardIDs []game.CardID) {
	if canUseObservedHiddenCards(*state, game.Alliance) {
		consumeHiddenCards(state, game.Alliance, game.HiddenCardZoneSupporters, len(cardIDs))
		return
	}

	state.Alliance.Supporters = removeCardsByID(state.Alliance.Supporters, cardIDs)
}
