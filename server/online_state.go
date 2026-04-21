package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

var errRevisionConflict = errors.New("revision conflict")

type authoritativeGameRecord struct {
	GameID        string         `json:"gameID"`
	Revision      int64          `json:"revision"`
	SavedAt       time.Time      `json:"savedAt"`
	RequiresLobby bool           `json:"requiresLobby,omitempty"`
	State         game.GameState `json:"state"`
}

type onlineStateRepository interface {
	create(gameID string, state game.GameState) (authoritativeGameRecord, error)
	createMultiplayer(gameID string, state game.GameState) (authoritativeGameRecord, error)
	load(gameID string) (authoritativeGameRecord, bool, error)
	save(gameID string, state game.GameState) (authoritativeGameRecord, error)
	saveIfRevision(gameID string, expectedRevision int64, state game.GameState) (authoritativeGameRecord, error)
}

type onlineStateStore struct {
	mu    sync.RWMutex
	games map[string]authoritativeGameRecord
	dir   string
}

func newOnlineStateStore(dir string) *onlineStateStore {
	return &onlineStateStore{
		games: map[string]authoritativeGameRecord{},
		dir:   dir,
	}
}

var store onlineStateRepository = newOnlineStateStore(filepath.Join(".rootbuddy-saves", "online"))

func newGameID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}

	return hex.EncodeToString(bytes)
}

func (s *onlineStateStore) create(gameID string, state game.GameState) (authoritativeGameRecord, error) {
	if gameID == "" {
		return authoritativeGameRecord{}, errors.New("game id is required")
	}

	return s.writeRecord(gameID, 1, false, state)
}

func (s *onlineStateStore) createMultiplayer(gameID string, state game.GameState) (authoritativeGameRecord, error) {
	if gameID == "" {
		return authoritativeGameRecord{}, errors.New("game id is required")
	}

	return s.writeRecord(gameID, 1, true, state)
}

