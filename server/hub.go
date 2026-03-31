package server

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/imdehydrated/rootbuddy/game"
)

const wsWriteTimeout = 10 * time.Second

type wsClient struct {
	conn        *websocket.Conn
	playerToken string
	faction     game.Faction
	joinCode    string
	send        chan []byte
	hub         *hub
	closeOnce   sync.Once
}

func (c *wsClient) close() {
	c.closeOnce.Do(func() {
		close(c.send)
		if c.conn != nil {
			_ = c.conn.Close()
		}
	})
}

func (c *wsClient) writePump() {
	defer c.close()

	for payload := range c.send {
		if c.conn == nil {
			continue
		}

		if err := c.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout)); err != nil {
			return
		}
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			return
		}
	}
}

func (c *wsClient) readPump() {
	defer handleClientClosed(c)

	if c.conn == nil {
		return
	}

	c.conn.SetReadLimit(1024)
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

type hub struct {
	mu      sync.RWMutex
	clients map[string]map[string]*wsClient
}

func newHub() *hub {
	return &hub{
		clients: map[string]map[string]*wsClient{},
	}
}

var globalHub = newHub()

func (h *hub) register(client *wsClient) *wsClient {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.joinCode] == nil {
		h.clients[client.joinCode] = map[string]*wsClient{}
	}

	replaced := h.clients[client.joinCode][client.playerToken]
	h.clients[client.joinCode][client.playerToken] = client
	return replaced
}

func (h *hub) disconnectPlayer(joinCode string, token string) {
	if joinCode == "" || token == "" {
		return
	}

	var client *wsClient

	h.mu.Lock()
	lobbyClients, ok := h.clients[joinCode]
	if ok {
		client = lobbyClients[token]
		delete(lobbyClients, token)
		if len(lobbyClients) == 0 {
			delete(h.clients, joinCode)
		}
	}
	h.mu.Unlock()

	if client != nil {
		client.close()
	}
}

func (h *hub) unregister(client *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	lobbyClients, ok := h.clients[client.joinCode]
	if !ok {
		return
	}
	if existing, ok := lobbyClients[client.playerToken]; !ok || existing != client {
		return
	}

	delete(lobbyClients, client.playerToken)
	if len(lobbyClients) == 0 {
		delete(h.clients, client.joinCode)
	}
}

func (h *hub) hasClient(joinCode string, token string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	lobbyClients, ok := h.clients[joinCode]
	if !ok {
		return false
	}
	_, ok = lobbyClients[token]
	return ok
}

func (h *hub) broadcastToLobby(joinCode string, makePayload func(faction game.Faction) []byte) {
	h.mu.RLock()
	lobbyClients := h.clients[joinCode]
	snapshot := make([]*wsClient, 0, len(lobbyClients))
	for _, client := range lobbyClients {
		snapshot = append(snapshot, client)
	}
	h.mu.RUnlock()

	if len(snapshot) == 0 {
		return
	}

	payloads := map[game.Faction][]byte{}
	for _, client := range snapshot {
		payload, ok := payloads[client.faction]
		if !ok {
			payload = makePayload(client.faction)
			payloads[client.faction] = payload
		}
		if payload == nil {
			continue
		}
		h.enqueue(client, payload)
	}
}

func (h *hub) enqueue(client *wsClient, payload []byte) {
	select {
	case client.send <- payload:
	default:
		go handleClientClosed(client)
	}
}

func (h *hub) broadcastLobbyState(joinCode string, lobby *Lobby) {
	if lobby == nil {
		return
	}

	staleClients := h.syncLobbyClients(lobby)

	payload := marshalSocketMessage(LobbyUpdateMessage{
		Type:  socketMessageLobbyUpdate,
		Lobby: cloneLobby(lobby),
	})
	if payload == nil {
		return
	}

	h.mu.RLock()
	lobbyClients := h.clients[joinCode]
	snapshot := make([]*wsClient, 0, len(lobbyClients))
	for _, client := range lobbyClients {
		snapshot = append(snapshot, client)
	}
	h.mu.RUnlock()

	for _, client := range snapshot {
		h.enqueue(client, payload)
	}

	for _, client := range staleClients {
		client.close()
	}
}

func (h *hub) broadcastGameStart(joinCode string, gameID string, revision int64, state game.GameState) {
	actionLog := actionLogs.get(gameID)
	h.broadcastToLobby(joinCode, func(faction game.Faction) []byte {
		return marshalSocketMessage(GameStartMessage{
			Type:      socketMessageGameStart,
			GameID:    gameID,
			Revision:  revision,
			State:     redactStateForPlayer(state, faction),
			ActionLog: actionLog,
		})
	})
}

func (h *hub) broadcastGameState(joinCode string, gameID string, revision int64, state game.GameState) {
	actionLog := actionLogs.get(gameID)
	h.broadcastToLobby(joinCode, func(faction game.Faction) []byte {
		return marshalSocketMessage(GameStateMessage{
			Type:      socketMessageGameState,
			GameID:    gameID,
			Revision:  revision,
			State:     redactStateForPlayer(state, faction),
			ActionLog: actionLog,
		})
	})
}

func (h *hub) broadcastBattlePrompt(joinCode string, session *battleSession) {
	h.broadcastToLobby(joinCode, func(faction game.Faction) []byte {
		var prompt *BattlePrompt
		if session != nil {
			prompt = battlePromptView(*session, faction)
		}
		return marshalSocketMessage(BattlePromptMessage{
			Type:   socketMessageBattlePrompt,
			Prompt: prompt,
		})
	})
}

func marshalSocketMessage(value any) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		log.Printf("failed to marshal websocket message: %v", err)
		return nil
	}
	return payload
}

func (h *hub) syncLobbyClients(lobby *Lobby) []*wsClient {
	if lobby == nil {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	lobbyClients, ok := h.clients[lobby.JoinCode]
	if !ok {
		return nil
	}

	playerByToken := make(map[string]PlayerSlot, len(lobby.Players))
	for _, player := range lobby.Players {
		playerByToken[player.PlayerToken] = player
	}

	staleClients := make([]*wsClient, 0)
	for token, client := range lobbyClients {
		player, exists := playerByToken[token]
		if !exists {
			delete(lobbyClients, token)
			staleClients = append(staleClients, client)
			continue
		}

		if player.HasFaction {
			client.faction = player.Faction
		} else {
			client.faction = 0
		}
	}

	if len(lobbyClients) == 0 {
		delete(h.clients, lobby.JoinCode)
	}

	return staleClients
}

func handleClientClosed(client *wsClient) {
	if client == nil || client.hub == nil {
		return
	}

	client.hub.unregister(client)
	client.close()

	if client.hub.hasClient(client.joinCode, client.playerToken) {
		return
	}

	lobby, changed, err := lobbies.setConnected(client.playerToken, false)
	if err != nil || !changed {
		return
	}

	client.hub.broadcastLobbyState(lobby.JoinCode, &lobby)
}
