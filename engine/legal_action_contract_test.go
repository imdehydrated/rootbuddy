package engine

import (
	"reflect"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestGeneratedLegalActionsApplyToValidStates(t *testing.T) {
	setup, err := SetupGame(SetupRequest{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
		Factions:      []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		MapID:         game.AutumnMapID,
		RandomSeed:    1010,
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	afterMarquiseSetup := applyLegalContractAction(t, setup, game.Action{
		Type: game.ActionMarquiseSetup,
		MarquiseSetup: &game.MarquiseSetupAction{
			Faction:             game.Marquise,
			KeepClearingID:      1,
			SawmillClearingID:   1,
			WorkshopClearingID:  5,
			RecruiterClearingID: 10,
		},
	})
	afterEyrieSetup := applyLegalContractAction(t, afterMarquiseSetup, game.Action{
		Type: game.ActionEyrieSetup,
		EyrieSetup: &game.EyrieSetupAction{
			Faction:    game.Eyrie,
			Leader:     game.LeaderBuilder,
			ClearingID: 3,
		},
	})

	fixtures := []struct {
		name  string
		state game.GameState
	}{
		{name: "setup marquise", state: setup},
		{name: "setup eyrie", state: afterMarquiseSetup},
		{name: "setup vagabond", state: afterEyrieSetup},
		{name: "marquise birdsong", state: legalContractMarquiseBirdsongState()},
		{name: "marquise daylight craft and dominance", state: legalContractMarquiseCraftDominanceState(t)},
		{name: "marquise daylight actions", state: legalContractMarquiseDaylightState(t)},
		{name: "marquise extra action", state: legalContractMarquiseExtraActionState(t)},
		{name: "eyrie birdsong", state: legalContractEyrieBirdsongState(t)},
		{name: "eyrie emergency orders", state: legalContractEyrieEmergencyState()},
		{name: "eyrie new roost", state: legalContractEyrieNewRoostState()},
		{name: "eyrie daylight", state: legalContractEyrieDaylightState(t)},
		{name: "eyrie evening", state: legalContractEyrieEveningState()},
		{name: "alliance birdsong", state: legalContractAllianceBirdsongState(t)},
		{name: "alliance daylight", state: legalContractAllianceDaylightState(t)},
		{name: "alliance evening", state: legalContractAllianceEveningState()},
		{name: "vagabond birdsong refresh", state: legalContractVagabondBirdsongRefreshState()},
		{name: "vagabond birdsong slip", state: legalContractVagabondBirdsongSlipState()},
		{name: "vagabond daylight ranger", state: legalContractVagabondDaylightState(t, game.CharRanger)},
		{name: "vagabond daylight thief", state: legalContractVagabondDaylightState(t, game.CharThief)},
		{name: "vagabond daylight tinker", state: legalContractVagabondDaylightState(t, game.CharTinker)},
		{name: "vagabond evening", state: legalContractVagabondEveningState(t)},
		{name: "vagabond evening discard", state: legalContractVagabondEveningDiscardState(t)},
		{name: "vagabond evening capacity", state: legalContractVagabondEveningCapacityState(t)},
		{name: "pending field hospitals", state: legalContractFieldHospitalsState(t)},
		{name: "pending outrage", state: legalContractOutrageState(t)},
	}

	covered := map[game.ActionType]bool{}
	for _, fixture := range fixtures {
		assertGeneratedActionsApplyToValidState(t, fixture.name, fixture.state, covered)
	}

	requiredActionTypes := []game.ActionType{
		game.ActionMarquiseSetup,
		game.ActionEyrieSetup,
		game.ActionVagabondSetup,
		game.ActionBirdsongWood,
		game.ActionCraft,
		game.ActionActivateDominance,
		game.ActionTakeDominance,
		game.ActionRecruit,
		game.ActionMovement,
		game.ActionBattle,
		game.ActionBuild,
		game.ActionOverwork,
		game.ActionMarquiseExtraAction,
		game.ActionAddToDecree,
		game.ActionEyrieEmergencyOrders,
		game.ActionEyrieNewRoost,
		game.ActionScoreRoosts,
		game.ActionSpreadSympathy,
		game.ActionRevolt,
		game.ActionMobilize,
		game.ActionTrain,
		game.ActionOrganize,
		game.ActionDaybreak,
		game.ActionSlip,
		game.ActionExplore,
		game.ActionAid,
		game.ActionQuest,
		game.ActionStrike,
		game.ActionRepair,
		game.ActionVagabondSteal,
		game.ActionVagabondDayLabor,
		game.ActionVagabondHideout,
		game.ActionVagabondRest,
		game.ActionEveningDraw,
		game.ActionVagabondDiscard,
		game.ActionVagabondItemCapacity,
		game.ActionFieldHospitals,
		game.ActionResolveOutrage,
		game.ActionPassPhase,
	}
	for _, actionType := range requiredActionTypes {
		if !covered[actionType] {
			t.Fatalf("legal-action contract fixtures did not cover action type %v; covered=%+v", actionType, covered)
		}
	}
}

func assertGeneratedActionsApplyToValidState(t *testing.T, name string, state game.GameState, covered map[game.ActionType]bool) {
	t.Helper()

	if err := ValidateState(state); err != nil {
		t.Fatalf("%s fixture is invalid before contract check: %v\nstate=%+v", name, err, state)
	}

	actions := ValidActions(CloneState(state))
	if len(actions) == 0 {
		t.Fatalf("%s fixture generated no legal actions", name)
	}

	for index, action := range actions {
		t.Run(name, func(t *testing.T) {
			before := CloneState(state)
			original := CloneState(before)
			next, err := ApplyLegalAction(before, action)
			if err != nil {
				t.Fatalf("generated action %d %+v failed legal application: %v", index, action, err)
			}
			if !reflect.DeepEqual(before, original) {
				t.Fatalf("generated action %d %+v mutated input state\nbefore=%+v\nafter=%+v", index, action, original, before)
			}
			if err := ValidateState(next); err != nil {
				t.Fatalf("generated action %d %+v produced invalid state: %v\nnext=%+v", index, action, err, next)
			}
			covered[action.Type] = true
		})
	}
}

func applyLegalContractAction(t *testing.T, state game.GameState, action game.Action) game.GameState {
	t.Helper()

	next, err := ApplyLegalAction(state, action)
	if err != nil {
		t.Fatalf("failed to apply contract setup action %+v: %v", action, err)
	}
	if err := ValidateState(next); err != nil {
		t.Fatalf("contract setup action %+v produced invalid state: %v", action, err)
	}
	return next
}

func legalContractCard(t *testing.T, id game.CardID) game.Card {
	t.Helper()

	card, ok := CardByID(id)
	if !ok {
		t.Fatalf("missing contract card id %d", id)
	}
	return card
}

func legalContractMarquiseBirdsongState() game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie},
		Map: game.Map{Clearings: []game.Clearing{
			{
				ID:         1,
				Suit:       game.Fox,
				BuildSlots: 2,
				Buildings:  []game.Building{{Faction: game.Marquise, Type: game.Sawmill}},
			},
		}},
		Marquise: game.MarquiseState{
			WarriorSupply:  25,
			WoodSupply:     8,
			SawmillsPlaced: 1,
		},
	}
}

