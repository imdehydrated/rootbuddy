package server

import "github.com/imdehydrated/rootbuddy/game"

type ValidActionsRequest struct {
	State game.GameState `json:"state"`
}

type ApplyActionRequest struct {
	State  game.GameState `json:"state"`
	Action game.Action    `json:"action"`
}

type ResolveBattleRequest struct {
	State        game.GameState       `json:"state"`
	Action       game.Action          `json:"action"`
	AttackerRoll int                  `json:"attackerRoll"`
	DefenderRoll int                  `json:"defenderRoll"`
	Modifiers    game.BattleModifiers `json:"modifiers"`
	UseModifiers bool                 `json:"useModifiers"`
}

type SetupRequest struct {
	GameMode          game.GameMode          `json:"gameMode"`
	PlayerFaction     game.Faction           `json:"playerFaction"`
	Factions          []game.Faction         `json:"factions"`
	MapID             game.MapID             `json:"mapID"`
	VagabondCharacter game.VagabondCharacter `json:"vagabondCharacter"`
	EyrieLeader       game.EyrieLeader       `json:"eyrieLeader"`
}
