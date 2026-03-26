package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestResolveBattleWithModifiersDefenderAmbushCanEndBattle(t *testing.T) {
	state := game.GameState{
		PlayerFaction: game.Eyrie,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
						game.Eyrie:    2,
					},
				},
			},
		},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{{ID: 12, Kind: game.AmbushCard, Suit: game.Bird}},
		},
	}

	resolved := ResolveBattleWithModifiers(state, game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
		},
	}, 3, 1, game.BattleModifiers{DefenderAmbush: true})

	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if !resolved.BattleResolution.DefenderAmbushed {
		t.Fatalf("expected defender ambush to be recorded")
	}
	if resolved.BattleResolution.AttackerLosses != 1 || resolved.BattleResolution.DefenderLosses != 0 {
		t.Fatalf("expected ambush to end battle with only attacker losses, got %+v", resolved.BattleResolution)
	}
}

func TestResolveBattleWithModifiersScoutingPartyPreventsAmbush(t *testing.T) {
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
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {30},
		},
	}

	resolved := ResolveBattleWithModifiers(state, game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
		},
	}, 2, 1, game.BattleModifiers{DefenderAmbush: true})

	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if resolved.BattleResolution.DefenderAmbushed {
		t.Fatalf("expected scouting party to prevent defender ambush, got %+v", resolved.BattleResolution)
	}
	if resolved.BattleResolution.AttackerLosses != 1 || resolved.BattleResolution.DefenderLosses != 2 {
		t.Fatalf("expected normal battle to resolve after ignored ambush, got %+v", resolved.BattleResolution)
	}
}

func TestResolveBattleWithModifiersArmorersIgnoreOnlyRolledHits(t *testing.T) {
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
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {1},
		},
	}

	resolved := ResolveBattleWithModifiers(state, game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
		},
	}, 1, 1, game.BattleModifiers{
		AttackerUsesArmorers: true,
		DefenderHitModifier:  1,
	})

	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if !resolved.BattleResolution.AttackerUsedArmorers {
		t.Fatalf("expected armorers use to be recorded")
	}
	if resolved.BattleResolution.AttackerLosses != 1 {
		t.Fatalf("expected armorers to ignore only the rolled hit and still take the extra hit, got %+v", resolved.BattleResolution)
	}
}

func TestApplyBattleResolutionConsumesHiddenAmbushAndArmorersAndScoresBrutalTactics(t *testing.T) {
	state := game.GameState{
		PlayerFaction: game.Marquise,
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
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {1, 5},
		},
		OtherHandCounts: map[game.Faction]int{
			game.Eyrie: 2,
		},
		VictoryPoints: map[game.Faction]int{},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:                   game.Marquise,
			ClearingID:                1,
			TargetFaction:             game.Eyrie,
			DefenderAmbushed:          true,
			AttackerUsedArmorers:      true,
			AttackerUsedBrutalTactics: true,
			AttackerLosses:            1,
			DefenderLosses:            1,
		},
	})

	if next.OtherHandCounts[game.Eyrie] != 1 {
		t.Fatalf("expected hidden defender ambush to consume one hand count, got %+v", next.OtherHandCounts)
	}
	if hasPersistentEffect(next, game.Marquise, "armorers") {
		t.Fatalf("expected used armorers to be discarded, got %+v", next.PersistentEffects)
	}
	if !hasPersistentEffect(next, game.Marquise, "brutal_tactics") {
		t.Fatalf("expected brutal tactics to remain in play, got %+v", next.PersistentEffects)
	}
	if next.VictoryPoints[game.Eyrie] != 1 {
		t.Fatalf("expected brutal tactics to award 1 VP to defender, got %+v", next.VictoryPoints)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 1 {
		t.Fatalf("expected used armorers to be discarded publicly, got %+v", next.DiscardPile)
	}
}
