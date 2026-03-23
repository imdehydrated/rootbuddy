package rules

import (
	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

var decreeOrder = []game.DecreeColumn{
	game.DecreeRecruit,
	game.DecreeMove,
	game.DecreeBattle,
	game.DecreeBuild,
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

func cardSuitByID(id game.CardID) game.Suit {
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

func decreeMatchesSuit(cardID game.CardID, suit game.Suit) bool {
	cardSuit := cardSuitByID(cardID)
	return cardSuit == game.Bird || cardSuit == suit
}

func isResolvedDecreeCard(state game.GameState, cardID game.CardID) bool {
	for _, resolvedID := range state.TurnProgress.ResolvedDecreeCardIDs {
		if resolvedID == cardID {
			return true
		}
	}
	return false
}

func unresolvedDecreeCards(state game.GameState, column game.DecreeColumn) []game.CardID {
	cards := decreeCardsByColumn(state.Eyrie.Decree, column)
	unresolved := make([]game.CardID, 0, len(cards))
	for _, cardID := range cards {
		if !isResolvedDecreeCard(state, cardID) {
			unresolved = append(unresolved, cardID)
		}
	}
	return unresolved
}

func currentDecreeColumn(state game.GameState) (game.DecreeColumn, []game.CardID, bool) {
	for i := state.TurnProgress.DecreeColumnsResolved; i < len(decreeOrder); i++ {
		column := decreeOrder[i]
		unresolved := unresolvedDecreeCards(state, column)
		if len(unresolved) == 0 {
			continue
		}

		return column, unresolved, true
	}

	return 0, nil, false
}

func roostClearings(state game.GameState) []game.Clearing {
	clearings := []game.Clearing{}
	for _, clearing := range state.Map.Clearings {
		for _, building := range clearing.Buildings {
			if building.Faction == game.Eyrie && building.Type == game.Roost {
				clearings = append(clearings, clearing)
				break
			}
		}
	}
	return clearings
}

func roostCountInClearing(c game.Clearing) int {
	count := 0
	for _, building := range c.Buildings {
		if building.Faction == game.Eyrie && building.Type == game.Roost {
			count++
		}
	}
	return count
}

func leaderVizierColumns(leader game.EyrieLeader) [2]game.DecreeColumn {
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

func availableNewLeaders(state game.GameState) []game.EyrieLeader {
	leaders := []game.EyrieLeader{}
	for _, leader := range state.Eyrie.AvailableLeaders {
		if leader != state.Eyrie.Leader {
			leaders = append(leaders, leader)
		}
	}
	return leaders
}
