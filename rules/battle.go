package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidBattles(faction game.Faction, m game.Map) []game.Action {
	battles := []game.Action{}
	for _, c := range m.Clearings {
		available := c.Warriors[faction]
		if available <= 0 {
			continue
		}

		seenTargets := map[game.Faction]bool{}
		for target, count := range c.Warriors {
			if target == faction || count <= 0 || seenTargets[target] {
				continue
			}

			battles = append(battles, game.Action{
				Type: game.ActionBattle,
				Battle: &game.BattleAction{
					Faction:       faction,
					ClearingID:    c.ID,
					TargetFaction: target,
				},
			})
			seenTargets[target] = true
		}

		for target, count := range c.Buildings {
			if target == faction || count <= 0 || seenTargets[target] {
				continue
			}

			battles = append(battles, game.Action{
				Type: game.ActionBattle,
				Battle: &game.BattleAction{
					Faction:       faction,
					ClearingID:    c.ID,
					TargetFaction: target,
				},
			})
			seenTargets[target] = true
		}
	}
	return battles
}
