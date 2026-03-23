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

	if marquiseActionLimitReached(state) {
		return []game.Action{MarquisePassPhaseAction()}
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
