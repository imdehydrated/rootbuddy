package engine

import (
	"errors"
	"reflect"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestApplyLegalActionAppliesGeneratedAction(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Sawmill},
					},
				},
			},
		},
		Marquise: game.MarquiseState{
			WoodSupply: 1,
		},
	}
	action := game.Action{
		Type: game.ActionBirdsongWood,
		BirdsongWood: &game.BirdsongWoodAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
			Amount:      1,
		},
	}

	next, err := ApplyLegalAction(state, action)

	if err != nil {
		t.Fatalf("expected legal action to apply, got %v", err)
	}
	if next.Map.Clearings[0].Wood != 1 || next.Marquise.WoodSupply != 0 {
		t.Fatalf("expected legal birdsong wood action to mutate state, got clearing=%+v marquise=%+v", next.Map.Clearings[0], next.Marquise)
	}
}

func TestApplyLegalActionRejectsNonGeneratedActionWithoutMutating(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Marquise: game.MarquiseState{
			WoodSupply: 1,
		},
	}
	originalInput := state
	originalClone := CloneState(state)
	illegal := game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
		},
	}

	next, err := ApplyLegalAction(state, illegal)

	if !errors.Is(err, ErrIllegalAction) {
		t.Fatalf("expected ErrIllegalAction, got %v", err)
	}
	if !reflect.DeepEqual(next, originalClone) {
		t.Fatalf("expected rejected action to return unchanged clone\nwant=%+v\ngot=%+v", originalClone, next)
	}
	if !reflect.DeepEqual(state, originalInput) {
		t.Fatalf("expected rejected action not to mutate input\nwant=%+v\ngot=%+v", originalInput, state)
	}
	if advanced := ApplyAction(state, illegal); reflect.DeepEqual(advanced, originalClone) {
		t.Fatalf("expected low-level ApplyAction to remain permissive for direct engine use")
	}
}

func TestApplyLegalActionRejectsAfterGameOver(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecycleGameOver,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
	}
	action := game.Action{
		Type: game.ActionPassPhase,
		PassPhase: &game.PassPhaseAction{
			Faction: game.Marquise,
		},
	}

	next, err := ApplyLegalAction(state, action)
	want := CloneState(state)

	if !errors.Is(err, ErrGameOver) {
		t.Fatalf("expected ErrGameOver, got %v", err)
	}
	if !reflect.DeepEqual(next, want) {
		t.Fatalf("expected game-over rejection to return unchanged clone\nwant=%+v\ngot=%+v", want, next)
	}
}

func TestApplyLegalActionRejectsRawBattleResolution(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 2,
						game.Eyrie:    1,
					},
				},
			},
		},
	}
	battle := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
		},
	}
	requireScenarioAction(t, state, battle)
	rawResolution := ResolveBattle(state, battle, 1, 0)

	if IsLegalAction(state, rawResolution) {
		t.Fatalf("expected raw battle resolution not to be part of generated legal actions")
	}
	next, err := ApplyLegalAction(state, rawResolution)
	if !errors.Is(err, ErrIllegalAction) {
		t.Fatalf("expected ErrIllegalAction for raw battle resolution, got %v", err)
	}
	if !reflect.DeepEqual(next, state) {
		t.Fatalf("expected rejected raw battle resolution to leave state unchanged\nwant=%+v\ngot=%+v", state, next)
	}
}
