package engine

import (
	"reflect"
	"testing"

	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

func scenarioCard(t *testing.T, suit game.Suit, kind game.CardKind) game.Card {
	t.Helper()

	for _, card := range carddata.BaseDeck() {
		if card.Suit == suit && card.Kind == kind {
			return card
		}
	}

	t.Fatalf("no %v card found for suit %v", kind, suit)
	return game.Card{}
}

func scenarioQuest(t *testing.T, id game.QuestID) game.Quest {
	t.Helper()

	for _, quest := range carddata.QuestDeck() {
		if quest.ID == id {
			return quest
		}
	}

	t.Fatalf("no quest found for id %d", id)
	return game.Quest{}
}

func requireScenarioAction(t *testing.T, state game.GameState, want game.Action) {
	t.Helper()

	if !containsAction(ValidActions(state), want) {
		t.Fatalf("expected legal scenario action %+v, got %+v", want, ValidActions(state))
	}
}

func applyScenarioAction(t *testing.T, state game.GameState, action game.Action) game.GameState {
	t.Helper()

	requireScenarioAction(t, state, action)
	next := ApplyAction(state, action)
	if err := ValidateState(next); err != nil {
		t.Fatalf("scenario action %+v produced invalid state: %v\nstate=%+v", action, err, next)
	}
	return next
}

func firstScenarioActionOfType(actions []game.Action, actionType game.ActionType) (game.Action, bool) {
	for _, action := range actions {
		if action.Type == actionType {
			return action, true
		}
	}
	return game.Action{}, false
}

func chooseScenarioRolloutAction(t *testing.T, state game.GameState) game.Action {
	t.Helper()

	actions := ValidActions(state)
	if len(actions) == 0 {
		t.Fatalf("expected scenario rollout action for state %+v", state)
	}

	if len(state.PendingFieldHospitals) > 0 || len(state.PendingOutrage) > 0 {
		return actions[0]
	}

	if state.FactionTurn == game.Vagabond && state.CurrentPhase == game.Birdsong {
		for _, actionType := range []game.ActionType{game.ActionDaybreak, game.ActionSlip, game.ActionPassPhase} {
			if action, ok := firstScenarioActionOfType(actions, actionType); ok {
				return action
			}
		}
	}

	if state.CurrentPhase == game.Daylight {
		if action, ok := firstScenarioActionOfType(actions, game.ActionPassPhase); ok {
			return action
		}
	}

	if state.CurrentPhase == game.Evening {
		for _, actionType := range []game.ActionType{
			game.ActionScoreRoosts,
			game.ActionVagabondRest,
			game.ActionEveningDraw,
			game.ActionEveningDiscard,
			game.ActionVagabondDiscard,
			game.ActionVagabondItemCapacity,
			game.ActionPassPhase,
		} {
			if action, ok := firstScenarioActionOfType(actions, actionType); ok {
				return action
			}
		}
	}

	if action, ok := firstScenarioActionOfType(actions, game.ActionPassPhase); ok {
		return action
	}

	return actions[0]
}

func applyNextScenarioRolloutAction(t *testing.T, state game.GameState) game.GameState {
	t.Helper()

	return applyScenarioAction(t, state, chooseScenarioRolloutAction(t, state))
}

func TestScenarioSetupAndFullRoundUsesOnlyGeneratedLegalActions(t *testing.T) {
	state, err := SetupGame(SetupRequest{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
		Factions:      []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		MapID:         game.AutumnMapID,
		RandomSeed:    707,
	})
	if err != nil {
		t.Fatalf("expected setup to succeed, got %v", err)
	}
	if err := ValidateState(state); err != nil {
		t.Fatalf("setup produced invalid state: %v", err)
	}

	state = applyScenarioAction(t, state, game.Action{
		Type: game.ActionMarquiseSetup,
		MarquiseSetup: &game.MarquiseSetupAction{
			Faction:             game.Marquise,
			KeepClearingID:      1,
			SawmillClearingID:   1,
			WorkshopClearingID:  5,
			RecruiterClearingID: 10,
		},
	})
	state = applyScenarioAction(t, state, game.Action{
		Type: game.ActionEyrieSetup,
		EyrieSetup: &game.EyrieSetupAction{
			Faction:    game.Eyrie,
			Leader:     game.LeaderBuilder,
			ClearingID: 3,
		},
	})
	state = applyScenarioAction(t, state, game.Action{
		Type: game.ActionVagabondSetup,
		VagabondSetup: &game.VagabondSetupAction{
			Faction:   game.Vagabond,
			Character: game.CharRanger,
			ForestID:  7,
		},
	})

	if state.GamePhase != game.LifecyclePlaying || state.SetupStage != game.SetupStageComplete {
		t.Fatalf("expected setup to enter playing state, got phase=%v setup=%v", state.GamePhase, state.SetupStage)
	}

	startRound := state.RoundNumber
	startFaction := state.FactionTurn
	visited := map[game.Faction]bool{}
	log := []game.ActionType{}
	completed := false
	for step := 0; step < 200; step++ {
		if state.RoundNumber == startRound+1 && state.FactionTurn == startFaction && state.CurrentPhase == game.Birdsong {
			completed = true
			break
		}
		if state.GamePhase == game.LifecycleGameOver {
			t.Fatalf("did not expect full-round smoke scenario to end the game; log=%+v state=%+v", log, state)
		}

		visited[state.FactionTurn] = true
		action := chooseScenarioRolloutAction(t, state)
		log = append(log, action.Type)
		state = applyScenarioAction(t, state, action)
	}
	if !completed {
		t.Fatalf("scenario did not complete a full round within step limit; log=%+v state=%+v", log, state)
	}

	if state.RoundNumber != startRound+1 {
		t.Fatalf("expected full-round scenario to advance from round %d to %d, got %d", startRound, startRound+1, state.RoundNumber)
	}
	for _, faction := range state.TurnOrder {
		if !visited[faction] {
			t.Fatalf("expected full-round scenario to visit %v in order %+v; visited=%+v log=%+v", faction, state.TurnOrder, visited, log)
		}
	}
}

func TestScenarioBattleActivationAndDominanceWinPath(t *testing.T) {
	dominance := scenarioCard(t, game.Fox, game.DominanceCard)
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		RoundNumber:  3,
		TurnOrder:    []game.Faction{game.Marquise},
		VictoryPoints: map[game.Faction]int{
			game.Marquise: 10,
		},
		Map: game.Map{
			ID: game.AutumnMapID,
			Clearings: []game.Clearing{
				{ID: 1, Suit: game.Fox, Warriors: map[game.Faction]int{game.Marquise: 2}},
				{ID: 5, Suit: game.Fox, Warriors: map[game.Faction]int{game.Marquise: 1, game.Eyrie: 1}},
				{ID: 9, Suit: game.Fox, Warriors: map[game.Faction]int{game.Marquise: 1}},
			},
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{dominance},
		},
	}

	battle := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    5,
			TargetFaction: game.Eyrie,
		},
	}
	requireScenarioAction(t, state, battle)
	resolvedBattle := ResolveBattle(state, battle, 1, 0)
	state = ApplyAction(state, resolvedBattle)
	if err := ValidateState(state); err != nil {
		t.Fatalf("battle resolution produced invalid scenario state: %v", err)
	}
	if state.Map.Clearings[1].Warriors[game.Eyrie] != 0 {
		t.Fatalf("expected battle scenario to remove Eyrie warrior, got %+v", state.Map.Clearings[1].Warriors)
	}

	state = applyScenarioAction(t, state, game.Action{
		Type: game.ActionActivateDominance,
		ActivateDominance: &game.ActivateDominanceAction{
			Faction: game.Marquise,
			CardID:  dominance.ID,
		},
	})
	if state.ActiveDominance[game.Marquise] != dominance.ID {
		t.Fatalf("expected active dominance after activation, got %+v", state.ActiveDominance)
	}

	for state.GamePhase != game.LifecycleGameOver {
		state = applyNextScenarioRolloutAction(t, state)
	}
	if state.Winner != game.Marquise {
		t.Fatalf("expected Marquise dominance win, got winner=%v state=%+v", state.Winner, state)
	}
	if got := ValidActions(state); len(got) != 0 {
		t.Fatalf("expected no legal actions after dominance scenario terminal state, got %+v", got)
	}
}

