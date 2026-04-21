package engine

import "github.com/imdehydrated/rootbuddy/game"

var autumnCornerClearings = []int{1, 2, 3, 4}

var oppositeAutumnCorner = map[int]int{
	1: 3,
	2: 4,
	3: 1,
	4: 2,
}

func cornerClearings(mapID game.MapID) []int {
	if mapID == game.AutumnMapID {
		return append([]int(nil), autumnCornerClearings...)
	}
	return nil
}

func oppositeCornerClearing(mapID game.MapID, clearingID int) int {
	if mapID != game.AutumnMapID {
		return 0
	}
	return oppositeAutumnCorner[clearingID]
}

func marquiseSetupSites(state game.GameState, keepClearingID int) []int {
	sites := []int{keepClearingID}
	clearing := findClearing(&state, keepClearingID)
	if clearing == nil {
		return sites
	}
	sites = append(sites, clearing.Adj...)
	return sites
}

func generateMarquiseBuildingPlacements(
	keepClearingID int,
	sites []int,
	remaining map[int]int,
	assignments map[game.BuildingType]int,
	buildingTypes []game.BuildingType,
	index int,
	actions *[]game.Action,
) {
	if index >= len(buildingTypes) {
		*actions = append(*actions, game.Action{
			Type: game.ActionMarquiseSetup,
			MarquiseSetup: &game.MarquiseSetupAction{
				Faction:             game.Marquise,
				KeepClearingID:      keepClearingID,
				SawmillClearingID:   assignments[game.Sawmill],
				WorkshopClearingID:  assignments[game.Workshop],
				RecruiterClearingID: assignments[game.Recruiter],
			},
		})
		return
	}

	buildingType := buildingTypes[index]
	for _, site := range sites {
		if remaining[site] <= 0 {
			continue
		}
		remaining[site]--
		assignments[buildingType] = site
		generateMarquiseBuildingPlacements(keepClearingID, sites, remaining, assignments, buildingTypes, index+1, actions)
		remaining[site]++
	}
}

func ValidSetupActions(state game.GameState) []game.Action {
	if state.GamePhase != game.LifecycleSetup {
		return nil
	}

	switch state.SetupStage {
	case game.SetupStageMarquise:
		return validMarquiseSetupActions(state)
	case game.SetupStageEyrie:
		return validEyrieSetupActions(state)
	case game.SetupStageVagabond:
		return validVagabondSetupActions(state)
	default:
		return nil
	}
}

func validMarquiseSetupActions(state game.GameState) []game.Action {
	actions := []game.Action{}
	buildingTypes := []game.BuildingType{game.Sawmill, game.Workshop, game.Recruiter}

	for _, keepClearingID := range cornerClearings(state.Map.ID) {
		remaining := map[int]int{}
		sites := marquiseSetupSites(state, keepClearingID)
		assignments := map[game.BuildingType]int{}
		for _, site := range sites {
			clearing := findClearing(&state, site)
			if clearing == nil {
				continue
			}
			remaining[site] = clearing.BuildSlots
		}
		generateMarquiseBuildingPlacements(keepClearingID, sites, remaining, assignments, buildingTypes, 0, &actions)
	}

	return actions
}

func legalEyrieStartingClearings(state game.GameState) []int {
	corners := cornerClearings(state.Map.ID)
	blocked := map[int]bool{}
	if state.Marquise.KeepClearingID != 0 {
		blocked[state.Marquise.KeepClearingID] = true
		opposite := oppositeCornerClearing(state.Map.ID, state.Marquise.KeepClearingID)
		if opposite != 0 && !blocked[opposite] {
			return []int{opposite}
		}
	}

	legal := []int{}
	for _, corner := range corners {
		if blocked[corner] {
			continue
		}
		legal = append(legal, corner)
	}
	return legal
}

