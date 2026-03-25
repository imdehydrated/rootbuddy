package engine

import (
	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

func ApplyAction(state game.GameState, action game.Action) game.GameState {
	next := cloneState(state)

	switch action.Type {
	case game.ActionRecruit:
		applyRecruit(&next, action)
	case game.ActionMovement:
		applyMovement(&next, action)
	case game.ActionBattleResolution:
		applyBattleResolution(&next, action)
	case game.ActionBuild:
		applyBuild(&next, action)
	case game.ActionOverwork:
		applyOverwork(&next, action)
	case game.ActionCraft:
		applyCraft(&next, action)
	case game.ActionDaybreak:
		applyDaybreak(&next, action)
	case game.ActionSlip:
		applySlip(&next, action)
	case game.ActionExplore:
		applyExplore(&next, action)
	case game.ActionAid:
		applyAid(&next, action)
	case game.ActionQuest:
		applyQuest(&next, action)
	case game.ActionStrike:
		applyStrike(&next, action)
	case game.ActionRepair:
		applyRepair(&next, action)
	case game.ActionSpreadSympathy:
		applySpreadSympathy(&next, action)
	case game.ActionRevolt:
		applyRevolt(&next, action)
	case game.ActionMobilize:
		applyMobilize(&next, action)
	case game.ActionTrain:
		applyTrain(&next, action)
	case game.ActionOrganize:
		applyOrganize(&next, action)
	case game.ActionAddToDecree:
		applyAddToDecree(&next, action)
	case game.ActionTurmoil:
		applyTurmoil(&next, action)
	case game.ActionBirdsongWood:
		applyBirdsongWood(&next, action)
	case game.ActionEveningDraw:
		applyEveningDraw(&next, action)
	case game.ActionScoreRoosts:
		applyScoreRoosts(&next, action)
	case game.ActionPassPhase:
		applyPassPhase(&next, action)
	case game.ActionAddCardToHand:
		applyAddCardToHand(&next, action)
	case game.ActionRemoveCardFromHand:
		applyRemoveCardFromHand(&next, action)
	case game.ActionOtherPlayerDraw:
		applyOtherPlayerDraw(&next, action)
	case game.ActionOtherPlayerPlay:
		applyOtherPlayerPlay(&next, action)
	}

	advanceTurnState(&next, action)

	return next
}

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

	if state.VictoryPoints != nil {
		next.VictoryPoints = make(map[game.Faction]int, len(state.VictoryPoints))
		for faction, points := range state.VictoryPoints {
			next.VictoryPoints[faction] = points
		}
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

func addAllianceSupporter(state *game.GameState, card game.Card) {
	if !allianceHasAnyBase(*state) && len(state.Alliance.Supporters) >= 5 {
		return
	}

	state.Alliance.Supporters = append(state.Alliance.Supporters, card)
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

		addAllianceSupporter(state, card)
		*hand = append((*hand)[:i], (*hand)[i+1:]...)
		return
	}
}

func hasAllianceSympathy(clearing game.Clearing) bool {
	for _, token := range clearing.Tokens {
		if token.Faction == game.Alliance && token.Type == game.TokenSympathy {
			return true
		}
	}

	return false
}

func applyRecruit(state *game.GameState, action game.Action) {
	if action.Recruit == nil {
		return
	}

	for _, clearingID := range action.Recruit.ClearingIDs {
		index := findClearingIndex(state.Map, clearingID)
		if index == -1 {
			continue
		}

		if state.Map.Clearings[index].Warriors == nil {
			state.Map.Clearings[index].Warriors = map[game.Faction]int{}
		}

		state.Map.Clearings[index].Warriors[action.Recruit.Faction]++
		switch action.Recruit.Faction {
		case game.Marquise:
			state.Marquise.WarriorSupply--
		case game.Eyrie:
			state.Eyrie.WarriorSupply--
		case game.Alliance:
			state.Alliance.WarriorSupply--
		}
	}

	if action.Recruit.Faction == game.Marquise {
		state.TurnProgress.RecruitUsed = true
	}
}

func applyMovement(state *game.GameState, action game.Action) {
	if action.Movement == nil {
		return
	}

	if action.Movement.Faction == game.Vagabond {
		if action.Movement.ToForestID != 0 {
			state.Vagabond.ClearingID = 0
			state.Vagabond.ForestID = action.Movement.ToForestID
			state.Vagabond.InForest = true
		} else {
			state.Vagabond.ClearingID = action.Movement.To
			state.Vagabond.ForestID = 0
			state.Vagabond.InForest = false
		}
		exhaustReadyItemsByType(state, game.ItemBoots, max(1, action.Movement.Count))

		toIndex := findClearingIndex(state.Map, action.Movement.To)
		if !state.Vagabond.InForest && toIndex != -1 && hasAllianceSympathy(state.Map.Clearings[toIndex]) {
			transferOutrageCard(state, action.Movement.Faction, state.Map.Clearings[toIndex].Suit)
		}
		return
	}

	fromIndex := findClearingIndex(state.Map, action.Movement.From)
	toIndex := findClearingIndex(state.Map, action.Movement.To)
	if fromIndex == -1 || toIndex == -1 {
		return
	}

	if state.Map.Clearings[fromIndex].Warriors == nil {
		return
	}

	moved := action.Movement.Count
	if moved <= 0 {
		moved = action.Movement.MaxCount
	}
	state.Map.Clearings[fromIndex].Warriors[action.Movement.Faction] -= moved

	if state.Map.Clearings[toIndex].Warriors == nil {
		state.Map.Clearings[toIndex].Warriors = map[game.Faction]int{}
	}
	state.Map.Clearings[toIndex].Warriors[action.Movement.Faction] += moved

	if action.Movement.Faction != game.Alliance && hasAllianceSympathy(state.Map.Clearings[toIndex]) {
		transferOutrageCard(state, action.Movement.Faction, state.Map.Clearings[toIndex].Suit)
	}
}

func removeWarriorLosses(clearing *game.Clearing, faction game.Faction, losses int) int {
	if losses <= 0 || clearing.Warriors == nil {
		return losses
	}

	available := clearing.Warriors[faction]
	if available <= 0 {
		return losses
	}

	removed := losses
	if removed > available {
		removed = available
	}

	clearing.Warriors[faction] = available - removed
	return losses - removed
}

func decrementPlacedBuildingCounter(state *game.GameState, buildingType game.BuildingType) {
	switch buildingType {
	case game.Sawmill:
		if state.Marquise.SawmillsPlaced > 0 {
			state.Marquise.SawmillsPlaced--
		}
	case game.Workshop:
		if state.Marquise.WorkshopsPlaced > 0 {
			state.Marquise.WorkshopsPlaced--
		}
	case game.Recruiter:
		if state.Marquise.RecruitersPlaced > 0 {
			state.Marquise.RecruitersPlaced--
		}
	}
}

func removeBuildingLosses(state *game.GameState, clearing *game.Clearing, faction game.Faction, losses int) {
	if losses <= 0 || len(clearing.Buildings) == 0 {
		return
	}

	remaining := make([]game.Building, 0, len(clearing.Buildings))
	for _, building := range clearing.Buildings {
		if losses > 0 && building.Faction == faction {
			if faction == game.Marquise {
				decrementPlacedBuildingCounter(state, building.Type)
			}
			if faction == game.Eyrie && building.Type == game.Roost && state.Eyrie.RoostsPlaced > 0 {
				state.Eyrie.RoostsPlaced--
			}
			if faction == game.Alliance && building.Type == game.Base {
				setAllianceBasePlaced(state, clearing.Suit, false)
			}
			losses--
			continue
		}
		remaining = append(remaining, building)
	}
	clearing.Buildings = remaining
}

func removeTokenLosses(state *game.GameState, clearing *game.Clearing, faction game.Faction, losses int) (int, int, int) {
	if losses <= 0 {
		return losses, 0, 0
	}

	removedTokens := 0
	removedSympathy := 0
	if len(clearing.Tokens) > 0 {
		remaining := make([]game.Token, 0, len(clearing.Tokens))
		for _, token := range clearing.Tokens {
			if losses > 0 && token.Faction == faction {
				if token.Faction == game.Alliance && token.Type == game.TokenSympathy && state.Alliance.SympathyPlaced > 0 {
					state.Alliance.SympathyPlaced--
					removedSympathy++
				}
				if token.Faction == game.Marquise && token.Type == game.TokenKeep {
					state.Marquise.KeepClearingID = 0
				}
				losses--
				removedTokens++
				continue
			}
			remaining = append(remaining, token)
		}
		clearing.Tokens = remaining
	}

	if losses > 0 && faction == game.Marquise && clearing.Wood > 0 {
		removedWood := losses
		if removedWood > clearing.Wood {
			removedWood = clearing.Wood
		}
		clearing.Wood -= removedWood
		losses -= removedWood
		removedTokens += removedWood
	}

	return losses, removedTokens, removedSympathy
}

func applyBattleResolution(state *game.GameState, action game.Action) {
	if action.BattleResolution == nil {
		return
	}

	index := findClearingIndex(state.Map, action.BattleResolution.ClearingID)
	if index == -1 {
		return
	}

	clearing := &state.Map.Clearings[index]
	if action.BattleResolution.Faction == game.Vagabond {
		exhaustReadyItemsByType(state, game.ItemSword, 1)
		damageVagabondItems(state, action.BattleResolution.AttackerLosses)
	} else {
		removeWarriorLosses(clearing, action.BattleResolution.Faction, action.BattleResolution.AttackerLosses)
	}

	if action.BattleResolution.TargetFaction == game.Vagabond {
		exhaustReadyItemsByType(state, game.ItemSword, 1)
		damageVagabondItems(state, action.BattleResolution.DefenderLosses)
		return
	}

	targetWarriorsBefore := 0
	if clearing.Warriors != nil {
		targetWarriorsBefore = clearing.Warriors[action.BattleResolution.TargetFaction]
	}
	remainingDefenderLosses := removeWarriorLosses(clearing, action.BattleResolution.TargetFaction, action.BattleResolution.DefenderLosses)
	beforeBuildings := len(clearing.Buildings)
	removeBuildingLosses(state, clearing, action.BattleResolution.TargetFaction, remainingDefenderLosses)
	removedBuildings := beforeBuildings - len(clearing.Buildings)
	remainingDefenderLosses -= removedBuildings
	_, removedTokens, removedSympathy := removeTokenLosses(state, clearing, action.BattleResolution.TargetFaction, remainingDefenderLosses)
	removedWarriors := targetWarriorsBefore
	if clearing.Warriors != nil {
		removedWarriors -= clearing.Warriors[action.BattleResolution.TargetFaction]
	}
	scoreBattleRemovals(state, action.BattleResolution.Faction, removedBuildings, removedTokens)

	if removedSympathy > 0 && action.BattleResolution.Faction != game.Alliance {
		transferOutrageCard(state, action.BattleResolution.Faction, clearing.Suit)
	}
	if action.BattleResolution.Faction == game.Vagabond && removedWarriors+removedBuildings+removedTokens > 0 {
		setVagabondRelationship(state, action.BattleResolution.TargetFaction, game.RelHostile)
	}
}

func applyBuild(state *game.GameState, action game.Action) {
	if action.Build == nil {
		return
	}

	index := findClearingIndex(state.Map, action.Build.ClearingID)
	if index == -1 {
		return
	}

	state.Map.Clearings[index].Buildings = append(
		state.Map.Clearings[index].Buildings,
		game.Building{
			Faction: action.Build.Faction,
			Type:    action.Build.BuildingType,
		},
	)

	switch action.Build.BuildingType {
	case game.Sawmill:
		scoreMarquiseBuilding(state, action.Build.BuildingType, state.Marquise.SawmillsPlaced)
		state.Marquise.SawmillsPlaced++
	case game.Workshop:
		scoreMarquiseBuilding(state, action.Build.BuildingType, state.Marquise.WorkshopsPlaced)
		state.Marquise.WorkshopsPlaced++
	case game.Recruiter:
		scoreMarquiseBuilding(state, action.Build.BuildingType, state.Marquise.RecruitersPlaced)
		state.Marquise.RecruitersPlaced++
	case game.Roost:
		state.Eyrie.RoostsPlaced++
	case game.Base:
		setAllianceBasePlaced(state, state.Map.Clearings[index].Suit, true)
	}

	for _, source := range action.Build.WoodSources {
		sourceIndex := findClearingIndex(state.Map, source.ClearingID)
		if sourceIndex == -1 {
			continue
		}

		wood := state.Map.Clearings[sourceIndex].Wood
		if source.Amount >= wood {
			state.Map.Clearings[sourceIndex].Wood = 0
			continue
		}

		state.Map.Clearings[sourceIndex].Wood -= source.Amount
	}
}

func applyOverwork(state *game.GameState, action game.Action) {
	if action.Overwork == nil {
		return
	}

	index := findClearingIndex(state.Map, action.Overwork.ClearingID)
	if index == -1 {
		return
	}

	if _, ok := removeCardFromFactionHand(state, game.Marquise, action.Overwork.CardID); !ok {
		return
	}
	state.Map.Clearings[index].Wood++
	DiscardCard(state, action.Overwork.CardID)
}

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
	DiscardCard(state, action.Craft.CardID)
	state.TurnProgress.UsedWorkshopClearings = append(
		state.TurnProgress.UsedWorkshopClearings,
		action.Craft.UsedWorkshopClearings...,
	)
}

