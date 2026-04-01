package server

import "github.com/imdehydrated/rootbuddy/game"

type LobbyResponse struct {
	Lobby Lobby      `json:"lobby"`
	Self  PlayerSlot `json:"self"`
}

type CreateLobbyResponse struct {
	Lobby       Lobby      `json:"lobby"`
	Self        PlayerSlot `json:"self"`
	PlayerToken string     `json:"playerToken"`
}

type JoinLobbyResponse struct {
	Lobby       Lobby      `json:"lobby"`
	Self        PlayerSlot `json:"self"`
	PlayerToken string     `json:"playerToken"`
}

type StartLobbyResponse struct {
	Lobby    Lobby          `json:"lobby"`
	Self     PlayerSlot     `json:"self"`
	State    game.GameState `json:"state"`
	GameID   string         `json:"gameID"`
	Revision int64          `json:"revision,omitempty"`
}

type LeaveLobbyResponse struct {
	Closed bool        `json:"closed"`
	Lobby  *Lobby      `json:"lobby,omitempty"`
	Self   *PlayerSlot `json:"self,omitempty"`
}
