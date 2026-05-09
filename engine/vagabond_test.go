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

func vagabondTestCardsOfSuit(t *testing.T, suit game.Suit, count int) []game.Card {
	t.Helper()

	cards := []game.Card{}
	for _, card := range carddata.BaseDeck() {
		if card.Suit == suit {
			cards = append(cards, card)
			if len(cards) == count {
				return cards
			}
		}
	}

	t.Fatalf("only found %d cards for suit %v, needed %d", len(cards), suit, count)
	return nil
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

func TestResolveBattleVagabondWithoutUndamagedSwordIsDefenseless(t *testing.T) {
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
				{Type: game.ItemSword, Status: game.ItemDamaged},
				{Type: game.ItemBoots, Status: game.ItemReady},
			},
		},
	}

	action := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Vagabond,
		},
	}

	resolved := ResolveBattle(state, action, 0, 3)
	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if resolved.BattleResolution.DefenderLosses != 1 {
		t.Fatalf("expected Vagabond to take defenseless hit with no undamaged sword, got %d", resolved.BattleResolution.DefenderLosses)
	}
	if resolved.BattleResolution.AttackerLosses != 0 {
		t.Fatalf("expected Vagabond with no undamaged sword to deal no rolled hits, got %d", resolved.BattleResolution.AttackerLosses)
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
	if next.Vagabond.Items[0].Zone != game.ItemZoneSatchel {
		t.Fatalf("expected refreshed boots to stay in satchel, got %v", next.Vagabond.Items[0].Zone)
	}
	if next.CurrentPhase != game.Birdsong || next.CurrentStep != game.StepBirdsong {
		t.Fatalf("expected vagabond to remain in birdsong after refresh, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
}

func TestApplyDaybreakMovesTrackItemsBetweenSatchelAndTrack(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Vagabond: game.VagabondState{
			Items: []game.Item{
				{Type: game.ItemCoin, Status: game.ItemExhausted, Zone: game.ItemZoneSatchel},
				{Type: game.ItemTea, Status: game.ItemReady, Zone: game.ItemZoneTrack},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionDaybreak,
		Daybreak: &game.DaybreakAction{
			Faction:              game.Vagabond,
			RefreshedItemIndexes: []int{0},
		},
	})

	if next.Vagabond.Items[0].Status != game.ItemReady || next.Vagabond.Items[0].Zone != game.ItemZoneTrack {
		t.Fatalf("expected refreshed coin to return to track, got %+v", next.Vagabond.Items[0])
	}
	if next.Vagabond.Items[1].Status != game.ItemReady || next.Vagabond.Items[1].Zone != game.ItemZoneTrack {
		t.Fatalf("expected refresh tea to stay ready on track, got %+v", next.Vagabond.Items[1])
	}
}

func TestApplyDaybreakUsesTwoRefreshesPerReadyTrackTea(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Vagabond: game.VagabondState{
			Items: []game.Item{
				{Type: game.ItemTea, Status: game.ItemReady, Zone: game.ItemZoneTrack},
				{Type: game.ItemBoots, Status: game.ItemExhausted, Zone: game.ItemZoneSatchel},
				{Type: game.ItemSword, Status: game.ItemExhausted, Zone: game.ItemZoneSatchel},
				{Type: game.ItemTorch, Status: game.ItemExhausted, Zone: game.ItemZoneSatchel},
				{Type: game.ItemHammer, Status: game.ItemExhausted, Zone: game.ItemZoneSatchel},
				{Type: game.ItemCrossbow, Status: game.ItemExhausted, Zone: game.ItemZoneSatchel},
				{Type: game.ItemBoots, Status: game.ItemExhausted, Zone: game.ItemZoneSatchel},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionDaybreak,
		Daybreak: &game.DaybreakAction{
			Faction:              game.Vagabond,
			RefreshedItemIndexes: []int{1, 2, 3, 4, 5, 6},
		},
	})

	ready := 0
	for _, item := range next.Vagabond.Items {
		if item.Status == game.ItemReady {
			ready++
		}
	}
	if ready != 6 {
		t.Fatalf("expected ready tea plus five refreshed items, got %+v", next.Vagabond.Items)
	}
	if next.Vagabond.Items[6].Status != game.ItemExhausted {
		t.Fatalf("expected direct refresh application to stop at legal limit, got %+v", next.Vagabond.Items[6])
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
				{
					ID:   1,
					Suit: game.Fox,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
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
	if vagabondRelationshipLevel(next, game.Marquise) != game.RelAmiable {
		t.Fatalf("expected relationship to improve to amiable, got %v", vagabondRelationshipLevel(next, game.Marquise))
	}
	if next.VictoryPoints[game.Vagabond] != 1 {
		t.Fatalf("expected first relationship improvement to score 1 VP, got %+v", next.VictoryPoints)
	}
}

func TestApplyActionAidRequiresThresholdsForLaterRelationshipSpaces(t *testing.T) {
	cards := vagabondTestCardsOfSuit(t, game.Fox, 2)
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			CardsInHand: cards,
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
			},
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelAmiable,
			},
		},
		VictoryPoints: map[game.Faction]int{},
	}

	afterFirst := ApplyAction(state, game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: game.Marquise,
			ClearingID:    1,
			CardID:        cards[0].ID,
			ItemIndex:     0,
		},
	})

	if afterFirst.Vagabond.Relationships[game.Marquise] != game.RelAmiable {
		t.Fatalf("expected first amiable aid not to improve yet, got %+v", afterFirst.Vagabond.Relationships)
	}
	if afterFirst.TurnProgress.VagabondAidCounts[game.Marquise] != 1 {
		t.Fatalf("expected one aid progress toward friendly, got %+v", afterFirst.TurnProgress.VagabondAidCounts)
	}
	if afterFirst.VictoryPoints[game.Vagabond] != 0 {
		t.Fatalf("expected no VP before relationship threshold, got %+v", afterFirst.VictoryPoints)
	}

	afterSecond := ApplyAction(afterFirst, game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: game.Marquise,
			ClearingID:    1,
			CardID:        cards[1].ID,
			ItemIndex:     1,
		},
	})

	if afterSecond.Vagabond.Relationships[game.Marquise] != game.RelFriendly {
		t.Fatalf("expected second amiable aid to improve to friendly, got %+v", afterSecond.Vagabond.Relationships)
	}
	if afterSecond.TurnProgress.VagabondAidCounts[game.Marquise] != 0 {
		t.Fatalf("expected aid progress to reset after improvement, got %+v", afterSecond.TurnProgress.VagabondAidCounts)
	}
	if afterSecond.VictoryPoints[game.Vagabond] != 2 {
		t.Fatalf("expected friendly improvement to score 2 VP, got %+v", afterSecond.VictoryPoints)
	}
}

