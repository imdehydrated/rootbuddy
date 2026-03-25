package carddata

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestBaseDeckIncludesCraftedItemsAndEffectIDs(t *testing.T) {
	deck := BaseDeck()

	byID := map[game.CardID]game.Card{}
	for _, card := range deck {
		byID[card.ID] = card
	}

	birdyBindle := byID[8]
	if birdyBindle.CraftedItem == nil || *birdyBindle.CraftedItem != game.ItemBag {
		t.Fatalf("expected Birdy Bindle to craft a bag, got %+v", birdyBindle.CraftedItem)
	}
	if birdyBindle.EffectID != "" {
		t.Fatalf("expected Birdy Bindle to have no effect id, got %q", birdyBindle.EffectID)
	}

	armorers := byID[1]
	if armorers.CraftedItem != nil {
		t.Fatalf("expected Armorers to craft no item, got %+v", armorers.CraftedItem)
	}
	if armorers.EffectID != "armorers" {
		t.Fatalf("expected Armorers effect id to be armorers, got %q", armorers.EffectID)
	}

	favor := byID[50]
	if favor.EffectID != "favor_foxes" {
		t.Fatalf("expected Favor of the Foxes effect id to be favor_foxes, got %q", favor.EffectID)
	}
}
