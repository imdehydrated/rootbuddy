package server

import (
	"errors"
	"reflect"

	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

var errInvalidActionForState = errors.New("action is not valid in current state")

func actionMatchesValidAction(state game.GameState, action game.Action) bool {
	for _, candidate := range engine.ValidActions(state) {
		if reflect.DeepEqual(candidate, action) {
			return true
		}
	}
	return false
}
