package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestHandleResolveBattleMultiplayerRequiresDefenderResponseSession(t *testing.T) {
	teardown := resetRealtimeTestState(t)
	defer teardown()

	gameID, hostToken, _, _, _ := startLobbyBackedGame(t)
	record := seedPendingAmbushBattle(t, gameID)
	visible := redactStateForPlayer(record.State, game.Marquise)

	body, _ := json.Marshal(ResolveBattleRequest{
		GameID:       gameID,
		State:        visible,
		Action:       battleAction(game.Marquise, game.Eyrie),
		AttackerRoll: 1,
		DefenderRoll: 0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/battles/resolve", bytes.NewReader(body))
	req.Header.Set("X-Player-Token", hostToken)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 when defender response is required, got %d body=%s", rec.Code, rec.Body.String())
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != errBattleSessionPendingResponse.Error() {
		t.Fatalf("unexpected resolve battle error: %+v", resp)
	}
	if resp.Revision != record.Revision {
		t.Fatalf("expected revision %d, got %d", record.Revision, resp.Revision)
	}
}

func TestBattleSessionOpenRespondResolveAndBroadcastPrompt(t *testing.T) {
	teardown := resetRealtimeTestState(t)
	defer teardown()

	gameID, hostToken, birdToken, _, _ := startLobbyBackedGame(t)
	record := seedPendingAmbushBattle(t, gameID)
	visible := redactStateForPlayer(record.State, game.Marquise)

	testServer := httptest.NewServer(NewServer())
	defer testServer.Close()

	hostConn := dialWebSocket(t, testServer.URL, hostToken)
	defer hostConn.Close()
	hostInitial := waitForGameState(t, hostConn, func(msg GameStateMessage) bool {
		return msg.GameID == gameID && msg.Revision == record.Revision
	})
	if hostInitial.State.PlayerFaction != game.Marquise {
		t.Fatalf("expected marquise reconnect state, got %v", hostInitial.State.PlayerFaction)
	}

	birdConn := dialWebSocket(t, testServer.URL, birdToken)
	defer birdConn.Close()
	birdInitial := waitForGameState(t, birdConn, func(msg GameStateMessage) bool {
		return msg.GameID == gameID && msg.Revision == record.Revision
	})
	if birdInitial.State.PlayerFaction != game.Eyrie {
		t.Fatalf("expected eyrie reconnect state, got %v", birdInitial.State.PlayerFaction)
	}

	var openResp BattlePromptResponse
	postJSON(t, testServer.URL, "/api/battles/open", BattleContextRequest{
		GameID: gameID,
		State:  visible,
		Action: battleAction(game.Marquise, game.Eyrie),
	}, hostToken, &openResp)

	if openResp.Prompt == nil || openResp.Prompt.Stage != BattlePromptWaitingDefender {
		t.Fatalf("expected attacker waiting prompt, got %+v", openResp.Prompt)
	}

	birdPrompt := waitForBattlePrompt(t, birdConn, func(msg BattlePromptMessage) bool {
		return msg.Prompt != nil && msg.Prompt.GameID == gameID && msg.Prompt.Stage == BattlePromptDefenderTurn
	})
	if !birdPrompt.Prompt.CanUseAmbush {
		t.Fatalf("expected defender prompt to allow ambush, got %+v", birdPrompt.Prompt)
	}
	if birdPrompt.Prompt.BattleContext.CanAttackerCounterAmbush {
		t.Fatalf("expected defender prompt to hide attacker counter-ambush availability, got %+v", birdPrompt.Prompt.BattleContext)
	}

	useAmbush := true
	var respondResp BattlePromptResponse
	postJSON(t, testServer.URL, "/api/battles/respond", BattleResponseRequest{
		GameID:    gameID,
		UseAmbush: &useAmbush,
	}, birdToken, &respondResp)

	if respondResp.Prompt == nil || respondResp.Prompt.Stage != BattlePromptReadyToResolve || !respondResp.Prompt.DefenderAmbush {
		t.Fatalf("expected defender ready-to-resolve prompt with ambush, got %+v", respondResp.Prompt)
	}

	hostPrompt := waitForBattlePrompt(t, hostConn, func(msg BattlePromptMessage) bool {
		return msg.Prompt != nil && msg.Prompt.GameID == gameID && msg.Prompt.Stage == BattlePromptReadyToResolve
	})
	if !hostPrompt.Prompt.DefenderAmbush {
		t.Fatalf("expected attacker prompt to reveal defender ambush choice, got %+v", hostPrompt.Prompt)
	}
	if hostPrompt.Prompt.BattleContext.CanDefenderAmbush {
		t.Fatalf("expected attacker prompt to hide defender ambush capability, got %+v", hostPrompt.Prompt.BattleContext)
	}

	resolveBody, _ := json.Marshal(ResolveBattleRequest{
		GameID:       gameID,
		State:        visible,
		Action:       battleAction(game.Marquise, game.Eyrie),
		AttackerRoll: 1,
		DefenderRoll: 0,
		Modifiers: game.BattleModifiers{
			DefenderAmbush: false,
		},
		UseModifiers: true,
	})
	resolveReq := httptest.NewRequest(http.MethodPost, "/api/battles/resolve", bytes.NewReader(resolveBody))
	resolveReq.Header.Set("X-Player-Token", hostToken)
	resolveRec := httptest.NewRecorder()

	NewServer().ServeHTTP(resolveRec, resolveReq)

	if resolveRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for multiplayer resolve battle, got %d body=%s", resolveRec.Code, resolveRec.Body.String())
	}

	var resolveResp ResolveBattleResponse
	if err := json.Unmarshal(resolveRec.Body.Bytes(), &resolveResp); err != nil {
		t.Fatalf("failed to decode resolve battle response: %v", err)
	}
	if resolveResp.Action.BattleResolution == nil || !resolveResp.Action.BattleResolution.DefenderAmbushed {
		t.Fatalf("expected stored defender ambush to be applied, got %+v", resolveResp.Action)
	}

	applyBody, _ := json.Marshal(ApplyActionRequest{
		GameID:         gameID,
		State:          visible,
		Action:         resolveResp.Action,
		ClientRevision: record.Revision,
	})
	applyReq := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(applyBody))
	applyReq.Header.Set("X-Player-Token", hostToken)
	applyRec := httptest.NewRecorder()

	NewServer().ServeHTTP(applyRec, applyReq)

	if applyRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for battle resolution apply, got %d body=%s", applyRec.Code, applyRec.Body.String())
	}

	if _, ok := battleSessions.get(gameID); ok {
		t.Fatalf("expected battle session to be cleared after applying battle resolution")
	}
}

