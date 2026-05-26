package engine

import (
	"sort"

	"github.com/imdehydrated/rootbuddy/game"
)

func vagabondRelationshipLevel(state game.GameState, faction game.Faction) game.RelationshipLevel {
	if state.Vagabond.Relationships == nil {
		return game.RelIndifferent
	}

	relationship, ok := state.Vagabond.Relationships[faction]
	if !ok {
		return game.RelIndifferent
	}

	return relationship
}

func setVagabondRelationship(state *game.GameState, faction game.Faction, relationship game.RelationshipLevel) {
	if state.Vagabond.Relationships == nil {
		state.Vagabond.Relationships = map[game.Faction]game.RelationshipLevel{}
	}

	state.Vagabond.Relationships[faction] = relationship
}

func vagabondAidProgress(state game.GameState, faction game.Faction) int {
	if state.TurnProgress.VagabondAidCounts == nil {
		return 0
	}

	return state.TurnProgress.VagabondAidCounts[faction]
}

func setVagabondAidProgress(state *game.GameState, faction game.Faction, count int) {
	if count <= 0 {
		if state.TurnProgress.VagabondAidCounts != nil {
			delete(state.TurnProgress.VagabondAidCounts, faction)
		}
		return
	}
	if state.TurnProgress.VagabondAidCounts == nil {
		state.TurnProgress.VagabondAidCounts = map[game.Faction]int{}
	}

	state.TurnProgress.VagabondAidCounts[faction] = count
}

func vagabondAidCostToImprove(relationship game.RelationshipLevel) int {
	switch relationship {
	case game.RelIndifferent:
		return 1
	case game.RelAmiable:
		return 2
	case game.RelFriendly:
		return 3
	default:
		return 0
	}
}

func vagabondRelationshipReward(relationship game.RelationshipLevel) int {
	switch relationship {
	case game.RelAmiable:
		return 1
	case game.RelFriendly, game.RelAllied:
		return 2
	default:
		return 0
	}
}

func resolveVagabondAidRelationship(state *game.GameState, faction game.Faction) {
	relationship := vagabondRelationshipLevel(*state, faction)
	switch relationship {
	case game.RelHostile:
		return
	case game.RelAllied:
		addVictoryPoints(state, game.Vagabond, 2)
		return
	}

	cost := vagabondAidCostToImprove(relationship)
	if cost == 0 {
		return
	}

	progress := vagabondAidProgress(*state, faction) + 1
	if progress < cost {
		setVagabondAidProgress(state, faction, progress)
		return
	}

	nextRelationship := relationship + 1
	setVagabondRelationship(state, faction, nextRelationship)
	setVagabondAidProgress(state, faction, 0)
	addVictoryPoints(state, game.Vagabond, vagabondRelationshipReward(nextRelationship))
}

func addCraftedItem(state *game.GameState, faction game.Faction, itemType game.ItemType) {
	if state.CraftedItems == nil {
		state.CraftedItems = map[game.Faction][]game.ItemType{}
	}

	state.CraftedItems[faction] = append(state.CraftedItems[faction], itemType)
}

func takeCraftedItem(state *game.GameState, faction game.Faction, index int) (game.ItemType, bool) {
	if index < 0 || state.CraftedItems == nil || index >= len(state.CraftedItems[faction]) {
		return 0, false
	}

	itemType := state.CraftedItems[faction][index]
	state.CraftedItems[faction] = append(state.CraftedItems[faction][:index], state.CraftedItems[faction][index+1:]...)
	return itemType, true
}

func gainVagabondItem(state *game.GameState, itemType game.ItemType) {
	state.Vagabond.Items = append(state.Vagabond.Items, game.NormalizeItemZone(game.Item{
		Type:   itemType,
		Status: game.ItemReady,
	}))
}

func vagabondItemIndex(state game.GameState, itemType game.ItemType, status game.ItemStatus) (int, bool) {
	for index, item := range state.Vagabond.Items {
		if item.Type == itemType && item.Status == status {
			return index, true
		}
	}

	return -1, false
}

