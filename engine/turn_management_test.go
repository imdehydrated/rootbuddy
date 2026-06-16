package engine

import (
	"reflect"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestApplyActionEveningDrawStaysInEveningForDiscard(t *testing.T) {
	tests := []struct {
		name   string
		state  game.GameState
		action game.Action
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := ApplyAction(tt.state, tt.action)

			if next.FactionTurn != tt.state.FactionTurn {
				t.Fatalf("expected faction %v to remain active for discard, got %v", tt.state.FactionTurn, next.FactionTurn)
			}
			if next.CurrentPhase != game.Evening || next.CurrentStep != game.StepEvening {
				t.Fatalf("expected turn to remain in evening for discard, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
			}
			if !next.TurnProgress.EveningDrawn {
				t.Fatalf("expected evening draw to be marked before discard")
			}
			if !reflect.DeepEqual(next.TurnOrder, tt.state.TurnOrder) {
				t.Fatalf("expected turn order to be preserved, got %+v", next.TurnOrder)
			}
		})
	}
}

func TestApplyActionEveningDiscardAdvancesTurnOrderAndResetsProgress(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		RoundNumber:  2,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		TurnProgress: game.TurnProgress{
			EveningDrawn: true,
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 8},
				{ID: 9},
				{ID: 10},
				{ID: 11},
				{ID: 12},
				{ID: 13},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionEveningDiscard,
		EveningDiscard: &game.EveningDiscardAction{
			Faction: game.Marquise,
			CardIDs: []game.CardID{
				8,
			},
			Count: 1,
		},
	})

	if next.FactionTurn != game.Eyrie {
		t.Fatalf("expected discard to advance to Eyrie, got %v", next.FactionTurn)
	}
	if next.RoundNumber != 2 {
		t.Fatalf("expected mid-round discard to keep round 2, got %d", next.RoundNumber)
	}
	if next.CurrentPhase != game.Birdsong || next.CurrentStep != game.StepBirdsong {
		t.Fatalf("expected next turn to begin at birdsong, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
	if !reflect.DeepEqual(next.TurnProgress, game.TurnProgress{}) {
		t.Fatalf("expected turn progress reset, got %+v", next.TurnProgress)
	}
	if hasCard(next.Marquise.CardsInHand, 8) {
		t.Fatalf("expected discarded card to leave Marquise hand, got %+v", next.Marquise.CardsInHand)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 8 {
		t.Fatalf("expected discarded card in discard pile, got %+v", next.DiscardPile)
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
		RoundNumber:  2,
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
	if next.RoundNumber != 3 {
		t.Fatalf("expected Vagabond capacity check to advance to round 3, got %d", next.RoundNumber)
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

func TestApplyActionEveningDiscardUsesDefaultTurnOrderWhenUnset(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		RoundNumber:  4,
		TurnProgress: game.TurnProgress{
			EveningDrawn: true,
		},
	}

	action := game.Action{
		Type: game.ActionEveningDiscard,
		EveningDiscard: &game.EveningDiscardAction{
			Faction: game.Marquise,
		},
	}

	next := ApplyAction(state, action)

	if next.FactionTurn != game.Eyrie {
		t.Fatalf("expected default turn order to advance marquise to eyrie, got %v", next.FactionTurn)
	}
	if next.RoundNumber != 4 {
		t.Fatalf("expected default mid-round advance to keep round 4, got %d", next.RoundNumber)
	}
	if !reflect.DeepEqual(next.TurnOrder, defaultTurnOrder) {
		t.Fatalf("expected default turn order to be stored, got %+v", next.TurnOrder)
	}
}

func TestBeginNextFactionTurnIncrementsRoundOnlyOnWrap(t *testing.T) {
	tests := []struct {
		name      string
		state     game.GameState
		wantTurn  game.Faction
		wantRound int
		wantOrder []game.Faction
	}{
		{
			name: "middle of explicit order keeps round",
			state: game.GameState{
				FactionTurn: game.Marquise,
				RoundNumber: 2,
				TurnOrder:   []game.Faction{game.Marquise, game.Eyrie, game.Alliance},
			},
			wantTurn:  game.Eyrie,
			wantRound: 2,
			wantOrder: []game.Faction{game.Marquise, game.Eyrie, game.Alliance},
		},
		{
			name: "last faction wraps to first and increments round",
			state: game.GameState{
				FactionTurn: game.Alliance,
				RoundNumber: 2,
				TurnOrder:   []game.Faction{game.Marquise, game.Eyrie, game.Alliance},
			},
			wantTurn:  game.Marquise,
			wantRound: 3,
			wantOrder: []game.Faction{game.Marquise, game.Eyrie, game.Alliance},
		},
		{
			name: "custom order wrap increments round",
			state: game.GameState{
				FactionTurn: game.Eyrie,
				RoundNumber: 5,
				TurnOrder:   []game.Faction{game.Vagabond, game.Marquise, game.Eyrie},
			},
			wantTurn:  game.Vagabond,
			wantRound: 6,
			wantOrder: []game.Faction{game.Vagabond, game.Marquise, game.Eyrie},
		},
		{
			name: "default fallback order stores order and increments on wrap",
			state: game.GameState{
				FactionTurn: game.Vagabond,
				RoundNumber: 3,
			},
			wantTurn:  game.Marquise,
			wantRound: 4,
			wantOrder: defaultTurnOrder,
		},
		{
			name: "uninitialized round becomes one on wrap",
			state: game.GameState{
				FactionTurn: game.Vagabond,
				TurnOrder:   []game.Faction{game.Marquise, game.Vagabond},
			},
			wantTurn:  game.Marquise,
			wantRound: 1,
			wantOrder: []game.Faction{game.Marquise, game.Vagabond},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beginNextFactionTurn(&tt.state)

			if tt.state.FactionTurn != tt.wantTurn {
				t.Fatalf("expected next faction %v, got %v", tt.wantTurn, tt.state.FactionTurn)
			}
			if tt.state.RoundNumber != tt.wantRound {
				t.Fatalf("expected round %d, got %d", tt.wantRound, tt.state.RoundNumber)
			}
			if !reflect.DeepEqual(tt.state.TurnOrder, tt.wantOrder) {
				t.Fatalf("expected turn order %+v, got %+v", tt.wantOrder, tt.state.TurnOrder)
			}
			if tt.state.CurrentPhase != game.Birdsong || tt.state.CurrentStep != game.StepBirdsong {
				t.Fatalf("expected next turn to begin at birdsong, got phase=%v step=%v", tt.state.CurrentPhase, tt.state.CurrentStep)
			}
			if !reflect.DeepEqual(tt.state.TurnProgress, game.TurnProgress{}) {
				t.Fatalf("expected turn progress reset, got %+v", tt.state.TurnProgress)
			}
		})
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