func (s *onlineStateStore) load(gameID string) (authoritativeGameRecord, bool, error) {
	if gameID == "" {
		return authoritativeGameRecord{}, false, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if record, ok := s.games[gameID]; ok {
		return cloneAuthoritativeGameRecord(record), true, nil
	}

	record, ok, err := s.loadRecordFromDisk(gameID)
	if err != nil || !ok {
		return authoritativeGameRecord{}, ok, err
	}

	s.games[gameID] = record
	return cloneAuthoritativeGameRecord(record), true, nil
}

func (s *onlineStateStore) save(gameID string, state game.GameState) (authoritativeGameRecord, error) {
	if gameID == "" {
		return authoritativeGameRecord{}, errors.New("game id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok, err := s.currentRecordLocked(gameID)
	if err != nil {
		return authoritativeGameRecord{}, err
	}

	revision := int64(1)
	requiresLobby := false
	if ok {
		revision = record.Revision + 1
		requiresLobby = record.RequiresLobby
	}

	return s.writeRecordLocked(gameID, revision, requiresLobby, state)
}

func (s *onlineStateStore) saveIfRevision(gameID string, expectedRevision int64, state game.GameState) (authoritativeGameRecord, error) {
	if gameID == "" {
		return authoritativeGameRecord{}, errors.New("game id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok, err := s.currentRecordLocked(gameID)
	if err != nil {
		return authoritativeGameRecord{}, err
	}
	if !ok {
		return authoritativeGameRecord{}, errors.New("unknown game id")
	}
	if record.Revision != expectedRevision {
		return authoritativeGameRecord{}, errRevisionConflict
	}

	return s.writeRecordLocked(gameID, expectedRevision+1, record.RequiresLobby, state)
}

func (s *onlineStateStore) writeRecord(gameID string, revision int64, requiresLobby bool, state game.GameState) (authoritativeGameRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.writeRecordLocked(gameID, revision, requiresLobby, state)
}

func (s *onlineStateStore) writeRecordLocked(gameID string, revision int64, requiresLobby bool, state game.GameState) (authoritativeGameRecord, error) {
	record := authoritativeGameRecord{
		GameID:        gameID,
		Revision:      revision,
		SavedAt:       time.Now().UTC(),
		RequiresLobby: requiresLobby,
		State:         engine.CloneState(state),
	}

	if err := s.persistRecordLocked(record); err != nil {
		return authoritativeGameRecord{}, err
	}

	s.games[gameID] = record
	return cloneAuthoritativeGameRecord(record), nil
}

func (s *onlineStateStore) currentRecordLocked(gameID string) (authoritativeGameRecord, bool, error) {
	if record, ok := s.games[gameID]; ok {
		return record, true, nil
	}

	record, ok, err := s.loadRecordFromDisk(gameID)
	if err != nil || !ok {
		return authoritativeGameRecord{}, ok, err
	}

	s.games[gameID] = record
	return record, true, nil
}

func (s *onlineStateStore) persistRecordLocked(record authoritativeGameRecord) error {
	if s.dir == "" {
		return nil
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath(record.GameID), payload, 0o644)
}

func (s *onlineStateStore) loadRecordFromDisk(gameID string) (authoritativeGameRecord, bool, error) {
	if s.dir == "" {
		return authoritativeGameRecord{}, false, nil
	}

	payload, err := os.ReadFile(s.filePath(gameID))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return authoritativeGameRecord{}, false, nil
		}
		return authoritativeGameRecord{}, false, err
	}

	var record authoritativeGameRecord
	if err := json.Unmarshal(payload, &record); err != nil {
		return authoritativeGameRecord{}, false, err
	}
	record.State = engine.CloneState(record.State)
	return record, true, nil
}

func (s *onlineStateStore) filePath(gameID string) string {
	return filepath.Join(s.dir, gameID+".json")
}

func cloneAuthoritativeGameRecord(record authoritativeGameRecord) authoritativeGameRecord {
	cloned := record
	cloned.State = engine.CloneState(record.State)
	return cloned
}

func mergeAuthoritativeState(visible game.GameState, authoritative game.GameState) game.GameState {
	merged := engine.CloneState(visible)
	merged.TrackAllHands = authoritative.TrackAllHands
	merged.Deck = append([]game.CardID(nil), authoritative.Deck...)
	merged.DiscardPile = append([]game.CardID(nil), authoritative.DiscardPile...)
	merged.QuestDeck = append([]game.QuestID(nil), authoritative.QuestDeck...)
	merged.QuestDiscard = append([]game.QuestID(nil), authoritative.QuestDiscard...)
	merged.OtherHandCounts = copyOtherHandCounts(authoritative.OtherHandCounts)
	merged.HiddenCards = append([]game.HiddenCard(nil), authoritative.HiddenCards...)
	merged.NextHiddenCardID = authoritative.NextHiddenCardID
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

func redactStateForPlayer(state game.GameState, perspective game.Faction) game.GameState {
	redacted := engine.CloneState(state)
	redacted.PlayerFaction = perspective
	redacted.TrackAllHands = false
	if redacted.GameMode != game.GameModeOnline {
		return redacted
	}
	redacted.HiddenCards = nil
	redacted.NextHiddenCardID = 0

	if redacted.OtherHandCounts == nil {
		redacted.OtherHandCounts = map[game.Faction]int{}
	}

	redacted.Deck = make([]game.CardID, len(state.Deck))
	redacted.QuestDeck = make([]game.QuestID, len(state.QuestDeck))

	for _, faction := range redacted.TurnOrder {
		if faction == perspective {
			continue
		}

		redacted.OtherHandCounts[faction] = factionHandCount(state, faction)
		clearFactionHand(&redacted, faction)
	}

	if perspective != game.Alliance {
		redacted.HiddenCards = hiddenAllianceSupporterPlaceholders(state)
		redacted.Alliance.Supporters = nil
	}

	return redacted
}

func hiddenAllianceSupporterPlaceholders(state game.GameState) []game.HiddenCard {
	count := len(state.Alliance.Supporters)
	if count == 0 {
		for _, hidden := range state.HiddenCards {
			if hidden.OwnerFaction == game.Alliance && hidden.Zone == game.HiddenCardZoneSupporters {
				count++
			}
		}
	}
	if count == 0 {
		return nil
	}

	hidden := make([]game.HiddenCard, 0, count)
	for i := 0; i < count; i++ {
		hidden = append(hidden, game.HiddenCard{
			ID:           i + 1,
			OwnerFaction: game.Alliance,
			Zone:         game.HiddenCardZoneSupporters,
		})
	}
	return hidden
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