func validEyrieSetupActions(state game.GameState) []game.Action {
	clearings := legalEyrieStartingClearings(state)
	actions := make([]game.Action, 0, len(state.Eyrie.AvailableLeaders)*len(clearings))
	for _, leader := range state.Eyrie.AvailableLeaders {
		for _, clearingID := range clearings {
			actions = append(actions, game.Action{
				Type: game.ActionEyrieSetup,
				EyrieSetup: &game.EyrieSetupAction{
					Faction:    game.Eyrie,
					Leader:     leader,
					ClearingID: clearingID,
				},
			})
		}
	}
	return actions
}

func supportedVagabondCharacters() []game.VagabondCharacter {
	return []game.VagabondCharacter{
		game.CharThief,
		game.CharTinker,
		game.CharRanger,
	}
}

func validVagabondSetupActions(state game.GameState) []game.Action {
	characters := supportedVagabondCharacters()
	actions := make([]game.Action, 0, len(characters)*len(state.Map.Forests))
	for _, character := range characters {
		for _, forest := range state.Map.Forests {
			actions = append(actions, game.Action{
				Type: game.ActionVagabondSetup,
				VagabondSetup: &game.VagabondSetupAction{
					Faction:   game.Vagabond,
					Character: character,
					ForestID:  forest.ID,
				},
			})
		}
	}
	return actions
}

func applyMarquiseSetup(state *game.GameState, action game.Action) {
	if action.MarquiseSetup == nil {
		return
	}

	state.Marquise.KeepClearingID = action.MarquiseSetup.KeepClearingID
	placeToken(state, game.Marquise, action.MarquiseSetup.KeepClearingID, game.TokenKeep)

	opposite := oppositeCornerClearing(state.Map.ID, action.MarquiseSetup.KeepClearingID)
	for _, clearing := range state.Map.Clearings {
		if clearing.ID == opposite {
			continue
		}
		placeWarriors(state, game.Marquise, clearing.ID, 1)
		state.Marquise.WarriorSupply--
	}

	placeBuilding(state, game.Marquise, action.MarquiseSetup.SawmillClearingID, game.Sawmill)
	placeBuilding(state, game.Marquise, action.MarquiseSetup.WorkshopClearingID, game.Workshop)
	placeBuilding(state, game.Marquise, action.MarquiseSetup.RecruiterClearingID, game.Recruiter)
	state.Marquise.SawmillsPlaced = 1
	state.Marquise.WorkshopsPlaced = 1
	state.Marquise.RecruitersPlaced = 1
}

func applyEyrieSetup(state *game.GameState, action game.Action) {
	if action.EyrieSetup == nil {
		return
	}

	if !eyrieLeaderAvailable(state.Eyrie.AvailableLeaders, action.EyrieSetup.Leader) {
		return
	}

	state.Eyrie.Leader = action.EyrieSetup.Leader
	state.Eyrie.AvailableLeaders = removeLeader(state.Eyrie.AvailableLeaders, action.EyrieSetup.Leader)
	placeBuilding(state, game.Eyrie, action.EyrieSetup.ClearingID, game.Roost)
	placeWarriors(state, game.Eyrie, action.EyrieSetup.ClearingID, 6)
	state.Eyrie.WarriorSupply -= 6
	state.Eyrie.RoostsPlaced = 1

	vizierColumns := vizierColumnsForLeader(state.Eyrie.Leader)
	appendCardToDecree(&state.Eyrie.Decree, vizierColumns[0], game.LoyalVizier1)
	appendCardToDecree(&state.Eyrie.Decree, vizierColumns[1], game.LoyalVizier2)
}

func applyVagabondSetup(state *game.GameState, action game.Action) {
	if action.VagabondSetup == nil {
		return
	}

	state.Vagabond.Character = action.VagabondSetup.Character
	state.Vagabond.Items = VagabondStartingItems(action.VagabondSetup.Character)
	state.Vagabond.ClearingID = 0
	state.Vagabond.ForestID = action.VagabondSetup.ForestID
	state.Vagabond.InForest = true
}

func eyrieLeaderAvailable(leaders []game.EyrieLeader, leader game.EyrieLeader) bool {
	for _, candidate := range leaders {
		if candidate == leader {
			return true
		}
	}
	return false
}
