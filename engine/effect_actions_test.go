package engine

import (
	"reflect"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidActionsIncludesBetterBurrowBankTargetsAtBirdsong(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance},
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {15},
		},
	}

	got := ValidActions(state)

	wantEyrie := game.Action{
		Type: game.ActionUsePersistentEffect,
		UsePersistentEffect: &game.UsePersistentEffectAction{
			Faction:       game.Marquise,
			EffectID:      "better_burrow_bank",
			TargetFaction: game.Eyrie,
		},
	}
	wantAlliance := game.Action{
		Type: game.ActionUsePersistentEffect,
		UsePersistentEffect: &game.UsePersistentEffectAction{
			Faction:       game.Marquise,
			EffectID:      "better_burrow_bank",
			TargetFaction: game.Alliance,
		},
	}

	if !containsAction(got, wantEyrie) || !containsAction(got, wantAlliance) {
		t.Fatalf("expected Better Burrow Bank targets, got %+v", got)
	}
}

func TestApplyActionUseBetterBurrowBankAssistTracksHiddenHands(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Vagabond,
		OtherHandCounts: map[game.Faction]int{
			game.Eyrie:    2,
			game.Alliance: 1,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionUsePersistentEffect,
		UsePersistentEffect: &game.UsePersistentEffectAction{
			Faction:       game.Eyrie,
			EffectID:      "better_burrow_bank",
			TargetFaction: game.Alliance,
		},
	})

	if next.OtherHandCounts[game.Eyrie] != 3 || next.OtherHandCounts[game.Alliance] != 2 {
		t.Fatalf("expected Better Burrow Bank to increment hidden hand counts, got %+v", next.OtherHandCounts)
	}
	if !reflect.DeepEqual(next.TurnProgress.UsedPersistentEffectIDs, []string{"better_burrow_bank"}) {
		t.Fatalf("expected Better Burrow Bank to be marked used, got %+v", next.TurnProgress.UsedPersistentEffectIDs)
	}
}

func TestValidActionsIncludesCommandWarrenBattle(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Rabbit,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
						game.Eyrie:    1,
					},
				},
			},
		},
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {19},
		},
	}

	want := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:        game.Marquise,
			ClearingID:     1,
			TargetFaction:  game.Eyrie,
			SourceEffectID: "command_warren",
		},
	}

	if !containsAction(ValidActions(state), want) {
		t.Fatalf("expected Command Warren battle action, got %+v", ValidActions(state))
	}
}

func TestApplyCommandWarrenBattleDoesNotSpendNormalAction(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Rabbit,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
						game.Eyrie:    1,
					},
				},
			},
		},
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {19},
		},
	}

	battle := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:        game.Marquise,
			ClearingID:     1,
			TargetFaction:  game.Eyrie,
			SourceEffectID: "command_warren",
		},
	}

	resolved := ResolveBattle(state, battle, 1, 0)
	next := ApplyAction(state, resolved)

	if next.TurnProgress.ActionsUsed != 0 {
		t.Fatalf("expected Command Warren battle not to spend Marquise daylight action, got %+v", next.TurnProgress)
	}
	if !reflect.DeepEqual(next.TurnProgress.UsedPersistentEffectIDs, []string{"command_warren"}) {
		t.Fatalf("expected Command Warren to be marked used, got %+v", next.TurnProgress.UsedPersistentEffectIDs)
	}
}

func TestApplyCobblerMoveStaysInEveningAndDoesNotSpendMarch(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Adj:  []int{2},
					Suit: game.Rabbit,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
				{
					ID:   2,
					Adj:  []int{1},
					Suit: game.Fox,
				},
			},
		},
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {17},
		},
	}

	move := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:        game.Marquise,
			Count:          1,
			MaxCount:       1,
			From:           1,
			To:             2,
			SourceEffectID: "cobbler",
		},
	}

	next := ApplyAction(state, move)

	if next.Map.Clearings[0].Warriors[game.Marquise] != 0 || next.Map.Clearings[1].Warriors[game.Marquise] != 1 {
		t.Fatalf("expected Cobbler move to move warrior, got %+v", next.Map.Clearings)
	}
	if next.CurrentPhase != game.Evening || next.CurrentStep != game.StepEvening {
		t.Fatalf("expected Cobbler move to remain in evening, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
	if next.TurnProgress.ActionsUsed != 0 || next.TurnProgress.MarchesUsed != 0 {
		t.Fatalf("expected Cobbler move not to spend Marquise daylight actions, got %+v", next.TurnProgress)
	}
	if !reflect.DeepEqual(next.TurnProgress.UsedPersistentEffectIDs, []string{"cobbler"}) {
		t.Fatalf("expected Cobbler to be marked used, got %+v", next.TurnProgress.UsedPersistentEffectIDs)
	}
}

func TestApplyActionUseTaxCollectorRemovesWarriorAndDrawsForHiddenFaction(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Vagabond,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 2,
					},
				},
			},
		},
		OtherHandCounts: map[game.Faction]int{
			game.Marquise: 1,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionUsePersistentEffect,
		UsePersistentEffect: &game.UsePersistentEffectAction{
			Faction:    game.Marquise,
			EffectID:   "tax_collector",
			ClearingID: 1,
		},
	})

	if next.Map.Clearings[0].Warriors[game.Marquise] != 1 {
		t.Fatalf("expected Tax Collector to remove one warrior, got %+v", next.Map.Clearings[0].Warriors)
	}
	if next.OtherHandCounts[game.Marquise] != 2 {
		t.Fatalf("expected Tax Collector to draw for hidden faction, got %+v", next.OtherHandCounts)
	}
}

