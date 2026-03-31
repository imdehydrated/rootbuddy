package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

func TestWebSocketGameStartUsesCurrentClaimedFactionPerspective(t *testing.T) {
	teardown := resetRealtimeTestState(t)
	defer teardown()

	testServer := httptest.NewServer(NewServer())
	defer testServer.Close()

	var createResp CreateLobbyResponse
	postJSON(t, testServer.URL, "/api/lobby/create", CreateLobbyRequest{
		DisplayName: "Host",
		Factions:    []game.Faction{game.Marquise, game.Eyrie},
	}, "", &createResp)

	var joinResp JoinLobbyResponse
	postJSON(t, testServer.URL, "/api/lobby/join", JoinLobbyRequest{
		JoinCode:    createResp.Lobby.JoinCode,
		DisplayName: "Bird",
	}, "", &joinResp)

	hostConn := dialWebSocket(t, testServer.URL, createResp.PlayerToken)
	defer hostConn.Close()
	waitForLobbyUpdate(t, hostConn, func(lobby Lobby) bool {
		return len(lobby.Players) == 2 && lobby.Players[0].Connected
	})

	joinConn := dialWebSocket(t, testServer.URL, joinResp.PlayerToken)
	defer joinConn.Close()
	waitForLobbyUpdate(t, hostConn, func(lobby Lobby) bool {
		return len(lobby.Players) == 2 && countConnected(lobby) == 2
	})
	waitForLobbyUpdate(t, joinConn, func(lobby Lobby) bool {
		return len(lobby.Players) == 2 && countConnected(lobby) == 2
	})

	marquise := game.Marquise
	postJSON(t, testServer.URL, "/api/lobby/claim-faction", ClaimFactionRequest{Faction: &marquise}, createResp.PlayerToken, nil)
	eyrie := game.Eyrie
	postJSON(t, testServer.URL, "/api/lobby/claim-faction", ClaimFactionRequest{Faction: &eyrie}, joinResp.PlayerToken, nil)

	ready := true
	postJSON(t, testServer.URL, "/api/lobby/ready", ReadyLobbyRequest{IsReady: &ready}, createResp.PlayerToken, nil)
	postJSON(t, testServer.URL, "/api/lobby/ready", ReadyLobbyRequest{IsReady: &ready}, joinResp.PlayerToken, nil)

	var startResp StartLobbyResponse
	postJSON(t, testServer.URL, "/api/lobby/start", map[string]any{}, createResp.PlayerToken, &startResp)

	hostStart := waitForGameStart(t, hostConn, func(msg GameStartMessage) bool {
		return msg.GameID == startResp.GameID && msg.Revision == startResp.Revision
	})
	if hostStart.State.PlayerFaction != game.Marquise {
		t.Fatalf("expected host game start perspective to be Marquise, got %v", hostStart.State.PlayerFaction)
	}

	joinStart := waitForGameStart(t, joinConn, func(msg GameStartMessage) bool {
		return msg.GameID == startResp.GameID && msg.Revision == startResp.Revision
	})
	if joinStart.State.PlayerFaction != game.Eyrie {
		t.Fatalf("expected joiner game start perspective to be Eyrie, got %v", joinStart.State.PlayerFaction)
	}
}