func TestBattleSessionAttackerResponseUsesServerOwnedRolls(t *testing.T) {
	teardown := resetRealtimeTestState(t)
	defer teardown()

	gameID, hostToken, birdToken, _, _ := startLobbyBackedGame(t)
	record := seedAttackerChoiceBattle(t, gameID)
	visible := redactStateForPlayer(record.State, game.Marquise)

	previousRoller := battleRoller
	battleRoller = func() (int, int, error) { return 3, 2, nil }
	defer func() {
		battleRoller = previousRoller
	}()

	testServer := httptest.NewServer(NewServer())
	defer testServer.Close()

	hostConn := dialWebSocket(t, testServer.URL, hostToken)
	defer hostConn.Close()
	waitForGameState(t, hostConn, func(msg GameStateMessage) bool {
		return msg.GameID == gameID && msg.Revision == record.Revision
	})

	birdConn := dialWebSocket(t, testServer.URL, birdToken)
	defer birdConn.Close()
	waitForGameState(t, birdConn, func(msg GameStateMessage) bool {
		return msg.GameID == gameID && msg.Revision == record.Revision
	})

	var openResp BattlePromptResponse
	postJSON(t, testServer.URL, "/api/battles/open", BattleContextRequest{
		GameID: gameID,
		State:  visible,
		Action: battleAction(game.Marquise, game.Eyrie),
	}, hostToken, &openResp)

	if openResp.Prompt == nil || openResp.Prompt.Stage != BattlePromptAttackerTurn {
		t.Fatalf("expected attacker prompt, got %+v", openResp.Prompt)
	}
	if openResp.Prompt.AttackerRoll != 3 || openResp.Prompt.DefenderRoll != 2 {
		t.Fatalf("expected attacker prompt to show server-owned rolls 3/2 before effect choices, got %+v", openResp.Prompt)
	}
	if !openResp.Prompt.CanUseAttackerArmorers || !openResp.Prompt.CanUseBrutalTactics {
		t.Fatalf("expected attacker prompt options, got %+v", openResp.Prompt)
	}
	sessionAfterOpen, ok := battleSessions.get(gameID)
	if !ok {
		t.Fatalf("expected battle session after open")
	}
	if responder := currentBattleResponder(sessionAfterOpen); responder != game.Marquise {
		t.Fatalf("expected attacker to remain current responder after post-roll prompt, got responder=%v session=%+v", responder, sessionAfterOpen)
	}

	prematureResolveBody, _ := json.Marshal(ResolveBattleRequest{
		GameID: gameID,
		State:  visible,
		Action: battleAction(game.Marquise, game.Eyrie),
	})
	prematureResolveReq := httptest.NewRequest(http.MethodPost, "/api/battles/resolve", bytes.NewReader(prematureResolveBody))
	prematureResolveReq.Header.Set("X-Player-Token", hostToken)
	prematureResolveRec := httptest.NewRecorder()
	NewServer().ServeHTTP(prematureResolveRec, prematureResolveReq)
	if prematureResolveRec.Code != http.StatusConflict {
		t.Fatalf("expected resolve to wait for post-roll attacker response, got %d body=%s", prematureResolveRec.Code, prematureResolveRec.Body.String())
	}

	useArmorers := true
	useBrutalTactics := true
	var respondResp BattlePromptResponse
	postJSON(t, testServer.URL, "/api/battles/respond", BattleResponseRequest{
		GameID:              gameID,
		UseAttackerArmorers: &useArmorers,
		UseBrutalTactics:    &useBrutalTactics,
	}, hostToken, &respondResp)

	if respondResp.Prompt == nil || respondResp.Prompt.Stage != BattlePromptReadyToResolve {
		t.Fatalf("expected ready-to-resolve attacker prompt, got %+v", respondResp.Prompt)
	}
	if !respondResp.Prompt.AttackerUsedArmorers || !respondResp.Prompt.AttackerUsedBrutalTactics {
		t.Fatalf("expected attacker choices to be recorded, got %+v", respondResp.Prompt)
	}

	observerPrompt := waitForBattlePrompt(t, birdConn, func(msg BattlePromptMessage) bool {
		return msg.Prompt != nil && msg.Prompt.GameID == gameID && msg.Prompt.Stage == BattlePromptWaitingAttacker
	})
	if observerPrompt.Prompt.WaitingOnFaction != game.Marquise {
		t.Fatalf("expected observer prompt to wait on attacker, got %+v", observerPrompt.Prompt)
	}

	resolveBody, _ := json.Marshal(ResolveBattleRequest{
		GameID:       gameID,
		State:        visible,
		Action:       battleAction(game.Marquise, game.Eyrie),
		AttackerRoll: 0,
		DefenderRoll: 0,
	})
	resolveReq := httptest.NewRequest(http.MethodPost, "/api/battles/resolve", bytes.NewReader(resolveBody))
	resolveReq.Header.Set("X-Player-Token", hostToken)
	resolveRec := httptest.NewRecorder()

	NewServer().ServeHTTP(resolveRec, resolveReq)

	if resolveRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for attacker-response battle resolve, got %d body=%s", resolveRec.Code, resolveRec.Body.String())
	}

	var resolveResp ResolveBattleResponse
	if err := json.Unmarshal(resolveRec.Body.Bytes(), &resolveResp); err != nil {
		t.Fatalf("failed to decode attacker-response battle resolve: %v", err)
	}
	if resolveResp.Action.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload, got %+v", resolveResp.Action)
	}
	if resolveResp.Action.BattleResolution.AttackerRoll != 3 || resolveResp.Action.BattleResolution.DefenderRoll != 2 {
		t.Fatalf("expected server-owned battle rolls 3/2, got %+v", resolveResp.Action.BattleResolution)
	}
	if !resolveResp.Action.BattleResolution.AttackerUsedArmorers || !resolveResp.Action.BattleResolution.AttackerUsedBrutalTactics {
		t.Fatalf("expected stored attacker choices in resolution, got %+v", resolveResp.Action.BattleResolution)
	}
}

