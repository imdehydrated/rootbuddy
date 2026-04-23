package rules

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestValidVagabondBirdsongActionsIncludesDaybreakAndSlip(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:  1,
					Adj: []int{2},
				},
				{
					ID:  2,
					Adj: []int{1},
				},
			},
			Forests: []game.Forest{
				{ID: 1, AdjacentClearings: []int{1}},
			},
		},
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemBoots, Status: game.ItemExhausted},
				{Type: game.ItemSword, Status: game.ItemReady},
			},
		},
	}

	got := ValidVagabondBirdsongActions(state)
	wantDaybreak := game.Action{
		Type: game.ActionDaybreak,
		Daybreak: &game.DaybreakAction{
			Faction:              game.Vagabond,
			RefreshedItemIndexes: []int{0},
		},
	}
	wantSlip := game.Action{
		Type: game.ActionSlip,
		Slip: &game.SlipAction{
			Faction: game.Vagabond,
			From:    1,
			To:      2,
		},
	}

	if !containsAction(got, wantDaybreak) {
		t.Fatalf("expected daybreak action %+v, got %+v", wantDaybreak, got)
	}
	if !containsAction(got, wantSlip) {
		t.Fatalf("expected slip action %+v, got %+v", wantSlip, got)
	}
	wantStaySlip := game.Action{
		Type: game.ActionSlip,
		Slip: &game.SlipAction{
			Faction: game.Vagabond,
			From:    1,
			To:      1,
		},
	}
	if !containsAction(got, wantStaySlip) {
		t.Fatalf("expected stay slip action %+v, got %+v", wantStaySlip, got)
	}
	wantForestSlip := game.Action{
		Type: game.ActionSlip,
		Slip: &game.SlipAction{
			Faction:    game.Vagabond,
			From:       1,
			ToForestID: 1,
		},
	}
	if !containsAction(got, wantForestSlip) {
		t.Fatalf("expected forest slip action %+v, got %+v", wantForestSlip, got)
	}
	for _, action := range got {
		if action.Type == game.ActionPassPhase {
			t.Fatalf("did not expect pass before slip has been resolved, got %+v", got)
		}
	}
}

func TestValidVagabondBirdsongActionsAllowsPassAfterSlip(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		TurnProgress: game.TurnProgress{
			HasSlipped: true,
		},
	}

	got := ValidVagabondBirdsongActions(state)
	wantPass := game.Action{
		Type: game.ActionPassPhase,
		PassPhase: &game.PassPhaseAction{
			Faction: game.Vagabond,
		},
	}
	if !containsAction(got, wantPass) {
		t.Fatalf("expected pass after slip has been resolved, got %+v", got)
	}
}

func TestValidVagabondBirdsongActionsIncludesStayInForestSlip(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Vagabond: game.VagabondState{
			InForest: true,
			ForestID: 2,
		},
		Map: game.Map{
			Forests: []game.Forest{
				{ID: 2, AdjacentClearings: []int{1}},
			},
		},
	}

	got := ValidVagabondBirdsongActions(state)
	wantStay := game.Action{
		Type: game.ActionSlip,
		Slip: &game.SlipAction{
			Faction:      game.Vagabond,
			FromForestID: 2,
			ToForestID:   2,
		},
	}
	if !containsAction(got, wantStay) {
		t.Fatalf("expected stay-in-forest slip action %+v, got %+v", wantStay, got)
	}
}

func TestValidVagabondMoveActionsCountsHostileBootTax(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:  1,
					Adj: []int{2},
				},
				{
					ID:  2,
					Adj: []int{1},
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemBoots, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
			},
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelHostile,
			},
		},
	}

	got := ValidVagabondMoveActions(state)
	want := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Vagabond,
			Count:    2,
			MaxCount: 2,
			From:     1,
			To:       2,
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected hostile-taxed move %+v, got %+v", want, got)
	}
}

func TestValidVagabondMoveActionsCanMoveBetweenClearingAndForest(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:  1,
					Adj: []int{2},
				},
				{
					ID:  2,
					Adj: []int{1},
				},
			},
			Forests: []game.Forest{
				{ID: 1, AdjacentClearings: []int{1, 2}},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemBoots, Status: game.ItemReady},
			},
		},
	}

	got := ValidVagabondMoveActions(state)
	wantForestMove := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:    game.Vagabond,
			Count:      1,
			MaxCount:   1,
			From:       1,
			ToForestID: 1,
		},
	}

	if !containsAction(got, wantForestMove) {
		t.Fatalf("expected clearing-to-forest move %+v, got %+v", wantForestMove, got)
	}

	state.Vagabond.ClearingID = 0
	state.Vagabond.ForestID = 1
	state.Vagabond.InForest = true
	state.Vagabond.Items[0].Status = game.ItemReady

	got = ValidVagabondMoveActions(state)
	wantForestExit := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:      game.Vagabond,
			Count:        1,
			MaxCount:     1,
			To:           2,
			FromForestID: 1,
		},
	}

	if !containsAction(got, wantForestExit) {
		t.Fatalf("expected forest-to-clearing move %+v, got %+v", wantForestExit, got)
	}
}

