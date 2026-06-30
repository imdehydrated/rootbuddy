package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

func firstAllianceTestCard(t *testing.T, suit game.Suit) game.Card {
	t.Helper()

	for _, card := range carddata.BaseDeck() {
		if card.Suit == suit {
			return card
		}
	}

	t.Fatalf("no card found for suit %v", suit)
	return game.Card{}
}

func allianceTestCards(t *testing.T, suit game.Suit, count int) []game.Card {
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

	t.Fatalf("only found %d card(s) for suit %v, wanted %d", len(cards), suit, count)
	return nil
}

func TestResolveBattleAllianceDefenderUsesHigherRoll(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 2,
						game.Alliance: 2,
					},
				},
			},
		},
	}

	action := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Alliance,
		},
	}

	resolved := ResolveBattle(state, action, 3, 1)
	if resolved.BattleResolution == nil {
		t.Fatalf("expected battle resolution payload")
	}
	if resolved.BattleResolution.DefenderLosses != 1 {
		t.Fatalf("expected attacker to use lower roll for 1 hit, got %d", resolved.BattleResolution.DefenderLosses)
	}
	if resolved.BattleResolution.AttackerLosses != 2 {
		t.Fatalf("expected alliance defender to use higher roll for 2 hits, got %d", resolved.BattleResolution.AttackerLosses)
	}
}

func TestApplySpreadSympathyRemovesSupportersAndScores(t *testing.T) {
	rabbitCard := firstAllianceTestCard(t, game.Rabbit)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Adj:  []int{2},
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Adj:  []int{1},
				},
			},
		},
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		VictoryPoints: map[game.Faction]int{
			game.Alliance: 0,
		},
		Alliance: game.AllianceState{
			SympathyPlaced: 1,
			Supporters:     []game.Card{rabbitCard},
		},
	}

	action := game.Action{
		Type: game.ActionSpreadSympathy,
		SpreadSympathy: &game.SpreadSympathyAction{
			Faction:          game.Alliance,
			ClearingID:       2,
			SupporterCardIDs: []game.CardID{rabbitCard.ID},
		},
	}

	next := ApplyAction(state, action)
	if len(next.Alliance.Supporters) != 0 {
		t.Fatalf("expected supporter to be spent, got %+v", next.Alliance.Supporters)
	}
	if next.Alliance.SympathyPlaced != 2 {
		t.Fatalf("expected sympathy count to increase to 2, got %d", next.Alliance.SympathyPlaced)
	}
	if next.VictoryPoints[game.Alliance] != 1 {
		t.Fatalf("expected second sympathy to score 1 VP, got %d", next.VictoryPoints[game.Alliance])
	}
	if len(next.Map.Clearings[1].Tokens) != 1 || next.Map.Clearings[1].Tokens[0].Type != game.TokenSympathy {
		t.Fatalf("expected sympathy token in clearing 2, got %+v", next.Map.Clearings[1].Tokens)
	}
	if !next.TurnProgress.SpreadSympathyStarted {
		t.Fatalf("expected spread sympathy to mark the Alliance Birdsong spread step as started")
	}
}

func TestAllianceBirdsongHidesRevoltAfterSpreadSympathyStarts(t *testing.T) {
	foxCard := firstAllianceTestCard(t, game.Fox)
	birdCard := firstAllianceTestCard(t, game.Bird)
	rabbitCard := firstAllianceTestCard(t, game.Rabbit)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Adj:  []int{2},
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Adj:  []int{1},
				},
			},
		},
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Birdsong,
		CurrentStep:  game.StepBirdsong,
		Alliance: game.AllianceState{
			SympathyPlaced: 1,
			Supporters:     []game.Card{foxCard, birdCard, rabbitCard},
		},
		TurnProgress: game.TurnProgress{
			SpreadSympathyStarted: true,
		},
	}

	actions := ValidActions(state)
	wantSpread := game.Action{
		Type: game.ActionSpreadSympathy,
		SpreadSympathy: &game.SpreadSympathyAction{
			Faction:          game.Alliance,
			ClearingID:       2,
			SupporterCardIDs: []game.CardID{rabbitCard.ID},
		},
	}
	unwantedRevolt := game.Action{
		Type: game.ActionRevolt,
		Revolt: &game.RevoltAction{
			Faction:          game.Alliance,
			ClearingID:       1,
			BaseSuit:         game.Fox,
			SupporterCardIDs: []game.CardID{foxCard.ID, birdCard.ID},
		},
	}

	if !containsAction(actions, wantSpread) {
		t.Fatalf("expected spread sympathy to remain available after the spread step starts, got %+v", actions)
	}
	if containsAction(actions, unwantedRevolt) {
		t.Fatalf("did not expect revolt after spread sympathy starts, got %+v", actions)
	}
}

