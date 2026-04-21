package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func decodeJSON(r *http.Request, value any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(value); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return io.ErrUnexpectedEOF
	}

	return nil
}

func writeError(w http.ResponseWriter, status int, resp *ErrorResponse) {
	if resp == nil {
		resp = &ErrorResponse{Error: "request failed"}
	}
	writeJSON(w, status, *resp)
}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleGameLog(w http.ResponseWriter, r *http.Request) {
	gameID := strings.TrimSpace(r.URL.Query().Get("gameID"))
	if gameID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "gameID is required"})
		return
	}

	record, errResp, status := loadValidatedRecord(gameID)
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}

	if record.RequiresLobby {
		if _, _, errResp, status := multiplayerPerspective(record, playerTokenFromRequest(r)); errResp != nil {
			writeError(w, status, errResp)
			return
		}
	}

	writeJSON(w, http.StatusOK, GameLogResponse{
		Entries:  actionLogs.get(gameID),
		GameID:   gameID,
		Revision: record.Revision,
	})
}

func HandleValidActions(w http.ResponseWriter, r *http.Request) {
	var req ValidActionsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	context, errResp, status := buildReadContext(req.GameID, req.State, playerTokenFromRequest(r))
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}
	if req.GameID == "" {
		if errResp := validateClientState(context.state); errResp != nil {
			writeError(w, http.StatusBadRequest, errResp)
			return
		}
	}
	if context.multiplayer && context.perspective != context.record.State.FactionTurn {
		writeJSON(w, http.StatusOK, ValidActionsResponse{
			Actions:  []game.Action{},
			GameID:   req.GameID,
			Revision: context.record.Revision,
		})
		return
	}

	writeJSON(w, http.StatusOK, ValidActionsResponse{
		Actions:  engine.ValidActions(context.state),
		GameID:   req.GameID,
		Revision: context.record.Revision,
	})
}

func HandleApplyAction(w http.ResponseWriter, r *http.Request) {
	var req ApplyActionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	context, errResp, status := buildApplyContext(req, playerTokenFromRequest(r))
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}
	if req.GameID == "" {
		if errResp := validateClientState(context.state); errResp != nil {
			writeError(w, http.StatusBadRequest, errResp)
			return
		}
	}

	effectiveReq := req
	effectiveReq.State = context.state
	if validationError := validateApplyActionRequest(effectiveReq); validationError != "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: validationError})
		return
	}
	if req.GameID != "" && context.multiplayer {
		switch req.Action.Type {
		case game.ActionBattleResolution:
			resolvedAction, err := battleSessions.resolvedActionForApply(req.GameID, context.record.Revision, req.Action)
			if err != nil {
				writeJSON(w, battleSessionErrorStatus(err), ErrorResponse{
					Error:    err.Error(),
					GameID:   req.GameID,
					Revision: context.record.Revision,
				})
				return
			}
			effectiveReq.Action = resolvedAction
		default:
			if !actionMatchesValidAction(context.state, req.Action) {
				writeError(w, http.StatusBadRequest, &ErrorResponse{
					Error:    errInvalidActionForState.Error(),
					GameID:   req.GameID,
					Revision: context.record.Revision,
				})
				return
			}
		}
	}

	next, effectResult := engine.ApplyActionDetailed(context.state, effectiveReq.Action)
	record, errResp, status := saveMutationResult(context, next, req.ClientRevision)
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}
	clearedBattlePrompt := false
	if req.GameID != "" {
		clearedBattlePrompt = battleSessions.clearIfPresent(req.GameID)
	}
	if req.GameID != "" && context.multiplayer {
		actionLogs.append(req.GameID, newActionLogEntry(context.state.RoundNumber, context.perspective, effectiveReq.Action))
	}
	var lobby Lobby
	hasLobby := false
	if req.GameID != "" {
		lobby, hasLobby = lobbies.getByGameID(req.GameID)
		if record.State.GamePhase == game.LifecycleGameOver && hasLobby {
			if closedLobby, _, err := lobbies.closeGameLobby(req.GameID); err == nil {
				lobby = closedLobby
			}
		}
		if hasLobby {
			if clearedBattlePrompt {
				globalHub.broadcastBattlePrompt(lobby.JoinCode, nil)
			}
			globalHub.broadcastGameState(lobby.JoinCode, record.GameID, record.Revision, record.State)
			if record.State.GamePhase == game.LifecycleGameOver {
				globalHub.broadcastLobbyState(lobby.JoinCode, &lobby)
			}
		}
	}

	if req.GameID != "" {
		effectResult = redactEffectResultForPlayer(context.state, next, effectiveReq.Action, effectResult)
		next = redactStateForPlayer(record.State, context.perspective)
	}
	if req.GameID == "" {
		if err := engine.ValidateState(next); err != nil {
			writeError(w, http.StatusInternalServerError, &ErrorResponse{Error: "post-mutation state is invalid"})
			return
		}
	}

	writeJSON(w, http.StatusOK, ApplyActionResponse{
		State:        next,
		EffectResult: effectResult,
		GameID:       req.GameID,
		Revision:     record.Revision,
	})
}

