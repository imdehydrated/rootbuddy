package rules

import "github.com/imdehydrated/rootbuddy/game"

func recruitClearingSubsets(clearingIDs []int, choose int) [][]int {
	if choose <= 0 || choose > len(clearingIDs) {
		return nil
	}

	subsets := [][]int{}
	current := make([]int, 0, choose)

	var build func(start int)
	build = func(start int) {
		if len(current) == choose {
			subset := make([]int, len(current))
			copy(subset, current)
			subsets = append(subsets, subset)
			return
		}

		remainingToPick := choose - len(current)
		maxStart := len(clearingIDs) - remainingToPick
		for i := start; i <= maxStart; i++ {
			current = append(current, clearingIDs[i])
			build(i + 1)
			current = current[:len(current)-1]
		}
	}

	build(0)
	return subsets
}

func ValidRecruitActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Marquise {
		return []game.Action{}
	}

	if state.CurrentStep != game.StepUnspecified {
		if state.CurrentStep != game.StepDaylightActions {
			return []game.Action{}
		}
	} else if state.CurrentPhase != game.Daylight {
		return []game.Action{}
	}

	if state.TurnProgress.ActionsUsed >= 3+state.TurnProgress.BonusActions {
		return []game.Action{}
	}

	if state.TurnProgress.RecruitUsed {
		return []game.Action{}
	}

	recruiterClearings := []int{}
	for _, clearing := range state.Map.Clearings {
		for _, building := range clearing.Buildings {
			if building.Faction == game.Marquise && building.Type == game.Recruiter {
				recruiterClearings = append(recruiterClearings, clearing.ID)
				break
			}
		}
	}

	if len(recruiterClearings) == 0 {
		return []game.Action{}
	}

	if state.Marquise.WarriorSupply <= 0 {
		return []game.Action{}
	}

	recruitCount := state.Marquise.WarriorSupply
	if recruitCount > len(recruiterClearings) {
		recruitCount = len(recruiterClearings)
	}

	actions := make([]game.Action, 0)
	for _, chosenClearings := range recruitClearingSubsets(recruiterClearings, recruitCount) {
		actions = append(actions, game.Action{
			Type: game.ActionRecruit,
			Recruit: &game.RecruitAction{
				Faction:     game.Marquise,
				ClearingIDs: chosenClearings,
			},
		})
	}

	return actions
}
