package server

import (
	"encoding/json"
	"io"
	"net/http"

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
		writeError(w, http.StatusForbidden, &ErrorResponse{
			Error:    "not your turn",
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

	next, effectResult := engine.ApplyActionDetailed(context.state, req.Action)
	record, errResp, status := saveMutationResult(context, next, req.ClientRevision)
	if errResp != nil {
		writeError(w, status, errResp)
		return
	}

	if req.GameID != "" {
		effectResult = redactEffectResultForPlayer(context.state, next, req.Action, effectResult)
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

	var action game.Action
	if req.UseModifiers {
		action = engine.ResolveBattleWithModifiers(
			context.state,
			req.Action,
			req.AttackerRoll,
			req.DefenderRoll,
			req.Modifiers,
		)
	} else {
		action = engine.ResolveBattle(
			context.state,
			req.Action,
			req.AttackerRoll,
			req.DefenderRoll,
		)
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

	writeJSON(w, http.StatusOK, BattleContextResponse{
		BattleContext: engine.BattleContext(context.state, req.Action),
		GameID:        req.GameID,
		Revision:      context.record.Revision,
	})
}

func HandleSetup(w http.ResponseWriter, r *http.Request) {
	var req SetupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	state, err := engine.SetupGame(engine.SetupRequest{
		GameMode:          req.GameMode,
		PlayerFaction:     req.PlayerFaction,
		Factions:          req.Factions,
		MapID:             req.MapID,
		VagabondCharacter: req.VagabondCharacter,
		EyrieLeader:       req.EyrieLeader,
		RandomSeed:        req.RandomSeed,
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
		perspective, multiplayer, errResp, status := multiplayerPerspective(req.GameID, playerTokenFromRequest(r))
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
