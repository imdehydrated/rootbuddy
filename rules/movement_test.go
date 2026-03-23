package rules

import (
	"reflect"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func containsAction(actions []game.Action, want game.Action) bool {
	for _, action := range actions {
		if reflect.DeepEqual(action, want) {
			return true
		}
	}
	return false
}

func TestValidMovementActions(t *testing.T) {
	tests := []struct {
		name          string
		faction       game.Faction
		board         game.Map
		wantActions   []game.Action
		unwantActions []game.Action
	}{
		{
			name:    "generates one action per selectable warrior count when origin is ruled",
			faction: game.Marquise,
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID:  1,
						Adj: []int{2},
						Warriors: map[game.Faction]int{
							game.Marquise: 3,
							game.WoodlandAlliance: 1,
						},
					},
					{
						ID:       2,
						Adj:      []int{1},
						Warriors: map[game.Faction]int{},
					},
				},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionMovement,
					Movement: &game.MovementAction{
						Faction:  game.Marquise,
						Count:    1,
						MaxCount: 3,
						From:     1,
						To:       2,
					},
				},
				{
					Type: game.ActionMovement,
					Movement: &game.MovementAction{
						Faction:  game.Marquise,
						Count:    3,
						MaxCount: 3,
						From:     1,
						To:       2,
					},
				},
			},
		},
		{
			name:    "generates movement when destination is ruled",
			faction: game.Marquise,
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID:  1,
						Adj: []int{2},
						Warriors: map[game.Faction]int{
							game.Marquise:         2,
							game.WoodlandAlliance: 3,
						},
					},
					{
						ID:  2,
						Adj: []int{1},
						Warriors: map[game.Faction]int{
							game.Marquise: 1,
						},
						Buildings: []game.Building{
							{Faction: game.Marquise, Type: game.Sawmill},
						},
					},
				},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionMovement,
					Movement: &game.MovementAction{
						Faction:  game.Marquise,
						Count:    2,
						MaxCount: 2,
						From:     1,
						To:       2,
					},
				},
			},
		},
		{
			name:    "generates no action when neither origin nor destination is ruled",
			faction: game.Marquise,
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID:  1,
						Adj: []int{2},
						Warriors: map[game.Faction]int{
							game.Marquise:         1,
							game.WoodlandAlliance: 2,
						},
					},
					{
						ID:  2,
						Adj: []int{1},
						Warriors: map[game.Faction]int{
							game.Eyrie: 1,
						},
					},
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionMovement,
					Movement: &game.MovementAction{
						Faction:  game.Marquise,
						Count:    1,
						MaxCount: 1,
						From:     1,
						To:       2,
					},
				},
			},
		},
		{
			name:    "no warriors at origin means no movement action",
			faction: game.Marquise,
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID:  1,
						Adj: []int{2},
						Warriors: map[game.Faction]int{
							game.Marquise: 0,
						},
						Buildings: []game.Building{
							{Faction: game.Marquise, Type: game.Sawmill},
						},
					},
					{
						ID:       2,
						Adj:      []int{1},
						Warriors: map[game.Faction]int{},
					},
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionMovement,
					Movement: &game.MovementAction{
						Faction:  game.Marquise,
						Count:    1,
						MaxCount: 1,
						From:     1,
						To:       2,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidMovementActions(tt.faction, tt.board)

			for _, want := range tt.wantActions {
				if !containsAction(got, want) {
					t.Fatalf("expected action %+v to be generated, but it was missing", want)
				}
			}

			for _, unwant := range tt.unwantActions {
				if containsAction(got, unwant) {
					t.Fatalf("expected action %+v to be absent, but it was generated", unwant)
				}
			}
		})
	}
}
