package engine

import (
	"sort"

	"github.com/imdehydrated/rootbuddy/game"
	"github.com/imdehydrated/rootbuddy/rules"
)

func validDominanceActions(state game.GameState) []game.Action {
	if state.CurrentPhase != game.Daylight || state.GamePhase != game.LifecyclePlaying {
		return nil
	}

	faction := state.FactionTurn
	actions := []game.Action{}
	hand := factionHand(state, faction)
	points := state.VictoryPoints[faction]

	if points >= 10 && !hasActiveDominance(state, faction) {
		for _, card := range hand {
			if card.Kind != game.DominanceCard {
				continue
			}

			if faction == game.Vagabond {
				if len(effectiveTurnOrder(state)) < 4 || state.CoalitionActive {
					continue
				}
				for _, targetFaction := range coalitionTargets(state) {
					actions = append(actions, game.Action{
						Type: game.ActionActivateDominance,
						ActivateDominance: &game.ActivateDominanceAction{
							Faction:       faction,
							CardID:        card.ID,
							TargetFaction: targetFaction,
						},
					})
				}
				continue
			}

			actions = append(actions, game.Action{
				Type: game.ActionActivateDominance,
				ActivateDominance: &game.ActivateDominanceAction{
					Faction: faction,
					CardID:  card.ID,
				},
			})
		}
	}

	for _, dominanceCardID := range state.AvailableDominance {
		dominanceCard, ok := CardByID(dominanceCardID)
		if !ok {
			continue
		}
		for _, spendCard := range hand {
			if !cardCanTakeDominance(spendCard, dominanceCard) {
				continue
			}
			actions = append(actions, game.Action{
				Type: game.ActionTakeDominance,
				TakeDominance: &game.TakeDominanceAction{
					Faction:         faction,
					DominanceCardID: dominanceCardID,
					SpentCardID:     spendCard.ID,
				},
			})
		}
	}

	return actions
}

func cardCanTakeDominance(spendCard game.Card, dominanceCard game.Card) bool {
	if spendCard.Suit != dominanceCard.Suit {
		return false
	}
	if dominanceCard.Suit == game.Bird {
		return spendCard.Suit == game.Bird
	}
	return true
}

func coalitionTargets(state game.GameState) []game.Faction {
	order := effectiveTurnOrder(state)
	if len(order) < 4 {
		return nil
	}

	lowest := 0
	targets := []game.Faction{}
	found := false
	for _, faction := range order {
		if faction == game.Vagabond {
			continue
		}
		points := state.VictoryPoints[faction]
		if !found || points < lowest {
			lowest = points
			targets = []game.Faction{faction}
			found = true
		} else if points == lowest {
			targets = append(targets, faction)
		}
	}

	return targets
}

func addAvailableDominance(state *game.GameState, cardID game.CardID) {
	if cardID <= 0 {
		return
	}
	for _, existing := range state.AvailableDominance {
		if existing == cardID {
			return
		}
	}
	state.AvailableDominance = append(state.AvailableDominance, cardID)
	sort.Slice(state.AvailableDominance, func(i, j int) bool {
		return state.AvailableDominance[i] < state.AvailableDominance[j]
	})
}

func removeAvailableDominance(state *game.GameState, cardID game.CardID) bool {
	for index, existing := range state.AvailableDominance {
		if existing != cardID {
			continue
		}
		state.AvailableDominance = append(state.AvailableDominance[:index], state.AvailableDominance[index+1:]...)
		return true
	}
	return false
}

func hasActiveDominance(state game.GameState, faction game.Faction) bool {
	if state.ActiveDominance == nil {
		return false
	}
	_, ok := state.ActiveDominance[faction]
	return ok
}

func birdDominanceCornerPairs(state game.GameState) [][2]int {
	switch state.Map.ID {
	case game.AutumnMapID:
		return [][2]int{{1, oppositeCornerClearing(state.Map.ID, 1)}, {2, oppositeCornerClearing(state.Map.ID, 2)}}
	default:
		return nil
	}
}

func rulesClearing(state game.GameState, clearingID int, faction game.Faction) bool {
	index := findClearingIndex(state.Map, clearingID)
	if index == -1 {
		return false
	}
	ruler, ruled := rules.Ruler(state.Map.Clearings[index])
	return ruled && ruler == faction
}

func winsByDominance(state game.GameState, faction game.Faction) bool {
	cardID, ok := state.ActiveDominance[faction]
	if !ok {
		return false
	}
	card, found := CardByID(cardID)
	if !found {
		return false
	}
	if faction == game.Vagabond {
		return false
	}

	if card.Suit == game.Bird {
		for _, pair := range birdDominanceCornerPairs(state) {
			if rulesClearing(state, pair[0], faction) && rulesClearing(state, pair[1], faction) {
				return true
			}
		}
		return false
	}

	count := 0
	for _, clearing := range state.Map.Clearings {
		if clearing.Suit != card.Suit {
			continue
		}
		ruler, ruled := rules.Ruler(clearing)
		if ruled && ruler == faction {
			count++
		}
	}

	return count >= 3
}

func checkBirdsongDominanceWin(state *game.GameState) {
	if state.GamePhase != game.LifecyclePlaying || state.CurrentPhase != game.Birdsong {
		return
	}
	if !winsByDominance(*state, state.FactionTurn) {
		return
	}

	setWinner(state, state.FactionTurn)
}

func setWinner(state *game.GameState, faction game.Faction) {
	state.Winner = faction
	if state.CoalitionActive && faction == state.CoalitionPartner {
		state.WinningCoalition = []game.Faction{faction, game.Vagabond}
	} else {
		state.WinningCoalition = nil
	}
	state.GamePhase = game.LifecycleGameOver
}