func TestHandleApplyActionMultiplayerRejectsTamperedResolvedBattleAction(t *testing.T) {
	teardown := resetRealtimeTestState(t)
	defer teardown()

	gameID, hostToken, _, _, _ := startLobbyBackedGame(t)
	record := seedAttackerChoiceBattle(t, gameID)
	visible := redactStateForPlayer(record.State, game.Marquise)

	previousRoller := battleRoller
	battleRoller = func() (int, int, error) { return 3, 2, nil }
	defer func() {
		battleRoller = previousRoller
	}()

	testServer := httptest.NewServer(NewServer())
	defer testServer.Close()

	var openResp BattlePromptResponse
	postJSON(t, testServer.URL, "/api/battles/open", BattleContextRequest{
		GameID: gameID,
		State:  visible,
		Action: battleAction(game.Marquise, game.Eyrie),
	}, hostToken, &openResp)

	useArmorers := true
	useBrutalTactics := true
	var respondResp BattlePromptResponse
	postJSON(t, testServer.URL, "/api/battles/respond", BattleResponseRequest{
		GameID:              gameID,
		UseAttackerArmorers: &useArmorers,
		UseBrutalTactics:    &useBrutalTactics,
	}, hostToken, &respondResp)

	resolveBody, _ := json.Marshal(ResolveBattleRequest{
		GameID:       gameID,
		State:        visible,
		Action:       battleAction(game.Marquise, game.Eyrie),
		AttackerRoll: 0,
		DefenderRoll: 0,
	})
	resolveReq := httptest.NewRequest(http.MethodPost, "/api/battles/resolve", bytes.NewReader(resolveBody))
	resolveReq.Header.Set("X-Player-Token", hostToken)
	resolveRec := httptest.NewRecorder()
	NewServer().ServeHTTP(resolveRec, resolveReq)

	if resolveRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for resolve battle, got %d body=%s", resolveRec.Code, resolveRec.Body.String())
	}

	var resolveResp ResolveBattleResponse
	if err := json.Unmarshal(resolveRec.Body.Bytes(), &resolveResp); err != nil {
		t.Fatalf("failed to decode resolve battle response: %v", err)
	}

	tampered := resolveResp.Action
	tampered.BattleResolution.AttackerLosses++

	applyBody, _ := json.Marshal(ApplyActionRequest{
		GameID:         gameID,
		State:          visible,
		Action:         tampered,
		ClientRevision: record.Revision,
	})
	applyReq := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(applyBody))
	applyReq.Header.Set("X-Player-Token", hostToken)
	applyRec := httptest.NewRecorder()
	NewServer().ServeHTTP(applyRec, applyReq)

	if applyRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for tampered resolved battle action, got %d body=%s", applyRec.Code, applyRec.Body.String())
	}

	resp := decodeErrorResponse(t, applyRec)
	if resp.Error != errBattleResolutionMismatch.Error() {
		t.Fatalf("unexpected apply error: %+v", resp)
	}
	if resp.Revision != record.Revision {
		t.Fatalf("expected revision %d, got %d", record.Revision, resp.Revision)
	}
}

