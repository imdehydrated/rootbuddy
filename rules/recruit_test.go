package rules

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidRecruitActions(t *testing.T) {
	tests := []struct {
		name          string
		state         game.GameState
		wantActions   []game.Action
		unwantActions []game.Action
	}{
		{
			name: "generates one recruit action when marquise has recruiters and enough supply",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID: 1,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Recruiter},
							},
						},
						{
							ID: 2,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Recruiter},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					WarriorSupply: 2,
				},
				TurnProgress: game.TurnProgress{},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionRecruit,
					Recruit: &game.RecruitAction{
						Faction:     game.Marquise,
						ClearingIDs: []int{1, 2},
					},
				},
			},
		},
		{
			name: "no recruit action on another factions turn",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID: 1,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Recruiter},
							},
						},
					},
				},
				FactionTurn:  game.Eyrie,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					WarriorSupply: 1,
				},
				TurnProgress: game.TurnProgress{},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionRecruit,
					Recruit: &game.RecruitAction{
						Faction:     game.Marquise,
						ClearingIDs: []int{1},
					},
				},
			},
		},
		{
			name: "no recruit action outside daylight",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID: 1,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Recruiter},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Birdsong,
				Marquise: game.MarquiseState{
					WarriorSupply: 1,
				},
				TurnProgress: game.TurnProgress{},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionRecruit,
					Recruit: &game.RecruitAction{
						Faction:     game.Marquise,
						ClearingIDs: []int{1},
					},
				},
			},
		},
		{
			name: "no recruit action after recruit already used this turn",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID: 1,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Recruiter},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					WarriorSupply: 1,
				},
				TurnProgress: game.TurnProgress{
					RecruitUsed: true,
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionRecruit,
					Recruit: &game.RecruitAction{
						Faction:     game.Marquise,
						ClearingIDs: []int{1},
					},
				},
			},
		},
		{
			name: "no recruit action when no recruiters are on the board",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:        1,
							Buildings: []game.Building{},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					WarriorSupply: 1,
				},
				TurnProgress: game.TurnProgress{},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionRecruit,
					Recruit: &game.RecruitAction{
						Faction:     game.Marquise,
						ClearingIDs: []int{1},
					},
				},
			},
		},
		{
			name: "no recruit action when supply is less than recruiter count",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID: 1,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Recruiter},
							},
						},
						{
							ID: 2,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Recruiter},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				Marquise: game.MarquiseState{
					WarriorSupply: 1,
				},
				TurnProgress: game.TurnProgress{},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionRecruit,
					Recruit: &game.RecruitAction{
						Faction:     game.Marquise,
						ClearingIDs: []int{1, 2},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidRecruitActions(tt.state)

			for _, want := range tt.wantActions {
				if !containsAction(got, want) {
					t.Fatalf("expected recruit action %+v to be generated, but it was missing", want)
				}
			}

			for _, unwant := range tt.unwantActions {
				if containsAction(got, unwant) {
					t.Fatalf("expected recruit action %+v to be absent, but it was generated", unwant)
				}
			}
		})
	}
}
