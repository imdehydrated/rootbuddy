package engine

import "github.com/imdehydrated/rootbuddy/game"

func cloneState(state game.GameState) game.GameState {
	next := state

	next.Map.Clearings = make([]game.Clearing, len(state.Map.Clearings))
	for i, clearing := range state.Map.Clearings {
		cloned := clearing

		if clearing.Adj != nil {
			cloned.Adj = make([]int, len(clearing.Adj))
			copy(cloned.Adj, clearing.Adj)
		}

		if clearing.RuinItems != nil {
			cloned.RuinItems = make([]game.ItemType, len(clearing.RuinItems))
			copy(cloned.RuinItems, clearing.RuinItems)
		}

		if clearing.Warriors != nil {
			cloned.Warriors = make(map[game.Faction]int, len(clearing.Warriors))
			for faction, count := range clearing.Warriors {
				cloned.Warriors[faction] = count
			}
		}

		if clearing.Buildings != nil {
			cloned.Buildings = make([]game.Building, len(clearing.Buildings))
			copy(cloned.Buildings, clearing.Buildings)
		}

		if clearing.Tokens != nil {
			cloned.Tokens = make([]game.Token, len(clearing.Tokens))
			copy(cloned.Tokens, clearing.Tokens)
		}

		next.Map.Clearings[i] = cloned
	}

	if state.Map.Forests != nil {
		next.Map.Forests = make([]game.Forest, len(state.Map.Forests))
		for i, forest := range state.Map.Forests {
			cloned := forest
			if forest.AdjacentClearings != nil {
				cloned.AdjacentClearings = make([]int, len(forest.AdjacentClearings))
				copy(cloned.AdjacentClearings, forest.AdjacentClearings)
			}
			next.Map.Forests[i] = cloned
		}
	}

	if state.TurnOrder != nil {
		next.TurnOrder = make([]game.Faction, len(state.TurnOrder))
		copy(next.TurnOrder, state.TurnOrder)
	}

	if state.Deck != nil {
		next.Deck = make([]game.CardID, len(state.Deck))
		copy(next.Deck, state.Deck)
	}

	if state.DiscardPile != nil {
		next.DiscardPile = make([]game.CardID, len(state.DiscardPile))
		copy(next.DiscardPile, state.DiscardPile)
	}

	if state.AvailableDominance != nil {
		next.AvailableDominance = make([]game.CardID, len(state.AvailableDominance))
		copy(next.AvailableDominance, state.AvailableDominance)
	}

	if state.VictoryPoints != nil {
		next.VictoryPoints = make(map[game.Faction]int, len(state.VictoryPoints))
		for faction, points := range state.VictoryPoints {
			next.VictoryPoints[faction] = points
		}
	}

	if state.ActiveDominance != nil {
		next.ActiveDominance = make(map[game.Faction]game.CardID, len(state.ActiveDominance))
		for faction, cardID := range state.ActiveDominance {
			next.ActiveDominance[faction] = cardID
		}
	}

	if state.WinningCoalition != nil {
		next.WinningCoalition = make([]game.Faction, len(state.WinningCoalition))
		copy(next.WinningCoalition, state.WinningCoalition)
	}

	if state.ItemSupply != nil {
		next.ItemSupply = make(map[game.ItemType]int, len(state.ItemSupply))
		for itemType, count := range state.ItemSupply {
			next.ItemSupply[itemType] = count
		}
	}

	if state.PersistentEffects != nil {
		next.PersistentEffects = make(map[game.Faction][]game.CardID, len(state.PersistentEffects))
		for faction, cardIDs := range state.PersistentEffects {
			next.PersistentEffects[faction] = cloneCardIDs(cardIDs)
		}
	}

	if state.QuestDeck != nil {
		next.QuestDeck = make([]game.QuestID, len(state.QuestDeck))
		copy(next.QuestDeck, state.QuestDeck)
	}

	if state.QuestDiscard != nil {
		next.QuestDiscard = make([]game.QuestID, len(state.QuestDiscard))
		copy(next.QuestDiscard, state.QuestDiscard)
	}

	if state.OtherHandCounts != nil {
		next.OtherHandCounts = make(map[game.Faction]int, len(state.OtherHandCounts))
		for faction, count := range state.OtherHandCounts {
			next.OtherHandCounts[faction] = count
		}
	}

	if state.HiddenCards != nil {
		next.HiddenCards = make([]game.HiddenCard, len(state.HiddenCards))
		copy(next.HiddenCards, state.HiddenCards)
	}

	if state.PendingFieldHospitals != nil {
		next.PendingFieldHospitals = make([]game.FieldHospitalsPending, len(state.PendingFieldHospitals))
		copy(next.PendingFieldHospitals, state.PendingFieldHospitals)
	}

	if state.Marquise.CardsInHand != nil {
		next.Marquise.CardsInHand = make([]game.Card, len(state.Marquise.CardsInHand))
		copy(next.Marquise.CardsInHand, state.Marquise.CardsInHand)
	}

	if state.Eyrie.CardsInHand != nil {
		next.Eyrie.CardsInHand = make([]game.Card, len(state.Eyrie.CardsInHand))
		copy(next.Eyrie.CardsInHand, state.Eyrie.CardsInHand)
	}

	if state.Eyrie.AvailableLeaders != nil {
		next.Eyrie.AvailableLeaders = make([]game.EyrieLeader, len(state.Eyrie.AvailableLeaders))
		copy(next.Eyrie.AvailableLeaders, state.Eyrie.AvailableLeaders)
	}

	next.Eyrie.Decree.Recruit = cloneCardIDs(state.Eyrie.Decree.Recruit)
	next.Eyrie.Decree.Move = cloneCardIDs(state.Eyrie.Decree.Move)
	next.Eyrie.Decree.Battle = cloneCardIDs(state.Eyrie.Decree.Battle)
	next.Eyrie.Decree.Build = cloneCardIDs(state.Eyrie.Decree.Build)

	if state.Alliance.CardsInHand != nil {
		next.Alliance.CardsInHand = make([]game.Card, len(state.Alliance.CardsInHand))
		copy(next.Alliance.CardsInHand, state.Alliance.CardsInHand)
	}

	if state.Alliance.Supporters != nil {
		next.Alliance.Supporters = make([]game.Card, len(state.Alliance.Supporters))
		copy(next.Alliance.Supporters, state.Alliance.Supporters)
	}

	if state.Vagabond.CardsInHand != nil {
		next.Vagabond.CardsInHand = make([]game.Card, len(state.Vagabond.CardsInHand))
		copy(next.Vagabond.CardsInHand, state.Vagabond.CardsInHand)
	}

	if state.Vagabond.Items != nil {
		next.Vagabond.Items = make([]game.Item, len(state.Vagabond.Items))
		copy(next.Vagabond.Items, state.Vagabond.Items)
	}

	if state.Vagabond.Relationships != nil {
		next.Vagabond.Relationships = make(map[game.Faction]game.RelationshipLevel, len(state.Vagabond.Relationships))
		for faction, relationship := range state.Vagabond.Relationships {
			next.Vagabond.Relationships[faction] = relationship
		}
	}

	if state.Vagabond.QuestsCompleted != nil {
		next.Vagabond.QuestsCompleted = make([]game.Quest, len(state.Vagabond.QuestsCompleted))
		copy(next.Vagabond.QuestsCompleted, state.Vagabond.QuestsCompleted)
	}

	if state.Vagabond.QuestsAvailable != nil {
		next.Vagabond.QuestsAvailable = make([]game.Quest, len(state.Vagabond.QuestsAvailable))
		copy(next.Vagabond.QuestsAvailable, state.Vagabond.QuestsAvailable)
	}

	if state.TurnProgress.UsedWorkshopClearings != nil {
		next.TurnProgress.UsedWorkshopClearings = make([]int, len(state.TurnProgress.UsedWorkshopClearings))
		copy(next.TurnProgress.UsedWorkshopClearings, state.TurnProgress.UsedWorkshopClearings)
	}

	if state.TurnProgress.ResolvedDecreeCardIDs != nil {
		next.TurnProgress.ResolvedDecreeCardIDs = make([]game.CardID, len(state.TurnProgress.ResolvedDecreeCardIDs))
		copy(next.TurnProgress.ResolvedDecreeCardIDs, state.TurnProgress.ResolvedDecreeCardIDs)
	}

	if state.TurnProgress.UsedPersistentEffectIDs != nil {
		next.TurnProgress.UsedPersistentEffectIDs = make([]string, len(state.TurnProgress.UsedPersistentEffectIDs))
		copy(next.TurnProgress.UsedPersistentEffectIDs, state.TurnProgress.UsedPersistentEffectIDs)
	}

	return next
}

