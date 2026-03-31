package server

import "github.com/imdehydrated/rootbuddy/game"

type ValidActionsResponse struct {
	Actions  []game.Action `json:"actions"`
	GameID   string        `json:"gameID,omitempty"`
	Revision int64         `json:"revision,omitempty"`
}

type ApplyActionResponse struct {
	State        game.GameState     `json:"state"`
	EffectResult *game.EffectResult `json:"effectResult,omitempty"`
	GameID       string             `json:"gameID,omitempty"`
	Revision     int64              `json:"revision,omitempty"`
}

type ResolveBattleResponse struct {
	Action   game.Action `json:"action"`
	GameID   string      `json:"gameID,omitempty"`
	Revision int64       `json:"revision,omitempty"`
}

type BattleContextResponse struct {
	BattleContext game.BattleContext `json:"battleContext"`
	GameID        string             `json:"gameID,omitempty"`
	Revision      int64              `json:"revision,omitempty"`
}

type BattlePromptResponse struct {
	Prompt   *BattlePrompt `json:"prompt,omitempty"`
	GameID   string        `json:"gameID,omitempty"`
	Revision int64         `json:"revision,omitempty"`
}

type SetupResponse struct {
	State    game.GameState `json:"state"`
	GameID   string         `json:"gameID,omitempty"`
	Revision int64          `json:"revision,omitempty"`
}

type LoadGameResponse struct {
	State    game.GameState `json:"state"`
	GameID   string         `json:"gameID,omitempty"`
	Revision int64          `json:"revision,omitempty"`
}

type GameLogResponse struct {
	Entries  []ActionLogEntry `json:"entries"`
	GameID   string           `json:"gameID,omitempty"`
	Revision int64            `json:"revision,omitempty"`
}

type ErrorResponse struct {
	Error    string          `json:"error"`
	GameID   string          `json:"gameID,omitempty"`
	Revision int64           `json:"revision,omitempty"`
	State    *game.GameState `json:"state,omitempty"`
}