func applySpreadSympathy(state *game.GameState, action game.Action) {
	if action.SpreadSympathy == nil {
		return
	}

	index := findClearingIndex(state.Map, action.SpreadSympathy.ClearingID)
	if index == -1 {
		return
	}

	state.Alliance.Supporters = removeCardsByID(state.Alliance.Supporters, action.SpreadSympathy.SupporterCardIDs)
	DiscardCards(state, action.SpreadSympathy.SupporterCardIDs)
	state.Map.Clearings[index].Tokens = append(state.Map.Clearings[index].Tokens, game.Token{
		Faction: game.Alliance,
		Type:    game.TokenSympathy,
	})
	scoreAllianceSympathy(state, state.Alliance.SympathyPlaced)
	state.Alliance.SympathyPlaced++
}

func removeEnemyPiecesForRevolt(state *game.GameState, clearing *game.Clearing) int {
	removedPieces := 0

	for faction, warriors := range clearing.Warriors {
		if faction == game.Alliance || warriors <= 0 {
			continue
		}
		removedPieces += warriors
		clearing.Warriors[faction] = 0
	}

	if len(clearing.Buildings) > 0 {
		remaining := make([]game.Building, 0, len(clearing.Buildings))
		for _, building := range clearing.Buildings {
			if building.Faction == game.Alliance {
				remaining = append(remaining, building)
				continue
			}

			if building.Faction == game.Marquise {
				decrementPlacedBuildingCounter(state, building.Type)
			}
			if building.Faction == game.Eyrie && building.Type == game.Roost && state.Eyrie.RoostsPlaced > 0 {
				state.Eyrie.RoostsPlaced--
			}
			if building.Faction == game.Alliance && building.Type == game.Base {
				setAllianceBasePlaced(state, clearing.Suit, false)
			}
			removedPieces++
		}
		clearing.Buildings = remaining
	}

	if len(clearing.Tokens) > 0 {
		remaining := make([]game.Token, 0, len(clearing.Tokens))
		for _, token := range clearing.Tokens {
			if token.Faction == game.Alliance {
				remaining = append(remaining, token)
				continue
			}

			if token.Faction == game.Marquise && token.Type == game.TokenKeep {
				state.Marquise.KeepClearingID = 0
			}
			removedPieces++
		}
		clearing.Tokens = remaining
	}

	if clearing.Wood > 0 {
		removedPieces += clearing.Wood
		clearing.Wood = 0
	}

	return removedPieces
}

