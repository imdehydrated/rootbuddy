package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imdehydrated/rootbuddy/engine"
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

	if _, err := original.create("persist-test", state); err != nil {
		t.Fatalf("failed to persist online state: %v", err)
	}

	reloaded := newOnlineStateStore(tempDir)
	loaded, ok, err := reloaded.load("persist-test")
	if err != nil {
		t.Fatalf("failed to reload persisted state: %v", err)
	}
	if !ok {
		t.Fatalf("expected persisted online state to reload from disk")
	}
	if loaded.Revision != 1 {
		t.Fatalf("expected initial revision 1, got %+v", loaded)
	}
	if loaded.State.RandomSeed != 17 {
		t.Fatalf("expected persisted state to retain random seed, got %+v", loaded)
	}
	if len(loaded.State.Marquise.CardsInHand) != 1 || loaded.State.Marquise.CardsInHand[0].ID != 24 {
		t.Fatalf("expected persisted hand to reload, got %+v", loaded.State.Marquise.CardsInHand)
	}
}

func TestOnlineStateStoreSaveIfRevisionDetectsStaleWrites(t *testing.T) {
	tempDir := t.TempDir()
	testStore := newOnlineStateStore(tempDir)
	state := game.GameState{
		GameMode:     game.GameModeOnline,
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
	}

	record, err := testStore.create("revision-test", state)
	if err != nil {
		t.Fatalf("failed to create record: %v", err)
	}

	updated := engine.CloneState(record.State)
	updated.RoundNumber = 2
	saved, err := testStore.saveIfRevision("revision-test", record.Revision, updated)
	if err != nil {
		t.Fatalf("failed to save matching revision: %v", err)
	}
	if saved.Revision != record.Revision+1 {
		t.Fatalf("expected incremented revision, got %+v", saved)
	}

	stale := engine.CloneState(record.State)
	stale.RoundNumber = 3
	_, err = testStore.saveIfRevision("revision-test", record.Revision, stale)
	if err != errRevisionConflict {
		t.Fatalf("expected revision conflict, got %v", err)
	}
}

func TestOnlineStateStorePersistsLobbyRequirement(t *testing.T) {
	tempDir := t.TempDir()
	testStore := newOnlineStateStore(tempDir)
	state := game.GameState{
		GameMode: game.GameModeOnline,
	}

	if _, err := testStore.createMultiplayer("lobby-backed", state); err != nil {
		t.Fatalf("failed to create multiplayer record: %v", err)
	}

	reloaded := newOnlineStateStore(tempDir)
	record, ok, err := reloaded.load("lobby-backed")
	if err != nil {
		t.Fatalf("failed to reload multiplayer record: %v", err)
	}
	if !ok {
		t.Fatalf("expected multiplayer record to reload")
	}
	if !record.RequiresLobby {
		t.Fatalf("expected multiplayer record to retain lobby requirement, got %+v", record)
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
		DiscardPile:   []game.CardID{9},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{{ID: 24, Name: "Bird Card"}},
		},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{{ID: 8, Name: "Birdy Bindle"}},
		},
	}
	record, err := store.create("load-test", authoritative)
	if err != nil {
		t.Fatalf("failed to save authoritative state: %v", err)
	}

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
	if resp.Revision != record.Revision {
		t.Fatalf("expected load revision %d, got %d", record.Revision, resp.Revision)
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
	if len(resp.State.DiscardPile) != 1 || resp.State.DiscardPile[0] != 9 {
		t.Fatalf("expected discard pile to remain public on load, got %+v", resp.State.DiscardPile)
	}
}

