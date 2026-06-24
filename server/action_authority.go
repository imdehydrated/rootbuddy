package server

import (
	"errors"

	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

var errInvalidActionForState = errors.New("action is not valid in current state")

func actionMatchesValidAction(state game.GameState, action game.Action) bool {
	return engine.IsLegalAction(state, action)
}
