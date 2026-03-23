package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidVagabondMoveActions(state game.GameState) []game.Action {
	actions := []game.Action{}
	if clearing, ok := vagabondCurrentClearing(state); ok {
		for _, adjacentID := range clearing.Adj {
			destination, ok := findClearingByID(state.Map, adjacentID)
			if !ok {
				continue
			}

			bootCost := 1 + hostileFactionCountInClearing(state, destination)
			if len(vagabondItemIndexes(state, game.ItemBoots, game.ItemReady)) < bootCost {
				continue
			}

			actions = append(actions, game.Action{
				Type: game.ActionMovement,
				Movement: &game.MovementAction{
					Faction:  game.Vagabond,
					Count:    bootCost,
					MaxCount: bootCost,
					From:     clearing.ID,
					To:       destination.ID,
				},
			})
		}

		if len(vagabondItemIndexes(state, game.ItemBoots, game.ItemReady)) > 0 {
			for _, forestID := range forestIDsAdjacentToClearing(state.Map, clearing.ID) {
				actions = append(actions, game.Action{
					Type: game.ActionMovement,
					Movement: &game.MovementAction{
						Faction:    game.Vagabond,
						Count:      1,
						MaxCount:   1,
						From:       clearing.ID,
						ToForestID: forestID,
					},
				})
			}
		}
	}

	if forest, ok := vagabondCurrentForest(state); ok {
		for _, adjacentID := range forest.AdjacentClearings {
			destination, ok := findClearingByID(state.Map, adjacentID)
			if !ok {
				continue
			}

			bootCost := 1 + hostileFactionCountInClearing(state, destination)
			if len(vagabondItemIndexes(state, game.ItemBoots, game.ItemReady)) < bootCost {
				continue
			}

			actions = append(actions, game.Action{
				Type: game.ActionMovement,
				Movement: &game.MovementAction{
					Faction:      game.Vagabond,
					Count:        bootCost,
					MaxCount:     bootCost,
					To:           destination.ID,
					FromForestID: forest.ID,
				},
			})
		}
	}

	return actions
}
