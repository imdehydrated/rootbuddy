package engine

import "github.com/imdehydrated/rootbuddy/game"

func marquiseHasFieldHospitalsCard(state game.GameState, suit game.Suit) bool {
	for _, card := range state.Marquise.CardsInHand {
		if cardMatchesSuitOrBird(card, suit) {
			return true
		}
	}
	return false
}

func queueFieldHospitals(state *game.GameState, clearing game.Clearing, warriorCount int) {
	if warriorCount <= 0 || state.Marquise.KeepClearingID <= 0 {
		return
	}

	state.PendingFieldHospitals = append(state.PendingFieldHospitals, game.FieldHospitalsPending{
		ClearingID:   clearing.ID,
		Suit:         clearing.Suit,
		WarriorCount: warriorCount,
	})
}

func pruneFieldHospitalsPending(state *game.GameState) {
	if len(state.PendingFieldHospitals) == 0 {
		return
	}

	if state.Marquise.KeepClearingID <= 0 {
		state.PendingFieldHospitals = nil
		return
	}

	remaining := make([]game.FieldHospitalsPending, 0, len(state.PendingFieldHospitals))
	for _, pending := range state.PendingFieldHospitals {
		if pending.WarriorCount <= 0 || pending.ClearingID <= 0 {
			continue
		}
		if !marquiseHasFieldHospitalsCard(*state, pending.Suit) {
			continue
		}
		remaining = append(remaining, pending)
	}

	if len(remaining) == 0 {
		state.PendingFieldHospitals = nil
		return
	}
	state.PendingFieldHospitals = remaining
}

func validFieldHospitalsActions(state game.GameState) []game.Action {
	if len(state.PendingFieldHospitals) == 0 {
		return nil
	}

	pending := state.PendingFieldHospitals[0]
	actions := []game.Action{
		{
			Type: game.ActionFieldHospitals,
			FieldHospitals: &game.FieldHospitalsAction{
				Faction:    game.Marquise,
				ClearingID: pending.ClearingID,
				Decline:    true,
			},
		},
	}

	for _, card := range state.Marquise.CardsInHand {
		if !cardMatchesSuitOrBird(card, pending.Suit) {
			continue
		}
		actions = append(actions, game.Action{
			Type: game.ActionFieldHospitals,
			FieldHospitals: &game.FieldHospitalsAction{
				Faction:    game.Marquise,
				ClearingID: pending.ClearingID,
				CardID:     card.ID,
			},
		})
	}

	return actions
}

func applyFieldHospitals(state *game.GameState, action game.Action) {
	if action.FieldHospitals == nil || len(state.PendingFieldHospitals) == 0 {
		return
	}

	pending := state.PendingFieldHospitals[0]
	if action.FieldHospitals.Faction != game.Marquise || action.FieldHospitals.ClearingID != pending.ClearingID {
		return
	}

	if action.FieldHospitals.Decline {
		state.PendingFieldHospitals = state.PendingFieldHospitals[1:]
		pruneFieldHospitalsPending(state)
		return
	}

	card, ok := CardByID(action.FieldHospitals.CardID)
	if !ok || !cardMatchesSuitOrBird(card, pending.Suit) {
		return
	}
	if _, ok := removeCardFromFactionHand(state, game.Marquise, action.FieldHospitals.CardID); !ok {
		return
	}

	DiscardCard(state, action.FieldHospitals.CardID)
	keepIndex := findClearingIndex(state.Map, state.Marquise.KeepClearingID)
	if keepIndex != -1 && state.Marquise.WarriorSupply > 0 {
		placed := pending.WarriorCount
		if placed > state.Marquise.WarriorSupply {
			placed = state.Marquise.WarriorSupply
		}
		if state.Map.Clearings[keepIndex].Warriors == nil {
			state.Map.Clearings[keepIndex].Warriors = map[game.Faction]int{}
		}
		state.Map.Clearings[keepIndex].Warriors[game.Marquise] += placed
		state.Marquise.WarriorSupply -= placed
	}

	state.PendingFieldHospitals = state.PendingFieldHospitals[1:]
	pruneFieldHospitalsPending(state)
}
