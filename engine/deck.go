package engine

import (
	"sort"

	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

func BuildDeck(cards []game.Card) []game.CardID {
	deck := make([]game.CardID, 0, len(cards))
	for _, card := range cards {
		if card.ID <= 0 {
			continue
		}
		deck = append(deck, card.ID)
	}
	return deck
}

func ShuffleDeck(state *game.GameState, deck []game.CardID) []game.CardID {
	shuffled := cloneCardIDs(deck)
	rng := nextShuffleRNG(state)
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

func ReshuffleDeck(state *game.GameState) {
	if len(state.Deck) > 0 || len(state.DiscardPile) == 0 {
		return
	}

	state.Deck = ShuffleDeck(state, state.DiscardPile)
	state.DiscardPile = nil
}

func tracksHandForFaction(state game.GameState, faction game.Faction) bool {
	return state.TrackAllHands || faction == state.PlayerFaction
}

func drawOneCardID(state *game.GameState) (game.CardID, bool) {
	if len(state.Deck) == 0 {
		ReshuffleDeck(state)
	}
	if len(state.Deck) == 0 {
		return 0, false
	}

	cardID := state.Deck[0]
	state.Deck = state.Deck[1:]
	return cardID, true
}

func DrawCards(state *game.GameState, faction game.Faction, count int) []game.Card {
	if state.GameMode != game.GameModeOnline || count <= 0 {
		return nil
	}

	drawn := []game.Card{}
	for i := 0; i < count; i++ {
		cardID, ok := drawOneCardID(state)
		if !ok {
			break
		}

		if tracksHandForFaction(*state, faction) {
			card, found := CardByID(cardID)
			if !found {
				continue
			}
			appendCardToFactionHand(state, faction, card)
			drawn = append(drawn, card)
			continue
		}

		incrementOtherHandCount(state, faction, 1)
	}

	return drawn
}

func DiscardCard(state *game.GameState, cardID game.CardID) {
	if cardID <= 0 {
		return
	}
	if card, ok := CardByID(cardID); ok && card.Kind == game.DominanceCard {
		addAvailableDominance(state, cardID)
		return
	}
	state.DiscardPile = append(state.DiscardPile, cardID)
}

func DiscardCards(state *game.GameState, cardIDs []game.CardID) {
	for _, cardID := range cardIDs {
		DiscardCard(state, cardID)
	}
}

func addKnownCardID(known map[game.CardID]struct{}, cardID game.CardID) {
	if cardID <= 0 {
		return
	}
	known[cardID] = struct{}{}
}

func addKnownCards(known map[game.CardID]struct{}, cards []game.Card) {
	for _, card := range cards {
		addKnownCardID(known, card.ID)
	}
}

func playerHand(state game.GameState) []game.Card {
	switch state.PlayerFaction {
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

func KnownCardIDs(state game.GameState) []game.CardID {
	known := map[game.CardID]struct{}{}

	for _, cardID := range state.DiscardPile {
		addKnownCardID(known, cardID)
	}

	for _, cardIDs := range state.PersistentEffects {
		for _, cardID := range cardIDs {
			addKnownCardID(known, cardID)
		}
	}

	addKnownCards(known, playerHand(state))

	if state.PlayerFaction == game.Alliance {
		addKnownCards(known, state.Alliance.Supporters)
	}

	if state.PlayerFaction == game.Eyrie {
		for _, cardID := range state.Eyrie.Decree.Recruit {
			addKnownCardID(known, cardID)
		}
		for _, cardID := range state.Eyrie.Decree.Move {
			addKnownCardID(known, cardID)
		}
		for _, cardID := range state.Eyrie.Decree.Battle {
			addKnownCardID(known, cardID)
		}
		for _, cardID := range state.Eyrie.Decree.Build {
			addKnownCardID(known, cardID)
		}
	}

	cardIDs := make([]game.CardID, 0, len(known))
	for cardID := range known {
		cardIDs = append(cardIDs, cardID)
	}
	sort.Slice(cardIDs, func(i, j int) bool {
		return cardIDs[i] < cardIDs[j]
	})

	return cardIDs
}

func UnknownCardCount(state game.GameState) int {
	unknown := len(carddata.BaseDeck()) - len(KnownCardIDs(state))
	if unknown < 0 {
		return 0
	}
	return unknown
}

func incrementOtherHandCount(state *game.GameState, faction game.Faction, count int) {
	if count <= 0 || faction == state.PlayerFaction {
		return
	}
	if state.GameMode == game.GameModeAssist {
		for i := 0; i < count; i++ {
			addHiddenCard(state, faction, game.HiddenCardZoneHand, 0)
		}
		return
	}
	if state.OtherHandCounts == nil {
		state.OtherHandCounts = map[game.Faction]int{}
	}
	state.OtherHandCounts[faction] += count
}

func decrementOtherHandCount(state *game.GameState, faction game.Faction, count int) {
	if count <= 0 || faction == state.PlayerFaction {
		return
	}
	if state.GameMode == game.GameModeAssist {
		consumeHiddenCards(state, faction, game.HiddenCardZoneHand, count)
		return
	}
	if state.OtherHandCounts == nil {
		return
	}
	state.OtherHandCounts[faction] -= count
	if state.OtherHandCounts[faction] < 0 {
		state.OtherHandCounts[faction] = 0
	}
}