func legalContractMarquiseCraftDominanceState(t *testing.T) game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightCraft,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie},
		VictoryPoints: map[game.Faction]int{
			game.Marquise: 10,
		},
		AvailableDominance: []game.CardID{27},
		ItemSupply:         InitialItemSupply(),
		Map: game.Map{Clearings: []game.Clearing{
			{
				ID:         1,
				Suit:       game.Fox,
				BuildSlots: 2,
				Buildings:  []game.Building{{Faction: game.Marquise, Type: game.Workshop}},
			},
		}},
		Marquise: game.MarquiseState{
			WarriorSupply:   25,
			WoodSupply:      8,
			WorkshopsPlaced: 1,
			CardsInHand: []game.Card{
				legalContractCard(t, 52),
				legalContractCard(t, 54),
				legalContractCard(t, 24),
			},
		},
	}
}

func legalContractMarquiseDaylightState(t *testing.T) game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie},
		ItemSupply:   InitialItemSupply(),
		Map: game.Map{Clearings: []game.Clearing{
			{
				ID:         1,
				Suit:       game.Fox,
				Adj:        []int{2},
				BuildSlots: 3,
				Warriors:   map[game.Faction]int{game.Marquise: 3, game.Eyrie: 1},
				Wood:       2,
				Buildings: []game.Building{
					{Faction: game.Marquise, Type: game.Sawmill},
					{Faction: game.Marquise, Type: game.Recruiter},
				},
			},
			{
				ID:         2,
				Suit:       game.Rabbit,
				Adj:        []int{1},
				BuildSlots: 2,
				Warriors:   map[game.Faction]int{game.Marquise: 1},
			},
		}},
		Marquise: game.MarquiseState{
			WarriorSupply:    21,
			WoodSupply:       6,
			SawmillsPlaced:   1,
			RecruitersPlaced: 1,
			CardsInHand:      []game.Card{legalContractCard(t, 53)},
		},
		Eyrie: game.EyrieState{
			WarriorSupply: 19,
		},
	}
}

