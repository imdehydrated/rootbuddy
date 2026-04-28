package engine

import "github.com/imdehydrated/rootbuddy/game"

func advanceTurnState(state *game.GameState, action game.Action) {
	if state.GamePhase == game.LifecycleSetup && state.SetupStage != game.SetupStageUnspecified {
		switch action.Type {
		case game.ActionMarquiseSetup, game.ActionEyrieSetup, game.ActionVagabondSetup:
			advanceSetupStage(state)
		}
		return
	}

	switch action.Type {
	case game.ActionAddToDecree:
		state.TurnProgress.BirdsongMainActionTaken = true
		if state.FactionTurn == game.Eyrie && !eyrieHasRoostOnMap(*state) {
			state.CurrentPhase = game.Birdsong
			state.CurrentStep = game.StepBirdsong
		} else {
			state.CurrentPhase = game.Daylight
			state.CurrentStep = game.DaylightEntryStep(state.FactionTurn)
		}
	case game.ActionEyrieEmergencyOrders:
		state.TurnProgress.BirdsongMainActionTaken = true
		state.TurnProgress.EyrieEmergencyResolved = true
		state.CurrentPhase = game.Birdsong
		state.CurrentStep = game.StepBirdsong
	case game.ActionEyrieNewRoost:
		state.TurnProgress.BirdsongMainActionTaken = true
		state.TurnProgress.EyrieNewRoostResolved = true
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.DaylightEntryStep(state.FactionTurn)
	case game.ActionBirdsongWood:
		state.TurnProgress.BirdsongMainActionTaken = true
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.DaylightEntryStep(state.FactionTurn)
	case game.ActionDaybreak:
		state.TurnProgress.BirdsongMainActionTaken = true
		state.TurnProgress.HasRefreshed = true
		state.CurrentPhase = game.Birdsong
		state.CurrentStep = game.StepBirdsong
	case game.ActionSlip:
		state.TurnProgress.BirdsongMainActionTaken = true
		state.CurrentPhase = game.Birdsong
		state.CurrentStep = game.StepBirdsong
		state.TurnProgress.HasSlipped = true
	case game.ActionRecruit:
		if action.Recruit != nil && action.Recruit.Faction == game.Alliance {
			state.TurnProgress.EveningMainActionTaken = true
		} else {
			state.TurnProgress.DaylightMainActionTaken = true
		}
		if action.Recruit != nil && action.Recruit.Faction == game.Alliance {
			state.CurrentPhase = game.Evening
			state.CurrentStep = game.StepEvening
			state.TurnProgress.OfficerActionsUsed++
		} else {
			state.CurrentStep = game.StepDaylightActions
		}
		if action.Recruit != nil && action.Recruit.Faction == game.Marquise {
			state.TurnProgress.ActionsUsed++
		}
		if action.Recruit != nil && action.Recruit.Faction == game.Eyrie {
			markResolvedDecreeCard(state, action.Recruit.DecreeCardID)
		}
	case game.ActionMovement:
		if action.Movement != nil && action.Movement.SourceEffectID == "" {
			switch state.CurrentPhase {
			case game.Daylight:
				state.TurnProgress.DaylightMainActionTaken = true
			case game.Evening:
				state.TurnProgress.EveningMainActionTaken = true
			}
		} else if action.Movement != nil && action.Movement.SourceEffectID != "" {
			markPersistentEffectUsed(state, action.Movement.SourceEffectID)
		}
		if action.Movement != nil && action.Movement.Faction == game.Alliance && action.Movement.SourceEffectID == "" {
			state.CurrentPhase = game.Evening
			state.CurrentStep = game.StepEvening
			state.TurnProgress.OfficerActionsUsed++
		} else if state.CurrentPhase == game.Evening {
			state.CurrentPhase = game.Evening
			state.CurrentStep = game.StepEvening
		} else {
			state.CurrentStep = game.StepDaylightActions
		}
		if action.Movement != nil && action.Movement.Faction == game.Marquise && action.Movement.SourceEffectID == "" {
			state.TurnProgress.ActionsUsed++
			state.TurnProgress.MarchesUsed++
		}
		if action.Movement != nil && action.Movement.Faction == game.Eyrie && action.Movement.SourceEffectID == "" {
			markResolvedDecreeCard(state, action.Movement.DecreeCardID)
		}
	case game.ActionBattleResolution, game.ActionBuild, game.ActionOverwork:
		state.CurrentStep = game.StepDaylightActions
		if action.Type == game.ActionBattleResolution && action.BattleResolution != nil {
			if action.BattleResolution.SourceEffectID == "" {
				if action.BattleResolution.Faction == game.Alliance {
					state.TurnProgress.EveningMainActionTaken = true
				} else {
					state.TurnProgress.DaylightMainActionTaken = true
				}
			} else {
				markPersistentEffectUsed(state, action.BattleResolution.SourceEffectID)
			}
		}
		if action.Type == game.ActionBuild || action.Type == game.ActionOverwork {
			state.TurnProgress.DaylightMainActionTaken = true
		}
		if action.Type == game.ActionBattleResolution &&
			action.BattleResolution != nil &&
			action.BattleResolution.Faction == game.Alliance &&
			action.BattleResolution.SourceEffectID == "" {
			state.CurrentPhase = game.Evening
			state.CurrentStep = game.StepEvening
			state.TurnProgress.OfficerActionsUsed++
		}
		switch {
		case action.Type == game.ActionBattleResolution &&
			action.BattleResolution != nil &&
			action.BattleResolution.Faction == game.Marquise &&
			action.BattleResolution.SourceEffectID == "":
			state.TurnProgress.ActionsUsed++
		case action.Type == game.ActionBattleResolution &&
			action.BattleResolution != nil &&
			action.BattleResolution.Faction == game.Eyrie &&
			action.BattleResolution.SourceEffectID == "":
			markResolvedDecreeCard(state, action.BattleResolution.DecreeCardID)
		case action.Type == game.ActionBuild && action.Build != nil && action.Build.Faction == game.Marquise:
			state.TurnProgress.ActionsUsed++
		case action.Type == game.ActionBuild && action.Build != nil && action.Build.Faction == game.Eyrie:
			markResolvedDecreeCard(state, action.Build.DecreeCardID)
		case action.Type == game.ActionOverwork:
			state.TurnProgress.ActionsUsed++
		}
	case game.ActionCraft:
		if action.Craft != nil &&
			game.DaylightEntryStep(action.Craft.Faction) == game.StepDaylightCraft &&
			effectiveStep(*state) == game.StepDaylightCraft {
			state.CurrentPhase = game.Daylight
			state.CurrentStep = game.StepDaylightCraft
		} else {
			state.TurnProgress.DaylightMainActionTaken = true
			state.CurrentStep = game.StepDaylightActions
		}
		state.TurnProgress.HasCrafted = true
	case game.ActionExplore, game.ActionAid, game.ActionQuest, game.ActionStrike, game.ActionRepair:
		state.TurnProgress.DaylightMainActionTaken = true
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.StepDaylightActions
	case game.ActionSpreadSympathy:
		state.TurnProgress.BirdsongMainActionTaken = true
		state.TurnProgress.SpreadSympathyStarted = true
		state.CurrentPhase = game.Birdsong
		state.CurrentStep = game.StepBirdsong
	case game.ActionRevolt:
		state.TurnProgress.BirdsongMainActionTaken = true
		state.CurrentPhase = game.Birdsong
		state.CurrentStep = game.StepBirdsong
	case game.ActionMobilize, game.ActionTrain:
		state.TurnProgress.DaylightMainActionTaken = true
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.StepDaylightActions
	case game.ActionOrganize:
		state.TurnProgress.EveningMainActionTaken = true
		state.CurrentPhase = game.Evening
		state.CurrentStep = game.StepEvening
		state.TurnProgress.OfficerActionsUsed++
	case game.ActionTurmoil:
		state.TurnProgress.DaylightMainActionTaken = true
		state.CurrentPhase = game.Evening
		state.CurrentStep = game.StepEvening
		state.TurnProgress.DecreeColumnsResolved = 0
		state.TurnProgress.DecreeCardsResolved = 0
	case game.ActionScoreRoosts:
		state.TurnProgress.EveningMainActionTaken = true
		state.CurrentPhase = game.Evening
		state.CurrentStep = game.StepEvening
	case game.ActionPassPhase:
		switch state.CurrentPhase {
		case game.Birdsong:
			if state.FactionTurn == game.Vagabond && !state.TurnProgress.HasSlipped {
				state.CurrentPhase = game.Birdsong
				state.CurrentStep = game.StepBirdsong
				return
			}
			state.CurrentPhase = game.Daylight
			state.CurrentStep = game.DaylightEntryStep(state.FactionTurn)
		case game.Daylight:
			if effectiveStep(*state) == game.StepDaylightCraft {
				state.CurrentStep = game.StepDaylightActions
			} else {
				state.CurrentPhase = game.Evening
				state.CurrentStep = game.StepEvening
			}
		case game.Evening:
			beginNextFactionTurn(state)
		}
	case game.ActionEveningDraw:
		state.TurnProgress.EveningMainActionTaken = true
		beginNextFactionTurn(state)
	case game.ActionDiscardEffect:
		state.CurrentStep = game.StepDaylightActions
	case game.ActionUsePersistentEffect:
		if action.UsePersistentEffect != nil && !effectStartsPhase(action.UsePersistentEffect.EffectID) {
			switch state.CurrentPhase {
			case game.Birdsong:
				state.TurnProgress.BirdsongMainActionTaken = true
			case game.Daylight:
				state.TurnProgress.DaylightMainActionTaken = true
			case game.Evening:
				state.TurnProgress.EveningMainActionTaken = true
			}
		}
	}
}

