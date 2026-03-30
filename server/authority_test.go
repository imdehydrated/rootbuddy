package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

type failingSaveRepository struct {
	record  authoritativeGameRecord
	saveErr error
}

func (r *failingSaveRepository) create(gameID string, state game.GameState) (authoritativeGameRecord, error) {
	record := authoritativeGameRecord{GameID: gameID, Revision: 1, State: engine.CloneState(state)}
	r.record = record
	return cloneAuthoritativeGameRecord(record), nil
}

func (r *failingSaveRepository) createMultiplayer(gameID string, state game.GameState) (authoritativeGameRecord, error) {
	record := authoritativeGameRecord{GameID: gameID, Revision: 1, RequiresLobby: true, State: engine.CloneState(state)}
	r.record = record
	return cloneAuthoritativeGameRecord(record), nil
}

func (r *failingSaveRepository) load(gameID string) (authoritativeGameRecord, bool, error) {
	if gameID != r.record.GameID {
		return authoritativeGameRecord{}, false, nil
	}
	return cloneAuthoritativeGameRecord(r.record), true, nil
}

func (r *failingSaveRepository) save(gameID string, state game.GameState) (authoritativeGameRecord, error) {
	return authoritativeGameRecord{}, r.saveErr
}

func (r *failingSaveRepository) saveIfRevision(gameID string, expectedRevision int64, state game.GameState) (authoritativeGameRecord, error) {
	return authoritativeGameRecord{}, r.saveErr
}

func startLobbyBackedGame(t *testing.T) (string, string, string, game.GameState, int64) {
	t.Helper()

	lobby, hostToken, err := lobbies.createLobby(CreateLobbyRequest{
		DisplayName: "Host",
		Factions:    []game.Faction{game.Marquise, game.Eyrie},
	})
	if err != nil {
		t.Fatalf("create lobby failed: %v", err)
	}

	_, birdToken, err := lobbies.joinLobby(lobby.JoinCode, "Bird")
	if err != nil {
		t.Fatalf("join lobby failed: %v", err)
	}

	marquise := game.Marquise
	if _, err := lobbies.claimFaction(hostToken, &marquise); err != nil {
		t.Fatalf("host claim failed: %v", err)
	}
	eyrie := game.Eyrie
	if _, err := lobbies.claimFaction(birdToken, &eyrie); err != nil {
		t.Fatalf("joiner claim failed: %v", err)
	}
	ready := true
	if _, err := lobbies.setReady(hostToken, &ready); err != nil {
		t.Fatalf("host ready failed: %v", err)
	}
	if _, err := lobbies.setReady(birdToken, &ready); err != nil {
		t.Fatalf("joiner ready failed: %v", err)
	}

	lobby, state, revision, err := lobbies.startLobby(hostToken)
	if err != nil {
		t.Fatalf("start lobby failed: %v", err)
	}

	return lobby.GameID, hostToken, birdToken, state, revision
}

func replaceAuthoritativeState(t *testing.T, gameID string, mutate func(state *game.GameState)) authoritativeGameRecord {
	t.Helper()

	record, ok, err := store.load(gameID)
	if err != nil {
		t.Fatalf("failed to load authoritative record: %v", err)
	}
	if !ok {
		t.Fatalf("expected authoritative record for %s", gameID)
	}

	state := engine.CloneState(record.State)
	mutate(&state)

	saved, err := store.save(gameID, state)
	if err != nil {
		t.Fatalf("failed to replace authoritative state: %v", err)
	}
	return saved
}

