package engine

import (
	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

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
	state.TurnProgress.CardsAddedToDecree = len(action.AddToDecree.CardIDs)

	for i, cardID := range action.AddToDecree.CardIDs {
		if i >= len(action.AddToDecree.Columns) {
			break
		}

		if _, ok := spendFactionHandCard(state, game.Eyrie, cardID); !ok {
			continue
		}

		appendCardToDecree(&state.Eyrie.Decree, action.AddToDecree.Columns[i], cardID)
	}
}

func applyEyrieEmergencyOrders(state *game.GameState, action game.Action) {
	if action.EyrieEmergency == nil {
		return
	}

	DrawCards(state, action.EyrieEmergency.Faction, action.EyrieEmergency.Count)
}

func eyrieHasRoostOnMap(state game.GameState) bool {
	for _, clearing := range state.Map.Clearings {
		for _, building := range clearing.Buildings {
			if building.Faction == game.Eyrie && building.Type == game.Roost {
				return true
			}
		}
	}
	return false
}

func clearingHasOpenBuildSlot(clearing game.Clearing) bool {
	usedSlots := len(clearing.Buildings)
	if clearing.Ruins {
		usedSlots++
	}
	return usedSlots < clearing.BuildSlots
}

func applyEyrieNewRoost(state *game.GameState, action game.Action) {
	if action.EyrieNewRoost == nil || eyrieHasRoostOnMap(*state) || state.Eyrie.WarriorSupply < 3 || state.Eyrie.RoostsPlaced >= 7 {
		return
	}

	index := findClearingIndex(state.Map, action.EyrieNewRoost.ClearingID)
	if index == -1 || !clearingHasOpenBuildSlot(state.Map.Clearings[index]) {
		return
	}

	state.Map.Clearings[index].Buildings = append(state.Map.Clearings[index].Buildings, game.Building{
		Faction: game.Eyrie,
		Type:    game.Roost,
	})
	if state.Map.Clearings[index].Warriors == nil {
		state.Map.Clearings[index].Warriors = map[game.Faction]int{}
	}
	state.Map.Clearings[index].Warriors[game.Eyrie] += 3
	state.Eyrie.WarriorSupply -= 3
	state.Eyrie.RoostsPlaced++
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

func hasAvailableNewLeader(state game.GameState) bool {
	for _, leader := range state.Eyrie.AvailableLeaders {
		if leader != state.Eyrie.Leader {
			return true
		}
	}
	return false
}

func recycleEyrieLeadersIfNeeded(state *game.GameState) {
	if hasAvailableNewLeader(*state) {
		return
	}
	state.Eyrie.AvailableLeaders = game.AllEyrieLeaders()
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
	recycleEyrieLeadersIfNeeded(state)
	state.Eyrie.AvailableLeaders = removeLeader(state.Eyrie.AvailableLeaders, state.Eyrie.Leader)
	state.Eyrie.AvailableLeaders = removeLeader(state.Eyrie.AvailableLeaders, action.Turmoil.NewLeader)
	state.Eyrie.Leader = action.Turmoil.NewLeader
	state.Eyrie.Decree = game.Decree{}
	state.TurnProgress.ResolvedDecreeCardIDs = nil

	vizierColumns := vizierColumnsForLeader(action.Turmoil.NewLeader)
	appendCardToDecree(&state.Eyrie.Decree, vizierColumns[0], game.LoyalVizier1)
	appendCardToDecree(&state.Eyrie.Decree, vizierColumns[1], game.LoyalVizier2)
}