func TestApplyActionUseRoyalClaimScoresRuledClearingsAndDiscards(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Sawmill},
					},
				},
				{
					ID:   3,
					Suit: game.Mouse,
					Warriors: map[game.Faction]int{
						game.Eyrie: 1,
					},
				},
			},
		},
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {7},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionUsePersistentEffect,
		UsePersistentEffect: &game.UsePersistentEffectAction{
			Faction:  game.Marquise,
			EffectID: "royal_claim",
		},
	})

	if next.VictoryPoints[game.Marquise] != 2 {
		t.Fatalf("expected Royal Claim to score ruled clearings, got %+v", next.VictoryPoints)
	}
	if len(next.PersistentEffects) != 0 {
		t.Fatalf("expected Royal Claim to leave persistent effects, got %+v", next.PersistentEffects)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 7 {
		t.Fatalf("expected Royal Claim to be discarded, got %+v", next.DiscardPile)
	}
}

func TestValidActionsIncludesStandAndDeliverTargetsAtBirdsong(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance},
		OtherHandCounts: map[game.Faction]int{
			game.Eyrie:    1,
			game.Alliance: 0,
		},
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {41},
		},
	}

	want := game.Action{
		Type: game.ActionUsePersistentEffect,
		UsePersistentEffect: &game.UsePersistentEffectAction{
			Faction:       game.Marquise,
			EffectID:      "stand_and_deliver",
			TargetFaction: game.Eyrie,
		},
	}

	if !containsAction(ValidActions(state), want) {
		t.Fatalf("expected Stand and Deliver target, got %+v", ValidActions(state))
	}
}

func TestValidActionsIncludesCodebreakersTargetsInEvening(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance},
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {28},
		},
	}

	want := game.Action{
		Type: game.ActionUsePersistentEffect,
		UsePersistentEffect: &game.UsePersistentEffectAction{
			Faction:       game.Marquise,
			EffectID:      "codebreakers",
			TargetFaction: game.Eyrie,
		},
	}

	if !containsAction(ValidActions(state), want) {
		t.Fatalf("expected Codebreakers target, got %+v", ValidActions(state))
	}
}

func TestApplyActionDetailedCodebreakersRevealsTrackedHand(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{
				{ID: 8, Name: "Birdy Bindle"},
				{ID: 12, Name: "Ambush"},
			},
		},
	}

	next, result := ApplyActionDetailed(state, game.Action{
		Type: game.ActionUsePersistentEffect,
		UsePersistentEffect: &game.UsePersistentEffectAction{
			Faction:       game.Marquise,
			EffectID:      "codebreakers",
			TargetFaction: game.Eyrie,
		},
	})

	if result == nil || len(result.Cards) != 2 {
		t.Fatalf("expected Codebreakers to reveal tracked hand, got %+v", result)
	}
	if !reflect.DeepEqual(next.TurnProgress.UsedPersistentEffectIDs, []string{"codebreakers"}) {
		t.Fatalf("expected Codebreakers to be marked used, got %+v", next.TurnProgress.UsedPersistentEffectIDs)
	}
}

func TestApplyActionDetailedStandAndDeliverTransfersRandomOnlineCard(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
		RandomSeed:    5,
		Marquise:      game.MarquiseState{},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{
				{ID: 8, Name: "Birdy Bindle"},
			},
		},
	}

	next, result := ApplyActionDetailed(state, game.Action{
		Type: game.ActionUsePersistentEffect,
		UsePersistentEffect: &game.UsePersistentEffectAction{
			Faction:       game.Marquise,
			EffectID:      "stand_and_deliver",
			TargetFaction: game.Eyrie,
		},
	})

	if result == nil || len(result.Cards) != 1 || result.Cards[0].ID != 8 {
		t.Fatalf("expected Stand and Deliver to report transferred card, got %+v", result)
	}
	if len(next.Marquise.CardsInHand) != 1 || next.Marquise.CardsInHand[0].ID != 8 {
		t.Fatalf("expected transferred card in actor hand, got %+v", next.Marquise.CardsInHand)
	}
	if len(next.Eyrie.CardsInHand) != 0 {
		t.Fatalf("expected target hand to lose transferred card, got %+v", next.Eyrie.CardsInHand)
	}
	if next.VictoryPoints[game.Eyrie] != 1 {
		t.Fatalf("expected target to score 1 VP, got %+v", next.VictoryPoints)
	}
}