func sympathyCountBySuit(board game.Map, suit game.Suit) int {
	count := 0
	for _, clearing := range board.Clearings {
		if clearing.Suit != suit || !hasAllianceSympathy(clearing) {
			continue
		}
		count++
	}

	return count
}

func applyRevolt(state *game.GameState, action game.Action) {
	if action.Revolt == nil {
		return
	}

	index := findClearingIndex(state.Map, action.Revolt.ClearingID)
	if index == -1 {
		return
	}

	state.Alliance.Supporters = removeCardsByID(state.Alliance.Supporters, action.Revolt.SupporterCardIDs)
	DiscardCards(state, action.Revolt.SupporterCardIDs)
	clearing := &state.Map.Clearings[index]
	removedPieces := removeEnemyPiecesForRevolt(state, clearing)
	clearing.Buildings = append(clearing.Buildings, game.Building{
		Faction: game.Alliance,
		Type:    game.Base,
	})
	setAllianceBasePlaced(state, action.Revolt.BaseSuit, true)

	if clearing.Warriors == nil {
		clearing.Warriors = map[game.Faction]int{}
	}

	recruitCount := sympathyCountBySuit(state.Map, action.Revolt.BaseSuit)
	if recruitCount > state.Alliance.WarriorSupply {
		recruitCount = state.Alliance.WarriorSupply
	}
	clearing.Warriors[game.Alliance] += recruitCount
	state.Alliance.WarriorSupply -= recruitCount
	state.Alliance.Officers++
	addVictoryPoints(state, game.Alliance, removedPieces)
}