func TestApplyActionAidAlliedFactionScoresTwoVP(t *testing.T) {
	foxCard := firstVagabondTestCard(t, game.Fox)
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			CardsInHand: []game.Card{foxCard},
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
			},
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelAllied,
			},
		},
		VictoryPoints: map[game.Faction]int{},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: game.Marquise,
			ClearingID:    1,
			CardID:        foxCard.ID,
			ItemIndex:     0,
		},
	})

	if next.Vagabond.Relationships[game.Marquise] != game.RelAllied {
		t.Fatalf("expected allied relationship to remain allied, got %+v", next.Vagabond.Relationships)
	}
	if next.VictoryPoints[game.Vagabond] != 2 {
		t.Fatalf("expected allied aid to score 2 VP, got %+v", next.VictoryPoints)
	}
}

func TestApplyActionAidHostileFactionDoesNotImproveRelationship(t *testing.T) {
	foxCard := firstVagabondTestCard(t, game.Fox)
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
			},
		},
		Vagabond: game.VagabondState{
			CardsInHand: []game.Card{foxCard},
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
			},
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelHostile,
			},
		},
		VictoryPoints: map[game.Faction]int{},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: game.Marquise,
			ClearingID:    1,
			CardID:        foxCard.ID,
			ItemIndex:     0,
		},
	})

	if next.Vagabond.Relationships[game.Marquise] != game.RelHostile {
		t.Fatalf("expected hostile aid not to improve relationship, got %+v", next.Vagabond.Relationships)
	}
	if next.VictoryPoints[game.Vagabond] != 0 {
		t.Fatalf("expected hostile aid not to score relationship VP, got %+v", next.VictoryPoints)
	}
	if len(next.Marquise.CardsInHand) != 1 || next.Vagabond.Items[0].Status != game.ItemExhausted {
		t.Fatalf("expected hostile aid to still transfer card and exhaust item, cards=%+v items=%+v", next.Marquise.CardsInHand, next.Vagabond.Items)
	}
}

