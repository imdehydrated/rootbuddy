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
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Buildings: []game.Building{
						{Faction: game.Eyrie, Type: game.Roost},
					},
				},
			},
		},
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

func TestApplyEyrieEmergencyOrdersDrawsAndStaysInBirdsong(t *testing.T) {
	foxCard := firstEyrieTestCard(t, game.Fox)
	state := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Eyrie,
		FactionTurn:   game.Eyrie,
		CurrentPhase:  game.Birdsong,
		CurrentStep:   game.StepBirdsong,
		Deck:          []game.CardID{foxCard.ID},
	}

	action := game.Action{
		Type: game.ActionEyrieEmergencyOrders,
		EyrieEmergency: &game.EyrieEmergencyOrdersAction{
			Faction: game.Eyrie,
			Count:   1,
		},
	}

	next := ApplyAction(state, action)
	if len(next.Eyrie.CardsInHand) != 1 || next.Eyrie.CardsInHand[0].ID != foxCard.ID {
		t.Fatalf("expected emergency orders to draw card %d, got %+v", foxCard.ID, next.Eyrie.CardsInHand)
	}
	if !next.TurnProgress.EyrieEmergencyResolved {
		t.Fatalf("expected emergency orders to mark the step resolved")
	}
	if next.CurrentPhase != game.Birdsong || next.CurrentStep != game.StepBirdsong {
		t.Fatalf("expected emergency orders to stay in birdsong, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
}

func TestApplyAddToDecreeWithoutRoostStaysInBirdsongForNewRoost(t *testing.T) {
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
	if next.CurrentPhase != game.Birdsong || next.CurrentStep != game.StepBirdsong {
		t.Fatalf("expected add-to-decree without roost to stay in birdsong, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
	if next.TurnProgress.CardsAddedToDecree != 1 {
		t.Fatalf("expected add-to-decree to mark one added card, got %+v", next.TurnProgress)
	}
}

func TestApplyEyrieNewRoostPlacesRoostWarriorsAndAdvances(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:         2,
					BuildSlots: 1,
				},
			},
		},
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Eyrie: game.EyrieState{
			WarriorSupply: 4,
		},
		TurnProgress: game.TurnProgress{
			CardsAddedToDecree: 1,
		},
	}

	action := game.Action{
		Type: game.ActionEyrieNewRoost,
		EyrieNewRoost: &game.EyrieNewRoostAction{
			Faction:    game.Eyrie,
			ClearingID: 2,
		},
	}

	next := ApplyAction(state, action)
	clearing := next.Map.Clearings[0]
	if len(clearing.Buildings) != 1 || clearing.Buildings[0].Faction != game.Eyrie || clearing.Buildings[0].Type != game.Roost {
		t.Fatalf("expected eyrie roost in clearing 2, got %+v", clearing.Buildings)
	}
	if clearing.Warriors[game.Eyrie] != 3 {
		t.Fatalf("expected three eyrie warriors in new roost clearing, got %+v", clearing.Warriors)
	}
	if next.Eyrie.WarriorSupply != 1 || next.Eyrie.RoostsPlaced != 1 {
		t.Fatalf("expected roost and warrior supply counters to update, got %+v", next.Eyrie)
	}
	if next.CurrentPhase != game.Daylight || next.CurrentStep != game.StepDaylightCraft {
		t.Fatalf("expected new roost to advance to daylight craft, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
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

func TestApplyTurmoilDoesNotDropVictoryPointsBelowZero(t *testing.T) {
	state := game.GameState{
		FactionTurn: game.Eyrie,
		Eyrie: game.EyrieState{
			Leader:           game.LeaderCommander,
			AvailableLeaders: []game.EyrieLeader{game.LeaderBuilder},
			Decree: game.Decree{
				Move:   []game.CardID{game.LoyalVizier1},
				Battle: []game.CardID{game.LoyalVizier2},
			},
		},
		VictoryPoints: map[game.Faction]int{
			game.Eyrie: 1,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionTurmoil,
		Turmoil: &game.TurmoilAction{
			Faction:   game.Eyrie,
			NewLeader: game.LeaderBuilder,
		},
	})

	if next.VictoryPoints[game.Eyrie] != 0 {
		t.Fatalf("expected turmoil VP loss to stop at zero, got %d", next.VictoryPoints[game.Eyrie])
	}
}

func TestApplyTurmoilRecyclesLeadersWhenNoneAreAvailable(t *testing.T) {
	state := game.GameState{
		FactionTurn: game.Eyrie,
		Eyrie: game.EyrieState{
			Leader: game.LeaderDespot,
			Decree: game.Decree{
				Move:  []game.CardID{game.LoyalVizier1},
				Build: []game.CardID{game.LoyalVizier2},
			},
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
		t.Fatalf("expected recycled leader choice to become builder, got %v", next.Eyrie.Leader)
	}
	if !eyrieLeaderAvailable(next.Eyrie.AvailableLeaders, game.LeaderCharismatic) ||
		!eyrieLeaderAvailable(next.Eyrie.AvailableLeaders, game.LeaderCommander) {
		t.Fatalf("expected unchosen recycled leaders to remain available, got %+v", next.Eyrie.AvailableLeaders)
	}
	if eyrieLeaderAvailable(next.Eyrie.AvailableLeaders, game.LeaderDespot) ||
		eyrieLeaderAvailable(next.Eyrie.AvailableLeaders, game.LeaderBuilder) {
		t.Fatalf("expected old and new leaders to be unavailable, got %+v", next.Eyrie.AvailableLeaders)
	}
}

func TestApplyTurmoilAfterRecycleContinuesUsingRemainingAvailableLeaders(t *testing.T) {
	state := game.GameState{
		FactionTurn: game.Eyrie,
		Eyrie: game.EyrieState{
			Leader:           game.LeaderBuilder,
			AvailableLeaders: []game.EyrieLeader{game.LeaderCharismatic, game.LeaderCommander},
			Decree: game.Decree{
				Recruit: []game.CardID{game.LoyalVizier1},
				Move:    []game.CardID{game.LoyalVizier2},
			},
		},
	}

	action := game.Action{
		Type: game.ActionTurmoil,
		Turmoil: &game.TurmoilAction{
			Faction:   game.Eyrie,
			NewLeader: game.LeaderCharismatic,
		},
	}

	next := ApplyAction(state, action)
	if next.Eyrie.Leader != game.LeaderCharismatic {
		t.Fatalf("expected next leader to be charismatic, got %v", next.Eyrie.Leader)
	}
	if len(next.Eyrie.AvailableLeaders) != 1 || next.Eyrie.AvailableLeaders[0] != game.LeaderCommander {
		t.Fatalf("expected only commander to remain available, got %+v", next.Eyrie.AvailableLeaders)
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

func TestEyrieEveningScoresRoostsThenDraws(t *testing.T) {
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
	if next.FactionTurn != game.Eyrie || next.CurrentPhase != game.Evening || next.CurrentStep != game.StepEvening {
		t.Fatalf("expected eyrie to remain in evening for draw, got faction=%v phase=%v step=%v", next.FactionTurn, next.CurrentPhase, next.CurrentStep)
	}

	actions = ValidActions(next)
	if len(actions) != 1 || actions[0].EveningDraw == nil || actions[0].EveningDraw.Count != 2 {
		t.Fatalf("expected eyrie evening draw action for 2 cards, got %+v", actions)
	}
}
