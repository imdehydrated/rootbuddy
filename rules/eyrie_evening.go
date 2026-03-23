package rules

import "github.com/imdehydrated/rootbuddy/game"

var roostScoreTrack = []int{0, 0, 1, 2, 3, 4, 4, 5}

func roostScore(roostsPlaced int) int {
	if roostsPlaced < 0 || roostsPlaced >= len(roostScoreTrack) {
		return 0
	}
	return roostScoreTrack[roostsPlaced]
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
