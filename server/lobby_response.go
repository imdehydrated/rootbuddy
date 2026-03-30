package server

import "github.com/imdehydrated/rootbuddy/game"

type LobbyResponse struct {
	Lobby Lobby `json:"lobby"`
}

type CreateLobbyResponse struct {
	Lobby       Lobby  `json:"lobby"`
	PlayerToken string `json:"playerToken"`
}

type JoinLobbyResponse struct {
	Lobby       Lobby  `json:"lobby"`
	PlayerToken string `json:"playerToken"`
}

type StartLobbyResponse struct {
	Lobby    Lobby          `json:"lobby"`
	State    game.GameState `json:"state"`
	GameID   string         `json:"gameID"`
	Revision int64          `json:"revision,omitempty"`
}

type LeaveLobbyResponse struct {
	Closed bool   `json:"closed"`
	Lobby  *Lobby `json:"lobby,omitempty"`
}
