package rules

import "github.com/imdehydrated/rootbuddy/game"

func marquiseDrawCount(recruitersPlaced int) int {
	switch {
	case recruitersPlaced >= 5:
		return 3
	case recruitersPlaced >= 2:
		return 2
	default:
		return 1
	}
}

func ValidMarquiseEveningActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Marquise {
		return []game.Action{}
	}

	if state.CurrentPhase != game.Evening {
		return []game.Action{}
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepEvening {
		return []game.Action{}
	}

	return []game.Action{
		{
			Type: game.ActionEveningDraw,
			EveningDraw: &game.EveningDrawAction{
				Faction: game.Marquise,
				Count:   marquiseDrawCount(state.Marquise.RecruitersPlaced),
			},
		},
	}
}
