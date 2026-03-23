package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidEyrieBattleActions(state game.GameState, cardID game.CardID) []game.Action {
	actions := []game.Action{}

	for _, action := range ValidBattles(game.Eyrie, state.Map) {
		if action.Battle == nil {
			continue
		}

		clearing, ok := findClearingByID(state.Map, action.Battle.ClearingID)
		if !ok {
			continue
		}

		if decreeMatchesSuit(cardID, clearing.Suit) {
			action.Battle.DecreeCardID = cardID
			actions = append(actions, action)
		}
	}

	return actions
}
