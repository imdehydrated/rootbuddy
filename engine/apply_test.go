package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func hasCard(cards []game.Card, id game.CardID) bool {
	for _, card := range cards {
		if card.ID == id {
			return true
		}
	}
	return false
}

func TestApplyActionRecruit(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
				},
			},
		},
		CurrentPhase: game.Daylight,
		CurrentStep:  game.StepDaylightActions,
		Marquise: game.MarquiseState{
			WarriorSupply: 2,
		},
	}

	action := game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
		},
	}

	next := ApplyAction(state, action)

	if next.Map.Clearings[0].Warriors[game.Marquise] != 1 {
		t.Fatalf("expected 1 marquise warrior after recruit, got %d", next.Map.Clearings[0].Warriors[game.Marquise])
	}
	if next.Marquise.WarriorSupply != 1 {
		t.Fatalf("expected warrior supply to decrease to 1, got %d", next.Marquise.WarriorSupply)
	}
	if !next.TurnProgress.RecruitUsed {
		t.Fatalf("expected recruit to be marked used after recruit action")
	}
	if next.TurnProgress.ActionsUsed != 1 {
		t.Fatalf("expected recruit to consume 1 action, got %d", next.TurnProgress.ActionsUsed)
	}
	if next.CurrentPhase != game.Daylight {
		t.Fatalf("expected recruit to remain in daylight, got %v", next.CurrentPhase)
	}
	if next.CurrentStep != game.StepDaylightActions {
		t.Fatalf("expected recruit to advance step to daylight actions, got %v", next.CurrentStep)
	}

	if state.Map.Clearings[0].Warriors != nil {
		t.Fatalf("expected original state warriors to remain nil, got %+v", state.Map.Clearings[0].Warriors)
	}
	if state.Marquise.WarriorSupply != 2 {
		t.Fatalf("expected original warrior supply to remain 2, got %d", state.Marquise.WarriorSupply)
	}
	if state.TurnProgress.RecruitUsed {
		t.Fatalf("expected original recruit-used flag to remain false")
	}
	if state.CurrentPhase != game.Daylight {
		t.Fatalf("expected original phase to remain daylight, got %v", state.CurrentPhase)
	}
	if state.CurrentStep != game.StepDaylightActions {
		t.Fatalf("expected original step to remain daylight actions, got %v", state.CurrentStep)
	}
}

func TestApplyActionOverwork(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   2,
					Wood: 0,
				},
			},
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 10, Name: "Spent Card"},
				{ID: 11, Name: "Kept Card"},
			},
		},
	}

	action := game.Action{
		Type: game.ActionOverwork,
		Overwork: &game.OverworkAction{
			Faction:    game.Marquise,
			ClearingID: 2,
			CardID:     10,
		},
	}

	next := ApplyAction(state, action)

	if next.Map.Clearings[0].Wood != 1 {
		t.Fatalf("expected overwork to add 1 wood, got %d", next.Map.Clearings[0].Wood)
	}
	if hasCard(next.Marquise.CardsInHand, 10) {
		t.Fatalf("expected spent card 10 to be removed from hand")
	}
	if !hasCard(next.Marquise.CardsInHand, 11) {
		t.Fatalf("expected unspent card 11 to remain in hand")
	}

	if state.Map.Clearings[0].Wood != 0 {
		t.Fatalf("expected original state wood to remain 0, got %d", state.Map.Clearings[0].Wood)
	}
	if !hasCard(state.Marquise.CardsInHand, 10) {
		t.Fatalf("expected original hand to still contain card 10")
	}
}