func TestApplyActionAidRequiresTargetPiecesInClearing(t *testing.T) {
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
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: game.Marquise,
			ClearingID:    1,
			CardID:        foxCard.ID,
			ItemIndex:     0,
		},
	})

	if len(next.Vagabond.CardsInHand) != 1 || next.Vagabond.Items[0].Status != game.ItemReady {
		t.Fatalf("expected invalid aid to leave hand and item unchanged, hand=%+v items=%+v", next.Vagabond.CardsInHand, next.Vagabond.Items)
	}
}

func TestApplyActionStrikeReturnsRemovedWarriorToSupply(t *testing.T) {
	state := game.GameState{
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
		Marquise: game.MarquiseState{
			WarriorSupply: 3,
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemCrossbow, Status: game.ItemReady},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionStrike,
		Strike: &game.StrikeAction{
			Faction:       game.Vagabond,
			ClearingID:    1,
			TargetFaction: game.Marquise,
		},
	})

	if next.Map.Clearings[0].Warriors[game.Marquise] != 0 {
		t.Fatalf("expected strike to remove marquise warrior, got %+v", next.Map.Clearings[0].Warriors)
	}
	if next.Marquise.WarriorSupply != 4 {
		t.Fatalf("expected struck marquise warrior to return to supply, got %d", next.Marquise.WarriorSupply)
	}
	if next.Vagabond.Items[0].Status != game.ItemExhausted {
		t.Fatalf("expected strike to exhaust crossbow, got %+v", next.Vagabond.Items)
	}
}

func TestApplyActionStrikeRequiresCrossbow(t *testing.T) {
	state := game.GameState{
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
		Marquise: game.MarquiseState{
			WarriorSupply: 3,
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemSword, Status: game.ItemReady},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionStrike,
		Strike: &game.StrikeAction{
			Faction:       game.Vagabond,
			ClearingID:    1,
			TargetFaction: game.Marquise,
		},
	})

	if next.Map.Clearings[0].Warriors[game.Marquise] != 1 {
		t.Fatalf("expected sword-only strike to leave marquise warrior, got %+v", next.Map.Clearings[0].Warriors)
	}
	if next.Marquise.WarriorSupply != 3 {
		t.Fatalf("expected sword-only strike not to return warrior to supply, got %d", next.Marquise.WarriorSupply)
	}
	if next.Vagabond.Items[0].Status != game.ItemReady {
		t.Fatalf("expected sword-only strike not to exhaust sword, got %+v", next.Vagabond.Items)
	}
}

func TestApplyActionStrikeRemovingAllianceBaseAppliesFallout(t *testing.T) {
	foxCard := firstVagabondTestCard(t, game.Fox)
	birdCard := firstVagabondTestCard(t, game.Bird)
	rabbitCard := firstVagabondTestCard(t, game.Rabbit)
	state := game.GameState{
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
		Alliance: game.AllianceState{
			Supporters:    []game.Card{foxCard, birdCard, rabbitCard},
			Officers:      2,
			FoxBasePlaced: true,
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemCrossbow, Status: game.ItemReady},
			},
		},
		VictoryPoints: map[game.Faction]int{},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionStrike,
		Strike: &game.StrikeAction{
			Faction:       game.Vagabond,
			ClearingID:    1,
			TargetFaction: game.Alliance,
		},
	})

	if len(next.Map.Clearings[0].Buildings) != 0 {
		t.Fatalf("expected strike to remove Alliance base, got %+v", next.Map.Clearings[0].Buildings)
	}
	if next.Alliance.FoxBasePlaced {
		t.Fatalf("expected fox base flag to clear")
	}
	if next.Alliance.Officers != 1 {
		t.Fatalf("expected base removal to remove half of 2 officers rounded up, got %d", next.Alliance.Officers)
	}
	if len(next.Alliance.Supporters) != 1 || next.Alliance.Supporters[0].ID != rabbitCard.ID {
		t.Fatalf("expected only off-suit supporter to remain, got %+v", next.Alliance.Supporters)
	}
	if len(next.DiscardPile) != 2 || next.DiscardPile[0] != foxCard.ID || next.DiscardPile[1] != birdCard.ID {
		t.Fatalf("expected fox and bird supporters in discard, got %+v", next.DiscardPile)
	}
	if next.VictoryPoints[game.Vagabond] != 1 {
		t.Fatalf("expected Vagabond to score removed base, got %+v", next.VictoryPoints)
	}
	if next.Vagabond.Items[0].Status != game.ItemExhausted {
		t.Fatalf("expected strike to exhaust crossbow, got %+v", next.Vagabond.Items)
	}
}

