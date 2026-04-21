package server

import "github.com/imdehydrated/rootbuddy/game"

type ValidActionsRequest struct {
	State  game.GameState `json:"state"`
	GameID string         `json:"gameID,omitempty"`
}

type ApplyActionRequest struct {
	State          game.GameState `json:"state"`
	Action         game.Action    `json:"action"`
	GameID         string         `json:"gameID,omitempty"`
	ClientRevision int64          `json:"clientRevision,omitempty"`
}

type ResolveBattleRequest struct {
	State        game.GameState       `json:"state"`
	Action       game.Action          `json:"action"`
	AttackerRoll int                  `json:"attackerRoll"`
	DefenderRoll int                  `json:"defenderRoll"`
	Modifiers    game.BattleModifiers `json:"modifiers"`
	UseModifiers bool                 `json:"useModifiers"`
	GameID       string               `json:"gameID,omitempty"`
}

type BattleContextRequest struct {
	State  game.GameState `json:"state"`
	Action game.Action    `json:"action"`
	GameID string         `json:"gameID,omitempty"`
}

type BattleResponseRequest struct {
	GameID              string `json:"gameID"`
	UseAmbush           *bool  `json:"useAmbush,omitempty"`
	UseDefenderArmorers *bool  `json:"useDefenderArmorers,omitempty"`
	UseSappers          *bool  `json:"useSappers,omitempty"`
	UseCounterAmbush    *bool  `json:"useCounterAmbush,omitempty"`
	UseAttackerArmorers *bool  `json:"useAttackerArmorers,omitempty"`
	UseBrutalTactics    *bool  `json:"useBrutalTactics,omitempty"`
}

type SetupRequest struct {
	GameMode      game.GameMode  `json:"gameMode"`
	PlayerFaction game.Faction   `json:"playerFaction"`
	Factions      []game.Faction `json:"factions"`
	MapID         game.MapID     `json:"mapID"`
	RandomSeed    int64          `json:"randomSeed,omitempty"`
}

type LoadGameRequest struct {
	GameID string `json:"gameID"`
}
