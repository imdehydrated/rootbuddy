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

func TestValidActionsReturnsRecruitActionsForRecruitStep(t *testing.T) {
	state := game.GameState{
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
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepRecruit,
		Marquise: game.MarquiseState{
			WarriorSupply: 1,
		},
	}

	got := ValidActions(state)
	want := game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected recruit action %+v, got %+v", want, got)
	}

	for _, action := range got {
		if action.Type != game.ActionRecruit {
			t.Fatalf("expected only recruit actions at recruit step, got %+v", got)
		}
	}
}

func TestValidActionsReturnsDaylightActionsForDaylightStep(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:       1,
					Adj:      []int{2},
					Warriors: map[game.Faction]int{game.Marquise: 1},
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
	}

	got := ValidActions(state)
	want := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Marquise,
			MaxCount: 1,
			From:     1,
			To:       2,
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected movement action %+v, got %+v", want, got)
	}

	for _, action := range got {
		if action.Type == game.ActionRecruit {
			t.Fatalf("did not expect recruit actions during daylight step, got %+v", got)
		}
	}
}

func TestValidActionsFallsBackToPhaseWhenStepIsUnspecified(t *testing.T) {
	state := game.GameState{
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
		FactionTurn:  game.Marquise,
		CurrentPhase: game.Birdsong,
		Marquise: game.MarquiseState{
			WarriorSupply: 1,
		},
	}

	got := ValidActions(state)

	if len(got) != 1 || got[0].Type != game.ActionRecruit {
		t.Fatalf("expected birdsong fallback to recruit action, got %+v", got)
	}
}

func TestEngineFlowRecruitMoveBattleResolve(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Adj:  []int{2},
					Suit: game.Fox,
					Buildings: []game.Building{
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
		CurrentStep:  game.StepRecruit,
		Marquise: game.MarquiseState{
			WarriorSupply: 1,
		},
	}

	recruitAction := game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
		},
	}

	actions := ValidActions(state)
	if !containsAction(actions, recruitAction) {
		t.Fatalf("expected recruit action %+v, got %+v", recruitAction, actions)
	}

	afterRecruit := ApplyAction(state, recruitAction)
	if afterRecruit.CurrentStep != game.StepDaylightActions {
		t.Fatalf("expected recruit to advance to daylight actions, got %v", afterRecruit.CurrentStep)
	}

	moveAction := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Marquise,
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
	if afterMove.Map.Clearings[1].Warriors[game.Marquise] != 1 {
		t.Fatalf("expected marquise warrior in clearing 2 after move, got %d", afterMove.Map.Clearings[1].Warriors[game.Marquise])
	}

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
	if resolved.Type != game.ActionBattleResolution || resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution action, got %+v", resolved)
	}

	afterBattle := ApplyAction(afterMove, resolved)
	if afterBattle.Map.Clearings[1].Warriors[game.Eyrie] != 0 {
		t.Fatalf("expected eyrie warrior to be removed after battle, got %d", afterBattle.Map.Clearings[1].Warriors[game.Eyrie])
	}

	actions = ValidActions(afterBattle)
	if containsAction(actions, battleAction) {
		t.Fatalf("did not expect battle action %+v after defender was removed, got %+v", battleAction, actions)
	}
}
