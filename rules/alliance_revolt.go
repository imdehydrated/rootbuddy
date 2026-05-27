package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidRevoltActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Alliance || state.CurrentPhase != game.Birdsong {
		return nil
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepBirdsong {
		return nil
	}

	if state.TurnProgress.SpreadSympathyStarted {
		return nil
	}

	actions := []game.Action{}
	for _, clearing := range state.Map.Clearings {
		if !hasAllianceSympathy(clearing) || allianceHasBaseInSuit(state, clearing.Suit) {
			continue
		}
		if !hasOpenBuildSlotAfterRevolt(clearing) {
			continue
		}

		supporterIDs := allianceSupporterCardIDs(state, clearing.Suit)
		for _, chosenSupporters := range supporterCardSubsets(supporterIDs, 2) {
			for _, damagedItemIndexes := range revoltVagabondDamageChoices(state, clearing.ID) {
				actions = append(actions, game.Action{
					Type: game.ActionRevolt,
					Revolt: &game.RevoltAction{
						Faction:                    game.Alliance,
						ClearingID:                 clearing.ID,
						BaseSuit:                   clearing.Suit,
						SupporterCardIDs:           chosenSupporters,
						DamagedVagabondItemIndexes: damagedItemIndexes,
					},
				})
			}
		}
	}

	return actions
}

func revoltVagabondDamageChoices(state game.GameState, clearingID int) [][]int {
	if state.Vagabond.InForest || state.Vagabond.ClearingID != clearingID || !game.AreEnemies(state, game.Alliance, game.Vagabond) {
		return [][]int{nil}
	}

	return vagabondDamageIndexChoices(state, 3)
}
