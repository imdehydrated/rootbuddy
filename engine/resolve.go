package engine

import "github.com/imdehydrated/rootbuddy/game"

func warriorCountInClearing(state game.GameState, clearingID int, faction game.Faction) int {
	index := findClearingIndex(state.Map, clearingID)
	if index == -1 {
		return 0
	}

	clearing := state.Map.Clearings[index]
	if clearing.Warriors == nil {
		return 0
	}

	return clearing.Warriors[faction]
}

func ResolveBattle(state game.GameState, action game.Action, attackerRoll int, defenderRoll int) game.Action {
	return ResolveBattleWithModifiers(state, action, attackerRoll, defenderRoll, game.BattleModifiers{})
}

func ResolveBattleWithModifiers(state game.GameState, action game.Action, attackerRoll int, defenderRoll int, modifiers game.BattleModifiers) game.Action {
	if action.Battle == nil {
		return game.Action{}
	}

	attackerHits := min(attackerRoll, warriorCountInClearing(state, action.Battle.ClearingID, action.Battle.Faction))
	defenderHits := min(defenderRoll, warriorCountInClearing(state, action.Battle.ClearingID, action.Battle.TargetFaction))

	attackerHits = max(0, attackerHits+modifiers.AttackerHitModifier)
	defenderHits = max(0, defenderHits+modifiers.DefenderHitModifier)

	if modifiers.IgnoreHitsToDefender {
		attackerHits = 0
	}
	if modifiers.IgnoreHitsToAttacker {
		defenderHits = 0
	}

	return game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:              action.Battle.Faction,
			ClearingID:           action.Battle.ClearingID,
			TargetFaction:        action.Battle.TargetFaction,
			AttackerRoll:         attackerRoll,
			DefenderRoll:         defenderRoll,
			AttackerHitModifier:  modifiers.AttackerHitModifier,
			DefenderHitModifier:  modifiers.DefenderHitModifier,
			IgnoreHitsToAttacker: modifiers.IgnoreHitsToAttacker,
			IgnoreHitsToDefender: modifiers.IgnoreHitsToDefender,
			AttackerLosses:       defenderHits,
			DefenderLosses:       attackerHits,
		},
	}
}
