package engine

import (
	"github.com/imdehydrated/rootbuddy/game"
	"github.com/imdehydrated/rootbuddy/rules"
)

func effectiveStep(state game.GameState) game.TurnStep {
	return state.TurnWindow().Step
}

func ValidActions(state game.GameState) []game.Action {
	if state.GamePhase == game.LifecycleSetup && state.SetupStage != game.SetupStageUnspecified {
		return ValidSetupActions(state)
	}

	switch state.FactionTurn {
	case game.Marquise:
		return validMarquiseActions(state)
	case game.Alliance:
		return validAllianceActions(state)
	case game.Eyrie:
		return validEyrieActions(state)
	case game.Vagabond:
		return validVagabondActions(state)
	default:
		return []game.Action{}
	}
}

func validMarquiseActions(state game.GameState) []game.Action {
	switch effectiveStep(state) {
	case game.StepBirdsong:
		actions := effectActions(state)
		return append(actions, rules.ValidMarquiseBirdsongWoodActions(state)...)
	case game.StepDaylightCraft:
		actions := rules.ValidCraftActions(state)
		actions = append(actions, effectActions(state)...)
		actions = append(actions, rules.MarquisePassPhaseAction())
		return actions
	case game.StepDaylightActions:
		actions := effectActions(state)
		return append(actions, rules.ValidMarquiseDaylightActions(state)...)
	case game.StepEvening:
		actions := effectActions(state)
		return append(actions, rules.ValidMarquiseEveningActions(state)...)
	default:
		return []game.Action{}
	}
}

func validEyrieActions(state game.GameState) []game.Action {
	switch effectiveStep(state) {
	case game.StepBirdsong:
		actions := effectActions(state)
		return append(actions, rules.ValidAddToDecreeActions(state)...)
	case game.StepDaylightCraft:
		actions := rules.ValidEyrieCraftActions(state)
		actions = append(actions, effectActions(state)...)
		actions = append(actions, game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Eyrie,
			},
		})
		return actions
	case game.StepDaylightActions:
		actions := effectActions(state)
		return append(actions, rules.ValidEyrieDaylightActions(state)...)
	case game.StepEvening:
		actions := effectActions(state)
		return append(actions, rules.ValidEyrieEveningActions(state)...)
	default:
		return []game.Action{}
	}
}

func validAllianceActions(state game.GameState) []game.Action {
	switch effectiveStep(state) {
	case game.StepBirdsong:
		actions := effectActions(state)
		actions = append(actions, rules.ValidRevoltActions(state)...)
		actions = append(actions, rules.ValidSpreadSympathyActions(state)...)
		actions = append(actions, game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Alliance,
			},
		})
		return actions
	case game.StepDaylightCraft:
		actions := rules.ValidAllianceCraftActions(state)
		actions = append(actions, effectActions(state)...)
		actions = append(actions, game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Alliance,
			},
		})
		return actions
	case game.StepDaylightActions:
		actions := effectActions(state)
		actions = append(actions, rules.ValidMobilizeActions(state)...)
		actions = append(actions, rules.ValidTrainActions(state)...)
		actions = append(actions, game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Alliance,
			},
		})
		return actions
	case game.StepEvening:
		actions := effectActions(state)
		return append(actions, rules.ValidAllianceEveningActions(state)...)
	default:
		return []game.Action{}
	}
}

func validVagabondActions(state game.GameState) []game.Action {
	switch effectiveStep(state) {
	case game.StepBirdsong:
		actions := effectActions(state)
		return append(actions, rules.ValidVagabondBirdsongActions(state)...)
	case game.StepDaylightCraft, game.StepDaylightActions:
		actions := effectActions(state)
		actions = append(actions, rules.ValidVagabondMoveActions(state)...)
		actions = append(actions, rules.ValidVagabondBattleActions(state)...)
		actions = append(actions, rules.ValidExploreActions(state)...)
		actions = append(actions, rules.ValidAidActions(state)...)
		actions = append(actions, rules.ValidQuestActions(state)...)
		actions = append(actions, rules.ValidStrikeActions(state)...)
		actions = append(actions, rules.ValidRepairActions(state)...)
		actions = append(actions, rules.ValidVagabondCraftActions(state)...)
		actions = append(actions, game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Vagabond,
			},
		})
		return actions
	case game.StepEvening:
		actions := effectActions(state)
		return append(actions, rules.ValidVagabondEveningActions(state)...)
	default:
		return []game.Action{}
	}
}
