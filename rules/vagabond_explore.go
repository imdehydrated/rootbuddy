package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidExploreActions(state game.GameState) []game.Action {
	if len(vagabondItemIndexes(state, game.ItemTorch, game.ItemReady)) == 0 {
		return nil
	}

	clearing, ok := vagabondCurrentClearing(state)
	if !ok || !clearing.Ruins || len(clearing.RuinItems) == 0 {
		return nil
	}

	actions := []game.Action{}
	for _, itemType := range clearing.RuinItems {
		actions = append(actions, game.Action{
			Type: game.ActionExplore,
			Explore: &game.ExploreAction{
				Faction:    game.Vagabond,
				ClearingID: clearing.ID,
				ItemType:   itemType,
			},
		})
	}

	return actions
}
