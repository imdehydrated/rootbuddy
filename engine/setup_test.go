package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestSetupGameOnlineFourPlayersBuildsInitialState(t *testing.T) {
	state, err := SetupGame(SetupRequest{
		GameMode:          game.GameModeOnline,
		PlayerFaction:     game.Marquise,
		Factions:          []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		MapID:             game.AutumnMapID,
		VagabondCharacter: game.CharThief,
		EyrieLeader:       game.LeaderCommander,
	})
	if err != nil {
		t.Fatalf("expected setup to succeed, got %v", err)
	}

	if state.GamePhase != game.LifecyclePlaying {
		t.Fatalf("expected playing phase, got %v", state.GamePhase)
	}
	if state.FactionTurn != game.Marquise || state.CurrentPhase != game.Birdsong || state.CurrentStep != game.StepBirdsong {
		t.Fatalf("expected Marquise birdsong start, got faction=%v phase=%v step=%v", state.FactionTurn, state.CurrentPhase, state.CurrentStep)
	}
	if len(state.Marquise.CardsInHand) != 3 {
		t.Fatalf("expected player Marquise to start with 3 cards, got %+v", state.Marquise.CardsInHand)
	}
	if state.OtherHandCounts[game.Eyrie] != 3 || state.OtherHandCounts[game.Alliance] != 3 || state.OtherHandCounts[game.Vagabond] != 3 {
		t.Fatalf("expected other hand counts to start at 3, got %+v", state.OtherHandCounts)
	}
	if len(state.Deck) != 39 {
		t.Fatalf("expected online deck to have 39 cards after setup, got %d", len(state.Deck))
	}
	if len(state.QuestDeck) != 12 || len(state.Vagabond.QuestsAvailable) != 3 {
		t.Fatalf("expected 12 quests remaining and 3 available, got deck=%d available=%d", len(state.QuestDeck), len(state.Vagabond.QuestsAvailable))
	}
	if state.ItemSupply[game.ItemCoin] != 2 {
		t.Fatalf("expected shared item supply to remain the 12-item market pool, got %+v", state.ItemSupply)
	}
}

func TestSetupGameAssistLeavesDeckEmptyAndPlayerHandUnknown(t *testing.T) {
	state, err := SetupGame(SetupRequest{
		GameMode:          game.GameModeAssist,
		PlayerFaction:     game.Vagabond,
		Factions:          []game.Faction{game.Marquise, game.Eyrie, game.Vagabond},
		MapID:             game.AutumnMapID,
		VagabondCharacter: game.CharRanger,
		EyrieLeader:       game.LeaderBuilder,
	})
	if err != nil {
		t.Fatalf("expected setup to succeed, got %v", err)
	}

	if len(state.Deck) != 0 {
		t.Fatalf("expected assist mode to leave deck empty, got %+v", state.Deck)
	}
	if len(state.Vagabond.CardsInHand) != 0 {
		t.Fatalf("expected assist mode player hand to remain empty until manual entry, got %+v", state.Vagabond.CardsInHand)
	}
	if state.OtherHandCounts[game.Marquise] != 3 || state.OtherHandCounts[game.Eyrie] != 3 {
		t.Fatalf("expected other factions to start with hand count 3, got %+v", state.OtherHandCounts)
	}
	if state.Vagabond.ForestID != 7 || !state.Vagabond.InForest {
		t.Fatalf("expected Vagabond to start in forest 7, got %+v", state.Vagabond)
	}
	if len(state.Vagabond.Items) != 4 || state.Vagabond.Items[2].Type != game.ItemCrossbow {
		t.Fatalf("expected Ranger starting items, got %+v", state.Vagabond.Items)
	}
}

func TestSetupGameTwoPlayersRemovesDominanceCards(t *testing.T) {
	state, err := SetupGame(SetupRequest{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		Factions:      []game.Faction{game.Marquise, game.Eyrie},
		MapID:         game.AutumnMapID,
		EyrieLeader:   game.LeaderBuilder,
	})
	if err != nil {
		t.Fatalf("expected setup to succeed, got %v", err)
	}

	if len(state.Deck) != 44 {
		t.Fatalf("expected 44 deck cards after removing four dominance cards and drawing six, got %d", len(state.Deck))
	}
}
