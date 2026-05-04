package engine

import "github.com/imdehydrated/rootbuddy/game"

func applySpreadSympathy(state *game.GameState, action game.Action) {
	if action.SpreadSympathy == nil {
		return
	}

	index := findClearingIndex(state.Map, action.SpreadSympathy.ClearingID)
	if index == -1 {
		return
	}

	spendAllianceSupporters(state, action.SpreadSympathy.SupporterCardIDs)
	DiscardCards(state, action.SpreadSympathy.SupporterCardIDs)
	state.Map.Clearings[index].Tokens = append(state.Map.Clearings[index].Tokens, game.Token{
		Faction: game.Alliance,
		Type:    game.TokenSympathy,
	})
	scoreAllianceSympathy(state, state.Alliance.SympathyPlaced)
	state.Alliance.SympathyPlaced++
}

func removeEnemyPiecesForRevolt(state *game.GameState, clearing *game.Clearing) int {
	removedPieces := 0

	for faction, warriors := range clearing.Warriors {
		if faction == game.Alliance || warriors <= 0 {
			continue
		}
		removedPieces += warriors
		clearing.Warriors[faction] = 0
	}

	if len(clearing.Buildings) > 0 {
		remaining := make([]game.Building, 0, len(clearing.Buildings))
		for _, building := range clearing.Buildings {
			if building.Faction == game.Alliance {
				remaining = append(remaining, building)
				continue
			}

			if building.Faction == game.Marquise {
				decrementPlacedBuildingCounter(state, building.Type)
			}
			if building.Faction == game.Eyrie && building.Type == game.Roost && state.Eyrie.RoostsPlaced > 0 {
				state.Eyrie.RoostsPlaced--
			}
			if building.Faction == game.Alliance && building.Type == game.Base {
				setAllianceBasePlaced(state, clearing.Suit, false)
			}
			removedPieces++
		}
		clearing.Buildings = remaining
	}

	if len(clearing.Tokens) > 0 {
		remaining := make([]game.Token, 0, len(clearing.Tokens))
		for _, token := range clearing.Tokens {
			if token.Faction == game.Alliance {
				remaining = append(remaining, token)
				continue
			}

			if token.Faction == game.Marquise && token.Type == game.TokenKeep {
				state.Marquise.KeepClearingID = 0
			}
			removedPieces++
		}
		clearing.Tokens = remaining
	}

	if clearing.Wood > 0 {
		removedPieces += clearing.Wood
		clearing.Wood = 0
	}

	return removedPieces
}

func sympathyCountBySuit(board game.Map, suit game.Suit) int {
	count := 0
	for _, clearing := range board.Clearings {
		if clearing.Suit != suit || !hasAllianceSympathy(clearing) {
			continue
		}
		count++
	}

	return count
}

func applyRevolt(state *game.GameState, action game.Action) {
	if action.Revolt == nil {
		return
	}

	index := findClearingIndex(state.Map, action.Revolt.ClearingID)
	if index == -1 {
		return
	}

	spendAllianceSupporters(state, action.Revolt.SupporterCardIDs)
	DiscardCards(state, action.Revolt.SupporterCardIDs)
	clearing := &state.Map.Clearings[index]
	removedPieces := removeEnemyPiecesForRevolt(state, clearing)
	clearing.Buildings = append(clearing.Buildings, game.Building{
		Faction: game.Alliance,
		Type:    game.Base,
	})
	setAllianceBasePlaced(state, action.Revolt.BaseSuit, true)

	if clearing.Warriors == nil {
		clearing.Warriors = map[game.Faction]int{}
	}

	recruitCount := sympathyCountBySuit(state.Map, action.Revolt.BaseSuit)
	if recruitCount > state.Alliance.WarriorSupply {
		recruitCount = state.Alliance.WarriorSupply
	}
	clearing.Warriors[game.Alliance] += recruitCount
	state.Alliance.WarriorSupply -= recruitCount
	state.Alliance.Officers++
	addVictoryPoints(state, game.Alliance, removedPieces)
}

func applyMobilize(state *game.GameState, action game.Action) {
	if action.Mobilize == nil {
		return
	}
	if canUseObservedHiddenCards(*state, game.Alliance) {
		if !moveHiddenCard(state, game.Alliance, game.HiddenCardZoneHand, game.HiddenCardZoneSupporters) {
			return
		}
		return
	}

	for _, card := range state.Alliance.CardsInHand {
		if card.ID != action.Mobilize.CardID {
			continue
		}

		if _, ok := spendFactionHandCard(state, game.Alliance, card.ID); !ok {
			return
		}
		gainAllianceSupporter(state, card)
		return
	}
}

func applyTrain(state *game.GameState, action game.Action) {
	if action.Train == nil {
		return
	}

	if _, ok := spendFactionHandCard(state, game.Alliance, action.Train.CardID); !ok {
		return
	}
	DiscardCard(state, action.Train.CardID)
	state.Alliance.Officers++
}

func applyOrganize(state *game.GameState, action game.Action) {
	if action.Organize == nil {
		return
	}

	index := findClearingIndex(state.Map, action.Organize.ClearingID)
	if index == -1 {
		return
	}

	clearing := &state.Map.Clearings[index]
	if clearing.Warriors == nil || clearing.Warriors[game.Alliance] <= 0 {
		return
	}

	clearing.Warriors[game.Alliance]--
	clearing.Tokens = append(clearing.Tokens, game.Token{
		Faction: game.Alliance,
		Type:    game.TokenSympathy,
	})
	scoreAllianceSympathy(state, state.Alliance.SympathyPlaced)
	state.Alliance.SympathyPlaced++
}
