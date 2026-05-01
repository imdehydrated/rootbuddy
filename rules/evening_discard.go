package rules

import "github.com/imdehydrated/rootbuddy/game"

func factionHandCardIDs(state game.GameState, faction game.Faction) ([]game.CardID, bool) {
	cards := []game.Card{}
	switch faction {
	case game.Marquise:
		cards = state.Marquise.CardsInHand
	case game.Eyrie:
		cards = state.Eyrie.CardsInHand
	case game.Alliance:
		cards = state.Alliance.CardsInHand
	default:
		return nil, false
	}

	if !state.TrackAllHands && faction != state.PlayerFaction && len(cards) == 0 && state.OtherHandCounts[faction] > 0 {
		return nil, false
	}

	cardIDs := make([]game.CardID, 0, len(cards))
	for _, card := range cards {
		cardIDs = append(cardIDs, card.ID)
	}
	return cardIDs, true
}

func factionHandLimitSize(state game.GameState, faction game.Faction) int {
	cardIDs, known := factionHandCardIDs(state, faction)
	if known {
		return len(cardIDs)
	}
	return state.OtherHandCounts[faction]
}

func eveningDiscardActions(state game.GameState, faction game.Faction) []game.Action {
	excess := factionHandLimitSize(state, faction) - 5
	if excess <= 0 {
		return []game.Action{{
			Type: game.ActionEveningDiscard,
			EveningDiscard: &game.EveningDiscardAction{
				Faction: faction,
			},
		}}
	}

	cardIDs, known := factionHandCardIDs(state, faction)
	if !known {
		return nil
	}

	actions := []game.Action{}
	for _, discardedIDs := range supporterCardSubsets(cardIDs, excess) {
		actions = append(actions, game.Action{
			Type: game.ActionEveningDiscard,
			EveningDiscard: &game.EveningDiscardAction{
				Faction: faction,
				CardIDs: discardedIDs,
				Count:   len(discardedIDs),
			},
		})
	}

	return actions
}
