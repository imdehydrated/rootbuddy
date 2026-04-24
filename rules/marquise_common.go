package rules

import "github.com/imdehydrated/rootbuddy/game"

func marquiseIsDaylightActionStep(state game.GameState) bool {
	if state.FactionTurn != game.Marquise {
		return false
	}

	if state.CurrentStep != game.StepUnspecified {
		return state.CurrentStep == game.StepDaylightActions
	}

	return state.CurrentPhase == game.Daylight
}

func marquiseActionLimitReached(state game.GameState) bool {
	return state.TurnProgress.ActionsUsed >= 3+state.TurnProgress.BonusActions
}

func marquiseRecruiterClearings(board game.Map) []int {
	recruiterClearings := []int{}
	for _, clearing := range board.Clearings {
		for _, building := range clearing.Buildings {
			if building.Faction == game.Marquise && building.Type == game.Recruiter {
				recruiterClearings = append(recruiterClearings, clearing.ID)
			}
		}
	}

	return recruiterClearings
}

func marquiseHasSawmill(clearing game.Clearing) bool {
	for _, building := range clearing.Buildings {
		if building.Faction == game.Marquise && building.Type == game.Sawmill {
			return true
		}
	}
	return false
}
