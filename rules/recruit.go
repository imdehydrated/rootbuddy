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
	return ValidMarquiseRecruitActions(state)
}

func ValidMarquiseRecruitActions(state game.GameState) []game.Action {
	if !marquiseIsDaylightActionStep(state) || marquiseActionLimitReached(state) || state.TurnProgress.RecruitUsed {
		return nil
	}

	recruiterClearings := marquiseRecruiterClearings(state.Map)
	if len(recruiterClearings) == 0 || state.Marquise.WarriorSupply <= 0 {
		return nil
	}

	recruitCount := state.Marquise.WarriorSupply
	if recruitCount > len(recruiterClearings) {
		recruitCount = len(recruiterClearings)
	}

	actions := make([]game.Action, 0, len(recruiterClearings))
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
