package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidateStateAcceptsActiveSetupStage(t *testing.T) {
	state := game.GameState{
		GamePhase:  game.LifecycleSetup,
		SetupStage: game.SetupStageMarquise,
	}

	if err := ValidateState(state); err != nil {
		t.Fatalf("expected setup stage to validate, got %v", err)
	}
}

func TestValidateStateRejectsSetupWithActiveStep(t *testing.T) {
	state := game.GameState{
		GamePhase:   game.LifecycleSetup,
		SetupStage:  game.SetupStageMarquise,
		CurrentStep: game.StepBirdsong,
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected setup state with active step to fail validation")
	}
}

func TestValidateStateRejectsNegativeOtherHandCount(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		OtherHandCounts: map[game.Faction]int{
			game.Eyrie: -1,
		},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected negative other-hand count to fail validation")
	}
}

func TestValidateStateRejectsCoalitionWithoutVagabondDominance(t *testing.T) {
	state := game.GameState{
		GamePhase:        game.LifecyclePlaying,
		SetupStage:       game.SetupStageComplete,
		CurrentPhase:     game.Birdsong,
		CurrentStep:      game.StepBirdsong,
		CoalitionActive:  true,
		CoalitionPartner: game.Marquise,
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected coalition without active Vagabond dominance to fail validation")
	}
}

func TestValidateStateRejectsDuplicateAvailableDominance(t *testing.T) {
	state := game.GameState{
		GamePhase:          game.LifecyclePlaying,
		SetupStage:         game.SetupStageComplete,
		CurrentPhase:       game.Birdsong,
		CurrentStep:        game.StepBirdsong,
		AvailableDominance: []game.CardID{14, 14},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected duplicate available dominance cards to fail validation")
	}
}

func TestValidateStateRejectsAssistPlaceholderCountMismatch(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		GamePhase:     game.LifecyclePlaying,
		SetupStage:    game.SetupStageComplete,
		CurrentPhase:  game.Birdsong,
		CurrentStep:   game.StepBirdsong,
		PlayerFaction: game.Marquise,
		OtherHandCounts: map[game.Faction]int{
			game.Eyrie: 2,
		},
		HiddenCards: []game.HiddenCard{
			{ID: 1, OwnerFaction: game.Eyrie, Zone: game.HiddenCardZoneHand},
		},
		NextHiddenCardID: 2,
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected assist placeholder mismatch to fail validation")
	}
}

func TestApplyActionProducesValidState(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
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
		Marquise: game.MarquiseState{
			WarriorSupply:    1,
			RecruitersPlaced: 1,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
		},
	})

	if err := ValidateState(next); err != nil {
		t.Fatalf("expected applied recruit state to validate, got %v", err)
	}
}

func TestValidateStateRejectsNegativeBoardCounts(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:       1,
					Suit:     game.Fox,
					Wood:     -1,
					Warriors: map[game.Faction]int{game.Marquise: -1},
				},
			},
		},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected negative board counts to fail validation")
	}
}

func TestValidateStateRejectsOverfilledBuildingSlots(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:         1,
					Suit:       game.Fox,
					BuildSlots: 1,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Sawmill},
						{Faction: game.Marquise, Type: game.Workshop},
					},
				},
			},
		},
		Marquise: game.MarquiseState{
			SawmillsPlaced:  1,
			WorkshopsPlaced: 1,
		},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected overfilled clearing to fail validation")
	}
}

func TestValidateStateRejectsWarriorSupplyLimitOverflow(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:       1,
					Suit:     game.Fox,
					Warriors: map[game.Faction]int{game.Marquise: 25},
				},
			},
		},
		Marquise: game.MarquiseState{
			WarriorSupply: 1,
		},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected warrior supply overflow to fail validation")
	}
}

func TestValidateStateRejectsFactionCounterMismatch(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:         1,
					Suit:       game.Fox,
					BuildSlots: 1,
					Buildings:  []game.Building{{Faction: game.Marquise, Type: game.Sawmill}},
				},
			},
		},
		Marquise: game.MarquiseState{
			SawmillsPlaced: 2,
		},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected faction counter mismatch to fail validation")
	}
}

func TestValidateStateRejectsAllianceBaseFlagMismatch(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Alliance: game.AllianceState{
			FoxBasePlaced: true,
		},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected Alliance base flag without board base to fail validation")
	}
}

func TestValidateStateRejectsInvalidVagabondItemZone(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Vagabond: game.VagabondState{
			Items: []game.Item{
				{Type: game.ItemTea, Status: game.ItemReady, Zone: game.ItemZoneSatchel},
			},
		},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected invalid Vagabond item zone to fail validation")
	}
}

func TestValidateStateRejectsDuplicateQuestZones(t *testing.T) {
	quest := scenarioQuest(t, 1)
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		QuestDeck:    []game.QuestID{quest.ID},
		Vagabond: game.VagabondState{
			QuestsAvailable: []game.Quest{quest},
		},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected duplicate quest zones to fail validation")
	}
}

func TestValidateStateRejectsDuplicateKnownCardZones(t *testing.T) {
	card := scenarioCard(t, game.Fox, game.ItemCard)
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Deck:         []game.CardID{card.ID},
		DiscardPile:  []game.CardID{card.ID},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected duplicate known card zones to fail validation")
	}
}

func TestValidateStateRejectsInvalidTurnOrderAndTerminalCoalition(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		FactionTurn:  game.Marquise,
		TurnOrder:    []game.Faction{game.Eyrie},
	}

	if err := ValidateState(state); err == nil {
		t.Fatalf("expected active faction outside turn order to fail validation")
	}

	state = game.GameState{
		GamePhase:        game.LifecyclePlaying,
		SetupStage:       game.SetupStageComplete,
		CurrentPhase:     game.Birdsong,
		CurrentStep:      game.StepBirdsong,
		WinningCoalition: []game.Faction{game.Marquise, game.Vagabond},
	}
	if err := ValidateState(state); err == nil {
		t.Fatalf("expected winning coalition before game over to fail validation")
	}
}
