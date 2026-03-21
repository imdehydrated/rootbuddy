package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidRecruitActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Marquise {
		return []game.Action{}
	}

	if state.CurrentPhase != game.Daylight {
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

	if state.Marquise.WarriorSupply < len(recruiterClearings) {
		return []game.Action{}
	}

	return []game.Action{
		{
			Type: game.ActionRecruit,
			Recruit: &game.RecruitAction{
				Faction:     game.Marquise,
				ClearingIDs: recruiterClearings,
			},
		},
	}
}
