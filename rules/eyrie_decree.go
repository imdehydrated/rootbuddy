package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidEyrieEmergencyOrdersActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Eyrie || state.CurrentPhase != game.Birdsong {
		return nil
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepBirdsong {
		return nil
	}

	if state.TurnProgress.CardsAddedToDecree > 0 {
		return nil
	}

	if state.TurnProgress.EyrieEmergencyResolved || len(state.Eyrie.CardsInHand) > 0 {
		return nil
	}

	return []game.Action{
		{
			Type: game.ActionEyrieEmergencyOrders,
			EyrieEmergency: &game.EyrieEmergencyOrdersAction{
				Faction: game.Eyrie,
				Count:   1,
			},
		},
	}
}

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

	if state.TurnProgress.CardsAddedToDecree > 0 {
		return []game.Action{}
	}

	if len(ValidEyrieEmergencyOrdersActions(state)) > 0 {
		return []game.Action{}
	}

	if len(state.Eyrie.CardsInHand) == 0 {
		return []game.Action{}
	}

	return addToDecreeActionsForCards(state.Eyrie.CardsInHand)
}

func ValidEyrieNewRoostActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Eyrie || state.CurrentPhase != game.Birdsong {
		return nil
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepBirdsong {
		return nil
	}

	if state.TurnProgress.CardsAddedToDecree == 0 || state.TurnProgress.EyrieNewRoostResolved || eyrieHasRoost(state) {
		return nil
	}

	if state.Eyrie.WarriorSupply < 3 || state.Eyrie.RoostsPlaced >= 7 {
		return nil
	}

	minWarriors := 0
	candidates := []game.Clearing{}
	for _, clearing := range state.Map.Clearings {
		if !hasOpenBuildSlot(clearing) {
			continue
		}

		warriors := totalWarriors(clearing)
		if len(candidates) == 0 || warriors < minWarriors {
			minWarriors = warriors
			candidates = []game.Clearing{clearing}
			continue
		}
		if warriors == minWarriors {
			candidates = append(candidates, clearing)
		}
	}

	actions := make([]game.Action, 0, len(candidates))
	for _, clearing := range candidates {
		actions = append(actions, game.Action{
			Type: game.ActionEyrieNewRoost,
			EyrieNewRoost: &game.EyrieNewRoostAction{
				Faction:    game.Eyrie,
				ClearingID: clearing.ID,
			},
		})
	}

	return actions
}

func ValidEyrieBirdsongActions(state game.GameState) []game.Action {
	if emergencyActions := ValidEyrieEmergencyOrdersActions(state); len(emergencyActions) > 0 {
		return emergencyActions
	}

	if decreeActions := ValidAddToDecreeActions(state); len(decreeActions) > 0 {
		return decreeActions
	}

	if newRoostActions := ValidEyrieNewRoostActions(state); len(newRoostActions) > 0 {
		return newRoostActions
	}

	if state.FactionTurn == game.Eyrie && state.CurrentPhase == game.Birdsong &&
		(state.CurrentStep == game.StepUnspecified || state.CurrentStep == game.StepBirdsong) &&
		state.TurnProgress.CardsAddedToDecree > 0 {
		return []game.Action{
			{
				Type: game.ActionPassPhase,
				PassPhase: &game.PassPhaseAction{
					Faction: game.Eyrie,
				},
			},
		}
	}

	return nil
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
