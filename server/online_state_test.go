package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestOnlineStateStorePersistsAndReloadsFromDisk(t *testing.T) {
	tempDir := t.TempDir()
	original := newOnlineStateStore(tempDir)
	state := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
		RandomSeed:    17,
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{{ID: 24, Name: "Bird Card"}},
		},
	}

	original.save("persist-test", state)

	reloaded := newOnlineStateStore(tempDir)
	loaded, ok := reloaded.load("persist-test")
	if !ok {
		t.Fatalf("expected persisted online state to reload from disk")
	}
	if loaded.RandomSeed != 17 {
		t.Fatalf("expected persisted state to retain random seed, got %+v", loaded)
	}
	if len(loaded.Marquise.CardsInHand) != 1 || loaded.Marquise.CardsInHand[0].ID != 24 {
		t.Fatalf("expected persisted hand to reload, got %+v", loaded.Marquise.CardsInHand)
	}
}

func TestHandleLoadGameReturnsRedactedOnlineState(t *testing.T) {
	previousStore := store
	store = newOnlineStateStore(t.TempDir())
	defer func() {
		store = previousStore
	}()

	authoritative := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
		TurnOrder:     []game.Faction{game.Marquise, game.Eyrie},
		Deck:          []game.CardID{8, 12},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{{ID: 24, Name: "Bird Card"}},
		},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{{ID: 8, Name: "Birdy Bindle"}},
		},
	}
	store.save("load-test", authoritative)

	body, _ := json.Marshal(LoadGameRequest{GameID: "load-test"})
	req := httptest.NewRequest(http.MethodPost, "/api/game/load", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for load game, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp LoadGameResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode load game response: %v", err)
	}
	if len(resp.State.Marquise.CardsInHand) != 1 {
		t.Fatalf("expected player hand to remain visible, got %+v", resp.State.Marquise.CardsInHand)
	}
	if len(resp.State.Eyrie.CardsInHand) != 0 {
		t.Fatalf("expected non-player hand to be redacted, got %+v", resp.State.Eyrie.CardsInHand)
	}
	if resp.State.OtherHandCounts[game.Eyrie] != 1 {
		t.Fatalf("expected redacted hand count for Eyrie, got %+v", resp.State.OtherHandCounts)
	}
	if len(resp.State.Deck) != 2 || resp.State.Deck[0] != 0 || resp.State.Deck[1] != 0 {
		t.Fatalf("expected deck order to stay hidden on load, got %+v", resp.State.Deck)
	}
}

func TestHandleLoadGameRejectsMissingGameID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/game/load", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing game id, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "gameID is required" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}
