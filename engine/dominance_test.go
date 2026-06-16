package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidActionsIncludesActivateDominanceAtTenPoints(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		VictoryPoints: map[game.Faction]int{
			game.Marquise: 10,
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 14, Name: "Dominance", Suit: game.Bird, Kind: game.DominanceCard},
			},
		},
	}

	actions := ValidActions(state)
	found := false
	for _, action := range actions {
		if action.Type == game.ActionActivateDominance && action.ActivateDominance != nil && action.ActivateDominance.CardID == 14 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected activate dominance action, got %+v", actions)
	}
}

func TestApplyActivateDominancePreventsFutureScoring(t *testing.T) {
	state := game.GameState{
		GamePhase: game.LifecyclePlaying,
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 14, Name: "Dominance", Suit: game.Bird, Kind: game.DominanceCard},
			},
		},
		VictoryPoints: map[game.Faction]int{
			game.Marquise: 10,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionActivateDominance,
		ActivateDominance: &game.ActivateDominanceAction{
			Faction: game.Marquise,
			CardID:  14,
		},
	})

	if !hasActiveDominance(next, game.Marquise) {
		t.Fatalf("expected active dominance to be tracked, got %+v", next.ActiveDominance)
	}
	addVictoryPoints(&next, game.Marquise, 5)
	if next.VictoryPoints[game.Marquise] != 10 {
		t.Fatalf("expected active dominance to block future scoring, got %+v", next.VictoryPoints)
	}
}

func TestValidActionsDoesNotOfferSecondActiveDominance(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		ActiveDominance: map[game.Faction]game.CardID{
			game.Marquise: 14,
		},
		VictoryPoints: map[game.Faction]int{
			game.Marquise: 12,
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 27, Name: "Dominance", Suit: game.Rabbit, Kind: game.DominanceCard},
			},
		},
	}

	for _, action := range ValidActions(state) {
		if action.Type == game.ActionActivateDominance {
			t.Fatalf("expected active dominance to block replacement activation, got %+v", action)
		}
	}
}

func TestDiscardDominanceCardMakesItAvailable(t *testing.T) {
	state := game.GameState{}

	DiscardCard(&state, 14)

	if len(state.AvailableDominance) != 1 || state.AvailableDominance[0] != 14 {
		t.Fatalf("expected discarded dominance card to become available, got %+v", state.AvailableDominance)
	}
	if len(state.DiscardPile) != 0 {
		t.Fatalf("expected dominance card to stay out of discard pile, got %+v", state.DiscardPile)
	}
}

func TestApplyTakeDominanceSpendsMatchingCard(t *testing.T) {
	state := game.GameState{
		GamePhase:          game.LifecyclePlaying,
		AvailableDominance: []game.CardID{27},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 24, Name: "A Visit to Friends", Suit: game.Rabbit},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionTakeDominance,
		TakeDominance: &game.TakeDominanceAction{
			Faction:         game.Marquise,
			DominanceCardID: 27,
			SpentCardID:     24,
		},
	})

	if len(next.AvailableDominance) != 0 {
		t.Fatalf("expected available dominance to be taken, got %+v", next.AvailableDominance)
	}
	if len(next.Marquise.CardsInHand) != 1 || next.Marquise.CardsInHand[0].ID != 27 {
		t.Fatalf("expected dominance card to enter hand, got %+v", next.Marquise.CardsInHand)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 24 {
		t.Fatalf("expected spent matching card to be discarded, got %+v", next.DiscardPile)
	}
}

func TestValidActionsTakeDominanceAllowsBirdCardAsWild(t *testing.T) {
	state := game.GameState{
		GamePhase:          game.LifecyclePlaying,
		FactionTurn:        game.Marquise,
		CurrentPhase:       game.Daylight,
		CurrentStep:        game.StepDaylightActions,
		AvailableDominance: []game.CardID{27},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 1, Name: "Armorers", Suit: game.Bird},
			},
		},
	}

	actions := ValidActions(state)
	for _, action := range actions {
		if action.Type == game.ActionTakeDominance &&
			action.TakeDominance != nil &&
			action.TakeDominance.DominanceCardID == 27 &&
			action.TakeDominance.SpentCardID == 1 {
			return
		}
	}
	t.Fatalf("expected bird card to take rabbit dominance, got %+v", actions)
}