func TestApplyRevoltRemovesEnemyPiecesAndPlacesBase(t *testing.T) {
	foxCard := firstAllianceTestCard(t, game.Fox)
	birdCard := firstAllianceTestCard(t, game.Bird)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:         1,
					Suit:       game.Fox,
					BuildSlots: 1,
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
						{Faction: game.Marquise, Type: game.TokenKeep},
					},
				},
				{
					ID:   2,
					Suit: game.Fox,
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
			},
		},
		VictoryPoints: map[game.Faction]int{
			game.Alliance: 0,
		},
		Marquise: game.MarquiseState{
			KeepClearingID: 1,
			WarriorSupply:  6,
			CardsInHand:    []game.Card{foxCard},
		},
		Alliance: game.AllianceState{
			Supporters:    []game.Card{foxCard, birdCard},
			WarriorSupply: 5,
		},
	}
	state.Map.Clearings[0].Warriors = map[game.Faction]int{
		game.Marquise: 1,
	}
	state.Map.Clearings[0].Buildings = []game.Building{
		{Faction: game.Marquise, Type: game.Sawmill},
	}
	state.Map.Clearings[0].Wood = 1

	action := game.Action{
		Type: game.ActionRevolt,
		Revolt: &game.RevoltAction{
			Faction:          game.Alliance,
			ClearingID:       1,
			BaseSuit:         game.Fox,
			SupporterCardIDs: []game.CardID{foxCard.ID, birdCard.ID},
		},
	}

	next := ApplyAction(state, action)
	clearing := next.Map.Clearings[0]

	if len(next.Alliance.Supporters) != 0 {
		t.Fatalf("expected supporters to be spent, got %+v", next.Alliance.Supporters)
	}
	if next.Alliance.Officers != 1 {
		t.Fatalf("expected revolt to grant 1 officer, got %d", next.Alliance.Officers)
	}
	if next.Alliance.WarriorSupply != 2 {
		t.Fatalf("expected revolt to place 2 warriors and 1 officer from supply, got supply %d", next.Alliance.WarriorSupply)
	}
	if clearing.Warriors[game.Alliance] != 2 {
		t.Fatalf("expected 2 alliance warriors after revolt, got %d", clearing.Warriors[game.Alliance])
	}
	if len(clearing.Buildings) != 1 || clearing.Buildings[0].Faction != game.Alliance || clearing.Buildings[0].Type != game.Base {
		t.Fatalf("expected alliance base after revolt, got %+v", clearing.Buildings)
	}
	if clearing.Warriors[game.Marquise] != 0 || clearing.Wood != 0 {
		t.Fatalf("expected enemy pieces to be removed, got warriors=%d wood=%d", clearing.Warriors[game.Marquise], clearing.Wood)
	}
	if next.Marquise.WarriorSupply != 7 {
		t.Fatalf("expected revolted marquise warrior to return to supply, got %d", next.Marquise.WarriorSupply)
	}
	if next.Marquise.KeepClearingID != 0 {
		t.Fatalf("expected keep token to be removed, got keep clearing %d", next.Marquise.KeepClearingID)
	}
	if next.VictoryPoints[game.Alliance] != 4 {
		t.Fatalf("expected revolt to score 4 VP from removed pieces, got %d", next.VictoryPoints[game.Alliance])
	}
	if !next.Alliance.FoxBasePlaced {
		t.Fatalf("expected fox base flag to be set")
	}
	if len(next.PendingFieldHospitals) != 0 {
		t.Fatalf("did not expect Field Hospitals when the keep is removed, got %+v", next.PendingFieldHospitals)
	}
}

