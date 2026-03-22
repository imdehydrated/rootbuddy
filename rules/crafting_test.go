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