func TestBattleSessionApplyBroadcastsPromptClear(t *testing.T) {
	teardown := resetRealtimeTestState(t)
	defer teardown()

	gameID, hostToken, birdToken, _, _ := startLobbyBackedGame(t)
	record := seedPendingAmbushBattle(t, gameID)
	visible := redactStateForPlayer(record.State, game.Marquise)

	testServer := httptest.NewServer(NewServer())
	defer testServer.Close()

	hostConn := dialWebSocket(t, testServer.URL, hostToken)
	defer hostConn.Close()
	waitForGameState(t, hostConn, func(msg GameStateMessage) bool {
		return msg.GameID == gameID && msg.Revision == record.Revision
	})

	birdConn := dialWebSocket(t, testServer.URL, birdToken)
	defer birdConn.Close()
	waitForGameState(t, birdConn, func(msg GameStateMessage) bool {
		return msg.GameID == gameID && msg.Revision == record.Revision
	})

	var openResp BattlePromptResponse
	postJSON(t, testServer.URL, "/api/battles/open", BattleContextRequest{
		GameID: gameID,
		State:  visible,
		Action: battleAction(game.Marquise, game.Eyrie),
	}, hostToken, &openResp)

	useAmbush := true
	var respondResp BattlePromptResponse
	postJSON(t, testServer.URL, "/api/battles/respond", BattleResponseRequest{
		GameID:    gameID,
		UseAmbush: &useAmbush,
	}, birdToken, &respondResp)

	resolveBody, _ := json.Marshal(ResolveBattleRequest{
		GameID:       gameID,
		State:        visible,
		Action:       battleAction(game.Marquise, game.Eyrie),
		AttackerRoll: 1,
		DefenderRoll: 0,
	})
	resolveReq := httptest.NewRequest(http.MethodPost, "/api/battles/resolve", bytes.NewReader(resolveBody))
	resolveReq.Header.Set("X-Player-Token", hostToken)
	resolveRec := httptest.NewRecorder()
	NewServer().ServeHTTP(resolveRec, resolveReq)

	if resolveRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for resolve battle, got %d body=%s", resolveRec.Code, resolveRec.Body.String())
	}

	var resolveResp ResolveBattleResponse
	if err := json.Unmarshal(resolveRec.Body.Bytes(), &resolveResp); err != nil {
		t.Fatalf("failed to decode resolve battle response: %v", err)
	}

	applyBody, _ := json.Marshal(ApplyActionRequest{
		GameID:         gameID,
		State:          visible,
		Action:         resolveResp.Action,
		ClientRevision: record.Revision,
	})
	applyReq := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(applyBody))
	applyReq.Header.Set("X-Player-Token", hostToken)
	applyRec := httptest.NewRecorder()
	NewServer().ServeHTTP(applyRec, applyReq)

	if applyRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for battle apply, got %d body=%s", applyRec.Code, applyRec.Body.String())
	}

	hostClear := waitForBattlePrompt(t, hostConn, func(msg BattlePromptMessage) bool {
		return msg.Prompt == nil
	})
	if hostClear.Prompt != nil {
		t.Fatalf("expected host prompt clear, got %+v", hostClear)
	}

	birdClear := waitForBattlePrompt(t, birdConn, func(msg BattlePromptMessage) bool {
		return msg.Prompt == nil
	})
	if birdClear.Prompt != nil {
		t.Fatalf("expected defender prompt clear, got %+v", birdClear)
	}
}

