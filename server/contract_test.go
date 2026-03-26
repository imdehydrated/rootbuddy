package server

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestSetupResponseJSONIncludesDeterministicStateFields(t *testing.T) {
	body, err := json.Marshal(SetupResponse{
		State: game.GameState{
			RandomSeed:   99,
			ShuffleCount: 3,
		},
		GameID: "game-123",
	})
	if err != nil {
		t.Fatalf("failed to marshal setup response: %v", err)
	}

	jsonText := string(body)
	for _, key := range []string{`"RandomSeed"`, `"ShuffleCount"`, `"gameID"`} {
		if !strings.Contains(jsonText, key) {
			t.Fatalf("expected setup response JSON to include %s, got %s", key, jsonText)
		}
	}
}

func TestBattleContextResponseJSONIncludesStableContractKeys(t *testing.T) {
	body, err := json.Marshal(BattleContextResponse{
		BattleContext: game.BattleContext{
			CanDefenderAmbush: true,
			Timing:            []game.BattleTimingStep{game.BattleTimingAmbush},
		},
		GameID: "game-123",
	})
	if err != nil {
		t.Fatalf("failed to marshal battle context response: %v", err)
	}

	jsonText := string(body)
	for _, key := range []string{`"battleContext"`, `"CanDefenderAmbush"`, `"Timing"`} {
		if !strings.Contains(jsonText, key) {
			t.Fatalf("expected battle context JSON to include %s, got %s", key, jsonText)
		}
	}
}
