package carddata

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestBaseDeckHas54Cards(t *testing.T) {
	deck := BaseDeck()

	if len(deck) != 54 {
		t.Fatalf("expected 54 cards in base deck, got %d", len(deck))
	}
}

func TestBaseDeckCardIDsAreUnique(t *testing.T) {
	deck := BaseDeck()
	seen := map[game.CardID]bool{}

	for _, card := range deck {
		if seen[card.ID] {
			t.Fatalf("found duplicate card ID %d", card.ID)
		}
		seen[card.ID] = true
	}
}

func TestBaseDeckCardsHaveBaseDeckTag(t *testing.T) {
	deck := BaseDeck()

	for _, card := range deck {
		if card.Deck != game.BaseDeck {
			t.Fatalf("card %d (%s) had deck %d, want %d", card.ID, card.Name, card.Deck, game.BaseDeck)
		}
	}
}

func TestBaseDeckSuitDistribution(t *testing.T) {
	deck := BaseDeck()
	counts := map[game.Suit]int{}

	for _, card := range deck {
		counts[card.Suit]++
	}

	if counts[game.Bird] != 14 {
		t.Fatalf("expected 14 bird cards, got %d", counts[game.Bird])
	}
	if counts[game.Rabbit] != 13 {
		t.Fatalf("expected 13 rabbit cards, got %d", counts[game.Rabbit])
	}
	if counts[game.Mouse] != 13 {
		t.Fatalf("expected 13 mouse cards, got %d", counts[game.Mouse])
	}
	if counts[game.Fox] != 14 {
		t.Fatalf("expected 14 fox cards, got %d", counts[game.Fox])
	}
}