func TestApplyRevoltRejectsClearingWithNoOpenSlotAfterRemovingEnemies(t *testing.T) {
	foxCard := firstAllianceTestCard(t, game.Fox)
	birdCard := firstAllianceTestCard(t, game.Bird)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:         1,
					Suit:       game.Fox,
					BuildSlots: 1,
					Ruins:      true,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
			},
		},
		Alliance: game.AllianceState{
			Supporters:    []game.Card{foxCard, birdCard},
			WarriorSupply: 5,
		},
		Marquise: game.MarquiseState{
			WarriorSupply: 5,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionRevolt,
		Revolt: &game.RevoltAction{
			Faction:          game.Alliance,
			ClearingID:       1,
			BaseSuit:         game.Fox,
			SupporterCardIDs: []game.CardID{foxCard.ID, birdCard.ID},
		},
	})

	if len(next.Alliance.Supporters) != 2 {
		t.Fatalf("expected invalid revolt not to spend supporters, got %+v", next.Alliance.Supporters)
	}
	if len(next.DiscardPile) != 0 {
		t.Fatalf("expected invalid revolt not to discard supporters, got %+v", next.DiscardPile)
	}
	if next.Map.Clearings[0].Warriors[game.Marquise] != 1 {
		t.Fatalf("expected invalid revolt not to remove enemy warriors, got %+v", next.Map.Clearings[0].Warriors)
	}
	if len(next.Map.Clearings[0].Buildings) != 0 || next.Alliance.FoxBasePlaced || next.Alliance.Officers != 0 {
		t.Fatalf("expected invalid revolt not to place base or add officer, clearing=%+v alliance=%+v", next.Map.Clearings[0], next.Alliance)
	}
}

func TestApplyRevoltDamagesVagabondInRevoltingClearing(t *testing.T) {
	foxCard := firstAllianceTestCard(t, game.Fox)
	birdCard := firstAllianceTestCard(t, game.Bird)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:         1,
					Suit:       game.Fox,
					BuildSlots: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
			},
		},
		Alliance: game.AllianceState{
			Supporters:    []game.Card{foxCard, birdCard},
			WarriorSupply: 5,
		},
		Marquise: game.MarquiseState{
			WarriorSupply: 5,
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
				{Type: game.ItemSword, Status: game.ItemReady},
				{Type: game.ItemTea, Status: game.ItemReady},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionRevolt,
		Revolt: &game.RevoltAction{
			Faction:                    game.Alliance,
			ClearingID:                 1,
			BaseSuit:                   game.Fox,
			SupporterCardIDs:           []game.CardID{foxCard.ID, birdCard.ID},
			DamagedVagabondItemIndexes: []int{1, 2, 3},
		},
	})

	damaged := 0
	for _, item := range next.Vagabond.Items {
		if item.Status == game.ItemDamaged {
			damaged++
		}
	}
	if damaged != 3 {
		t.Fatalf("expected revolt to damage three Vagabond items, got %+v", next.Vagabond.Items)
	}
	if next.Map.Clearings[0].Warriors[game.Marquise] != 0 {
		t.Fatalf("expected revolt to still remove enemy warriors, got %+v", next.Map.Clearings[0].Warriors)
	}
	if len(next.Map.Clearings[0].Buildings) != 1 || next.Map.Clearings[0].Buildings[0].Type != game.Base {
		t.Fatalf("expected revolt to still place the Alliance base, got %+v", next.Map.Clearings[0].Buildings)
	}
}

func TestApplyRevoltDoesNotDamageCoalitionPartnerVagabond(t *testing.T) {
	foxCard := firstAllianceTestCard(t, game.Fox)
	birdCard := firstAllianceTestCard(t, game.Bird)

	state := game.GameState{
		CoalitionActive:  true,
		CoalitionPartner: game.Alliance,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:         1,
					Suit:       game.Fox,
					BuildSlots: 1,
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
			},
		},
		Alliance: game.AllianceState{
			Supporters:    []game.Card{foxCard, birdCard},
			WarriorSupply: 5,
		},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
				{Type: game.ItemSword, Status: game.ItemReady},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionRevolt,
		Revolt: &game.RevoltAction{
			Faction:          game.Alliance,
			ClearingID:       1,
			BaseSuit:         game.Fox,
			SupporterCardIDs: []game.CardID{foxCard.ID, birdCard.ID},
		},
	})

	for _, item := range next.Vagabond.Items {
		if item.Status == game.ItemDamaged {
			t.Fatalf("did not expect coalition partner Vagabond to be damaged, got %+v", next.Vagabond.Items)
		}
	}
}

