package rules

import "github.com/imdehydrated/rootbuddy/game"

func allianceDrawCount(state game.GameState) int {
	return 1 + len(allianceBaseClearings(state))
}

func ValidAllianceRecruitActions(state game.GameState) []game.Action {
	baseClearings := allianceBaseClearings(state)
	if len(baseClearings) == 0 || state.Alliance.WarriorSupply <= 0 {
		return nil
	}

	clearingIDs := make([]int, 0, len(baseClearings))
	for _, clearing := range baseClearings {
		clearingIDs = append(clearingIDs, clearing.ID)
	}

	recruitCount := state.Alliance.WarriorSupply
	if recruitCount > len(clearingIDs) {
		recruitCount = len(clearingIDs)
	}

	actions := []game.Action{}
	for _, chosenClearings := range recruitClearingSubsets(clearingIDs, recruitCount) {
		actions = append(actions, game.Action{
			Type: game.ActionRecruit,
			Recruit: &game.RecruitAction{
				Faction:     game.Alliance,
				ClearingIDs: chosenClearings,
			},
		})
	}

	return actions
}

func ValidOrganizeActions(state game.GameState) []game.Action {
	if state.Alliance.SympathyPlaced >= len(allianceSympathyTrack) {
		return nil
	}

	actions := []game.Action{}
	for _, clearing := range state.Map.Clearings {
		if hasAllianceSympathy(clearing) || hasKeepToken(clearing) {
			continue
		}
		if clearing.Warriors[game.Alliance] <= 0 {
			continue
		}

		actions = append(actions, game.Action{
			Type: game.ActionOrganize,
			Organize: &game.OrganizeAction{
				Faction:    game.Alliance,
				ClearingID: clearing.ID,
			},
		})
	}

	return actions
}

func ValidAllianceEveningActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Alliance || state.CurrentPhase != game.Evening {
		return nil
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepEvening {
		return nil
	}

	if state.TurnProgress.OfficerActionsUsed > state.Alliance.Officers {
		return nil
	}

	if state.TurnProgress.EveningDrawn && !state.TurnProgress.EveningDiscardResolved {
		return eveningDiscardActions(state, game.Alliance)
	}

	actions := []game.Action{
		{
			Type: game.ActionEveningDraw,
			EveningDraw: &game.EveningDrawAction{
				Faction: game.Alliance,
				Count:   allianceDrawCount(state),
			},
		},
	}

	if state.TurnProgress.OfficerActionsUsed >= state.Alliance.Officers {
		return actions
	}

	actions = append(actions, ValidAllianceRecruitActions(state)...)
	actions = append(actions, ValidMovementActions(game.Alliance, state.Map)...)
	actions = append(actions, ValidBattlesInState(game.Alliance, state)...)
	actions = append(actions, ValidOrganizeActions(state)...)

	return actions
}
