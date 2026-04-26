package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

func firstVagabondTestCard(t *testing.T, suit game.Suit) game.Card {
	t.Helper()

	for _, card := range carddata.BaseDeck() {
		if card.Suit == suit {
			return card
		}
	}

	t.Fatalf("no card found for suit %v", suit)
	return game.Card{}
}

func TestResolveBattleVagabondCapsHitsByExhaustedSwords(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 3,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemSword, Status: game.ItemExhausted},
				{Type: game.ItemSword, Status: game.ItemReady},
			},
		},
	}

	action := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Vagabond,
			ClearingID:    1,
			TargetFaction: game.Marquise,
		},
	}

	resolved := ResolveBattle(state, action, 3, 0)
	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if resolved.BattleResolution.DefenderLosses != 2 {
		t.Fatalf("expected vagabond hits to be capped at 2 swords, got %d", resolved.BattleResolution.DefenderLosses)
	}
}

func TestApplyDaybreakMarksVagabondRefreshResolved(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Vagabond: game.VagabondState{
			Items: []game.Item{
				{Type: game.ItemBoots, Status: game.ItemExhausted},
			},
		},
	}

	action := game.Action{
		Type: game.ActionDaybreak,
		Daybreak: &game.DaybreakAction{
			Faction:              game.Vagabond,
			RefreshedItemIndexes: []int{0},
		},
	}

	next := ApplyAction(state, action)
	if !next.TurnProgress.HasRefreshed {
		t.Fatalf("expected daybreak to mark refresh resolved")
	}
	if next.Vagabond.Items[0].Status != game.ItemReady {
		t.Fatalf("expected exhausted item to refresh, got %v", next.Vagabond.Items[0].Status)
	}
	if next.CurrentPhase != game.Birdsong || next.CurrentStep != game.StepBirdsong {
		t.Fatalf("expected vagabond to remain in birdsong after refresh, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
}

func TestApplyActionVagabondMovementExhaustsBootsAndMoves(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{ID: 1},
				{ID: 2},
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

	action := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Vagabond,
			Count:    1,
			MaxCount: 1,
			From:     1,
			To:       2,
		},
	}

	next := ApplyAction(state, action)
	if next.Vagabond.ClearingID != 2 {
		t.Fatalf("expected vagabond to move to clearing 2, got %d", next.Vagabond.ClearingID)
	}
	if next.Vagabond.Items[0].Status != game.ItemExhausted {
		t.Fatalf("expected boots to exhaust after moving, got %v", next.Vagabond.Items[0].Status)
	}

	invalidForestMove := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:    game.Vagabond,
			Count:      1,
			MaxCount:   1,
			From:       2,
			ToForestID: 1,
		},
	}

	next.Vagabond.Items[0].Status = game.ItemReady
	next = ApplyAction(next, invalidForestMove)
	if next.Vagabond.InForest || next.Vagabond.ForestID != 0 || next.Vagabond.ClearingID != 2 {
		t.Fatalf("expected daylight move to forest to be ignored, got %+v", next.Vagabond)
	}
	if next.Vagabond.Items[0].Status != game.ItemReady {
		t.Fatalf("expected invalid forest move not to exhaust boots, got %v", next.Vagabond.Items[0].Status)
	}
}

func TestApplyActionExploreTakesRuinItemAndScores(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:        6,
					Ruins:     true,
					RuinItems: []game.ItemType{game.ItemCoin},
				},
			},
		},
		VictoryPoints: map[game.Faction]int{
			game.Vagabond: 0,
		},
		Vagabond: game.VagabondState{
			ClearingID: 6,
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
			},
		},
	}

	action := game.Action{
		Type: game.ActionExplore,
		Explore: &game.ExploreAction{
			Faction:    game.Vagabond,
			ClearingID: 6,
			ItemType:   game.ItemCoin,
		},
	}

	next := ApplyAction(state, action)
	if next.Vagabond.Items[0].Status != game.ItemExhausted {
		t.Fatalf("expected torch to exhaust, got %v", next.Vagabond.Items[0].Status)
	}
	if len(next.Vagabond.Items) != 2 || next.Vagabond.Items[1].Type != game.ItemCoin {
		t.Fatalf("expected coin item to be gained, got %+v", next.Vagabond.Items)
	}
	if next.Map.Clearings[0].Ruins {
		t.Fatalf("expected ruins to be cleared after taking last ruin item")
	}
	if next.VictoryPoints[game.Vagabond] != 1 {
		t.Fatalf("expected explore to score 1 VP, got %d", next.VictoryPoints[game.Vagabond])
	}
}

