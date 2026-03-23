package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidMarquiseMovementActions(state game.GameState) []game.Action {
	if !marquiseIsDaylightActionStep(state) || marquiseActionLimitReached(state) || state.TurnProgress.MarchesUsed >= 2 {
		return nil
	}

	return ValidMovementActions(game.Marquise, state.Map)
}
