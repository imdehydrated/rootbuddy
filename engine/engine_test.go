package engine

import (
	"reflect"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func containsAction(actions []game.Action, want game.Action) bool {
	for _, action := range actions {
		if reflect.DeepEqual(action, want) {
			return true
		}
	}
	return false
}

func TestValidActionsReturnsBirdsongWoodActionForBirdsong(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Sawmill},
					},
				},
			},
		},
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
	}

	got := ValidActions(state)
	want := game.Action{
		Type: game.ActionBirdsongWood,
		BirdsongWood: &game.BirdsongWoodAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
			Amount:      1,
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected birdsong wood action %+v, got %+v", want, got)
	}
}

func TestValidActionsReturnsRecruitAndMovementActionsForDaylightStep(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:       1,
					Adj:      []int{2},
					Warriors: map[game.Faction]int{game.Marquise: 1},
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Recruiter},
					},
				},
				{
					ID:  2,
					Adj: []int{1},
				},
			},
		},
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		Marquise: game.MarquiseState{
			WarriorSupply: 1,
		},
	}

	got := ValidActions(state)
	wantRecruit := game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
		},
	}
	wantMovement := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Marquise,
			Count:    1,
			MaxCount: 1,
			From:     1,
			To:       2,
		},
	}

	if !containsAction(got, wantRecruit) {
		t.Fatalf("expected recruit action %+v, got %+v", wantRecruit, got)
	}
	if !containsAction(got, wantMovement) {
		t.Fatalf("expected movement action %+v, got %+v", wantMovement, got)
	}
}

func TestValidActionsReturnsOnlyPassPhaseWhenActionLimitIsReached(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		TurnProgress: game.TurnProgress{
			ActionsUsed: 3,
		},
	}

	got := ValidActions(state)
	if len(got) != 1 || got[0].Type != game.ActionPassPhase {
		t.Fatalf("expected only pass-phase action at action limit, got %+v", got)
	}
}

func TestEngineFlowBirdsongRecruitMoveBattleResolve(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Adj:  []int{2},
					Suit: game.Fox,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Sawmill},
						{Faction: game.Marquise, Type: game.Recruiter},
					},
				},
				{
					ID:   2,
					Adj:  []int{1},
					Suit: game.Rabbit,
					Warriors: map[game.Faction]int{
						game.Eyrie: 1,
					},
				},
			},
		},
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Marquise: game.MarquiseState{
			WarriorSupply: 1,
		},
	}

	birdsongAction := game.Action{
		Type: game.ActionBirdsongWood,
		BirdsongWood: &game.BirdsongWoodAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
			Amount:      1,
		},
	}

	actions := ValidActions(state)
	if !containsAction(actions, birdsongAction) {
		t.Fatalf("expected birdsong action %+v, got %+v", birdsongAction, actions)
	}

	afterBirdsong := ApplyAction(state, birdsongAction)
	if afterBirdsong.CurrentPhase != game.Daylight || afterBirdsong.CurrentStep != game.StepDaylightCraft {
		t.Fatalf("expected birdsong to advance to daylight craft, got phase=%v step=%v", afterBirdsong.CurrentPhase, afterBirdsong.CurrentStep)
	}

	recruitAction := game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
		},
	}

	afterCraftPass := ApplyAction(afterBirdsong, game.Action{
		Type: game.ActionPassPhase,
		PassPhase: &game.PassPhaseAction{
			Faction: game.Marquise,
		},
	})
	if afterCraftPass.CurrentPhase != game.Daylight || afterCraftPass.CurrentStep != game.StepDaylightActions {
		t.Fatalf("expected craft pass to advance to daylight actions, got phase=%v step=%v", afterCraftPass.CurrentPhase, afterCraftPass.CurrentStep)
	}

	actions = ValidActions(afterCraftPass)
	if !containsAction(actions, recruitAction) {
		t.Fatalf("expected recruit action %+v, got %+v", recruitAction, actions)
	}

	afterRecruit := ApplyAction(afterCraftPass, recruitAction)
	moveAction := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Marquise,
			Count:    1,
			MaxCount: 1,
			From:     1,
			To:       2,
		},
	}

	actions = ValidActions(afterRecruit)
	if !containsAction(actions, moveAction) {
		t.Fatalf("expected movement action %+v after recruit, got %+v", moveAction, actions)
	}

	afterMove := ApplyAction(afterRecruit, moveAction)
	battleAction := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    2,
			TargetFaction: game.Eyrie,
		},
	}

	actions = ValidActions(afterMove)
	if !containsAction(actions, battleAction) {
		t.Fatalf("expected battle action %+v after move, got %+v", battleAction, actions)
	}

	resolved := ResolveBattle(afterMove, battleAction, 1, 0)
	afterBattle := ApplyAction(afterMove, resolved)
	if afterBattle.Map.Clearings[1].Warriors[game.Eyrie] != 0 {
		t.Fatalf("expected eyrie warrior to be removed after battle, got %d", afterBattle.Map.Clearings[1].Warriors[game.Eyrie])
	}
	if afterBattle.TurnProgress.ActionsUsed != 3 {
		t.Fatalf("expected recruit, move, and battle to consume 3 actions, got %d", afterBattle.TurnProgress.ActionsUsed)
	}
}