func TestHandleApplyActionBroadcastsGameStateToConnectedLobbyPlayers(t *testing.T) {
	teardown := resetRealtimeTestState(t)
	defer teardown()

	testServer := httptest.NewServer(NewServer())
	defer testServer.Close()

	var createResp CreateLobbyResponse
	postJSON(t, testServer.URL, "/api/lobby/create", CreateLobbyRequest{
		DisplayName: "Host",
		Factions:    []game.Faction{game.Marquise, game.Eyrie},
	}, "", &createResp)

	var joinResp JoinLobbyResponse
	postJSON(t, testServer.URL, "/api/lobby/join", JoinLobbyRequest{
		JoinCode:    createResp.Lobby.JoinCode,
		DisplayName: "Bird",
	}, "", &joinResp)

	marquise := game.Marquise
	postJSON(t, testServer.URL, "/api/lobby/claim-faction", ClaimFactionRequest{Faction: &marquise}, createResp.PlayerToken, nil)
	eyrie := game.Eyrie
	postJSON(t, testServer.URL, "/api/lobby/claim-faction", ClaimFactionRequest{Faction: &eyrie}, joinResp.PlayerToken, nil)

	ready := true
	postJSON(t, testServer.URL, "/api/lobby/ready", ReadyLobbyRequest{IsReady: &ready}, createResp.PlayerToken, nil)
	postJSON(t, testServer.URL, "/api/lobby/ready", ReadyLobbyRequest{IsReady: &ready}, joinResp.PlayerToken, nil)

	var startResp StartLobbyResponse
	postJSON(t, testServer.URL, "/api/lobby/start", map[string]any{}, createResp.PlayerToken, &startResp)

	hostConn := dialWebSocket(t, testServer.URL, createResp.PlayerToken)
	defer hostConn.Close()
	hostInitial := waitForGameState(t, hostConn, func(msg GameStateMessage) bool {
		return msg.GameID == startResp.GameID && msg.Revision == startResp.Revision
	})
	if hostInitial.State.PlayerFaction != game.Marquise {
		t.Fatalf("expected host game state perspective to be Marquise, got %v", hostInitial.State.PlayerFaction)
	}

	joinConn := dialWebSocket(t, testServer.URL, joinResp.PlayerToken)
	defer joinConn.Close()
	joinInitial := waitForGameState(t, joinConn, func(msg GameStateMessage) bool {
		return msg.GameID == startResp.GameID && msg.Revision == startResp.Revision
	})
	if joinInitial.State.PlayerFaction != game.Eyrie {
		t.Fatalf("expected joiner game state perspective to be Eyrie, got %v", joinInitial.State.PlayerFaction)
	}

	var setupAction game.Action
	for _, action := range engine.ValidActions(startResp.State) {
		if action.Type == game.ActionMarquiseSetup {
			setupAction = action
			break
		}
	}
	if setupAction.Type != game.ActionMarquiseSetup {
		t.Fatalf("expected marquise setup action, got %+v", setupAction)
	}

	var applyResp ApplyActionResponse
	postJSON(t, testServer.URL, "/api/actions/apply", ApplyActionRequest{
		GameID:         startResp.GameID,
		State:          startResp.State,
		Action:         setupAction,
		ClientRevision: startResp.Revision,
	}, createResp.PlayerToken, &applyResp)

	hostUpdate := waitForGameState(t, hostConn, func(msg GameStateMessage) bool {
		return msg.GameID == startResp.GameID && msg.Revision == applyResp.Revision
	})
	if hostUpdate.State.PlayerFaction != game.Marquise {
		t.Fatalf("expected host broadcast perspective to be Marquise, got %v", hostUpdate.State.PlayerFaction)
	}
	if len(hostUpdate.ActionLog) != 1 || hostUpdate.ActionLog[0].ActionType != game.ActionMarquiseSetup {
		t.Fatalf("expected host broadcast to include one marquise setup log entry, got %+v", hostUpdate.ActionLog)
	}

	joinUpdate := waitForGameState(t, joinConn, func(msg GameStateMessage) bool {
		return msg.GameID == startResp.GameID && msg.Revision == applyResp.Revision
	})
	if joinUpdate.State.PlayerFaction != game.Eyrie {
		t.Fatalf("expected joiner broadcast perspective to be Eyrie, got %v", joinUpdate.State.PlayerFaction)
	}
	if len(joinUpdate.ActionLog) != 1 || joinUpdate.ActionLog[0].ActionType != game.ActionMarquiseSetup {
		t.Fatalf("expected joiner broadcast to include one marquise setup log entry, got %+v", joinUpdate.ActionLog)
	}
}

