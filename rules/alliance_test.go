package rules

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidSpreadSympathyActionsRequiresAdjacencyAfterFirstToken(t *testing.T) {
	rabbitCard := firstCardOfSuit(t, game.Rabbit)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Adj:  []int{2},
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Adj:  []int{1},
				},
				{
					ID:   3,
					Suit: game.Rabbit,
				},
			},
		},
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Alliance: game.AllianceState{
			SympathyPlaced: 1,
			Supporters:     []game.Card{rabbitCard},
		},
	}

	got := ValidSpreadSympathyActions(state)
	want := game.Action{
		Type: game.ActionSpreadSympathy,
		SpreadSympathy: &game.SpreadSympathyAction{
			Faction:          game.Alliance,
			ClearingID:       2,
			SupporterCardIDs: []game.CardID{rabbitCard.ID},
		},
	}
	unwant := game.Action{
		Type: game.ActionSpreadSympathy,
		SpreadSympathy: &game.SpreadSympathyAction{
			Faction:          game.Alliance,
			ClearingID:       3,
			SupporterCardIDs: []game.CardID{rabbitCard.ID},
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected adjacent sympathy action %+v, got %+v", want, got)
	}
	if containsAction(got, unwant) {
		t.Fatalf("did not expect non-adjacent sympathy action %+v, got %+v", unwant, got)
	}
}

func TestValidSpreadSympathyActionsAppliesMartialLaw(t *testing.T) {
	foxCard := firstCardOfSuit(t, game.Fox)
	birdCard := firstCardOfSuit(t, game.Bird)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Warriors: map[game.Faction]int{
						game.Marquise: 3,
					},
				},
			},
		},
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Alliance: game.AllianceState{
			Supporters: []game.Card{foxCard, birdCard},
		},
	}

	got := ValidSpreadSympathyActions(state)
	want := game.Action{
		Type: game.ActionSpreadSympathy,
		SpreadSympathy: &game.SpreadSympathyAction{
			Faction:          game.Alliance,
			ClearingID:       1,
			SupporterCardIDs: []game.CardID{foxCard.ID, birdCard.ID},
		},
	}
	unwantSingle := game.Action{
		Type: game.ActionSpreadSympathy,
		SpreadSympathy: &game.SpreadSympathyAction{
			Faction:          game.Alliance,
			ClearingID:       1,
			SupporterCardIDs: []game.CardID{foxCard.ID},
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected martial-law sympathy action %+v, got %+v", want, got)
	}
	if containsAction(got, unwantSingle) {
		t.Fatalf("did not expect single-supporter martial-law action %+v, got %+v", unwantSingle, got)
	}
}

func TestValidSpreadSympathyActionsIgnoresAllianceWarriorsForMartialLaw(t *testing.T) {
	foxCard := firstCardOfSuit(t, game.Fox)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Warriors: map[game.Faction]int{
						game.Alliance: 3,
					},
				},
			},
		},
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Alliance: game.AllianceState{
			Supporters: []game.Card{foxCard},
		},
	}

	got := ValidSpreadSympathyActions(state)
	want := game.Action{
		Type: game.ActionSpreadSympathy,
		SpreadSympathy: &game.SpreadSympathyAction{
			Faction:          game.Alliance,
			ClearingID:       1,
			SupporterCardIDs: []game.CardID{foxCard.ID},
		},
	}
	if !containsAction(got, want) {
		t.Fatalf("expected alliance warriors not to trigger martial law %+v, got %+v", want, got)
	}
}

func TestValidRevoltActionsRequiresSympathyAndTwoSupporters(t *testing.T) {
	foxCard := firstCardOfSuit(t, game.Fox)
	birdCard := firstCardOfSuit(t, game.Bird)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   4,
					Suit: game.Fox,
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
			},
		},
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Alliance: game.AllianceState{
			Supporters: []game.Card{foxCard, birdCard},
		},
	}

	got := ValidRevoltActions(state)
	want := game.Action{
		Type: game.ActionRevolt,
		Revolt: &game.RevoltAction{
			Faction:          game.Alliance,
			ClearingID:       4,
			BaseSuit:         game.Fox,
			SupporterCardIDs: []game.CardID{foxCard.ID, birdCard.ID},
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected revolt action %+v, got %+v", want, got)
	}
}

