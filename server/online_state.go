package server

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

type onlineStateStore struct {
	mu    sync.RWMutex
	games map[string]game.GameState
}

var store = &onlineStateStore{
	games: map[string]game.GameState{},
}

func newGameID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}

	return hex.EncodeToString(bytes)
}

func (s *onlineStateStore) save(gameID string, state game.GameState) {
	if gameID == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.games[gameID] = engine.CloneState(state)
}

func (s *onlineStateStore) load(gameID string) (game.GameState, bool) {
	if gameID == "" {
		return game.GameState{}, false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.games[gameID]
	if !ok {
		return game.GameState{}, false
	}

	return engine.CloneState(state), true
}

func mergeAuthoritativeState(visible game.GameState, authoritative game.GameState) game.GameState {
	merged := engine.CloneState(visible)
	merged.TrackAllHands = authoritative.TrackAllHands
	merged.Deck = append([]game.CardID(nil), authoritative.Deck...)
	merged.DiscardPile = append([]game.CardID(nil), authoritative.DiscardPile...)
	merged.QuestDeck = append([]game.QuestID(nil), authoritative.QuestDeck...)
	merged.QuestDiscard = append([]game.QuestID(nil), authoritative.QuestDiscard...)
	merged.OtherHandCounts = copyOtherHandCounts(authoritative.OtherHandCounts)
	merged.Marquise.CardsInHand = append([]game.Card(nil), authoritative.Marquise.CardsInHand...)
	merged.Eyrie.CardsInHand = append([]game.Card(nil), authoritative.Eyrie.CardsInHand...)
	merged.Alliance.CardsInHand = append([]game.Card(nil), authoritative.Alliance.CardsInHand...)
	merged.Alliance.Supporters = append([]game.Card(nil), authoritative.Alliance.Supporters...)
	merged.Vagabond.CardsInHand = append([]game.Card(nil), authoritative.Vagabond.CardsInHand...)
	return merged
}

func copyOtherHandCounts(source map[game.Faction]int) map[game.Faction]int {
	if source == nil {
		return nil
	}

	cloned := make(map[game.Faction]int, len(source))
	for faction, count := range source {
		cloned[faction] = count
	}
	return cloned
}

func redactStateForPlayer(state game.GameState) game.GameState {
	redacted := engine.CloneState(state)
	redacted.TrackAllHands = false
	if redacted.GameMode != game.GameModeOnline {
		return redacted
	}

	if redacted.OtherHandCounts == nil {
		redacted.OtherHandCounts = map[game.Faction]int{}
	}

	redacted.Deck = make([]game.CardID, len(state.Deck))
	redacted.QuestDeck = make([]game.QuestID, len(state.QuestDeck))

	for _, faction := range redacted.TurnOrder {
		if faction == redacted.PlayerFaction {
			continue
		}

		redacted.OtherHandCounts[faction] = factionHandCount(state, faction)
		clearFactionHand(&redacted, faction)
	}

	if redacted.PlayerFaction != game.Alliance {
		redacted.Alliance.Supporters = nil
	}

	return redacted
}

func redactEffectResultForPlayer(before game.GameState, after game.GameState, action game.Action, result *game.EffectResult) *game.EffectResult {
	if result == nil || action.UsePersistentEffect == nil {
		return result
	}

	playerFaction := after.PlayerFaction
	switch action.UsePersistentEffect.EffectID {
	case "codebreakers":
		if action.UsePersistentEffect.Faction != playerFaction {
			return nil
		}
	case "stand_and_deliver":
		if action.UsePersistentEffect.Faction != playerFaction && action.UsePersistentEffect.TargetFaction != playerFaction {
			return nil
		}
	}

	cloned := &game.EffectResult{
		EffectID: result.EffectID,
		Message:  result.Message,
	}
	if len(result.Cards) > 0 {
		cloned.Cards = append([]game.Card(nil), result.Cards...)
	}

	return cloned
}

func factionHandCount(state game.GameState, faction game.Faction) int {
	switch faction {
	case game.Marquise:
		return len(state.Marquise.CardsInHand)
	case game.Eyrie:
		return len(state.Eyrie.CardsInHand)
	case game.Alliance:
		return len(state.Alliance.CardsInHand)
	case game.Vagabond:
		return len(state.Vagabond.CardsInHand)
	default:
		return 0
	}
}

func clearFactionHand(state *game.GameState, faction game.Faction) {
	switch faction {
	case game.Marquise:
		state.Marquise.CardsInHand = nil
	case game.Eyrie:
		state.Eyrie.CardsInHand = nil
	case game.Alliance:
		state.Alliance.CardsInHand = nil
	case game.Vagabond:
		state.Vagabond.CardsInHand = nil
	}
}
