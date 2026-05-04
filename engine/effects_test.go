package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestApplyActionCraftFavorResolvesAndScores(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Warriors: map[game.Faction]int{
						game.Eyrie: 2,
					},
					Buildings: []game.Building{
						{Faction: game.Eyrie, Type: game.Roost},
					},
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Warriors: map[game.Faction]int{
						game.Eyrie: 1,
					},
				},
			},
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 50, Name: "Favor of the Foxes"},
			},
		},
		Eyrie: game.EyrieState{
			RoostsPlaced: 1,
		},
		Alliance: game.AllianceState{
			SympathyPlaced: 1,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction: game.Marquise,
			CardID:  50,
		},
	})

	if next.Map.Clearings[0].Warriors[game.Eyrie] != 0 {
		t.Fatalf("expected fox clearing warriors to be removed, got %+v", next.Map.Clearings[0].Warriors)
	}
	if len(next.Map.Clearings[0].Buildings) != 0 {
		t.Fatalf("expected fox clearing buildings to be removed, got %+v", next.Map.Clearings[0].Buildings)
	}
	if len(next.Map.Clearings[0].Tokens) != 0 {
		t.Fatalf("expected fox clearing tokens to be removed, got %+v", next.Map.Clearings[0].Tokens)
	}
	if next.Map.Clearings[1].Warriors[game.Eyrie] != 1 {
		t.Fatalf("expected non-fox clearing to stay unchanged, got %+v", next.Map.Clearings[1].Warriors)
	}
	if next.Eyrie.RoostsPlaced != 0 {
		t.Fatalf("expected removed roost to update placed count, got %d", next.Eyrie.RoostsPlaced)
	}
	if next.Alliance.SympathyPlaced != 0 {
		t.Fatalf("expected removed sympathy to update placed count, got %d", next.Alliance.SympathyPlaced)
	}
	if next.VictoryPoints[game.Marquise] != 2 {
		t.Fatalf("expected 2 VP for removed building and token, got %+v", next.VictoryPoints)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 50 {
		t.Fatalf("expected favor card to discard after resolving, got %+v", next.DiscardPile)
	}
}

func TestApplyActionCraftFavorDamagesVagabondInAffectedClearing(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   4,
					Suit: game.Mouse,
				},
			},
		},
		PlayerFaction: game.Vagabond,
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{
				{ID: 36, Name: "Favor of the Mice"},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 4,
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
				{Type: game.ItemTea, Status: game.ItemReady},
				{Type: game.ItemSword, Status: game.ItemReady},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction: game.Eyrie,
			CardID:  36,
		},
	})

	damaged := 0
	for _, item := range next.Vagabond.Items {
		if item.Status == game.ItemDamaged {
			damaged++
		}
	}
	if damaged != 3 {
		t.Fatalf("expected favor to damage 3 Vagabond items, got %+v", next.Vagabond.Items)
	}
}

func TestApplyActionCraftFavorDoesNotHitCoalitionPartnerVagabond(t *testing.T) {
	state := game.GameState{
		CoalitionActive:  true,
		CoalitionPartner: game.Eyrie,
		ActiveDominance: map[game.Faction]game.CardID{
			game.Vagabond: 14,
		},
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   4,
					Suit: game.Mouse,
				},
			},
		},
		PlayerFaction: game.Vagabond,
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{
				{ID: 36, Name: "Favor of the Mice"},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 4,
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
				{Type: game.ItemTea, Status: game.ItemReady},
				{Type: game.ItemSword, Status: game.ItemReady},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction: game.Eyrie,
			CardID:  36,
		},
	})

	damaged := 0
	for _, item := range next.Vagabond.Items {
		if item.Status == game.ItemDamaged {
			damaged++
		}
	}
	if damaged != 0 {
		t.Fatalf("expected coalition partner favor to leave Vagabond untouched, got %+v", next.Vagabond.Items)
	}
}

func TestApplyActionDiscardEffectMovesPersistentCardToDiscard(t *testing.T) {
	state := game.GameState{
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {15},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionDiscardEffect,
		DiscardEffect: &game.DiscardEffectAction{
			Faction: game.Marquise,
			CardID:  15,
		},
	})

	if len(next.PersistentEffects) != 0 {
		t.Fatalf("expected persistent effect to be removed from play, got %+v", next.PersistentEffects)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 15 {
		t.Fatalf("expected discarded effect to move to discard pile, got %+v", next.DiscardPile)
	}
}

func TestApplyActionCraftScoresItemCardVictoryPoints(t *testing.T) {
	state := game.GameState{
		ItemSupply: map[game.ItemType]int{
			game.ItemCoin: 1,
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 21, Name: "Bake Sale"},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction: game.Marquise,
			CardID:  21,
		},
	})

	if next.VictoryPoints[game.Marquise] != 3 {
		t.Fatalf("expected item craft to score printed VP, got %+v", next.VictoryPoints)
	}
	if next.ItemSupply[game.ItemCoin] != 0 {
		t.Fatalf("expected crafted item to deduct shared supply, got %+v", next.ItemSupply)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 21 {
		t.Fatalf("expected item craft to discard card after scoring, got %+v", next.DiscardPile)
	}
}

func TestApplyActionCraftEyrieDisdainScoresOnePointForItems(t *testing.T) {
	state := game.GameState{
		ItemSupply: map[game.ItemType]int{
			game.ItemCoin: 1,
		},
		Eyrie: game.EyrieState{
			Leader: game.LeaderCommander,
			CardsInHand: []game.Card{
				{ID: 21, Name: "Bake Sale"},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction: game.Eyrie,
			CardID:  21,
		},
	})

	if next.VictoryPoints[game.Eyrie] != 1 {
		t.Fatalf("expected Eyrie Disdain for Trade to score 1 VP, got %+v", next.VictoryPoints)
	}
}

func TestApplyActionCraftEyrieBuilderScoresPrintedItemPoints(t *testing.T) {
	state := game.GameState{
		ItemSupply: map[game.ItemType]int{
			game.ItemCoin: 1,
		},
		Eyrie: game.EyrieState{
			Leader: game.LeaderBuilder,
			CardsInHand: []game.Card{
				{ID: 21, Name: "Bake Sale"},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction: game.Eyrie,
			CardID:  21,
		},
	})

	if next.VictoryPoints[game.Eyrie] != 3 {
		t.Fatalf("expected Builder to ignore Disdain and score printed VP, got %+v", next.VictoryPoints)
	}
}
