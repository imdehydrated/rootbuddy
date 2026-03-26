package engine

import "github.com/imdehydrated/rootbuddy/game"

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

		state.Map.Clearings[index].Warriors[action.Recruit.Faction]++
		switch action.Recruit.Faction {
		case game.Marquise:
			state.Marquise.WarriorSupply--
		case game.Eyrie:
			state.Eyrie.WarriorSupply--
		case game.Alliance:
			state.Alliance.WarriorSupply--
		}
	}

	if action.Recruit.Faction == game.Marquise {
		state.TurnProgress.RecruitUsed = true
	}
}

func applyMovement(state *game.GameState, action game.Action) {
	if action.Movement == nil {
		return
	}

	if action.Movement.Faction == game.Vagabond {
		if action.Movement.ToForestID != 0 {
			state.Vagabond.ClearingID = 0
			state.Vagabond.ForestID = action.Movement.ToForestID
			state.Vagabond.InForest = true
		} else {
			state.Vagabond.ClearingID = action.Movement.To
			state.Vagabond.ForestID = 0
			state.Vagabond.InForest = false
		}
		exhaustReadyItemsByType(state, game.ItemBoots, max(1, action.Movement.Count))

		toIndex := findClearingIndex(state.Map, action.Movement.To)
		if !state.Vagabond.InForest && toIndex != -1 && hasAllianceSympathy(state.Map.Clearings[toIndex]) {
			transferOutrageCard(state, action.Movement.Faction, state.Map.Clearings[toIndex].Suit)
		}
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

	moved := action.Movement.Count
	if moved <= 0 {
		moved = action.Movement.MaxCount
	}
	state.Map.Clearings[fromIndex].Warriors[action.Movement.Faction] -= moved

	if state.Map.Clearings[toIndex].Warriors == nil {
		state.Map.Clearings[toIndex].Warriors = map[game.Faction]int{}
	}
	state.Map.Clearings[toIndex].Warriors[action.Movement.Faction] += moved

	if action.Movement.Faction != game.Alliance && hasAllianceSympathy(state.Map.Clearings[toIndex]) {
		transferOutrageCard(state, action.Movement.Faction, state.Map.Clearings[toIndex].Suit)
	}
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
			if faction == game.Eyrie && building.Type == game.Roost && state.Eyrie.RoostsPlaced > 0 {
				state.Eyrie.RoostsPlaced--
			}
			if faction == game.Alliance && building.Type == game.Base {
				setAllianceBasePlaced(state, clearing.Suit, false)
			}
			losses--
			continue
		}
		remaining = append(remaining, building)
	}
	clearing.Buildings = remaining
}

func removeTokenLosses(state *game.GameState, clearing *game.Clearing, faction game.Faction, losses int) (int, int, int) {
	if losses <= 0 {
		return losses, 0, 0
	}

	removedTokens := 0
	removedSympathy := 0
	if len(clearing.Tokens) > 0 {
		remaining := make([]game.Token, 0, len(clearing.Tokens))
		for _, token := range clearing.Tokens {
			if losses > 0 && token.Faction == faction {
				if token.Faction == game.Alliance && token.Type == game.TokenSympathy && state.Alliance.SympathyPlaced > 0 {
					state.Alliance.SympathyPlaced--
					removedSympathy++
				}
				if token.Faction == game.Marquise && token.Type == game.TokenKeep {
					state.Marquise.KeepClearingID = 0
				}
				losses--
				removedTokens++
				continue
			}
			remaining = append(remaining, token)
		}
		clearing.Tokens = remaining
	}

	if losses > 0 && faction == game.Marquise && clearing.Wood > 0 {
		removedWood := losses
		if removedWood > clearing.Wood {
			removedWood = clearing.Wood
		}
		clearing.Wood -= removedWood
		losses -= removedWood
		removedTokens += removedWood
	}

	return losses, removedTokens, removedSympathy
}

func applyBattleCardSideEffects(state *game.GameState, action game.Action) {
	if action.BattleResolution == nil {
		return
	}

	suit := clearingSuit(*state, action.BattleResolution.ClearingID)
	if action.BattleResolution.DefenderAmbushed {
		consumeAmbushCard(state, action.BattleResolution.TargetFaction, suit)
		if action.BattleResolution.AttackerCounterAmbush {
			consumeAmbushCard(state, action.BattleResolution.Faction, suit)
		}
	}

	if action.BattleResolution.AttackerUsedArmorers {
		consumePersistentEffect(state, action.BattleResolution.Faction, "armorers")
	}
	if action.BattleResolution.DefenderUsedArmorers {
		consumePersistentEffect(state, action.BattleResolution.TargetFaction, "armorers")
	}
	if action.BattleResolution.AttackerUsedBrutalTactics {
		addVictoryPoints(state, action.BattleResolution.TargetFaction, 1)
	}
}

