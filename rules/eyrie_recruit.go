package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidEyrieRecruitActions(state game.GameState, cardID game.CardID) []game.Action {
	actions := []game.Action{}

	recruitsPerAction := 1
	if state.Eyrie.Leader == game.LeaderCharismatic {
		recruitsPerAction = 2
	}
	if state.Eyrie.WarriorSupply < recruitsPerAction {
		recruitsPerAction = state.Eyrie.WarriorSupply
	}
	if recruitsPerAction <= 0 {
		return actions
	}

	for _, clearing := range roostClearings(state) {
		if !decreeMatchesSuit(cardID, clearing.Suit) {
			continue
		}

		clearingIDs := make([]int, recruitsPerAction)
		for i := range clearingIDs {
			clearingIDs[i] = clearing.ID
		}

		actions = append(actions, game.Action{
			Type: game.ActionRecruit,
			Recruit: &game.RecruitAction{
				Faction:      game.Eyrie,
				ClearingIDs:  clearingIDs,
				DecreeCardID: cardID,
			},
		})
	}

	return actions
}
