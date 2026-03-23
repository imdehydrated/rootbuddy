package engine

import (
	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

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
	case game.ActionAddToDecree:
		applyAddToDecree(&next, action)
	case game.ActionTurmoil:
		applyTurmoil(&next, action)
	case game.ActionBirdsongWood:
		applyBirdsongWood(&next, action)
	case game.ActionEveningDraw:
		applyEveningDraw(&next, action)
	case game.ActionScoreRoosts:
		applyScoreRoosts(&next, action)
	case game.ActionPassPhase:
		applyPassPhase(&next, action)
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

		if clearing.Tokens != nil {
			cloned.Tokens = make([]game.Token, len(clearing.Tokens))
			copy(cloned.Tokens, clearing.Tokens)
		}

		next.Map.Clearings[i] = cloned
	}

	if state.TurnOrder != nil {
		next.TurnOrder = make([]game.Faction, len(state.TurnOrder))
		copy(next.TurnOrder, state.TurnOrder)
	}

	if state.VictoryPoints != nil {
		next.VictoryPoints = make(map[game.Faction]int, len(state.VictoryPoints))
		for faction, points := range state.VictoryPoints {
			next.VictoryPoints[faction] = points
		}
	}

	if state.Marquise.CardsInHand != nil {
		next.Marquise.CardsInHand = make([]game.Card, len(state.Marquise.CardsInHand))
		copy(next.Marquise.CardsInHand, state.Marquise.CardsInHand)
	}

	if state.Eyrie.CardsInHand != nil {
		next.Eyrie.CardsInHand = make([]game.Card, len(state.Eyrie.CardsInHand))
		copy(next.Eyrie.CardsInHand, state.Eyrie.CardsInHand)
	}

	if state.Eyrie.AvailableLeaders != nil {
		next.Eyrie.AvailableLeaders = make([]game.EyrieLeader, len(state.Eyrie.AvailableLeaders))
		copy(next.Eyrie.AvailableLeaders, state.Eyrie.AvailableLeaders)
	}

	next.Eyrie.Decree.Recruit = cloneCardIDs(state.Eyrie.Decree.Recruit)
	next.Eyrie.Decree.Move = cloneCardIDs(state.Eyrie.Decree.Move)
	next.Eyrie.Decree.Battle = cloneCardIDs(state.Eyrie.Decree.Battle)
	next.Eyrie.Decree.Build = cloneCardIDs(state.Eyrie.Decree.Build)

	if state.Alliance.CardsInHand != nil {
		next.Alliance.CardsInHand = make([]game.Card, len(state.Alliance.CardsInHand))
		copy(next.Alliance.CardsInHand, state.Alliance.CardsInHand)
	}

	if state.Alliance.Supporters != nil {
		next.Alliance.Supporters = make([]game.Card, len(state.Alliance.Supporters))
		copy(next.Alliance.Supporters, state.Alliance.Supporters)
	}

	if state.Vagabond.CardsInHand != nil {
		next.Vagabond.CardsInHand = make([]game.Card, len(state.Vagabond.CardsInHand))
		copy(next.Vagabond.CardsInHand, state.Vagabond.CardsInHand)
	}

	if state.Vagabond.Items != nil {
		next.Vagabond.Items = make([]game.Item, len(state.Vagabond.Items))
		copy(next.Vagabond.Items, state.Vagabond.Items)
	}

	if state.Vagabond.Relationships != nil {
		next.Vagabond.Relationships = make(map[game.Faction]game.RelationshipLevel, len(state.Vagabond.Relationships))
		for faction, relationship := range state.Vagabond.Relationships {
			next.Vagabond.Relationships[faction] = relationship
		}
	}

	if state.Vagabond.QuestsCompleted != nil {
		next.Vagabond.QuestsCompleted = make([]game.Card, len(state.Vagabond.QuestsCompleted))
		copy(next.Vagabond.QuestsCompleted, state.Vagabond.QuestsCompleted)
	}

	if state.Vagabond.QuestsAvailable != nil {
		next.Vagabond.QuestsAvailable = make([]game.Card, len(state.Vagabond.QuestsAvailable))
		copy(next.Vagabond.QuestsAvailable, state.Vagabond.QuestsAvailable)
	}

	if state.TurnProgress.UsedWorkshopClearings != nil {
		next.TurnProgress.UsedWorkshopClearings = make([]int, len(state.TurnProgress.UsedWorkshopClearings))
		copy(next.TurnProgress.UsedWorkshopClearings, state.TurnProgress.UsedWorkshopClearings)
	}

	if state.TurnProgress.ResolvedDecreeCardIDs != nil {
		next.TurnProgress.ResolvedDecreeCardIDs = make([]game.CardID, len(state.TurnProgress.ResolvedDecreeCardIDs))
		copy(next.TurnProgress.ResolvedDecreeCardIDs, state.TurnProgress.ResolvedDecreeCardIDs)
	}

	return next
}

