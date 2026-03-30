package server

import (
	"log"
	"net/http"

	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

type gameRequestContext struct {
	record      authoritativeGameRecord
	state       game.GameState
	perspective game.Faction
	multiplayer bool
}

func validateClientState(state game.GameState) *ErrorResponse {
	if err := engine.ValidateState(state); err != nil {
		return &ErrorResponse{Error: err.Error()}
	}
	return nil
}

func loadValidatedRecord(gameID string) (authoritativeGameRecord, *ErrorResponse, int) {
	record, ok, err := store.load(gameID)
	if err != nil {
		log.Printf("online state load failed gameID=%s err=%v", gameID, err)
		return authoritativeGameRecord{}, &ErrorResponse{Error: "failed to load authoritative game state", GameID: gameID}, http.StatusInternalServerError
	}
	if !ok {
		return authoritativeGameRecord{}, &ErrorResponse{Error: "unknown game id", GameID: gameID}, http.StatusBadRequest
	}
	if err := engine.ValidateState(record.State); err != nil {
		log.Printf("invalid authoritative state gameID=%s revision=%d err=%v", gameID, record.Revision, err)
		return authoritativeGameRecord{}, &ErrorResponse{Error: "authoritative game state is invalid", GameID: gameID, Revision: record.Revision}, http.StatusInternalServerError
	}
	return record, nil, http.StatusOK
}

func multiplayerPerspective(record authoritativeGameRecord, token string) (game.Faction, bool, *ErrorResponse, int) {
	if !record.RequiresLobby {
		return 0, false, nil, http.StatusOK
	}

	gameID := record.GameID
	lobby, ok := lobbies.getByGameID(gameID)
	if !ok {
		return 0, true, &ErrorResponse{
			Error:    errLobbySessionUnavailable.Error(),
			GameID:   gameID,
			Revision: record.Revision,
		}, http.StatusConflict
	}
	if token == "" {
		return 0, true, &ErrorResponse{Error: errPlayerTokenRequired.Error(), GameID: gameID}, http.StatusUnauthorized
	}

	playerFaction, claimed := lobby.claimedFaction(token)
	if !claimed {
		return 0, true, &ErrorResponse{Error: errPlayerNotFound.Error(), GameID: gameID}, http.StatusUnauthorized
	}

	return playerFaction, true, nil, http.StatusOK
}

func multiplayerConflictResponse(record authoritativeGameRecord, perspective game.Faction) *ErrorResponse {
	visible := redactStateForPlayer(record.State, perspective)
	return &ErrorResponse{
		Error:    errRevisionConflict.Error(),
		GameID:   record.GameID,
		Revision: record.Revision,
		State:    &visible,
	}
}

func buildReadContext(gameID string, fallbackState game.GameState, token string) (gameRequestContext, *ErrorResponse, int) {
	if gameID == "" {
		return gameRequestContext{
			state:       fallbackState,
			perspective: fallbackState.PlayerFaction,
		}, nil, http.StatusOK
	}

	record, errResp, status := loadValidatedRecord(gameID)
	if errResp != nil {
		return gameRequestContext{}, errResp, status
	}

	perspective, multiplayer, errResp, status := multiplayerPerspective(record, token)
	if errResp != nil {
		return gameRequestContext{}, errResp, status
	}

	if multiplayer {
		state := engine.CloneState(record.State)
		state.PlayerFaction = perspective
		return gameRequestContext{
			record:      record,
			state:       state,
			perspective: perspective,
			multiplayer: true,
		}, nil, http.StatusOK
	}

	state := mergeAuthoritativeState(fallbackState, record.State)
	if errResp := validateClientState(state); errResp != nil {
		return gameRequestContext{}, errResp, http.StatusBadRequest
	}

	return gameRequestContext{
		record:      record,
		state:       state,
		perspective: state.PlayerFaction,
	}, nil, http.StatusOK
}

func buildApplyContext(req ApplyActionRequest, token string) (gameRequestContext, *ErrorResponse, int) {
	if req.GameID == "" {
		return gameRequestContext{
			state:       req.State,
			perspective: req.State.PlayerFaction,
		}, nil, http.StatusOK
	}

	record, errResp, status := loadValidatedRecord(req.GameID)
	if errResp != nil {
		return gameRequestContext{}, errResp, status
	}

	perspective, multiplayer, errResp, status := multiplayerPerspective(record, token)
	if errResp != nil {
		return gameRequestContext{}, errResp, status
	}

	if multiplayer {
		if req.ClientRevision <= 0 {
			return gameRequestContext{}, &ErrorResponse{Error: "clientRevision is required for multiplayer game actions", GameID: req.GameID}, http.StatusBadRequest
		}
		if req.ClientRevision != record.Revision {
			return gameRequestContext{}, multiplayerConflictResponse(record, perspective), http.StatusConflict
		}
		if perspective != record.State.FactionTurn {
			return gameRequestContext{}, &ErrorResponse{Error: "not your turn", GameID: req.GameID, Revision: record.Revision}, http.StatusForbidden
		}

		state := engine.CloneState(record.State)
		state.PlayerFaction = perspective
		return gameRequestContext{
			record:      record,
			state:       state,
			perspective: perspective,
			multiplayer: true,
		}, nil, http.StatusOK
	}

	state := mergeAuthoritativeState(req.State, record.State)
	if errResp := validateClientState(state); errResp != nil {
		return gameRequestContext{}, errResp, http.StatusBadRequest
	}

	return gameRequestContext{
		record:      record,
		state:       state,
		perspective: state.PlayerFaction,
	}, nil, http.StatusOK
}

func saveMutationResult(ctx gameRequestContext, next game.GameState, clientRevision int64) (authoritativeGameRecord, *ErrorResponse, int) {
	if ctx.record.GameID == "" {
		return authoritativeGameRecord{State: next}, nil, http.StatusOK
	}

	if err := engine.ValidateState(next); err != nil {
		log.Printf("invalid post-mutation state gameID=%s revision=%d err=%v", ctx.record.GameID, ctx.record.Revision, err)
		return authoritativeGameRecord{}, &ErrorResponse{Error: "post-mutation authoritative state is invalid", GameID: ctx.record.GameID, Revision: ctx.record.Revision}, http.StatusInternalServerError
	}

	var (
		record authoritativeGameRecord
		err    error
	)
	if ctx.multiplayer {
		record, err = store.saveIfRevision(ctx.record.GameID, clientRevision, next)
	} else {
		record, err = store.save(ctx.record.GameID, next)
	}
	if err == errRevisionConflict {
		latest, loadErrResp, loadStatus := loadValidatedRecord(ctx.record.GameID)
		if loadErrResp != nil {
			return authoritativeGameRecord{}, loadErrResp, loadStatus
		}
		return authoritativeGameRecord{}, multiplayerConflictResponse(latest, ctx.perspective), http.StatusConflict
	}
	if err != nil {
		log.Printf("online state save failed gameID=%s revision=%d err=%v", ctx.record.GameID, ctx.record.Revision, err)
		return authoritativeGameRecord{}, &ErrorResponse{Error: "failed to persist authoritative game state", GameID: ctx.record.GameID}, http.StatusInternalServerError
	}

	if err := engine.ValidateState(record.State); err != nil {
		log.Printf("invalid saved authoritative state gameID=%s revision=%d err=%v", record.GameID, record.Revision, err)
		return authoritativeGameRecord{}, &ErrorResponse{Error: "saved authoritative game state is invalid", GameID: record.GameID, Revision: record.Revision}, http.StatusInternalServerError
	}

	return record, nil, http.StatusOK
}
