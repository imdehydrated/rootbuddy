package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerTokenRequired.Error()})
		return
	}

	lobby, ok := lobbies.getByToken(token)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: errPlayerNotFound.Error()})
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed token=%s err=%v", token, err)
		return
	}

	faction, _ := lobby.claimedFaction(token)
	client := &wsClient{
		conn:        conn,
		playerToken: token,
		faction:     faction,
		joinCode:    lobby.JoinCode,
		send:        make(chan []byte, 8),
		hub:         globalHub,
	}

	replaced := globalHub.register(client)
	if replaced != nil {
		replaced.close()
	}

	go client.writePump()

	sendInitial := true
	updatedLobby, changed, err := lobbies.setConnected(token, true)
	if err == nil && changed {
		globalHub.broadcastLobbyState(updatedLobby.JoinCode, &updatedLobby)
		lobby = updatedLobby
		if lobby.State != LobbyInGame {
			sendInitial = false
		}
	}

	if sendInitial {
		sendInitialSocketState(client, lobby)
	}
	client.readPump()
}

func sendInitialSocketState(client *wsClient, lobby Lobby) {
	if lobby.GameID != "" {
		record, errResp, _ := loadValidatedRecord(lobby.GameID)
		if errResp == nil {
			payload := marshalSocketMessage(GameStateMessage{
				Type:      socketMessageGameState,
				GameID:    record.GameID,
				Revision:  record.Revision,
				State:     redactStateForPlayer(record.State, client.faction),
				ActionLog: actionLogs.get(record.GameID),
			})
			if payload != nil {
				client.hub.enqueue(client, payload)
			}
			if session, ok := battleSessions.get(record.GameID); ok {
				promptPayload := marshalSocketMessage(BattlePromptMessage{
					Type:   socketMessageBattlePrompt,
					Prompt: battlePromptView(session, client.faction),
				})
				if promptPayload != nil {
					client.hub.enqueue(client, promptPayload)
				}
			}
			return
		}

		payload := marshalSocketMessage(SessionErrorMessage{
			Type:   socketMessageSessionError,
			GameID: lobby.GameID,
			Error:  errResp.Error,
		})
		if payload != nil {
			client.hub.enqueue(client, payload)
		}
		return
	}

	payload := marshalSocketMessage(LobbyUpdateMessage{
		Type:  socketMessageLobbyUpdate,
		Lobby: lobby,
	})
	if payload != nil {
		client.hub.enqueue(client, payload)
	}
}
