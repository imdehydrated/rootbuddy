package rules

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

func firstCardOfSuit(t *testing.T, suit game.Suit) game.Card {
	t.Helper()

	for _, card := range carddata.BaseDeck() {
		if card.Suit == suit {
			return card
		}
	}

	t.Fatalf("no card found for suit %v", suit)
	return game.Card{}
}

func TestValidAddToDecreeActions(t *testing.T) {
	foxCard := firstCardOfSuit(t, game.Fox)
	birdCard := firstCardOfSuit(t, game.Bird)

	state := game.GameState{
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{foxCard, birdCard},
		},
	}

	got := ValidAddToDecreeActions(state)

	wantSingle := game.Action{
		Type: game.ActionAddToDecree,
		AddToDecree: &game.AddToDecreeAction{
			Faction: game.Eyrie,
			CardIDs: []game.CardID{foxCard.ID},
			Columns: []game.DecreeColumn{game.DecreeRecruit},
		},
	}
	wantPair := game.Action{
		Type: game.ActionAddToDecree,
		AddToDecree: &game.AddToDecreeAction{
			Faction: game.Eyrie,
			CardIDs: []game.CardID{foxCard.ID, birdCard.ID},
			Columns: []game.DecreeColumn{game.DecreeRecruit, game.DecreeMove},
		},
	}

	if !containsAction(got, wantSingle) {
		t.Fatalf("expected single-card decree action %+v, got %+v", wantSingle, got)
	}
	if !containsAction(got, wantPair) {
		t.Fatalf("expected two-card decree action %+v, got %+v", wantPair, got)
	}
}

func TestValidEyrieRecruitActionsCharismaticRecruitsTwoWarriors(t *testing.T) {
	foxCard := firstCardOfSuit(t, game.Fox)
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Buildings: []game.Building{
						{Faction: game.Eyrie, Type: game.Roost},
					},
				},
			},
		},
		Eyrie: game.EyrieState{
			Leader:        game.LeaderCharismatic,
			WarriorSupply: 2,
		},
	}

	got := ValidEyrieRecruitActions(state, foxCard.ID)
	want := game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:      game.Eyrie,
			ClearingIDs:  []int{1, 1},
			DecreeCardID: foxCard.ID,
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected charismatic recruit action %+v, got %+v", want, got)
	}
}

func TestValidEyrieDaylightActionsReturnsTurmoilWhenCurrentCardIsUnresolvable(t *testing.T) {
	foxCard := firstCardOfSuit(t, game.Fox)
	state := game.GameState{
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		Eyrie: game.EyrieState{
			Leader:           game.LeaderCommander,
			AvailableLeaders: []game.EyrieLeader{game.LeaderBuilder, game.LeaderDespot},
			Decree: game.Decree{
				Recruit: []game.CardID{foxCard.ID},
			},
		},
	}

	got := ValidEyrieDaylightActions(state)
	want := game.Action{
		Type: game.ActionTurmoil,
		Turmoil: &game.TurmoilAction{
			Faction:   game.Eyrie,
			NewLeader: game.LeaderBuilder,
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected turmoil action %+v, got %+v", want, got)
	}
}

func TestValidEyrieDaylightActionsRecyclesLeadersWhenNoneAreAvailable(t *testing.T) {
	foxCard := firstCardOfSuit(t, game.Fox)
	state := game.GameState{
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		Eyrie: game.EyrieState{
			Leader: game.LeaderDespot,
			Decree: game.Decree{
				Recruit: []game.CardID{foxCard.ID},
			},
		},
	}

	got := ValidEyrieDaylightActions(state)
	wantLeaders := []game.EyrieLeader{
		game.LeaderBuilder,
		game.LeaderCharismatic,
		game.LeaderCommander,
	}
	unwanted := game.Action{
		Type: game.ActionTurmoil,
		Turmoil: &game.TurmoilAction{
			Faction:   game.Eyrie,
			NewLeader: game.LeaderDespot,
		},
	}

	for _, leader := range wantLeaders {
		want := game.Action{
			Type: game.ActionTurmoil,
			Turmoil: &game.TurmoilAction{
				Faction:   game.Eyrie,
				NewLeader: leader,
			},
		}
		if !containsAction(got, want) {
			t.Fatalf("expected recycled turmoil action %+v, got %+v", want, got)
		}
	}
	if containsAction(got, unwanted) {
		t.Fatalf("did not expect current leader to be available after recycling, got %+v", got)
	}
}

func TestValidEyrieDaylightActionsAllowsResolvingCardsInAnyOrderWithinColumn(t *testing.T) {
	foxCard := firstCardOfSuit(t, game.Fox)
	rabbitCard := firstCardOfSuit(t, game.Rabbit)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   2,
					Suit: game.Rabbit,
					Buildings: []game.Building{
						{Faction: game.Eyrie, Type: game.Roost},
					},
				},
			},
		},
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		Eyrie: game.EyrieState{
			Leader:           game.LeaderBuilder,
			AvailableLeaders: []game.EyrieLeader{game.LeaderCommander},
			WarriorSupply:    1,
			Decree: game.Decree{
				Recruit: []game.CardID{foxCard.ID, rabbitCard.ID},
			},
		},
	}

	got := ValidEyrieDaylightActions(state)
	want := game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:      game.Eyrie,
			ClearingIDs:  []int{2},
			DecreeCardID: rabbitCard.ID,
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected rabbit recruit action %+v, got %+v", want, got)
	}

	for _, action := range got {
		if action.Type == game.ActionTurmoil {
			t.Fatalf("did not expect turmoil while another card in the column is still resolvable, got %+v", got)
		}
	}
}

func TestValidEyrieEveningActionsScoreThenDraw(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		Eyrie: game.EyrieState{
			RoostsPlaced: 3,
		},
	}

	got := ValidEyrieEveningActions(state)
	wantScore := game.Action{
		Type: game.ActionScoreRoosts,
		ScoreRoosts: &game.ScoreRoostsAction{
			Faction: game.Eyrie,
			Points:  2,
		},
	}
	if !containsAction(got, wantScore) {
		t.Fatalf("expected score-roosts action %+v, got %+v", wantScore, got)
	}

	state.TurnProgress.EveningMainActionTaken = true
	got = ValidEyrieEveningActions(state)
	wantDraw := game.Action{
		Type: game.ActionEveningDraw,
		EveningDraw: &game.EveningDrawAction{
			Faction: game.Eyrie,
			Count:   2,
		},
	}
	if !containsAction(got, wantDraw) {
		t.Fatalf("expected eyrie evening draw action %+v, got %+v", wantDraw, got)
	}
}
