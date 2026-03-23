package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidEyrieTurmoilActions(state game.GameState) []game.Action {
	actions := []game.Action{}
	for _, leader := range availableNewLeaders(state) {
		actions = append(actions, game.Action{
			Type: game.ActionTurmoil,
			Turmoil: &game.TurmoilAction{
				Faction:   game.Eyrie,
				NewLeader: leader,
			},
		})
	}
	return actions
}
