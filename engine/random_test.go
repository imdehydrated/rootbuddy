package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestRollBattleDiceIsDeterministicForSeedAndBattleRollCount(t *testing.T) {
	state := game.GameState{
		RandomSeed:      12345,
		BattleRollCount: 7,
	}

	firstAttacker, firstDefender, err := RollBattleDice(state)
	if err != nil {
		t.Fatalf("roll battle dice: %v", err)
	}
	secondAttacker, secondDefender, err := RollBattleDice(state)
	if err != nil {
		t.Fatalf("roll battle dice again: %v", err)
	}

	if firstAttacker != secondAttacker || firstDefender != secondDefender {
		t.Fatalf("expected same seed and roll count to produce same rolls, got %d/%d then %d/%d", firstAttacker, firstDefender, secondAttacker, secondDefender)
	}
	if firstAttacker < 0 || firstAttacker > 3 || firstDefender < 0 || firstDefender > 3 {
		t.Fatalf("expected battle dice in range 0-3, got %d/%d", firstAttacker, firstDefender)
	}
}

func TestApplyBattleResolutionAdvancesBattleRollCount(t *testing.T) {
	state := game.GameState{
		RandomSeed:      12345,
		BattleRollCount: 2,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
						game.Eyrie:    1,
					},
				},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
			AttackerRoll:  0,
			DefenderRoll:  0,
		},
	})

	if next.BattleRollCount != 3 {
		t.Fatalf("expected battle roll count to advance after valid battle resolution, got %d", next.BattleRollCount)
	}
}

func TestApplyRejectedBattleResolutionDoesNotAdvanceBattleRollCount(t *testing.T) {
	state := game.GameState{
		RandomSeed:      12345,
		BattleRollCount: 2,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
						game.Eyrie:    1,
					},
				},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:       game.Marquise,
			ClearingID:    99,
			TargetFaction: game.Eyrie,
			AttackerRoll:  0,
			DefenderRoll:  0,
		},
	})

	if next.BattleRollCount != 2 {
		t.Fatalf("expected rejected battle resolution to leave battle roll count unchanged, got %d", next.BattleRollCount)
	}
}
