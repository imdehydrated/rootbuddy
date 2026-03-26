package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidateStateAcceptsActiveSetupStage(t *testing.T) {
	state := game.GameState{
		GamePhase:  game.LifecycleSetup,
		SetupStage: game.SetupStageMarquise,
	}

	if err := ValidateState(state); err != nil {
		t.Fatalf("expected setup stage to validate, got %v", err)
	}
}

func TestValidateStateRejectsSetupWithActiveStep(t *testing.T) {
	state := game.GameState{
		GamePhase:   game.LifecycleSetup,
		SetupStage:  game.SetupStageMarquise,
		CurrentStep: game.StepBirdsong,
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected setup state with active step to fail validation")
	}
}

func TestValidateStateRejectsNegativeOtherHandCount(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		OtherHandCounts: map[game.Faction]int{
			game.Eyrie: -1,
		},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected negative other-hand count to fail validation")
	}
}

func TestApplyActionProducesValidState(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
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
		Marquise: game.MarquiseState{
			WarriorSupply: 1,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
		},
	})

	if err := ValidateState(next); err != nil {
		t.Fatalf("expected applied recruit state to validate, got %v", err)
	}
}
