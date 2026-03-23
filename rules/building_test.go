package rules

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidBuilds(t *testing.T) {
	tests := []struct {
		name          string
		board         game.Map
		marquise      game.MarquiseState
		wantActions   []game.Action
		unwantActions []game.Action
	}{
		{
			name: "first sawmill can be built in ruled clearing with open slot",
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID:         1,
						BuildSlots: 2,
						Warriors: map[game.Faction]int{
							game.Marquise: 1,
						},
					},
				},
			},
			marquise: game.MarquiseState{
				SawmillsPlaced:   0,
				WorkshopsPlaced:  6,
				RecruitersPlaced: 6,
			},
			wantActions: []game.Action{
				{
					Type: game.ActionBuild,
					Build: &game.BuildAction{
						Faction:      game.Marquise,
						ClearingID:   1,
						BuildingType: game.Sawmill,
						WoodSources:  []game.WoodSource{},
					},
				},
			},
		},
		{
			name: "cannot build in clearing the marquise does not rule",
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID:         1,
						BuildSlots: 2,
						Warriors: map[game.Faction]int{
							game.Marquise:         1,
							game.WoodlandAlliance: 2,
						},
					},
				},
			},
			marquise: game.MarquiseState{
				SawmillsPlaced:   0,
				WorkshopsPlaced:  6,
				RecruitersPlaced: 6,
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionBuild,
					Build: &game.BuildAction{
						Faction:      game.Marquise,
						ClearingID:   1,
						BuildingType: game.Sawmill,
						WoodSources:  []game.WoodSource{},
					},
				},
			},
		},
		{
			name: "can build using enough wood from ruled connected network with explicit wood sources",
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID:         1,
						BuildSlots: 2,
						Adj:        []int{2},
						Wood:       2,
						Warriors: map[game.Faction]int{
							game.Marquise: 2,
						},
					},
					{
						ID:         2,
						BuildSlots: 2,
						Adj:        []int{1},
						Warriors: map[game.Faction]int{
							game.Marquise: 1,
						},
					},
				},
			},
			marquise: game.MarquiseState{
				SawmillsPlaced:   2,
				WorkshopsPlaced:  6,
				RecruitersPlaced: 6,
			},
			wantActions: []game.Action{
				{
					Type: game.ActionBuild,
					Build: &game.BuildAction{
						Faction:      game.Marquise,
						ClearingID:   2,
						BuildingType: game.Sawmill,
						WoodSources: []game.WoodSource{
							{ClearingID: 1, Amount: 2},
						},
					},
				},
			},
		},
		{
			name: "cannot build when ruled wood network has insufficient wood",
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID:         1,
						BuildSlots: 2,
						Adj:        []int{2},
						Wood:       1,
						Warriors: map[game.Faction]int{
							game.Marquise: 2,
						},
					},
					{
						ID:         2,
						BuildSlots: 2,
						Adj:        []int{1},
						Warriors: map[game.Faction]int{
							game.Marquise: 1,
						},
					},
				},
			},
			marquise: game.MarquiseState{
				SawmillsPlaced:   2,
				WorkshopsPlaced:  6,
				RecruitersPlaced: 6,
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionBuild,
					Build: &game.BuildAction{
						Faction:      game.Marquise,
						ClearingID:   2,
						BuildingType: game.Sawmill,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidBuilds(tt.board, tt.marquise)

			for _, want := range tt.wantActions {
				if !containsAction(got, want) {
					t.Fatalf("expected build action %+v to be generated, but it was missing", want)
				}
			}

			for _, unwant := range tt.unwantActions {
				if containsAction(got, unwant) {
					t.Fatalf("expected build action %+v to be absent, but it was generated", unwant)
				}
			}
		})
	}
}
