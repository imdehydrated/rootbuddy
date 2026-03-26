package engine

import (
	"math/rand"
	"time"

	"github.com/imdehydrated/rootbuddy/game"
)

func ensureRandomSeed(state *game.GameState) {
	if state.RandomSeed == 0 {
		state.RandomSeed = time.Now().UnixNano()
	}
}

func nextShuffleRNG(state *game.GameState) *rand.Rand {
	ensureRandomSeed(state)
	seed := state.RandomSeed + state.ShuffleCount
	state.ShuffleCount++
	return rand.New(rand.NewSource(seed))
}
