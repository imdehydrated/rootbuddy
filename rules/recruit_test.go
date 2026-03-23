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
			name: "generates one recruit action during daylight actions when marquise has recruiters and enough supply",
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
				CurrentStep:  game.StepDaylightActions,
				Marquise: game.MarquiseState{
					WarriorSupply: 2,
				},
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
				CurrentStep:  game.StepDaylightActions,
				Marquise: game.MarquiseState{
					WarriorSupply: 1,
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
			name: "no recruit action outside daylight actions",
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
				CurrentStep:  game.StepBirdsong,
				Marquise: game.MarquiseState{
					WarriorSupply: 1,
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
				CurrentStep:  game.StepDaylightActions,
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
			name: "no recruit action when action limit is exhausted",
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
				CurrentStep:  game.StepDaylightActions,
				Marquise: game.MarquiseState{
					WarriorSupply: 1,
				},
				TurnProgress: game.TurnProgress{
					ActionsUsed: 3,
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
			name: "limited supply generates one recruit action per valid recruiter subset",
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
						{
							ID: 3,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Recruiter},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				CurrentStep:  game.StepDaylightActions,
				Marquise: game.MarquiseState{
					WarriorSupply: 2,
				},
			},
			wantActions: []game.Action{
				{
					Type: game.ActionRecruit,
					Recruit: &game.RecruitAction{
						Faction:     game.Marquise,
						ClearingIDs: []int{1, 2},
					},
				},
				{
					Type: game.ActionRecruit,
					Recruit: &game.RecruitAction{
						Faction:     game.Marquise,
						ClearingIDs: []int{1, 3},
					},
				},
				{
					Type: game.ActionRecruit,
					Recruit: &game.RecruitAction{
						Faction:     game.Marquise,
						ClearingIDs: []int{2, 3},
					},
				},
			},
			unwantActions: []game.Action{
				{
					Type: game.ActionRecruit,
					Recruit: &game.RecruitAction{
						Faction:     game.Marquise,
						ClearingIDs: []int{1, 2, 3},
					},
				},
			},
		},
		{
			name: "zero supply produces no recruit actions even with recruiters on the map",
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
				CurrentStep:  game.StepDaylightActions,
				Marquise: game.MarquiseState{
					WarriorSupply: 0,
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
