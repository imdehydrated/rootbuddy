package rules

import "github.com/imdehydrated/rootbuddy/game"

func Ruler(c game.Clearing) (game.Faction, bool) {
	scores := map[game.Faction]int{}

	for faction, warriors := range c.Warriors {
		if faction == game.Vagabond {
			continue
		}
		scores[faction] += warriors
	}

	for faction, buildings := range c.Buildings {
		scores[faction] += buildings
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

	if !hasRuler || tied {
		return 0, false
	}
	return ruler, true
}
