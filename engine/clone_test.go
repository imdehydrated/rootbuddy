package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestCloneStateCopiesExpandedCollections(t *testing.T) {
	state := game.GameState{
		Deck:        []game.CardID{1, 2, 3},
		DiscardPile: []game.CardID{4, 5},
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
	}

	cloned := cloneState(state)
	cloned.Deck[0] = 99
	cloned.DiscardPile[0] = 98
	cloned.ItemSupply[game.ItemBag] = 0
	cloned.PersistentEffects[game.Marquise][0] = 97
	cloned.QuestDeck[0] = 96
	cloned.QuestDiscard[0] = 95
	cloned.OtherHandCounts[game.Eyrie] = 1

	if state.Deck[0] != 1 {
		t.Fatalf("expected deck to be deep-copied, got %+v", state.Deck)
	}
	if state.DiscardPile[0] != 4 {
		t.Fatalf("expected discard pile to be deep-copied, got %+v", state.DiscardPile)
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
}