func TestValidExploreActionsUsesRuinItems(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:        3,
					Ruins:     true,
					RuinItems: []game.ItemType{game.ItemCoin, game.ItemBag},
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 3,
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
			},
		},
	}

	got := ValidExploreActions(state)
	wantCoin := game.Action{
		Type: game.ActionExplore,
		Explore: &game.ExploreAction{
			Faction:    game.Vagabond,
			ClearingID: 3,
			ItemType:   game.ItemCoin,
		},
	}
	wantBag := game.Action{
		Type: game.ActionExplore,
		Explore: &game.ExploreAction{
			Faction:    game.Vagabond,
			ClearingID: 3,
			ItemType:   game.ItemBag,
		},
	}

	if !containsAction(got, wantCoin) || !containsAction(got, wantBag) {
		t.Fatalf("expected ruin-item explore actions, got %+v", got)
	}
}

func TestValidAidActionsSkipsHostileFactions(t *testing.T) {
	foxCard := firstCardOfSuit(t, game.Fox)
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
						game.Alliance: 1,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			CardsInHand: []game.Card{
				foxCard,
			},
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
			},
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelHostile,
			},
		},
	}

	got := ValidAidActions(state)
	wantAlliance := game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: game.Alliance,
			ClearingID:    1,
			CardID:        foxCard.ID,
		},
	}
	unwantMarquise := game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: game.Marquise,
			ClearingID:    1,
			CardID:        foxCard.ID,
		},
	}

	if !containsAction(got, wantAlliance) {
		t.Fatalf("expected alliance aid action %+v, got %+v", wantAlliance, got)
	}
	if containsAction(got, unwantMarquise) {
		t.Fatalf("did not expect hostile marquise aid action %+v, got %+v", unwantMarquise, got)
	}
}

func TestValidVagabondBattleActionsTargetsMarquiseWood(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Wood: 1,
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemSword, Status: game.ItemReady},
			},
		},
	}

	got := ValidVagabondBattleActions(state)
	want := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Vagabond,
			ClearingID:    1,
			TargetFaction: game.Marquise,
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected battle against marquise wood %+v, got %+v", want, got)
	}
}

func TestValidVagabondBattleActionsSkipsCoalitionPartner(t *testing.T) {
	state := game.GameState{
		CoalitionActive:  true,
		CoalitionPartner: game.Marquise,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemSword, Status: game.ItemReady},
			},
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelHostile,
			},
		},
	}

	if got := ValidVagabondBattleActions(state); len(got) != 0 {
		t.Fatalf("did not expect battle actions against coalition partner, got %+v", got)
	}
}

func TestValidVagabondMoveActionsIgnoresCoalitionPartnerHostileTax(t *testing.T) {
	state := game.GameState{
		CoalitionActive:  true,
		CoalitionPartner: game.Marquise,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:  1,
					Adj: []int{2},
				},
				{
					ID:  2,
					Adj: []int{1},
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemBoots, Status: game.ItemReady},
			},
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelHostile,
			},
		},
	}

	want := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Vagabond,
			Count:    1,
			MaxCount: 1,
			From:     1,
			To:       2,
		},
	}

	got := ValidVagabondMoveActions(state)
	if !containsAction(got, want) {
		t.Fatalf("expected coalition partner clearing to avoid hostile boot tax %+v, got %+v", want, got)
	}
}

func TestValidAidActionsAllowsCoalitionPartnerDespiteHostileRelationship(t *testing.T) {
	foxCard := firstCardOfSuit(t, game.Fox)
	state := game.GameState{
		CoalitionActive:  true,
		CoalitionPartner: game.Marquise,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			CardsInHand: []game.Card{
				foxCard,
			},
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
			},
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelHostile,
			},
		},
	}

	want := game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: game.Marquise,
			ClearingID:    1,
			CardID:        foxCard.ID,
		},
	}

	got := ValidAidActions(state)
	if !containsAction(got, want) {
		t.Fatalf("expected aid action for coalition partner %+v, got %+v", want, got)
	}
}

func TestValidQuestActionsUsesQuestRequirements(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
				{Type: game.ItemSword, Status: game.ItemReady},
			},
			QuestsAvailable: []game.Quest{
				{
					ID:            10,
					Name:          "Scavenge",
					Suit:          game.Fox,
					RequiredItems: []game.ItemType{game.ItemTorch, game.ItemBoots},
				},
			},
		},
	}

	got := ValidQuestActions(state)
	want := game.Action{
		Type: game.ActionQuest,
		Quest: &game.QuestAction{
			Faction:     game.Vagabond,
			QuestID:     10,
			ItemIndexes: []int{0, 1},
			Reward:      game.QuestRewardVictoryPoints,
		},
	}
	unwant := game.Action{
		Type: game.ActionQuest,
		Quest: &game.QuestAction{
			Faction:     game.Vagabond,
			QuestID:     10,
			ItemIndexes: []int{0, 2},
			Reward:      game.QuestRewardVictoryPoints,
		},
	}

	if !containsAction(got, want) {
		t.Fatalf("expected quest action %+v, got %+v", want, got)
	}
	if containsAction(got, unwant) {
		t.Fatalf("did not expect mismatched quest item choice %+v, got %+v", unwant, got)
	}
}
