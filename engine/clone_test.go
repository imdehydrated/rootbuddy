package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestCloneStateCopiesExpandedCollections(t *testing.T) {
	state := game.GameState{
		Deck:               []game.CardID{1, 2, 3},
		DiscardPile:        []game.CardID{4, 5},
		AvailableDominance: []game.CardID{14, 27},
		ActiveDominance: map[game.Faction]game.CardID{
			game.Marquise: 14,
		},
		WinningCoalition: []game.Faction{game.Marquise, game.Vagabond},
		ItemSupply: map[game.ItemType]int{
			game.ItemBag: 2,
		},
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {15, 17},
		},
		QuestDeck:    []game.QuestID{1, 2},
		QuestDiscard: []game.QuestID{3},
		OtherHandCounts: map[game.Faction]int{
			game.Eyrie: 4,
		},
		HiddenCards: []game.HiddenCard{
			{ID: 1, OwnerFaction: game.Eyrie, Zone: game.HiddenCardZoneHand},
		},
	}

	cloned := cloneState(state)
	cloned.Deck[0] = 99
	cloned.DiscardPile[0] = 98
	cloned.AvailableDominance[0] = 97
	cloned.ActiveDominance[game.Marquise] = 96
	cloned.WinningCoalition[0] = game.Eyrie
	cloned.ItemSupply[game.ItemBag] = 0
	cloned.PersistentEffects[game.Marquise][0] = 97
	cloned.QuestDeck[0] = 96
	cloned.QuestDiscard[0] = 95
	cloned.OtherHandCounts[game.Eyrie] = 1
	cloned.HiddenCards[0].Zone = game.HiddenCardZoneSupporters

	if state.Deck[0] != 1 {
		t.Fatalf("expected deck to be deep-copied, got %+v", state.Deck)
	}
	if state.DiscardPile[0] != 4 {
		t.Fatalf("expected discard pile to be deep-copied, got %+v", state.DiscardPile)
	}
	if state.AvailableDominance[0] != 14 {
		t.Fatalf("expected available dominance to be deep-copied, got %+v", state.AvailableDominance)
	}
	if state.ActiveDominance[game.Marquise] != 14 {
		t.Fatalf("expected active dominance to be deep-copied, got %+v", state.ActiveDominance)
	}
	if state.WinningCoalition[0] != game.Marquise {
		t.Fatalf("expected winning coalition to be deep-copied, got %+v", state.WinningCoalition)
	}
	if state.ItemSupply[game.ItemBag] != 2 {
		t.Fatalf("expected item supply to be deep-copied, got %+v", state.ItemSupply)
	}
	if state.PersistentEffects[game.Marquise][0] != 15 {
		t.Fatalf("expected persistent effects to be deep-copied, got %+v", state.PersistentEffects)
	}
	if state.QuestDeck[0] != 1 {
		t.Fatalf("expected quest deck to be deep-copied, got %+v", state.QuestDeck)
	}
	if state.QuestDiscard[0] != 3 {
		t.Fatalf("expected quest discard to be deep-copied, got %+v", state.QuestDiscard)
	}
	if state.OtherHandCounts[game.Eyrie] != 4 {
		t.Fatalf("expected other hand counts to be deep-copied, got %+v", state.OtherHandCounts)
	}
	if state.HiddenCards[0].Zone != game.HiddenCardZoneHand {
		t.Fatalf("expected hidden cards to be deep-copied, got %+v", state.HiddenCards)
	}
}
