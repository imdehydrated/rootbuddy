package engine

import (
	"reflect"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestApplyActionEveningDrawAdvancesTurnOrderAndResetsProgress(t *testing.T) {
	tests := []struct {
		name        string
		state       game.GameState
		action      game.Action
		wantFaction game.Faction
	}{
		{
			name: "marquise draw advances to eyrie",
			state: game.GameState{
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Evening,
				CurrentStep:  game.StepEvening,
				TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
				TurnProgress: game.TurnProgress{
					ActionsUsed:           3,
					MarchesUsed:           2,
					RecruitUsed:           true,
					UsedWorkshopClearings: []int{1, 2},
					HasCrafted:            true,
				},
			},
			action: game.Action{
				Type: game.ActionEveningDraw,
				EveningDraw: &game.EveningDrawAction{
					Faction: game.Marquise,
					Count:   1,
				},
			},
			wantFaction: game.Eyrie,
		},
		{
			name: "alliance draw advances to vagabond",
			state: game.GameState{
				FactionTurn:  game.Alliance,
				CurrentPhase: game.Evening,
				CurrentStep:  game.StepEvening,
				TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
				TurnProgress: game.TurnProgress{
					OfficerActionsUsed: 1,
					HasOrganized:       true,
				},
			},
			action: game.Action{
				Type: game.ActionEveningDraw,
				EveningDraw: &game.EveningDrawAction{
					Faction: game.Alliance,
					Count:   2,
				},
			},
			wantFaction: game.Vagabond,
		},
		{
			name: "vagabond draw wraps to marquise",
			state: game.GameState{
				FactionTurn:  game.Vagabond,
				CurrentPhase: game.Evening,
				CurrentStep:  game.StepEvening,
				TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
				TurnProgress: game.TurnProgress{
					HasSlipped: true,
				},
				Vagabond: game.VagabondState{
					InForest: true,
					Items: []game.Item{
						{Type: game.ItemSword, Status: game.ItemDamaged},
					},
				},
			},
			action: game.Action{
				Type: game.ActionEveningDraw,
				EveningDraw: &game.EveningDrawAction{
					Faction: game.Vagabond,
					Count:   0,
				},
			},
			wantFaction: game.Marquise,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := ApplyAction(tt.state, tt.action)

			if next.FactionTurn != tt.wantFaction {
				t.Fatalf("expected next faction %v, got %v", tt.wantFaction, next.FactionTurn)
			}
			if next.CurrentPhase != game.Birdsong || next.CurrentStep != game.StepBirdsong {
				t.Fatalf("expected next turn to begin at birdsong, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
			}
			if !reflect.DeepEqual(next.TurnProgress, game.TurnProgress{}) {
				t.Fatalf("expected turn progress reset, got %+v", next.TurnProgress)
			}
			if !reflect.DeepEqual(next.TurnOrder, tt.state.TurnOrder) {
				t.Fatalf("expected turn order to be preserved, got %+v", next.TurnOrder)
			}
		})
	}
}

func TestApplyActionScoreRoostsAdvancesToAllianceAndKeepsScore(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		VictoryPoints: map[game.Faction]int{
			game.Eyrie: 1,
		},
		TurnProgress: game.TurnProgress{
			DecreeColumnsResolved: 4,
			ResolvedDecreeCardIDs: []game.CardID{9, 10},
		},
	}

	action := game.Action{
		Type: game.ActionScoreRoosts,
		ScoreRoosts: &game.ScoreRoostsAction{
			Faction: game.Eyrie,
			Points:  3,
		},
	}

	next := ApplyAction(state, action)

	if next.VictoryPoints[game.Eyrie] != 4 {
		t.Fatalf("expected eyrie score to persist after turn advance, got %d", next.VictoryPoints[game.Eyrie])
	}
	if next.FactionTurn != game.Alliance {
		t.Fatalf("expected alliance turn after eyrie evening, got %v", next.FactionTurn)
	}
	if next.CurrentPhase != game.Birdsong || next.CurrentStep != game.StepBirdsong {
		t.Fatalf("expected alliance birdsong after eyrie evening, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
	if !reflect.DeepEqual(next.TurnProgress, game.TurnProgress{}) {
		t.Fatalf("expected turn progress reset after eyrie evening, got %+v", next.TurnProgress)
	}
}

func TestApplyActionEveningDrawUsesDefaultTurnOrderWhenUnset(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
	}

	action := game.Action{
		Type: game.ActionEveningDraw,
		EveningDraw: &game.EveningDrawAction{
			Faction: game.Marquise,
			Count:   1,
		},
	}

	next := ApplyAction(state, action)

	if next.FactionTurn != game.Eyrie {
		t.Fatalf("expected default turn order to advance marquise to eyrie, got %v", next.FactionTurn)
	}
	if !reflect.DeepEqual(next.TurnOrder, defaultTurnOrder) {
		t.Fatalf("expected default turn order to be stored, got %+v", next.TurnOrder)
	}
}
