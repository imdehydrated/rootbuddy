package rules

import (
	"sort"

	"github.com/imdehydrated/rootbuddy/game"
)

func ValidBattles(faction game.Faction, m game.Map) []game.Action {
	battles := []game.Action{}
	for _, c := range m.Clearings {
		available := c.Warriors[faction]
		if available <= 0 {
			continue
		}

		targets := map[game.Faction]bool{}
		for target, count := range c.Warriors {
			if target == faction || count <= 0 {
				continue
			}

			targets[target] = true
		}

		for _, building := range c.Buildings {
			target := building.Faction
			if target == faction {
				continue
			}

			targets[target] = true
		}

		for _, token := range c.Tokens {
			target := token.Faction
			if target == faction {
				continue
			}

			targets[target] = true
		}

		if c.Wood > 0 && faction != game.Marquise {
			targets[game.Marquise] = true
		}

		targetFactions := make([]game.Faction, 0, len(targets))
		for target := range targets {
			targetFactions = append(targetFactions, target)
		}
		sort.Slice(targetFactions, func(i, j int) bool {
			return targetFactions[i] < targetFactions[j]
		})

		for _, target := range targetFactions {
			battles = append(battles, game.Action{
				Type: game.ActionBattle,
				Battle: &game.BattleAction{
					Faction:       faction,
					ClearingID:    c.ID,
					TargetFaction: target,
				},
			})
		}
	}
	return battles
}

func ValidBattlesInState(faction game.Faction, state game.GameState) []game.Action {
	battles := []game.Action{}
	for _, action := range ValidBattles(faction, state.Map) {
		if action.Battle == nil || !game.AreEnemies(state, faction, action.Battle.TargetFaction) {
			continue
		}
		battles = append(battles, action)
	}
	if faction == game.Vagabond || state.Vagabond.InForest || state.Vagabond.ClearingID == 0 {
		return battles
	}

	for _, clearing := range state.Map.Clearings {
		if clearing.ID != state.Vagabond.ClearingID || clearing.Warriors[faction] <= 0 {
			continue
		}

		if game.AreEnemies(state, faction, game.Vagabond) {
			battles = append(battles, game.Action{
				Type: game.ActionBattle,
				Battle: &game.BattleAction{
					Faction:       faction,
					ClearingID:    clearing.ID,
					TargetFaction: game.Vagabond,
				},
			})
		}
		break
	}

	return battles
}
