package rules

import "github.com/imdehydrated/rootbuddy/game"

func MarquisePassPhaseAction() game.Action {
	return game.Action{
		Type: game.ActionPassPhase,
		PassPhase: &game.PassPhaseAction{
			Faction: game.Marquise,
		},
	}
}

func ValidMarquiseDaylightActions(state game.GameState) []game.Action {
	if !marquiseIsDaylightActionStep(state) {
		return nil
	}

	if marquiseActionLimitReached(state) && state.TurnProgress.MarchesUsed == 0 {
		actions := ValidMarquiseExtraActionActions(state)
		actions = append(actions, MarquisePassPhaseAction())
		return actions
	}

	actions := []game.Action{}
	actions = append(actions, ValidMarquiseRecruitActions(state)...)
	actions = append(actions, ValidMarquiseMovementActions(state)...)
	actions = append(actions, ValidMarquiseBattleActions(state)...)
	actions = append(actions, ValidMarquiseBuildActions(state)...)
	actions = append(actions, ValidMarquiseOverworkActions(state)...)
	actions = append(actions, MarquisePassPhaseAction())

	return actions
}

func ValidMarquiseExtraActionActions(state game.GameState) []game.Action {
	if !marquiseIsDaylightActionStep(state) || !marquiseActionLimitReached(state) || state.TurnProgress.MarchesUsed != 0 {
		return nil
	}

	actions := []game.Action{}
	for _, card := range state.Marquise.CardsInHand {
		if card.Suit != game.Bird {
			continue
		}
		actions = append(actions, game.Action{
			Type: game.ActionMarquiseExtraAction,
			MarquiseExtraAction: &game.MarquiseExtraActionAction{
				Faction: game.Marquise,
				CardID:  card.ID,
			},
		})
	}

	return actions
}
