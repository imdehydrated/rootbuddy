package rules

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidOverworkActions(t *testing.T) {
	tests := []struct {
		name          string
		state         game.GameState
		wantActions   []game.Action
		unwantActions []game.Action
	}{
		{
			name: "matching suit card can overwork a marquise sawmill",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   1,
							Suit: game.Fox,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Sawmill},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					WoodSupply: 8,
					CardsInHand: []game.Card{
						{ID: 10, Deck: game.BaseDeck, Name: "Fox Card", Suit: game.Fox, Kind: game.ItemCard},
					},
				},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionOverwork,
					Overwork: &game.OverworkAction{
						Faction:    game.Marquise,
						ClearingID: 1,
						CardID:     10,
					},
				},
			},
		},
		{
			name: "bird card can overwork any marquise sawmill",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   2,
							Suit: game.Rabbit,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Sawmill},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					WoodSupply: 8,
					CardsInHand: []game.Card{
						{ID: 20, Deck: game.BaseDeck, Name: "Bird Card", Suit: game.Bird, Kind: game.AmbushCard},
					},
				},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionOverwork,
					Overwork: &game.OverworkAction{
						Faction:    game.Marquise,
						ClearingID: 2,
						CardID:     20,
					},
				},
			},
		},
		{
			name: "no overwork action without a marquise sawmill",
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
					WoodSupply: 8,
					CardsInHand: []game.Card{
						{ID: 10, Deck: game.BaseDeck, Name: "Fox Card", Suit: game.Fox, Kind: game.ItemCard},
					},
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionOverwork,
					Overwork: &game.OverworkAction{
						Faction:    game.Marquise,
						ClearingID: 1,
						CardID:     10,
					},
				},
			},
		},
		{
			name: "non matching non bird card cannot overwork",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   1,
							Suit: game.Mouse,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Sawmill},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					WoodSupply: 8,
					CardsInHand: []game.Card{
						{ID: 11, Deck: game.BaseDeck, Name: "Fox Card", Suit: game.Fox, Kind: game.ItemCard},
					},
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionOverwork,
					Overwork: &game.OverworkAction{
						Faction:    game.Marquise,
						ClearingID: 1,
						CardID:     11,
					},
				},
			},
		},
		{
			name: "no overwork action when marquise wood supply is empty",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   1,
							Suit: game.Fox,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Sawmill},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					CardsInHand: []game.Card{
						{ID: 12, Deck: game.BaseDeck, Name: "Fox Card", Suit: game.Fox, Kind: game.ItemCard},
					},
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionOverwork,
					Overwork: &game.OverworkAction{
						Faction:    game.Marquise,
						ClearingID: 1,
						CardID:     12,
					},
				},
			},
		},
		{
			name: "multiple matching cards generate multiple overwork actions",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   3,
							Suit: game.Rabbit,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Sawmill},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					WoodSupply: 8,
					CardsInHand: []game.Card{
						{ID: 30, Deck: game.BaseDeck, Name: "Rabbit Card", Suit: game.Rabbit, Kind: game.ItemCard},
						{ID: 31, Deck: game.BaseDeck, Name: "Bird Card", Suit: game.Bird, Kind: game.AmbushCard},
					},
				},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionOverwork,
					Overwork: &game.OverworkAction{
						Faction:    game.Marquise,
						ClearingID: 3,
						CardID:     30,
					},
				},
				{
					Type: game.ActionOverwork,
					Overwork: &game.OverworkAction{
						Faction:    game.Marquise,
						ClearingID: 3,
						CardID:     31,
					},
				},
			},
		},
		{
			name: "no overwork action outside daylight or on another factions turn",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   4,
							Suit: game.Fox,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Sawmill},
							},
						},
					},
				},
				FactionTurn:  game.Eyrie,
				CurrentPhase: game.Birdsong,
				Marquise: game.MarquiseState{
					WoodSupply: 8,
					CardsInHand: []game.Card{
						{ID: 40, Deck: game.BaseDeck, Name: "Fox Card", Suit: game.Fox, Kind: game.ItemCard},
					},
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionOverwork,
					Overwork: &game.OverworkAction{
						Faction:    game.Marquise,
						ClearingID: 4,
						CardID:     40,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidOverworkActions(tt.state)

			for _, want := range tt.wantActions {
				if !containsAction(got, want) {
					t.Fatalf("expected overwork action %+v to be generated, but it was missing", want)
				}
			}

			for _, unwant := range tt.unwantActions {
				if containsAction(got, unwant) {
					t.Fatalf("expected overwork action %+v to be absent, but it was generated", unwant)
				}
			}
		})
	}
}
