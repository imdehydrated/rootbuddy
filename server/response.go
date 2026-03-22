package server

import "github.com/imdehydrated/rootbuddy/game"

type ValidActionsResponse struct {
	Actions []game.Action `json:"actions"`
}

type ApplyActionResponse struct {
	State game.GameState `json:"state"`
}

type ResolveBattleResponse struct {
	Action game.Action `json:"action"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