func TestApplyBattleResolutionRemovingAllianceBaseAppliesFallout(t *testing.T) {
	foxCard := firstAllianceTestCard(t, game.Fox)
	birdCard := firstAllianceTestCard(t, game.Bird)
	rabbitCards := allianceTestCards(t, game.Rabbit, 6)
	supporters := append([]game.Card{foxCard, birdCard}, rabbitCards...)

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
			Supporters:    supporters,
			Officers:      3,
			FoxBasePlaced: true,
		},
		VictoryPoints: map[game.Faction]int{},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionBattleResolution,
		BattleResolution: &game.BattleResolutionAction{
			Faction:        game.Marquise,
			ClearingID:     1,
			TargetFaction:  game.Alliance,
			DefenderLosses: 1,
		},
	})

	if len(next.Map.Clearings[0].Buildings) != 0 {
		t.Fatalf("expected Alliance base to be removed, got %+v", next.Map.Clearings[0].Buildings)
	}
	if next.Alliance.FoxBasePlaced {
		t.Fatalf("expected fox base flag to clear")
	}
	if next.Alliance.Officers != 1 {
		t.Fatalf("expected base removal to remove half of 3 officers rounded up, got %d", next.Alliance.Officers)
	}
	if len(next.Alliance.Supporters) != 5 {
		t.Fatalf("expected no-base supporter cap to leave five supporters, got %+v", next.Alliance.Supporters)
	}
	for _, supporter := range next.Alliance.Supporters {
		if supporter.Suit == game.Fox || supporter.Suit == game.Bird {
			t.Fatalf("expected fox and bird supporters to be discarded after fox base removal, got %+v", next.Alliance.Supporters)
		}
	}
	if len(next.DiscardPile) != 3 ||
		next.DiscardPile[0] != foxCard.ID ||
		next.DiscardPile[1] != birdCard.ID ||
		next.DiscardPile[2] != rabbitCards[5].ID {
		t.Fatalf("expected matching supporters and no-base excess supporter in discard, got %+v", next.DiscardPile)
	}
	if next.VictoryPoints[game.Marquise] != 1 {
		t.Fatalf("expected attacker to score removed base, got %+v", next.VictoryPoints)
	}
}

func TestApplyAllianceRecruitPlacesOneWarriorAtOneBase(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Buildings: []game.Building{
						{Faction: game.Alliance, Type: game.Base},
					},
				},
				{
					ID: 2,
					Buildings: []game.Building{
						{Faction: game.Alliance, Type: game.Base},
					},
				},
			},
		},
		Alliance: game.AllianceState{
			WarriorSupply: 5,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Alliance,
			ClearingIDs: []int{2},
		},
	})

	if next.Map.Clearings[0].Warriors[game.Alliance] != 0 || next.Map.Clearings[1].Warriors[game.Alliance] != 1 {
		t.Fatalf("expected recruit to place one warrior at chosen base only, got %+v", next.Map.Clearings)
	}
	if next.Alliance.WarriorSupply != 4 {
		t.Fatalf("expected Alliance warrior supply to decrease by one, got %d", next.Alliance.WarriorSupply)
	}
}

func TestApplyAllianceRecruitRejectsMultiBasePayload(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID: 1,
					Buildings: []game.Building{
						{Faction: game.Alliance, Type: game.Base},
					},
				},
				{
					ID: 2,
					Buildings: []game.Building{
						{Faction: game.Alliance, Type: game.Base},
					},
				},
			},
		},
		Alliance: game.AllianceState{
			WarriorSupply: 5,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitAction{
			Faction:     game.Alliance,
			ClearingIDs: []int{1, 2},
		},
	})

	if next.Map.Clearings[0].Warriors != nil || next.Map.Clearings[1].Warriors != nil {
		t.Fatalf("expected multi-base recruit payload to be rejected, got %+v", next.Map.Clearings)
	}
	if next.Alliance.WarriorSupply != 5 {
		t.Fatalf("expected rejected recruit to leave supply unchanged, got %d", next.Alliance.WarriorSupply)
	}
}

