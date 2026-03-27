package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestSetupGameStartsInLifecycleSetup(t *testing.T) {
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

	if state.GamePhase != game.LifecycleSetup {
		t.Fatalf("expected setup lifecycle, got %v", state.GamePhase)
	}
	if state.SetupStage != game.SetupStageMarquise || state.FactionTurn != game.Marquise {
		t.Fatalf("expected Marquise setup stage first, got stage=%v faction=%v", state.SetupStage, state.FactionTurn)
	}
	if len(state.Deck) != 0 {
		t.Fatalf("expected deck to remain undealt until setup is complete, got %+v", state.Deck)
	}
	if state.Marquise.KeepClearingID != 0 || state.Eyrie.RoostsPlaced != 0 || state.Vagabond.ForestID != 0 {
		t.Fatalf("expected no setup choices to be resolved yet, got marquise=%+v eyrie=%+v vagabond=%+v", state.Marquise, state.Eyrie, state.Vagabond)
	}
}

func TestValidSetupActionsGenerateLegalMarquiseChoices(t *testing.T) {
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

	actions := ValidActions(state)
	if len(actions) == 0 {
		t.Fatalf("expected Marquise setup actions, got none")
	}

	foundKeepOne := false
	for _, action := range actions {
		if action.Type != game.ActionMarquiseSetup || action.MarquiseSetup == nil {
			t.Fatalf("expected only Marquise setup actions, got %+v", action)
		}
		if action.MarquiseSetup.KeepClearingID == 1 {
			foundKeepOne = true
			legalSites := map[int]bool{1: true, 5: true, 9: true, 10: true}
			if !legalSites[action.MarquiseSetup.SawmillClearingID] ||
				!legalSites[action.MarquiseSetup.WorkshopClearingID] ||
				!legalSites[action.MarquiseSetup.RecruiterClearingID] {
				t.Fatalf("expected keep-1 setup buildings to stay in keep/adjacent clearings, got %+v", action.MarquiseSetup)
			}
		}
	}

	if !foundKeepOne {
		t.Fatalf("expected to generate at least one keep-1 Marquise setup action, got %+v", actions)
	}
}

func TestSetupActionsAdvanceToPlayingStateAndDealHands(t *testing.T) {
	state, err := SetupGame(SetupRequest{
		GameMode:          game.GameModeOnline,
		PlayerFaction:     game.Marquise,
		Factions:          []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		MapID:             game.AutumnMapID,
		VagabondCharacter: game.CharRanger,
		EyrieLeader:       game.LeaderBuilder,
	})
	if err != nil {
		t.Fatalf("expected setup to succeed, got %v", err)
	}

	state = ApplyAction(state, game.Action{
		Type: game.ActionMarquiseSetup,
		MarquiseSetup: &game.MarquiseSetupAction{
			Faction:             game.Marquise,
			KeepClearingID:      1,
			SawmillClearingID:   1,
			WorkshopClearingID:  5,
			RecruiterClearingID: 10,
		},
	})
	if state.SetupStage != game.SetupStageEyrie || state.FactionTurn != game.Eyrie {
		t.Fatalf("expected setup to advance to Eyrie, got stage=%v faction=%v", state.SetupStage, state.FactionTurn)
	}

	state = ApplyAction(state, game.Action{
		Type: game.ActionEyrieSetup,
		EyrieSetup: &game.EyrieSetupAction{
			Faction:    game.Eyrie,
			ClearingID: 3,
		},
	})
	if state.SetupStage != game.SetupStageVagabond || state.FactionTurn != game.Vagabond {
		t.Fatalf("expected setup to advance to Vagabond, got stage=%v faction=%v", state.SetupStage, state.FactionTurn)
	}

	state = ApplyAction(state, game.Action{
		Type: game.ActionVagabondSetup,
		VagabondSetup: &game.VagabondSetupAction{
			Faction:  game.Vagabond,
			ForestID: 7,
		},
	})

	if state.GamePhase != game.LifecyclePlaying {
		t.Fatalf("expected setup to transition into playing state, got %v", state.GamePhase)
	}
	if state.SetupStage != game.SetupStageComplete {
		t.Fatalf("expected setup stage complete, got %v", state.SetupStage)
	}
	if state.FactionTurn != game.Marquise || state.CurrentPhase != game.Birdsong || state.CurrentStep != game.StepBirdsong {
		t.Fatalf("expected Marquise birdsong start after setup, got faction=%v phase=%v step=%v", state.FactionTurn, state.CurrentPhase, state.CurrentStep)
	}
	if len(state.Marquise.CardsInHand) != 3 {
		t.Fatalf("expected player Marquise to start with 3 cards after setup, got %+v", state.Marquise.CardsInHand)
	}
	if state.OtherHandCounts[game.Eyrie] != 3 || state.OtherHandCounts[game.Alliance] != 3 || state.OtherHandCounts[game.Vagabond] != 3 {
		t.Fatalf("expected other hand counts to start at 3 after setup, got %+v", state.OtherHandCounts)
	}
	if len(state.Deck) != 39 {
		t.Fatalf("expected online deck to have 39 cards after final setup, got %d", len(state.Deck))
	}
}

