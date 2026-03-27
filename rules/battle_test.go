package rules

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidBattles(t *testing.T) {
	tests := []struct {
		name          string
		faction       game.Faction
		board         game.Map
		wantBattles   []game.Action
		unwantBattles []game.Action
	}{
		{
			name:    "attacker with warriors can battle one enemy faction with warriors",
			faction: game.Marquise,
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID: 1,
						Warriors: map[game.Faction]int{
							game.Marquise: 2,
							game.Eyrie:    1,
						},
						Buildings: []game.Building{},
					},
				},
			},
			wantBattles: []game.Action{
				{
					Type: game.ActionBattle,
					Battle: &game.BattleAction{
						Faction:       game.Marquise,
						ClearingID:    1,
						TargetFaction: game.Eyrie,
					},
				},
			},
		},
		{
			name:    "attacker with warriors can battle multiple enemy factions in same clearing",
			faction: game.Marquise,
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID: 1,
						Warriors: map[game.Faction]int{
							game.Marquise:         2,
							game.Eyrie:            1,
							game.WoodlandAlliance: 1,
						},
						Buildings: []game.Building{},
					},
				},
			},
			wantBattles: []game.Action{
				{
					Type: game.ActionBattle,
					Battle: &game.BattleAction{
						Faction:       game.Marquise,
						ClearingID:    1,
						TargetFaction: game.Eyrie,
					},
				},
				{
					Type: game.ActionBattle,
					Battle: &game.BattleAction{
						Faction:       game.Marquise,
						ClearingID:    1,
						TargetFaction: game.WoodlandAlliance,
					},
				},
			},
		},
		{
			name:    "attacker with no warriors cannot battle",
			faction: game.Marquise,
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID: 1,
						Warriors: map[game.Faction]int{
							game.Eyrie: 1,
						},
						Buildings: []game.Building{
							{Faction: game.WoodlandAlliance, Type: game.Sawmill},
						},
					},
				},
			},
			unwantBattles: []game.Action{
				{
					Type: game.ActionBattle,
					Battle: &game.BattleAction{
						Faction:       game.Marquise,
						ClearingID:    1,
						TargetFaction: game.Eyrie,
					},
				},
				{
					Type: game.ActionBattle,
					Battle: &game.BattleAction{
						Faction:       game.Marquise,
						ClearingID:    1,
						TargetFaction: game.WoodlandAlliance,
					},
				},
			},
		},
		{
			name:    "enemy with only buildings can still be battled",
			faction: game.Marquise,
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID: 1,
						Warriors: map[game.Faction]int{
							game.Marquise: 1,
						},
						Buildings: []game.Building{
							{Faction: game.Eyrie, Type: game.Sawmill},
							{Faction: game.Eyrie, Type: game.Workshop},
						},
					},
				},
			},
			wantBattles: []game.Action{
				{
					Type: game.ActionBattle,
					Battle: &game.BattleAction{
						Faction:       game.Marquise,
						ClearingID:    1,
						TargetFaction: game.Eyrie,
					},
				},
			},
		},
		{
			name:    "clearing with only attacker pieces generates no battles",
			faction: game.Marquise,
			board: game.Map{
				Clearings: []game.Clearing{
					{
						ID: 1,
						Warriors: map[game.Faction]int{
							game.Marquise: 2,
						},
						Buildings: []game.Building{
							{Faction: game.Marquise, Type: game.Sawmill},
						},
					},
				},
			},
			unwantBattles: []game.Action{
				{
					Type: game.ActionBattle,
					Battle: &game.BattleAction{
						Faction:       game.Marquise,
						ClearingID:    1,
						TargetFaction: game.Eyrie,
					},
				},
				{
					Type: game.ActionBattle,
					Battle: &game.BattleAction{
						Faction:       game.Marquise,
						ClearingID:    1,
						TargetFaction: game.WoodlandAlliance,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidBattles(tt.faction, tt.board)

			for _, want := range tt.wantBattles {
				if !containsAction(got, want) {
					t.Fatalf("expected battle %+v to be generated, but it was missing", want)
				}
			}

			for _, unwant := range tt.unwantBattles {
				if containsAction(got, unwant) {
					t.Fatalf("expected battle %+v to be absent, but it was generated", unwant)
				}
			}
		})
	}
}

func TestValidBattlesInStateSkipsCoalitionPartnerTargets(t *testing.T) {
	state := game.GameState{
		CoalitionActive:  true,
		CoalitionPartner: game.Marquise,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
		},
	}

	got := ValidBattlesInState(game.Marquise, state)
	unwant := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Vagabond,
		},
	}

	if containsAction(got, unwant) {
		t.Fatalf("did not expect battle against coalition Vagabond partner, got %+v", got)
	}
}