func legalContractMarquiseExtraActionState(t *testing.T) game.GameState {
	state := legalContractMarquiseDaylightState(t)
	state.TurnProgress.ActionsUsed = 3
	state.Marquise.CardsInHand = []game.Card{legalContractCard(t, 11)}
	return state
}

func legalContractEyrieBirdsongState(t *testing.T) game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie},
		Map: game.Map{Clearings: []game.Clearing{
			{ID: 1, Suit: game.Fox, BuildSlots: 2},
		}},
		Eyrie: game.EyrieState{
			WarriorSupply:    20,
			Leader:           game.LeaderBuilder,
			AvailableLeaders: []game.EyrieLeader{game.LeaderCharismatic, game.LeaderCommander, game.LeaderDespot},
			CardsInHand:      []game.Card{legalContractCard(t, 24), legalContractCard(t, 52)},
			Decree: game.Decree{
				Recruit: []game.CardID{game.LoyalVizier1},
				Move:    []game.CardID{game.LoyalVizier2},
			},
		},
	}
}

func legalContractEyrieEmergencyState() game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie},
		Deck:         []game.CardID{52},
		Map: game.Map{Clearings: []game.Clearing{
			{
				ID:         1,
				Suit:       game.Fox,
				BuildSlots: 1,
				Buildings:  []game.Building{{Faction: game.Eyrie, Type: game.Roost}},
			},
		}},
		Eyrie: game.EyrieState{
			WarriorSupply: 20,
			RoostsPlaced:  1,
			Leader:        game.LeaderBuilder,
		},
	}
}

func legalContractEyrieNewRoostState() game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie},
		Map: game.Map{Clearings: []game.Clearing{
			{ID: 1, Suit: game.Fox, BuildSlots: 1},
			{
				ID:         2,
				Suit:       game.Rabbit,
				BuildSlots: 1,
				Warriors:   map[game.Faction]int{game.Marquise: 1},
			},
		}},
		Marquise: game.MarquiseState{
			WarriorSupply: 24,
			WoodSupply:    8,
		},
		Eyrie: game.EyrieState{
			WarriorSupply: 3,
			Leader:        game.LeaderBuilder,
		},
		TurnProgress: game.TurnProgress{
			CardsAddedToDecree: 1,
		},
	}
}

