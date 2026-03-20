package rules

import "github.com/imdehydrated/rootbuddy/game"

func findClearingByID(m game.Map, id int) (game.Clearing, bool) {
	for _, c := range m.Clearings {
		if c.ID == id {
			return c, true
		}
	}
	return game.Clearing{}, false
}

func ValidMovementActions(faction game.Faction, m game.Map) []game.Action {
	moves := []game.Action{}
	for _, origin := range m.Clearings {
		available := origin.Warriors[faction]
		if available <= 0 {
			continue
		}

		originRuler, originRuled := Ruler(origin)

		for _, adjid := range origin.Adj {
			destination, ok := findClearingByID(m, adjid)
			if !ok {
				continue
			}
			destinationRuler, destinationRuled := Ruler(destination)

			rulesOrigin := originRuled && originRuler == faction
			rulesDestination := destinationRuled && destinationRuler == faction
			if !rulesOrigin && !rulesDestination {
				continue
			}

			moves = append(moves, game.Action{
				Type: game.ActionMovement,
				Movement: &game.MovementAction{
					Faction:  faction,
					MaxCount: available,
					From:     origin.ID,
					To:       destination.ID,
				},
			})
		}
	}
	return moves
}
