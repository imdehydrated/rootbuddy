package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestLobbyLifecycleAndTurnEnforcement(t *testing.T) {
	previousStore := store
	previousLobbies := lobbies
	store = newOnlineStateStore(t.TempDir())
	lobbies = newLobbyStore()
	defer func() {
		store = previousStore
		lobbies = previousLobbies
	}()

	server := NewServer()

	createBody, _ := json.Marshal(CreateLobbyRequest{
		DisplayName: "Host",
		Factions:    []game.Faction{game.Marquise, game.Eyrie},
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/lobby/create", bytes.NewReader(createBody))
	createRec := httptest.NewRecorder()
	server.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from create lobby, got %d body=%s", createRec.Code, createRec.Body.String())
	}

	var createResp CreateLobbyResponse
	if err := json.Unmarshal(createRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("failed to decode create lobby response: %v", err)
	}
	if matched := regexp.MustCompile(`^[A-Z0-9]{6}$`).MatchString(createResp.Lobby.JoinCode); !matched {
		t.Fatalf("expected 6-char uppercase join code, got %q", createResp.Lobby.JoinCode)
	}
	if createResp.PlayerToken == "" {
		t.Fatalf("expected host player token")
	}
	if !createResp.Lobby.Players[0].IsHost {
		t.Fatalf("expected creator to be host, got %+v", createResp.Lobby.Players)
	}

	joinBody, _ := json.Marshal(JoinLobbyRequest{
		JoinCode:    createResp.Lobby.JoinCode,
		DisplayName: "Bird",
	})
	joinReq := httptest.NewRequest(http.MethodPost, "/api/lobby/join", bytes.NewReader(joinBody))
	joinRec := httptest.NewRecorder()
	server.ServeHTTP(joinRec, joinReq)

	if joinRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from join lobby, got %d body=%s", joinRec.Code, joinRec.Body.String())
	}

	var joinResp JoinLobbyResponse
	if err := json.Unmarshal(joinRec.Body.Bytes(), &joinResp); err != nil {
		t.Fatalf("failed to decode join lobby response: %v", err)
	}
	if len(joinResp.Lobby.Players) != 2 {
		t.Fatalf("expected two players after join, got %+v", joinResp.Lobby.Players)
	}

	marquise := game.Marquise
	hostClaimBody, _ := json.Marshal(ClaimFactionRequest{Faction: &marquise})
	hostClaimReq := httptest.NewRequest(http.MethodPost, "/api/lobby/claim-faction", bytes.NewReader(hostClaimBody))
	hostClaimReq.Header.Set("X-Player-Token", createResp.PlayerToken)
	hostClaimRec := httptest.NewRecorder()
	server.ServeHTTP(hostClaimRec, hostClaimReq)

	if hostClaimRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from host claim, got %d body=%s", hostClaimRec.Code, hostClaimRec.Body.String())
	}

	eyrie := game.Eyrie
	joinClaimBody, _ := json.Marshal(ClaimFactionRequest{Faction: &eyrie})
	joinClaimReq := httptest.NewRequest(http.MethodPost, "/api/lobby/claim-faction", bytes.NewReader(joinClaimBody))
	joinClaimReq.Header.Set("X-Player-Token", joinResp.PlayerToken)
	joinClaimRec := httptest.NewRecorder()
	server.ServeHTTP(joinClaimRec, joinClaimReq)

	if joinClaimRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from joiner claim, got %d body=%s", joinClaimRec.Code, joinClaimRec.Body.String())
	}

	ready := true
	for _, token := range []string{createResp.PlayerToken, joinResp.PlayerToken} {
		readyBody, _ := json.Marshal(ReadyLobbyRequest{IsReady: &ready})
		readyReq := httptest.NewRequest(http.MethodPost, "/api/lobby/ready", bytes.NewReader(readyBody))
		readyReq.Header.Set("X-Player-Token", token)
		readyRec := httptest.NewRecorder()
		server.ServeHTTP(readyRec, readyReq)

		if readyRec.Code != http.StatusOK {
			t.Fatalf("expected 200 from ready update, got %d body=%s", readyRec.Code, readyRec.Body.String())
		}
	}

	stateReq := httptest.NewRequest(http.MethodGet, "/api/lobby/state", nil)
	stateReq.Header.Set("X-Player-Token", joinResp.PlayerToken)
	stateRec := httptest.NewRecorder()
	server.ServeHTTP(stateRec, stateReq)

	if stateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from lobby state, got %d body=%s", stateRec.Code, stateRec.Body.String())
	}

	var stateResp LobbyResponse
	if err := json.Unmarshal(stateRec.Body.Bytes(), &stateResp); err != nil {
		t.Fatalf("failed to decode lobby state response: %v", err)
	}
	if !stateResp.Lobby.Players[1].HasFaction || stateResp.Lobby.Players[1].Faction != game.Eyrie {
		t.Fatalf("expected claimed faction in lobby state, got %+v", stateResp.Lobby.Players)
	}

	startReq := httptest.NewRequest(http.MethodPost, "/api/lobby/start", bytes.NewBufferString(`{}`))
	startReq.Header.Set("X-Player-Token", createResp.PlayerToken)
	startRec := httptest.NewRecorder()
	server.ServeHTTP(startRec, startReq)

	if startRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from start lobby, got %d body=%s", startRec.Code, startRec.Body.String())
	}

	var startResp StartLobbyResponse
	if err := json.Unmarshal(startRec.Body.Bytes(), &startResp); err != nil {
		t.Fatalf("failed to decode start response: %v", err)
	}
	if startResp.GameID == "" || startResp.Lobby.GameID == "" {
		t.Fatalf("expected start response to include game id, got %+v", startResp)
	}
	if startResp.Revision <= 0 {
		t.Fatalf("expected start response revision, got %+v", startResp)
	}
	if startResp.Lobby.State != LobbyInGame {
		t.Fatalf("expected in-game lobby state, got %+v", startResp.Lobby)
	}
	if startResp.State.PlayerFaction != game.Marquise {
		t.Fatalf("expected host perspective in start response, got %+v", startResp.State.PlayerFaction)
	}

	loadReq := httptest.NewRequest(http.MethodPost, "/api/game/load", bytes.NewBufferString(`{"gameID":"`+startResp.GameID+`"}`))
	loadReq.Header.Set("X-Player-Token", joinResp.PlayerToken)
	loadRec := httptest.NewRecorder()
	server.ServeHTTP(loadRec, loadReq)

	if loadRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from multiplayer load, got %d body=%s", loadRec.Code, loadRec.Body.String())
	}

	var loadResp LoadGameResponse
	if err := json.Unmarshal(loadRec.Body.Bytes(), &loadResp); err != nil {
		t.Fatalf("failed to decode load response: %v", err)
	}
	if loadResp.State.PlayerFaction != game.Eyrie {
		t.Fatalf("expected load to use token perspective, got %+v", loadResp.State.PlayerFaction)
	}

	wrongTurnBody, _ := json.Marshal(ApplyActionRequest{
		GameID:         startResp.GameID,
		State:          loadResp.State,
		ClientRevision: loadResp.Revision,
		Action: game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Eyrie,
			},
		},
	})
	wrongTurnReq := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(wrongTurnBody))
	wrongTurnReq.Header.Set("X-Player-Token", joinResp.PlayerToken)
	wrongTurnRec := httptest.NewRecorder()
	server.ServeHTTP(wrongTurnRec, wrongTurnReq)

	if wrongTurnRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for wrong-turn apply, got %d body=%s", wrongTurnRec.Code, wrongTurnRec.Body.String())
	}

	errResp := decodeErrorResponse(t, wrongTurnRec)
	if errResp.Error != "not your turn" {
		t.Fatalf("unexpected wrong-turn response: %+v", errResp)
	}
}