func cloneCardIDs(cardIDs []game.CardID) []game.CardID {
	if cardIDs == nil {
		return nil
	}

	cloned := make([]game.CardID, len(cardIDs))
	copy(cloned, cardIDs)
	return cloned
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

		state.Map.Clearings[index].Warriors[action.Recruit.Faction]++
		switch action.Recruit.Faction {
		case game.Marquise:
			state.Marquise.WarriorSupply--
		case game.Eyrie:
			state.Eyrie.WarriorSupply--
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
	beforeBuildings := len(clearing.Buildings)
	removeBuildingLosses(state, clearing, action.BattleResolution.TargetFaction, remainingDefenderLosses)
	removedBuildings := beforeBuildings - len(clearing.Buildings)
	scoreBattleRemovals(state, action.BattleResolution.Faction, removedBuildings, 0)
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

	state.Map.Clearings[index].Wood++
	state.Marquise.CardsInHand = removeCardByID(state.Marquise.CardsInHand, action.Overwork.CardID)
}

func applyCraft(state *game.GameState, action game.Action) {
	if action.Craft == nil {
		return
	}

	switch action.Craft.Faction {
	case game.Marquise:
		state.Marquise.CardsInHand = removeCardByID(state.Marquise.CardsInHand, action.Craft.CardID)
	case game.Eyrie:
		state.Eyrie.CardsInHand = removeCardByID(state.Eyrie.CardsInHand, action.Craft.CardID)
	}
	state.TurnProgress.UsedWorkshopClearings = append(
		state.TurnProgress.UsedWorkshopClearings,
		action.Craft.UsedWorkshopClearings...,
	)
}

func appendCardToDecree(decree *game.Decree, column game.DecreeColumn, cardID game.CardID) {
	switch column {
	case game.DecreeRecruit:
		decree.Recruit = append(decree.Recruit, cardID)
	case game.DecreeMove:
		decree.Move = append(decree.Move, cardID)
	case game.DecreeBattle:
		decree.Battle = append(decree.Battle, cardID)
	case game.DecreeBuild:
		decree.Build = append(decree.Build, cardID)
	}
}

func applyAddToDecree(state *game.GameState, action game.Action) {
	if action.AddToDecree == nil {
		return
	}

	state.TurnProgress.ResolvedDecreeCardIDs = nil
	state.TurnProgress.DecreeColumnsResolved = 0
	state.TurnProgress.DecreeCardsResolved = 0

	for i, cardID := range action.AddToDecree.CardIDs {
		if i >= len(action.AddToDecree.Columns) {
			break
		}

		appendCardToDecree(&state.Eyrie.Decree, action.AddToDecree.Columns[i], cardID)
		state.Eyrie.CardsInHand = removeCardByID(state.Eyrie.CardsInHand, cardID)
	}
}

func removeLeader(leaders []game.EyrieLeader, remove game.EyrieLeader) []game.EyrieLeader {
	filtered := make([]game.EyrieLeader, 0, len(leaders))
	for _, leader := range leaders {
		if leader != remove {
			filtered = append(filtered, leader)
		}
	}
	return filtered
}

func eyrieCardSuit(id game.CardID) game.Suit {
	if id == game.LoyalVizier1 || id == game.LoyalVizier2 {
		return game.Bird
	}

	for _, card := range carddata.BaseDeck() {
		if card.ID == id {
			return card.Suit
		}
	}

	return game.Bird
}

func birdCardsInDecree(decree game.Decree) int {
	count := 0
	for _, column := range [][]game.CardID{decree.Recruit, decree.Move, decree.Battle, decree.Build} {
		for _, cardID := range column {
			if eyrieCardSuit(cardID) == game.Bird {
				count++
			}
		}
	}
	return count
}

func vizierColumnsForLeader(leader game.EyrieLeader) [2]game.DecreeColumn {
	switch leader {
	case game.LeaderBuilder:
		return [2]game.DecreeColumn{game.DecreeRecruit, game.DecreeMove}
	case game.LeaderCharismatic:
		return [2]game.DecreeColumn{game.DecreeRecruit, game.DecreeBattle}
	case game.LeaderCommander:
		return [2]game.DecreeColumn{game.DecreeMove, game.DecreeBattle}
	default:
		return [2]game.DecreeColumn{game.DecreeMove, game.DecreeBuild}
	}
}

func applyTurmoil(state *game.GameState, action game.Action) {
	if action.Turmoil == nil {
		return
	}

	if state.VictoryPoints == nil {
		state.VictoryPoints = map[game.Faction]int{}
	}
	state.VictoryPoints[game.Eyrie] -= birdCardsInDecree(state.Eyrie.Decree)

	state.Eyrie.AvailableLeaders = removeLeader(state.Eyrie.AvailableLeaders, state.Eyrie.Leader)
	state.Eyrie.AvailableLeaders = removeLeader(state.Eyrie.AvailableLeaders, action.Turmoil.NewLeader)
	state.Eyrie.Leader = action.Turmoil.NewLeader
	state.Eyrie.Decree = game.Decree{}
	state.TurnProgress.ResolvedDecreeCardIDs = nil

	vizierColumns := vizierColumnsForLeader(action.Turmoil.NewLeader)
	appendCardToDecree(&state.Eyrie.Decree, vizierColumns[0], game.LoyalVizier1)
	appendCardToDecree(&state.Eyrie.Decree, vizierColumns[1], game.LoyalVizier2)
}

func applyBirdsongWood(state *game.GameState, action game.Action) {
	if action.BirdsongWood == nil {
		return
	}

	for _, clearingID := range action.BirdsongWood.ClearingIDs {
		index := findClearingIndex(state.Map, clearingID)
		if index == -1 {
			continue
		}

		state.Map.Clearings[index].Wood += action.BirdsongWood.Amount
	}
}

func applyEveningDraw(state *game.GameState, action game.Action) {
	if action.EveningDraw == nil {
		return
	}
}

func applyScoreRoosts(state *game.GameState, action game.Action) {
	if action.ScoreRoosts == nil {
		return
	}

	addVictoryPoints(state, action.ScoreRoosts.Faction, action.ScoreRoosts.Points)
}

func applyPassPhase(state *game.GameState, action game.Action) {
	if action.PassPhase == nil {
		return
	}
}

func advanceTurnState(state *game.GameState, action game.Action) {
	switch action.Type {
	case game.ActionAddToDecree:
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.StepDaylightActions
	case game.ActionBirdsongWood:
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.StepDaylightActions
	case game.ActionRecruit:
		state.CurrentStep = game.StepDaylightActions
		if action.Recruit != nil && action.Recruit.Faction == game.Marquise {
			state.TurnProgress.ActionsUsed++
		}
		if action.Recruit != nil && action.Recruit.Faction == game.Eyrie {
			markResolvedDecreeCard(state, action.Recruit.DecreeCardID)
		}
	case game.ActionMovement:
		state.CurrentStep = game.StepDaylightActions
		if action.Movement != nil && action.Movement.Faction == game.Marquise {
			state.TurnProgress.ActionsUsed++
			state.TurnProgress.MarchesUsed++
		}
		if action.Movement != nil && action.Movement.Faction == game.Eyrie {
			markResolvedDecreeCard(state, action.Movement.DecreeCardID)
		}
	case game.ActionBattleResolution, game.ActionBuild, game.ActionOverwork:
		state.CurrentStep = game.StepDaylightActions
		switch {
		case action.Type == game.ActionBattleResolution && action.BattleResolution != nil && action.BattleResolution.Faction == game.Marquise:
			state.TurnProgress.ActionsUsed++
		case action.Type == game.ActionBattleResolution && action.BattleResolution != nil && action.BattleResolution.Faction == game.Eyrie:
			markResolvedDecreeCard(state, action.BattleResolution.DecreeCardID)
		case action.Type == game.ActionBuild && action.Build != nil && action.Build.Faction == game.Marquise:
			state.TurnProgress.ActionsUsed++
		case action.Type == game.ActionBuild && action.Build != nil && action.Build.Faction == game.Eyrie:
			markResolvedDecreeCard(state, action.Build.DecreeCardID)
		case action.Type == game.ActionOverwork:
			state.TurnProgress.ActionsUsed++
		}
	case game.ActionCraft:
		state.CurrentStep = game.StepDaylightActions
		state.TurnProgress.HasCrafted = true
	case game.ActionTurmoil:
		state.CurrentPhase = game.Evening
		state.CurrentStep = game.StepEvening
		state.TurnProgress.DecreeColumnsResolved = 0
		state.TurnProgress.DecreeCardsResolved = 0
	case game.ActionPassPhase:
		switch state.CurrentPhase {
		case game.Birdsong:
			state.CurrentPhase = game.Daylight
			state.CurrentStep = game.StepDaylightActions
		case game.Daylight:
			if state.CurrentStep == game.StepDaylightCraft {
				state.CurrentStep = game.StepDaylightActions
			} else {
				state.CurrentPhase = game.Evening
				state.CurrentStep = game.StepEvening
			}
		}
	}
}

func decreeCardsByColumn(decree game.Decree, column game.DecreeColumn) []game.CardID {
	switch column {
	case game.DecreeRecruit:
		return decree.Recruit
	case game.DecreeMove:
		return decree.Move
	case game.DecreeBattle:
		return decree.Battle
	case game.DecreeBuild:
		return decree.Build
	default:
		return nil
	}
}

func currentDecreeCard(state game.GameState) (game.DecreeColumn, game.CardID, bool) {
	for i := state.TurnProgress.DecreeColumnsResolved; i < 4; i++ {
		column := game.DecreeColumn(i)
		for _, cardID := range decreeCardsByColumn(state.Eyrie.Decree, column) {
			if !decreeCardResolved(state, cardID) {
				return column, cardID, true
			}
		}
	}

	return 0, 0, false
}

func decreeCardResolved(state game.GameState, cardID game.CardID) bool {
	for _, resolvedID := range state.TurnProgress.ResolvedDecreeCardIDs {
		if resolvedID == cardID {
			return true
		}
	}
	return false
}

func markResolvedDecreeCard(state *game.GameState, cardID game.CardID) {
	if cardID == 0 || decreeCardResolved(*state, cardID) {
		return
	}

	state.TurnProgress.ResolvedDecreeCardIDs = append(state.TurnProgress.ResolvedDecreeCardIDs, cardID)
	state.TurnProgress.DecreeCardsResolved = 0

	for state.TurnProgress.DecreeColumnsResolved < 4 {
		column := game.DecreeColumn(state.TurnProgress.DecreeColumnsResolved)
		allResolved := true
		for _, columnCardID := range decreeCardsByColumn(state.Eyrie.Decree, column) {
			if !decreeCardResolved(*state, columnCardID) {
				allResolved = false
				break
			}
		}
		if !allResolved {
			return
		}
		state.TurnProgress.DecreeColumnsResolved++
	}
}