func HandleResolveBattle(w http.ResponseWriter, r *http.Request) {
	var req ResolveBattleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	context, errResp, status := buildReadContext(req.GameID, req.State, playerTokenFromRequest(r))
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}
	if req.GameID == "" {
		if errResp := validateClientState(context.state); errResp != nil {
			writeError(w, http.StatusBadRequest, errResp)
			return
		}
	}
	if context.multiplayer && context.perspective != context.record.State.FactionTurn {
		writeError(w, http.StatusForbidden, &ErrorResponse{
			Error:    "not your turn",
			GameID:   req.GameID,
			Revision: context.record.Revision,
		})
		return
	}

	effectiveReq := req
	effectiveReq.State = context.state
	if validationError := validateResolveBattleRequest(effectiveReq); validationError != "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: validationError})
		return
	}
	if context.multiplayer && !actionMatchesValidAction(context.state, req.Action) {
		writeError(w, http.StatusBadRequest, &ErrorResponse{
			Error:    errInvalidActionForState.Error(),
			GameID:   req.GameID,
			Revision: context.record.Revision,
		})
		return
	}

	modifiers := req.Modifiers
	attackerRoll := req.AttackerRoll
	defenderRoll := req.DefenderRoll
	var battleContext game.BattleContext
	if context.multiplayer {
		battleContext = engine.BattleContext(context.state, req.Action)
		if requiresDefenderResponse(battleContext) || requiresAttackerBaseResponse(battleContext) {
			session, ok := battleSessions.get(req.GameID)
			if !ok {
				writeJSON(w, http.StatusConflict, pendingBattleSessionError(req.GameID, context.record.Revision, battleContext))
				return
			}
			if session.Revision != context.record.Revision || !battleActionsMatch(session.Action, req.Action) {
				battleSessions.clear(req.GameID)
				if lobby, ok := lobbies.getByGameID(req.GameID); ok {
					globalHub.broadcastBattlePrompt(lobby.JoinCode, nil)
				}
				writeJSON(w, http.StatusConflict, ErrorResponse{
					Error:    errBattleSessionStale.Error(),
					GameID:   req.GameID,
					Revision: context.record.Revision,
				})
				return
			}
			if responder := currentBattleResponder(session); responder != 0 {
				errResp := ErrorResponse{
					Error:    errBattleSessionPendingResponse.Error(),
					GameID:   req.GameID,
					Revision: context.record.Revision,
				}
				if responder == session.AttackerFaction {
					errResp.Error = errBattleSessionPendingAttacker.Error()
				}
				writeJSON(w, http.StatusConflict, errResp)
				return
			}
			modifiers.DefenderAmbush = session.DefenderAmbush
			modifiers.DefenderUsesArmorers = session.DefenderUsedArmorers
			modifiers.DefenderUsesSappers = session.DefenderUsedSappers
			modifiers.AttackerCounterAmbush = session.AttackerCounterAmbush
			modifiers.AttackerUsesArmorers = session.AttackerUsedArmorers
			modifiers.AttackerUsesBrutalTactics = session.AttackerUsedBrutalTactics
		}
		rolledAttacker, rolledDefender, err := battleRoller()
		if err != nil {
			writeError(w, http.StatusInternalServerError, &ErrorResponse{
				Error:    "failed to roll battle dice",
				GameID:   req.GameID,
				Revision: context.record.Revision,
			})
			return
		}
		attackerRoll = rolledAttacker
		defenderRoll = rolledDefender
	}

	useModifiers := req.UseModifiers
	if context.multiplayer && (modifiers.DefenderAmbush || modifiers.DefenderUsesArmorers || modifiers.DefenderUsesSappers || modifiers.AttackerCounterAmbush || modifiers.AttackerUsesArmorers || modifiers.AttackerUsesBrutalTactics) {
		useModifiers = true
	}

	var action game.Action
	if useModifiers {
		action = engine.ResolveBattleWithModifiers(
			context.state,
			req.Action,
			attackerRoll,
			defenderRoll,
			modifiers,
		)
	} else {
		action = engine.ResolveBattle(
			context.state,
			req.Action,
			attackerRoll,
			defenderRoll,
		)
	}
	if context.multiplayer {
		if _, err := battleSessions.recordResolution(req.GameID, context.record.Revision, req.Action, battleContext, action); err != nil {
			writeJSON(w, battleSessionErrorStatus(err), ErrorResponse{
				Error:    err.Error(),
				GameID:   req.GameID,
				Revision: context.record.Revision,
			})
			return
		}
	}

	writeJSON(w, http.StatusOK, ResolveBattleResponse{
		Action:   action,
		GameID:   req.GameID,
		Revision: context.record.Revision,
	})
}

