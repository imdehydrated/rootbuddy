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

	writeJSON(w, http.StatusOK, ApplyActionResponse{
		State: engine.ApplyAction(req.State, req.Action),
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

	var action game.Action
	if req.UseModifiers {
		action = engine.ResolveBattleWithModifiers(
			req.State,
			req.Action,
			req.AttackerRoll,
			req.DefenderRoll,
			req.Modifiers,
		)
	} else {
		action = engine.ResolveBattle(
			req.State,
			req.Action,
			req.AttackerRoll,
			req.DefenderRoll,
		)
	}

	writeJSON(w, http.StatusOK, ResolveBattleResponse{
		Action: action,
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
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SetupResponse{State: state})
}