func setVagabondItemStatus(state *game.GameState, index int, status game.ItemStatus) {
	state.Vagabond.Items[index].Status = status
	state.Vagabond.Items[index].Zone = game.ItemZoneForStatus(state.Vagabond.Items[index].Type, status)
}

func exhaustReadyItemsByType(state *game.GameState, itemType game.ItemType, count int) int {
	exhausted := 0
	for exhausted < count {
		index, ok := vagabondItemIndex(*state, itemType, game.ItemReady)
		if !ok {
			return exhausted
		}

		setVagabondItemStatus(state, index, game.ItemExhausted)
		exhausted++
	}

	return exhausted
}

func exhaustAnyReadyItems(state *game.GameState, count int) int {
	exhausted := 0
	for exhausted < count {
		found := false
		for index, item := range state.Vagabond.Items {
			if item.Status != game.ItemReady {
				continue
			}

			setVagabondItemStatus(state, index, game.ItemExhausted)
			exhausted++
			found = true
			break
		}
		if !found {
			return exhausted
		}
	}

	return exhausted
}

func exhaustReadyItemsByIndexes(state *game.GameState, indexes []int) bool {
	seen := map[int]bool{}
	for _, index := range indexes {
		if index < 0 || index >= len(state.Vagabond.Items) || seen[index] {
			return false
		}
		if state.Vagabond.Items[index].Status != game.ItemReady {
			return false
		}
		seen[index] = true
	}

	for _, index := range indexes {
		setVagabondItemStatus(state, index, game.ItemExhausted)
	}

	return true
}

func repairDamagedItem(state *game.GameState, index int) bool {
	if index < 0 || index >= len(state.Vagabond.Items) {
		return false
	}
	if state.Vagabond.Items[index].Status != game.ItemDamaged {
		return false
	}

	setVagabondItemStatus(state, index, game.ItemReady)
	return true
}

func repairAllDamagedItems(state *game.GameState) {
	for index, item := range state.Vagabond.Items {
		if item.Status == game.ItemDamaged {
			setVagabondItemStatus(state, index, game.ItemReady)
		}
	}
}

func damageVagabondItems(state *game.GameState, count int) int {
	damaged := 0
	for damaged < count {
		found := false
		for index, item := range state.Vagabond.Items {
			if item.Status == game.ItemDamaged {
				continue
			}

			setVagabondItemStatus(state, index, game.ItemDamaged)
			damaged++
			found = true
			break
		}
		if !found {
			return damaged
		}
	}

	return damaged
}

func vagabondReadySwordCount(state game.GameState) int {
	count := 0
	for _, item := range state.Vagabond.Items {
		if item.Type == game.ItemSword && item.Status == game.ItemReady {
			count++
		}
	}

	return count
}

func vagabondExhaustedSwordCount(state game.GameState) int {
	count := 0
	for _, item := range state.Vagabond.Items {
		if item.Type == game.ItemSword && item.Status == game.ItemExhausted {
			count++
		}
	}

	return count
}

func vagabondBattleHitCap(state game.GameState) int {
	return vagabondReadySwordCount(state) + vagabondExhaustedSwordCount(state)
}

func appendCardToFactionHand(state *game.GameState, faction game.Faction, card game.Card) {
	if !tracksHandForFaction(*state, faction) {
		incrementOtherHandCount(state, faction, 1)
		return
	}

	switch faction {
	case game.Marquise:
		state.Marquise.CardsInHand = append(state.Marquise.CardsInHand, card)
	case game.Eyrie:
		state.Eyrie.CardsInHand = append(state.Eyrie.CardsInHand, card)
	case game.Alliance:
		state.Alliance.CardsInHand = append(state.Alliance.CardsInHand, card)
	case game.Vagabond:
		state.Vagabond.CardsInHand = append(state.Vagabond.CardsInHand, card)
	}
}

func removeCardFromCards(cards []game.Card, id game.CardID) ([]game.Card, game.Card, bool) {
	for i, card := range cards {
		if card.ID != id {
			continue
		}

		return append(cards[:i], cards[i+1:]...), card, true
	}

	return cards, game.Card{}, false
}

