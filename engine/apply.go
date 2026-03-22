package engine

import "github.com/imdehydrated/rootbuddy/game"

func ApplyAction(state game.GameState, action game.Action) game.GameState {
	next := cloneState(state)

	switch action.Type {
	case game.ActionRecruit:
		applyRecruit(&next, action)
	case game.ActionMovement:
		applyMovement(&next, action)
	case game.ActionBattleResolution:
		applyBattleResolution(&next, action)
	case game.ActionBuild:
		applyBuild(&next, action)
	case game.ActionOverwork:
		applyOverwork(&next, action)
	case game.ActionCraft:
		applyCraft(&next, action)
	}

	advanceTurnState(&next, action)

	return next
}

func cloneState(state game.GameState) game.GameState {
	next := state

	next.Map.Clearings = make([]game.Clearing, len(state.Map.Clearings))
	for i, clearing := range state.Map.Clearings {
		cloned := clearing

		if clearing.Adj != nil {
			cloned.Adj = make([]int, len(clearing.Adj))
			copy(cloned.Adj, clearing.Adj)
		}

		if clearing.Warriors != nil {
			cloned.Warriors = make(map[game.Faction]int, len(clearing.Warriors))
			for faction, count := range clearing.Warriors {
				cloned.Warriors[faction] = count
			}
		}

		if clearing.Buildings != nil {
			cloned.Buildings = make([]game.Building, len(clearing.Buildings))
			copy(cloned.Buildings, clearing.Buildings)
		}

		next.Map.Clearings[i] = cloned
	}

	if state.Marquise.CardsInHand != nil {
		next.Marquise.CardsInHand = make([]game.Card, len(state.Marquise.CardsInHand))
		copy(next.Marquise.CardsInHand, state.Marquise.CardsInHand)
	}

	if state.TurnProgress.UsedWorkshopClearings != nil {
		next.TurnProgress.UsedWorkshopClearings = make([]int, len(state.TurnProgress.UsedWorkshopClearings))
		copy(next.TurnProgress.UsedWorkshopClearings, state.TurnProgress.UsedWorkshopClearings)
	}

	return next
}

func findClearingIndex(m game.Map, id int) int {
	for i, clearing := range m.Clearings {
		if clearing.ID == id {
			return i
		}
	}
	return -1
}

func removeCardByID(cards []game.Card, id game.CardID) []game.Card {
	for i, card := range cards {
		if card.ID == id {
			return append(cards[:i], cards[i+1:]...)
		}
	}
	return cards
}

func applyRecruit(state *game.GameState, action game.Action) {
	if action.Recruit == nil {
		return
	}

	for _, clearingID := range action.Recruit.ClearingIDs {
		index := findClearingIndex(state.Map, clearingID)
		if index == -1 {
			continue
		}

		if state.Map.Clearings[index].Warriors == nil {
			state.Map.Clearings[index].Warriors = map[game.Faction]int{}
		}

		state.Map.Clearings[index].Warriors[game.Marquise]++
		state.Marquise.WarriorSupply--
	}

	state.TurnProgress.RecruitUsed = true
}

func applyMovement(state *game.GameState, action game.Action) {
	if action.Movement == nil {
		return
	}

	fromIndex := findClearingIndex(state.Map, action.Movement.From)
	toIndex := findClearingIndex(state.Map, action.Movement.To)
	if fromIndex == -1 || toIndex == -1 {
		return
	}

	if state.Map.Clearings[fromIndex].Warriors == nil {
		return
	}

	moved := action.Movement.MaxCount
	state.Map.Clearings[fromIndex].Warriors[action.Movement.Faction] -= moved

	if state.Map.Clearings[toIndex].Warriors == nil {
		state.Map.Clearings[toIndex].Warriors = map[game.Faction]int{}
	}
	state.Map.Clearings[toIndex].Warriors[action.Movement.Faction] += moved
}

func removeWarriorLosses(clearing *game.Clearing, faction game.Faction, losses int) int {
	if losses <= 0 || clearing.Warriors == nil {
		return losses
	}

	available := clearing.Warriors[faction]
	if available <= 0 {
		return losses
	}

	removed := losses
	if removed > available {
		removed = available
	}

	clearing.Warriors[faction] = available - removed
	return losses - removed
}

func decrementPlacedBuildingCounter(state *game.GameState, buildingType game.BuildingType) {
	switch buildingType {
	case game.Sawmill:
		if state.Marquise.SawmillsPlaced > 0 {
			state.Marquise.SawmillsPlaced--
		}
	case game.Workshop:
		if state.Marquise.WorkshopsPlaced > 0 {
			state.Marquise.WorkshopsPlaced--
		}
	case game.Recruiter:
		if state.Marquise.RecruitersPlaced > 0 {
			state.Marquise.RecruitersPlaced--
		}
	}
}

func removeBuildingLosses(state *game.GameState, clearing *game.Clearing, faction game.Faction, losses int) {
	if losses <= 0 || len(clearing.Buildings) == 0 {
		return
	}

	remaining := make([]game.Building, 0, len(clearing.Buildings))
	for _, building := range clearing.Buildings {
		if losses > 0 && building.Faction == faction {
			if faction == game.Marquise {
				decrementPlacedBuildingCounter(state, building.Type)
			}
			losses--
			continue
		}
		remaining = append(remaining, building)
	}
	clearing.Buildings = remaining
}

func applyBattleResolution(state *game.GameState, action game.Action) {
	if action.BattleResolution == nil {
		return
	}

	index := findClearingIndex(state.Map, action.BattleResolution.ClearingID)
	if index == -1 {
		return
	}

	clearing := &state.Map.Clearings[index]
	removeWarriorLosses(clearing, action.BattleResolution.Faction, action.BattleResolution.AttackerLosses)
	remainingDefenderLosses := removeWarriorLosses(clearing, action.BattleResolution.TargetFaction, action.BattleResolution.DefenderLosses)
	removeBuildingLosses(state, clearing, action.BattleResolution.TargetFaction, remainingDefenderLosses)
}

func applyBuild(state *game.GameState, action game.Action) {
	if action.Build == nil {
		return
	}

	index := findClearingIndex(state.Map, action.Build.ClearingID)
	if index == -1 {
		return
	}

	state.Map.Clearings[index].Buildings = append(
		state.Map.Clearings[index].Buildings,
		game.Building{
			Faction:      action.Build.Faction,
			Type:         action.Build.BuildingType,
		},
	)

	switch action.Build.BuildingType {
	case game.Sawmill:
		state.Marquise.SawmillsPlaced++
	case game.Workshop:
		state.Marquise.WorkshopsPlaced++
	case game.Recruiter:
		state.Marquise.RecruitersPlaced++
	}
}

func applyOverwork(state *game.GameState, action game.Action) {
	if action.Overwork == nil {
		return
	}

	index := findClearingIndex(state.Map, action.Overwork.ClearingID)
	if index == -1 {
		return
	}

	state.Map.Clearings[index].Wood++
	state.Marquise.CardsInHand = removeCardByID(state.Marquise.CardsInHand, action.Overwork.CardID)
}

func applyCraft(state *game.GameState, action game.Action) {
	if action.Craft == nil {
		return
	}

	state.Marquise.CardsInHand = removeCardByID(state.Marquise.CardsInHand, action.Craft.CardID)
	state.TurnProgress.UsedWorkshopClearings = append(
		state.TurnProgress.UsedWorkshopClearings,
		action.Craft.UsedWorkshopClearings...,
	)
}

func advanceTurnState(state *game.GameState, action game.Action) {
	switch action.Type {
	case game.ActionRecruit:
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.StepDaylightActions
	}
}
