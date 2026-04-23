package engine

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestApplyActionCraftObservedHiddenCardInAssistMode(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Marquise,
		OtherHandCounts: map[game.Faction]int{
			game.Eyrie: 2,
		},
		ItemSupply: map[game.ItemType]int{
			game.ItemCoin: 1,
		},
		VictoryPoints: map[game.Faction]int{},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction: game.Eyrie,
			CardID:  21,
		},
	})

	if next.OtherHandCounts[game.Eyrie] != 1 {
		t.Fatalf("expected hidden craft to spend one hand count, got %+v", next.OtherHandCounts)
	}
	if next.VictoryPoints[game.Eyrie] != 3 {
		t.Fatalf("expected observed craft to score printed VP, got %+v", next.VictoryPoints)
	}
	if next.ItemSupply[game.ItemCoin] != 0 {
		t.Fatalf("expected observed craft to deduct item supply, got %+v", next.ItemSupply)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 21 {
		t.Fatalf("expected crafted card to enter discard, got %+v", next.DiscardPile)
	}
}

func TestApplyActionAddToDecreeObservedHiddenCardDecrementsHandCount(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Marquise,
		OtherHandCounts: map[game.Faction]int{
			game.Eyrie: 3,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionAddToDecree,
		AddToDecree: &game.AddToDecreeAction{
			Faction: game.Eyrie,
			CardIDs: []game.CardID{24},
			Columns: []game.DecreeColumn{game.DecreeMove},
		},
	})

	if next.OtherHandCounts[game.Eyrie] != 2 {
		t.Fatalf("expected hidden decree add to spend one hand count, got %+v", next.OtherHandCounts)
	}
	if len(next.Eyrie.Decree.Move) != 1 || next.Eyrie.Decree.Move[0] != 24 {
		t.Fatalf("expected public decree card to be recorded, got %+v", next.Eyrie.Decree)
	}
}

func TestApplyActionTrainObservedHiddenCardInAssistMode(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Marquise,
		OtherHandCounts: map[game.Faction]int{
			game.Alliance: 2,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionTrain,
		Train: &game.TrainAction{
			Faction: game.Alliance,
			CardID:  24,
		},
	})

	if next.OtherHandCounts[game.Alliance] != 1 {
		t.Fatalf("expected hidden train to spend one hand count, got %+v", next.OtherHandCounts)
	}
	if next.Alliance.Officers != 1 {
		t.Fatalf("expected train to grant an officer, got %+v", next.Alliance)
	}
	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 24 {
		t.Fatalf("expected trained card to enter discard, got %+v", next.DiscardPile)
	}
}

func TestApplyActionMobilizeMovesHiddenHandPlaceholderToSupporters(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Marquise,
		HiddenCards: []game.HiddenCard{
			{ID: 1, OwnerFaction: game.Alliance, Zone: game.HiddenCardZoneHand},
		},
		NextHiddenCardID: 2,
		OtherHandCounts: map[game.Faction]int{
			game.Alliance: 1,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionMobilize,
		Mobilize: &game.MobilizeAction{
			Faction: game.Alliance,
			CardID:  24,
		},
	})

	if hiddenCardCount(next, game.Alliance, game.HiddenCardZoneHand) != 0 {
		t.Fatalf("expected mobilize to move hidden hand placeholder out of hand, got %+v", next.HiddenCards)
	}
	if hiddenCardCount(next, game.Alliance, game.HiddenCardZoneSupporters) != 1 {
		t.Fatalf("expected mobilize to move hidden placeholder into supporters, got %+v", next.HiddenCards)
	}
}

func TestApplyActionSpreadSympathyObservedSupportersInAssistMode(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Marquise,
		Map: game.Map{
			Clearings: []game.Clearing{
				{ID: 4, Suit: game.Mouse},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionSpreadSympathy,
		SpreadSympathy: &game.SpreadSympathyAction{
			Faction:          game.Alliance,
			ClearingID:       4,
			SupporterCardIDs: []game.CardID{24, 48},
		},
	})

	if len(next.Map.Clearings[0].Tokens) != 1 || next.Map.Clearings[0].Tokens[0].Faction != game.Alliance {
		t.Fatalf("expected observed spread sympathy to place sympathy, got %+v", next.Map.Clearings[0].Tokens)
	}
	if len(next.DiscardPile) != 2 {
		t.Fatalf("expected spent supporters to be discarded, got %+v", next.DiscardPile)
	}
}

func TestApplyActionAidObservedHiddenVagabondCardInAssistMode(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Marquise,
		OtherHandCounts: map[game.Faction]int{
			game.Vagabond: 2,
		},
		Map: game.Map{
			Clearings: []game.Clearing{
				{ID: 1, Suit: game.Rabbit},
			},
		},
		Marquise: game.MarquiseState{},
		Vagabond: game.VagabondState{
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
			},
			Relationships: map[game.Faction]game.RelationshipLevel{
				game.Marquise: game.RelIndifferent,
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionAid,
		Aid: &game.AidAction{
			Faction:       game.Vagabond,
			TargetFaction: game.Marquise,
			ClearingID:    1,
			CardID:        24,
			ItemIndex:     0,
		},
	})

	if next.OtherHandCounts[game.Vagabond] != 1 {
		t.Fatalf("expected observed aid to spend one hidden Vagabond card, got %+v", next.OtherHandCounts)
	}
	if len(next.Marquise.CardsInHand) != 1 || next.Marquise.CardsInHand[0].ID != 24 {
		t.Fatalf("expected observed aid to transfer the public card, got %+v", next.Marquise.CardsInHand)
	}
	if next.Vagabond.Relationships[game.Marquise] != game.RelFriendly {
		t.Fatalf("expected aid to improve relationship, got %+v", next.Vagabond.Relationships)
	}
}
