package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidMarquiseMovementActions(state game.GameState) []game.Action {
	if !marquiseIsDaylightActionStep(state) || state.TurnProgress.MarchesUsed >= 2 {
		return nil
	}
	if marquiseActionLimitReached(state) && state.TurnProgress.MarchesUsed == 0 {
		return nil
	}

	return ValidMovementActions(game.Marquise, state.Map)
}
