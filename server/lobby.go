package server

import (
	"time"

	"github.com/imdehydrated/rootbuddy/game"
)

type LobbyState int

const (
	LobbyWaiting LobbyState = iota
	LobbyInGame
	LobbyClosed
)

type PlayerSlot struct {
	PlayerToken string       `json:"-"`
	DisplayName string       `json:"displayName"`
	Faction     game.Faction `json:"faction"`
	HasFaction  bool         `json:"hasFaction"`
	IsHost      bool         `json:"isHost"`
	IsReady     bool         `json:"isReady"`
	Connected   bool         `json:"connected"`
}

type Lobby struct {
	JoinCode          string                 `json:"joinCode"`
	GameID            string                 `json:"gameID,omitempty"`
	State             LobbyState             `json:"state"`
	HostToken         string                 `json:"-"`
	Players           []PlayerSlot           `json:"players"`
	Factions          []game.Faction         `json:"factions"`
	MapID             game.MapID             `json:"mapID"`
	VagabondCharacter game.VagabondCharacter `json:"vagabondCharacter"`
	EyrieLeader       game.EyrieLeader       `json:"eyrieLeader"`
	CreatedAt         time.Time              `json:"createdAt"`
}

var defaultLobbyFactions = []game.Faction{
	game.Marquise,
	game.Eyrie,
	game.Alliance,
	game.Vagabond,
}

func cloneLobby(lobby *Lobby) Lobby {
	if lobby == nil {
		return Lobby{}
	}

	cloned := *lobby
	cloned.Players = append([]PlayerSlot(nil), lobby.Players...)
	cloned.Factions = append([]game.Faction(nil), lobby.Factions...)
	return cloned
}

func (l Lobby) playerIndex(token string) int {
	for index := range l.Players {
		if l.Players[index].PlayerToken == token {
			return index
		}
	}
	return -1
}

func (l Lobby) claimedFaction(token string) (game.Faction, bool) {
	index := l.playerIndex(token)
	if index == -1 || !l.Players[index].HasFaction {
		return 0, false
	}
	return l.Players[index].Faction, true
}

func (l Lobby) claimedFactions() []game.Faction {
	claimed := make([]game.Faction, 0, len(l.Players))
	for _, faction := range l.Factions {
		for _, player := range l.Players {
			if player.HasFaction && player.Faction == faction {
				claimed = append(claimed, faction)
				break
			}
		}
	}
	return claimed
}

func (l *Lobby) setHost(token string) {
	l.HostToken = token
	for index := range l.Players {
		l.Players[index].IsHost = l.Players[index].PlayerToken == token
	}
}
