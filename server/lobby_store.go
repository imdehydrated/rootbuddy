package server

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

var (
	errPlayerTokenRequired     = errors.New("player token is required")
	errUnknownJoinCode         = errors.New("unknown join code")
	errLobbyNotFound           = errors.New("lobby not found")
	errPlayerNotFound          = errors.New("player not found")
	errLobbyFull               = errors.New("lobby is full")
	errLobbyNotWaiting         = errors.New("lobby is not accepting changes")
	errHostOnly                = errors.New("only the host can perform this action")
	errFactionRequired         = errors.New("claim a faction before readying")
	errFactionUnavailable      = errors.New("faction is not available in this lobby")
	errFactionClaimed          = errors.New("faction is already claimed")
	errLobbyNotReady           = errors.New("all players must claim a unique faction and be ready")
	errLobbySessionUnavailable = errors.New("multiplayer session is unavailable")
	errGameIDGeneration        = errors.New("failed to generate game id")
	errCannotLeaveInGame       = errors.New("cannot leave an in-progress lobby")
)

type lobbyStore struct {
	mu      sync.RWMutex
	byCode  map[string]*Lobby
	byGame  map[string]*Lobby
	byToken map[string]string
}

func newLobbyStore() *lobbyStore {
	return &lobbyStore{
		byCode:  map[string]*Lobby{},
		byGame:  map[string]*Lobby{},
		byToken: map[string]string{},
	}
}

var lobbies = newLobbyStore()

