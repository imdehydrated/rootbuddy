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
						Buildings: []game.Building{},
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
						Buildings: []game.Building{},
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
					},
				},
			},
		},
		{
			name: "cannot build when no open slot remains",
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID:         1,
						BuildSlots: 1,
						Warriors: map[game.Faction]int{
							game.Marquise: 1,
						},
						Buildings: []game.Building{
							{Faction: game.Marquise, Type: game.Sawmill},
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
					},
				},
			},
		},
		{
			name: "ruin occupies a slot and blocks building when it fills the clearing",
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID:         1,
						BuildSlots: 1,
						Ruins:      true,
						Warriors: map[game.Faction]int{
							game.Marquise: 1,
						},
						Buildings: []game.Building{},
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
					},
				},
			},
		},
		{
			name: "can build using enough wood from ruled connected network",
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
						Buildings: []game.Building{},
					},
					{
						ID:         2,
						BuildSlots: 2,
						Adj:        []int{1},
						Warriors: map[game.Faction]int{
							game.Marquise: 1,
						},
						Buildings: []game.Building{},
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
						Buildings: []game.Building{},
					},
					{
						ID:         2,
						BuildSlots: 2,
						Adj:        []int{1},
						Warriors: map[game.Faction]int{
							game.Marquise: 1,
						},
						Buildings: []game.Building{},
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
		{
			name: "cannot build building type with no remaining supply",
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID:         1,
						BuildSlots: 2,
						Warriors: map[game.Faction]int{
							game.Marquise: 1,
						},
						Buildings: []game.Building{},
					},
				},
			},
			marquise: game.MarquiseState{
				SawmillsPlaced:   6,
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