func TestScenarioVagabondItemsQuestsAidAndHostileRelationship(t *testing.T) {
	foxCard := scenarioCard(t, game.Fox, game.ItemCard)
	state := game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		RoundNumber:  2,
		TurnOrder:    []game.Faction{game.Vagabond, game.Marquise},
		VictoryPoints: map[game.Faction]int{
			game.Vagabond: 0,
		},
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:        1,
					Suit:      game.Fox,
					Ruins:     true,
					RuinItems: []game.ItemType{game.ItemBag},
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			Character:  game.CharRanger,
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady, Zone: game.ItemZoneSatchel},
				{Type: game.ItemBoots, Status: game.ItemReady, Zone: game.ItemZoneSatchel},
				{Type: game.ItemHammer, Status: game.ItemReady, Zone: game.ItemZoneSatchel},
				{Type: game.ItemSword, Status: game.ItemDamaged, Zone: game.ItemZoneDamaged},
				{Type: game.ItemCoin, Status: game.ItemReady, Zone: game.ItemZoneTrack},
				{Type: game.ItemTea, Status: game.ItemReady, Zone: game.ItemZoneTrack},
				{Type: game.ItemCrossbow, Status: game.ItemReady, Zone: game.ItemZoneSatchel},
			},
			CardsInHand:     []game.Card{foxCard},
			Relationships:   map[game.Faction]game.RelationshipLevel{game.Marquise: game.RelIndifferent},
			QuestsAvailable: []game.Quest{scenarioQuest(t, 1)},
		},
	}
	if err := ValidateState(state); err != nil {
		t.Fatalf("scenario initial state invalid: %v", err)
	}

	state = applyScenarioAction(t, state, game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: game.Marquise,
			ClearingID:    1,
			CardID:        foxCard.ID,
			ItemIndex:     1,
		},
	})
	if state.Vagabond.Relationships[game.Marquise] != game.RelAmiable {
		t.Fatalf("expected Aid to improve Marquise relationship to Amiable, got %+v", state.Vagabond.Relationships)
	}
	if len(state.Marquise.CardsInHand) != 1 || state.Marquise.CardsInHand[0].ID != foxCard.ID {
		t.Fatalf("expected Aid to transfer card to Marquise, got %+v", state.Marquise.CardsInHand)
	}

	state = applyScenarioAction(t, state, game.Action{
		Type: game.ActionExplore,
		Explore: &game.ExploreAction{
			Faction:    game.Vagabond,
			ClearingID: 1,
			ItemType:   game.ItemBag,
		},
	})
	if len(state.Map.Clearings[0].RuinItems) != 0 || len(state.Vagabond.Items) != 8 {
		t.Fatalf("expected Explore to take ruin item, clearing=%+v items=%+v", state.Map.Clearings[0], state.Vagabond.Items)
	}

	state = applyScenarioAction(t, state, game.Action{
		Type: game.ActionRepair,
		Repair: &game.RepairAction{
			Faction:   game.Vagabond,
			ItemIndex: 3,
		},
	})
	if state.Vagabond.Items[3].Status != game.ItemReady {
		t.Fatalf("expected Repair to restore damaged sword, got %+v", state.Vagabond.Items[3])
	}

	beforeQuestVP := state.VictoryPoints[game.Vagabond]
	state = applyScenarioAction(t, state, game.Action{
		Type: game.ActionQuest,
		Quest: &game.QuestAction{
			Faction:     game.Vagabond,
			QuestID:     1,
			ItemIndexes: []int{4, 5},
			Reward:      game.QuestRewardVictoryPoints,
		},
	})
	if len(state.Vagabond.QuestsCompleted) != 1 || state.VictoryPoints[game.Vagabond] <= beforeQuestVP {
		t.Fatalf("expected Quest to complete and score, quests=%+v vp=%+v", state.Vagabond.QuestsCompleted, state.VictoryPoints)
	}

	state = applyScenarioAction(t, state, game.Action{
		Type: game.ActionStrike,
		Strike: &game.StrikeAction{
			Faction:       game.Vagabond,
			ClearingID:    1,
			TargetFaction: game.Marquise,
		},
	})
	if state.Map.Clearings[0].Warriors[game.Marquise] != 0 {
		t.Fatalf("expected Strike to remove Marquise warrior, got %+v", state.Map.Clearings[0].Warriors)
	}
	if state.Vagabond.Relationships[game.Marquise] != game.RelHostile {
		t.Fatalf("expected Strike warrior removal to make Marquise hostile, got %+v", state.Vagabond.Relationships)
	}

	cloneActions := ValidActions(CloneState(state))
	if !reflect.DeepEqual(ValidActions(state), cloneActions) {
		t.Fatalf("expected scenario state legal actions to remain stable across clone\nstate=%+v\nclone=%+v", ValidActions(state), cloneActions)
	}
}
