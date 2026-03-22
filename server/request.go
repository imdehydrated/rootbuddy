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
	State         game.GameState      `json:"state"`
	Action        game.Action         `json:"action"`
	AttackerRoll  int                 `json:"attackerRoll"`
	DefenderRoll  int                 `json:"defenderRoll"`
	Modifiers     game.BattleModifiers `json:"modifiers"`
	UseModifiers  bool                `json:"useModifiers"`
}
