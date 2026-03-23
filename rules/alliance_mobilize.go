package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidMobilizeActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Alliance || state.CurrentPhase != game.Daylight {
		return nil
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepDaylightActions {
		return nil
	}

	if !allianceHasAnyBase(state) && len(state.Alliance.Supporters) >= 5 {
		return nil
	}

	actions := make([]game.Action, 0, len(state.Alliance.CardsInHand))
	for _, card := range state.Alliance.CardsInHand {
		actions = append(actions, game.Action{
			Type: game.ActionMobilize,
			Mobilize: &game.MobilizeAction{
				Faction: game.Alliance,
				CardID:  card.ID,
			},
		})
	}

	return actions
}