func TestAssistSetupCreatesHiddenPlaceholdersByZone(t *testing.T) {
	state, err := SetupGame(SetupRequest{
		GameMode:          game.GameModeAssist,
		PlayerFaction:     game.Marquise,
		Factions:          []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		MapID:             game.AutumnMapID,
		VagabondCharacter: game.CharRanger,
		EyrieLeader:       game.LeaderBuilder,
	})
	if err != nil {
		t.Fatalf("expected setup to succeed, got %v", err)
	}

	state = ApplyAction(state, game.Action{
		Type: game.ActionMarquiseSetup,
		MarquiseSetup: &game.MarquiseSetupAction{
			Faction:             game.Marquise,
			KeepClearingID:      1,
			SawmillClearingID:   1,
			WorkshopClearingID:  5,
			RecruiterClearingID: 10,
		},
	})
	state = ApplyAction(state, game.Action{
		Type: game.ActionEyrieSetup,
		EyrieSetup: &game.EyrieSetupAction{
			Faction:    game.Eyrie,
			ClearingID: 3,
		},
	})
	state = ApplyAction(state, game.Action{
		Type: game.ActionVagabondSetup,
		VagabondSetup: &game.VagabondSetupAction{
			Faction:  game.Vagabond,
			ForestID: 7,
		},
	})

	if hiddenCardCount(state, game.Eyrie, game.HiddenCardZoneHand) != 3 {
		t.Fatalf("expected Eyrie hidden hand placeholders, got %+v", state.HiddenCards)
	}
	if hiddenCardCount(state, game.Vagabond, game.HiddenCardZoneHand) != 3 {
		t.Fatalf("expected Vagabond hidden hand placeholders, got %+v", state.HiddenCards)
	}
	if hiddenCardCount(state, game.Alliance, game.HiddenCardZoneSupporters) != 3 {
		t.Fatalf("expected Alliance hidden supporter placeholders, got %+v", state.HiddenCards)
	}
}

func TestValidEyrieSetupActionsPreferOppositeCorner(t *testing.T) {
	state, err := SetupGame(SetupRequest{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Eyrie,
		Factions:      []game.Faction{game.Marquise, game.Eyrie},
		MapID:         game.AutumnMapID,
		EyrieLeader:   game.LeaderDespot,
	})
	if err != nil {
		t.Fatalf("expected setup to succeed, got %v", err)
	}

	state = ApplyAction(state, game.Action{
		Type: game.ActionMarquiseSetup,
		MarquiseSetup: &game.MarquiseSetupAction{
			Faction:             game.Marquise,
			KeepClearingID:      1,
			SawmillClearingID:   1,
			WorkshopClearingID:  5,
			RecruiterClearingID: 10,
		},
	})

	actions := ValidActions(state)
	if len(actions) != 1 || actions[0].EyrieSetup == nil || actions[0].EyrieSetup.ClearingID != 3 {
		t.Fatalf("expected only opposite-corner Eyrie setup action, got %+v", actions)
	}
}

func TestSetupGameWithSameSeedProducesDeterministicRuinsAndDeck(t *testing.T) {
	setup := func() game.GameState {
		state, err := SetupGame(SetupRequest{
			GameMode:          game.GameModeOnline,
			PlayerFaction:     game.Marquise,
			Factions:          []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
			MapID:             game.AutumnMapID,
			VagabondCharacter: game.CharRanger,
			EyrieLeader:       game.LeaderBuilder,
			RandomSeed:        12345,
		})
		if err != nil {
			t.Fatalf("expected setup to succeed, got %v", err)
		}

		state = ApplyAction(state, game.Action{
			Type: game.ActionMarquiseSetup,
			MarquiseSetup: &game.MarquiseSetupAction{
				Faction:             game.Marquise,
				KeepClearingID:      1,
				SawmillClearingID:   1,
				WorkshopClearingID:  5,
				RecruiterClearingID: 10,
			},
		})
		state = ApplyAction(state, game.Action{
			Type: game.ActionEyrieSetup,
			EyrieSetup: &game.EyrieSetupAction{
				Faction:    game.Eyrie,
				ClearingID: 3,
			},
		})
		state = ApplyAction(state, game.Action{
			Type: game.ActionVagabondSetup,
			VagabondSetup: &game.VagabondSetupAction{
				Faction:  game.Vagabond,
				ForestID: 7,
			},
		})

		return state
	}

	first := setup()
	second := setup()

	if first.RandomSeed != second.RandomSeed || first.RandomSeed != 12345 {
		t.Fatalf("expected deterministic setups to preserve seed 12345, got %d and %d", first.RandomSeed, second.RandomSeed)
	}
	if len(first.Deck) == 0 || len(second.Deck) == 0 {
		t.Fatalf("expected deterministic setups to deal from a shuffled deck")
	}
	if first.Deck[0] != second.Deck[0] || first.Deck[1] != second.Deck[1] || first.Deck[2] != second.Deck[2] {
		t.Fatalf("expected same seed to preserve deck order, got %+v vs %+v", first.Deck[:3], second.Deck[:3])
	}
	for i := range first.Map.Clearings {
		if len(first.Map.Clearings[i].RuinItems) != len(second.Map.Clearings[i].RuinItems) {
			t.Fatalf("expected ruin item counts to match for clearing %d", first.Map.Clearings[i].ID)
		}
		if len(first.Map.Clearings[i].RuinItems) == 1 && first.Map.Clearings[i].RuinItems[0] != second.Map.Clearings[i].RuinItems[0] {
			t.Fatalf("expected ruin item order to match for clearing %d, got %v vs %v", first.Map.Clearings[i].ID, first.Map.Clearings[i].RuinItems, second.Map.Clearings[i].RuinItems)
		}
	}
}