func applyMobilize(state *game.GameState, action game.Action) {
	if action.Mobilize == nil {
		return
	}

	for _, card := range state.Alliance.CardsInHand {
		if card.ID != action.Mobilize.CardID {
			continue
		}

		if _, ok := removeCardFromFactionHand(state, game.Alliance, card.ID); !ok {
			return
		}
		addAllianceSupporter(state, card)
		return
	}
}

func applyTrain(state *game.GameState, action game.Action) {
	if action.Train == nil {
		return
	}

	if _, ok := removeCardFromFactionHand(state, game.Alliance, action.Train.CardID); !ok {
		return
	}
	DiscardCard(state, action.Train.CardID)
	state.Alliance.Officers++
}

func applyOrganize(state *game.GameState, action game.Action) {
	if action.Organize == nil {
		return
	}

	index := findClearingIndex(state.Map, action.Organize.ClearingID)
	if index == -1 {
		return
	}

	clearing := &state.Map.Clearings[index]
	if clearing.Warriors == nil || clearing.Warriors[game.Alliance] <= 0 {
		return
	}

	clearing.Warriors[game.Alliance]--
	clearing.Tokens = append(clearing.Tokens, game.Token{
		Faction: game.Alliance,
		Type:    game.TokenSympathy,
	})
	scoreAllianceSympathy(state, state.Alliance.SympathyPlaced)
	state.Alliance.SympathyPlaced++
}

