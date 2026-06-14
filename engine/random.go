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

func battleRollRNG(state game.GameState) *rand.Rand {
	seed := int64(uint64(state.RandomSeed) ^
		(uint64(state.BattleRollCount+1) * 0x9e3779b97f4a7c15) ^
		0x626174746c65)
	return rand.New(rand.NewSource(seed))
}

func RollBattleDice(state game.GameState) (int, int, error) {
	rng := battleRollRNG(state)
	return rng.Intn(4), rng.Intn(4), nil
}