func TestHandleResolveBattleStaleSessionBroadcastsPromptClear(t *testing.T) {
	teardown := resetRealtimeTestState(t)
	defer teardown()

	gameID, hostToken, birdToken, _, _ := startLobbyBackedGame(t)
	record := seedPendingAmbushBattle(t, gameID)
	visible := redactStateForPlayer(record.State, game.Marquise)

	testServer := httptest.NewServer(NewServer())
	defer testServer.Close()

	hostConn := dialWebSocket(t, testServer.URL, hostToken)
	defer hostConn.Close()
	waitForGameState(t, hostConn, func(msg GameStateMessage) bool {
		return msg.GameID == gameID && msg.Revision == record.Revision
	})

	birdConn := dialWebSocket(t, testServer.URL, birdToken)
	defer birdConn.Close()
	waitForGameState(t, birdConn, func(msg GameStateMessage) bool {
		return msg.GameID == gameID && msg.Revision == record.Revision
	})

	var openResp BattlePromptResponse
	postJSON(t, testServer.URL, "/api/battles/open", BattleContextRequest{
		GameID: gameID,
		State:  visible,
		Action: battleAction(game.Marquise, game.Eyrie),
	}, hostToken, &openResp)

	replaceAuthoritativeState(t, gameID, func(state *game.GameState) {
		state.RoundNumber++
	})

	resolveBody, _ := json.Marshal(ResolveBattleRequest{
		GameID:       gameID,
		State:        visible,
		Action:       battleAction(game.Marquise, game.Eyrie),
		AttackerRoll: 1,
		DefenderRoll: 0,
	})
	resolveReq := httptest.NewRequest(http.MethodPost, "/api/battles/resolve", bytes.NewReader(resolveBody))
	resolveReq.Header.Set("X-Player-Token", hostToken)
	resolveRec := httptest.NewRecorder()
	NewServer().ServeHTTP(resolveRec, resolveReq)

	if resolveRec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for stale battle session, got %d body=%s", resolveRec.Code, resolveRec.Body.String())
	}

	clearPrompt := waitForBattlePrompt(t, birdConn, func(msg BattlePromptMessage) bool {
		return msg.Prompt == nil
	})
	if clearPrompt.Prompt != nil {
		t.Fatalf("expected stale-session prompt clear, got %+v", clearPrompt)
	}
}

