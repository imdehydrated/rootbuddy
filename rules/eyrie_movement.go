package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidEyrieMovementActions(state game.GameState, cardID game.CardID) []game.Action {
	actions := []game.Action{}

	for _, action := range ValidMovementActions(game.Eyrie, state.Map) {
		if action.Movement == nil {
			continue
		}

		clearing, ok := findClearingByID(state.Map, action.Movement.From)
		if !ok {
			continue
		}

		if decreeMatchesSuit(cardID, clearing.Suit) {
			action.Movement.DecreeCardID = cardID
			actions = append(actions, action)
		}
	}

	return actions
}