func TestApplyOrganizeReturnsAllianceWarriorToSupply(t *testing.T) {
	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Rabbit,
					Warriors: map[game.Faction]int{
						game.Alliance: 1,
					},
				},
			},
		},
		Alliance: game.AllianceState{
			WarriorSupply: 4,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionOrganize,
		Organize: &game.OrganizeAction{
			Faction:    game.Alliance,
			ClearingID: 1,
		},
	})

	if next.Map.Clearings[0].Warriors[game.Alliance] != 0 {
		t.Fatalf("expected organize to remove the alliance warrior, got %+v", next.Map.Clearings[0].Warriors)
	}
	if next.Alliance.WarriorSupply != 5 {
		t.Fatalf("expected organized alliance warrior to return to supply, got %d", next.Alliance.WarriorSupply)
	}
}

func TestApplyMovementIntoSympathyTransfersOutrageCard(t *testing.T) {
	rabbitCard := firstAllianceTestCard(t, game.Rabbit)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Adj:  []int{2},
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Adj:  []int{1},
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
			},
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{rabbitCard},
		},
		Alliance: game.AllianceState{},
	}

	action := game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Marquise,
			Count:    1,
			MaxCount: 1,
			From:     1,
			To:       2,
		},
	}

	next := ApplyAction(state, action)
	if len(next.Marquise.CardsInHand) != 0 {
		t.Fatalf("expected matching card to leave marquise hand, got %+v", next.Marquise.CardsInHand)
	}
	if len(next.Alliance.Supporters) != 1 || next.Alliance.Supporters[0].ID != rabbitCard.ID {
		t.Fatalf("expected outrage to move rabbit card to supporters, got %+v", next.Alliance.Supporters)
	}
}

func TestApplyVagabondMovementIntoSympathyDoesNotTriggerOutrage(t *testing.T) {
	rabbitCard := firstAllianceTestCard(t, game.Rabbit)
	foxCard := firstAllianceTestCard(t, game.Fox)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Adj:  []int{2},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Adj:  []int{1},
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
			},
		},
		Deck: []game.CardID{foxCard.ID},
		Vagabond: game.VagabondState{
			ClearingID: 1,
			CardsInHand: []game.Card{
				rabbitCard,
			},
			Items: []game.Item{
				{Type: game.ItemBoots, Status: game.ItemReady},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Vagabond,
			Count:    1,
			MaxCount: 1,
			From:     1,
			To:       2,
		},
	})

	if next.Vagabond.ClearingID != 2 {
		t.Fatalf("expected Vagabond to move into sympathetic clearing, got %+v", next.Vagabond)
	}
	if len(next.Vagabond.CardsInHand) != 1 || next.Vagabond.CardsInHand[0].ID != rabbitCard.ID {
		t.Fatalf("did not expect Vagabond movement to pay Outrage, got hand %+v", next.Vagabond.CardsInHand)
	}
	if len(next.Alliance.Supporters) != 0 {
		t.Fatalf("did not expect Vagabond pawn movement to add supporters, got %+v", next.Alliance.Supporters)
	}
	if len(next.Deck) != 1 || next.Deck[0] != foxCard.ID {
		t.Fatalf("did not expect Vagabond movement to draw fallback supporter, got deck %+v", next.Deck)
	}
}