func TestApplyActionAidTransfersCardAndImprovesRelationship(t *testing.T) {
	foxCard := firstVagabondTestCard(t, game.Fox)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{ID: 1, Suit: game.Fox},
			},
		},
		Vagabond: game.VagabondState{
			CardsInHand: []game.Card{foxCard},
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
			},
		},
	}

	action := game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: game.Marquise,
			ClearingID:    1,
			CardID:        foxCard.ID,
			ItemIndex:     1,
		},
	}

	next := ApplyAction(state, action)
	if len(next.Vagabond.CardsInHand) != 0 {
		t.Fatalf("expected aided card to leave vagabond hand, got %+v", next.Vagabond.CardsInHand)
	}
	if len(next.Marquise.CardsInHand) != 1 || next.Marquise.CardsInHand[0].ID != foxCard.ID {
		t.Fatalf("expected card to move to marquise hand, got %+v", next.Marquise.CardsInHand)
	}
	if next.Vagabond.Items[0].Status != game.ItemReady || next.Vagabond.Items[1].Status != game.ItemExhausted {
		t.Fatalf("expected aid to exhaust selected item only, got %+v", next.Vagabond.Items)
	}
	if vagabondRelationshipLevel(next, game.Marquise) != game.RelFriendly {
		t.Fatalf("expected relationship to improve to friendly, got %v", vagabondRelationshipLevel(next, game.Marquise))
	}
}

func TestApplyActionBattleResolutionDamagesVagabondItems(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 2,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemSword, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
				{Type: game.ItemTorch, Status: game.ItemReady},
			},
		},
	}

	action := game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:        game.Marquise,
			ClearingID:     1,
			TargetFaction:  game.Vagabond,
			AttackerLosses: 0,
			DefenderLosses: 2,
		},
	}

	next := ApplyAction(state, action)
	damaged := 0
	for _, item := range next.Vagabond.Items {
		if item.Status == game.ItemDamaged {
			damaged++
		}
	}
	if damaged != 2 {
		t.Fatalf("expected 2 damaged vagabond items, got %+v", next.Vagabond.Items)
	}
}

func TestApplyActionVagabondEveningInForestRepairsDamagedItems(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		Vagabond: game.VagabondState{
			InForest: true,
			Items: []game.Item{
				{Type: game.ItemSword, Status: game.ItemDamaged},
				{Type: game.ItemBoots, Status: game.ItemDamaged},
			},
		},
	}

	action := game.Action{
		Type: game.ActionEveningDraw,
		EveningDraw: &game.EveningDrawAction{
			Faction: game.Vagabond,
			Count:   1,
		},
	}

	next := ApplyAction(state, action)
	for _, item := range next.Vagabond.Items {
		if item.Status != game.ItemReady {
			t.Fatalf("expected forest rest to repair all items, got %+v", next.Vagabond.Items)
		}
	}
}

func TestApplyActionQuestExhaustsRequiredItemsAndScores(t *testing.T) {
	state := game.GameState{
		VictoryPoints: map[game.Faction]int{
			game.Vagabond: 0,
		},
		Vagabond: game.VagabondState{
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
				{Type: game.ItemSword, Status: game.ItemReady},
			},
			QuestsAvailable: []game.Quest{
				{
					ID:            9,
					Name:          "Scavenge",
					Suit:          game.Fox,
					RequiredItems: []game.ItemType{game.ItemTorch, game.ItemBoots},
				},
			},
		},
	}

	action := game.Action{
		Type: game.ActionQuest,
		Quest: &game.QuestAction{
			Faction:     game.Vagabond,
			QuestID:     9,
			ItemIndexes: []int{0, 1},
			Reward:      game.QuestRewardVictoryPoints,
		},
	}

	next := ApplyAction(state, action)
	if next.Vagabond.Items[0].Status != game.ItemExhausted || next.Vagabond.Items[1].Status != game.ItemExhausted {
		t.Fatalf("expected quest to exhaust required items, got %+v", next.Vagabond.Items)
	}
	if next.Vagabond.Items[2].Status != game.ItemReady {
		t.Fatalf("expected unrelated item to stay ready, got %+v", next.Vagabond.Items)
	}
	if len(next.Vagabond.QuestsAvailable) != 0 || len(next.Vagabond.QuestsCompleted) != 1 {
		t.Fatalf("expected quest to move from available to completed, got available=%+v completed=%+v", next.Vagabond.QuestsAvailable, next.Vagabond.QuestsCompleted)
	}
	if next.VictoryPoints[game.Vagabond] != 1 {
		t.Fatalf("expected first completed fox quest to score 1 VP, got %d", next.VictoryPoints[game.Vagabond])
	}
}
