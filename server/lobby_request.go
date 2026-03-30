package server

import "github.com/imdehydrated/rootbuddy/game"

type CreateLobbyRequest struct {
	DisplayName       string                 `json:"displayName"`
	Factions          []game.Faction         `json:"factions,omitempty"`
	MapID             game.MapID             `json:"mapID,omitempty"`
	VagabondCharacter game.VagabondCharacter `json:"vagabondCharacter,omitempty"`
	EyrieLeader       game.EyrieLeader       `json:"eyrieLeader,omitempty"`
}

type JoinLobbyRequest struct {
	JoinCode    string `json:"joinCode"`
	DisplayName string `json:"displayName"`
}

type ClaimFactionRequest struct {
	Faction *game.Faction `json:"faction"`
}

type ReadyLobbyRequest struct {
	IsReady *bool `json:"isReady,omitempty"`
}
