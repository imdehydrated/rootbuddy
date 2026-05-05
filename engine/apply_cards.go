package engine

import "github.com/imdehydrated/rootbuddy/game"

func applyCraft(state *game.GameState, action game.Action) {
	if action.Craft == nil {
		return
	}

	card, found := CardByID(action.Craft.CardID)
	if found && card.CraftedItem != nil && !DeductItem(state, *card.CraftedItem) {
		return
	}

	if _, ok := spendFactionHandCard(state, action.Craft.Faction, action.Craft.CardID); !ok {
		return
	}

	if action.Craft.Faction == game.Vagabond {
		exhaustReadyItemsByType(state, game.ItemHammer, len(action.Craft.UsedWorkshopClearings))
		if found && card.CraftedItem != nil {
			state.Vagabond.Items = append(state.Vagabond.Items, game.Item{
				Type:   *card.CraftedItem,
				Status: game.ItemReady,
			})
			state.Vagabond.Items[len(state.Vagabond.Items)-1] = game.NormalizeItemZone(state.Vagabond.Items[len(state.Vagabond.Items)-1])
		}
	}
	if found {
		resolveCraftedCard(state, action.Craft.Faction, card)
	} else {
		DiscardCard(state, action.Craft.CardID)
	}
	state.TurnProgress.UsedWorkshopClearings = append(
		state.TurnProgress.UsedWorkshopClearings,
		action.Craft.UsedWorkshopClearings...,
	)
}

func applyBirdsongWood(state *game.GameState, action game.Action) {
	if action.BirdsongWood == nil {
		return
	}

	for _, clearingID := range action.BirdsongWood.ClearingIDs {
		index := findClearingIndex(state.Map, clearingID)
		if index == -1 {
			continue
		}

		state.Map.Clearings[index].Wood += action.BirdsongWood.Amount
	}
}

func applyEveningDraw(state *game.GameState, action game.Action) {
	if action.EveningDraw == nil {
		return
	}

	DrawCards(state, action.EveningDraw.Faction, action.EveningDraw.Count)
}

func applyEveningDiscard(state *game.GameState, action game.Action) {
	if action.EveningDiscard == nil {
		return
	}

	for _, cardID := range action.EveningDiscard.CardIDs {
		if _, ok := spendFactionHandCard(state, action.EveningDiscard.Faction, cardID); ok {
			DiscardCard(state, cardID)
		}
	}
}

func applyScoreRoosts(state *game.GameState, action game.Action) {
	if action.ScoreRoosts == nil {
		return
	}

	addVictoryPoints(state, action.ScoreRoosts.Faction, action.ScoreRoosts.Points)
}

func applyPassPhase(state *game.GameState, action game.Action) {
	if action.PassPhase == nil {
		return
	}
}

func applyMarquiseExtraAction(state *game.GameState, action game.Action) {
	if action.MarquiseExtraAction == nil || action.MarquiseExtraAction.Faction != game.Marquise {
		return
	}

	card, ok := game.Card{}, false
	for _, handCard := range state.Marquise.CardsInHand {
		if handCard.ID == action.MarquiseExtraAction.CardID {
			card = handCard
			ok = true
			break
		}
	}
	if !ok && canUseObservedHiddenCards(*state, game.Marquise) {
		card, ok = CardByID(action.MarquiseExtraAction.CardID)
	}
	if !ok || card.Suit != game.Bird {
		return
	}
	if _, ok := spendFactionHandCard(state, game.Marquise, action.MarquiseExtraAction.CardID); !ok {
		return
	}

	DiscardCard(state, action.MarquiseExtraAction.CardID)
	state.TurnProgress.BonusActions++
}

func applyAddCardToHand(state *game.GameState, action game.Action) {
	if action.AddCardToHand == nil {
		return
	}

	card, ok := CardByID(action.AddCardToHand.CardID)
	if !ok {
		return
	}

	appendCardToFactionHand(state, action.AddCardToHand.Faction, card)
}

func applyRemoveCardFromHand(state *game.GameState, action game.Action) {
	if action.RemoveCardFromHand == nil {
		return
	}

	if _, ok := spendFactionHandCard(state, action.RemoveCardFromHand.Faction, action.RemoveCardFromHand.CardID); !ok {
		return
	}

	DiscardCard(state, action.RemoveCardFromHand.CardID)
}

func applyOtherPlayerDraw(state *game.GameState, action game.Action) {
	if action.OtherPlayerDraw == nil {
		return
	}

	if state.GameMode == game.GameModeOnline {
		DrawCards(state, action.OtherPlayerDraw.Faction, action.OtherPlayerDraw.Count)
		return
	}

	incrementOtherHandCount(state, action.OtherPlayerDraw.Faction, action.OtherPlayerDraw.Count)
}

func applyOtherPlayerPlay(state *game.GameState, action game.Action) {
	if action.OtherPlayerPlay == nil {
		return
	}

	decrementOtherHandCount(state, action.OtherPlayerPlay.Faction, 1)
	DiscardCard(state, action.OtherPlayerPlay.CardID)
}

func applyDiscardEffect(state *game.GameState, action game.Action) {
	if action.DiscardEffect == nil {
		return
	}
	if !removePersistentEffect(state, action.DiscardEffect.Faction, action.DiscardEffect.CardID) {
		return
	}

	DiscardCard(state, action.DiscardEffect.CardID)
}

func applyActivateDominance(state *game.GameState, action game.Action) {
	if action.ActivateDominance == nil {
		return
	}
	if hasActiveDominance(*state, action.ActivateDominance.Faction) {
		return
	}

	if _, ok := spendFactionHandCard(state, action.ActivateDominance.Faction, action.ActivateDominance.CardID); !ok {
		return
	}

	if state.ActiveDominance == nil {
		state.ActiveDominance = map[game.Faction]game.CardID{}
	}
	state.ActiveDominance[action.ActivateDominance.Faction] = action.ActivateDominance.CardID

	if action.ActivateDominance.Faction == game.Vagabond {
		state.CoalitionActive = true
		state.CoalitionPartner = action.ActivateDominance.TargetFaction
	}
}

func applyTakeDominance(state *game.GameState, action game.Action) {
	if action.TakeDominance == nil {
		return
	}

	if !removeAvailableDominance(state, action.TakeDominance.DominanceCardID) {
		return
	}
	if _, ok := spendFactionHandCard(state, action.TakeDominance.Faction, action.TakeDominance.SpentCardID); !ok {
		addAvailableDominance(state, action.TakeDominance.DominanceCardID)
		return
	}

	DiscardCard(state, action.TakeDominance.SpentCardID)
	card, ok := CardByID(action.TakeDominance.DominanceCardID)
	if !ok {
		addAvailableDominance(state, action.TakeDominance.DominanceCardID)
		return
	}
	appendCardToFactionHand(state, action.TakeDominance.Faction, card)
}
