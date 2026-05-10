package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidVagabondBattleActions(state game.GameState) []game.Action {
	if len(vagabondItemIndexes(state, game.ItemSword, game.ItemReady)) == 0 {
		return nil
	}

	clearing, ok := vagabondCurrentClearing(state)
	if !ok {
		return nil
	}

	actions := []game.Action{}
	for _, targetFaction := range vagabondFactionsInClearing(clearing) {
		if !game.AreEnemies(state, game.Vagabond, targetFaction) {
			continue
		}
		actions = append(actions, vagabondBattleAction(clearing.ID, targetFaction, 0, false))
		for _, alliedFaction := range vagabondAlliedBattleFactions(state, clearing, targetFaction) {
			actions = append(actions, vagabondBattleAction(clearing.ID, targetFaction, alliedFaction, true))
		}
	}

	return actions
}

func vagabondBattleAction(clearingID int, targetFaction game.Faction, alliedFaction game.Faction, useAlliedFaction bool) game.Action {
	return game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:          game.Vagabond,
			ClearingID:       clearingID,
			TargetFaction:    targetFaction,
			UseAlliedFaction: useAlliedFaction,
			AlliedFaction:    alliedFaction,
		},
	}
}

func vagabondAlliedBattleFactions(state game.GameState, clearing game.Clearing, targetFaction game.Faction) []game.Faction {
	factions := []game.Faction{}
	for _, faction := range []game.Faction{game.Marquise, game.Alliance, game.Eyrie} {
		if faction == targetFaction || vagabondRelationshipLevel(state, faction) != game.RelAllied {
			continue
		}
		if clearing.Warriors != nil && clearing.Warriors[faction] > 0 {
			factions = append(factions, faction)
		}
	}

	return factions
}
