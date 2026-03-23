package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidMarquiseBattleActions(state game.GameState) []game.Action {
	if !marquiseIsDaylightActionStep(state) || marquiseActionLimitReached(state) {
		return nil
	}

	return ValidBattlesInState(game.Marquise, state)
}
