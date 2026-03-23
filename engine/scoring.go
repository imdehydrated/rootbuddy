package engine

import "github.com/imdehydrated/rootbuddy/game"

var marquiseBuildingTrack = []int{0, 1, 2, 3, 3, 4}
var allianceSympathyTrack = []int{0, 1, 1, 1, 2, 2, 2, 3, 3, 4}

func addVictoryPoints(state *game.GameState, faction game.Faction, points int) {
	if points <= 0 {
		return
	}

	if state.VictoryPoints == nil {
		state.VictoryPoints = map[game.Faction]int{}
	}

	state.VictoryPoints[faction] += points
}

func scoreMarquiseBuilding(state *game.GameState, buildingType game.BuildingType, alreadyPlaced int) {
	if alreadyPlaced < 0 || alreadyPlaced >= len(marquiseBuildingTrack) {
		return
	}

	addVictoryPoints(state, game.Marquise, marquiseBuildingTrack[alreadyPlaced])
}

func scoreBattleRemovals(state *game.GameState, faction game.Faction, removedBuildings int, removedTokens int) {
	points := removedBuildings + removedTokens
	if faction == game.Eyrie && state.Eyrie.Leader == game.LeaderDespot {
		points += removedBuildings + removedTokens
	}
	addVictoryPoints(state, faction, points)
}

func scoreAllianceSympathy(state *game.GameState, alreadyPlaced int) {
	if alreadyPlaced < 0 || alreadyPlaced >= len(allianceSympathyTrack) {
		return
	}

	addVictoryPoints(state, game.Alliance, allianceSympathyTrack[alreadyPlaced])
}