func TestApplyTakeDominanceAllowsBirdCardAsWild(t *testing.T) {
	state := game.GameState{
		GamePhase:          game.LifecyclePlaying,
		AvailableDominance: []game.CardID{27},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 1, Name: "Armorers", Suit: game.Bird},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionTakeDominance,
		TakeDominance: &game.TakeDominanceAction{
			Faction:         game.Marquise,
			DominanceCardID: 27,
			SpentCardID:     1,
		},
	})

	if len(next.AvailableDominance) != 0 {
		t.Fatalf("expected available dominance to be taken, got %+v", next.AvailableDominance)
	}
	if len(next.Marquise.CardsInHand) != 1 || next.Marquise.CardsInHand[0].ID != 27 {
		t.Fatalf("expected rabbit dominance card to enter hand, got %+v", next.Marquise.CardsInHand)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 1 {
		t.Fatalf("expected spent bird card to be discarded, got %+v", next.DiscardPile)
	}
}

func TestApplyTakeDominanceRejectsOffSuitNonBirdCard(t *testing.T) {
	state := game.GameState{
		GamePhase:          game.LifecyclePlaying,
		AvailableDominance: []game.CardID{27},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 28, Name: "Codebreakers", Suit: game.Mouse},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionTakeDominance,
		TakeDominance: &game.TakeDominanceAction{
			Faction:         game.Marquise,
			DominanceCardID: 27,
			SpentCardID:     28,
		},
	})

	if len(next.AvailableDominance) != 1 || next.AvailableDominance[0] != 27 {
		t.Fatalf("expected rabbit dominance to remain available, got %+v", next.AvailableDominance)
	}
	if len(next.Marquise.CardsInHand) != 1 || next.Marquise.CardsInHand[0].ID != 28 {
		t.Fatalf("expected off-suit card to remain in hand, got %+v", next.Marquise.CardsInHand)
	}
	if len(next.DiscardPile) != 0 {
		t.Fatalf("expected no discard for invalid dominance take, got %+v", next.DiscardPile)
	}
}

