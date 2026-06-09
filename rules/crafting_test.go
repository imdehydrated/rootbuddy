package rules

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidCraftActions(t *testing.T) {
	tests := []struct {
		name          string
		state         game.GameState
		wantActions   []game.Action
		unwantActions []game.Action
	}{
		{
			name: "craftable card with one matching workshop generates craft action",
			state: game.GameState{
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
				Marquise: game.MarquiseState{
					CardsInHand: []game.Card{
						{
							ID:           100,
							Deck:         game.BaseDeck,
							Name:         "Fox Craft",
							Suit:         game.Fox,
							Kind:         game.ItemCard,
							CraftingCost: game.CraftingCost{Fox: 1},
							VP:           1,
						},
					},
				},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionCraft,
					Craft: &game.CraftAction{
						Faction:               game.Marquise,
						CardID:                100,
						UsedWorkshopClearings: []int{1},
					},
				},
			},
		},
		{
			name: "card requiring two suits uses two matching workshop clearings",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   1,
							Suit: game.Fox,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Workshop},
							},
						},
						{
							ID:   2,
							Suit: game.Rabbit,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Workshop},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					CardsInHand: []game.Card{
						{
							ID:           101,
							Deck:         game.BaseDeck,
							Name:         "Mixed Craft",
							Suit:         game.Mouse,
							Kind:         game.PersistentEffectCard,
							CraftingCost: game.CraftingCost{Fox: 1, Rabbit: 1},
							VP:           2,
						},
					},
				},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionCraft,
					Craft: &game.CraftAction{
						Faction:               game.Marquise,
						CardID:                101,
						UsedWorkshopClearings: []int{1, 2},
					},
				},
			},
		},
		{
			name: "insufficient workshops means no craft action",
			state: game.GameState{
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
				Marquise: game.MarquiseState{
					CardsInHand: []game.Card{
						{
							ID:           102,
							Deck:         game.BaseDeck,
							Name:         "Too Expensive",
							Suit:         game.Mouse,
							Kind:         game.PersistentEffectCard,
							CraftingCost: game.CraftingCost{Fox: 1, Rabbit: 1},
							VP:           2,
						},
					},
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionCraft,
					Craft: &game.CraftAction{
						Faction:               game.Marquise,
						CardID:                102,
						UsedWorkshopClearings: []int{1},
					},
				},
			},
		},
		{
			name: "already used workshop cannot be reused for crafting",
			state: game.GameState{
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
				Marquise: game.MarquiseState{
					CardsInHand: []game.Card{
						{
							ID:           103,
							Deck:         game.BaseDeck,
							Name:         "Blocked Craft",
							Suit:         game.Fox,
							Kind:         game.ItemCard,
							CraftingCost: game.CraftingCost{Fox: 1},
							VP:           1,
						},
					},
				},
				TurnProgress: game.TurnProgress{
					UsedWorkshopClearings: []int{1},
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionCraft,
					Craft: &game.CraftAction{
						Faction:               game.Marquise,
						CardID:                103,
						UsedWorkshopClearings: []int{1},
					},
				},
			},
		},
		{
			name: "non crafted cards do not generate craft actions",
			state: game.GameState{
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
				Marquise: game.MarquiseState{
					CardsInHand: []game.Card{
						{
							ID:           104,
							Deck:         game.BaseDeck,
							Name:         "Fox Ambush",
							Suit:         game.Fox,
							Kind:         game.AmbushCard,
							CraftingCost: game.CraftingCost{},
							VP:           0,
						},
					},
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionCraft,
					Craft: &game.CraftAction{
						Faction:               game.Marquise,
						CardID:                104,
						UsedWorkshopClearings: []int{1},
					},
				},
			},
		},
		{
			name: "wrong turn or phase means no craft actions",
			state: game.GameState{
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
				FactionTurn:  game.Eyrie,
				CurrentPhase: game.Birdsong,
				Marquise: game.MarquiseState{
					CardsInHand: []game.Card{
						{
							ID:           105,
							Deck:         game.BaseDeck,
							Name:         "Fox Craft",
							Suit:         game.Fox,
							Kind:         game.ItemCard,
							CraftingCost: game.CraftingCost{Fox: 1},
							VP:           1,
						},
					},
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionCraft,
					Craft: &game.CraftAction{
						Faction:               game.Marquise,
						CardID:                105,
						UsedWorkshopClearings: []int{1},
					},
				},
			},
		},
		{
			name: "two workshops in one clearing can satisfy repeated same-suit cost",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   1,
							Suit: game.Fox,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Workshop},
								{Faction: game.Marquise, Type: game.Workshop},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					CardsInHand: []game.Card{
						{
							ID:           106,
							Deck:         game.BaseDeck,
							Name:         "Double Fox Craft",
							Suit:         game.Fox,
							Kind:         game.ItemCard,
							CraftingCost: game.CraftingCost{Fox: 2},
							VP:           2,
						},
					},
				},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionCraft,
					Craft: &game.CraftAction{
						Faction:               game.Marquise,
						CardID:                106,
						UsedWorkshopClearings: []int{1, 1},
					},
				},
			},
		},
		{
			name: "any-cost can use any remaining workshop suits",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   1,
							Suit: game.Fox,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Workshop},
							},
						},
						{
							ID:   2,
							Suit: game.Rabbit,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Workshop},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					CardsInHand: []game.Card{
						{
							ID:           107,
							Deck:         game.ExilesAndPartisansDeck,
							Name:         "Any Two Craft",
							Suit:         game.Mouse,
							Kind:         game.OneTimeEffectCard,
							CraftingCost: game.CraftingCost{Any: 2},
							VP:           2,
						},
					},
				},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionCraft,
					Craft: &game.CraftAction{
						Faction:               game.Marquise,
						CardID:                107,
						UsedWorkshopClearings: []int{1, 2},
					},
				},
			},
		},
		{
			name: "mixed exact and any cost uses exact match first then any remaining workshop",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   1,
							Suit: game.Fox,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Workshop},
							},
						},
						{
							ID:   2,
							Suit: game.Mouse,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Workshop},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					CardsInHand: []game.Card{
						{
							ID:           108,
							Deck:         game.ExilesAndPartisansDeck,
							Name:         "Mixed Any Craft",
							Suit:         game.Rabbit,
							Kind:         game.OneTimeEffectCard,
							CraftingCost: game.CraftingCost{Fox: 1, Any: 1},
							VP:           2,
						},
					},
				},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionCraft,
					Craft: &game.CraftAction{
						Faction:               game.Marquise,
						CardID:                108,
						UsedWorkshopClearings: []int{1, 2},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidCraftActions(tt.state)

			for _, want := range tt.wantActions {
				if !containsAction(got, want) {
					t.Fatalf("expected craft action %+v to be generated, but it was missing", want)
				}
			}

			for _, unwant := range tt.unwantActions {
				if containsAction(got, unwant) {
					t.Fatalf("expected craft action %+v to be absent, but it was generated", unwant)
				}
			}
		})
	}
}