func TestRedactStateForPlayerKeepsNonOwnerHandsAndSupportersCountOnly(t *testing.T) {
	authoritative := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
		TurnOrder:     []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{{ID: 10, Name: "Cat Card", Suit: game.Fox}},
		},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{{ID: 20, Name: "Bird Card", Suit: game.Bird}},
		},
		Alliance: game.AllianceState{
			CardsInHand: []game.Card{{ID: 30, Name: "Alliance Hand", Suit: game.Mouse}},
			Supporters: []game.Card{
				{ID: 31, Name: "Alliance Supporter", Suit: game.Rabbit},
				{ID: 32, Name: "Alliance Bird Supporter", Suit: game.Bird},
			},
		},
		Vagabond: game.VagabondState{
			CardsInHand: []game.Card{{ID: 40, Name: "Vagabond Card", Suit: game.Fox}},
		},
		HiddenCards: []game.HiddenCard{
			{
				ID:           99,
				OwnerFaction: game.Alliance,
				Zone:         game.HiddenCardZoneSupporters,
				KnownCardID:  31,
			},
		},
		OtherHandCounts: map[game.Faction]int{
			game.Eyrie:    7,
			game.Alliance: 7,
			game.Vagabond: 7,
		},
	}

	visible := redactStateForPlayer(authoritative, game.Marquise)

	if len(visible.Marquise.CardsInHand) != 1 || visible.Marquise.CardsInHand[0].ID != 10 {
		t.Fatalf("expected own hand to remain visible, got %+v", visible.Marquise.CardsInHand)
	}
	if len(visible.Eyrie.CardsInHand) != 0 || len(visible.Alliance.CardsInHand) != 0 || len(visible.Vagabond.CardsInHand) != 0 {
		t.Fatalf("expected non-owner hands to be redacted, got eyrie=%+v alliance=%+v vagabond=%+v", visible.Eyrie.CardsInHand, visible.Alliance.CardsInHand, visible.Vagabond.CardsInHand)
	}
	if visible.OtherHandCounts[game.Eyrie] != 1 || visible.OtherHandCounts[game.Alliance] != 1 || visible.OtherHandCounts[game.Vagabond] != 1 {
		t.Fatalf("expected non-owner hand counts to reflect authoritative hands, got %+v", visible.OtherHandCounts)
	}
	if len(visible.Alliance.Supporters) != 0 {
		t.Fatalf("expected non-Alliance perspective to hide supporter identities, got %+v", visible.Alliance.Supporters)
	}
	if len(visible.HiddenCards) != 2 {
		t.Fatalf("expected supporter placeholders only, got %+v", visible.HiddenCards)
	}
	for _, hidden := range visible.HiddenCards {
		if hidden.OwnerFaction != game.Alliance || hidden.Zone != game.HiddenCardZoneSupporters {
			t.Fatalf("expected Alliance supporter placeholder, got %+v", hidden)
		}
		if hidden.KnownCardID != 0 {
			t.Fatalf("expected supporter placeholder to omit known card identity, got %+v", hidden)
		}
	}
}

func TestRedactStateForAlliancePerspectiveKeepsSupportersVisible(t *testing.T) {
	authoritative := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Alliance,
		TrackAllHands: true,
		TurnOrder:     []game.Faction{game.Marquise, game.Alliance},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{{ID: 10, Name: "Cat Card", Suit: game.Fox}},
		},
		Alliance: game.AllianceState{
			CardsInHand: []game.Card{{ID: 30, Name: "Alliance Hand", Suit: game.Mouse}},
			Supporters:  []game.Card{{ID: 31, Name: "Alliance Supporter", Suit: game.Rabbit}},
		},
	}

	visible := redactStateForPlayer(authoritative, game.Alliance)

	if len(visible.Alliance.CardsInHand) != 1 || visible.Alliance.CardsInHand[0].ID != 30 {
		t.Fatalf("expected Alliance hand to remain visible, got %+v", visible.Alliance.CardsInHand)
	}
	if len(visible.Alliance.Supporters) != 1 || visible.Alliance.Supporters[0].ID != 31 {
		t.Fatalf("expected Alliance supporters to remain visible to Alliance, got %+v", visible.Alliance.Supporters)
	}
	if len(visible.Marquise.CardsInHand) != 0 || visible.OtherHandCounts[game.Marquise] != 1 {
		t.Fatalf("expected non-owner hand count only, hand=%+v counts=%+v", visible.Marquise.CardsInHand, visible.OtherHandCounts)
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
