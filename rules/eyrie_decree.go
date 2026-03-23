package rules

import "github.com/imdehydrated/rootbuddy/game"

func addToDecreeActionsForCards(cards []game.Card) []game.Action {
	actions := []game.Action{}

	appendAssignments := func(selected []game.Card) {
		columns := make([]game.DecreeColumn, len(selected))
		cardIDs := make([]game.CardID, len(selected))
		for i, card := range selected {
			cardIDs[i] = card.ID
		}

		var assign func(index int)
		assign = func(index int) {
			if index == len(selected) {
				actionColumns := make([]game.DecreeColumn, len(columns))
				copy(actionColumns, columns)
				actionCardIDs := make([]game.CardID, len(cardIDs))
				copy(actionCardIDs, cardIDs)
				actions = append(actions, game.Action{
					Type: game.ActionAddToDecree,
					AddToDecree: &game.AddToDecreeAction{
						Faction: game.Eyrie,
						CardIDs: actionCardIDs,
						Columns: actionColumns,
					},
				})
				return
			}

			for _, column := range decreeOrder {
				columns[index] = column
				assign(index + 1)
			}
		}

		assign(0)
	}

	for i, first := range cards {
		appendAssignments([]game.Card{first})
		for j := i + 1; j < len(cards); j++ {
			second := cards[j]
			birdCount := 0
			if first.Suit == game.Bird {
				birdCount++
			}
			if second.Suit == game.Bird {
				birdCount++
			}
			if birdCount > 1 {
				continue
			}
			appendAssignments([]game.Card{first, second})
		}
	}

	return actions
}

func ValidAddToDecreeActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Eyrie {
		return []game.Action{}
	}

	if state.CurrentPhase != game.Birdsong {
		return []game.Action{}
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepBirdsong {
		return []game.Action{}
	}

	if len(state.Eyrie.CardsInHand) == 0 {
		return []game.Action{}
	}

	return addToDecreeActionsForCards(state.Eyrie.CardsInHand)
}

func ValidEyrieDaylightActions(state game.GameState) []game.Action {
	column, cardIDs, ok := currentDecreeColumn(state)
	if !ok {
		return []game.Action{
			{
				Type: game.ActionPassPhase,
				PassPhase: &game.PassPhaseAction{
					Faction: game.Eyrie,
				},
			},
		}
	}

	actions := []game.Action{}
	for _, cardID := range cardIDs {
		switch column {
		case game.DecreeRecruit:
			actions = append(actions, ValidEyrieRecruitActions(state, cardID)...)
		case game.DecreeMove:
			actions = append(actions, ValidEyrieMovementActions(state, cardID)...)
		case game.DecreeBattle:
			actions = append(actions, ValidEyrieBattleActions(state, cardID)...)
		case game.DecreeBuild:
			actions = append(actions, ValidEyrieBuildActions(state, cardID)...)
		}
	}

	if len(actions) == 0 {
		return ValidEyrieTurmoilActions(state)
	}

	return actions
}