func legalContractEyrieDaylightState(t *testing.T) game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie},
		ItemSupply:   InitialItemSupply(),
		Map: game.Map{Clearings: []game.Clearing{
			{
				ID:         1,
				Suit:       game.Fox,
				Adj:        []int{2},
				BuildSlots: 2,
				Warriors:   map[game.Faction]int{game.Eyrie: 4, game.Marquise: 1},
				Buildings:  []game.Building{{Faction: game.Eyrie, Type: game.Roost}},
			},
			{
				ID:         2,
				Suit:       game.Rabbit,
				Adj:        []int{1},
				BuildSlots: 2,
				Warriors:   map[game.Faction]int{game.Eyrie: 1},
			},
		}},
		Marquise: game.MarquiseState{
			WarriorSupply: 24,
			WoodSupply:    8,
		},
		Eyrie: game.EyrieState{
			WarriorSupply: 15,
			RoostsPlaced:  1,
			Leader:        game.LeaderBuilder,
			Decree: game.Decree{
				Recruit: []game.CardID{game.LoyalVizier1},
				Move:    []game.CardID{game.LoyalVizier2},
				Battle:  []game.CardID{52},
				Build:   []game.CardID{24},
			},
			CardsInHand: []game.Card{legalContractCard(t, 11)},
		},
	}
}

func legalContractEyrieEveningState() game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Eyrie,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie},
		Map: game.Map{Clearings: []game.Clearing{
			{
				ID:        1,
				Suit:      game.Fox,
				Buildings: []game.Building{{Faction: game.Eyrie, Type: game.Roost}},
			},
			{
				ID:        2,
				Suit:      game.Rabbit,
				Buildings: []game.Building{{Faction: game.Eyrie, Type: game.Roost}},
			},
			{
				ID:        3,
				Suit:      game.Mouse,
				Buildings: []game.Building{{Faction: game.Eyrie, Type: game.Roost}},
			},
		}},
		Eyrie: game.EyrieState{
			WarriorSupply: 20,
			RoostsPlaced:  3,
			Leader:        game.LeaderBuilder,
		},
	}
}

func legalContractAllianceBirdsongState(t *testing.T) game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Alliance, game.Marquise},
		Map: game.Map{Clearings: []game.Clearing{
			{
				ID:         1,
				Suit:       game.Fox,
				BuildSlots: 1,
				Warriors:   map[game.Faction]int{game.Marquise: 1},
				Tokens:     []game.Token{{Faction: game.Alliance, Type: game.TokenSympathy}},
			},
			{ID: 2, Suit: game.Rabbit, Adj: []int{1}, BuildSlots: 1},
		}},
		Marquise: game.MarquiseState{
			WarriorSupply: 24,
			WoodSupply:    8,
		},
		Alliance: game.AllianceState{
			WarriorSupply:  10,
			SympathyPlaced: 1,
			Supporters:     []game.Card{legalContractCard(t, 52), legalContractCard(t, 53), legalContractCard(t, 24)},
		},
	}
}

func legalContractAllianceDaylightState(t *testing.T) game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Alliance, game.Marquise},
		ItemSupply:   InitialItemSupply(),
		Map: game.Map{Clearings: []game.Clearing{
			{
				ID:        1,
				Suit:      game.Fox,
				Buildings: []game.Building{{Faction: game.Alliance, Type: game.Base}},
			},
		}},
		Alliance: game.AllianceState{
			WarriorSupply:  10,
			FoxBasePlaced:  true,
			CardsInHand:    []game.Card{legalContractCard(t, 52), legalContractCard(t, 11)},
			Supporters:     []game.Card{legalContractCard(t, 24)},
			SympathyPlaced: 0,
		},
	}
}

func legalContractAllianceEveningState() game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Alliance, game.Marquise},
		Map: game.Map{Clearings: []game.Clearing{
			{
				ID:       1,
				Suit:     game.Fox,
				Adj:      []int{2},
				Warriors: map[game.Faction]int{game.Alliance: 2},
				Buildings: []game.Building{
					{Faction: game.Alliance, Type: game.Base},
				},
			},
			{
				ID:       2,
				Suit:     game.Rabbit,
				Adj:      []int{1},
				Warriors: map[game.Faction]int{game.Marquise: 1},
			},
		}},
		Marquise: game.MarquiseState{
			WarriorSupply: 24,
			WoodSupply:    8,
		},
		Alliance: game.AllianceState{
			WarriorSupply: 6,
			Officers:      2,
			FoxBasePlaced: true,
		},
	}
}

