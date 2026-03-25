package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestInitialItemSupplyMatchesBaseGameCounts(t *testing.T) {
	supply := InitialItemSupply()

	want := map[game.ItemType]int{
		game.ItemTea:      2,
		game.ItemCoin:     2,
		game.ItemCrossbow: 1,
		game.ItemHammer:   1,
		game.ItemSword:    2,
		game.ItemBoots:    2,
		game.ItemBag:      2,
	}

	for itemType, count := range want {
		if supply[itemType] != count {
			t.Fatalf("expected %v supply count %d, got %d", itemType, count, supply[itemType])
		}
	}

	if _, ok := supply[game.ItemTorch]; ok {
		t.Fatalf("expected torch to be excluded from the shared craftable supply, got %+v", supply)
	}
}

func TestRuinItemsMatchBaseGameSet(t *testing.T) {
	got := RuinItems()
	want := []game.ItemType{
		game.ItemSword,
		game.ItemHammer,
		game.ItemBag,
		game.ItemCoin,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d ruin items, got %+v", len(want), got)
	}
	for index, itemType := range want {
		if got[index] != itemType {
			t.Fatalf("expected ruin item %d to be %v, got %+v", index, itemType, got)
		}
	}
}

func TestVagabondStartingItemsAreSeparateFromSharedSupply(t *testing.T) {
	tests := []struct {
		name      string
		character game.VagabondCharacter
		want      []game.ItemType
	}{
		{
			name:      "thief",
			character: game.CharThief,
			want:      []game.ItemType{game.ItemBoots, game.ItemTorch, game.ItemTea, game.ItemSword},
		},
		{
			name:      "tinker",
			character: game.CharTinker,
			want:      []game.ItemType{game.ItemBoots, game.ItemTorch, game.ItemBag, game.ItemHammer},
		},
		{
			name:      "ranger",
			character: game.CharRanger,
			want:      []game.ItemType{game.ItemBoots, game.ItemTorch, game.ItemCrossbow, game.ItemSword},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VagabondStartingItems(tt.character)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d starting items, got %+v", len(tt.want), got)
			}
			for index, itemType := range tt.want {
				if got[index].Type != itemType || got[index].Status != game.ItemReady {
					t.Fatalf("expected starting item %d to be ready %v, got %+v", index, itemType, got)
				}
			}
		})
	}
}

func TestApplyCraftDeductsItemSupplyAndAddsVagabondItem(t *testing.T) {
	hammer := game.ItemHammer
	sword := game.ItemSword
	state := game.GameState{
		ItemSupply: map[game.ItemType]int{
			game.ItemSword: 1,
		},
		Vagabond: game.VagabondState{
			CardsInHand: []game.Card{
				{
					ID:          33,
					CraftedItem: &sword,
				},
			},
			Items: []game.Item{
				{Type: hammer, Status: game.ItemReady},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction:               game.Vagabond,
			CardID:                33,
			UsedWorkshopClearings: []int{1},
		},
	})

	if next.ItemSupply[game.ItemSword] != 0 {
		t.Fatalf("expected sword supply to be deducted to 0, got %+v", next.ItemSupply)
	}
	if len(next.Vagabond.Items) != 2 || next.Vagabond.Items[1].Type != game.ItemSword {
		t.Fatalf("expected crafted sword to be added to Vagabond items, got %+v", next.Vagabond.Items)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 33 {
		t.Fatalf("expected crafted card to be discarded, got %+v", next.DiscardPile)
	}
}
