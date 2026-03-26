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

	if _, ok := removeCardFromFactionHand(state, action.Craft.Faction, action.Craft.CardID); !ok {
		return
	}

	if action.Craft.Faction == game.Vagabond {
		exhaustReadyItemsByType(state, game.ItemHammer, len(action.Craft.UsedWorkshopClearings))
		if found && card.CraftedItem != nil {
			state.Vagabond.Items = append(state.Vagabond.Items, game.Item{
				Type:   *card.CraftedItem,
				Status: game.ItemReady,
			})
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

	if action.EveningDraw.Faction == game.Vagabond && state.Vagabond.InForest {
		repairAllDamagedItems(state)
	}

	DrawCards(state, action.EveningDraw.Faction, action.EveningDraw.Count)
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

	if _, ok := removeCardFromFactionHand(state, action.RemoveCardFromHand.Faction, action.RemoveCardFromHand.CardID); !ok {
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
