package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidVagabondBattleActions(state game.GameState) []game.Action {
	if len(vagabondItemIndexes(state, game.ItemSword, game.ItemReady)) == 0 {
		return nil
	}

	clearing, ok := vagabondCurrentClearing(state)
	if !ok {
		return nil
	}

	actions := []game.Action{}
	for _, targetFaction := range vagabondFactionsInClearing(clearing) {
		actions = append(actions, game.Action{
			Type: game.ActionBattle,
			Battle: &game.BattleAction{
				Faction:       game.Vagabond,
				ClearingID:    clearing.ID,
				TargetFaction: targetFaction,
			},
		})
	}

	return actions
}
