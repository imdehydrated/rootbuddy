package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidMarquiseBuildActions(state game.GameState) []game.Action {
	if !marquiseIsDaylightActionStep(state) || marquiseActionLimitReached(state) {
		return nil
	}

	return ValidBuilds(state.Map, state.Marquise)
}
