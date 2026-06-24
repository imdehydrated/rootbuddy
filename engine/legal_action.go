package engine

import (
	"errors"
	"reflect"

	"github.com/imdehydrated/rootbuddy/game"
)

var (
	ErrIllegalAction = errors.New("action is not valid in current state")
	ErrGameOver      = errors.New("cannot apply action after game over")
)

func IsLegalAction(state game.GameState, action game.Action) bool {
	for _, candidate := range ValidActions(state) {
		if reflect.DeepEqual(candidate, action) {
			return true
		}
	}
	return false
}

func ApplyLegalAction(state game.GameState, action game.Action) (game.GameState, error) {
	next, _, err := ApplyLegalActionDetailed(state, action)
	return next, err
}

func ApplyLegalActionDetailed(state game.GameState, action game.Action) (game.GameState, *game.EffectResult, error) {
	if state.GamePhase == game.LifecycleGameOver {
		return CloneState(state), nil, ErrGameOver
	}
	if !IsLegalAction(state, action) {
		return CloneState(state), nil, ErrIllegalAction
	}

	next, result := ApplyActionDetailed(state, action)
	return next, result, nil
}