func cloneCardIDs(cardIDs []game.CardID) []game.CardID {
	if cardIDs == nil {
		return nil
	}

	cloned := make([]game.CardID, len(cardIDs))
	copy(cloned, cardIDs)
	return cloned
}

func findClearingIndex(m game.Map, id int) int {
	for i, clearing := range m.Clearings {
		if clearing.ID == id {
			return i
		}
	}
	return -1
}

func removeCardByID(cards []game.Card, id game.CardID) []game.Card {
	for i, card := range cards {
		if card.ID == id {
			return append(cards[:i], cards[i+1:]...)
		}
	}
	return cards
}

func removeCardsByID(cards []game.Card, ids []game.CardID) []game.Card {
	remaining := cards
	for _, id := range ids {
		remaining = removeCardByID(remaining, id)
	}

	return remaining
}

func removeCardFromFactionHand(state *game.GameState, faction game.Faction, cardID game.CardID) (game.Card, bool) {
	var (
		remaining []game.Card
		card      game.Card
		ok        bool
	)

	switch faction {
	case game.Marquise:
		remaining, card, ok = removeCardFromCards(state.Marquise.CardsInHand, cardID)
		if ok {
			state.Marquise.CardsInHand = remaining
		}
	case game.Eyrie:
		remaining, card, ok = removeCardFromCards(state.Eyrie.CardsInHand, cardID)
		if ok {
			state.Eyrie.CardsInHand = remaining
		}
	case game.Alliance:
		remaining, card, ok = removeCardFromCards(state.Alliance.CardsInHand, cardID)
		if ok {
			state.Alliance.CardsInHand = remaining
		}
	case game.Vagabond:
		remaining, card, ok = removeCardFromCards(state.Vagabond.CardsInHand, cardID)
		if ok {
			state.Vagabond.CardsInHand = remaining
		}
	}

	return card, ok
}

