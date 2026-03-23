package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidTrainActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Alliance || state.CurrentPhase != game.Daylight {
		return nil
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepDaylightActions {
		return nil
	}

	actions := []game.Action{}
	for _, card := range state.Alliance.CardsInHand {
		if card.Suit == game.Bird {
			if !allianceHasAnyBase(state) {
				continue
			}
		} else if !allianceHasBaseInSuit(state, card.Suit) {
			continue
		}

		actions = append(actions, game.Action{
			Type: game.ActionTrain,
			Train: &game.TrainAction{
				Faction: game.Alliance,
				CardID:  card.ID,
			},
		})
	}

	return actions
}
