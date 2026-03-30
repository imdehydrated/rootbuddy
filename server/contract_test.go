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
		GameID:   "game-123",
		Revision: 4,
	})
	if err != nil {
		t.Fatalf("failed to marshal setup response: %v", err)
	}

	jsonText := string(body)
	for _, key := range []string{`"RandomSeed"`, `"ShuffleCount"`, `"gameID"`, `"revision"`} {
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
		GameID:   "game-123",
		Revision: 7,
	})
	if err != nil {
		t.Fatalf("failed to marshal battle context response: %v", err)
	}

	jsonText := string(body)
	for _, key := range []string{`"battleContext"`, `"CanDefenderAmbush"`, `"Timing"`, `"revision"`} {
		if !strings.Contains(jsonText, key) {
			t.Fatalf("expected battle context JSON to include %s, got %s", key, jsonText)
		}
	}
}

func TestLoadGameResponseJSONIncludesStableContractKeys(t *testing.T) {
	body, err := json.Marshal(LoadGameResponse{
		State: game.GameState{
			RandomSeed: 11,
		},
		GameID:   "game-123",
		Revision: 2,
	})
	if err != nil {
		t.Fatalf("failed to marshal load game response: %v", err)
	}

	jsonText := string(body)
	for _, key := range []string{`"state"`, `"RandomSeed"`, `"gameID"`, `"revision"`} {
		if !strings.Contains(jsonText, key) {
			t.Fatalf("expected load game response JSON to include %s, got %s", key, jsonText)
		}
	}
}

func TestGameStateMessageJSONIncludesRevisionAndType(t *testing.T) {
	body, err := json.Marshal(GameStateMessage{
		Type:     socketMessageGameState,
		GameID:   "game-123",
		Revision: 9,
		State: game.GameState{
			RandomSeed: 21,
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal game state message: %v", err)
	}

	jsonText := string(body)
	for _, key := range []string{`"type"`, `"gameID"`, `"revision"`, `"state"`, `"RandomSeed"`} {
		if !strings.Contains(jsonText, key) {
			t.Fatalf("expected game state message JSON to include %s, got %s", key, jsonText)
		}
	}
}
