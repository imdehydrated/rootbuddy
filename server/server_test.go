package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func decodeErrorResponse(t *testing.T, rec *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	return resp
}

func TestHandleHealthCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for health check, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected JSON content type for health check, got %q", got)
	}
}

func TestHandleAPIHealthCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for API health check, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected CORS allow-origin header, got %q", got)
	}
}

func TestHandleCORSPreflight(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/actions/valid", nil)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for CORS preflight, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected CORS allow-origin header, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
		t.Fatalf("expected CORS allow-methods header, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, X-Player-Token" {
		t.Fatalf("expected CORS allow-headers header, got %q", got)
	}
}

func TestHandleValidActions(t *testing.T) {
	body, _ := json.Marshal(ValidActionsRequest{
		State: game.GameState{
			Map: game.Map{
				Clearings: []game.Clearing{
					{
						ID: 1,
						Buildings: []game.Building{
							{Faction: game.Marquise, Type: game.Recruiter},
						},
					},
				},
			},
			FactionTurn:  game.Marquise,
			CurrentPhase: game.Daylight,
			CurrentStep:  game.StepDaylightActions,
			Marquise: game.MarquiseState{
				WarriorSupply: 1,
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/actions/valid", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid actions, got %d", rec.Code)
	}

	var resp ValidActionsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode valid actions response: %v", err)
	}
	foundRecruit := false
	for _, action := range resp.Actions {
		if action.Type == game.ActionRecruit {
			foundRecruit = true
			break
		}
	}
	if !foundRecruit {
		t.Fatalf("expected recruit action response, got %+v", resp.Actions)
	}
}

func TestHandleValidActionsReturnsDaylightMovement(t *testing.T) {
	body, _ := json.Marshal(ValidActionsRequest{
		State: game.GameState{
			Map: game.Map{
				Clearings: []game.Clearing{
					{
						ID:       1,
						Adj:      []int{2},
						Warriors: map[game.Faction]int{game.Marquise: 1},
					},
					{
						ID:  2,
						Adj: []int{1},
					},
				},
			},
			FactionTurn:  game.Marquise,
			CurrentPhase: game.Daylight,
			CurrentStep:  game.StepDaylightActions,
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/actions/valid", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid actions, got %d", rec.Code)
	}

	var resp ValidActionsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode valid actions response: %v", err)
	}
	if len(resp.Actions) == 0 || resp.Actions[0].Type != game.ActionMovement {
		t.Fatalf("expected at least one movement action response, got %+v", resp.Actions)
	}
}

func TestHandleValidActionsBadRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/actions/valid", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid valid-actions body, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "invalid request body" {
		t.Fatalf("expected invalid request body error, got %+v", resp)
	}
}

func TestHandleValidActionsRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/actions/valid",
		bytes.NewBufferString(`{"state":{},"unexpected":true}`),
	)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown request fields, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "invalid request body" {
		t.Fatalf("expected invalid request body error, got %+v", resp)
	}
}

func TestHandleApplyAction(t *testing.T) {
	body, _ := json.Marshal(ApplyActionRequest{
		State: game.GameState{
			Map: game.Map{
				Clearings: []game.Clearing{
					{ID: 1},
				},
			},
			CurrentPhase: game.Daylight,
			CurrentStep:  game.StepDaylightActions,
			Marquise: game.MarquiseState{
				WarriorSupply: 1,
			},
		},
		Action: game.Action{
			Type: game.ActionRecruit,
			Recruit: &game.RecruitAction{
				Faction:     game.Marquise,
				ClearingIDs: []int{1},
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for apply action, got %d", rec.Code)
	}

	var resp ApplyActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode apply action response: %v", err)
	}
	if resp.State.Map.Clearings[0].Warriors[game.Marquise] != 1 {
		t.Fatalf("expected recruit to add one warrior, got %+v", resp.State.Map.Clearings[0].Warriors)
	}
}

func TestHandleApplyActionBadRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/actions/apply", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid apply-action body, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "invalid request body" {
		t.Fatalf("expected invalid request body error, got %+v", resp)
	}
}

func TestHandleApplyActionRejectsTrailingJSON(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/actions/apply",
		bytes.NewBufferString(`{"state":{},"action":{"type":0}} {}`),
	)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for trailing JSON payload, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "invalid request body" {
		t.Fatalf("expected invalid request body error, got %+v", resp)
	}
}

