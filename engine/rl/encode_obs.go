package rl

import (
	rootengine "github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

type observationInput struct {
	perspective game.Faction
	omniscient  bool
	state       rootengine.TrainingPublicState
	hidden      rootengine.TrainingHiddenCounts
}

func EncodeObservation(obs rootengine.TrainingObservation) []float32 {
	return encodeObservationState(observationInput{
		perspective: obs.Perspective,
		omniscient:  obs.Omniscient,
		state:       obs.State,
		hidden:      obs.Hidden,
	})
}

func encodeObservationState(input observationInput) []float32 {
	out := make([]float32, 0, 1024)
	state := input.state

	out = appendFactionOneHot(out, input.perspective)
	out = appendBool(out, input.omniscient)
	out = appendOneHot(out, int(state.GameMode), gameModeCount)
	out = appendOneHot(out, int(state.GamePhase), lifecycleCount)
	out = appendOneHot(out, int(state.SetupStage), setupStageCount)
	out = appendFactionOneHot(out, state.PlayerFaction)
	out = appendOptionalFactionOneHot(out, state.Winner, state.GamePhase == game.LifecycleGameOver)
	out = appendNorm(out, state.RoundNumber, 30)
	out = appendFactionOneHot(out, state.FactionTurn)
	out = appendOneHot(out, int(state.CurrentPhase), phaseCount)
	out = appendOneHot(out, int(state.CurrentStep), stepCount)
	out = appendTurnOrder(out, state.TurnOrder)
	out = appendFactionScores(out, state.VictoryPoints)
	out = appendDominance(out, state.ActiveDominance)
	out = appendCoalition(out, state.CoalitionActive, state.CoalitionPartner, state.WinningCoalition)
	out = appendHiddenCounts(out, input.hidden, state)
	out = appendItemSupply(out, state.ItemSupply)
	out = appendCraftedItems(out, state.CraftedItems)
	out = appendPersistentEffects(out, state.PersistentEffects)
	out = appendTurnProgress(out, state.TurnProgress)
	out = appendMapFeatures(out, state.Map)
	out = appendMarquiseState(out, state.Marquise)
	out = appendEyrieState(out, state.Eyrie)
	out = appendAllianceState(out, state.Alliance, input.perspective)
	out = appendVagabondState(out, state.Vagabond)

	return out
}

func appendTurnOrder(out []float32, turnOrder []game.Faction) []float32 {
	for slot := 0; slot < factionCount; slot++ {
		if slot >= len(turnOrder) {
			out = appendOneHot(out, -1, factionCount)
			continue
		}
		out = appendFactionOneHot(out, turnOrder[slot])
	}
	return out
}

func appendFactionScores(out []float32, points map[game.Faction]int) []float32 {
	for _, faction := range orderedFactions {
		out = appendNorm(out, points[faction], 30)
	}
	return out
}

func appendDominance(out []float32, active map[game.Faction]game.CardID) []float32 {
	for _, faction := range orderedFactions {
		out = appendBool(out, active[faction] != 0)
	}
	return out
}

func appendCoalition(out []float32, active bool, partner game.Faction, winners []game.Faction) []float32 {
	out = appendBool(out, active)
	out = appendOptionalFactionOneHot(out, partner, active)

	winnerSet := map[game.Faction]bool{}
	for _, faction := range winners {
		winnerSet[faction] = true
	}
	for _, faction := range orderedFactions {
		out = appendBool(out, winnerSet[faction])
	}
	return out
}

func appendHiddenCounts(out []float32, hidden rootengine.TrainingHiddenCounts, state rootengine.TrainingPublicState) []float32 {
	for _, faction := range orderedFactions {
		out = appendNorm(out, hidden.Hands[faction], 12)
	}
	out = appendNorm(out, hidden.AllianceSupporters, 20)
	out = appendNorm(out, hidden.Deck, baseDeckCardCount)
	out = appendNorm(out, hidden.QuestDeck, 15)
	out = appendNorm(out, len(state.DiscardPile), baseDeckCardCount)
	out = appendNorm(out, len(state.AvailableDominance), 4)
	out = appendNorm(out, len(state.QuestDiscard), 15)
	out = appendNorm(out, len(state.PendingFieldHospitals), 4)
	out = appendNorm(out, len(state.PendingOutrage), 4)
	return out
}

func appendItemSupply(out []float32, supply map[game.ItemType]int) []float32 {
	for _, itemType := range orderedItemTypes {
		out = appendNorm(out, supply[itemType], 2)
	}
	return out
}

func appendCraftedItems(out []float32, crafted map[game.Faction][]game.ItemType) []float32 {
	for _, faction := range orderedFactions {
		counts := make([]int, itemTypeCount)
		for _, itemType := range crafted[faction] {
			if index := itemTypeIndex(itemType); index >= 0 {
				counts[index]++
			}
		}
		for _, count := range counts {
			out = appendNorm(out, count, 4)
		}
	}
	return out
}

func appendPersistentEffects(out []float32, effects map[game.Faction][]game.CardID) []float32 {
	for _, faction := range orderedFactions {
		out = appendNorm(out, len(effects[faction]), 8)
	}
	return out
}

func appendTurnProgress(out []float32, progress game.TurnProgress) []float32 {
	out = appendNorm(out, progress.ActionsUsed, 6)
	out = appendNorm(out, progress.BonusActions, 6)
	out = appendNorm(out, progress.MarchesUsed, 2)
	out = appendBool(out, progress.RecruitUsed)
	out = appendNorm(out, len(progress.UsedWorkshopClearings), maxClearings)
	out = appendBool(out, progress.HasCrafted)
	out = appendNorm(out, progress.DecreeColumnsResolved, decreeColumnCount)
	out = appendNorm(out, progress.DecreeCardsResolved, 12)
	out = appendNorm(out, len(progress.ResolvedDecreeCardIDs), 12)
	out = appendNorm(out, progress.CardsAddedToDecree, 4)
	out = appendBool(out, progress.EyrieEmergencyResolved)
	out = appendBool(out, progress.EyrieNewRoostResolved)
	out = appendNorm(out, progress.OfficerActionsUsed, 10)
	out = appendBool(out, progress.HasOrganized)
	out = appendBool(out, progress.HasRefreshed)
	out = appendBool(out, progress.HasSlipped)
	out = appendBool(out, progress.VagabondRestResolved)
	out = appendBool(out, progress.VagabondEveningDrawn)
	for _, faction := range orderedFactions {
		out = appendNorm(out, progress.VagabondAidCounts[faction], 3)
	}
	out = appendBool(out, progress.VagabondDiscardResolved)
	out = appendBool(out, progress.VagabondCapacityChecked)
	out = appendBool(out, progress.EveningDrawn)
	out = appendBool(out, progress.EveningDiscardResolved)
	out = appendNorm(out, len(progress.UsedPersistentEffectIDs), 8)
	out = appendBool(out, progress.BirdsongMainActionTaken)
	out = appendBool(out, progress.SpreadSympathyStarted)
	out = appendBool(out, progress.DaylightMainActionTaken)
	out = appendBool(out, progress.EveningMainActionTaken)
	return out
}

func appendMapFeatures(out []float32, board game.Map) []float32 {
	clearings := map[int]game.Clearing{}
	for _, clearing := range board.Clearings {
		clearings[clearing.ID] = clearing
	}
	for clearingID := 1; clearingID <= maxClearings; clearingID++ {
		out = appendClearing(out, clearings[clearingID])
	}

	forests := map[int]game.Forest{}
	for _, forest := range board.Forests {
		forests[forest.ID] = forest
	}
	for forestID := 1; forestID <= maxForests; forestID++ {
		out = appendForest(out, forests[forestID])
	}
	return out
}

func appendClearing(out []float32, clearing game.Clearing) []float32 {
	out = appendBool(out, clearing.ID != 0)
	out = appendOneHot(out, int(clearing.Suit), suitCount)
	out = appendNorm(out, clearing.BuildSlots, 3)
	out = appendBool(out, clearing.Ruins)
	out = appendNorm(out, clearing.Wood, 8)

	ruinItems := make([]int, itemTypeCount)
	for _, itemType := range clearing.RuinItems {
		if index := itemTypeIndex(itemType); index >= 0 {
			ruinItems[index]++
		}
	}
	for _, count := range ruinItems {
		out = appendNorm(out, count, 4)
	}

	for _, faction := range orderedFactions {
		out = appendNorm(out, clearing.Warriors[faction], 25)
	}

	buildings := make(map[game.Faction][]int, factionCount)
	for _, faction := range orderedFactions {
		buildings[faction] = make([]int, buildingTypeCount)
	}
	for _, building := range clearing.Buildings {
		if index := buildingTypeIndex(building.Type); index >= 0 {
			buildings[building.Faction][index]++
		}
	}
	for _, faction := range orderedFactions {
		for _, count := range buildings[faction] {
			out = appendNorm(out, count, 3)
		}
	}

	tokens := make(map[game.Faction][]int, factionCount)
	for _, faction := range orderedFactions {
		tokens[faction] = make([]int, tokenTypeCount)
	}
	for _, token := range clearing.Tokens {
		if index := tokenTypeIndex(token.Type); index >= 0 {
			tokens[token.Faction][index]++
		}
	}
	for _, faction := range orderedFactions {
		for _, count := range tokens[faction] {
			out = appendNorm(out, count, 2)
		}
	}

	adjacent := map[int]bool{}
	for _, clearingID := range clearing.Adj {
		adjacent[clearingID] = true
	}
	for clearingID := 1; clearingID <= maxClearings; clearingID++ {
		out = appendBool(out, adjacent[clearingID])
	}
	return out
}

func appendForest(out []float32, forest game.Forest) []float32 {
	out = appendBool(out, forest.ID != 0)
	adjacent := map[int]bool{}
	for _, clearingID := range forest.AdjacentClearings {
		adjacent[clearingID] = true
	}
	for clearingID := 1; clearingID <= maxClearings; clearingID++ {
		out = appendBool(out, adjacent[clearingID])
	}
	return out
}

func appendMarquiseState(out []float32, state game.MarquiseState) []float32 {
	out = appendCardSummary(out, state.CardsInHand)
	out = appendNorm(out, state.WarriorSupply, 25)
	out = appendNorm(out, state.WoodSupply, 8)
	out = appendNorm(out, state.SawmillsPlaced, 6)
	out = appendNorm(out, state.WorkshopsPlaced, 6)
	out = appendNorm(out, state.RecruitersPlaced, 6)
	out = appendOneHot(out, clearingSlot(state.KeepClearingID), maxClearings)
	return out
}

func appendEyrieState(out []float32, state game.EyrieState) []float32 {
	out = appendCardSummary(out, state.CardsInHand)
	out = appendNorm(out, state.WarriorSupply, 20)
	out = appendNorm(out, state.RoostsPlaced, 7)
	out = appendOptionalLeader(out, state.Leader)
	available := map[game.EyrieLeader]bool{}
	for _, leader := range state.AvailableLeaders {
		available[leader] = true
	}
	for _, leader := range game.AllEyrieLeaders() {
		out = appendBool(out, available[leader])
	}
	out = appendNorm(out, len(state.Decree.Recruit), 12)
	out = appendNorm(out, len(state.Decree.Move), 12)
	out = appendNorm(out, len(state.Decree.Battle), 12)
	out = appendNorm(out, len(state.Decree.Build), 12)
	out = appendBool(out, state.CraftedThisTurn)
	return out
}

func appendAllianceState(out []float32, state game.AllianceState, perspective game.Faction) []float32 {
	out = appendCardSummary(out, state.CardsInHand)
	if perspective == game.Alliance {
		out = appendCardSummary(out, state.Supporters)
	} else {
		out = appendCardSummary(out, nil)
	}
	out = appendNorm(out, state.WarriorSupply, 10)
	out = appendNorm(out, len(state.Supporters), 20)
	out = appendNorm(out, state.Officers, 10)
	out = appendBool(out, state.FoxBasePlaced)
	out = appendBool(out, state.RabbitBasePlaced)
	out = appendBool(out, state.MouseBasePlaced)
	out = appendNorm(out, state.SympathyPlaced, 10)
	return out
}

func appendVagabondState(out []float32, state game.VagabondState) []float32 {
	out = appendCardSummary(out, state.CardsInHand)
	out = appendOptionalCharacter(out, state.Character)
	out = appendOneHot(out, clearingSlot(state.ClearingID), maxClearings)
	out = appendOneHot(out, forestSlot(state.ForestID), maxForests)
	out = appendBool(out, state.InForest)

	itemCounts := make([][]int, itemTypeCount)
	for i := range itemCounts {
		itemCounts[i] = make([]int, itemStatusCount)
	}
	for _, item := range state.Items {
		typeIndex := itemTypeIndex(item.Type)
		statusIndex := int(item.Status)
		if typeIndex >= 0 && statusIndex >= 0 && statusIndex < itemStatusCount {
			itemCounts[typeIndex][statusIndex]++
		}
	}
	for _, byStatus := range itemCounts {
		for _, count := range byStatus {
			out = appendNorm(out, count, 8)
		}
	}

	for _, faction := range orderedFactions {
		if faction == game.Vagabond {
			out = appendOneHot(out, -1, relationshipCount)
			continue
		}
		out = appendOneHot(out, int(state.Relationships[faction]), relationshipCount)
	}
	out = appendNorm(out, len(state.QuestsCompleted), 15)
	out = appendQuestSuitCounts(out, state.QuestsCompleted, 15)
	out = appendNorm(out, len(state.QuestsAvailable), 3)
	out = appendQuestSuitCounts(out, state.QuestsAvailable, 3)
	return out
}

func appendCardSummary(out []float32, cards []game.Card) []float32 {
	out = appendNorm(out, len(cards), 12)
	suits := make([]int, suitCount)
	kinds := make([]int, cardKindCount)
	for _, card := range cards {
		if int(card.Suit) >= 0 && int(card.Suit) < suitCount {
			suits[card.Suit]++
		}
		if int(card.Kind) >= 0 && int(card.Kind) < cardKindCount {
			kinds[card.Kind]++
		}
	}
	for _, count := range suits {
		out = appendNorm(out, count, 12)
	}
	for _, count := range kinds {
		out = appendNorm(out, count, 12)
	}
	return out
}

func appendQuestSuitCounts(out []float32, quests []game.Quest, scale float32) []float32 {
	counts := make([]int, suitCount)
	for _, quest := range quests {
		if int(quest.Suit) >= 0 && int(quest.Suit) < suitCount {
			counts[quest.Suit]++
		}
	}
	for _, count := range counts {
		out = appendNorm(out, count, scale)
	}
	return out
}

func appendOptionalLeader(out []float32, leader game.EyrieLeader) []float32 {
	if int(leader) < 0 || int(leader) >= leaderCount {
		return appendOneHot(out, 0, leaderCount+1)
	}
	return appendOneHot(out, int(leader)+1, leaderCount+1)
}

func appendOptionalCharacter(out []float32, character game.VagabondCharacter) []float32 {
	if int(character) < 0 || int(character) >= characterCount {
		return appendOneHot(out, 0, characterCount+1)
	}
	return appendOneHot(out, int(character)+1, characterCount+1)
}
