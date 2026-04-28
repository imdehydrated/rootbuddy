package engine

import "github.com/imdehydrated/rootbuddy/game"

type actionHandler func(*game.GameState, game.Action)

var applyActionHandlers = map[game.ActionType]actionHandler{
	game.ActionRecruit:              applyRecruit,
	game.ActionMovement:             applyMovement,
	game.ActionBattleResolution:     applyBattleResolution,
	game.ActionBuild:                applyBuild,
	game.ActionOverwork:             applyOverwork,
	game.ActionCraft:                applyCraft,
	game.ActionDaybreak:             applyDaybreak,
	game.ActionSlip:                 applySlip,
	game.ActionExplore:              applyExplore,
	game.ActionAid:                  applyAid,
	game.ActionQuest:                applyQuest,
	game.ActionStrike:               applyStrike,
	game.ActionRepair:               applyRepair,
	game.ActionSpreadSympathy:       applySpreadSympathy,
	game.ActionRevolt:               applyRevolt,
	game.ActionMobilize:             applyMobilize,
	game.ActionTrain:                applyTrain,
	game.ActionOrganize:             applyOrganize,
	game.ActionAddToDecree:          applyAddToDecree,
	game.ActionTurmoil:              applyTurmoil,
	game.ActionBirdsongWood:         applyBirdsongWood,
	game.ActionEveningDraw:          applyEveningDraw,
	game.ActionScoreRoosts:          applyScoreRoosts,
	game.ActionPassPhase:            applyPassPhase,
	game.ActionAddCardToHand:        applyAddCardToHand,
	game.ActionRemoveCardFromHand:   applyRemoveCardFromHand,
	game.ActionOtherPlayerDraw:      applyOtherPlayerDraw,
	game.ActionOtherPlayerPlay:      applyOtherPlayerPlay,
	game.ActionDiscardEffect:        applyDiscardEffect,
	game.ActionActivateDominance:    applyActivateDominance,
	game.ActionTakeDominance:        applyTakeDominance,
	game.ActionMarquiseSetup:        applyMarquiseSetup,
	game.ActionEyrieSetup:           applyEyrieSetup,
	game.ActionVagabondSetup:        applyVagabondSetup,
	game.ActionEyrieEmergencyOrders: applyEyrieEmergencyOrders,
	game.ActionEyrieNewRoost:        applyEyrieNewRoost,
}

func ApplyAction(state game.GameState, action game.Action) game.GameState {
	next, _ := ApplyActionDetailed(state, action)
	return next
}

func ApplyActionDetailed(state game.GameState, action game.Action) (game.GameState, *game.EffectResult) {
	next := cloneState(state)
	materializeAssistHandPlaceholders(&next)
	var result *game.EffectResult

	switch action.Type {
	case game.ActionUsePersistentEffect:
		result = usePersistentEffect(&next, action)
	case game.ActionBattle:
	default:
		if handler, ok := applyActionHandlers[action.Type]; ok {
			handler(&next, action)
		}
	}

	advanceTurnState(&next, action)
	return next, result
}
