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
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
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
