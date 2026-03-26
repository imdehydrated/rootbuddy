package engine

import "github.com/imdehydrated/rootbuddy/game"

func CloneState(state game.GameState) game.GameState {
	return cloneState(state)
}