func removeQuestByID(quests []game.Quest, id game.QuestID) ([]game.Quest, game.Quest, bool) {
	for i, quest := range quests {
		if quest.ID != id {
			continue
		}

		return append(quests[:i], quests[i+1:]...), quest, true
	}

	return quests, game.Quest{}, false
}

func removeRuinItem(clearing *game.Clearing, itemType game.ItemType) bool {
	for index, ruinItem := range clearing.RuinItems {
		if ruinItem != itemType {
			continue
		}

		clearing.RuinItems = append(clearing.RuinItems[:index], clearing.RuinItems[index+1:]...)
		if len(clearing.RuinItems) == 0 {
			clearing.Ruins = false
		}
		return true
	}

	return false
}

func removeOneFactionPieceForStrike(state *game.GameState, clearing *game.Clearing, faction game.Faction) (bool, int, int, int) {
	if clearing.Warriors != nil && clearing.Warriors[faction] > 0 {
		clearing.Warriors[faction]--
		returnRemovedWarriorsToSupply(state, clearing, faction, 1)
		return true, 1, 0, 0
	}

	for index, building := range clearing.Buildings {
		if building.Faction != faction {
			continue
		}

		if faction == game.Marquise {
			decrementPlacedBuildingCounter(state, building.Type)
		}
		if faction == game.Eyrie && building.Type == game.Roost && state.Eyrie.RoostsPlaced > 0 {
			state.Eyrie.RoostsPlaced--
		}
		if faction == game.Alliance && building.Type == game.Base {
			removeAllianceBase(state, clearing.Suit)
		}

		clearing.Buildings = append(clearing.Buildings[:index], clearing.Buildings[index+1:]...)
		return true, 0, 1, 0
	}

	for index, token := range clearing.Tokens {
		if token.Faction != faction {
			continue
		}

		if token.Faction == game.Alliance && token.Type == game.TokenSympathy && state.Alliance.SympathyPlaced > 0 {
			state.Alliance.SympathyPlaced--
		}
		if token.Faction == game.Marquise && token.Type == game.TokenKeep {
			state.Marquise.KeepClearingID = 0
		}

		clearing.Tokens = append(clearing.Tokens[:index], clearing.Tokens[index+1:]...)
		return true, 0, 0, 1
	}

	if faction == game.Marquise && clearing.Wood > 0 {
		clearing.Wood--
		returnMarquiseWoodToSupply(state, 1)
		return true, 0, 0, 1
	}

	return false, 0, 0, 0
}

func questCountBySuit(quests []game.Quest, suit game.Suit) int {
	count := 0
	for _, quest := range quests {
		if quest.Suit == suit {
			count++
		}
	}

	return count
}

func questItemIndexesValid(state game.GameState, requiredItems []game.ItemType, indexes []int) bool {
	if len(requiredItems) != len(indexes) {
		return false
	}

	remaining := map[game.ItemType]int{}
	for _, itemType := range requiredItems {
		remaining[itemType]++
	}

	seen := map[int]bool{}
	for _, index := range indexes {
		if index < 0 || index >= len(state.Vagabond.Items) || seen[index] {
			return false
		}

		item := state.Vagabond.Items[index]
		if item.Status != game.ItemReady || remaining[item.Type] == 0 {
			return false
		}

		remaining[item.Type]--
		seen[index] = true
	}

	for _, count := range remaining {
		if count != 0 {
			return false
		}
	}

	return true
}

func applyDaybreak(state *game.GameState, action game.Action) {
	if action.Daybreak == nil {
		return
	}

	refreshLimit := 3 + 2*vagabondTrackItemCount(*state, game.ItemTea)
	refreshed := 0
	seen := map[int]bool{}
	for _, index := range action.Daybreak.RefreshedItemIndexes {
		if refreshed >= refreshLimit {
			return
		}
		if index < 0 || index >= len(state.Vagabond.Items) {
			continue
		}
		if seen[index] || state.Vagabond.Items[index].Status != game.ItemExhausted {
			continue
		}

		setVagabondItemStatus(state, index, game.ItemReady)
		seen[index] = true
		refreshed++
	}
}

