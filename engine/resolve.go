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

	battleSuit := clearingSuit(state, action.Battle.ClearingID)
	attackerHasScoutingParty := hasPersistentEffect(state, action.Battle.Faction, "scouting_party")
	defenderAmbushed := false
	attackerCounterAmbush := false
	ambushHitsToAttacker := 0
	simulatedState := cloneState(state)

	if modifiers.DefenderAmbush && !attackerHasScoutingParty && canFactionPlayAmbush(state, action.Battle.TargetFaction, battleSuit) {
		defenderAmbushed = true
		if modifiers.AttackerCounterAmbush && canFactionPlayAmbush(state, action.Battle.Faction, battleSuit) {
			attackerCounterAmbush = true
		} else {
			ambushHitsToAttacker = applyHypotheticalAmbushHits(&simulatedState, action.Battle.Faction, action.Battle.ClearingID, 2)
			if !attackersRemainAfterAmbush(simulatedState, action.Battle.Faction, action.Battle.ClearingID) {
				return game.Action{
					Type: game.ActionBattleResolution,
					BattleResolution: &game.BattleResolutionAction{
						Faction:               action.Battle.Faction,
						ClearingID:            action.Battle.ClearingID,
						TargetFaction:         action.Battle.TargetFaction,
						DecreeCardID:          action.Battle.DecreeCardID,
						AttackerRoll:          attackerRoll,
						DefenderRoll:          defenderRoll,
						DefenderAmbushed:      true,
						AmbushHitsToAttacker:  ambushHitsToAttacker,
						AttackerCounterAmbush: false,
						AttackerLosses:        ambushHitsToAttacker,
						DefenderLosses:        0,
						SourceEffectID:        action.Battle.SourceEffectID,
					},
				}
			}
		}
	}

	if action.Battle.TargetFaction == game.Alliance && attackerRoll > defenderRoll {
		attackerRoll, defenderRoll = defenderRoll, attackerRoll
	}

	attackerUsedArmorers := modifiers.AttackerUsesArmorers && hasPersistentEffect(state, action.Battle.Faction, "armorers")
	defenderUsedArmorers := modifiers.DefenderUsesArmorers && hasPersistentEffect(state, action.Battle.TargetFaction, "armorers")
	attackerUsedBrutalTactics := modifiers.AttackerUsesBrutalTactics && hasPersistentEffect(state, action.Battle.Faction, "brutal_tactics")
	defenderUsedSappers := modifiers.DefenderUsesSappers && hasPersistentEffect(state, action.Battle.TargetFaction, "sappers")

	if action.Battle.Faction == game.Eyrie && state.Eyrie.Leader == game.LeaderCommander {
		modifiers.AttackerHitModifier++
	}

	attackerCap := warriorCountInClearing(simulatedState, action.Battle.ClearingID, action.Battle.Faction)
	if action.Battle.Faction == game.Vagabond {
		attackerCap = vagabondBattleHitCap(simulatedState)
	}

	defenderCap := warriorCountInClearing(simulatedState, action.Battle.ClearingID, action.Battle.TargetFaction)
	if action.Battle.TargetFaction == game.Vagabond {
		defenderCap = vagabondBattleHitCap(simulatedState)
	}

	attackerRolledHits := min(attackerRoll, attackerCap)
	defenderRolledHits := min(defenderRoll, defenderCap)

	if attackerUsedArmorers {
		defenderRolledHits = 0
	}
	if defenderUsedArmorers {
		attackerRolledHits = 0
	}

	attackerExtraHits := modifiers.AttackerHitModifier
	defenderExtraHits := modifiers.DefenderHitModifier
	if attackerUsedBrutalTactics {
		attackerExtraHits++
	}
	if defenderUsedSappers {
		defenderExtraHits++
	}

	attackerHits := max(0, attackerRolledHits+attackerExtraHits)
	defenderHits := max(0, defenderRolledHits+defenderExtraHits)

	if modifiers.IgnoreHitsToDefender {
		attackerHits = 0
	}
	if modifiers.IgnoreHitsToAttacker {
		defenderHits = 0
	}

	return game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:                   action.Battle.Faction,
			ClearingID:                action.Battle.ClearingID,
			TargetFaction:             action.Battle.TargetFaction,
			DecreeCardID:              action.Battle.DecreeCardID,
			AttackerRoll:              attackerRoll,
			DefenderRoll:              defenderRoll,
			AttackerHitModifier:       modifiers.AttackerHitModifier,
			DefenderHitModifier:       modifiers.DefenderHitModifier,
			IgnoreHitsToAttacker:      modifiers.IgnoreHitsToAttacker,
			IgnoreHitsToDefender:      modifiers.IgnoreHitsToDefender,
			DefenderAmbushed:          defenderAmbushed,
			AttackerCounterAmbush:     attackerCounterAmbush,
			AttackerUsedArmorers:      attackerUsedArmorers,
			DefenderUsedArmorers:      defenderUsedArmorers,
			AttackerUsedBrutalTactics: attackerUsedBrutalTactics,
			DefenderUsedSappers:       defenderUsedSappers,
			AmbushHitsToAttacker:      ambushHitsToAttacker,
			AttackerLosses:            ambushHitsToAttacker + defenderHits,
			DefenderLosses:            attackerHits,
			SourceEffectID:            action.Battle.SourceEffectID,
		},
	}
}
