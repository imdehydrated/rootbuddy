package server

import "github.com/imdehydrated/rootbuddy/game"

const (
	socketMessageLobbyUpdate  = "lobby.update"
	socketMessageGameStart    = "game.start"
	socketMessageGameState    = "game.state"
	socketMessageBattlePrompt = "battle.prompt"
	socketMessageConflict     = "conflict"
	socketMessageSessionError = "session.error"
)

type LobbyUpdateMessage struct {
	Type  string `json:"type"`
	Lobby Lobby  `json:"lobby"`
}

type GameStartMessage struct {
	Type      string           `json:"type"`
	GameID    string           `json:"gameID"`
	Revision  int64            `json:"revision"`
	State     game.GameState   `json:"state"`
	ActionLog []ActionLogEntry `json:"actionLog,omitempty"`
}

type GameStateMessage struct {
	Type      string           `json:"type"`
	GameID    string           `json:"gameID"`
	Revision  int64            `json:"revision"`
	State     game.GameState   `json:"state"`
	ActionLog []ActionLogEntry `json:"actionLog,omitempty"`
}

type BattlePromptMessage struct {
	Type   string        `json:"type"`
	Prompt *BattlePrompt `json:"prompt,omitempty"`
}

type ConflictMessage struct {
	Type     string         `json:"type"`
	GameID   string         `json:"gameID"`
	Revision int64          `json:"revision"`
	State    game.GameState `json:"state"`
	Error    string         `json:"error"`
}

type SessionErrorMessage struct {
	Type   string `json:"type"`
	GameID string `json:"gameID,omitempty"`
	Error  string `json:"error"`
}