func applySlip(state *game.GameState, action game.Action) {
	if action.Slip == nil {
		return
	}

	if action.Slip.ToForestID != 0 {
		state.Vagabond.ClearingID = 0
		state.Vagabond.ForestID = action.Slip.ToForestID
		state.Vagabond.InForest = true
		return
	}

	state.Vagabond.ClearingID = action.Slip.To
	state.Vagabond.ForestID = 0
	state.Vagabond.InForest = false
}

func applyExplore(state *game.GameState, action game.Action) {
	if action.Explore == nil {
		return
	}

	index := findClearingIndex(state.Map, action.Explore.ClearingID)
	if index == -1 {
		return
	}

	if exhaustReadyItemsByType(state, game.ItemTorch, 1) == 0 {
		return
	}
	if !removeRuinItem(&state.Map.Clearings[index], action.Explore.ItemType) {
		return
	}

	state.Vagabond.Items = append(state.Vagabond.Items, game.Item{
		Type:   action.Explore.ItemType,
		Status: game.ItemReady,
	})
	state.Vagabond.Items[len(state.Vagabond.Items)-1] = game.NormalizeItemZone(state.Vagabond.Items[len(state.Vagabond.Items)-1])
	addVictoryPoints(state, game.Vagabond, 1)
}

func applyAid(state *game.GameState, action game.Action) {
	if action.Aid == nil || action.Aid.Faction != game.Vagabond {
		return
	}

	clearing := findClearing(state, action.Aid.ClearingID)
	if clearing == nil {
		return
	}
	if !clearingHasFactionPiecesForAid(*clearing, action.Aid.TargetFaction) {
		return
	}

	aidCard, found := CardByID(action.Aid.CardID)
	if !found || (aidCard.Suit != clearing.Suit && aidCard.Suit != game.Bird) {
		return
	}

	if action.Aid.ItemIndex < 0 ||
		action.Aid.ItemIndex >= len(state.Vagabond.Items) ||
		state.Vagabond.Items[action.Aid.ItemIndex].Status != game.ItemReady {
		return
	}
	if action.Aid.TakeItemIndex != nil {
		takeIndex := *action.Aid.TakeItemIndex
		if takeIndex < 0 || state.CraftedItems == nil || takeIndex >= len(state.CraftedItems[action.Aid.TargetFaction]) {
			return
		}
	}

	card, ok := game.Card{}, false
	if tracksHandForFaction(*state, game.Vagabond) {
		var cards []game.Card
		cards, card, ok = removeCardFromCards(state.Vagabond.CardsInHand, action.Aid.CardID)
		if ok {
			state.Vagabond.CardsInHand = cards
		}
	} else {
		card, ok = spendFactionHandCard(state, game.Vagabond, action.Aid.CardID)
	}
	if !ok {
		return
	}
	setVagabondItemStatus(state, action.Aid.ItemIndex, game.ItemExhausted)

	appendCardToFactionHand(state, action.Aid.TargetFaction, card)
	if action.Aid.TakeItemIndex != nil {
		if itemType, ok := takeCraftedItem(state, action.Aid.TargetFaction, *action.Aid.TakeItemIndex); ok {
			gainVagabondItem(state, itemType)
		}
	}
	resolveVagabondAidRelationship(state, action.Aid.TargetFaction)
}

func clearingHasFactionPiecesForAid(clearing game.Clearing, faction game.Faction) bool {
	if clearing.Warriors != nil && clearing.Warriors[faction] > 0 {
		return true
	}
	for _, building := range clearing.Buildings {
		if building.Faction == faction {
			return true
		}
	}
	for _, token := range clearing.Tokens {
		if token.Faction == faction {
			return true
		}
	}
	if faction == game.Marquise && clearing.Wood > 0 {
		return true
	}

	return false
}

