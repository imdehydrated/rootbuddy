package server

import (
	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

type battleRollFunc func(game.GameState) (int, int, error)

var battleRoller battleRollFunc = defaultBattleRoller

func defaultBattleRoller(state game.GameState) (int, int, error) {
	return engine.RollBattleDice(state)
}
