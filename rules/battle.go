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

		for _, building := range c.Buildings {
			target := building.Faction
			if target == faction || seenTargets[target] {
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

		for _, token := range c.Tokens {
			target := token.Faction
			if target == faction || seenTargets[target] {
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

		if c.Wood > 0 && faction != game.Marquise && !seenTargets[game.Marquise] {
			battles = append(battles, game.Action{
				Type: game.ActionBattle,
				Battle: &game.BattleAction{
					Faction:       faction,
					ClearingID:    c.ID,
					TargetFaction: game.Marquise,
				},
			})
		}
	}
	return battles
}

func ValidBattlesInState(faction game.Faction, state game.GameState) []game.Action {
	battles := ValidBattles(faction, state.Map)
	if faction == game.Vagabond || state.Vagabond.InForest || state.Vagabond.ClearingID == 0 {
		return battles
	}

	for _, clearing := range state.Map.Clearings {
		if clearing.ID != state.Vagabond.ClearingID || clearing.Warriors[faction] <= 0 {
			continue
		}

		battles = append(battles, game.Action{
			Type: game.ActionBattle,
			Battle: &game.BattleAction{
				Faction:       faction,
				ClearingID:    clearing.ID,
				TargetFaction: game.Vagabond,
			},
		})
		break
	}

	return battles
}