func applyQuest(state *game.GameState, action game.Action) {
	if action.Quest == nil {
		return
	}

	remaining, quest, ok := removeQuestByID(state.Vagabond.QuestsAvailable, action.Quest.QuestID)
	if !ok {
		return
	}
	if !questItemIndexesValid(*state, quest.RequiredItems, action.Quest.ItemIndexes) {
		return
	}
	if !exhaustReadyItemsByIndexes(state, action.Quest.ItemIndexes) {
		return
	}

	state.Vagabond.QuestsAvailable = remaining
	state.Vagabond.QuestsCompleted = append(state.Vagabond.QuestsCompleted, quest)
	drawAvailableQuest(state)
	if action.Quest.Reward == game.QuestRewardVictoryPoints {
		addVictoryPoints(state, game.Vagabond, questCountBySuit(state.Vagabond.QuestsCompleted, quest.Suit))
		return
	}

	DrawCards(state, game.Vagabond, 2)
}

func applyStrike(state *game.GameState, action game.Action) {
	if action.Strike == nil {
		return
	}
	if !game.AreEnemies(*state, game.Vagabond, action.Strike.TargetFaction) {
		return
	}

	index := findClearingIndex(state.Map, action.Strike.ClearingID)
	if index == -1 {
		return
	}
	if exhaustReadyItemsByType(state, game.ItemCrossbow, 1) == 0 {
		return
	}

	removed, removedWarriors, removedBuildings, removedTokens := removeOneFactionPieceForStrike(state, &state.Map.Clearings[index], action.Strike.TargetFaction)
	if !removed {
		return
	}

	scoreBattleRemovals(state, game.Vagabond, removedBuildings, removedTokens)
	if removedWarriors > 0 {
		setVagabondRelationship(state, action.Strike.TargetFaction, game.RelHostile)
	}
}

func applyRepair(state *game.GameState, action game.Action) {
	if action.Repair == nil {
		return
	}
	if exhaustReadyItemsByType(state, game.ItemHammer, 1) == 0 {
		return
	}

	repairDamagedItem(state, action.Repair.ItemIndex)
}

func applyVagabondSteal(state *game.GameState, action game.Action) {
	if action.VagabondSteal == nil || action.VagabondSteal.Faction != game.Vagabond {
		return
	}
	if state.Vagabond.Character != game.CharThief {
		return
	}
	clearing := findClearing(state, action.VagabondSteal.ClearingID)
	if clearing == nil ||
		state.Vagabond.InForest ||
		state.Vagabond.ClearingID != action.VagabondSteal.ClearingID ||
		!clearingHasFactionPiecesForAid(*clearing, action.VagabondSteal.TargetFaction) ||
		factionHandSize(*state, action.VagabondSteal.TargetFaction) == 0 {
		return
	}
	if _, ok := vagabondItemIndex(*state, game.ItemTorch, game.ItemReady); !ok {
		return
	}

	card, knownCard, ok := stealVagabondCard(state, action.VagabondSteal.TargetFaction, action.VagabondSteal.ObservedCardID)
	if !ok {
		return
	}
	if exhaustReadyItemsByType(state, game.ItemTorch, 1) == 0 {
		return
	}
	if knownCard {
		appendCardToFactionHand(state, game.Vagabond, card)
	} else if !tracksHandForFaction(*state, game.Vagabond) {
		incrementOtherHandCount(state, game.Vagabond, 1)
	}
}

func applyVagabondDayLabor(state *game.GameState, action game.Action) {
	if action.VagabondDayLabor == nil || action.VagabondDayLabor.Faction != game.Vagabond {
		return
	}
	if state.Vagabond.Character != game.CharTinker {
		return
	}
	clearing := findClearing(state, action.VagabondDayLabor.ClearingID)
	if clearing == nil ||
		state.Vagabond.InForest ||
		state.Vagabond.ClearingID != action.VagabondDayLabor.ClearingID {
		return
	}

	card, ok := CardByID(action.VagabondDayLabor.CardID)
	if !ok || !cardMatchesSuitOrBird(card, clearing.Suit) {
		return
	}
	if exhaustReadyItemsByType(state, game.ItemTorch, 1) == 0 {
		return
	}
	if !removeCardIDFromDiscard(state, action.VagabondDayLabor.CardID) {
		return
	}

	appendCardToFactionHand(state, game.Vagabond, card)
}