func HandleBattleContext(w http.ResponseWriter, r *http.Request) {
	var req BattleContextRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	context, errResp, status := buildReadContext(req.GameID, req.State, playerTokenFromRequest(r))
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}
	if req.GameID == "" {
		if errResp := validateClientState(context.state); errResp != nil {
			writeError(w, http.StatusBadRequest, errResp)
			return
		}
	}
	if context.multiplayer && context.perspective != context.record.State.FactionTurn {
		writeError(w, http.StatusForbidden, &ErrorResponse{
			Error:    "not your turn",
			GameID:   req.GameID,
			Revision: context.record.Revision,
		})
		return
	}

	effectiveReq := req
	effectiveReq.State = context.state
	if validationError := validateBattleContextRequest(effectiveReq); validationError != "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: validationError})
		return
	}
	if context.multiplayer && !actionMatchesValidAction(context.state, req.Action) {
		writeError(w, http.StatusBadRequest, &ErrorResponse{
			Error:    errInvalidActionForState.Error(),
			GameID:   req.GameID,
			Revision: context.record.Revision,
		})
		return
	}

	writeJSON(w, http.StatusOK, BattleContextResponse{
		BattleContext: engine.BattleContext(context.state, req.Action),
		GameID:        req.GameID,
		Revision:      context.record.Revision,
	})
}

func HandleOpenBattle(w http.ResponseWriter, r *http.Request) {
	var req BattleContextRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	context, errResp, status := buildReadContext(req.GameID, req.State, playerTokenFromRequest(r))
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}
	if req.GameID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "battle opening requires gameID"})
		return
	}
	if !context.multiplayer {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "battle opening is only supported for lobby-backed multiplayer games"})
		return
	}
	if context.perspective != context.record.State.FactionTurn {
		writeError(w, http.StatusForbidden, &ErrorResponse{
			Error:    "not your turn",
			GameID:   req.GameID,
			Revision: context.record.Revision,
		})
		return
	}

	effectiveReq := req
	effectiveReq.State = context.state
	if validationError := validateBattleContextRequest(effectiveReq); validationError != "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: validationError})
		return
	}
	if !actionMatchesValidAction(context.state, req.Action) {
		writeError(w, http.StatusBadRequest, &ErrorResponse{
			Error:    errInvalidActionForState.Error(),
			GameID:   req.GameID,
			Revision: context.record.Revision,
		})
		return
	}

	battleContext := engine.BattleContext(context.state, req.Action)
	if !requiresDefenderResponse(battleContext) && !requiresAttackerBaseResponse(battleContext) {
		writeJSON(w, http.StatusOK, BattlePromptResponse{
			Prompt: battlePromptView(battleSession{
				GameID:          req.GameID,
				Revision:        context.record.Revision,
				Action:          req.Action,
				BattleContext:   battleContext,
				AttackerFaction: req.Action.Battle.Faction,
				DefenderFaction: req.Action.Battle.TargetFaction,
			}, context.perspective),
			GameID:   req.GameID,
			Revision: context.record.Revision,
		})
		return
	}

	session, _ := battleSessions.open(req.GameID, context.record.Revision, req.Action, battleContext)
	if lobby, ok := lobbies.getByGameID(req.GameID); ok {
		globalHub.broadcastBattlePrompt(lobby.JoinCode, &session)
	}

	writeJSON(w, http.StatusOK, BattlePromptResponse{
		Prompt:   battlePromptView(session, context.perspective),
		GameID:   req.GameID,
		Revision: context.record.Revision,
	})
}

