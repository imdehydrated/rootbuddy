package server

import (
	"errors"
	"net/http"
	"strings"
)

func HandleCreateLobby(w http.ResponseWriter, r *http.Request) {
	var req CreateLobbyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.DisplayName == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "displayName is required"})
		return
	}

	lobby, playerToken, err := lobbies.createLobby(req)
	if err != nil {
		writeJSON(w, lobbyErrorStatus(err), ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, CreateLobbyResponse{
		Lobby:       lobby,
		PlayerToken: playerToken,
	})
}

func HandleJoinLobby(w http.ResponseWriter, r *http.Request) {
	var req JoinLobbyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	req.JoinCode = strings.ToUpper(strings.TrimSpace(req.JoinCode))
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.JoinCode == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "joinCode is required"})
		return
	}
	if req.DisplayName == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "displayName is required"})
		return
	}

	lobby, playerToken, err := lobbies.joinLobby(req.JoinCode, req.DisplayName)
	if err != nil {
		writeJSON(w, lobbyErrorStatus(err), ErrorResponse{Error: err.Error()})
		return
	}

	globalHub.broadcastLobbyState(lobby.JoinCode, &lobby)

	writeJSON(w, http.StatusOK, JoinLobbyResponse{
		Lobby:       lobby,
		PlayerToken: playerToken,
	})
}

func HandleLobbyState(w http.ResponseWriter, r *http.Request) {
	token := playerTokenFromRequest(r)
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerTokenRequired.Error()})
		return
	}

	lobby, ok := lobbies.getByToken(token)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerNotFound.Error()})
		return
	}

	writeJSON(w, http.StatusOK, LobbyResponse{Lobby: lobby})
}

func HandleClaimFaction(w http.ResponseWriter, r *http.Request) {
	token := playerTokenFromRequest(r)
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerTokenRequired.Error()})
		return
	}

	var req ClaimFactionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	lobby, err := lobbies.claimFaction(token, req.Faction)
	if err != nil {
		writeJSON(w, lobbyErrorStatus(err), ErrorResponse{Error: err.Error()})
		return
	}

	globalHub.broadcastLobbyState(lobby.JoinCode, &lobby)

	writeJSON(w, http.StatusOK, LobbyResponse{Lobby: lobby})
}

func HandleLobbyReady(w http.ResponseWriter, r *http.Request) {
	token := playerTokenFromRequest(r)
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerTokenRequired.Error()})
		return
	}

	var req ReadyLobbyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	lobby, err := lobbies.setReady(token, req.IsReady)
	if err != nil {
		writeJSON(w, lobbyErrorStatus(err), ErrorResponse{Error: err.Error()})
		return
	}

	globalHub.broadcastLobbyState(lobby.JoinCode, &lobby)

	writeJSON(w, http.StatusOK, LobbyResponse{Lobby: lobby})
}

func HandleStartLobby(w http.ResponseWriter, r *http.Request) {
	token := playerTokenFromRequest(r)
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerTokenRequired.Error()})
		return
	}

	lobby, state, revision, err := lobbies.startLobby(token)
	if err != nil {
		writeJSON(w, lobbyErrorStatus(err), ErrorResponse{Error: err.Error()})
		return
	}

	record, errResp, _ := loadValidatedRecord(lobby.GameID)
	if errResp == nil {
		globalHub.broadcastGameStart(lobby.JoinCode, record.GameID, record.Revision, record.State)
	}

	writeJSON(w, http.StatusOK, StartLobbyResponse{
		Lobby:    lobby,
		State:    state,
		GameID:   lobby.GameID,
		Revision: revision,
	})
}

func HandleLeaveLobby(w http.ResponseWriter, r *http.Request) {
	token := playerTokenFromRequest(r)
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerTokenRequired.Error()})
		return
	}

	currentLobby, _ := lobbies.getByToken(token)

	lobby, closed, err := lobbies.leaveLobby(token)
	if err != nil {
		writeJSON(w, lobbyErrorStatus(err), ErrorResponse{Error: err.Error()})
		return
	}

	if currentLobby.JoinCode != "" {
		globalHub.disconnectPlayer(currentLobby.JoinCode, token)
	}
	if !closed && lobby != nil {
		globalHub.broadcastLobbyState(lobby.JoinCode, lobby)
	}

	writeJSON(w, http.StatusOK, LeaveLobbyResponse{
		Closed: closed,
		Lobby:  lobby,
	})
}

func lobbyErrorStatus(err error) int {
	switch {
	case errors.Is(err, errUnknownJoinCode):
		return http.StatusNotFound
	case errors.Is(err, errPlayerTokenRequired), errors.Is(err, errLobbyNotFound), errors.Is(err, errPlayerNotFound):
		return http.StatusUnauthorized
	case errors.Is(err, errHostOnly):
		return http.StatusForbidden
	case errors.Is(err, errLobbyFull), errors.Is(err, errLobbyNotWaiting), errors.Is(err, errFactionClaimed), errors.Is(err, errLobbyNotReady), errors.Is(err, errCannotLeaveInGame):
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}

func battleSessionErrorStatus(err error) int {
	switch {
	case errors.Is(err, errBattleSessionNotFound):
		return http.StatusNotFound
	case errors.Is(err, errBattleSessionPendingResponse), errors.Is(err, errBattleSessionStale), errors.Is(err, errBattleResponseAlreadyProvided):
		return http.StatusConflict
	case errors.Is(err, errBattleResponseForbidden):
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}
