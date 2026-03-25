package rules

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func testItemCard(id game.CardID, craftedItem game.ItemType, cost game.CraftingCost) game.Card {
	item := craftedItem
	return game.Card{
		ID:           id,
		Kind:         game.ItemCard,
		CraftingCost: cost,
		CraftedItem:  &item,
	}
}

func TestValidCraftActionsSkipsUnavailableItemSupply(t *testing.T) {
	actions := ValidCraftActions(game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Workshop},
					},
				},
			},
		},
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		ItemSupply: map[game.ItemType]int{
			game.ItemSword: 0,
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				testItemCard(900, game.ItemSword, game.CraftingCost{Fox: 1}),
			},
		},
	})

	if len(actions) != 0 {
		t.Fatalf("expected no Marquise craft actions when item supply is exhausted, got %+v", actions)
	}
}

func TestValidEyrieCraftActionsSkipsUnavailableItemSupply(t *testing.T) {
	actions := ValidEyrieCraftActions(game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Buildings: []game.Building{
						{Faction: game.Eyrie, Type: game.Roost},
					},
				},
			},
		},
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Daylight,
		ItemSupply: map[game.ItemType]int{
			game.ItemSword: 0,
		},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{
				testItemCard(901, game.ItemSword, game.CraftingCost{Fox: 1}),
			},
		},
	})

	if len(actions) != 0 {
		t.Fatalf("expected no Eyrie craft actions when item supply is exhausted, got %+v", actions)
	}
}

func TestValidAllianceCraftActionsSkipsUnavailableItemSupply(t *testing.T) {
	actions := ValidAllianceCraftActions(game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Buildings: []game.Building{
						{Faction: game.Alliance, Type: game.Base},
					},
				},
			},
		},
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Daylight,
		ItemSupply: map[game.ItemType]int{
			game.ItemSword: 0,
		},
		Alliance: game.AllianceState{
			CardsInHand: []game.Card{
				testItemCard(902, game.ItemSword, game.CraftingCost{Fox: 1}),
			},
			FoxBasePlaced: true,
		},
	})

	if len(actions) != 0 {
		t.Fatalf("expected no Alliance craft actions when item supply is exhausted, got %+v", actions)
	}
}

func TestValidVagabondCraftActionsSkipsUnavailableItemSupply(t *testing.T) {
	actions := ValidVagabondCraftActions(game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
				},
			},
		},
		ItemSupply: map[game.ItemType]int{
			game.ItemSword: 0,
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemHammer, Status: game.ItemReady},
			},
			CardsInHand: []game.Card{
				testItemCard(903, game.ItemSword, game.CraftingCost{Fox: 1}),
			},
		},
	})

	if len(actions) != 0 {
		t.Fatalf("expected no Vagabond craft actions when item supply is exhausted, got %+v", actions)
	}
}
