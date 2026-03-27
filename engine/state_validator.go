package engine

import "github.com/imdehydrated/rootbuddy/game"

func ValidateState(state game.GameState) error {
	if err := state.TurnWindow().Validate(); err != nil {
		return err
	}

	for _, count := range state.OtherHandCounts {
		if count < 0 {
			return errInvalidState("other hand counts cannot be negative")
		}
	}

	seenHiddenCardIDs := map[int]struct{}{}
	maxHiddenCardID := 0
	derivedHandCounts := map[game.Faction]int{}
	for _, hidden := range state.HiddenCards {
		if hidden.ID <= 0 {
			return errInvalidState("hidden card ids must be positive")
		}
		if _, exists := seenHiddenCardIDs[hidden.ID]; exists {
			return errInvalidState("hidden card ids must be unique")
		}
		seenHiddenCardIDs[hidden.ID] = struct{}{}
		if hidden.ID > maxHiddenCardID {
			maxHiddenCardID = hidden.ID
		}
		if hidden.Zone == game.HiddenCardZoneHand && hidden.OwnerFaction != state.PlayerFaction {
			derivedHandCounts[hidden.OwnerFaction]++
		}
	}
	if state.NextHiddenCardID > 0 && state.NextHiddenCardID <= maxHiddenCardID {
		return errInvalidState("next hidden card id must be greater than existing hidden card ids")
	}
	if state.GameMode == game.GameModeAssist && len(state.HiddenCards) > 0 {
		for faction, count := range derivedHandCounts {
			if state.OtherHandCounts[faction] != count {
				return errInvalidState("other hand counts must match hidden hand placeholders in assist mode")
			}
		}
		for faction, count := range state.OtherHandCounts {
			if derivedHandCounts[faction] != count {
				return errInvalidState("other hand counts must match hidden hand placeholders in assist mode")
			}
		}
	}

	for _, cardID := range state.Deck {
		if cardID < 0 {
			return errInvalidState("deck cannot contain negative card ids")
		}
	}

	for _, cardID := range state.DiscardPile {
		if cardID < 0 {
			return errInvalidState("discard pile cannot contain negative card ids")
		}
	}

	seenAvailableDominance := map[game.CardID]struct{}{}
	for _, cardID := range state.AvailableDominance {
		card, ok := CardByID(cardID)
		if !ok || card.Kind != game.DominanceCard {
			return errInvalidState("available dominance must contain only dominance cards")
		}
		if _, exists := seenAvailableDominance[cardID]; exists {
			return errInvalidState("available dominance cannot contain duplicates")
		}
		seenAvailableDominance[cardID] = struct{}{}
	}

	for _, cardID := range state.ActiveDominance {
		card, ok := CardByID(cardID)
		if !ok || card.Kind != game.DominanceCard {
			return errInvalidState("active dominance must contain only dominance cards")
		}
		if _, exists := seenAvailableDominance[cardID]; exists {
			return errInvalidState("dominance card cannot be both active and available")
		}
	}

	if state.CoalitionActive {
		if !hasActiveDominance(state, game.Vagabond) {
			return errInvalidState("coalition requires an active Vagabond dominance card")
		}
		if state.CoalitionPartner == game.Vagabond {
			return errInvalidState("Vagabond cannot form a coalition with itself")
		}
	}

	return nil
}

type invalidStateError string

func (msg invalidStateError) Error() string {
	return string(msg)
}

func errInvalidState(message string) error {
	return invalidStateError(message)
}