func TestHandleApplyActionRejectsBattleInitiation(t *testing.T) {
	body, _ := json.Marshal(ApplyActionRequest{
		State: game.GameState{},
		Action: game.Action{
			Type: game.ActionBattle,
			Battle: &game.BattleAction{
				Faction:       game.Marquise,
				ClearingID:    1,
				TargetFaction: game.Eyrie,
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when applying unresolved battle, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "battle initiation cannot be applied directly; resolve it first" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleApplyActionRejectsMissingMovementPayload(t *testing.T) {
	body, _ := json.Marshal(ApplyActionRequest{
		State: game.GameState{},
		Action: game.Action{
			Type: game.ActionMovement,
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing movement payload, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "movement payload is required" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleApplyActionAcceptsVagabondForestMovement(t *testing.T) {
	body, _ := json.Marshal(ApplyActionRequest{
		State: game.GameState{
			Map: game.Map{
				Clearings: []game.Clearing{
					{ID: 1},
				},
				Forests: []game.Forest{
					{ID: 1, AdjacentClearings: []int{1}},
				},
			},
			Vagabond: game.VagabondState{
				ClearingID: 1,
				Items: []game.Item{
					{Type: game.ItemBoots, Status: game.ItemReady},
				},
			},
			GamePhase:    game.LifecyclePlaying,
			SetupStage:   game.SetupStageComplete,
			FactionTurn:  game.Vagabond,
			CurrentPhase: game.Daylight,
			CurrentStep:  game.StepDaylightActions,
		},
		Action: game.Action{
			Type: game.ActionMovement,
			Movement: &game.MovementAction{
				Faction:    game.Vagabond,
				Count:      1,
				MaxCount:   1,
				From:       1,
				ToForestID: 1,
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for Vagabond forest move apply action, got %d", rec.Code)
	}
}

func TestHandleApplyActionAcceptsPassPhase(t *testing.T) {
	body, _ := json.Marshal(ApplyActionRequest{
		State: game.GameState{
			CurrentPhase: game.Birdsong,
			CurrentStep:  game.StepBirdsong,
		},
		Action: game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Marquise,
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for pass phase apply action, got %d", rec.Code)
	}

	var resp ApplyActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode pass phase apply response: %v", err)
	}
	if resp.State.CurrentPhase != game.Daylight || resp.State.CurrentStep != game.StepDaylightActions {
		t.Fatalf("expected pass phase to advance to daylight actions, got phase=%v step=%v", resp.State.CurrentPhase, resp.State.CurrentStep)
	}
}

func TestHandleApplyActionAcceptsDiscardEffect(t *testing.T) {
	body, _ := json.Marshal(ApplyActionRequest{
		State: game.GameState{
			GamePhase:    game.LifecyclePlaying,
			SetupStage:   game.SetupStageComplete,
			FactionTurn:  game.Marquise,
			CurrentPhase: game.Daylight,
			CurrentStep:  game.StepDaylightActions,
			PersistentEffects: map[game.Faction][]game.CardID{
				game.Marquise: {15},
			},
		},
		Action: game.Action{
			Type: game.ActionDiscardEffect,
			DiscardEffect: &game.DiscardEffectAction{
				Faction: game.Marquise,
				CardID:  15,
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for discard effect apply action, got %d", rec.Code)
	}

	var resp ApplyActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode discard effect apply response: %v", err)
	}
	if len(resp.State.PersistentEffects) != 0 {
		t.Fatalf("expected discarded effect to be removed from persistent effects, got %+v", resp.State.PersistentEffects)
	}
	if len(resp.State.DiscardPile) != 1 || resp.State.DiscardPile[0] != 15 {
		t.Fatalf("expected discarded effect to move to discard pile, got %+v", resp.State.DiscardPile)
	}
}

func TestHandleApplyActionAcceptsUsePersistentEffect(t *testing.T) {
	body, _ := json.Marshal(ApplyActionRequest{
		State: game.GameState{
			PersistentEffects: map[game.Faction][]game.CardID{
				game.Marquise: {7},
			},
			Map: game.Map{
				Clearings: []game.Clearing{
					{
						ID: 1,
						Warriors: map[game.Faction]int{
							game.Marquise: 1,
						},
					},
				},
			},
		},
		Action: game.Action{
			Type: game.ActionUsePersistentEffect,
			UsePersistentEffect: &game.UsePersistentEffectAction{
				Faction:  game.Marquise,
				EffectID: "royal_claim",
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for use persistent effect apply action, got %d", rec.Code)
	}
}

func TestHandleApplyActionAcceptsActivateDominance(t *testing.T) {
	body, _ := json.Marshal(ApplyActionRequest{
		State: game.GameState{
			GamePhase: game.LifecyclePlaying,
			Marquise: game.MarquiseState{
				CardsInHand: []game.Card{
					{ID: 14, Name: "Dominance", Suit: game.Bird, Kind: game.DominanceCard},
				},
			},
		},
		Action: game.Action{
			Type: game.ActionActivateDominance,
			ActivateDominance: &game.ActivateDominanceAction{
				Faction: game.Marquise,
				CardID:  14,
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for activate dominance apply action, got %d", rec.Code)
	}

	var resp ApplyActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode activate dominance response: %v", err)
	}
	if resp.State.ActiveDominance[game.Marquise] != 14 {
		t.Fatalf("expected active dominance to be tracked, got %+v", resp.State.ActiveDominance)
	}
}

func TestHandleApplyActionAcceptsTakeDominance(t *testing.T) {
	body, _ := json.Marshal(ApplyActionRequest{
		State: game.GameState{
			GamePhase:          game.LifecyclePlaying,
			AvailableDominance: []game.CardID{27},
			Marquise: game.MarquiseState{
				CardsInHand: []game.Card{
					{ID: 24, Name: "A Visit to Friends", Suit: game.Rabbit},
				},
			},
		},
		Action: game.Action{
			Type: game.ActionTakeDominance,
			TakeDominance: &game.TakeDominanceAction{
				Faction:         game.Marquise,
				DominanceCardID: 27,
				SpentCardID:     24,
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for take dominance apply action, got %d", rec.Code)
	}

	var resp ApplyActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode take dominance response: %v", err)
	}
	if len(resp.State.Marquise.CardsInHand) != 1 || resp.State.Marquise.CardsInHand[0].ID != 27 {
		t.Fatalf("expected taken dominance card in hand, got %+v", resp.State.Marquise.CardsInHand)
	}
}

func TestHandleResolveBattle(t *testing.T) {
	body, _ := json.Marshal(ResolveBattleRequest{
		State: game.GameState{
			Map: game.Map{
				Clearings: []game.Clearing{
					{
						ID: 1,
						Warriors: map[game.Faction]int{
							game.Marquise: 2,
							game.Eyrie:    1,
						},
					},
				},
			},
		},
		Action: game.Action{
			Type: game.ActionBattle,
			Battle: &game.BattleAction{
				Faction:       game.Marquise,
				ClearingID:    1,
				TargetFaction: game.Eyrie,
			},
		},
		AttackerRoll: 1,
		DefenderRoll: 0,
	})

	req := httptest.NewRequest(http.MethodPost, "/battles/resolve", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for resolve battle, got %d", rec.Code)
	}

	var resp ResolveBattleResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode resolve battle response: %v", err)
	}
	if resp.Action.Type != game.ActionBattleResolution || resp.Action.BattleResolution == nil {
		t.Fatalf("expected resolved battle action, got %+v", resp.Action)
	}
}

func TestHandleResolveBattleWithModifiers(t *testing.T) {
	body, _ := json.Marshal(ResolveBattleRequest{
		State: game.GameState{
			Map: game.Map{
				Clearings: []game.Clearing{
					{
						ID: 1,
						Warriors: map[game.Faction]int{
							game.Marquise: 2,
							game.Eyrie:    2,
						},
					},
				},
			},
		},
		Action: game.Action{
			Type: game.ActionBattle,
			Battle: &game.BattleAction{
				Faction:       game.Marquise,
				ClearingID:    1,
				TargetFaction: game.Eyrie,
			},
		},
		AttackerRoll: 1,
		DefenderRoll: 1,
		UseModifiers: true,
		Modifiers: game.BattleModifiers{
			AttackerHitModifier: 1,
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/battles/resolve", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for resolve battle with modifiers, got %d", rec.Code)
	}

	var resp ResolveBattleResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode resolve battle response: %v", err)
	}
	if resp.Action.BattleResolution == nil {
		t.Fatalf("expected resolved battle payload, got %+v", resp.Action)
	}
	if resp.Action.BattleResolution.DefenderLosses != 2 {
		t.Fatalf("expected modifier-adjusted defender losses of 2, got %d", resp.Action.BattleResolution.DefenderLosses)
	}
	if resp.Action.BattleResolution.AttackerHitModifier != 1 {
		t.Fatalf("expected attacker hit modifier to be recorded, got %d", resp.Action.BattleResolution.AttackerHitModifier)
	}
}

func TestHandleResolveBattleBadRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/battles/resolve", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid resolve-battle body, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "invalid request body" {
		t.Fatalf("expected invalid request body error, got %+v", resp)
	}
}

func TestHandleResolveBattleRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/battles/resolve",
		bytes.NewBufferString(`{"state":{},"action":{"type":1},"attackerRoll":1,"defenderRoll":1,"extra":true}`),
	)
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown resolve-battle fields, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "invalid request body" {
		t.Fatalf("expected invalid request body error, got %+v", resp)
	}
}

func TestHandleResolveBattleRejectsWrongActionType(t *testing.T) {
	body, _ := json.Marshal(ResolveBattleRequest{
		State: game.GameState{},
		Action: game.Action{
			Type: game.ActionMovement,
		},
		AttackerRoll: 1,
		DefenderRoll: 1,
	})

	req := httptest.NewRequest(http.MethodPost, "/battles/resolve", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for wrong resolve-battle action type, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "battle resolution requires a battle action" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleResolveBattleRejectsCoalitionPartnerTarget(t *testing.T) {
	body, _ := json.Marshal(ResolveBattleRequest{
		State: game.GameState{
			CoalitionActive:  true,
			CoalitionPartner: game.Marquise,
			ActiveDominance: map[game.Faction]game.CardID{
				game.Vagabond: 14,
			},
		},
		Action: game.Action{
			Type: game.ActionBattle,
			Battle: &game.BattleAction{
				Faction:       game.Marquise,
				ClearingID:    1,
				TargetFaction: game.Vagabond,
			},
		},
		AttackerRoll: 1,
		DefenderRoll: 1,
	})

	req := httptest.NewRequest(http.MethodPost, "/battles/resolve", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for coalition-partner battle target, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "battle action must target an enemy faction" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleResolveBattleRejectsOutOfRangeRolls(t *testing.T) {
	body, _ := json.Marshal(ResolveBattleRequest{
		State: game.GameState{},
		Action: game.Action{
			Type: game.ActionBattle,
			Battle: &game.BattleAction{
				Faction:       game.Marquise,
				ClearingID:    1,
				TargetFaction: game.Eyrie,
			},
		},
		AttackerRoll: 4,
		DefenderRoll: -1,
	})

	req := httptest.NewRequest(http.MethodPost, "/battles/resolve", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for out-of-range battle rolls, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "battle rolls must be between 0 and 3" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleBattleContext(t *testing.T) {
	body, _ := json.Marshal(BattleContextRequest{
		State: game.GameState{
			GameMode:      game.GameModeAssist,
			PlayerFaction: game.Marquise,
			PersistentEffects: map[game.Faction][]game.CardID{
				game.Marquise: {30},
				game.Eyrie:    {1, 3},
			},
			OtherHandCounts: map[game.Faction]int{
				game.Eyrie: 2,
			},
			Map: game.Map{
				Clearings: []game.Clearing{
					{
						ID:   1,
						Suit: game.Fox,
					},
				},
			},
		},
		Action: game.Action{
			Type: game.ActionBattle,
			Battle: &game.BattleAction{
				Faction:       game.Marquise,
				ClearingID:    1,
				TargetFaction: game.Eyrie,
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/battles/context", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for battle context, got %d", rec.Code)
	}

	var resp BattleContextResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode battle context response: %v", err)
	}
	if !resp.BattleContext.AttackerHasScoutingParty {
		t.Fatalf("expected scouting party to be surfaced in battle context, got %+v", resp.BattleContext)
	}
	if resp.BattleContext.CanDefenderAmbush {
		t.Fatalf("expected scouting party to suppress defender ambush, got %+v", resp.BattleContext)
	}
	if !resp.BattleContext.CanDefenderArmorers || !resp.BattleContext.CanDefenderSappers {
		t.Fatalf("expected defender persistent battle effects to be exposed, got %+v", resp.BattleContext)
	}
}

func TestHandleSetup(t *testing.T) {
	body, _ := json.Marshal(SetupRequest{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		Factions:      []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		MapID:         game.AutumnMapID,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/game/setup", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for setup, got %d", rec.Code)
	}

	var resp SetupResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode setup response: %v", err)
	}
	if resp.State.GamePhase != game.LifecycleSetup {
		t.Fatalf("expected setup lifecycle state, got %+v", resp.State)
	}
	if resp.GameID == "" {
		t.Fatalf("expected online setup to return a game ID, got %+v", resp)
	}
	if resp.Revision <= 0 {
		t.Fatalf("expected online setup to return a revision, got %+v", resp)
	}
	if resp.State.SetupStage != game.SetupStageMarquise {
		t.Fatalf("expected setup to begin at Marquise stage, got %+v", resp.State)
	}
	if len(resp.State.Map.Clearings) != 12 {
		t.Fatalf("expected autumn map clearings, got %+v", resp.State.Map.Clearings)
	}
}

func TestOnlineSetupApplyRedactsHiddenHandsAfterFinalSetup(t *testing.T) {
	setupBody, _ := json.Marshal(SetupRequest{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		Factions:      []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		MapID:         game.AutumnMapID,
	})

	setupReq := httptest.NewRequest(http.MethodPost, "/api/game/setup", bytes.NewReader(setupBody))
	setupRec := httptest.NewRecorder()
	NewServer().ServeHTTP(setupRec, setupReq)

	if setupRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for setup, got %d", setupRec.Code)
	}

	var setupResp SetupResponse
	if err := json.Unmarshal(setupRec.Body.Bytes(), &setupResp); err != nil {
		t.Fatalf("failed to decode setup response: %v", err)
	}

	state := setupResp.State
	gameID := setupResp.GameID

	applyAndDecode := func(action game.Action) game.GameState {
		body, _ := json.Marshal(ApplyActionRequest{
			State:  state,
			Action: action,
			GameID: gameID,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		NewServer().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for apply action, got %d body=%s", rec.Code, rec.Body.String())
		}
		var resp ApplyActionResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode apply response: %v", err)
		}
		state = resp.State
		return state
	}

	state = applyAndDecode(game.Action{
		Type: game.ActionMarquiseSetup,
		MarquiseSetup: &game.MarquiseSetupAction{
			Faction:             game.Marquise,
			KeepClearingID:      1,
			SawmillClearingID:   1,
			WorkshopClearingID:  5,
			RecruiterClearingID: 10,
		},
	})
	state = applyAndDecode(game.Action{
		Type: game.ActionEyrieSetup,
		EyrieSetup: &game.EyrieSetupAction{
			Faction:    game.Eyrie,
			Leader:     game.LeaderBuilder,
			ClearingID: 3,
		},
	})
	state = applyAndDecode(game.Action{
		Type: game.ActionVagabondSetup,
		VagabondSetup: &game.VagabondSetupAction{
			Faction:   game.Vagabond,
			Character: game.CharThief,
			ForestID:  7,
		},
	})

	if state.GamePhase != game.LifecyclePlaying {
		t.Fatalf("expected final setup action to enter playing state, got %+v", state)
	}
	if len(state.Marquise.CardsInHand) != 3 {
		t.Fatalf("expected player hand to stay visible, got %+v", state.Marquise.CardsInHand)
	}
	if len(state.Eyrie.CardsInHand) != 0 || len(state.Alliance.CardsInHand) != 0 || len(state.Vagabond.CardsInHand) != 0 {
		t.Fatalf("expected non-player hands to be redacted, got eyrie=%+v alliance=%+v vagabond=%+v", state.Eyrie.CardsInHand, state.Alliance.CardsInHand, state.Vagabond.CardsInHand)
	}
	if state.OtherHandCounts[game.Eyrie] != 3 || state.OtherHandCounts[game.Alliance] != 3 || state.OtherHandCounts[game.Vagabond] != 3 {
		t.Fatalf("expected redacted other hand counts after setup, got %+v", state.OtherHandCounts)
	}
	hiddenSupporters := 0
	for _, hidden := range state.HiddenCards {
		if hidden.OwnerFaction == game.Alliance && hidden.Zone == game.HiddenCardZoneSupporters {
			hiddenSupporters++
		}
	}
	if hiddenSupporters != 3 {
		t.Fatalf("expected redacted Alliance supporter count after setup, got %+v", state.HiddenCards)
	}
	if len(state.Deck) != 39 {
		t.Fatalf("expected redacted deck to preserve remaining count only, got %+v", state.Deck)
	}
	for _, cardID := range state.Deck {
		if cardID != 0 {
			t.Fatalf("expected redacted deck order placeholders, got %+v", state.Deck)
		}
	}
}

func TestHandleApplyActionRejectsUnknownGameID(t *testing.T) {
	body, _ := json.Marshal(ApplyActionRequest{
		GameID: "missing",
		State:  game.GameState{},
		Action: game.Action{
			Type: game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{
				Faction: game.Marquise,
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown game ID, got %d", rec.Code)
	}

	resp := decodeErrorResponse(t, rec)
	if resp.Error != "unknown game id" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleApplyActionReturnsCodebreakersRevealForPlayerPerspective(t *testing.T) {
	gameID := "codebreakers-test"
	authoritative := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
		TurnOrder:     []game.Faction{game.Marquise, game.Eyrie},
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {28},
		},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{
				{ID: 8, Name: "Birdy Bindle"},
				{ID: 12, Name: "Ambush"},
			},
		},
	}
	record, err := store.create(gameID, authoritative)
	if err != nil {
		t.Fatalf("failed to save authoritative state: %v", err)
	}
	visible := redactStateForPlayer(authoritative, game.Marquise)

	body, _ := json.Marshal(ApplyActionRequest{
		GameID: gameID,
		State:  visible,
		Action: game.Action{
			Type: game.ActionUsePersistentEffect,
			UsePersistentEffect: &game.UsePersistentEffectAction{
				Faction:       game.Marquise,
				EffectID:      "codebreakers",
				TargetFaction: game.Eyrie,
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for Codebreakers apply action, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp ApplyActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode Codebreakers apply response: %v", err)
	}
	if resp.EffectResult == nil || len(resp.EffectResult.Cards) != 2 {
		t.Fatalf("expected Codebreakers reveal result, got %+v", resp.EffectResult)
	}
	if resp.Revision != record.Revision+1 {
		t.Fatalf("expected apply response revision %d, got %d", record.Revision+1, resp.Revision)
	}
	if len(resp.State.Eyrie.CardsInHand) != 0 {
		t.Fatalf("expected redacted visible state to keep Eyrie hand hidden, got %+v", resp.State.Eyrie.CardsInHand)
	}
}

func TestHandleApplyActionReturnsStandAndDeliverTransferredCardForPlayer(t *testing.T) {
	gameID := "stand-test"
	authoritative := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
		RandomSeed:    3,
		TurnOrder:     []game.Faction{game.Marquise, game.Eyrie},
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {41},
		},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{
				{ID: 8, Name: "Birdy Bindle"},
			},
		},
	}
	record, err := store.create(gameID, authoritative)
	if err != nil {
		t.Fatalf("failed to save authoritative state: %v", err)
	}
	visible := redactStateForPlayer(authoritative, game.Marquise)

	body, _ := json.Marshal(ApplyActionRequest{
		GameID: gameID,
		State:  visible,
		Action: game.Action{
			Type: game.ActionUsePersistentEffect,
			UsePersistentEffect: &game.UsePersistentEffectAction{
				Faction:       game.Marquise,
				EffectID:      "stand_and_deliver",
				TargetFaction: game.Eyrie,
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/actions/apply", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for Stand and Deliver apply action, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp ApplyActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode Stand and Deliver apply response: %v", err)
	}
	if resp.EffectResult == nil || len(resp.EffectResult.Cards) != 1 || resp.EffectResult.Cards[0].ID != 8 {
		t.Fatalf("expected transferred card result, got %+v", resp.EffectResult)
	}
	if resp.Revision != record.Revision+1 {
		t.Fatalf("expected apply response revision %d, got %d", record.Revision+1, resp.Revision)
	}
	if len(resp.State.Marquise.CardsInHand) != 1 || resp.State.Marquise.CardsInHand[0].ID != 8 {
		t.Fatalf("expected visible state to include transferred card, got %+v", resp.State.Marquise.CardsInHand)
	}
}
