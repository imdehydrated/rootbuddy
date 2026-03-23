package rules

import "github.com/imdehydrated/rootbuddy/game"

func canSpreadSympathy(clearing game.Clearing, state game.GameState) bool {
	if hasAllianceSympathy(clearing) || hasKeepToken(clearing) {
		return false
	}

	if state.Alliance.SympathyPlaced == 0 {
		return true
	}

	return adjacentToAllianceSympathy(clearing, state.Map)
}

func ValidSpreadSympathyActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Alliance || state.CurrentPhase != game.Birdsong {
		return nil
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepBirdsong {
		return nil
	}

	if state.Alliance.SympathyPlaced >= len(allianceSympathyTrack) {
		return nil
	}

	cost := allianceSupporterCost(state.Alliance.SympathyPlaced)
	actions := []game.Action{}

	for _, clearing := range state.Map.Clearings {
		if !canSpreadSympathy(clearing, state) {
			continue
		}

		matchingCards := allianceSupporterCardIDs(state, clearing.Suit)
		for _, supporterIDs := range supporterCardSubsets(matchingCards, cost) {
			actions = append(actions, game.Action{
				Type: game.ActionSpreadSympathy,
				SpreadSympathy: &game.SpreadSympathyAction{
					Faction:          game.Alliance,
					ClearingID:       clearing.ID,
					SupporterCardIDs: supporterIDs,
				},
			})
		}
	}

	return actions
}
