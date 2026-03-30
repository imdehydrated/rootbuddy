package server

import "github.com/imdehydrated/rootbuddy/game"

type ValidActionsResponse struct {
	Actions []game.Action `json:"actions"`
	GameID  string        `json:"gameID,omitempty"`
}

type ApplyActionResponse struct {
	State        game.GameState     `json:"state"`
	EffectResult *game.EffectResult `json:"effectResult,omitempty"`
	GameID       string             `json:"gameID,omitempty"`
}

type ResolveBattleResponse struct {
	Action game.Action `json:"action"`
	GameID string      `json:"gameID,omitempty"`
}

type BattleContextResponse struct {
	BattleContext game.BattleContext `json:"battleContext"`
	GameID        string             `json:"gameID,omitempty"`
}

type SetupResponse struct {
	State  game.GameState `json:"state"`
	GameID string         `json:"gameID,omitempty"`
}

type LoadGameResponse struct {
	State  game.GameState `json:"state"`
	GameID string         `json:"gameID,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