func (s *lobbyStore) createLobby(req CreateLobbyRequest) (Lobby, string, error) {
	factions, err := normalizeLobbyFactions(req.Factions)
	if err != nil {
		return Lobby{}, "", err
	}

	playerToken, err := newPlayerToken()
	if err != nil {
		return Lobby{}, "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	joinCode, err := s.newUniqueJoinCodeLocked()
	if err != nil {
		return Lobby{}, "", err
	}

	lobby := &Lobby{
		JoinCode:  joinCode,
		State:     LobbyWaiting,
		HostToken: playerToken,
		Players: []PlayerSlot{{
			PlayerToken: playerToken,
			DisplayName: req.DisplayName,
			IsHost:      true,
			Connected:   false,
		}},
		Factions:  factions,
		MapID:     req.MapID,
		CreatedAt: time.Now().UTC(),
	}

	if lobby.MapID == "" {
		lobby.MapID = game.AutumnMapID
	}

	s.byCode[joinCode] = lobby
	s.byToken[playerToken] = joinCode

	return cloneLobby(lobby), playerToken, nil
}

func (s *lobbyStore) joinLobby(joinCode string, displayName string) (Lobby, string, error) {
	playerToken, err := newPlayerToken()
	if err != nil {
		return Lobby{}, "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lobby, ok := s.byCode[joinCode]
	if !ok {
		return Lobby{}, "", errUnknownJoinCode
	}
	if lobby.State != LobbyWaiting {
		return Lobby{}, "", errLobbyNotWaiting
	}
	if len(lobby.Players) >= len(lobby.Factions) {
		return Lobby{}, "", errLobbyFull
	}

	lobby.Players = append(lobby.Players, PlayerSlot{
		PlayerToken: playerToken,
		DisplayName: displayName,
		Connected:   false,
	})
	s.byToken[playerToken] = joinCode

	return cloneLobby(lobby), playerToken, nil
}

func (s *lobbyStore) getByToken(token string) (Lobby, bool) {
	if token == "" {
		return Lobby{}, false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	joinCode, ok := s.byToken[token]
	if !ok {
		return Lobby{}, false
	}

	lobby, ok := s.byCode[joinCode]
	if !ok {
		return Lobby{}, false
	}

	return cloneLobby(lobby), true
}

func (s *lobbyStore) getByGameID(gameID string) (Lobby, bool) {
	if gameID == "" {
		return Lobby{}, false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	lobby, ok := s.byGame[gameID]
	if !ok {
		return Lobby{}, false
	}

	return cloneLobby(lobby), true
}

func (s *lobbyStore) claimFaction(token string, faction *game.Faction) (Lobby, error) {
	if token == "" {
		return Lobby{}, errPlayerTokenRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lobby, player, err := s.playerLobbyLocked(token)
	if err != nil {
		return Lobby{}, err
	}
	if lobby.State != LobbyWaiting {
		return Lobby{}, errLobbyNotWaiting
	}

	if faction == nil {
		player.HasFaction = false
		player.Faction = 0
		player.IsReady = false
		return cloneLobby(lobby), nil
	}

	if !lobbySupportsFaction(*lobby, *faction) {
		return Lobby{}, errFactionUnavailable
	}
	for _, slot := range lobby.Players {
		if slot.PlayerToken == token || !slot.HasFaction {
			continue
		}
		if slot.Faction == *faction {
			return Lobby{}, errFactionClaimed
		}
	}

	player.Faction = *faction
	player.HasFaction = true
	player.IsReady = false
	return cloneLobby(lobby), nil
}

func (s *lobbyStore) setReady(token string, ready *bool) (Lobby, error) {
	if token == "" {
		return Lobby{}, errPlayerTokenRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lobby, player, err := s.playerLobbyLocked(token)
	if err != nil {
		return Lobby{}, err
	}
	if lobby.State != LobbyWaiting {
		return Lobby{}, errLobbyNotWaiting
	}
	if !player.HasFaction {
		return Lobby{}, errFactionRequired
	}

	if ready == nil {
		player.IsReady = !player.IsReady
	} else {
		player.IsReady = *ready
	}
	return cloneLobby(lobby), nil
}

func (s *lobbyStore) setConnected(token string, connected bool) (Lobby, bool, error) {
	if token == "" {
		return Lobby{}, false, errPlayerTokenRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lobby, player, err := s.playerLobbyLocked(token)
	if err != nil {
		return Lobby{}, false, err
	}
	if player.Connected == connected {
		return cloneLobby(lobby), false, nil
	}

	player.Connected = connected
	return cloneLobby(lobby), true, nil
}

func (s *lobbyStore) startLobby(token string) (Lobby, game.GameState, int64, error) {
	if token == "" {
		return Lobby{}, game.GameState{}, 0, errPlayerTokenRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lobby, player, err := s.playerLobbyLocked(token)
	if err != nil {
		return Lobby{}, game.GameState{}, 0, err
	}
	if lobby.State != LobbyWaiting {
		return Lobby{}, game.GameState{}, 0, errLobbyNotWaiting
	}
	if !player.IsHost {
		return Lobby{}, game.GameState{}, 0, errHostOnly
	}
	if !lobbyReadyToStart(*lobby) {
		return Lobby{}, game.GameState{}, 0, errLobbyNotReady
	}

	hostFaction, ok := lobby.claimedFaction(token)
	if !ok {
		return Lobby{}, game.GameState{}, 0, errLobbyNotReady
	}
	randomSeed, err := multiplayerRandomSeedSource()
	if err != nil {
		return Lobby{}, game.GameState{}, 0, err
	}

	authoritative, err := engine.SetupGame(engine.SetupRequest{
		GameMode:      game.GameModeOnline,
		PlayerFaction: hostFaction,
		Factions:      lobby.claimedFactions(),
		MapID:         lobby.MapID,
		RandomSeed:    randomSeed,
	})
	if err != nil {
		return Lobby{}, game.GameState{}, 0, err
	}
	if err := engine.ValidateState(authoritative); err != nil {
		return Lobby{}, game.GameState{}, 0, err
	}

	gameID := newGameID()
	if gameID == "" {
		return Lobby{}, game.GameState{}, 0, errGameIDGeneration
	}

	authoritative.TrackAllHands = true
	record, err := store.createMultiplayer(gameID, authoritative)
	if err != nil {
		return Lobby{}, game.GameState{}, 0, err
	}
	actionLogs.ensureGame(gameID)

	lobby.GameID = gameID
	lobby.State = LobbyInGame
	s.byGame[gameID] = lobby

	return cloneLobby(lobby), redactStateForPlayer(record.State, hostFaction), record.Revision, nil
}

func (s *lobbyStore) closeGameLobby(gameID string) (Lobby, bool, error) {
	if gameID == "" {
		return Lobby{}, false, errLobbyNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lobby, ok := s.byGame[gameID]
	if !ok {
		return Lobby{}, false, errLobbyNotFound
	}
	if lobby.State == LobbyClosed {
		return cloneLobby(lobby), false, nil
	}

	lobby.State = LobbyClosed
	return cloneLobby(lobby), true, nil
}

func (s *lobbyStore) leaveLobby(token string) (*Lobby, bool, error) {
	if token == "" {
		return nil, false, errPlayerTokenRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lobby, playerIndex, err := s.playerLobbyIndexLocked(token)
	if err != nil {
		return nil, false, err
	}
	if lobby.State != LobbyWaiting {
		return nil, false, errCannotLeaveInGame
	}

	delete(s.byToken, token)
	lobby.Players = append(lobby.Players[:playerIndex], lobby.Players[playerIndex+1:]...)

	if len(lobby.Players) == 0 {
		delete(s.byCode, lobby.JoinCode)
		if lobby.GameID != "" {
			delete(s.byGame, lobby.GameID)
		}
		return nil, true, nil
	}

	if lobby.HostToken == token {
		lobby.setHost(lobby.Players[0].PlayerToken)
	}

	cloned := cloneLobby(lobby)
	return &cloned, false, nil
}

func (s *lobbyStore) playerFactionByGame(gameID string, token string) (game.Faction, bool) {
	lobby, ok := s.getByGameID(gameID)
	if !ok {
		return 0, false
	}

	return lobby.claimedFaction(token)
}

func (s *lobbyStore) playerLobbyLocked(token string) (*Lobby, *PlayerSlot, error) {
	lobby, index, err := s.playerLobbyIndexLocked(token)
	if err != nil {
		return nil, nil, err
	}
	return lobby, &lobby.Players[index], nil
}

func (s *lobbyStore) playerLobbyIndexLocked(token string) (*Lobby, int, error) {
	joinCode, ok := s.byToken[token]
	if !ok {
		return nil, -1, errPlayerNotFound
	}

	lobby, ok := s.byCode[joinCode]
	if !ok {
		return nil, -1, errLobbyNotFound
	}

	index := lobby.playerIndex(token)
	if index == -1 {
		return nil, -1, errPlayerNotFound
	}

	return lobby, index, nil
}

func (s *lobbyStore) newUniqueJoinCodeLocked() (string, error) {
	for attempt := 0; attempt < 32; attempt++ {
		joinCode, err := newJoinCode()
		if err != nil {
			return "", err
		}
		if _, exists := s.byCode[joinCode]; !exists {
			return joinCode, nil
		}
	}

	return "", errors.New("failed to generate join code")
}

func normalizeLobbyFactions(requested []game.Faction) ([]game.Faction, error) {
	if len(requested) == 0 {
		return append([]game.Faction(nil), defaultLobbyFactions...), nil
	}

	present := map[game.Faction]bool{}
	for _, faction := range requested {
		if !lobbySupportsFactionList(faction) {
			return nil, errFactionUnavailable
		}
		if present[faction] {
			return nil, errors.New("lobby factions must be unique")
		}
		present[faction] = true
	}
	if len(present) < 2 || len(present) > 4 {
		return nil, errors.New("lobby requires between 2 and 4 factions")
	}

	ordered := make([]game.Faction, 0, len(present))
	for _, faction := range defaultLobbyFactions {
		if present[faction] {
			ordered = append(ordered, faction)
		}
	}
	return ordered, nil
}

func lobbySupportsFaction(lobby Lobby, faction game.Faction) bool {
	for _, available := range lobby.Factions {
		if available == faction {
			return true
		}
	}
	return false
}

func lobbySupportsFactionList(faction game.Faction) bool {
	for _, available := range defaultLobbyFactions {
		if available == faction {
			return true
		}
	}
	return false
}

func lobbyReadyToStart(lobby Lobby) bool {
	if len(lobby.Players) < 2 || len(lobby.Players) > 4 {
		return false
	}

	claimed := map[game.Faction]bool{}
	for _, player := range lobby.Players {
		if !player.HasFaction || !player.IsReady {
			return false
		}
		if claimed[player.Faction] {
			return false
		}
		claimed[player.Faction] = true
	}

	return true
}

func newPlayerToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func newJoinCode() (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.Grow(len(bytes))
	for _, value := range bytes {
		builder.WriteByte(alphabet[int(value)%len(alphabet)])
	}
	return builder.String(), nil
}