func legalContractVagabondDaylightState(t *testing.T, character game.VagabondCharacter) game.GameState {
	items := []game.Item{
		game.NormalizeItemZone(game.Item{Type: game.ItemTorch, Status: game.ItemReady}),
		game.NormalizeItemZone(game.Item{Type: game.ItemBoots, Status: game.ItemReady}),
		game.NormalizeItemZone(game.Item{Type: game.ItemHammer, Status: game.ItemReady}),
		game.NormalizeItemZone(game.Item{Type: game.ItemSword, Status: game.ItemReady}),
		game.NormalizeItemZone(game.Item{Type: game.ItemCrossbow, Status: game.ItemReady}),
		game.NormalizeItemZone(game.Item{Type: game.ItemTea, Status: game.ItemDamaged, DamagedSide: game.ItemReady}),
		game.NormalizeItemZone(game.Item{Type: game.ItemCoin, Status: game.ItemDamaged, DamagedSide: game.ItemExhausted}),
		game.NormalizeItemZone(game.Item{Type: game.ItemBag, Status: game.ItemDamaged, DamagedSide: game.ItemReady}),
	}

	return game.GameState{
		GameMode:      game.GameModeOnline,
		TrackAllHands: true,
		GamePhase:     game.LifecyclePlaying,
		SetupStage:    game.SetupStageComplete,
		FactionTurn:   game.Vagabond,
		CurrentPhase:  game.Daylight,
		CurrentStep:   game.StepDaylightActions,
		RoundNumber:   1,
		TurnOrder:     []game.Faction{game.Vagabond, game.Marquise, game.Eyrie},
		VictoryPoints: map[game.Faction]int{
			game.Vagabond: 0,
		},
		ItemSupply: InitialItemSupply(),
		DiscardPile: []game.CardID{
			53,
		},
		CraftedItems: map[game.Faction][]game.ItemType{
			game.Marquise: []game.ItemType{game.ItemBoots},
		},
		Map: game.Map{Clearings: []game.Clearing{
			{
				ID:        1,
				Suit:      game.Fox,
				Adj:       []int{2},
				Ruins:     true,
				RuinItems: []game.ItemType{game.ItemBag},
				Warriors: map[game.Faction]int{
					game.Marquise: 1,
				},
			},
			{
				ID:   2,
				Suit: game.Rabbit,
				Adj:  []int{1},
			},
		}},
		Marquise: game.MarquiseState{
			WarriorSupply: 24,
			WoodSupply:    8,
			CardsInHand:   []game.Card{legalContractCard(t, 24)},
		},
		Vagabond: game.VagabondState{
			Character:       character,
			ClearingID:      1,
			Items:           items,
			CardsInHand:     []game.Card{legalContractCard(t, 52), legalContractCard(t, 11)},
			Relationships:   map[game.Faction]game.RelationshipLevel{game.Marquise: game.RelIndifferent, game.Eyrie: game.RelIndifferent},
			QuestsAvailable: []game.Quest{scenarioQuest(t, 4)},
		},
	}
}

func legalContractVagabondBirdsongRefreshState() game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Vagabond, game.Marquise},
		Map: game.Map{Clearings: []game.Clearing{
			{ID: 1, Suit: game.Fox},
		}},
		Vagabond: game.VagabondState{
			Character:  game.CharRanger,
			ClearingID: 1,
			Items: []game.Item{
				game.NormalizeItemZone(game.Item{Type: game.ItemTorch, Status: game.ItemExhausted}),
				game.NormalizeItemZone(game.Item{Type: game.ItemBoots, Status: game.ItemExhausted}),
				game.NormalizeItemZone(game.Item{Type: game.ItemSword, Status: game.ItemReady}),
			},
			Relationships: map[game.Faction]game.RelationshipLevel{game.Marquise: game.RelIndifferent},
		},
	}
}

