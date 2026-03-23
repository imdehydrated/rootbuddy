package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidEyrieBuildActions(state game.GameState, cardID game.CardID) []game.Action {
	actions := []game.Action{}
	if state.Eyrie.RoostsPlaced >= 7 {
		return actions
	}

	for _, clearing := range state.Map.Clearings {
		if !decreeMatchesSuit(cardID, clearing.Suit) {
			continue
		}

		hasRoost := false
		for _, building := range clearing.Buildings {
			if building.Faction == game.Eyrie && building.Type == game.Roost {
				hasRoost = true
				break
			}
		}
		if hasRoost {
			continue
		}

		ruler, ruled := Ruler(clearing)
		if !ruled || ruler != game.Eyrie {
			continue
		}

		usedSlots := len(clearing.Buildings)
		if clearing.Ruins {
			usedSlots++
		}
		if usedSlots >= clearing.BuildSlots {
			continue
		}

		actions = append(actions, game.Action{
			Type: game.ActionBuild,
			Build: &game.BuildAction{
				Faction:      game.Eyrie,
				ClearingID:   clearing.ID,
				BuildingType: game.Roost,
				DecreeCardID: cardID,
			},
		})
	}

	return actions
}