func TestValidCraftActionsIncludesFavorVagabondDamageChoices(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   4,
					Suit: game.Mouse,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Workshop},
						{Faction: game.Marquise, Type: game.Workshop},
						{Faction: game.Marquise, Type: game.Workshop},
					},
				},
			},
		},
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{
					ID:           36,
					Name:         "Favor of the Mice",
					Kind:         game.OneTimeEffectCard,
					EffectID:     "favor_mice",
					CraftingCost: game.CraftingCost{Mouse: 3},
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 4,
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
				{Type: game.ItemSword, Status: game.ItemReady},
				{Type: game.ItemTea, Status: game.ItemReady},
			},
		},
	}

	got := ValidCraftActions(state)
	want := game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction:                    game.Marquise,
			CardID:                     36,
			UsedWorkshopClearings:      []int{4, 4, 4},
			DamagedVagabondItemIndexes: []int{0, 1, 2},
		},
	}
	if !containsAction(got, want) {
		t.Fatalf("expected favor craft damage choice %+v, got %+v", want, got)
	}
}

func TestValidCraftActionsEnumeratesDistinctCraftRoutes(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Workshop},
					},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Workshop},
					},
				},
				{
					ID:   3,
					Suit: game.Mouse,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Workshop},
					},
				},
			},
		},
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{
					ID:           900,
					Name:         "Royal Claim",
					Kind:         game.OneTimeEffectCard,
					CraftingCost: game.CraftingCost{Any: 2},
				},
			},
		},
	}

	got := ValidCraftActions(state)
	wantRoutes := [][]int{{1, 2}, {1, 3}, {2, 3}}
	for _, route := range wantRoutes {
		want := game.Action{
			Type: game.ActionCraft,
			Craft: &game.CraftAction{
				Faction:               game.Marquise,
				CardID:                900,
				UsedWorkshopClearings: route,
			},
		}
		if !containsAction(got, want) {
			t.Fatalf("expected craft route %+v in actions %+v", route, got)
		}
	}

	if gotCount := countCraftActionsForCard(got, 900); gotCount != len(wantRoutes) {
		t.Fatalf("expected %d distinct craft routes, got %d actions: %+v", len(wantRoutes), gotCount, got)
	}
}