func legalContractVagabondBirdsongSlipState() game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Vagabond, game.Marquise},
		Map: game.Map{
			Clearings: []game.Clearing{
				{ID: 1, Suit: game.Fox, Adj: []int{2}},
				{ID: 2, Suit: game.Rabbit, Adj: []int{1}},
			},
			Forests: []game.Forest{
				{ID: 1, AdjacentClearings: []int{1, 2}},
			},
		},
		Vagabond: game.VagabondState{
			Character:  game.CharRanger,
			ClearingID: 1,
			Items: []game.Item{
				game.NormalizeItemZone(game.Item{Type: game.ItemTorch, Status: game.ItemReady}),
				game.NormalizeItemZone(game.Item{Type: game.ItemBoots, Status: game.ItemReady}),
			},
			Relationships: map[game.Faction]game.RelationshipLevel{game.Marquise: game.RelIndifferent},
		},
		TurnProgress: game.TurnProgress{
			HasRefreshed: true,
		},
	}
}

func legalContractVagabondEveningState(t *testing.T) game.GameState {
	state := legalContractVagabondDaylightState(t, game.CharRanger)
	state.CurrentPhase = game.Evening
	state.CurrentStep = game.StepEvening
	state.TurnProgress = game.TurnProgress{}
	state.Vagabond.CardsInHand = []game.Card{legalContractCard(t, 52), legalContractCard(t, 11)}
	return state
}

func legalContractVagabondEveningDiscardState(t *testing.T) game.GameState {
	state := legalContractVagabondEveningState(t)
	state.TurnProgress.VagabondRestResolved = true
	state.TurnProgress.VagabondEveningDrawn = true
	return state
}

func legalContractVagabondEveningCapacityState(t *testing.T) game.GameState {
	state := legalContractVagabondEveningDiscardState(t)
	state.TurnProgress.VagabondDiscardResolved = true
	return state
}

func legalContractFieldHospitalsState(t *testing.T) game.GameState {
	return game.GameState{
		GamePhase:    game.LifecyclePlaying,
		SetupStage:   game.SetupStageComplete,
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		RoundNumber:  1,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie},
		Map: game.Map{Clearings: []game.Clearing{
			{
				ID:         1,
				Suit:       game.Fox,
				Tokens:     []game.Token{{Faction: game.Marquise, Type: game.TokenKeep}},
				BuildSlots: 1,
			},
			{ID: 2, Suit: game.Fox, BuildSlots: 1},
		}},
		Marquise: game.MarquiseState{
			WarriorSupply:  25,
			WoodSupply:     8,
			KeepClearingID: 1,
			CardsInHand:    []game.Card{legalContractCard(t, 52)},
		},
		PendingFieldHospitals: []game.FieldHospitalsPending{
			{ClearingID: 2, Suit: game.Fox, WarriorCount: 2},
		},
	}
}

func legalContractOutrageState(t *testing.T) game.GameState {
	return game.GameState{
		GameMode:      game.GameModeOnline,
		TrackAllHands: true,
		GamePhase:     game.LifecyclePlaying,
		SetupStage:    game.SetupStageComplete,
		FactionTurn:   game.Marquise,
		CurrentPhase:  game.Daylight,
		CurrentStep:   game.StepDaylightActions,
		RoundNumber:   1,
		TurnOrder:     []game.Faction{game.Marquise, game.Alliance},
		Map: game.Map{Clearings: []game.Clearing{
			{ID: 1, Suit: game.Fox},
		}},
		Marquise: game.MarquiseState{
			WarriorSupply: 25,
			WoodSupply:    8,
			CardsInHand:   []game.Card{legalContractCard(t, 52)},
		},
		Alliance: game.AllianceState{
			WarriorSupply: 10,
		},
		PendingOutrage: []game.OutragePending{
			{Faction: game.Marquise, Suit: game.Fox},
		},
	}
}
