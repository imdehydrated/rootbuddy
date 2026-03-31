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
		ActionLog: []ActionLogEntry{{
			RoundNumber: 2,
			Faction:     game.Marquise,
			ActionType:  game.ActionRecruit,
			Summary:     "Recruit in clearings [1]",
			Timestamp:   1234,
		}},
	})
	if err != nil {
		t.Fatalf("failed to marshal game state message: %v", err)
	}

	jsonText := string(body)
	for _, key := range []string{`"type"`, `"gameID"`, `"revision"`, `"state"`, `"RandomSeed"`, `"actionLog"`, `"summary"`} {
		if !strings.Contains(jsonText, key) {
			t.Fatalf("expected game state message JSON to include %s, got %s", key, jsonText)
		}
	}
}

func TestBattlePromptMessageJSONIncludesPromptFields(t *testing.T) {
	body, err := json.Marshal(BattlePromptMessage{
		Type: socketMessageBattlePrompt,
		Prompt: &BattlePrompt{
			GameID:           "game-123",
			Revision:         5,
			Action:           game.Action{Type: game.ActionBattle},
			Stage:            BattlePromptDefenderTurn,
			WaitingOnFaction: game.Eyrie,
			BattleContext: game.BattleContext{
				CanDefenderAmbush: true,
			},
			CanUseAmbush: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal battle prompt message: %v", err)
	}

	jsonText := string(body)
	for _, key := range []string{`"type"`, `"prompt"`, `"gameID"`, `"stage"`, `"waitingOnFaction"`, `"canUseAmbush"`} {
		if !strings.Contains(jsonText, key) {
			t.Fatalf("expected battle prompt message JSON to include %s, got %s", key, jsonText)
		}
	}
}
