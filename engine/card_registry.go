package engine

import (
	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

var cardRegistry = buildCardRegistry()

func buildCardRegistry() map[game.CardID]game.Card {
	registry := make(map[game.CardID]game.Card, len(carddata.BaseDeck()))
	for _, card := range carddata.BaseDeck() {
		registry[card.ID] = card
	}
	return registry
}

func CardByID(id game.CardID) (game.Card, bool) {
	card, ok := cardRegistry[id]
	return card, ok
}