func setAllianceBasePlaced(state *game.GameState, suit game.Suit, placed bool) {
	switch suit {
	case game.Fox:
		state.Alliance.FoxBasePlaced = placed
	case game.Rabbit:
		state.Alliance.RabbitBasePlaced = placed
	case game.Mouse:
		state.Alliance.MouseBasePlaced = placed
	}
}

func allianceHasAnyBase(state game.GameState) bool {
	return state.Alliance.FoxBasePlaced || state.Alliance.RabbitBasePlaced || state.Alliance.MouseBasePlaced
}

func addAllianceSupporter(state *game.GameState, card game.Card) bool {
	if !allianceHasAnyBase(*state) && len(state.Alliance.Supporters) >= 5 {
		return false
	}

	state.Alliance.Supporters = append(state.Alliance.Supporters, card)
	return true
}

func gainAllianceSupporter(state *game.GameState, card game.Card) {
	if !addAllianceSupporter(state, card) {
		DiscardCard(state, card.ID)
	}
}

func drawAllianceSupporter(state *game.GameState) {
	cardID, ok := drawOneCardID(state)
	if !ok {
		return
	}

	card, found := CardByID(cardID)
	if !found {
		return
	}
	gainAllianceSupporter(state, card)
}

func cardMatchesSuitOrBird(card game.Card, suit game.Suit) bool {
	return card.Suit == suit || card.Suit == game.Bird
}

func transferOutrageCard(state *game.GameState, faction game.Faction, suit game.Suit) {
	if faction == game.Alliance {
		return
	}

	var hand *[]game.Card
	switch faction {
	case game.Marquise:
		hand = &state.Marquise.CardsInHand
	case game.Eyrie:
		hand = &state.Eyrie.CardsInHand
	case game.Vagabond:
		hand = &state.Vagabond.CardsInHand
	default:
		return
	}

	for i, card := range *hand {
		if !cardMatchesSuitOrBird(card, suit) {
			continue
		}

		gainAllianceSupporter(state, card)
		*hand = append((*hand)[:i], (*hand)[i+1:]...)
		return
	}

	drawAllianceSupporter(state)
}

func hasAllianceSympathy(clearing game.Clearing) bool {
	for _, token := range clearing.Tokens {
		if token.Faction == game.Alliance && token.Type == game.TokenSympathy {
			return true
		}
	}

	return false
}
