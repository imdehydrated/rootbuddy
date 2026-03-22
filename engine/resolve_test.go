package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestResolveBattleCapsHitsByWarriorsPresent(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 2,
						game.Eyrie:    1,
					},
				},
			},
		},
	}

	initiated := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
		},
	}

	resolved := ResolveBattle(state, initiated, 3, 3)

	if resolved.Type != game.ActionBattleResolution {
		t.Fatalf("expected battle resolution action type, got %+v", resolved.Type)
	}
	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if resolved.BattleResolution.DefenderLosses != 2 {
		t.Fatalf("expected defender losses to be capped at 2, got %d", resolved.BattleResolution.DefenderLosses)
	}
	if resolved.BattleResolution.AttackerLosses != 1 {
		t.Fatalf("expected attacker losses to be capped at 1, got %d", resolved.BattleResolution.AttackerLosses)
	}
}

func TestResolveBattleDefenderWithoutWarriorsDealsNoHits(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 3,
					},
					Buildings: []game.Building{
						{Faction: game.Eyrie, Type: game.Sawmill},
					},
				},
			},
		},
	}

	initiated := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
		},
	}

	resolved := ResolveBattle(state, initiated, 2, 3)

	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if resolved.BattleResolution.DefenderLosses != 2 {
		t.Fatalf("expected defender losses to be 2, got %d", resolved.BattleResolution.DefenderLosses)
	}
	if resolved.BattleResolution.AttackerLosses != 0 {
		t.Fatalf("expected attacker losses to be 0 when defender has no warriors, got %d", resolved.BattleResolution.AttackerLosses)
	}
}

func TestResolveBattleReturnsZeroActionWithoutBattlePayload(t *testing.T) {
	state := game.GameState{}
	resolved := ResolveBattle(state, game.Action{}, 1, 1)

	if resolved.Type != 0 || resolved.BattleResolution != nil {
		t.Fatalf("expected zero-value action for missing battle payload, got %+v", resolved)
	}
}

func TestResolveBattleWithModifiersAddsAttackerHits(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 2,
						game.Eyrie:    3,
					},
				},
			},
		},
	}

	initiated := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
		},
	}

	resolved := ResolveBattleWithModifiers(state, initiated, 1, 1, game.BattleModifiers{
		AttackerHitModifier: 1,
	})

	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if resolved.BattleResolution.DefenderLosses != 2 {
		t.Fatalf("expected defender losses to increase to 2, got %d", resolved.BattleResolution.DefenderLosses)
	}
	if resolved.BattleResolution.AttackerHitModifier != 1 {
		t.Fatalf("expected attacker hit modifier to be recorded, got %d", resolved.BattleResolution.AttackerHitModifier)
	}
}

func TestResolveBattleWithModifiersAddsDefenderHits(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 3,
						game.Eyrie:    2,
					},
				},
			},
		},
	}

	initiated := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
		},
	}

	resolved := ResolveBattleWithModifiers(state, initiated, 1, 1, game.BattleModifiers{
		DefenderHitModifier: 1,
	})

	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if resolved.BattleResolution.AttackerLosses != 2 {
		t.Fatalf("expected attacker losses to increase to 2, got %d", resolved.BattleResolution.AttackerLosses)
	}
}

func TestResolveBattleWithModifiersCanIgnoreHitsToAttacker(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 2,
						game.Eyrie:    2,
					},
				},
			},
		},
	}

	initiated := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
		},
	}

	resolved := ResolveBattleWithModifiers(state, initiated, 1, 2, game.BattleModifiers{
		IgnoreHitsToAttacker: true,
	})

	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if resolved.BattleResolution.AttackerLosses != 0 {
		t.Fatalf("expected attacker losses to be ignored, got %d", resolved.BattleResolution.AttackerLosses)
	}
	if !resolved.BattleResolution.IgnoreHitsToAttacker {
		t.Fatalf("expected ignore-hits-to-attacker flag to be recorded")
	}
}

func TestResolveBattleWithModifiersCanIgnoreHitsToDefender(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 2,
						game.Eyrie:    2,
					},
				},
			},
		},
	}

	initiated := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
		},
	}

	resolved := ResolveBattleWithModifiers(state, initiated, 2, 1, game.BattleModifiers{
		IgnoreHitsToDefender: true,
	})

	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if resolved.BattleResolution.DefenderLosses != 0 {
		t.Fatalf("expected defender losses to be ignored, got %d", resolved.BattleResolution.DefenderLosses)
	}
	if !resolved.BattleResolution.IgnoreHitsToDefender {
		t.Fatalf("expected ignore-hits-to-defender flag to be recorded")
	}
}
