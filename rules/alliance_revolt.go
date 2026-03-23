package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidRevoltActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Alliance || state.CurrentPhase != game.Birdsong {
		return nil
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepBirdsong {
		return nil
	}

	actions := []game.Action{}
	for _, clearing := range state.Map.Clearings {
		if !hasAllianceSympathy(clearing) || allianceHasBaseInSuit(state, clearing.Suit) {
			continue
		}

		supporterIDs := allianceSupporterCardIDs(state, clearing.Suit)
		for _, chosenSupporters := range supporterCardSubsets(supporterIDs, 2) {
			actions = append(actions, game.Action{
				Type: game.ActionRevolt,
				Revolt: &game.RevoltAction{
					Faction:          game.Alliance,
					ClearingID:       clearing.ID,
					BaseSuit:         clearing.Suit,
					SupporterCardIDs: chosenSupporters,
				},
			})
		}
	}

	return actions
}