func resetRealtimeTestState(t *testing.T) func() {
	t.Helper()

	previousStore := store
	previousLobbies := lobbies
	previousHub := globalHub
	previousActionLogs := actionLogs
	previousBattleSessions := battleSessions

	store = newOnlineStateStore(t.TempDir())
	lobbies = newLobbyStore()
	globalHub = newHub()
	actionLogs = newActionLogStore()
	battleSessions = newBattleSessionStore()

	return func() {
		store = previousStore
		lobbies = previousLobbies
		globalHub = previousHub
		actionLogs = previousActionLogs
		battleSessions = previousBattleSessions
	}
}

func postJSON(t *testing.T, baseURL string, path string, body any, token string, out any) {
	t.Helper()

	var payload []byte
	var err error
	if body == nil {
		payload = []byte("{}")
	} else {
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal %s request: %v", path, err)
		}
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("failed to build %s request: %v", path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Player-Token", token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s request failed: %v", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from %s, got %d", path, resp.StatusCode)
	}

	if out == nil {
		return
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("failed to decode %s response: %v", path, err)
	}
}

func dialWebSocket(t *testing.T, serverURL string, token string) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/api/ws?token=" + url.QueryEscape(token)
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("failed to dial websocket status=%d err=%v", status, err)
	}
	return conn
}

func waitForLobbyUpdate(t *testing.T, conn *websocket.Conn, match func(Lobby) bool) Lobby {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for attempt := 0; attempt < 12; attempt++ {
		var msg LobbyUpdateMessage
		if !readSocketMessage(t, conn, deadline, socketMessageLobbyUpdate, &msg) {
			continue
		}
		if match == nil || match(msg.Lobby) {
			return msg.Lobby
		}
	}

	t.Fatalf("timed out waiting for matching lobby update")
	return Lobby{}
}

func waitForGameStart(t *testing.T, conn *websocket.Conn, match func(GameStartMessage) bool) GameStartMessage {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for attempt := 0; attempt < 12; attempt++ {
		var msg GameStartMessage
		if !readSocketMessage(t, conn, deadline, socketMessageGameStart, &msg) {
			continue
		}
		if match == nil || match(msg) {
			return msg
		}
	}

	t.Fatalf("timed out waiting for matching game start")
	return GameStartMessage{}
}

func waitForGameState(t *testing.T, conn *websocket.Conn, match func(GameStateMessage) bool) GameStateMessage {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for attempt := 0; attempt < 12; attempt++ {
		var msg GameStateMessage
		if !readSocketMessage(t, conn, deadline, socketMessageGameState, &msg) {
			continue
		}
		if match == nil || match(msg) {
			return msg
		}
	}

	t.Fatalf("timed out waiting for matching game state")
	return GameStateMessage{}
}

func waitForBattlePrompt(t *testing.T, conn *websocket.Conn, match func(BattlePromptMessage) bool) BattlePromptMessage {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for attempt := 0; attempt < 12; attempt++ {
		var msg BattlePromptMessage
		if !readSocketMessage(t, conn, deadline, socketMessageBattlePrompt, &msg) {
			continue
		}
		if match == nil || match(msg) {
			return msg
		}
	}

	t.Fatalf("timed out waiting for matching battle prompt")
	return BattlePromptMessage{}
}

func readSocketMessage(t *testing.T, conn *websocket.Conn, deadline time.Time, expectedType string, out any) bool {
	t.Helper()

	if err := conn.SetReadDeadline(deadline); err != nil {
		t.Fatalf("failed to set websocket read deadline: %v", err)
	}

	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read websocket message: %v", err)
	}

	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("failed to decode websocket envelope: %v payload=%s", err, string(payload))
	}
	if envelope.Type != expectedType {
		return false
	}

	if err := json.Unmarshal(payload, out); err != nil {
		t.Fatalf("failed to decode websocket payload: %v payload=%s", err, string(payload))
	}
	return true
}

func countConnected(lobby Lobby) int {
	connected := 0
	for _, player := range lobby.Players {
		if player.Connected {
			connected++
		}
	}
	return connected
}