func HandleBattleSession(w http.ResponseWriter, r *http.Request) {
	gameID := strings.TrimSpace(r.URL.Query().Get("gameID"))
	if gameID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "gameID is required"})
		return
	}

	record, errResp, status := loadValidatedRecord(gameID)
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}

	perspective, multiplayer, errResp, status := multiplayerPerspective(record, playerTokenFromRequest(r))
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}
	if !multiplayer {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "battle sessions are only supported for lobby-backed multiplayer games"})
		return
	}

	session, ok := battleSessions.get(gameID)
	if !ok || session.Revision != record.Revision {
		if ok && session.Revision != record.Revision {
			battleSessions.clear(gameID)
			if lobby, lobbyOK := lobbies.getByGameID(gameID); lobbyOK {
				globalHub.broadcastBattlePrompt(lobby.JoinCode, nil)
			}
		}
		writeJSON(w, http.StatusOK, BattlePromptResponse{
			GameID:   gameID,
			Revision: record.Revision,
		})
		return
	}

	writeJSON(w, http.StatusOK, BattlePromptResponse{
		Prompt:   battlePromptView(session, perspective),
		GameID:   gameID,
		Revision: record.Revision,
	})
}

func HandleBattleResponse(w http.ResponseWriter, r *http.Request) {
	token := playerTokenFromRequest(r)
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerTokenRequired.Error()})
		return
	}

	var req BattleResponseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.GameID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "gameID is required"})
		return
	}

	record, errResp, status := loadValidatedRecord(req.GameID)
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}
	perspective, multiplayer, errResp, status := multiplayerPerspective(record, token)
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}
	if !multiplayer {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "battle responses are only supported for lobby-backed multiplayer games"})
		return
	}

	session, err := battleSessions.applyResponse(req.GameID, record.Revision, perspective, req)
	if err != nil {
		writeJSON(w, battleSessionErrorStatus(err), ErrorResponse{
			Error:    err.Error(),
			GameID:   req.GameID,
			Revision: record.Revision,
		})
		return
	}

	if lobby, ok := lobbies.getByGameID(req.GameID); ok {
		globalHub.broadcastBattlePrompt(lobby.JoinCode, &session)
	}

	writeJSON(w, http.StatusOK, BattlePromptResponse{
		Prompt:   battlePromptView(session, perspective),
		GameID:   req.GameID,
		Revision: record.Revision,
	})
}

func HandleSetup(w http.ResponseWriter, r *http.Request) {
	var req SetupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	state, err := engine.SetupGame(engine.SetupRequest{
		GameMode:      req.GameMode,
		PlayerFaction: req.PlayerFaction,
		TrackAllHands: req.GameMode == game.GameModeOnline,
		Factions:      req.Factions,
		MapID:         req.MapID,
		RandomSeed:    req.RandomSeed,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if errResp := validateClientState(state); errResp != nil {
		writeError(w, http.StatusInternalServerError, &ErrorResponse{Error: "generated setup state is invalid"})
		return
	}

	gameID := ""
	revision := int64(0)
	if state.GameMode == game.GameModeOnline {
		gameID = newGameID()
		authoritative := engine.CloneState(state)
		authoritative.TrackAllHands = true
		record, err := store.create(gameID, authoritative)
		if err != nil {
			writeError(w, http.StatusInternalServerError, &ErrorResponse{Error: "failed to persist authoritative game state"})
			return
		}
		state = redactStateForPlayer(record.State, req.PlayerFaction)
		revision = record.Revision
	}

	writeJSON(w, http.StatusOK, SetupResponse{
		State:    state,
		GameID:   gameID,
		Revision: revision,
	})
}

func HandleLoadGame(w http.ResponseWriter, r *http.Request) {
	var req LoadGameRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.GameID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "gameID is required"})
		return
	}

	record, errResp, status := loadValidatedRecord(req.GameID)
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}

	state := record.State
	if state.GameMode == game.GameModeOnline {
		perspective, multiplayer, errResp, status := multiplayerPerspective(record, playerTokenFromRequest(r))
		if errResp != nil {
			writeError(w, status, errResp)
			return
		}
		if !multiplayer {
			perspective = state.PlayerFaction
		}

		state = redactStateForPlayer(state, perspective)
	}

	writeJSON(w, http.StatusOK, LoadGameResponse{
		State:    state,
		GameID:   req.GameID,
		Revision: record.Revision,
	})
}
