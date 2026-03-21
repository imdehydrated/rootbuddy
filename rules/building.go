package rules

import "github.com/imdehydrated/rootbuddy/game"

func buildCost(placed int) int {
	costs := []int{0, 1, 2, 3, 3, 4}
	if placed < 0 || placed >= len(costs) {
		return -1
	}
	return costs[placed]
}

func availableBuildings(marquise game.MarquiseState) []game.BuildingType {
	buildings := []game.BuildingType{}

	if marquise.SawmillsPlaced < 6 {
		buildings = append(buildings, game.Sawmill)
	}
	if marquise.WorkshopsPlaced < 6 {
		buildings = append(buildings, game.Workshop)
	}
	if marquise.RecruitersPlaced < 6 {
		buildings = append(buildings, game.Recruiter)
	}

	return buildings
}

func placedCount(marquise game.MarquiseState, buildingType game.BuildingType) int {
	switch buildingType {
	case game.Sawmill:
		return marquise.SawmillsPlaced
	case game.Workshop:
		return marquise.WorkshopsPlaced
	case game.Recruiter:
		return marquise.RecruitersPlaced
	default:
		return -1
	}
}

func woodNetworkWood(startID int, m game.Map) int {
	total := 0
	queue := []int{startID}
	visited := map[int]bool{}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		c, ok := findClearingByID(m, currentID)
		if !ok {
			continue
		}

		ruler, ruled := Ruler(c)
		if !ruled || ruler != game.Marquise {
			continue
		}

		total += c.Wood

		for _, adjID := range c.Adj {
			if !visited[adjID] {
				queue = append(queue, adjID)
			}
		}
	}
	return total
}

func ValidBuilds(m game.Map, marquise game.MarquiseState) []game.Action {
	actions := []game.Action{}

	for _, c := range m.Clearings {
		ruler, ok := Ruler(c)
		if !ok || ruler != game.Marquise {
			continue
		}

		used := len(c.Buildings)
		if c.Ruins {
			used++
		}
		openSlots := c.BuildSlots - used
		if openSlots <= 0 {
			continue
		}

		networkWood := woodNetworkWood(c.ID, m)

		for _, buildingType := range availableBuildings(marquise) {
			placed := placedCount(marquise, buildingType)
			cost := buildCost(placed)
			if cost < 0 {
				continue
			}
			if networkWood < cost {
				continue
			}

			actions = append(actions, game.Action{
				Type: game.ActionBuild,
				Build: &game.BuildAction{
					Faction:      game.Marquise,
					ClearingID:   c.ID,
					BuildingType: buildingType,
				},
			})
		}
	}
	return actions
}