func TestVagabondBattleScoresInfamyForHostilePieces(t *testing.T) {
	state := game.GameState{
		GamePhase:   game.LifecyclePlaying,
		FactionTurn: game.Vagabond,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Workshop},
					},
				},
			},
		},
		Marquise: game.MarquiseState{
			WorkshopsPlaced: 1,
		},
		Vagabond: game.VagabondState{
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelHostile,
			},
		},
		VictoryPoints: map[game.Faction]int{},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:        game.Vagabond,
			ClearingID:     1,
			TargetFaction:  game.Marquise,
			DefenderLosses: 2,
			DefenderPieceLosses: []game.BattlePieceLoss{
				{Kind: game.BattlePieceBuilding, BuildingType: game.Workshop},
			},
		},
	})

	if next.VictoryPoints[game.Vagabond] != 3 {
		t.Fatalf("expected 1 building VP plus 2 infamy VP, got %+v", next.VictoryPoints)
	}
}

func TestVagabondBattleOnlyTurnsNonHostileFactionHostileForWarriorRemoval(t *testing.T) {
	state := game.GameState{
		GamePhase:   game.LifecyclePlaying,
		FactionTurn: game.Vagabond,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Buildings: []game.Building{
						{Faction: game.Marquise, Type: game.Workshop},
					},
				},
			},
		},
		Marquise: game.MarquiseState{
			WorkshopsPlaced: 1,
		},
		Vagabond: game.VagabondState{
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelIndifferent,
			},
		},
		VictoryPoints: map[game.Faction]int{},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:        game.Vagabond,
			ClearingID:     1,
			TargetFaction:  game.Marquise,
			DefenderLosses: 1,
			DefenderPieceLosses: []game.BattlePieceLoss{
				{Kind: game.BattlePieceBuilding, BuildingType: game.Workshop},
			},
		},
	})

	if next.Vagabond.Relationships[game.Marquise] != game.RelIndifferent {
		t.Fatalf("expected building removal not to make Marquise hostile, got %+v", next.Vagabond.Relationships)
	}
	if next.VictoryPoints[game.Vagabond] != 1 {
		t.Fatalf("expected only building VP with no infamy, got %+v", next.VictoryPoints)
	}
}

func TestVagabondBattleNewHostileInfamySkipsFirstWarrior(t *testing.T) {
	state := game.GameState{
		GamePhase:   game.LifecyclePlaying,
		FactionTurn: game.Vagabond,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 2,
					},
					Tokens: []game.Token{
						{Faction: game.Marquise, Type: game.TokenKeep},
					},
				},
			},
		},
		Marquise: game.MarquiseState{
			KeepClearingID: 1,
		},
		Vagabond: game.VagabondState{
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelIndifferent,
			},
		},
		VictoryPoints: map[game.Faction]int{},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:        game.Vagabond,
			ClearingID:     1,
			TargetFaction:  game.Marquise,
			DefenderLosses: 3,
			DefenderPieceLosses: []game.BattlePieceLoss{
				{Kind: game.BattlePieceToken, TokenType: game.TokenKeep},
			},
		},
	})

	if next.Vagabond.Relationships[game.Marquise] != game.RelHostile {
		t.Fatalf("expected warrior removal to make Marquise hostile, got %+v", next.Vagabond.Relationships)
	}
	if next.VictoryPoints[game.Vagabond] != 3 {
		t.Fatalf("expected 1 token VP plus 2 infamy VP after skipping first warrior, got %+v", next.VictoryPoints)
	}
}

