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

			actions = append(actions, vagabondMoveAction(clearing.ID, destination.ID, bootCost, 0, 0))
			actions = append(actions, vagabondAlliedMoveActions(state, clearing, destination.ID, bootCost)...)
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

func vagabondMoveAction(from int, to int, bootCost int, alliedFaction game.Faction, alliedWarriors int) game.Action {
	return game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:        game.Vagabond,
			Count:          bootCost,
			MaxCount:       bootCost,
			From:           from,
			To:             to,
			AlliedFaction:  alliedFaction,
			AlliedWarriors: alliedWarriors,
		},
	}
}

func vagabondAlliedMoveActions(state game.GameState, clearing game.Clearing, to int, bootCost int) []game.Action {
	actions := []game.Action{}
	for _, faction := range []game.Faction{game.Marquise, game.Alliance, game.Eyrie} {
		if vagabondRelationshipLevel(state, faction) != game.RelAllied {
			continue
		}
		available := 0
		if clearing.Warriors != nil {
			available = clearing.Warriors[faction]
		}
		for count := 1; count <= available; count++ {
			actions = append(actions, vagabondMoveAction(clearing.ID, to, bootCost, faction, count))
		}
	}

	return actions
}
