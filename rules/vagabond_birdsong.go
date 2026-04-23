package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidVagabondBirdsongActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Vagabond || state.CurrentPhase != game.Birdsong {
		return nil
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepBirdsong {
		return nil
	}

	actions := []game.Action{}
	exhausted := vagabondExhaustedItemIndexes(state)
	refreshLimit := 3 + len(vagabondItemIndexes(state, game.ItemTea, game.ItemReady))
	if refreshLimit > len(exhausted) {
		refreshLimit = len(exhausted)
	}
	if refreshLimit > 0 {
		for _, refreshedIndexes := range chooseItemIndexSubsets(exhausted, refreshLimit) {
			actions = append(actions, game.Action{
				Type: game.ActionDaybreak,
				Daybreak: &game.DaybreakAction{
					Faction:              game.Vagabond,
					RefreshedItemIndexes: refreshedIndexes,
				},
			})
		}
	}

	if !state.TurnProgress.HasSlipped {
		if clearing, ok := vagabondCurrentClearing(state); ok {
			actions = append(actions, game.Action{
				Type: game.ActionSlip,
				Slip: &game.SlipAction{
					Faction: game.Vagabond,
					From:    clearing.ID,
					To:      clearing.ID,
				},
			})

			for _, adjacentID := range clearing.Adj {
				actions = append(actions, game.Action{
					Type: game.ActionSlip,
					Slip: &game.SlipAction{
						Faction: game.Vagabond,
						From:    clearing.ID,
						To:      adjacentID,
					},
				})
			}

			for _, forestID := range forestIDsAdjacentToClearing(state.Map, clearing.ID) {
				actions = append(actions, game.Action{
					Type: game.ActionSlip,
					Slip: &game.SlipAction{
						Faction:    game.Vagabond,
						From:       clearing.ID,
						ToForestID: forestID,
					},
				})
			}
		}

		if forest, ok := vagabondCurrentForest(state); ok {
			actions = append(actions, game.Action{
				Type: game.ActionSlip,
				Slip: &game.SlipAction{
					Faction:      game.Vagabond,
					FromForestID: forest.ID,
					ToForestID:   forest.ID,
				},
			})

			for _, adjacentID := range forest.AdjacentClearings {
				actions = append(actions, game.Action{
					Type: game.ActionSlip,
					Slip: &game.SlipAction{
						Faction:      game.Vagabond,
						To:           adjacentID,
						FromForestID: forest.ID,
					},
				})
			}
		}
	}

	if state.TurnProgress.HasSlipped {
		actions = append(actions, game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Vagabond,
			},
		})
	}

	return actions
}
