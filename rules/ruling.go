package rules

import "github.com/imdehydrated/rootbuddy/game"

func buildingCountByFaction(c game.Clearing, faction game.Faction) int {
	count := 0
	for _, building := range c.Buildings {
		if building.Faction == faction {
			count++
		}
	}
	return count
}

func Ruler(c game.Clearing) (game.Faction, bool) {
	scores := map[game.Faction]int{}

	for faction, warriors := range c.Warriors {
		if faction == game.Vagabond {
			continue
		}
		scores[faction] += warriors
	}

	for _, building := range c.Buildings {
		scores[building.Faction]++
	}

	var ruler game.Faction
	maxScore := 0
	hasRuler := false
	tied := false

	for faction, score := range scores {
		if score <= 0 {
			continue
		}
		if !hasRuler || score > maxScore {
			ruler = faction
			maxScore = score
			hasRuler = true
			tied = false
		} else if score == maxScore {
			tied = true
		}
	}

	if !hasRuler {
		return 0, false
	}
	if tied {
		if scores[game.Eyrie] == maxScore {
			return game.Eyrie, true
		}
		return 0, false
	}
	return ruler, true
}