func TestApplyMovementIntoSympathyDrawsSupporterWhenNoMatchingCard(t *testing.T) {
	foxCard := firstAllianceTestCard(t, game.Fox)
	rabbitCard := firstAllianceTestCard(t, game.Rabbit)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Adj:  []int{2},
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Adj:  []int{1},
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
			},
		},
		Deck: []game.CardID{rabbitCard.ID},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{foxCard},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Marquise,
			Count:    1,
			MaxCount: 1,
			From:     1,
			To:       2,
		},
	})

	if len(next.Marquise.CardsInHand) != 1 || next.Marquise.CardsInHand[0].ID != foxCard.ID {
		t.Fatalf("expected non-matching hand to remain after Outrage draw fallback, got %+v", next.Marquise.CardsInHand)
	}
	if len(next.Alliance.Supporters) != 1 || next.Alliance.Supporters[0].ID != rabbitCard.ID {
		t.Fatalf("expected Alliance to draw supporter from deck, got %+v", next.Alliance.Supporters)
	}
	if len(next.Deck) != 0 {
		t.Fatalf("expected supporter draw to consume deck card, got %+v", next.Deck)
	}
}

func TestApplyMovementIntoSympathyQueuesOutrageForHiddenAssistHand(t *testing.T) {
	rabbitCard := firstAllianceTestCard(t, game.Rabbit)

	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Alliance,
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Adj:  []int{2},
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Adj:  []int{1},
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
			},
		},
		Deck: []game.CardID{rabbitCard.ID},
		OtherHandCounts: map[game.Faction]int{
			game.Marquise: 1,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Marquise,
			Count:    1,
			MaxCount: 1,
			From:     1,
			To:       2,
		},
	})

	if len(next.PendingOutrage) != 1 || next.PendingOutrage[0].Faction != game.Marquise || next.PendingOutrage[0].Suit != game.Rabbit {
		t.Fatalf("expected pending Marquise rabbit Outrage, got %+v", next.PendingOutrage)
	}
	if len(next.Alliance.Supporters) != 0 {
		t.Fatalf("expected no auto supporter before Outrage is resolved, got %+v", next.Alliance.Supporters)
	}
	if len(next.Deck) != 1 || next.Deck[0] != rabbitCard.ID {
		t.Fatalf("expected no fallback draw before Outrage is resolved, got deck %+v", next.Deck)
	}
	if hiddenCardCount(next, game.Marquise, game.HiddenCardZoneHand) != 1 {
		t.Fatalf("expected hidden Marquise hand placeholder to remain, got %+v", next.HiddenCards)
	}
}

func TestApplyResolveOutrageRevealedHiddenCardBecomesSupporter(t *testing.T) {
	rabbitCard := firstAllianceTestCard(t, game.Rabbit)

	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Alliance,
		PendingOutrage: []game.OutragePending{
			{Faction: game.Marquise, Suit: game.Rabbit},
		},
		OtherHandCounts: map[game.Faction]int{
			game.Marquise: 1,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionResolveOutrage,
		ResolveOutrage: &game.ResolveOutrageAction{
			Faction: game.Marquise,
			Suit:    game.Rabbit,
			CardID:  rabbitCard.ID,
		},
	})

	if len(next.PendingOutrage) != 0 {
		t.Fatalf("expected Outrage pending queue to clear, got %+v", next.PendingOutrage)
	}
	if hiddenCardCount(next, game.Marquise, game.HiddenCardZoneHand) != 0 || next.OtherHandCounts[game.Marquise] != 0 {
		t.Fatalf("expected revealed hidden hand card to be consumed, hidden=%+v counts=%+v", next.HiddenCards, next.OtherHandCounts)
	}
	if len(next.Alliance.Supporters) != 1 || next.Alliance.Supporters[0].ID != rabbitCard.ID {
		t.Fatalf("expected revealed card to become supporter, got %+v", next.Alliance.Supporters)
	}
}

func TestValidActionsReturnsPendingOutrageResolution(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Alliance,
		FactionTurn:   game.Marquise,
		PendingOutrage: []game.OutragePending{
			{Faction: game.Marquise, Suit: game.Rabbit},
		},
		OtherHandCounts: map[game.Faction]int{
			game.Marquise: 1,
		},
	}

	want := game.Action{
		Type: game.ActionResolveOutrage,
		ResolveOutrage: &game.ResolveOutrageAction{
			Faction:       game.Marquise,
			Suit:          game.Rabbit,
			DrawSupporter: true,
		},
	}

	got := ValidActions(state)
	if len(got) != 1 || !containsAction(got, want) {
		t.Fatalf("expected only pending Outrage resolution, got %+v", got)
	}
}

