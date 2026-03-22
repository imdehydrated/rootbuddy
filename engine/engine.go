package engine

import (
	"github.com/imdehydrated/rootbuddy/game"
	"github.com/imdehydrated/rootbuddy/rules"
)

func effectiveStep(state game.GameState) game.TurnStep {
	if state.CurrentStep != game.StepUnspecified {
		return state.CurrentStep
	}

	switch state.CurrentPhase {
	case game.Birdsong:
		return game.StepRecruit
	case game.Daylight:
		return game.StepDaylightActions
	case game.Evening:
		return game.StepEvening
	default:
		return game.StepUnspecified
	}
}

func ValidActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Marquise {
		return []game.Action{}
	}

	switch effectiveStep(state) {
	case game.StepRecruit:
		recruitState := state
		recruitState.CurrentStep = game.StepRecruit
		return rules.ValidRecruitActions(recruitState)
	case game.StepDaylightActions:
		actions := []game.Action{}
		actions = append(actions, rules.ValidMovementActions(game.Marquise, state.Map)...)
		actions = append(actions, rules.ValidBattles(game.Marquise, state.Map)...)
		actions = append(actions, rules.ValidBuilds(state.Map, state.Marquise)...)
		actions = append(actions, rules.ValidOverworkActions(state)...)
		actions = append(actions, rules.ValidCraftActions(state)...)
		return actions
	default:
		return []game.Action{}
	}
}
