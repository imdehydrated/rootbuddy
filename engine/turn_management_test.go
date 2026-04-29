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
			name: "eyrie draw advances to alliance",
			state: game.GameState{
				FactionTurn:  game.Eyrie,
				CurrentPhase: game.Evening,
				CurrentStep:  game.StepEvening,
				TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
				TurnProgress: game.TurnProgress{
					EveningMainActionTaken: true,
				},
			},
			action: game.Action{
				Type: game.ActionEveningDraw,
				EveningDraw: &game.EveningDrawAction{
					Faction: game.Eyrie,
					Count:   1,
				},
			},
			wantFaction: game.Alliance,
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

func TestApplyActionVagabondEveningDrawStaysInEveningForDiscard(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		TurnProgress: game.TurnProgress{
			VagabondRestResolved: true,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionEveningDraw,
		EveningDraw: &game.EveningDrawAction{
			Faction: game.Vagabond,
			Count:   1,
		},
	})

	if next.FactionTurn != game.Vagabond {
		t.Fatalf("expected Vagabond turn to continue after draw, got %v", next.FactionTurn)
	}
	if next.CurrentPhase != game.Evening || next.CurrentStep != game.StepEvening {
		t.Fatalf("expected Vagabond to remain in evening after draw, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
	if !next.TurnProgress.VagabondEveningDrawn {
		t.Fatalf("expected Vagabond evening draw to be marked resolved")
	}
}

func TestApplyActionVagabondItemCapacityAdvancesTurnOrderAndResetsProgress(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		TurnProgress: game.TurnProgress{
			VagabondRestResolved:    true,
			VagabondEveningDrawn:    true,
			VagabondDiscardResolved: true,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionVagabondItemCapacity,
		VagabondCapacity: &game.VagabondItemCapacityAction{
			Faction: game.Vagabond,
		},
	})

	if next.FactionTurn != game.Marquise {
		t.Fatalf("expected Vagabond capacity check to advance to Marquise, got %v", next.FactionTurn)
	}
	if next.CurrentPhase != game.Birdsong || next.CurrentStep != game.StepBirdsong {
		t.Fatalf("expected next turn to begin at birdsong, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
	if !reflect.DeepEqual(next.TurnProgress, game.TurnProgress{}) {
		t.Fatalf("expected turn progress reset, got %+v", next.TurnProgress)
	}
}

func TestApplyActionScoreRoostsStaysInEyrieEveningForDraw(t *testing.T) {
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
		t.Fatalf("expected eyrie score to persist after scoring roosts, got %d", next.VictoryPoints[game.Eyrie])
	}
	if next.FactionTurn != game.Eyrie {
		t.Fatalf("expected eyrie turn to continue for evening draw, got %v", next.FactionTurn)
	}
	if next.CurrentPhase != game.Evening || next.CurrentStep != game.StepEvening {
		t.Fatalf("expected eyrie evening draw step after scoring, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
	if !next.TurnProgress.EveningMainActionTaken {
		t.Fatalf("expected evening scoring to be marked complete before draw")
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

func TestApplyActionPassPhaseDoesNotSkipVagabondSlip(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
	}

	action := game.Action{
		Type: game.ActionPassPhase,
		PassPhase: &game.PassPhaseAction{
			Faction: game.Vagabond,
		},
	}

	next := ApplyAction(state, action)
	if next.CurrentPhase != game.Birdsong || next.CurrentStep != game.StepBirdsong {
		t.Fatalf("expected vagabond to remain in birdsong before slip, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
	if next.TurnProgress.HasSlipped {
		t.Fatalf("expected pass phase not to mark slip resolved")
	}
}

func TestApplyActionStaySlipResolvesVagabondSlip(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Vagabond: game.VagabondState{
			ClearingID: 3,
		},
	}

	action := game.Action{
		Type: game.ActionSlip,
		Slip: &game.SlipAction{
			Faction: game.Vagabond,
			From:    3,
			To:      3,
		},
	}

	next := ApplyAction(state, action)
	if !next.TurnProgress.HasSlipped {
		t.Fatalf("expected stay slip to mark slip resolved")
	}
	if next.Vagabond.ClearingID != 3 || next.Vagabond.InForest {
		t.Fatalf("expected vagabond to remain in clearing 3, got clearing=%d inForest=%v", next.Vagabond.ClearingID, next.Vagabond.InForest)
	}
	if next.CurrentPhase != game.Birdsong || next.CurrentStep != game.StepBirdsong {
		t.Fatalf("expected vagabond to remain in birdsong after resolving slip, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
}