func TestValidCraftActionsEnumeratesExactSuitAlternatives(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Workshop},
					},
				},
				{
					ID:   4,
					Suit: game.Fox,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Workshop},
					},
				},
			},
		},
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{
					ID:           901,
					Name:         "Fox Craft",
					Kind:         game.ItemCard,
					CraftingCost: game.CraftingCost{Fox: 1},
				},
			},
		},
	}

	got := ValidCraftActions(state)
	for _, route := range [][]int{{1}, {4}} {
		want := game.Action{
			Type: game.ActionCraft,
			Craft: &game.CraftAction{
				Faction:               game.Marquise,
				CardID:                901,
				UsedWorkshopClearings: route,
			},
		}
		if !containsAction(got, want) {
			t.Fatalf("expected exact-suit craft route %+v in actions %+v", route, got)
		}
	}

	if gotCount := countCraftActionsForCard(got, 901); gotCount != 2 {
		t.Fatalf("expected two exact-suit routes, got %d actions: %+v", gotCount, got)
	}
}

func TestBaseFactionCraftActionsEnumerateRoutes(t *testing.T) {
	tests := []struct {
		name     string
		actions  func(game.GameState) []game.Action
		state    game.GameState
		cardID   game.CardID
		faction  game.Faction
		wantUses [][]int
	}{
		{
			name:    "Eyrie roost alternatives",
			actions: ValidEyrieCraftActions,
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{ID: 1, Suit: game.Fox, Buildings: []game.Building{{Faction: game.Eyrie, Type: game.Roost}}},
						{ID: 2, Suit: game.Fox, Buildings: []game.Building{{Faction: game.Eyrie, Type: game.Roost}}},
					},
				},
				FactionTurn:  game.Eyrie,
				CurrentPhase: game.Daylight,
				Eyrie: game.EyrieState{
					CardsInHand: []game.Card{{ID: 902, Name: "Fox Craft", Kind: game.PersistentEffectCard, EffectID: "eyrie_route_test", CraftingCost: game.CraftingCost{Fox: 1}}},
				},
			},
			cardID:   902,
			faction:  game.Eyrie,
			wantUses: [][]int{{1}, {2}},
		},
		{
			name:    "Alliance base any-cost alternatives",
			actions: ValidAllianceCraftActions,
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{ID: 1, Suit: game.Fox, Buildings: []game.Building{{Faction: game.Alliance, Type: game.Base}}},
						{ID: 2, Suit: game.Rabbit, Buildings: []game.Building{{Faction: game.Alliance, Type: game.Base}}},
						{ID: 3, Suit: game.Mouse, Buildings: []game.Building{{Faction: game.Alliance, Type: game.Base}}},
					},
				},
				FactionTurn:  game.Alliance,
				CurrentPhase: game.Daylight,
				Alliance: game.AllianceState{
					CardsInHand: []game.Card{{ID: 903, Name: "Any Craft", Kind: game.OneTimeEffectCard, CraftingCost: game.CraftingCost{Any: 2}}},
				},
			},
			cardID:   903,
			faction:  game.Alliance,
			wantUses: [][]int{{1, 2}, {1, 3}, {2, 3}},
		},
		{
			name:    "Vagabond hammers can satisfy repeated cost",
			actions: ValidVagabondCraftActions,
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{{ID: 1, Suit: game.Rabbit}},
				},
				FactionTurn:  game.Vagabond,
				CurrentPhase: game.Daylight,
				Vagabond: game.VagabondState{
					ClearingID:    1,
					CardsInHand:   []game.Card{{ID: 904, Name: "Double Hammer Craft", Kind: game.OneTimeEffectCard, CraftingCost: game.CraftingCost{Rabbit: 2}}},
					Items:         []game.Item{{Type: game.ItemHammer, Status: game.ItemReady}, {Type: game.ItemHammer, Status: game.ItemReady}},
					Relationships: map[game.Faction]game.RelationshipLevel{},
				},
			},
			cardID:   904,
			faction:  game.Vagabond,
			wantUses: [][]int{{1, 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.actions(tt.state)
			for _, route := range tt.wantUses {
				want := game.Action{
					Type: game.ActionCraft,
					Craft: &game.CraftAction{
						Faction:               tt.faction,
						CardID:                tt.cardID,
						UsedWorkshopClearings: route,
					},
				}
				if !containsAction(got, want) {
					t.Fatalf("expected craft route %+v in actions %+v", route, got)
				}
			}

			if gotCount := countCraftActionsForCard(got, tt.cardID); gotCount != len(tt.wantUses) {
				t.Fatalf("expected %d craft routes, got %d actions: %+v", len(tt.wantUses), gotCount, got)
			}
		})
	}
}

func countCraftActionsForCard(actions []game.Action, cardID game.CardID) int {
	count := 0
	for _, action := range actions {
		if action.Craft != nil && action.Craft.CardID == cardID {
			count++
		}
	}
	return count
}