func TestValidActionsUsesDedicatedCraftStepBeforeFactionActions(t *testing.T) {
	craftCard := game.Card{
		ID:   9001,
		Name: "Test Craft",
		Suit: game.Fox,
		Kind: game.OneTimeEffectCard,
		CraftingCost: game.CraftingCost{
			Fox: 1,
		},
	}

	tests := []struct {
		name        string
		state       game.GameState
		wantFaction game.Faction
	}{
		{
			name: "marquise",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   1,
							Suit: game.Fox,
							Buildings: []game.Building{
								{Faction: game.Marquise, Type: game.Workshop},
							},
						},
					},
				},
				FactionTurn:  game.Marquise,
				CurrentPhase: game.Daylight,
				CurrentStep:  game.StepDaylightCraft,
				Marquise: game.MarquiseState{
					CardsInHand: []game.Card{craftCard},
				},
			},
			wantFaction: game.Marquise,
		},
		{
			name: "eyrie",
			state: game.GameState{
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
				FactionTurn:  game.Eyrie,
				CurrentPhase: game.Daylight,
				CurrentStep:  game.StepDaylightCraft,
				Eyrie: game.EyrieState{
					CardsInHand: []game.Card{craftCard},
				},
			},
			wantFaction: game.Eyrie,
		},
		{
			name: "alliance",
			state: game.GameState{
				Map: game.Map{
					Clearings: []game.Clearing{
						{
							ID:   1,
							Suit: game.Fox,
							Buildings: []game.Building{
								{Faction: game.Alliance, Type: game.Base},
							},
						},
					},
				},
				FactionTurn:  game.Alliance,
				CurrentPhase: game.Daylight,
				CurrentStep:  game.StepDaylightCraft,
				Alliance: game.AllianceState{
					CardsInHand: []game.Card{craftCard},
				},
			},
			wantFaction: game.Alliance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := ValidActions(tt.state)
			wantCraft := game.Action{
				Type: game.ActionCraft,
				Craft: &game.CraftAction{
					Faction:               tt.wantFaction,
					CardID:                craftCard.ID,
					UsedWorkshopClearings: []int{1},
				},
			}
			if !containsAction(actions, wantCraft) {
				t.Fatalf("expected craft action %+v, got %+v", wantCraft, actions)
			}

			next := ApplyAction(tt.state, wantCraft)
			if next.CurrentPhase != game.Daylight || next.CurrentStep != game.StepDaylightCraft {
				t.Fatalf("expected craft to stay in craft step for another craft/pass choice, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
			}

			afterPass := ApplyAction(next, game.Action{
				Type: game.ActionPassPhase,
				PassPhase: &game.PassPhaseAction{
					Faction: tt.wantFaction,
				},
			})
			if afterPass.CurrentPhase != game.Daylight || afterPass.CurrentStep != game.StepDaylightActions {
				t.Fatalf("expected craft pass to advance to action step, got phase=%v step=%v", afterPass.CurrentPhase, afterPass.CurrentStep)
			}
		})
	}
}

func TestImplicitDaylightCraftStepCanCraftAndPass(t *testing.T) {
	craftCard := game.Card{
		ID:   9002,
		Name: "Implicit Craft",
		Suit: game.Fox,
		Kind: game.OneTimeEffectCard,
		CraftingCost: game.CraftingCost{
			Fox: 1,
		},
	}
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Workshop},
					},
				},
			},
		},
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Daylight,
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{craftCard},
		},
	}

	actions := ValidActions(state)
	craftAction := game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction:               game.Marquise,
			CardID:                craftCard.ID,
			UsedWorkshopClearings: []int{1},
		},
	}
	if !containsAction(actions, craftAction) {
		t.Fatalf("expected implicit daylight craft action %+v, got %+v", craftAction, actions)
	}

	afterCraft := ApplyAction(state, craftAction)
	if afterCraft.CurrentPhase != game.Daylight || afterCraft.CurrentStep != game.StepDaylightCraft {
		t.Fatalf("expected implicit craft to remain in craft step, got phase=%v step=%v", afterCraft.CurrentPhase, afterCraft.CurrentStep)
	}

	afterPass := ApplyAction(state, game.Action{
		Type: game.ActionPassPhase,
		PassPhase: &game.PassPhaseAction{
			Faction: game.Marquise,
		},
	})
	if afterPass.CurrentPhase != game.Daylight || afterPass.CurrentStep != game.StepDaylightActions {
		t.Fatalf("expected implicit craft pass to advance to actions, got phase=%v step=%v", afterPass.CurrentPhase, afterPass.CurrentStep)
	}
}
