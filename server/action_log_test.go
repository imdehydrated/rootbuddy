package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

func TestHandleGameLogReturnsEntriesForLobbyBackedGame(t *testing.T) {
	teardown := resetRealtimeTestState(t)
	defer teardown()

	gameID, hostToken, _, hostState, revision := startLobbyBackedGame(t)

	var setupAction game.Action
	for _, action := range validActionsForState(t, hostState) {
		if action.Type == game.ActionMarquiseSetup {
			setupAction = action
			break
		}
	}
	if setupAction.Type != game.ActionMarquiseSetup {
		t.Fatalf("expected marquise setup action, got %+v", setupAction)
	}

	applyBody, _ := json.Marshal(ApplyActionRequest{
		GameID:         gameID,
		State:          hostState,
		Action:         setupAction,
		ClientRevision: revision,
	})
	applyReq := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(applyBody))
	applyReq.Header.Set("X-Player-Token", hostToken)
	applyRec := httptest.NewRecorder()

	NewServer().ServeHTTP(applyRec, applyReq)

	if applyRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from apply action, got %d body=%s", applyRec.Code, applyRec.Body.String())
	}

	logReq := httptest.NewRequest(http.MethodGet, "/api/game/log?gameID="+gameID, nil)
	logReq.Header.Set("X-Player-Token", hostToken)
	logRec := httptest.NewRecorder()

	NewServer().ServeHTTP(logRec, logReq)

	if logRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from game log, got %d body=%s", logRec.Code, logRec.Body.String())
	}

	var logResp GameLogResponse
	if err := json.Unmarshal(logRec.Body.Bytes(), &logResp); err != nil {
		t.Fatalf("failed to decode game log response: %v", err)
	}
	if logResp.GameID != gameID {
		t.Fatalf("expected gameID %s, got %s", gameID, logResp.GameID)
	}
	if len(logResp.Entries) != 1 {
		t.Fatalf("expected one log entry, got %+v", logResp.Entries)
	}
	if logResp.Entries[0].Faction != game.Marquise {
		t.Fatalf("expected marquise log faction, got %+v", logResp.Entries[0])
	}
	if logResp.Entries[0].ActionType != game.ActionMarquiseSetup {
		t.Fatalf("expected marquise setup log type, got %+v", logResp.Entries[0])
	}
	if logResp.Entries[0].Summary == "" {
		t.Fatalf("expected non-empty summary, got %+v", logResp.Entries[0])
	}
}

func TestHandleApplyActionClosesLobbyWhenGameEnds(t *testing.T) {
	teardown := resetRealtimeTestState(t)
	defer teardown()

	gameID, _, birdToken, _, _ := startLobbyBackedGame(t)
	record := replaceAuthoritativeState(t, gameID, func(state *game.GameState) {
		state.GameMode = game.GameModeOnline
		state.TrackAllHands = true
		state.GamePhase = game.LifecyclePlaying
		state.SetupStage = game.SetupStageComplete
		state.RoundNumber = 5
		state.FactionTurn = game.Eyrie
		state.CurrentPhase = game.Evening
		state.CurrentStep = game.StepEvening
		state.TurnOrder = []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond}
		state.VictoryPoints = map[game.Faction]int{
			game.Eyrie: 29,
		}
		state.Eyrie.RoostsPlaced = 2
	})

	visible := redactStateForPlayer(record.State, game.Eyrie)
	body, _ := json.Marshal(ApplyActionRequest{
		GameID:         gameID,
		State:          visible,
		ClientRevision: record.Revision,
		Action: game.Action{
			Type: game.ActionScoreRoosts,
			ScoreRoosts: &game.ScoreRoostsAction{
				Faction: game.Eyrie,
				Points:  1,
			},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(body))
	req.Header.Set("X-Player-Token", birdToken)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from winning apply action, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp ApplyActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode winning apply response: %v", err)
	}
	if resp.State.GamePhase != game.LifecycleGameOver || resp.State.Winner != game.Eyrie {
		t.Fatalf("expected winning state after roost scoring, got %+v", resp.State)
	}

	lobby, ok := lobbies.getByGameID(gameID)
	if !ok {
		t.Fatalf("expected lobby to remain addressable by game ID after closure")
	}
	if lobby.State != LobbyClosed {
		t.Fatalf("expected lobby state to be closed after game over, got %+v", lobby)
	}

	logResp := actionLogs.get(gameID)
	if len(logResp) != 1 || logResp[0].ActionType != game.ActionScoreRoosts {
		t.Fatalf("expected score-roosts log entry after game over, got %+v", logResp)
	}
}

func validActionsForState(t *testing.T, state game.GameState) []game.Action {
	t.Helper()

	actions := engine.ValidActions(state)
	return actions
}
