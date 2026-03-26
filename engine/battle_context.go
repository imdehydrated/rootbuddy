package engine

import "github.com/imdehydrated/rootbuddy/game"

func BattleContext(state game.GameState, action game.Action) game.BattleContext {
	context := game.BattleContext{
		Action: action,
		Timing: []game.BattleTimingStep{
			game.BattleTimingAmbush,
			game.BattleTimingCounterAmbush,
			game.BattleTimingModifiers,
			game.BattleTimingRolls,
		},
	}

	if action.Battle == nil {
		return context
	}

	context.ClearingSuit = clearingSuit(state, action.Battle.ClearingID)
	context.AttackerHasScoutingParty = hasPersistentEffect(state, action.Battle.Faction, "scouting_party")
	context.CanDefenderAmbush = !context.AttackerHasScoutingParty &&
		canFactionPlayAmbush(state, action.Battle.TargetFaction, context.ClearingSuit)
	context.CanAttackerCounterAmbush = canFactionPlayAmbush(state, action.Battle.Faction, context.ClearingSuit)
	context.CanAttackerArmorers = hasPersistentEffect(state, action.Battle.Faction, "armorers")
	context.CanDefenderArmorers = hasPersistentEffect(state, action.Battle.TargetFaction, "armorers")
	context.CanAttackerBrutalTactics = hasPersistentEffect(state, action.Battle.Faction, "brutal_tactics")
	context.CanDefenderSappers = hasPersistentEffect(state, action.Battle.TargetFaction, "sappers")
	context.AssistDefenderAmbushPromptRequired = state.GameMode == game.GameModeAssist &&
		action.Battle.TargetFaction != state.PlayerFaction &&
		context.CanDefenderAmbush

	return context
}
