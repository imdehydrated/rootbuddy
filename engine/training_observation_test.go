package engine

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestNewTrainingObservationRedactsHiddenInformationAndRNGState(t *testing.T) {
	state := game.GameState{
		GameMode:        game.GameModeOnline,
		TrackAllHands:   true,
		RandomSeed:      123,
		ShuffleCount:    4,
		BattleRollCount: 2,
		PlayerFaction:   game.Marquise,
		FactionTurn:     game.Marquise,
		CurrentPhase:    game.Daylight,
		CurrentStep:     game.StepDaylightActions,
		TurnOrder:       []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		Deck:            []game.CardID{8, 9, 10},
		QuestDeck:       []game.QuestID{1, 2},
		QuestDiscard:    []game.QuestID{3},
		DiscardPile:     []game.CardID{53},
		OtherHandCounts: map[game.Faction]int{
			game.Eyrie: 4,
		},
		HiddenCards: []game.HiddenCard{
			{ID: 1, OwnerFaction: game.Eyrie, Zone: game.HiddenCardZoneHand, KnownCardID: 31},
			{ID: 2, OwnerFaction: game.Alliance, Zone: game.HiddenCardZoneSupporters, KnownCardID: 24},
		},
		NextHiddenCardID: 3,
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{trainingObservationTestCard(t, 52)},
		},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{trainingObservationTestCard(t, 11)},
		},
		Alliance: game.AllianceState{
			CardsInHand: []game.Card{trainingObservationTestCard(t, 23)},
			Supporters:  []game.Card{trainingObservationTestCard(t, 24), trainingObservationTestCard(t, 25)},
		},
		Vagabond: game.VagabondState{
			CardsInHand: []game.Card{trainingObservationTestCard(t, 37)},
		},
	}

	obs := NewTrainingObservation(state, TrainingObservationOptions{Perspective: game.Marquise})

	if obs.DebugState != nil || obs.Omniscient {
		t.Fatalf("expected public observation without debug state, got %+v", obs)
	}
	if got := obs.Hidden.Hands[game.Eyrie]; got != 1 {
		t.Fatalf("expected Eyrie hand count only, got %+v", obs.Hidden.Hands)
	}
	if got := obs.Hidden.Hands[game.Alliance]; got != 1 {
		t.Fatalf("expected Alliance hand count only, got %+v", obs.Hidden.Hands)
	}
	if got := obs.Hidden.Hands[game.Vagabond]; got != 1 {
		t.Fatalf("expected Vagabond hand count only, got %+v", obs.Hidden.Hands)
	}
	if obs.Hidden.AllianceSupporters != 2 || obs.Hidden.Deck != 3 || obs.Hidden.QuestDeck != 2 {
		t.Fatalf("expected hidden zone counts for supporters/decks, got %+v", obs.Hidden)
	}
	if len(obs.State.Marquise.CardsInHand) != 1 || obs.State.Marquise.CardsInHand[0].ID != 52 {
		t.Fatalf("expected own hand to remain visible, got %+v", obs.State.Marquise.CardsInHand)
	}
	if len(obs.State.Eyrie.CardsInHand) != 0 || len(obs.State.Alliance.CardsInHand) != 0 || len(obs.State.Vagabond.CardsInHand) != 0 {
		t.Fatalf("expected non-owner hands to be hidden, eyrie=%+v alliance=%+v vagabond=%+v", obs.State.Eyrie.CardsInHand, obs.State.Alliance.CardsInHand, obs.State.Vagabond.CardsInHand)
	}
	if len(obs.State.Alliance.Supporters) != 0 {
		t.Fatalf("expected non-Alliance supporter identities hidden, got %+v", obs.State.Alliance.Supporters)
	}
	if len(obs.State.QuestDiscard) != 1 || obs.State.QuestDiscard[0] != 3 {
		t.Fatalf("expected public quest discard to remain visible, got %+v", obs.State.QuestDiscard)
	}

	payload, err := json.Marshal(obs)
	if err != nil {
		t.Fatalf("marshal training observation: %v", err)
	}
	jsonText := string(payload)
	for _, forbidden := range []string{
		`"RandomSeed"`,
		`"ShuffleCount"`,
		`"BattleRollCount"`,
		`"Deck":[`,
		`"QuestDeck"`,
		`"HiddenCards"`,
		`"NextHiddenCardID"`,
		`"KnownCardID"`,
		`"debugState"`,
	} {
		if strings.Contains(jsonText, forbidden) {
			t.Fatalf("public training observation leaked %s in %s", forbidden, jsonText)
		}
	}
	if !strings.Contains(jsonText, `"hidden"`) || !strings.Contains(jsonText, `"deck":3`) || !strings.Contains(jsonText, `"questDeck":2`) {
		t.Fatalf("expected explicit hidden counts in observation JSON, got %s", jsonText)
	}
}

func TestNewTrainingObservationShowsAllianceSupportersOnlyToAlliance(t *testing.T) {
	supporter := trainingObservationTestCard(t, 24)
	state := game.GameState{
		TurnOrder: []game.Faction{game.Marquise, game.Alliance},
		Alliance: game.AllianceState{
			Supporters: []game.Card{supporter},
		},
	}

	obs := NewTrainingObservation(state, TrainingObservationOptions{Perspective: game.Alliance})

	if obs.Hidden.AllianceSupporters != 0 {
		t.Fatalf("expected Alliance perspective not to hide its own supporter count, got %+v", obs.Hidden)
	}
	if len(obs.State.Alliance.Supporters) != 1 || obs.State.Alliance.Supporters[0].ID != supporter.ID {
		t.Fatalf("expected Alliance supporters visible to Alliance, got %+v", obs.State.Alliance.Supporters)
	}
}

func TestNewTrainingObservationOmniscientModeRequiresExplicitDebugState(t *testing.T) {
	state := game.GameState{
		RandomSeed:   99,
		ShuffleCount: 7,
		Deck:         []game.CardID{8},
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{trainingObservationTestCard(t, 11)},
		},
	}

	obs := NewTrainingObservation(state, TrainingObservationOptions{
		Perspective: game.Marquise,
		Omniscient:  true,
	})

	if !obs.Omniscient || obs.DebugState == nil {
		t.Fatalf("expected explicit debug state in omniscient mode, got %+v", obs)
	}
	if obs.DebugState.RandomSeed != 99 || obs.DebugState.ShuffleCount != 7 || len(obs.DebugState.Deck) != 1 {
		t.Fatalf("expected debug state to preserve simulator internals, got %+v", obs.DebugState)
	}
	if len(obs.State.Eyrie.CardsInHand) != 0 {
		t.Fatalf("expected public state to remain redacted even in omniscient observation, got %+v", obs.State.Eyrie.CardsInHand)
	}
}

func trainingObservationTestCard(t *testing.T, id game.CardID) game.Card {
	t.Helper()

	card, ok := CardByID(id)
	if !ok {
		t.Fatalf("missing card id %d", id)
	}
	return card
}