func TestApplyActionMovement(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:       1,
					Warriors: map[game.Faction]int{game.Marquise: 3},
				},
				{
					ID:       2,
					Warriors: map[game.Faction]int{},
				},
			},
		},
	}

	action := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Marquise,
			Count:    2,
			MaxCount: 2,
			From:     1,
			To:       2,
		},
	}

	next := ApplyAction(state, action)

	if next.Map.Clearings[0].Warriors[game.Marquise] != 1 {
		t.Fatalf("expected 1 marquise warrior left in origin, got %d", next.Map.Clearings[0].Warriors[game.Marquise])
	}
	if next.Map.Clearings[1].Warriors[game.Marquise] != 2 {
		t.Fatalf("expected 2 marquise warriors in destination, got %d", next.Map.Clearings[1].Warriors[game.Marquise])
	}

	if state.Map.Clearings[0].Warriors[game.Marquise] != 3 {
		t.Fatalf("expected original origin warriors to remain 3, got %d", state.Map.Clearings[0].Warriors[game.Marquise])
	}
	if state.Map.Clearings[1].Warriors[game.Marquise] != 0 {
		t.Fatalf("expected original destination warriors to remain 0, got %d", state.Map.Clearings[1].Warriors[game.Marquise])
	}
}

func TestApplyActionBuild(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:        3,
					Wood:      2,
					Buildings: []game.Building{},
				},
			},
		},
		Marquise: game.MarquiseState{
			SawmillsPlaced: 1,
		},
	}

	action := game.Action{
		Type: game.ActionBuild,
		Build: &game.BuildAction{
			Faction:      game.Marquise,
			ClearingID:   3,
			BuildingType: game.Sawmill,
			WoodSources: []game.WoodSource{
				{ClearingID: 3, Amount: 1},
			},
		},
	}

	next := ApplyAction(state, action)

	if len(next.Map.Clearings[0].Buildings) != 1 {
		t.Fatalf("expected 1 building after build, got %d", len(next.Map.Clearings[0].Buildings))
	}
	built := next.Map.Clearings[0].Buildings[0]
	if built.Faction != game.Marquise || built.Type != game.Sawmill {
		t.Fatalf("expected marquise sawmill to be built, got %+v", built)
	}
	if next.Marquise.SawmillsPlaced != 2 {
		t.Fatalf("expected sawmills placed to increase to 2, got %d", next.Marquise.SawmillsPlaced)
	}
	if next.Map.Clearings[0].Wood != 1 {
		t.Fatalf("expected build to deduct 1 wood, got %d", next.Map.Clearings[0].Wood)
	}
	if next.VictoryPoints[game.Marquise] != 1 {
		t.Fatalf("expected second sawmill to score 1 point, got %d", next.VictoryPoints[game.Marquise])
	}

	if len(state.Map.Clearings[0].Buildings) != 0 {
		t.Fatalf("expected original clearing buildings to remain empty, got %+v", state.Map.Clearings[0].Buildings)
	}
	if state.Marquise.SawmillsPlaced != 1 {
		t.Fatalf("expected original sawmills placed to remain 1, got %d", state.Marquise.SawmillsPlaced)
	}
	if state.Map.Clearings[0].Wood != 2 {
		t.Fatalf("expected original clearing wood to remain 2, got %d", state.Map.Clearings[0].Wood)
	}
}

func TestApplyActionBattleResolutionRemovesWarriors(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 3,
						game.Eyrie:    2,
					},
				},
			},
		},
	}

	action := game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:        game.Marquise,
			ClearingID:     1,
			TargetFaction:  game.Eyrie,
			AttackerRoll:   2,
			DefenderRoll:   1,
			AttackerLosses: 1,
			DefenderLosses: 2,
		},
	}

	next := ApplyAction(state, action)

	if next.Map.Clearings[0].Warriors[game.Marquise] != 2 {
		t.Fatalf("expected marquise warriors to decrease to 2, got %d", next.Map.Clearings[0].Warriors[game.Marquise])
	}
	if next.Map.Clearings[0].Warriors[game.Eyrie] != 0 {
		t.Fatalf("expected eyrie warriors to decrease to 0, got %d", next.Map.Clearings[0].Warriors[game.Eyrie])
	}

	if state.Map.Clearings[0].Warriors[game.Marquise] != 3 {
		t.Fatalf("expected original marquise warriors to remain 3, got %d", state.Map.Clearings[0].Warriors[game.Marquise])
	}
	if state.Map.Clearings[0].Warriors[game.Eyrie] != 2 {
		t.Fatalf("expected original eyrie warriors to remain 2, got %d", state.Map.Clearings[0].Warriors[game.Eyrie])
	}
}

