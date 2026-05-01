package rules

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidMarquiseEveningActionsRequiresDiscardAfterDraw(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		TurnProgress: game.TurnProgress{
			EveningDrawn: true,
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 8},
				{ID: 9},
				{ID: 10},
				{ID: 11},
				{ID: 12},
				{ID: 13},
			},
		},
	}

	got := ValidMarquiseEveningActions(state)
	want := game.Action{
		Type: game.ActionEveningDiscard,
		EveningDiscard: &game.EveningDiscardAction{
			Faction: game.Marquise,
			CardIDs: []game.CardID{
				8,
			},
			Count: 1,
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected Marquise discard choice %+v, got %+v", want, got)
	}
	for _, action := range got {
		if action.Type == game.ActionEveningDraw {
			t.Fatalf("did not expect another draw before discard is resolved, got %+v", got)
		}
	}
}

func TestValidEyrieEveningActionsAllowsNoDiscardWhenAtHandLimit(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		TurnProgress: game.TurnProgress{
			EveningMainActionTaken: true,
			EveningDrawn:           true,
		},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{
				{ID: 8},
				{ID: 9},
				{ID: 10},
				{ID: 11},
				{ID: 12},
			},
		},
	}

	got := ValidEyrieEveningActions(state)
	want := game.Action{
		Type: game.ActionEveningDiscard,
		EveningDiscard: &game.EveningDiscardAction{
			Faction: game.Eyrie,
		},
	}

	if len(got) != 1 || !containsAction(got, want) {
		t.Fatalf("expected only Eyrie no-discard confirmation %+v, got %+v", want, got)
	}
}

func TestValidAllianceEveningActionsDoesNotGenerateIdentitylessHiddenDiscard(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Marquise,
		FactionTurn:   game.Alliance,
		CurrentPhase:  game.Evening,
		CurrentStep:   game.StepEvening,
		TurnProgress: game.TurnProgress{
			EveningDrawn: true,
		},
		OtherHandCounts: map[game.Faction]int{
			game.Alliance: 7,
		},
	}

	got := ValidAllianceEveningActions(state)
	if len(got) != 0 {
		t.Fatalf("expected no generated count-only hidden discard because discarded cards are public, got %+v", got)
	}
}