func seedPendingAmbushBattle(t *testing.T, gameID string) authoritativeGameRecord {
	t.Helper()

	return replaceAuthoritativeState(t, gameID, func(state *game.GameState) {
		state.GameMode = game.GameModeOnline
		state.TrackAllHands = true
		state.GamePhase = game.LifecyclePlaying
		state.SetupStage = game.SetupStageComplete
		state.RoundNumber = 3
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
			{ID: 12, Name: "Ambush! (Fox)", Suit: game.Fox, Kind: game.AmbushCard},
		}
		state.OtherHandCounts = map[game.Faction]int{
			game.Eyrie: 1,
		}
	})
}

func seedAttackerChoiceBattle(t *testing.T, gameID string) authoritativeGameRecord {
	t.Helper()

	return replaceAuthoritativeState(t, gameID, func(state *game.GameState) {
		state.GameMode = game.GameModeOnline
		state.TrackAllHands = true
		state.GamePhase = game.LifecyclePlaying
		state.SetupStage = game.SetupStageComplete
		state.RoundNumber = 4
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
						game.Marquise: 3,
						game.Eyrie:    2,
					},
				},
			},
		}
		state.PersistentEffects = map[game.Faction][]game.CardID{
			game.Marquise: {1, 5},
		}
		state.Marquise.CardsInHand = nil
		state.Eyrie.CardsInHand = nil
		state.OtherHandCounts = map[game.Faction]int{}
	})
}

func battleAction(attacker game.Faction, defender game.Faction) game.Action {
	return game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       attacker,
			ClearingID:    1,
			TargetFaction: defender,
		},
	}
}
