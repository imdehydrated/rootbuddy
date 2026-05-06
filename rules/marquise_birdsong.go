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

	passAction := game.Action{
		Type: game.ActionPassPhase,
		PassPhase: &game.PassPhaseAction{
			Faction: game.Marquise,
		},
	}

	clearingIDs := []int{}
	for _, clearing := range state.Map.Clearings {
		if marquiseHasSawmill(clearing) {
			clearingIDs = append(clearingIDs, clearing.ID)
		}
	}

	if len(clearingIDs) == 0 || state.Marquise.WoodSupply <= 0 {
		return []game.Action{passAction}
	}

	if state.Marquise.WoodSupply < len(clearingIDs) {
		return limitedWoodPlacementActions(clearingIDs, state.Marquise.WoodSupply)
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

func limitedWoodPlacementActions(clearingIDs []int, count int) []game.Action {
	if count <= 0 {
		return nil
	}

	actions := []game.Action{}
	selected := make([]int, 0, count)
	var choose func(start int)
	choose = func(start int) {
		if len(selected) == count {
			clearingIDs := append([]int(nil), selected...)
			actions = append(actions, game.Action{
				Type: game.ActionBirdsongWood,
				BirdsongWood: &game.BirdsongWoodAction{
					Faction:     game.Marquise,
					ClearingIDs: clearingIDs,
					Amount:      1,
				},
			})
			return
		}

		remainingNeeded := count - len(selected)
		for index := start; index <= len(clearingIDs)-remainingNeeded; index++ {
			selected = append(selected, clearingIDs[index])
			choose(index + 1)
			selected = selected[:len(selected)-1]
		}
	}
	choose(0)
	return actions
}
