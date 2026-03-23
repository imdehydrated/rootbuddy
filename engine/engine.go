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
		return game.StepBirdsong
	case game.Daylight:
		return game.StepDaylightActions
	case game.Evening:
		return game.StepEvening
	default:
		return game.StepUnspecified
	}
}

func ValidActions(state game.GameState) []game.Action {
	switch state.FactionTurn {
	case game.Marquise:
		return validMarquiseActions(state)
	case game.Eyrie:
		return validEyrieActions(state)
	default:
		return []game.Action{}
	}
}

func validMarquiseActions(state game.GameState) []game.Action {
	switch effectiveStep(state) {
	case game.StepBirdsong:
		return rules.ValidMarquiseBirdsongWoodActions(state)
	case game.StepDaylightCraft:
		actions := rules.ValidCraftActions(state)
		actions = append(actions, game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Marquise,
			},
		})
		return actions
	case game.StepDaylightActions:
		if state.TurnProgress.ActionsUsed >= 3+state.TurnProgress.BonusActions {
			return []game.Action{
				{
					Type: game.ActionPassPhase,
					PassPhase: &game.PassPhaseAction{
						Faction: game.Marquise,
					},
				},
			}
		}

		actions := []game.Action{}
		actions = append(actions, rules.ValidRecruitActions(state)...)
		if state.TurnProgress.MarchesUsed < 2 {
			actions = append(actions, rules.ValidMovementActions(game.Marquise, state.Map)...)
		}
		actions = append(actions, rules.ValidBattles(game.Marquise, state.Map)...)
		actions = append(actions, rules.ValidBuilds(state.Map, state.Marquise)...)
		actions = append(actions, rules.ValidOverworkActions(state)...)
		actions = append(actions, game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Marquise,
			},
		})
		return actions
	case game.StepEvening:
		return rules.ValidMarquiseEveningActions(state)
	default:
		return []game.Action{}
	}
}

func validEyrieActions(state game.GameState) []game.Action {
	switch effectiveStep(state) {
	case game.StepBirdsong:
		return rules.ValidAddToDecreeActions(state)
	case game.StepDaylightCraft:
		actions := rules.ValidEyrieCraftActions(state)
		actions = append(actions, game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Eyrie,
			},
		})
		return actions
	case game.StepDaylightActions:
		return rules.ValidEyrieDaylightActions(state)
	case game.StepEvening:
		return rules.ValidEyrieEveningActions(state)
	default:
		return []game.Action{}
	}
}
