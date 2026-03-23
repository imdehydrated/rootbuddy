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

func ruledWoodNetwork(startID int, m game.Map) []game.Clearing {
	network := []game.Clearing{}
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

		network = append(network, c)

		for _, adjID := range c.Adj {
			if !visited[adjID] {
				queue = append(queue, adjID)
			}
		}
	}

	return network
}

func woodSourcesForBuild(startID int, cost int, m game.Map) ([]game.WoodSource, bool) {
	if cost == 0 {
		return []game.WoodSource{}, true
	}

	remaining := cost
	sources := []game.WoodSource{}

	for _, clearing := range ruledWoodNetwork(startID, m) {
		if clearing.Wood <= 0 {
			continue
		}

		used := clearing.Wood
		if used > remaining {
			used = remaining
		}

		sources = append(sources, game.WoodSource{
			ClearingID: clearing.ID,
			Amount:     used,
		})
		remaining -= used

		if remaining == 0 {
			return sources, true
		}
	}

	return nil, false
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

		for _, buildingType := range availableBuildings(marquise) {
			placed := placedCount(marquise, buildingType)
			cost := buildCost(placed)
			if cost < 0 {
				continue
			}

			woodSources, ok := woodSourcesForBuild(c.ID, cost, m)
			if !ok {
				continue
			}

			actions = append(actions, game.Action{
				Type: game.ActionBuild,
				Build: &game.BuildAction{
					Faction:      game.Marquise,
					ClearingID:   c.ID,
					BuildingType: buildingType,
					WoodSources:  woodSources,
				},
			})
		}
	}
	return actions
}