func applyBattleResolution(state *game.GameState, action game.Action) {
	if action.BattleResolution == nil {
		return
	}

	index := findClearingIndex(state.Map, action.BattleResolution.ClearingID)
	if index == -1 {
		return
	}

	applyBattleCardSideEffects(state, action)

	clearing := &state.Map.Clearings[index]
	if action.BattleResolution.Faction == game.Vagabond {
		exhaustReadyItemsByType(state, game.ItemSword, 1)
		damageVagabondItems(state, action.BattleResolution.AttackerLosses)
	} else {
		removeWarriorLosses(clearing, action.BattleResolution.Faction, action.BattleResolution.AttackerLosses)
	}

	if action.BattleResolution.TargetFaction == game.Vagabond {
		exhaustReadyItemsByType(state, game.ItemSword, 1)
		damageVagabondItems(state, action.BattleResolution.DefenderLosses)
		return
	}

	targetWarriorsBefore := 0
	if clearing.Warriors != nil {
		targetWarriorsBefore = clearing.Warriors[action.BattleResolution.TargetFaction]
	}
	remainingDefenderLosses := removeWarriorLosses(clearing, action.BattleResolution.TargetFaction, action.BattleResolution.DefenderLosses)
	beforeBuildings := len(clearing.Buildings)
	removeBuildingLosses(state, clearing, action.BattleResolution.TargetFaction, remainingDefenderLosses)
	removedBuildings := beforeBuildings - len(clearing.Buildings)
	remainingDefenderLosses -= removedBuildings
	_, removedTokens, removedSympathy := removeTokenLosses(state, clearing, action.BattleResolution.TargetFaction, remainingDefenderLosses)
	removedWarriors := targetWarriorsBefore
	if clearing.Warriors != nil {
		removedWarriors -= clearing.Warriors[action.BattleResolution.TargetFaction]
	}
	scoreBattleRemovals(state, action.BattleResolution.Faction, removedBuildings, removedTokens)

	if removedSympathy > 0 && action.BattleResolution.Faction != game.Alliance {
		transferOutrageCard(state, action.BattleResolution.Faction, clearing.Suit)
	}
	if action.BattleResolution.Faction == game.Vagabond && removedWarriors+removedBuildings+removedTokens > 0 {
		setVagabondRelationship(state, action.BattleResolution.TargetFaction, game.RelHostile)
	}
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
			Faction: action.Build.Faction,
			Type:    action.Build.BuildingType,
		},
	)

	switch action.Build.BuildingType {
	case game.Sawmill:
		scoreMarquiseBuilding(state, action.Build.BuildingType, state.Marquise.SawmillsPlaced)
		state.Marquise.SawmillsPlaced++
	case game.Workshop:
		scoreMarquiseBuilding(state, action.Build.BuildingType, state.Marquise.WorkshopsPlaced)
		state.Marquise.WorkshopsPlaced++
	case game.Recruiter:
		scoreMarquiseBuilding(state, action.Build.BuildingType, state.Marquise.RecruitersPlaced)
		state.Marquise.RecruitersPlaced++
	case game.Roost:
		state.Eyrie.RoostsPlaced++
	case game.Base:
		setAllianceBasePlaced(state, state.Map.Clearings[index].Suit, true)
	}

	for _, source := range action.Build.WoodSources {
		sourceIndex := findClearingIndex(state.Map, source.ClearingID)
		if sourceIndex == -1 {
			continue
		}

		wood := state.Map.Clearings[sourceIndex].Wood
		if source.Amount >= wood {
			state.Map.Clearings[sourceIndex].Wood = 0
			continue
		}

		state.Map.Clearings[sourceIndex].Wood -= source.Amount
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

	if _, ok := removeCardFromFactionHand(state, game.Marquise, action.Overwork.CardID); !ok {
		return
	}
	state.Map.Clearings[index].Wood++
	DiscardCard(state, action.Overwork.CardID)
}