func TestResolveBattleVagabondHitCapCountsAllUndamagedSwords(t *testing.T) {
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
				{Type: game.ItemSword, Status: game.ItemReady},
				{Type: game.ItemSword, Status: game.ItemReady},
				{Type: game.ItemSword, Status: game.ItemExhausted},
				{Type: game.ItemSword, Status: game.ItemDamaged},
			},
		},
	}

	resolved := ResolveBattle(state, game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Vagabond,
			ClearingID:    1,
			TargetFaction: game.Marquise,
		},
	}, 3, 0)

	if resolved.BattleResolution == nil || resolved.BattleResolution.DefenderLosses != 3 {
		t.Fatalf("expected three undamaged swords to cap Vagabond hits at 3, got %+v", resolved)
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
				{Type: game.ItemTea, Status: game.ItemDamaged},
			},
		},
	}

	action := game.Action{
		Type: game.ActionVagabondRest,
		VagabondRest: &game.VagabondRestAction{
			Faction: game.Vagabond,
		},
	}

	next := ApplyAction(state, action)
	for _, item := range next.Vagabond.Items {
		if item.Status != game.ItemReady {
			t.Fatalf("expected forest rest to repair all items, got %+v", next.Vagabond.Items)
		}
	}
	if next.Vagabond.Items[0].Zone != game.ItemZoneSatchel {
		t.Fatalf("expected repaired sword to return to satchel, got %+v", next.Vagabond.Items[0])
	}
	if next.Vagabond.Items[1].Zone != game.ItemZoneTrack {
		t.Fatalf("expected repaired tea to return to track, got %+v", next.Vagabond.Items[1])
	}
}

func TestApplyVagabondDiscardRemovesSelectedCards(t *testing.T) {
	foxCard := firstVagabondTestCard(t, game.Fox)
	rabbitCard := firstVagabondTestCard(t, game.Rabbit)
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		Vagabond: game.VagabondState{
			CardsInHand: []game.Card{foxCard, rabbitCard},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionVagabondDiscard,
		VagabondDiscard: &game.VagabondDiscardAction{
			Faction: game.Vagabond,
			CardIDs: []game.CardID{
				foxCard.ID,
			},
		},
	})

	if len(next.Vagabond.CardsInHand) != 1 || next.Vagabond.CardsInHand[0].ID != rabbitCard.ID {
		t.Fatalf("expected only rabbit card to remain, got %+v", next.Vagabond.CardsInHand)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != foxCard.ID {
		t.Fatalf("expected fox card in discard pile, got %+v", next.DiscardPile)
	}
	if !next.TurnProgress.VagabondDiscardResolved {
		t.Fatalf("expected discard step to be marked resolved")
	}
}

func TestApplyVagabondItemCapacityRemovesSelectedItems(t *testing.T) {
	state := game.GameState{
		FactionTurn:  game.Vagabond,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		Vagabond: game.VagabondState{
			Items: []game.Item{
				{Type: game.ItemBoots, Status: game.ItemReady},
				{Type: game.ItemSword, Status: game.ItemReady},
				{Type: game.ItemTorch, Status: game.ItemReady},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionVagabondItemCapacity,
		VagabondCapacity: &game.VagabondItemCapacityAction{
			Faction:     game.Vagabond,
			ItemIndexes: []int{1},
		},
	})

	if len(next.Vagabond.Items) != 2 {
		t.Fatalf("expected one item removed, got %+v", next.Vagabond.Items)
	}
	if next.Vagabond.Items[0].Type != game.ItemBoots || next.Vagabond.Items[1].Type != game.ItemTorch {
		t.Fatalf("expected selected sword to be removed, got %+v", next.Vagabond.Items)
	}
	if next.FactionTurn != game.Marquise {
		t.Fatalf("expected capacity check to advance turn, got %v", next.FactionTurn)
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
