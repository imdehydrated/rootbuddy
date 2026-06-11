package engine

import "github.com/imdehydrated/rootbuddy/game"

func queueOutrage(state *game.GameState, faction game.Faction, suit game.Suit) {
	if faction == game.Alliance {
		return
	}
	state.PendingOutrage = append(state.PendingOutrage, game.OutragePending{
		Faction: faction,
		Suit:    suit,
	})
}

func validOutrageActions(state game.GameState) []game.Action {
	if len(state.PendingOutrage) == 0 {
		return nil
	}

	pending := state.PendingOutrage[0]
	actions := []game.Action{}

	if tracksHandForFaction(state, pending.Faction) {
		for _, card := range factionHand(state, pending.Faction) {
			if !cardMatchesSuitOrBird(card, pending.Suit) {
				continue
			}
			actions = append(actions, game.Action{
				Type: game.ActionResolveOutrage,
				ResolveOutrage: &game.ResolveOutrageAction{
					Faction: pending.Faction,
					Suit:    pending.Suit,
					CardID:  card.ID,
				},
			})
		}
		if len(actions) > 0 {
			return actions
		}
	}

	return []game.Action{
		{
			Type: game.ActionResolveOutrage,
			ResolveOutrage: &game.ResolveOutrageAction{
				Faction:       pending.Faction,
				Suit:          pending.Suit,
				DrawSupporter: true,
			},
		},
	}
}

func applyResolveOutrage(state *game.GameState, action game.Action) {
	if action.ResolveOutrage == nil || len(state.PendingOutrage) == 0 {
		return
	}

	pending := state.PendingOutrage[0]
	if action.ResolveOutrage.Faction != pending.Faction || action.ResolveOutrage.Suit != pending.Suit {
		return
	}

	if action.ResolveOutrage.DrawSupporter {
		if tracksHandForFaction(*state, pending.Faction) && factionHasMatchingOutrageCard(*state, pending.Faction, pending.Suit) {
			return
		}
		drawAllianceSupporter(state)
		state.PendingOutrage = state.PendingOutrage[1:]
		return
	}

	card, ok := CardByID(action.ResolveOutrage.CardID)
	if !ok || !cardMatchesSuitOrBird(card, pending.Suit) {
		return
	}

	if tracksHandForFaction(*state, pending.Faction) {
		if _, ok := removeCardFromFactionHand(state, pending.Faction, card.ID); !ok {
			return
		}
	} else if !consumeHiddenCard(state, pending.Faction, game.HiddenCardZoneHand) {
		return
	}

	gainAllianceSupporter(state, card)
	state.PendingOutrage = state.PendingOutrage[1:]
}

func factionHasMatchingOutrageCard(state game.GameState, faction game.Faction, suit game.Suit) bool {
	for _, card := range factionHand(state, faction) {
		if cardMatchesSuitOrBird(card, suit) {
			return true
		}
	}
	return false
}