func appendCardToDecree(decree *game.Decree, column game.DecreeColumn, cardID game.CardID) {
	switch column {
	case game.DecreeRecruit:
		decree.Recruit = append(decree.Recruit, cardID)
	case game.DecreeMove:
		decree.Move = append(decree.Move, cardID)
	case game.DecreeBattle:
		decree.Battle = append(decree.Battle, cardID)
	case game.DecreeBuild:
		decree.Build = append(decree.Build, cardID)
	}
}

func applyAddToDecree(state *game.GameState, action game.Action) {
	if action.AddToDecree == nil {
		return
	}

	state.TurnProgress.ResolvedDecreeCardIDs = nil
	state.TurnProgress.DecreeColumnsResolved = 0
	state.TurnProgress.DecreeCardsResolved = 0

	for i, cardID := range action.AddToDecree.CardIDs {
		if i >= len(action.AddToDecree.Columns) {
			break
		}

		appendCardToDecree(&state.Eyrie.Decree, action.AddToDecree.Columns[i], cardID)
		state.Eyrie.CardsInHand = removeCardByID(state.Eyrie.CardsInHand, cardID)
	}
}

func removeLeader(leaders []game.EyrieLeader, remove game.EyrieLeader) []game.EyrieLeader {
	filtered := make([]game.EyrieLeader, 0, len(leaders))
	for _, leader := range leaders {
		if leader != remove {
			filtered = append(filtered, leader)
		}
	}
	return filtered
}

func eyrieCardSuit(id game.CardID) game.Suit {
	if id == game.LoyalVizier1 || id == game.LoyalVizier2 {
		return game.Bird
	}

	for _, card := range carddata.BaseDeck() {
		if card.ID == id {
			return card.Suit
		}
	}

	return game.Bird
}

func birdCardsInDecree(decree game.Decree) int {
	count := 0
	for _, column := range [][]game.CardID{decree.Recruit, decree.Move, decree.Battle, decree.Build} {
		for _, cardID := range column {
			if eyrieCardSuit(cardID) == game.Bird {
				count++
			}
		}
	}
	return count
}