func TestApplyResolveOutrageNoMatchDrawsSupporter(t *testing.T) {
	rabbitCard := firstAllianceTestCard(t, game.Rabbit)

	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Alliance,
		Deck:          []game.CardID{rabbitCard.ID},
		PendingOutrage: []game.OutragePending{
			{Faction: game.Marquise, Suit: game.Rabbit},
		},
		OtherHandCounts: map[game.Faction]int{
			game.Marquise: 1,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionResolveOutrage,
		ResolveOutrage: &game.ResolveOutrageAction{
			Faction:       game.Marquise,
			Suit:          game.Rabbit,
			DrawSupporter: true,
		},
	})

	if len(next.PendingOutrage) != 0 {
		t.Fatalf("expected Outrage pending queue to clear, got %+v", next.PendingOutrage)
	}
	if hiddenCardCount(next, game.Marquise, game.HiddenCardZoneHand) != 1 {
		t.Fatalf("expected no-match reveal to leave hidden hand count unchanged, got %+v", next.HiddenCards)
	}
	if len(next.Alliance.Supporters) != 1 || next.Alliance.Supporters[0].ID != rabbitCard.ID {
		t.Fatalf("expected fallback draw to add supporter, got %+v", next.Alliance.Supporters)
	}
	if len(next.Deck) != 0 {
		t.Fatalf("expected fallback draw to consume deck card, got %+v", next.Deck)
	}
}

func TestApplyMovementIntoSympathyDiscardsSupporterWhenStackIsCapped(t *testing.T) {
	rabbitCard := firstAllianceTestCard(t, game.Rabbit)
	foxCard := firstAllianceTestCard(t, game.Fox)

	state := game.GameState{
		Map: game.Map{
			Clearings: []game.Clearing{
				{
					ID:   1,
					Suit: game.Fox,
					Adj:  []int{2},
					Warriors: map[game.Faction]int{
						game.Marquise: 1,
					},
				},
				{
					ID:   2,
					Suit: game.Rabbit,
					Adj:  []int{1},
					Tokens: []game.Token{
						{Faction: game.Alliance, Type: game.TokenSympathy},
					},
				},
			},
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{rabbitCard},
		},
		Alliance: game.AllianceState{
			Supporters: []game.Card{foxCard, foxCard, foxCard, foxCard, foxCard},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionMovement,
		Movement: &game.MovementAction{
			Faction:  game.Marquise,
			Count:    1,
			MaxCount: 1,
			From:     1,
			To:       2,
		},
	})

	if len(next.Marquise.CardsInHand) != 0 {
		t.Fatalf("expected matching Outrage card to leave Marquise hand, got %+v", next.Marquise.CardsInHand)
	}
	if len(next.Alliance.Supporters) != 5 {
		t.Fatalf("expected capped supporter stack to stay at 5, got %+v", next.Alliance.Supporters)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != rabbitCard.ID {
		t.Fatalf("expected capped supporter gain to discard the card, got %+v", next.DiscardPile)
	}
}

func TestAllianceEveningDrawStaysInEveningForDiscard(t *testing.T) {
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
		FactionTurn:  game.Alliance,
		CurrentPhase: game.Evening,
		CurrentStep:  game.StepEvening,
		TurnOrder:    []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		Alliance: game.AllianceState{
			Officers:      1,
			WarriorSupply: 1,
		},
	}

	draw := game.Action{
		Type: game.ActionEveningDraw,
		EveningDraw: &game.EveningDrawAction{
			Faction: game.Alliance,
			Count:   2,
		},
	}

	next := ApplyAction(state, draw)
	if next.FactionTurn != game.Alliance {
		t.Fatalf("expected alliance draw to keep Alliance active for discard, got %v", next.FactionTurn)
	}
	if next.CurrentPhase != game.Evening || next.CurrentStep != game.StepEvening {
		t.Fatalf("expected alliance draw to remain in evening for discard, got phase=%v step=%v", next.CurrentPhase, next.CurrentStep)
	}
	if !next.TurnProgress.EveningDrawn {
		t.Fatalf("expected alliance evening draw to be marked before discard")
	}
}
