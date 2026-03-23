package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidMarquiseBirdsongWoodActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Marquise {
		return []game.Action{}
	}

	if state.CurrentPhase != game.Birdsong {
		return []game.Action{}
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepBirdsong {
		return []game.Action{}
	}

	clearingIDs := []int{}
	for _, clearing := range state.Map.Clearings {
		if hasMarquiseSawmill(clearing) {
			clearingIDs = append(clearingIDs, clearing.ID)
		}
	}

	if len(clearingIDs) == 0 {
		return []game.Action{
			{
				Type: game.ActionPassPhase,
				PassPhase: &game.PassPhaseAction{
					Faction: game.Marquise,
				},
			},
		}
	}

	return []game.Action{
		{
			Type: game.ActionBirdsongWood,
			BirdsongWood: &game.BirdsongWoodAction{
				Faction:     game.Marquise,
				ClearingIDs: clearingIDs,
				Amount:      1,
			},
		},
	}
}