func vizierColumnsForLeader(leader game.EyrieLeader) [2]game.DecreeColumn {
	switch leader {
	case game.LeaderBuilder:
		return [2]game.DecreeColumn{game.DecreeRecruit, game.DecreeMove}
	case game.LeaderCharismatic:
		return [2]game.DecreeColumn{game.DecreeRecruit, game.DecreeBattle}
	case game.LeaderCommander:
		return [2]game.DecreeColumn{game.DecreeMove, game.DecreeBattle}
	default:
		return [2]game.DecreeColumn{game.DecreeMove, game.DecreeBuild}
	}
}

func applyTurmoil(state *game.GameState, action game.Action) {
	if action.Turmoil == nil {
		return
	}

	if state.VictoryPoints == nil {
		state.VictoryPoints = map[game.Faction]int{}
	}
	state.VictoryPoints[game.Eyrie] -= birdCardsInDecree(state.Eyrie.Decree)

	DiscardCards(state, state.Eyrie.Decree.Recruit)
	DiscardCards(state, state.Eyrie.Decree.Move)
	DiscardCards(state, state.Eyrie.Decree.Battle)
	DiscardCards(state, state.Eyrie.Decree.Build)
	state.Eyrie.AvailableLeaders = removeLeader(state.Eyrie.AvailableLeaders, state.Eyrie.Leader)
	state.Eyrie.AvailableLeaders = removeLeader(state.Eyrie.AvailableLeaders, action.Turmoil.NewLeader)
	state.Eyrie.Leader = action.Turmoil.NewLeader
	state.Eyrie.Decree = game.Decree{}
	state.TurnProgress.ResolvedDecreeCardIDs = nil

	vizierColumns := vizierColumnsForLeader(action.Turmoil.NewLeader)
	appendCardToDecree(&state.Eyrie.Decree, vizierColumns[0], game.LoyalVizier1)
	appendCardToDecree(&state.Eyrie.Decree, vizierColumns[1], game.LoyalVizier2)
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

func advanceTurnState(state *game.GameState, action game.Action) {
	switch action.Type {
	case game.ActionAddToDecree:
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.StepDaylightActions
	case game.ActionBirdsongWood:
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.StepDaylightActions
	case game.ActionDaybreak:
		state.CurrentPhase = game.Birdsong
		state.CurrentStep = game.StepBirdsong
	case game.ActionSlip:
		state.CurrentPhase = game.Birdsong
		state.CurrentStep = game.StepBirdsong
		state.TurnProgress.HasSlipped = true
	case game.ActionRecruit:
		if action.Recruit != nil && action.Recruit.Faction == game.Alliance {
			state.CurrentPhase = game.Evening
			state.CurrentStep = game.StepEvening
			state.TurnProgress.OfficerActionsUsed++
		} else {
			state.CurrentStep = game.StepDaylightActions
		}
		if action.Recruit != nil && action.Recruit.Faction == game.Marquise {
			state.TurnProgress.ActionsUsed++
		}
		if action.Recruit != nil && action.Recruit.Faction == game.Eyrie {
			markResolvedDecreeCard(state, action.Recruit.DecreeCardID)
		}
	case game.ActionMovement:
		if action.Movement != nil && action.Movement.Faction == game.Alliance {
			state.CurrentPhase = game.Evening
			state.CurrentStep = game.StepEvening
			state.TurnProgress.OfficerActionsUsed++
		} else {
			state.CurrentStep = game.StepDaylightActions
		}
		if action.Movement != nil && action.Movement.Faction == game.Marquise {
			state.TurnProgress.ActionsUsed++
			state.TurnProgress.MarchesUsed++
		}
		if action.Movement != nil && action.Movement.Faction == game.Eyrie {
			markResolvedDecreeCard(state, action.Movement.DecreeCardID)
		}
	case game.ActionBattleResolution, game.ActionBuild, game.ActionOverwork:
		state.CurrentStep = game.StepDaylightActions
		if action.Type == game.ActionBattleResolution && action.BattleResolution != nil && action.BattleResolution.Faction == game.Alliance {
			state.CurrentPhase = game.Evening
			state.CurrentStep = game.StepEvening
			state.TurnProgress.OfficerActionsUsed++
		}
		switch {
		case action.Type == game.ActionBattleResolution && action.BattleResolution != nil && action.BattleResolution.Faction == game.Marquise:
			state.TurnProgress.ActionsUsed++
		case action.Type == game.ActionBattleResolution && action.BattleResolution != nil && action.BattleResolution.Faction == game.Eyrie:
			markResolvedDecreeCard(state, action.BattleResolution.DecreeCardID)
		case action.Type == game.ActionBuild && action.Build != nil && action.Build.Faction == game.Marquise:
			state.TurnProgress.ActionsUsed++
		case action.Type == game.ActionBuild && action.Build != nil && action.Build.Faction == game.Eyrie:
			markResolvedDecreeCard(state, action.Build.DecreeCardID)
		case action.Type == game.ActionOverwork:
			state.TurnProgress.ActionsUsed++
		}
	case game.ActionCraft:
		state.CurrentStep = game.StepDaylightActions
		state.TurnProgress.HasCrafted = true
	case game.ActionExplore, game.ActionAid, game.ActionQuest, game.ActionStrike, game.ActionRepair:
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.StepDaylightActions
	case game.ActionSpreadSympathy, game.ActionRevolt:
		state.CurrentPhase = game.Birdsong
		state.CurrentStep = game.StepBirdsong
	case game.ActionMobilize, game.ActionTrain:
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.StepDaylightActions
	case game.ActionOrganize:
		state.CurrentPhase = game.Evening
		state.CurrentStep = game.StepEvening
		state.TurnProgress.OfficerActionsUsed++
	case game.ActionTurmoil:
		state.CurrentPhase = game.Evening
		state.CurrentStep = game.StepEvening
		state.TurnProgress.DecreeColumnsResolved = 0
		state.TurnProgress.DecreeCardsResolved = 0
	case game.ActionScoreRoosts:
		beginNextFactionTurn(state)
	case game.ActionPassPhase:
		switch state.CurrentPhase {
		case game.Birdsong:
			state.CurrentPhase = game.Daylight
			state.CurrentStep = game.StepDaylightActions
		case game.Daylight:
			if state.CurrentStep == game.StepDaylightCraft {
				state.CurrentStep = game.StepDaylightActions
			} else {
				state.CurrentPhase = game.Evening
				state.CurrentStep = game.StepEvening
			}
		case game.Evening:
			beginNextFactionTurn(state)
		}
	case game.ActionEveningDraw:
		beginNextFactionTurn(state)
	}
}

func decreeCardsByColumn(decree game.Decree, column game.DecreeColumn) []game.CardID {
	switch column {
	case game.DecreeRecruit:
		return decree.Recruit
	case game.DecreeMove:
		return decree.Move
	case game.DecreeBattle:
		return decree.Battle
	case game.DecreeBuild:
		return decree.Build
	default:
		return nil
	}
}

func currentDecreeCard(state game.GameState) (game.DecreeColumn, game.CardID, bool) {
	for i := state.TurnProgress.DecreeColumnsResolved; i < 4; i++ {
		column := game.DecreeColumn(i)
		for _, cardID := range decreeCardsByColumn(state.Eyrie.Decree, column) {
			if !decreeCardResolved(state, cardID) {
				return column, cardID, true
			}
		}
	}

	return 0, 0, false
}

func decreeCardResolved(state game.GameState, cardID game.CardID) bool {
	for _, resolvedID := range state.TurnProgress.ResolvedDecreeCardIDs {
		if resolvedID == cardID {
			return true
		}
	}
	return false
}

func markResolvedDecreeCard(state *game.GameState, cardID game.CardID) {
	if cardID == 0 || decreeCardResolved(*state, cardID) {
		return
	}

	state.TurnProgress.ResolvedDecreeCardIDs = append(state.TurnProgress.ResolvedDecreeCardIDs, cardID)
	state.TurnProgress.DecreeCardsResolved = 0

	for state.TurnProgress.DecreeColumnsResolved < 4 {
		column := game.DecreeColumn(state.TurnProgress.DecreeColumnsResolved)
		allResolved := true
		for _, columnCardID := range decreeCardsByColumn(state.Eyrie.Decree, column) {
			if !decreeCardResolved(*state, columnCardID) {
				allResolved = false
				break
			}
		}
		if !allResolved {
			return
		}
		state.TurnProgress.DecreeColumnsResolved++
	}
}
