package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidEyrieRecruitActions(state game.GameState, cardID game.CardID) []game.Action {
	recruitsPerRoost := 1
	if state.Eyrie.Leader == game.LeaderCharismatic {
		recruitsPerRoost = 2
	}
	if state.Eyrie.WarriorSupply <= 0 {
		return nil
	}

	clearingIDs := []int{}
	for _, clearing := range roostClearings(state) {
		if !decreeMatchesSuit(cardID, clearing.Suit) {
			continue
		}

		for i := 0; i < recruitsPerRoost; i++ {
			clearingIDs = append(clearingIDs, clearing.ID)
		}
	}
	if len(clearingIDs) == 0 {
		return nil
	}

	recruitCount := state.Eyrie.WarriorSupply
	if recruitCount >= len(clearingIDs) {
		return []game.Action{eyrieRecruitAction(cardID, clearingIDs)}
	}

	actions := []game.Action{}
	for _, chosenClearings := range recruitClearingSubsets(clearingIDs, recruitCount) {
		actions = append(actions, eyrieRecruitAction(cardID, chosenClearings))
	}
	return actions
}

func eyrieRecruitAction(cardID game.CardID, clearingIDs []int) game.Action {
	return game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:      game.Eyrie,
			ClearingIDs:  clearingIDs,
			DecreeCardID: cardID,
		},
	}
}