func TestHandleValidActionsMultiplayerIgnoresClientStateTampering(t *testing.T) {
	previousStore := store
	previousLobbies := lobbies
	store = newOnlineStateStore(t.TempDir())
	lobbies = newLobbyStore()
	defer func() {
		store = previousStore
		lobbies = previousLobbies
	}()

	gameID, hostToken, _, hostState, revision := startLobbyBackedGame(t)
	tampered := engine.CloneState(hostState)
	tampered.GamePhase = game.LifecyclePlaying
	tampered.SetupStage = game.SetupStageComplete
	tampered.CurrentPhase = game.Evening
	tampered.CurrentStep = game.StepEvening
	tampered.Map.Clearings = nil

	body, _ := json.Marshal(ValidActionsRequest{
		GameID: gameID,
		State:  tampered,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/actions/valid", bytes.NewReader(body))
	req.Header.Set("X-Player-Token", hostToken)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for multiplayer valid actions, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp ValidActionsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode valid actions response: %v", err)
	}
	if resp.Revision != revision {
		t.Fatalf("expected revision %d, got %d", revision, resp.Revision)
	}

	foundSetup := false
	for _, action := range resp.Actions {
		if action.Type == game.ActionMarquiseSetup {
			foundSetup = true
			break
		}
	}
	if !foundSetup {
		t.Fatalf("expected authoritative setup actions, got %+v", resp.Actions)
	}
}

func TestHandleValidActionsMultiplayerReturnsEmptyForInactivePlayer(t *testing.T) {
	previousStore := store
	previousLobbies := lobbies
	store = newOnlineStateStore(t.TempDir())
	lobbies = newLobbyStore()
	defer func() {
		store = previousStore
		lobbies = previousLobbies
	}()

	gameID, _, birdToken, hostState, revision := startLobbyBackedGame(t)

	body, _ := json.Marshal(ValidActionsRequest{
		GameID: gameID,
		State:  hostState,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/actions/valid", bytes.NewReader(body))
	req.Header.Set("X-Player-Token", birdToken)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for inactive multiplayer valid actions, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp ValidActionsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode valid actions response: %v", err)
	}
	if resp.Revision != revision {
		t.Fatalf("expected revision %d, got %d", revision, resp.Revision)
	}
	if len(resp.Actions) != 0 {
		t.Fatalf("expected no actions for inactive player, got %+v", resp.Actions)
	}
}

func TestHandleLoadGameFailsWhenLobbyBackedSessionIsUnavailable(t *testing.T) {
	previousStore := store
	previousLobbies := lobbies
	store = newOnlineStateStore(t.TempDir())
	lobbies = newLobbyStore()
	defer func() {
		store = previousStore
		lobbies = previousLobbies
	}()

	gameID, hostToken, _, _, revision := startLobbyBackedGame(t)
	lobbies = newLobbyStore()

	body, _ := json.Marshal(LoadGameRequest{GameID: gameID})
	req := httptest.NewRequest(http.MethodPost, "/api/game/load", bytes.NewReader(body))
	req.Header.Set("X-Player-Token", hostToken)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 when lobby-backed session is unavailable, got %d body=%s", rec.Code, rec.Body.String())
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != errLobbySessionUnavailable.Error() {
		t.Fatalf("unexpected error response: %+v", resp)
	}
	if resp.Revision != revision {
		t.Fatalf("expected revision %d, got %d", revision, resp.Revision)
	}
}

func TestHandleBattleContextMultiplayerUsesAuthoritativeHiddenHands(t *testing.T) {
	previousStore := store
	previousLobbies := lobbies
	store = newOnlineStateStore(t.TempDir())
	lobbies = newLobbyStore()
	defer func() {
		store = previousStore
		lobbies = previousLobbies
	}()

	gameID, hostToken, _, hostState, _ := startLobbyBackedGame(t)
	record := replaceAuthoritativeState(t, gameID, func(state *game.GameState) {
		state.GamePhase = game.LifecyclePlaying
		state.SetupStage = game.SetupStageComplete
		state.FactionTurn = game.Marquise
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.StepDaylightActions
		state.TurnOrder = []game.Faction{game.Marquise, game.Eyrie}
		state.Map = game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Warriors: map[game.Faction]int{
						game.Marquise: 2,
						game.Eyrie:    1,
					},
				},
			},
		}
		state.Marquise.CardsInHand = nil
		state.Eyrie.CardsInHand = []game.Card{
			{ID: 24, Name: "A Visit to Friends", Suit: game.Rabbit},
		}
		state.OtherHandCounts = map[game.Faction]int{
			game.Eyrie: 1,
		}
	})

	body, _ := json.Marshal(BattleContextRequest{
		GameID: gameID,
		State:  hostState,
		Action: game.Action{
			Type: game.ActionBattle,
			Battle: &game.BattleAction{
				Faction:       game.Marquise,
				ClearingID:    1,
				TargetFaction: game.Eyrie,
			},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/battles/context", bytes.NewReader(body))
	req.Header.Set("X-Player-Token", hostToken)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for multiplayer battle context, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp BattleContextResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode battle context response: %v", err)
	}
	if resp.Revision != record.Revision {
		t.Fatalf("expected revision %d, got %d", record.Revision, resp.Revision)
	}
	if resp.BattleContext.CanDefenderAmbush {
		t.Fatalf("expected authoritative hidden hand to suppress false ambush, got %+v", resp.BattleContext)
	}
}

func TestHandleResolveBattleMultiplayerIgnoresFalseHiddenAmbushClaim(t *testing.T) {
	previousStore := store
	previousLobbies := lobbies
	store = newOnlineStateStore(t.TempDir())
	lobbies = newLobbyStore()
	defer func() {
		store = previousStore
		lobbies = previousLobbies
	}()

	gameID, hostToken, _, hostState, _ := startLobbyBackedGame(t)
	record := replaceAuthoritativeState(t, gameID, func(state *game.GameState) {
		state.GamePhase = game.LifecyclePlaying
		state.SetupStage = game.SetupStageComplete
		state.FactionTurn = game.Marquise
		state.CurrentPhase = game.Daylight
		state.CurrentStep = game.StepDaylightActions
		state.TurnOrder = []game.Faction{game.Marquise, game.Eyrie}
		state.Map = game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Warriors: map[game.Faction]int{
						game.Marquise: 2,
						game.Eyrie:    1,
					},
				},
			},
		}
		state.Eyrie.CardsInHand = []game.Card{
			{ID: 24, Name: "A Visit to Friends", Suit: game.Rabbit},
		}
		state.OtherHandCounts = map[game.Faction]int{
			game.Eyrie: 1,
		}
	})

	body, _ := json.Marshal(ResolveBattleRequest{
		GameID: gameID,
		State:  hostState,
		Action: game.Action{
			Type: game.ActionBattle,
			Battle: &game.BattleAction{
				Faction:       game.Marquise,
				ClearingID:    1,
				TargetFaction: game.Eyrie,
			},
		},
		AttackerRoll: 1,
		DefenderRoll: 0,
		UseModifiers: true,
		Modifiers: game.BattleModifiers{
			DefenderAmbush: true,
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/battles/resolve", bytes.NewReader(body))
	req.Header.Set("X-Player-Token", hostToken)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for multiplayer resolve battle, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp ResolveBattleResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode resolve battle response: %v", err)
	}
	if resp.Revision != record.Revision {
		t.Fatalf("expected revision %d, got %d", record.Revision, resp.Revision)
	}
	if resp.Action.BattleResolution == nil {
		t.Fatalf("expected battle resolution, got %+v", resp.Action)
	}
	if resp.Action.BattleResolution.DefenderAmbushed {
		t.Fatalf("expected false ambush claim to be ignored, got %+v", resp.Action.BattleResolution)
	}
}

func TestHandleApplyActionMultiplayerReturnsConflictForStaleRevision(t *testing.T) {
	previousStore := store
	previousLobbies := lobbies
	store = newOnlineStateStore(t.TempDir())
	lobbies = newLobbyStore()
	defer func() {
		store = previousStore
		lobbies = previousLobbies
	}()

	gameID, hostToken, _, hostState, revision := startLobbyBackedGame(t)

	var setupAction game.Action
	for _, action := range engine.ValidActions(hostState) {
		if action.Type == game.ActionMarquiseSetup {
			setupAction = action
			break
		}
	}
	if setupAction.Type != game.ActionMarquiseSetup {
		t.Fatalf("expected marquise setup action, got %+v", setupAction)
	}

	body, _ := json.Marshal(ApplyActionRequest{
		GameID:         gameID,
		State:          hostState,
		Action:         setupAction,
		ClientRevision: revision + 1,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(body))
	req.Header.Set("X-Player-Token", hostToken)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for stale multiplayer apply, got %d body=%s", rec.Code, rec.Body.String())
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != errRevisionConflict.Error() {
		t.Fatalf("expected revision conflict, got %+v", resp)
	}
	if resp.Revision != revision {
		t.Fatalf("expected latest revision %d, got %d", revision, resp.Revision)
	}
	if resp.State == nil || resp.State.PlayerFaction != game.Marquise {
		t.Fatalf("expected conflict response to include latest redacted state, got %+v", resp)
	}
}

func TestHandleLoadGameReturnsServerErrorForInvalidAuthoritativeState(t *testing.T) {
	previousStore := store
	store = newOnlineStateStore(t.TempDir())
	defer func() {
		store = previousStore
	}()

	if _, err := store.create("invalid-state", game.GameState{
		GameMode:      game.GameModeOnline,
		GamePhase:     game.LifecycleSetup,
		SetupStage:    game.SetupStageMarquise,
		CurrentStep:   game.StepBirdsong,
		PlayerFaction: game.Marquise,
	}); err != nil {
		t.Fatalf("failed to seed invalid authoritative state: %v", err)
	}

	body, _ := json.Marshal(LoadGameRequest{GameID: "invalid-state"})
	req := httptest.NewRequest(http.MethodPost, "/api/game/load", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for invalid authoritative state, got %d body=%s", rec.Code, rec.Body.String())
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "authoritative game state is invalid" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleApplyActionReturnsServerErrorWhenPersistenceFails(t *testing.T) {
	previousStore := store
	previousLobbies := lobbies
	store = &failingSaveRepository{
		record: authoritativeGameRecord{
			GameID:   "persist-fail",
			Revision: 1,
			State: game.GameState{
				GameMode:      game.GameModeOnline,
				GamePhase:     game.LifecyclePlaying,
				SetupStage:    game.SetupStageComplete,
				PlayerFaction: game.Marquise,
				FactionTurn:   game.Marquise,
				CurrentPhase:  game.Birdsong,
				CurrentStep:   game.StepBirdsong,
				TurnOrder:     []game.Faction{game.Marquise, game.Eyrie},
			},
		},
		saveErr: errors.New("disk full"),
	}
	lobbies = newLobbyStore()
	defer func() {
		store = previousStore
		lobbies = previousLobbies
	}()

	visible := redactStateForPlayer(store.(*failingSaveRepository).record.State, game.Marquise)
	body, _ := json.Marshal(ApplyActionRequest{
		GameID: "persist-fail",
		State:  visible,
		Action: game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Marquise,
			},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for persistence failure, got %d body=%s", rec.Code, rec.Body.String())
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "failed to persist authoritative game state" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}
