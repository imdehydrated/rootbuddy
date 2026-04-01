package server

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var battleRoller = defaultBattleRoller

func defaultBattleRoller() (int, int, error) {
	attackerRoll, err := rollBattleDie()
	if err != nil {
		return 0, 0, err
	}
	defenderRoll, err := rollBattleDie()
	if err != nil {
		return 0, 0, err
	}
	return attackerRoll, defenderRoll, nil
}

func rollBattleDie() (int, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(4))
	if err != nil {
		return 0, fmt.Errorf("roll battle die: %w", err)
	}
	return int(value.Int64()), nil
}
