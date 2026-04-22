package rules

import "github.com/imdehydrated/rootbuddy/game"

var roostScoreTrack = []int{0, 0, 1, 2, 3, 4, 4, 5}
var roostDrawBonusTrack = []int{0, 0, 0, 1, 1, 1, 2, 2}

func roostScore(roostsPlaced int) int {
	if roostsPlaced < 0 || roostsPlaced >= len(roostScoreTrack) {
		return 0
	}
	return roostScoreTrack[roostsPlaced]
}

func eyrieDrawCount(roostsPlaced int) int {
	if roostsPlaced < 0 {
		return 1
	}
	if roostsPlaced >= len(roostDrawBonusTrack) {
		return 1 + roostDrawBonusTrack[len(roostDrawBonusTrack)-1]
	}
	return 1 + roostDrawBonusTrack[roostsPlaced]
}

func ValidEyrieEveningActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Eyrie {
		return []game.Action{}
	}

	if state.CurrentPhase != game.Evening {
		return []game.Action{}
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepEvening {
		return []game.Action{}
	}

	if state.TurnProgress.EveningMainActionTaken {
		return []game.Action{
			{
				Type: game.ActionEveningDraw,
				EveningDraw: &game.EveningDrawAction{
					Faction: game.Eyrie,
					Count:   eyrieDrawCount(state.Eyrie.RoostsPlaced),
				},
			},
		}
	}

	return []game.Action{
		{
			Type: game.ActionScoreRoosts,
			ScoreRoosts: &game.ScoreRoostsAction{
				Faction: game.Eyrie,
				Points:  roostScore(state.Eyrie.RoostsPlaced),
			},
		},
	}
}
