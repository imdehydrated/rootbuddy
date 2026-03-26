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

	return nil
}

type invalidStateError string

func (msg invalidStateError) Error() string {
	return string(msg)
}

func errInvalidState(message string) error {
	return invalidStateError(message)
}
