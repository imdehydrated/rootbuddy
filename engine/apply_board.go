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
			return
		}
		state.Vagabond.ClearingID = action.Movement.To
		state.Vagabond.ForestID = 0
		state.Vagabond.InForest = false
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

func returnWarriorsToSupply(state *game.GameState, faction game.Faction, count int) {
	if count <= 0 {
		return
	}

	switch faction {
	case game.Marquise:
		state.Marquise.WarriorSupply += count
	case game.Eyrie:
		state.Eyrie.WarriorSupply += count
	case game.Alliance:
		state.Alliance.WarriorSupply += count
	}
}

func returnRemovedWarriorsToSupply(state *game.GameState, clearing *game.Clearing, faction game.Faction, count int) {
	returnWarriorsToSupply(state, faction, count)
	if faction == game.Marquise && clearing != nil {
		queueFieldHospitals(state, *clearing, count)
	}
}

func removeWarriorLosses(state *game.GameState, clearing *game.Clearing, faction game.Faction, losses int) int {
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
	returnRemovedWarriorsToSupply(state, clearing, faction, removed)
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

func removeBuildingLosses(state *game.GameState, clearing *game.Clearing, faction game.Faction, losses int) int {
	if losses <= 0 || len(clearing.Buildings) == 0 {
		return 0
	}

	removed := 0
	remaining := make([]game.Building, 0, len(clearing.Buildings))
	for _, building := range clearing.Buildings {
		if losses > 0 && building.Faction == faction {
			removeBuildingFromTracks(state, clearing, building)
			losses--
			removed++
			continue
		}
		remaining = append(remaining, building)
	}
	clearing.Buildings = remaining
	return removed
}

func removeBuildingFromTracks(state *game.GameState, clearing *game.Clearing, building game.Building) {
	if building.Faction == game.Marquise {
		decrementPlacedBuildingCounter(state, building.Type)
	}
	if building.Faction == game.Eyrie && building.Type == game.Roost && state.Eyrie.RoostsPlaced > 0 {
		state.Eyrie.RoostsPlaced--
	}
	if building.Faction == game.Alliance && building.Type == game.Base {
		setAllianceBasePlaced(state, clearing.Suit, false)
	}
}

func removeSelectedBuildingLoss(state *game.GameState, clearing *game.Clearing, faction game.Faction, buildingType game.BuildingType) bool {
	for index, building := range clearing.Buildings {
		if building.Faction != faction || building.Type != buildingType {
			continue
		}

		removeBuildingFromTracks(state, clearing, building)
		clearing.Buildings = append(clearing.Buildings[:index], clearing.Buildings[index+1:]...)
		return true
	}

	return false
}

func removeSelectedTokenLoss(state *game.GameState, clearing *game.Clearing, faction game.Faction, tokenType game.TokenType) (bool, bool) {
	for index, token := range clearing.Tokens {
		if token.Faction != faction || token.Type != tokenType {
			continue
		}

		removedSympathy := false
		if token.Faction == game.Alliance && token.Type == game.TokenSympathy && state.Alliance.SympathyPlaced > 0 {
			state.Alliance.SympathyPlaced--
			removedSympathy = true
		}
		if token.Faction == game.Marquise && token.Type == game.TokenKeep {
			state.Marquise.KeepClearingID = 0
		}
		clearing.Tokens = append(clearing.Tokens[:index], clearing.Tokens[index+1:]...)
		return true, removedSympathy
	}

	return false, false
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

type battleRemovalSummary struct {
	warriors  int
	buildings int
	tokens    int
	sympathy  int
}

func applySelectedPieceLosses(state *game.GameState, clearing *game.Clearing, faction game.Faction, losses int, selected []game.BattlePieceLoss) (int, battleRemovalSummary) {
	summary := battleRemovalSummary{}
	if losses <= 0 || len(selected) == 0 {
		return losses, summary
	}

	for _, loss := range selected {
		if losses <= 0 {
			break
		}

		switch loss.Kind {
		case game.BattlePieceBuilding:
			if removeSelectedBuildingLoss(state, clearing, faction, loss.BuildingType) {
				losses--
				summary.buildings++
			}
		case game.BattlePieceToken:
			removed, removedSympathy := removeSelectedTokenLoss(state, clearing, faction, loss.TokenType)
			if removed {
				losses--
				summary.tokens++
				if removedSympathy {
					summary.sympathy++
				}
			}
		case game.BattlePieceWood:
			if faction == game.Marquise && clearing.Wood > 0 {
				clearing.Wood--
				losses--
				summary.tokens++
			}
		}
	}

	return losses, summary
}

func applyNonVagabondBattleLosses(state *game.GameState, clearing *game.Clearing, faction game.Faction, losses int, selected []game.BattlePieceLoss) battleRemovalSummary {
	summary := battleRemovalSummary{}
	if losses <= 0 {
		return summary
	}

	warriorsBefore := 0
	if clearing.Warriors != nil {
		warriorsBefore = clearing.Warriors[faction]
	}
	remainingLosses := removeWarriorLosses(state, clearing, faction, losses)
	if clearing.Warriors != nil {
		summary.warriors = warriorsBefore - clearing.Warriors[faction]
	}

	selectedSummary := battleRemovalSummary{}
	remainingLosses, selectedSummary = applySelectedPieceLosses(state, clearing, faction, remainingLosses, selected)
	summary.buildings += selectedSummary.buildings
	summary.tokens += selectedSummary.tokens
	summary.sympathy += selectedSummary.sympathy

	removedBuildings := removeBuildingLosses(state, clearing, faction, remainingLosses)
	summary.buildings += removedBuildings
	remainingLosses -= removedBuildings

	_, removedTokens, removedSympathy := removeTokenLosses(state, clearing, faction, remainingLosses)
	summary.tokens += removedTokens
	summary.sympathy += removedSympathy

	return summary
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
	if !game.AreEnemies(*state, action.BattleResolution.Faction, action.BattleResolution.TargetFaction) {
		return
	}

	index := findClearingIndex(state.Map, action.BattleResolution.ClearingID)
	if index == -1 {
		return
	}

	applyBattleCardSideEffects(state, action)

	clearing := &state.Map.Clearings[index]
	attackerSummary := battleRemovalSummary{}
	if action.BattleResolution.Faction == game.Vagabond {
		exhaustReadyItemsByType(state, game.ItemSword, 1)
		damageVagabondItems(state, action.BattleResolution.AttackerLosses)
	} else {
		attackerSummary = applyNonVagabondBattleLosses(
			state,
			clearing,
			action.BattleResolution.Faction,
			action.BattleResolution.AttackerLosses,
			action.BattleResolution.AttackerPieceLosses,
		)
		scoreBattleRemovals(state, action.BattleResolution.TargetFaction, attackerSummary.buildings, attackerSummary.tokens)
		if attackerSummary.sympathy > 0 && action.BattleResolution.TargetFaction != game.Alliance {
			transferOutrageCard(state, action.BattleResolution.TargetFaction, clearing.Suit)
		}
	}

	if action.BattleResolution.TargetFaction == game.Vagabond {
		exhaustReadyItemsByType(state, game.ItemSword, 1)
		damageVagabondItems(state, action.BattleResolution.DefenderLosses)
		if attackerSummary.warriors+attackerSummary.buildings+attackerSummary.tokens > 0 {
			setVagabondRelationship(state, action.BattleResolution.Faction, game.RelHostile)
		}
		return
	}

	defenderWasHostileToVagabond := game.VagabondHostileTo(*state, action.BattleResolution.TargetFaction)
	defenderSummary := applyNonVagabondBattleLosses(
		state,
		clearing,
		action.BattleResolution.TargetFaction,
		action.BattleResolution.DefenderLosses,
		action.BattleResolution.DefenderPieceLosses,
	)
	scoreBattleRemovals(state, action.BattleResolution.Faction, defenderSummary.buildings, defenderSummary.tokens)

	if defenderSummary.sympathy > 0 && action.BattleResolution.Faction != game.Alliance {
		transferOutrageCard(state, action.BattleResolution.Faction, clearing.Suit)
	}
	if action.BattleResolution.Faction == game.Vagabond && state.FactionTurn == game.Vagabond {
		removedPieces := defenderSummary.warriors + defenderSummary.buildings + defenderSummary.tokens
		infamyPieces := 0
		if defenderWasHostileToVagabond {
			infamyPieces = removedPieces
		} else if defenderSummary.warriors > 0 {
			infamyPieces = removedPieces - 1
		}
		addVictoryPoints(state, game.Vagabond, infamyPieces)
	}
	if action.BattleResolution.Faction == game.Vagabond && !defenderWasHostileToVagabond && defenderSummary.warriors > 0 {
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

	if _, ok := spendFactionHandCard(state, game.Marquise, action.Overwork.CardID); !ok {
		return
	}
	state.Map.Clearings[index].Wood++
	DiscardCard(state, action.Overwork.CardID)
}
