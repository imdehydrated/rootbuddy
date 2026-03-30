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

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleValidActions(w http.ResponseWriter, r *http.Request) {
	var req ValidActionsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	writeJSON(w, http.StatusOK, ValidActionsResponse{
		Actions: engine.ValidActions(req.State),
		GameID:  req.GameID,
	})
}

func HandleApplyAction(w http.ResponseWriter, r *http.Request) {
	var req ApplyActionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if validationError := validateApplyActionRequest(req); validationError != "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: validationError})
		return
	}

	state := req.State
	perspective := state.PlayerFaction
	if req.GameID != "" {
		authoritative, ok := store.load(req.GameID)
		if !ok {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "unknown game id"})
			return
		}
		state = mergeAuthoritativeState(req.State, authoritative)

		if lobby, ok := lobbies.getByGameID(req.GameID); ok {
			token := playerTokenFromRequest(r)
			if token == "" {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerTokenRequired.Error()})
				return
			}

			playerFaction, claimed := lobby.claimedFaction(token)
			if !claimed {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerNotFound.Error()})
				return
			}
			if playerFaction != authoritative.FactionTurn {
				writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "not your turn"})
				return
			}

			perspective = playerFaction
			state.PlayerFaction = playerFaction
		}
	}

	next, effectResult := engine.ApplyActionDetailed(state, req.Action)
	if req.GameID != "" {
		store.save(req.GameID, next)
		effectResult = redactEffectResultForPlayer(state, next, req.Action, effectResult)
		next = redactStateForPlayer(next, perspective)
	}

	writeJSON(w, http.StatusOK, ApplyActionResponse{
		State:        next,
		EffectResult: effectResult,
		GameID:       req.GameID,
	})
}

func HandleResolveBattle(w http.ResponseWriter, r *http.Request) {
	var req ResolveBattleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if validationError := validateResolveBattleRequest(req); validationError != "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: validationError})
		return
	}

	state := req.State
	if req.GameID != "" {
		authoritative, ok := store.load(req.GameID)
		if !ok {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "unknown game id"})
			return
		}
		state = mergeAuthoritativeState(req.State, authoritative)
	}

	var action game.Action
	if req.UseModifiers {
		action = engine.ResolveBattleWithModifiers(
			state,
			req.Action,
			req.AttackerRoll,
			req.DefenderRoll,
			req.Modifiers,
		)
	} else {
		action = engine.ResolveBattle(
			state,
			req.Action,
			req.AttackerRoll,
			req.DefenderRoll,
		)
	}

	writeJSON(w, http.StatusOK, ResolveBattleResponse{
		Action: action,
		GameID: req.GameID,
	})
}

func HandleBattleContext(w http.ResponseWriter, r *http.Request) {
	var req BattleContextRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if validationError := validateBattleContextRequest(req); validationError != "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: validationError})
		return
	}

	state := req.State
	if req.GameID != "" {
		authoritative, ok := store.load(req.GameID)
		if !ok {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "unknown game id"})
			return
		}
		state = mergeAuthoritativeState(req.State, authoritative)
	}

	writeJSON(w, http.StatusOK, BattleContextResponse{
		BattleContext: engine.BattleContext(state, req.Action),
		GameID:        req.GameID,
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

	gameID := ""
	if state.GameMode == game.GameModeOnline {
		gameID = newGameID()
		authoritative := engine.CloneState(state)
		authoritative.TrackAllHands = true
		store.save(gameID, authoritative)
		state = redactStateForPlayer(authoritative, req.PlayerFaction)
	}

	writeJSON(w, http.StatusOK, SetupResponse{
		State:  state,
		GameID: gameID,
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

	state, ok := store.load(req.GameID)
	if !ok {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "unknown game id"})
		return
	}

	if state.GameMode == game.GameModeOnline {
		perspective := state.PlayerFaction
		if lobby, ok := lobbies.getByGameID(req.GameID); ok {
			token := playerTokenFromRequest(r)
			if token == "" {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerTokenRequired.Error()})
				return
			}

			playerFaction, claimed := lobby.claimedFaction(token)
			if !claimed {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerNotFound.Error()})
				return
			}
			perspective = playerFaction
		}

		state = redactStateForPlayer(state, perspective)
	}

	writeJSON(w, http.StatusOK, LoadGameResponse{
		State:  state,
		GameID: req.GameID,
	})
}
