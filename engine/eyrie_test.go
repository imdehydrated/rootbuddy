package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

func firstEyrieTestCard(t *testing.T, suit game.Suit) game.Card {
	t.Helper()

	for _, card := range carddata.BaseDeck() {
		if card.Suit == suit {
			return card
		}
	}

	t.Fatalf("no card found for suit %v", suit)
	return game.Card{}
}

func TestValidActionsEyrieBirdsongReturnsAddToDecreeAction(t *testing.T) {
	foxCard := firstEyrieTestCard(t, game.Fox)
	state := game.GameState{
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{foxCard},
		},
	}

	got := ValidActions(state)
	want := game.Action{
		Type: game.ActionAddToDecree,
		AddToDecree: &game.AddToDecreeAction{
			Faction: game.Eyrie,
			CardIDs: []game.CardID{foxCard.ID},
			Columns: []game.DecreeColumn{game.DecreeRecruit},
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected add-to-decree action %+v, got %+v", want, got)
	}
}

func TestApplyActionAddToDecreeAdvancesToDaylightAndRemovesCard(t *testing.T) {
	foxCard := firstEyrieTestCard(t, game.Fox)
	state := game.GameState{
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{foxCard},
		},
	}

	action := game.Action{
		Type: game.ActionAddToDecree,
		AddToDecree: &game.AddToDecreeAction{
			Faction: game.Eyrie,
			CardIDs: []game.CardID{foxCard.ID},
			Columns: []game.DecreeColumn{game.DecreeRecruit},
		},
	}

	next := ApplyAction(state, action)
	if next.CurrentPhase != game.Daylight || next.CurrentStep != game.StepDaylightCraft {
		t.Fatalf("expected add-to-decree to advance to daylight craft, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
	if len(next.Eyrie.CardsInHand) != 0 {
		t.Fatalf("expected decree card to be removed from hand, got %+v", next.Eyrie.CardsInHand)
	}
	if len(next.Eyrie.Decree.Recruit) != 1 || next.Eyrie.Decree.Recruit[0] != foxCard.ID {
		t.Fatalf("expected decree recruit column to contain card %d, got %+v", foxCard.ID, next.Eyrie.Decree.Recruit)
	}
}

func TestApplyTurmoilReassignsLeaderAndViziers(t *testing.T) {
	foxCard := firstEyrieTestCard(t, game.Fox)
	state := game.GameState{
		FactionTurn: game.Eyrie,
		Eyrie: game.EyrieState{
			Leader:           game.LeaderCommander,
			AvailableLeaders: []game.EyrieLeader{game.LeaderBuilder, game.LeaderDespot},
			Decree: game.Decree{
				Recruit: []game.CardID{foxCard.ID},
				Move:    []game.CardID{game.LoyalVizier1},
				Battle:  []game.CardID{game.LoyalVizier2},
			},
		},
		VictoryPoints: map[game.Faction]int{
			game.Eyrie: 5,
		},
	}

	action := game.Action{
		Type: game.ActionTurmoil,
		Turmoil: &game.TurmoilAction{
			Faction:   game.Eyrie,
			NewLeader: game.LeaderBuilder,
		},
	}

	next := ApplyAction(state, action)
	if next.Eyrie.Leader != game.LeaderBuilder {
		t.Fatalf("expected new leader to be builder, got %v", next.Eyrie.Leader)
	}
	if next.VictoryPoints[game.Eyrie] != 3 {
		t.Fatalf("expected the two bird cards in decree to cost 2 VP, got %d", next.VictoryPoints[game.Eyrie])
	}
	if len(next.Eyrie.Decree.Recruit) != 1 || next.Eyrie.Decree.Recruit[0] != game.LoyalVizier1 {
		t.Fatalf("expected loyal vizier in recruit column, got %+v", next.Eyrie.Decree.Recruit)
	}
	if len(next.Eyrie.Decree.Move) != 1 || next.Eyrie.Decree.Move[0] != game.LoyalVizier2 {
		t.Fatalf("expected loyal vizier in move column, got %+v", next.Eyrie.Decree.Move)
	}
}

func TestResolveBattleCommanderAddsOneHit(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Eyrie:    2,
						game.Marquise: 2,
					},
				},
			},
		},
		Eyrie: game.EyrieState{
			Leader: game.LeaderCommander,
		},
	}

	action := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Eyrie,
			ClearingID:    1,
			TargetFaction: game.Marquise,
		},
	}

	resolved := ResolveBattle(state, action, 1, 1)
	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload, got %+v", resolved)
	}
	if resolved.BattleResolution.DefenderLosses != 2 {
		t.Fatalf("expected commander bonus to raise defender losses to 2, got %d", resolved.BattleResolution.DefenderLosses)
	}
}

func TestEyrieEveningScoresRoosts(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		Eyrie: game.EyrieState{
			RoostsPlaced: 3,
		},
	}

	actions := ValidActions(state)
	if len(actions) != 1 || actions[0].ScoreRoosts == nil || actions[0].ScoreRoosts.Points != 2 {
		t.Fatalf("expected evening roost scoring action worth 2 points, got %+v", actions)
	}

	next := ApplyAction(state, actions[0])
	if next.VictoryPoints[game.Eyrie] != 2 {
		t.Fatalf("expected eyrie VP to increase to 2, got %d", next.VictoryPoints[game.Eyrie])
	}
}