func decreeCardsByColumn(decree game.Decree, column game.DecreeColumn) []game.CardID {
	switch column {
	case game.DecreeRecruit:
		return decree.Recruit
	case game.DecreeMove:
		return decree.Move
	case game.DecreeBattle:
		return decree.Battle
	case game.DecreeBuild:
		return decree.Build
	default:
		return nil
	}
}

func currentDecreeCard(state game.GameState) (game.DecreeColumn, game.CardID, bool) {
	for i := state.TurnProgress.DecreeColumnsResolved; i < 4; i++ {
		column := game.DecreeColumn(i)
		for _, cardID := range decreeCardsByColumn(state.Eyrie.Decree, column) {
			if !decreeCardResolved(state, cardID) {
				return column, cardID, true
			}
		}
	}

	return 0, 0, false
}

func decreeCardResolved(state game.GameState, cardID game.CardID) bool {
	for _, resolvedID := range state.TurnProgress.ResolvedDecreeCardIDs {
		if resolvedID == cardID {
			return true
		}
	}
	return false
}

func markResolvedDecreeCard(state *game.GameState, cardID game.CardID) {
	if cardID == 0 || decreeCardResolved(*state, cardID) {
		return
	}

	state.TurnProgress.ResolvedDecreeCardIDs = append(state.TurnProgress.ResolvedDecreeCardIDs, cardID)
	state.TurnProgress.DecreeCardsResolved = 0

	for state.TurnProgress.DecreeColumnsResolved < 4 {
		column := game.DecreeColumn(state.TurnProgress.DecreeColumnsResolved)
		allResolved := true
		for _, columnCardID := range decreeCardsByColumn(state.Eyrie.Decree, column) {
			if !decreeCardResolved(*state, columnCardID) {
				allResolved = false
				break
			}
		}
		if !allResolved {
			return
		}
		state.TurnProgress.DecreeColumnsResolved++
	}
}
