package rules

import "github.com/imdehydrated/rootbuddy/game"

func ValidQuestActions(state game.GameState) []game.Action {
	clearing, ok := vagabondCurrentClearing(state)
	if !ok {
		return nil
	}

	actions := []game.Action{}
	for _, quest := range state.Vagabond.QuestsAvailable {
		if quest.Suit != clearing.Suit && quest.Suit != game.Bird {
			continue
		}

		itemChoices := readyItemIndexChoicesForTypes(state, quest.RequiredItems)
		for _, itemIndexes := range itemChoices {
			actions = append(actions,
				game.Action{
					Type: game.ActionQuest,
					Quest: &game.QuestAction{
						Faction:     game.Vagabond,
						QuestID:     quest.ID,
						ItemIndexes: itemIndexes,
						Reward:      game.QuestRewardVictoryPoints,
					},
				},
				game.Action{
					Type: game.ActionQuest,
					Quest: &game.QuestAction{
						Faction:     game.Vagabond,
						QuestID:     quest.ID,
						ItemIndexes: itemIndexes,
						Reward:      game.QuestRewardDrawCards,
					},
				},
			)
		}

	}

	return actions
}
