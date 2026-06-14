package server

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestDefaultBattleRollerUsesRecordedGameState(t *testing.T) {
	state := game.GameState{
		RandomSeed:      9876,
		BattleRollCount: 4,
	}

	firstAttacker, firstDefender, err := defaultBattleRoller(state)
	if err != nil {
		t.Fatalf("default battle roller: %v", err)
	}
	secondAttacker, secondDefender, err := defaultBattleRoller(state)
	if err != nil {
		t.Fatalf("default battle roller again: %v", err)
	}

	if firstAttacker != secondAttacker || firstDefender != secondDefender {
		t.Fatalf("expected server battle roller to be deterministic for same state, got %d/%d then %d/%d", firstAttacker, firstDefender, secondAttacker, secondDefender)
	}
}
