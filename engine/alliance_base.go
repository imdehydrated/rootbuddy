package engine

import "github.com/imdehydrated/rootbuddy/game"

const allianceNoBaseSupporterLimit = 5

func cardMatchesBaseRemovalSuit(card game.Card, suit game.Suit) bool {
	return card.Suit == suit || card.Suit == game.Bird
}

func clearingHasAllianceBase(clearing game.Clearing) bool {
	for _, building := range clearing.Buildings {
		if building.Faction == game.Alliance && building.Type == game.Base {
			return true
		}
	}

	return false
}

func clearingHasOpenBuildSlotAfterRevolt(clearing game.Clearing) bool {
	usedSlots := 0
	for _, building := range clearing.Buildings {
		if building.Faction == game.Alliance {
			usedSlots++
		}
	}
	if clearing.Ruins {
		usedSlots++
	}
	return usedSlots < clearing.BuildSlots
}

func discardAllianceSupportersMatchingSuit(state *game.GameState, suit game.Suit) {
	remaining := make([]game.Card, 0, len(state.Alliance.Supporters))
	for _, supporter := range state.Alliance.Supporters {
		if cardMatchesBaseRemovalSuit(supporter, suit) {
			DiscardCard(state, supporter.ID)
			continue
		}
		remaining = append(remaining, supporter)
	}
	state.Alliance.Supporters = remaining

	hidden := make([]game.HiddenCard, 0, len(state.HiddenCards))
	for _, hiddenCard := range state.HiddenCards {
		if hiddenCard.OwnerFaction != game.Alliance ||
			hiddenCard.Zone != game.HiddenCardZoneSupporters ||
			hiddenCard.KnownCardID <= 0 {
			hidden = append(hidden, hiddenCard)
			continue
		}

		card, ok := CardByID(hiddenCard.KnownCardID)
		if !ok || !cardMatchesBaseRemovalSuit(card, suit) {
			hidden = append(hidden, hiddenCard)
			continue
		}

		DiscardCard(state, hiddenCard.KnownCardID)
	}
	state.HiddenCards = hidden
}

func removeHalfAllianceOfficersRoundedUp(state *game.GameState) {
	if state.Alliance.Officers <= 0 {
		return
	}

	state.Alliance.Officers -= (state.Alliance.Officers + 1) / 2
}

func enforceAllianceNoBaseSupporterLimit(state *game.GameState) {
	if allianceHasAnyBase(*state) {
		return
	}

	if len(state.Alliance.Supporters) > allianceNoBaseSupporterLimit {
		excess := state.Alliance.Supporters[allianceNoBaseSupporterLimit:]
		for _, supporter := range excess {
			DiscardCard(state, supporter.ID)
		}
		state.Alliance.Supporters = state.Alliance.Supporters[:allianceNoBaseSupporterLimit]
	}

	knownSupporterCount := len(state.Alliance.Supporters)
	hiddenLimit := allianceNoBaseSupporterLimit - knownSupporterCount
	if hiddenLimit < 0 {
		hiddenLimit = 0
	}

	keptHiddenSupporters := 0
	hidden := make([]game.HiddenCard, 0, len(state.HiddenCards))
	for _, hiddenCard := range state.HiddenCards {
		if hiddenCard.OwnerFaction != game.Alliance || hiddenCard.Zone != game.HiddenCardZoneSupporters {
			hidden = append(hidden, hiddenCard)
			continue
		}

		if keptHiddenSupporters < hiddenLimit {
			keptHiddenSupporters++
			hidden = append(hidden, hiddenCard)
			continue
		}

		DiscardCard(state, hiddenCard.KnownCardID)
	}
	state.HiddenCards = hidden
}

func removeAllianceBase(state *game.GameState, suit game.Suit) {
	discardAllianceSupportersMatchingSuit(state, suit)
	removeHalfAllianceOfficersRoundedUp(state)
	setAllianceBasePlaced(state, suit, false)
	enforceAllianceNoBaseSupporterLimit(state)
}
