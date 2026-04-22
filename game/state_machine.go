package game

import "fmt"

type TurnWindow struct {
	Lifecycle  GameLifecycle
	SetupStage SetupStage
	Phase      Phase
	Step       TurnStep
}

func DaylightEntryStep(faction Faction) TurnStep {
	switch faction {
	case Marquise, Eyrie, Alliance:
		return StepDaylightCraft
	default:
		return StepDaylightActions
	}
}

func (state GameState) TurnWindow() TurnWindow {
	lifecycle := state.GamePhase
	if lifecycle == LifecycleSetup && state.SetupStage == SetupStageUnspecified {
		lifecycle = LifecyclePlaying
	}

	window := TurnWindow{
		Lifecycle:  lifecycle,
		SetupStage: state.SetupStage,
		Phase:      state.CurrentPhase,
		Step:       state.CurrentStep,
	}

	if window.Lifecycle == LifecycleSetup {
		return window
	}

	if window.Step != StepUnspecified {
		return window
	}

	switch window.Phase {
	case Birdsong:
		window.Step = StepBirdsong
	case Daylight:
		window.Step = DaylightEntryStep(state.FactionTurn)
	case Evening:
		window.Step = StepEvening
	}

	return window
}

func (window TurnWindow) Validate() error {
	switch window.Lifecycle {
	case LifecycleSetup:
		if window.SetupStage == SetupStageUnspecified || window.SetupStage == SetupStageComplete {
			return fmt.Errorf("setup lifecycle requires an active setup stage")
		}
		if window.Step != StepUnspecified {
			return fmt.Errorf("setup lifecycle cannot have an active step")
		}
	case LifecyclePlaying:
		if window.SetupStage != SetupStageComplete && window.SetupStage != SetupStageUnspecified {
			return fmt.Errorf("playing lifecycle cannot use setup stage %v", window.SetupStage)
		}
		switch window.Phase {
		case Birdsong:
			if window.Step != StepBirdsong {
				return fmt.Errorf("birdsong phase requires birdsong step")
			}
		case Daylight:
			if window.Step != StepDaylightCraft && window.Step != StepDaylightActions {
				return fmt.Errorf("daylight phase requires daylight craft/actions step")
			}
		case Evening:
			if window.Step != StepEvening {
				return fmt.Errorf("evening phase requires evening step")
			}
		default:
			return fmt.Errorf("playing lifecycle has unknown phase %v", window.Phase)
		}
	case LifecycleGameOver:
		if window.SetupStage != SetupStageComplete && window.SetupStage != SetupStageUnspecified {
			return fmt.Errorf("game over lifecycle cannot use setup stage %v", window.SetupStage)
		}
	default:
		return fmt.Errorf("unknown lifecycle %v", window.Lifecycle)
	}

	return nil
}
