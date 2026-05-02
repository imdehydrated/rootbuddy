package rules

import "github.com/imdehydrated/rootbuddy/game"

func vagabondDrawCount(state game.GameState) int {
	coinCount := len(vagabondItemIndexesInZone(state, game.ItemCoin, game.ItemZoneTrack))
	if coinCount > 3 {
		return 3
	}

	return 1 + coinCount
}

func vagabondDiscardCardActions(state game.GameState) []game.Action {
	excess := len(state.Vagabond.CardsInHand) - 5
	if excess <= 0 {
		return []game.Action{
			{
				Type: game.ActionVagabondDiscard,
				VagabondDiscard: &game.VagabondDiscardAction{
					Faction: game.Vagabond,
				},
			},
		}
	}

	cardIDs := make([]game.CardID, 0, len(state.Vagabond.CardsInHand))
	for _, card := range state.Vagabond.CardsInHand {
		cardIDs = append(cardIDs, card.ID)
	}

	actions := []game.Action{}
	for _, discardedIDs := range supporterCardSubsets(cardIDs, excess) {
		actions = append(actions, game.Action{
			Type: game.ActionVagabondDiscard,
			VagabondDiscard: &game.VagabondDiscardAction{
				Faction: game.Vagabond,
				CardIDs: discardedIDs,
			},
		})
	}

	return actions
}

func vagabondCapacityLimit(state game.GameState) int {
	bags := len(vagabondItemIndexesInZone(state, game.ItemBag, game.ItemZoneTrack))
	return 6 + bags*2
}

func vagabondItemCapacityActions(state game.GameState) []game.Action {
	capacityIndexes := vagabondItemIndexesInZones(state, game.ItemZoneSatchel, game.ItemZoneDamaged)
	excess := len(capacityIndexes) - vagabondCapacityLimit(state)
	if excess <= 0 {
		return []game.Action{
			{
				Type: game.ActionVagabondItemCapacity,
				VagabondCapacity: &game.VagabondItemCapacityAction{
					Faction: game.Vagabond,
				},
			},
		}
	}

	actions := []game.Action{}
	for _, itemIndexes := range chooseItemIndexSubsets(capacityIndexes, excess) {
		actions = append(actions, game.Action{
			Type: game.ActionVagabondItemCapacity,
			VagabondCapacity: &game.VagabondItemCapacityAction{
				Faction:     game.Vagabond,
				ItemIndexes: itemIndexes,
			},
		})
	}

	return actions
}

func ValidVagabondEveningActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Vagabond || state.CurrentPhase != game.Evening {
		return nil
	}

	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepEvening {
		return nil
	}

	if !state.TurnProgress.VagabondRestResolved {
		return []game.Action{{
			Type: game.ActionVagabondRest,
			VagabondRest: &game.VagabondRestAction{
				Faction: game.Vagabond,
			},
		}}
	}

	if !state.TurnProgress.VagabondEveningDrawn {
		return []game.Action{{
			Type: game.ActionEveningDraw,
			EveningDraw: &game.EveningDrawAction{
				Faction: game.Vagabond,
				Count:   vagabondDrawCount(state),
			},
		}}
	}

	if !state.TurnProgress.VagabondDiscardResolved {
		return vagabondDiscardCardActions(state)
	}

	if !state.TurnProgress.VagabondCapacityChecked {
		return vagabondItemCapacityActions(state)
	}

	return []game.Action{{
		Type: game.ActionPassPhase,
		PassPhase: &game.PassPhaseAction{
			Faction: game.Vagabond,
		},
	}}
}