func applyVagabondHideout(state *game.GameState, action game.Action) {
	if action.VagabondHideout == nil || action.VagabondHideout.Faction != game.Vagabond {
		return
	}
	if state.Vagabond.Character != game.CharRanger {
		return
	}
	if !hideoutRepairIndexesValid(*state, action.VagabondHideout.ItemIndexes) {
		return
	}
	if exhaustReadyItemsByType(state, game.ItemTorch, 1) == 0 {
		return
	}

	for _, itemIndex := range action.VagabondHideout.ItemIndexes {
		repairDamagedItem(state, itemIndex)
	}
}

func applyVagabondRest(state *game.GameState, action game.Action) {
	if action.VagabondRest == nil || action.VagabondRest.Faction != game.Vagabond {
		return
	}

	if state.Vagabond.InForest {
		repairAllDamagedItems(state)
	}
}

func applyVagabondDiscard(state *game.GameState, action game.Action) {
	if action.VagabondDiscard == nil || action.VagabondDiscard.Faction != game.Vagabond {
		return
	}

	for _, cardID := range action.VagabondDiscard.CardIDs {
		if _, ok := spendFactionHandCard(state, game.Vagabond, cardID); ok {
			DiscardCard(state, cardID)
		}
	}
}

func applyVagabondItemCapacity(state *game.GameState, action game.Action) {
	if action.VagabondCapacity == nil || action.VagabondCapacity.Faction != game.Vagabond {
		return
	}

	indexes := append([]int(nil), action.VagabondCapacity.ItemIndexes...)
	sort.Sort(sort.Reverse(sort.IntSlice(indexes)))
	for _, index := range indexes {
		if index < 0 || index >= len(state.Vagabond.Items) {
			continue
		}
		state.Vagabond.Items = append(state.Vagabond.Items[:index], state.Vagabond.Items[index+1:]...)
	}
}

func vagabondItemIndexesByStatus(state game.GameState, itemType game.ItemType, status game.ItemStatus) []int {
	indexes := []int{}
	for index, item := range state.Vagabond.Items {
		if item.Type == itemType && item.Status == status {
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func vagabondTrackItemCount(state game.GameState, itemType game.ItemType) int {
	count := 0
	for _, item := range state.Vagabond.Items {
		if item.Type == itemType && game.ItemCurrentZone(item) == game.ItemZoneTrack {
			count++
		}
	}
	return count
}

func stealVagabondCard(state *game.GameState, targetFaction game.Faction, observedCardID game.CardID) (game.Card, bool, bool) {
	if state.GameMode == game.GameModeAssist && !tracksHandForFaction(*state, targetFaction) {
		if observedCardID > 0 {
			card, ok := CardByID(observedCardID)
			if !ok {
				return game.Card{}, false, false
			}
			decrementOtherHandCount(state, targetFaction, 1)
			return card, true, true
		}
		decrementOtherHandCount(state, targetFaction, 1)
		return game.Card{}, false, true
	}

	card, ok := removeRandomCardFromFactionHand(state, targetFaction)
	return card, ok, ok
}

func removeCardIDFromDiscard(state *game.GameState, cardID game.CardID) bool {
	for index, discardID := range state.DiscardPile {
		if discardID != cardID {
			continue
		}
		state.DiscardPile = append(state.DiscardPile[:index], state.DiscardPile[index+1:]...)
		return true
	}

	return false
}

func hideoutRepairIndexesValid(state game.GameState, itemIndexes []int) bool {
	damagedCount := 0
	for _, item := range state.Vagabond.Items {
		if item.Status == game.ItemDamaged {
			damagedCount++
		}
	}
	if damagedCount < 3 || len(itemIndexes) != 3 {
		return false
	}

	seen := map[int]bool{}
	for _, itemIndex := range itemIndexes {
		if itemIndex < 0 || itemIndex >= len(state.Vagabond.Items) || seen[itemIndex] {
			return false
		}
		if state.Vagabond.Items[itemIndex].Status != game.ItemDamaged {
			return false
		}
		seen[itemIndex] = true
	}

	return true
}

func vagabondReadyItemCount(state game.GameState) int {
	count := 0
	for _, item := range state.Vagabond.Items {
		if item.Status == game.ItemReady {
			count++
		}
	}
	return count
}
