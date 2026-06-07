package engine

import "github.com/imdehydrated/rootbuddy/game"

var marquiseBuildingTrack = []int{0, 1, 2, 3, 3, 4}
var allianceSympathyTrack = []int{0, 1, 1, 1, 2, 2, 2, 3, 3, 4}

func addVictoryPoints(state *game.GameState, faction game.Faction, points int) {
	if points <= 0 {
		return
	}
	if hasActiveDominance(*state, faction) {
		return
	}
	if faction == game.Vagabond && state.CoalitionActive {
		return
	}

	if state.VictoryPoints == nil {
		state.VictoryPoints = map[game.Faction]int{}
	}

	state.VictoryPoints[faction] += points
}

func victoryPointSnapshot(state game.GameState) map[game.Faction]int {
	if state.VictoryPoints == nil {
		return nil
	}

	snapshot := make(map[game.Faction]int, len(state.VictoryPoints))
	for faction, points := range state.VictoryPoints {
		snapshot[faction] = points
	}
	return snapshot
}

func resolveVictoryPointWin(state *game.GameState, before map[game.Faction]int) {
	if state.GamePhase == game.LifecycleGameOver || state.VictoryPoints == nil {
		return
	}

	winners := []game.Faction{}
	for faction, points := range state.VictoryPoints {
		if points < 30 {
			continue
		}
		if before != nil && before[faction] >= 30 {
			continue
		}
		winners = append(winners, faction)
	}
	if len(winners) == 0 {
		return
	}

	setWinner(state, chooseVictoryPointWinner(*state, winners))
}

func chooseVictoryPointWinner(state game.GameState, winners []game.Faction) game.Faction {
	for _, faction := range winners {
		if faction == state.FactionTurn {
			return faction
		}
	}

	for _, faction := range effectiveTurnOrder(state) {
		for _, winner := range winners {
			if winner == faction {
				return winner
			}
		}
	}

	return winners[0]
}

func scoreMarquiseBuilding(state *game.GameState, buildingType game.BuildingType, alreadyPlaced int) {
	if alreadyPlaced < 0 || alreadyPlaced >= len(marquiseBuildingTrack) {
		return
	}

	addVictoryPoints(state, game.Marquise, marquiseBuildingTrack[alreadyPlaced])
}

func scoreBattleRemovals(state *game.GameState, faction game.Faction, removedBuildings int, removedTokens int) {
	points := removedBuildings + removedTokens
	if points > 0 && faction == game.Eyrie && state.Eyrie.Leader == game.LeaderDespot {
		points++
	}
	addVictoryPoints(state, faction, points)
}

func scoreAllianceSympathy(state *game.GameState, alreadyPlaced int) {
	if alreadyPlaced < 0 || alreadyPlaced >= len(allianceSympathyTrack) {
		return
	}

	addVictoryPoints(state, game.Alliance, allianceSympathyTrack[alreadyPlaced])
}
