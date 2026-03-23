package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidRepairActions(state game.GameState) []game.Action {
	if len(vagabondItemIndexes(state, game.ItemHammer, game.ItemReady)) == 0 {
		return nil
	}

	actions := []game.Action{}
	for index, item := range state.Vagabond.Items {
		if item.Status != game.ItemDamaged {
			continue
		}

		actions = append(actions, game.Action{
			Type: game.ActionRepair,
			Repair: &game.RepairAction{
				Faction:   game.Vagabond,
				ItemIndex: index,
			},
		})
	}

	return actions
}