func TestApplyActionBattleResolutionSpillsIntoBuildings(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Eyrie: 1,
					},
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Workshop},
						{Faction: game.Eyrie, Type: game.Sawmill},
					},
				},
			},
		},
		Marquise: game.MarquiseState{
			WorkshopsPlaced: 1,
		},
	}

	action := game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:        game.Eyrie,
			ClearingID:     1,
			TargetFaction:  game.Marquise,
			AttackerRoll:   3,
			DefenderRoll:   0,
			AttackerLosses: 0,
			DefenderLosses: 1,
		},
	}

	next := ApplyAction(state, action)

	if len(next.Map.Clearings[0].Buildings) != 1 {
		t.Fatalf("expected one building left after battle resolution, got %d", len(next.Map.Clearings[0].Buildings))
	}
	if next.Map.Clearings[0].Buildings[0].Faction != game.Eyrie {
		t.Fatalf("expected remaining building to belong to eyrie, got %+v", next.Map.Clearings[0].Buildings[0])
	}
	if next.Marquise.WorkshopsPlaced != 0 {
		t.Fatalf("expected marquise workshop count to decrease to 0, got %d", next.Marquise.WorkshopsPlaced)
	}

	if len(state.Map.Clearings[0].Buildings) != 2 {
		t.Fatalf("expected original buildings to remain unchanged, got %+v", state.Map.Clearings[0].Buildings)
	}
	if state.Marquise.WorkshopsPlaced != 1 {
		t.Fatalf("expected original workshop count to remain 1, got %d", state.Marquise.WorkshopsPlaced)
	}
}

func TestApplyActionCraft(t *testing.T) {
	state := game.GameState{
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 20, Name: "Crafted Card"},
				{ID: 21, Name: "Other Card"},
			},
		},
		TurnProgress: game.TurnProgress{
			UsedWorkshopClearings: []int{1},
		},
	}

	action := game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction:               game.Marquise,
			CardID:                20,
			UsedWorkshopClearings: []int{2, 2},
		},
	}

	next := ApplyAction(state, action)

	if hasCard(next.Marquise.CardsInHand, 20) {
		t.Fatalf("expected crafted card 20 to be removed from hand")
	}
	if !hasCard(next.Marquise.CardsInHand, 21) {
		t.Fatalf("expected other card 21 to remain in hand")
	}
	if len(next.TurnProgress.UsedWorkshopClearings) != 3 {
		t.Fatalf("expected 3 recorded used workshops, got %d", len(next.TurnProgress.UsedWorkshopClearings))
	}
	if next.TurnProgress.UsedWorkshopClearings[0] != 1 ||
		next.TurnProgress.UsedWorkshopClearings[1] != 2 ||
		next.TurnProgress.UsedWorkshopClearings[2] != 2 {
		t.Fatalf("expected used workshops [1 2 2], got %+v", next.TurnProgress.UsedWorkshopClearings)
	}

	if !hasCard(state.Marquise.CardsInHand, 20) {
		t.Fatalf("expected original hand to still contain card 20")
	}
	if len(state.TurnProgress.UsedWorkshopClearings) != 1 || state.TurnProgress.UsedWorkshopClearings[0] != 1 {
		t.Fatalf("expected original used workshops to remain [1], got %+v", state.TurnProgress.UsedWorkshopClearings)
	}
}

func TestApplyRecruitChangesSubsequentValidActions(t *testing.T) {
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

	recruit := game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Marquise,
			ClearingIDs: []int{1},
		},
	}

	next := ApplyAction(state, recruit)
	actions := ValidActions(next)
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

	if containsAction(actions, recruit) {
		t.Fatalf("did not expect recruit action to remain legal after applying recruit, got %+v", actions)
	}
	if !containsAction(actions, wantMovement) {
		t.Fatalf("expected movement action %+v after recruit transition, got %+v", wantMovement, actions)
	}
}