func TestBeginNextFactionTurnChecksDominanceVictory(t *testing.T) {
	state := game.GameState{
		GamePhase:   game.LifecyclePlaying,
		FactionTurn: game.Alliance,
		TurnOrder:   []game.Faction{game.Alliance, game.Marquise},
		ActiveDominance: map[game.Faction]game.CardID{
			game.Marquise: 14,
		},
		Map: game.Map{
			ID: game.AutumnMapID,
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
				{
					ID:   3,
					Suit: game.Rabbit,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
	}

	beginNextFactionTurn(&state)

	if state.GamePhase != game.LifecycleGameOver || state.Winner != game.Marquise {
		t.Fatalf("expected bird dominance to win at start of birdsong, got phase=%v winner=%v", state.GamePhase, state.Winner)
	}
	if got := ValidActions(state); len(got) != 0 {
		t.Fatalf("expected no legal actions after dominance win, got %+v", got)
	}
}

func TestBirdDominanceAcceptsSecondAutumnOppositeCornerPair(t *testing.T) {
	state := game.GameState{
		GamePhase:   game.LifecyclePlaying,
		FactionTurn: game.Alliance,
		TurnOrder:   []game.Faction{game.Alliance, game.Marquise},
		ActiveDominance: map[game.Faction]game.CardID{
			game.Marquise: 14,
		},
		Map: game.Map{
			ID: game.AutumnMapID,
			Clearings: []game.Clearing{
				{
					ID:   2,
					Suit: game.Mouse,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
				{
					ID:   4,
					Suit: game.Rabbit,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
	}

	beginNextFactionTurn(&state)

	if state.GamePhase != game.LifecycleGameOver || state.Winner != game.Marquise {
		t.Fatalf("expected bird dominance to win with corners 2 and 4, got phase=%v winner=%v", state.GamePhase, state.Winner)
	}
}

func TestBirdDominanceRequiresOppositeAutumnCorners(t *testing.T) {
	state := game.GameState{
		GamePhase:   game.LifecyclePlaying,
		FactionTurn: game.Alliance,
		TurnOrder:   []game.Faction{game.Alliance, game.Marquise},
		ActiveDominance: map[game.Faction]game.CardID{
			game.Marquise: 14,
		},
		Map: game.Map{
			ID: game.AutumnMapID,
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
					Suit: game.Mouse,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
	}

	beginNextFactionTurn(&state)

	if state.GamePhase == game.LifecycleGameOver {
		t.Fatalf("did not expect bird dominance to win with non-opposite corners, got phase=%v winner=%v", state.GamePhase, state.Winner)
	}
}

func TestValidActionsIncludesVagabondCoalitionTarget(t *testing.T) {
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		VictoryPoints: map[game.Faction]int{
			game.Marquise: 12,
			game.Eyrie:    7,
			game.Alliance: 7,
			game.Vagabond: 10,
		},
		Vagabond: game.VagabondState{
			CardsInHand: []game.Card{
				{ID: 40, Name: "Dominance", Suit: game.Mouse, Kind: game.DominanceCard},
			},
		},
	}

	actions := ValidActions(state)
	count := 0
	for _, action := range actions {
		if action.Type == game.ActionActivateDominance && action.ActivateDominance != nil {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected coalition target actions for tied lowest players, got %+v", actions)
	}
}

func TestApplyVagabondCoalitionWithHostileFactionResetsRelationship(t *testing.T) {
	state := game.GameState{
		GamePhase: game.LifecyclePlaying,
		Vagabond: game.VagabondState{
			CardsInHand: []game.Card{
				{ID: 40, Name: "Dominance", Suit: game.Mouse, Kind: game.DominanceCard},
			},
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelHostile,
				game.Eyrie:    game.RelFriendly,
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionActivateDominance,
		ActivateDominance: &game.ActivateDominanceAction{
			Faction:       game.Vagabond,
			CardID:        40,
			TargetFaction: game.Marquise,
		},
	})

	if !next.CoalitionActive || next.CoalitionPartner != game.Marquise {
		t.Fatalf("expected Vagabond coalition with Marquise, got active=%v partner=%v", next.CoalitionActive, next.CoalitionPartner)
	}
	if next.Vagabond.Relationships[game.Marquise] != game.RelIndifferent {
		t.Fatalf("expected hostile coalition partner to reset to indifferent, got %+v", next.Vagabond.Relationships)
	}
	if next.Vagabond.Relationships[game.Eyrie] != game.RelFriendly {
		t.Fatalf("expected unrelated relationships to remain unchanged, got %+v", next.Vagabond.Relationships)
	}
}

func TestApplyVagabondCoalitionPreservesNonHostileRelationship(t *testing.T) {
	state := game.GameState{
		GamePhase: game.LifecyclePlaying,
		Vagabond: game.VagabondState{
			CardsInHand: []game.Card{
				{ID: 40, Name: "Dominance", Suit: game.Mouse, Kind: game.DominanceCard},
			},
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelFriendly,
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionActivateDominance,
		ActivateDominance: &game.ActivateDominanceAction{
			Faction:       game.Vagabond,
			CardID:        40,
			TargetFaction: game.Marquise,
		},
	})

	if next.Vagabond.Relationships[game.Marquise] != game.RelFriendly {
		t.Fatalf("expected non-hostile coalition partner relationship to remain friendly, got %+v", next.Vagabond.Relationships)
	}
}

func TestCoalitionSharesPartnerVictory(t *testing.T) {
	state := game.GameState{
		GamePhase:        game.LifecyclePlaying,
		CoalitionActive:  true,
		CoalitionPartner: game.Marquise,
		ActiveDominance: map[game.Faction]game.CardID{
			game.Vagabond: 14,
		},
		VictoryPoints: map[game.Faction]int{
			game.Marquise: 29,
		},
	}

	before := victoryPointSnapshot(state)
	addVictoryPoints(&state, game.Marquise, 1)
	resolveVictoryPointWin(&state, before)

	if state.GamePhase != game.LifecycleGameOver || state.Winner != game.Marquise {
		t.Fatalf("expected coalition partner victory to end the game, got phase=%v winner=%v", state.GamePhase, state.Winner)
	}
	if len(state.WinningCoalition) != 2 || state.WinningCoalition[0] != game.Marquise || state.WinningCoalition[1] != game.Vagabond {
		t.Fatalf("expected coalition winners to include partner and Vagabond, got %+v", state.WinningCoalition)
	}
	if got := ValidActions(state); len(got) != 0 {
		t.Fatalf("expected no legal actions after coalition win, got %+v", got)
	}
}
