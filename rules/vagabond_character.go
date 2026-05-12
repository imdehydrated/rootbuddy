package rules

import (
	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

func ValidVagabondCharacterActions(state game.GameState) []game.Action {
	if state.FactionTurn != game.Vagabond || state.CurrentPhase != game.Daylight {
		return nil
	}
	if state.CurrentStep != game.StepUnspecified && state.CurrentStep != game.StepDaylightActions && state.CurrentStep != game.StepDaylightCraft {
		return nil
	}
	if len(vagabondItemIndexes(state, game.ItemTorch, game.ItemReady)) == 0 {
		return nil
	}

	switch state.Vagabond.Character {
	case game.CharThief:
		return validVagabondStealActions(state)
	case game.CharTinker:
		return validVagabondDayLaborActions(state)
	case game.CharRanger:
		return validVagabondHideoutActions(state)
	default:
		return nil
	}
}

func validVagabondStealActions(state game.GameState) []game.Action {
	clearing, ok := vagabondCurrentClearing(state)
	if !ok {
		return nil
	}

	actions := []game.Action{}
	for _, targetFaction := range vagabondFactionsInClearing(clearing) {
		if targetFaction == game.Vagabond || vagabondFactionHandCount(state, targetFaction) == 0 {
			continue
		}
		actions = append(actions, game.Action{
			Type: game.ActionVagabondSteal,
			VagabondSteal: &game.VagabondStealAction{
				Faction:       game.Vagabond,
				ClearingID:    clearing.ID,
				TargetFaction: targetFaction,
			},
		})
	}

	return actions
}

func validVagabondDayLaborActions(state game.GameState) []game.Action {
	clearing, ok := vagabondCurrentClearing(state)
	if !ok {
		return nil
	}

	actions := []game.Action{}
	for _, cardID := range state.DiscardPile {
		card, ok := vagabondDiscardCardByID(cardID)
		if !ok || !matchesSuitOrBird(card, clearing.Suit) {
			continue
		}
		actions = append(actions, game.Action{
			Type: game.ActionVagabondDayLabor,
			VagabondDayLabor: &game.VagabondDayLaborAction{
				Faction:    game.Vagabond,
				ClearingID: clearing.ID,
				CardID:     cardID,
			},
		})
	}

	return actions
}

func vagabondDiscardCardByID(cardID game.CardID) (game.Card, bool) {
	for _, card := range carddata.BaseDeck() {
		if card.ID == cardID {
			return card, true
		}
	}

	return game.Card{}, false
}

func validVagabondHideoutActions(state game.GameState) []game.Action {
	damagedIndexes := vagabondItemIndexes(state, game.ItemTea, game.ItemDamaged)
	for _, itemType := range []game.ItemType{game.ItemCoin, game.ItemCrossbow, game.ItemHammer, game.ItemSword, game.ItemTorch, game.ItemBoots, game.ItemBag} {
		damagedIndexes = append(damagedIndexes, vagabondItemIndexes(state, itemType, game.ItemDamaged)...)
	}
	if len(damagedIndexes) < 3 {
		return nil
	}

	actions := []game.Action{}
	for _, itemIndexes := range chooseItemIndexSubsets(damagedIndexes, 3) {
		actions = append(actions, game.Action{
			Type: game.ActionVagabondHideout,
			VagabondHideout: &game.VagabondHideoutAction{
				Faction:     game.Vagabond,
				ItemIndexes: itemIndexes,
			},
		})
	}

	return actions
}

func vagabondFactionHandCount(state game.GameState, faction game.Faction) int {
	switch faction {
	case game.Marquise:
		if len(state.Marquise.CardsInHand) > 0 {
			return len(state.Marquise.CardsInHand)
		}
	case game.Alliance:
		if len(state.Alliance.CardsInHand) > 0 {
			return len(state.Alliance.CardsInHand)
		}
	case game.Eyrie:
		if len(state.Eyrie.CardsInHand) > 0 {
			return len(state.Eyrie.CardsInHand)
		}
	}

	return state.OtherHandCounts[faction]
}