func TestValidMobilizeActionsStopsAtSupporterLimitWithoutBase(t *testing.T) {
	foxCard := firstCardOfSuit(t, game.Fox)

	state := game.GameState{
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		Alliance: game.AllianceState{
			CardsInHand: []game.Card{foxCard},
			Supporters:  []game.Card{foxCard, foxCard, foxCard, foxCard, foxCard},
		},
	}

	got := ValidMobilizeActions(state)
	if len(got) != 0 {
		t.Fatalf("expected mobilize to stop at supporter limit, got %+v", got)
	}
}

func TestValidAllianceEveningActionsIncludesDrawAndOfficerActionChoices(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Adj:  []int{2},
					Warriors: map[game.Faction]int{
						game.Alliance: 1,
					},
					Buildings: []game.Building{
						{Faction: game.Alliance, Type: game.Base},
					},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Adj:  []int{1},
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		Alliance: game.AllianceState{
			Officers:      1,
			WarriorSupply: 1,
		},
	}

	got := ValidAllianceEveningActions(state)
	wantDraw := game.Action{
		Type: game.ActionEveningDraw,
		EveningDraw: &game.EveningDrawAction{
			Faction: game.Alliance,
			Count:   2,
		},
	}
	wantRecruit := game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Alliance,
			ClearingIDs: []int{1},
		},
	}
	wantMove := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Alliance,
			Count:    1,
			MaxCount: 1,
			From:     1,
			To:       2,
		},
	}

	if !containsAction(got, wantDraw) {
		t.Fatalf("expected draw action %+v, got %+v", wantDraw, got)
	}
	if !containsAction(got, wantRecruit) {
		t.Fatalf("expected recruit action %+v, got %+v", wantRecruit, got)
	}
	if !containsAction(got, wantMove) {
		t.Fatalf("expected move action %+v, got %+v", wantMove, got)
	}
}

func TestValidAllianceEveningDrawUsesBasesNotOfficers(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		Alliance: game.AllianceState{
			Officers: 3,
		},
	}

	got := ValidAllianceEveningActions(state)
	wantDraw := game.Action{
		Type: game.ActionEveningDraw,
		EveningDraw: &game.EveningDrawAction{
			Faction: game.Alliance,
			Count:   1,
		},
	}
	unwantedOfficerDraw := game.Action{
		Type: game.ActionEveningDraw,
		EveningDraw: &game.EveningDrawAction{
			Faction: game.Alliance,
			Count:   4,
		},
	}

	if !containsAction(got, wantDraw) {
		t.Fatalf("expected base draw action %+v, got %+v", wantDraw, got)
	}
	if containsAction(got, unwantedOfficerDraw) {
		t.Fatalf("expected officers not to increase draw count, got %+v", got)
	}
}

func TestValidAllianceEveningDrawCountsPlacedBases(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:        1,
					Suit:      game.Fox,
					Buildings: []game.Building{{Faction: game.Alliance, Type: game.Base}},
				},
				{
					ID:        2,
					Suit:      game.Rabbit,
					Buildings: []game.Building{{Faction: game.Alliance, Type: game.Base}},
				},
				{
					ID:        3,
					Suit:      game.Mouse,
					Buildings: []game.Building{{Faction: game.Alliance, Type: game.Base}},
				},
			},
		},
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
	}

	got := ValidAllianceEveningActions(state)
	wantDraw := game.Action{
		Type: game.ActionEveningDraw,
		EveningDraw: &game.EveningDrawAction{
			Faction: game.Alliance,
			Count:   4,
		},
	}

	if !containsAction(got, wantDraw) {
		t.Fatalf("expected draw action to count placed bases %+v, got %+v", wantDraw, got)
	}
}

func TestValidBattlesTargetsAllianceSympathy(t *testing.T) {
	board := game.Map{
		Clearings: []game.Clearing{
			{
				ID: 1,
				Warriors: map[game.Faction]int{
					game.Marquise: 1,
				},
				Tokens: []game.Token{
					{Faction: game.Alliance, Type: game.TokenSympathy},
				},
			},
		},
	}

	got := ValidBattles(game.Marquise, board)
	want := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Alliance,
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected sympathy battle target %+v, got %+v", want, got)
	}
}