func TestLobbyLeaveTransfersHost(t *testing.T) {
	previousLobbies := lobbies
	lobbies = newLobbyStore()
	defer func() {
		lobbies = previousLobbies
	}()

	lobby, hostToken, err := lobbies.createLobby(CreateLobbyRequest{
		DisplayName: "Host",
		Factions:    []game.Faction{game.Marquise, game.Eyrie},
	})
	if err != nil {
		t.Fatalf("create lobby failed: %v", err)
	}

	joined, playerToken, err := lobbies.joinLobby(lobby.JoinCode, "Bird")
	if err != nil {
		t.Fatalf("join lobby failed: %v", err)
	}
	if len(joined.Players) != 2 {
		t.Fatalf("expected two players before leave, got %+v", joined.Players)
	}

	remaining, closed, err := lobbies.leaveLobby(hostToken)
	if err != nil {
		t.Fatalf("leave lobby failed: %v", err)
	}
	if closed {
		t.Fatalf("expected lobby to remain open after host leaves")
	}
	if remaining == nil || remaining.HostToken != playerToken {
		t.Fatalf("expected host transfer to remaining player, got %+v", remaining)
	}
	if len(remaining.Players) != 1 || !remaining.Players[0].IsHost {
		t.Fatalf("expected remaining player to become host, got %+v", remaining.Players)
	}
}
