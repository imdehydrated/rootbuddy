package engine

import (
	"reflect"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestDrawCardsAddsCardsToPlayerHandInOnlineMode(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		Deck:          []game.CardID{8, 9},
	}

	drawn := DrawCards(&state, game.Marquise, 2)

	if !reflect.DeepEqual(state.Deck, []game.CardID{}) {
		t.Fatalf("expected deck to be emptied, got %+v", state.Deck)
	}
	if len(drawn) != 2 || drawn[0].ID != 8 || drawn[1].ID != 9 {
		t.Fatalf("expected cards 8 and 9 to be drawn, got %+v", drawn)
	}
	if len(state.Marquise.CardsInHand) != 2 || state.Marquise.CardsInHand[0].ID != 8 || state.Marquise.CardsInHand[1].ID != 9 {
		t.Fatalf("expected cards to be added to Marquise hand, got %+v", state.Marquise.CardsInHand)
	}
}

func TestDrawCardsTracksOtherFactionCountOnlyInOnlineMode(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		Deck:          []game.CardID{8, 9},
	}

	drawn := DrawCards(&state, game.Eyrie, 2)

	if len(drawn) != 0 {
		t.Fatalf("expected other faction draw to return no revealed cards, got %+v", drawn)
	}
	if len(state.Eyrie.CardsInHand) != 0 {
		t.Fatalf("expected other faction hand to stay untracked, got %+v", state.Eyrie.CardsInHand)
	}
	if state.OtherHandCounts[game.Eyrie] != 2 {
		t.Fatalf("expected other hand count to increase to 2, got %+v", state.OtherHandCounts)
	}
	if len(state.Deck) != 0 {
		t.Fatalf("expected deck to lose the hidden draws, got %+v", state.Deck)
	}
}

func TestApplyActionCraftDiscardsCraftedCard(t *testing.T) {
	state := game.GameState{
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{
				{ID: 20, Name: "Crafted Card"},
			},
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionCraft,
		Craft: &game.CraftAction{
			Faction: game.Marquise,
			CardID:  20,
		},
	})

	if len(next.DiscardPile) != 1 || next.DiscardPile[0] != 20 {
		t.Fatalf("expected crafted card to move to discard, got %+v", next.DiscardPile)
	}
}

func TestApplyActionTurmoilDiscardsDecreeCardsButNotViziers(t *testing.T) {
	state := game.GameState{
		Eyrie: game.EyrieState{
			Leader: game.LeaderBuilder,
			Decree: game.Decree{
				Recruit: []game.CardID{8, game.LoyalVizier1},
				Move:    []game.CardID{9, game.LoyalVizier2},
			},
			AvailableLeaders: []game.EyrieLeader{
				game.LeaderCharismatic,
				game.LeaderCommander,
				game.LeaderDespot,
			},
		},
		VictoryPoints: map[game.Faction]int{
			game.Eyrie: 5,
		},
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionTurmoil,
		Turmoil: &game.TurmoilAction{
			Faction:   game.Eyrie,
			NewLeader: game.LeaderCommander,
		},
	})

	if !reflect.DeepEqual(next.DiscardPile, []game.CardID{8, 9}) {
		t.Fatalf("expected only non-vizier decree cards to be discarded, got %+v", next.DiscardPile)
	}
}

func TestKnownCardIDsCountsOnlyPlayerVisibleCards(t *testing.T) {
	state := game.GameState{
		PlayerFaction: game.Eyrie,
		DiscardPile:   []game.CardID{25},
		PersistentEffects: map[game.Faction][]game.CardID{
			game.Marquise: {15},
		},
		Marquise: game.MarquiseState{
			CardsInHand: []game.Card{{ID: 8}},
		},
		Eyrie: game.EyrieState{
			CardsInHand: []game.Card{{ID: 30}},
			Decree: game.Decree{
				Recruit: []game.CardID{game.LoyalVizier1, 32},
			},
		},
	}

	got := KnownCardIDs(state)
	want := []game.CardID{15, 25, 30, 32}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected known cards %+v, got %+v", want, got)
	}
	if UnknownCardCount(state) != 50 {
		t.Fatalf("expected 50 unknown cards, got %d", UnknownCardCount(state))
	}
}

func TestApplyActionOtherPlayerDrawAndPlayTrackCountsAndDiscard(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeAssist,
		PlayerFaction: game.Marquise,
	}

	next := ApplyAction(state, game.Action{
		Type: game.ActionOtherPlayerDraw,
		OtherPlayerDraw: &game.OtherPlayerDrawAction{
			Faction: game.Eyrie,
			Count:   2,
		},
	})
	next = ApplyAction(next, game.Action{
		Type: game.ActionOtherPlayerPlay,
		OtherPlayerPlay: &game.OtherPlayerPlayAction{
			Faction: game.Eyrie,
			CardID:  25,
		},
	})

	if next.OtherHandCounts[game.Eyrie] != 1 {
		t.Fatalf("expected other hand count to net to 1, got %+v", next.OtherHandCounts)
	}
	if !reflect.DeepEqual(next.DiscardPile, []game.CardID{25}) {
		t.Fatalf("expected played card to be discarded, got %+v", next.DiscardPile)
	}
}

func TestApplyActionQuestDrawRewardDrawsCardsInOnlineMode(t *testing.T) {
	state := game.GameState{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Vagabond,
		Deck:          []game.CardID{8, 9},
		Vagabond: game.VagabondState{
			Items: []game.Item{
				{Type: game.ItemTorch, Status: game.ItemReady},
				{Type: game.ItemBoots, Status: game.ItemReady},
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

	next := ApplyAction(state, game.Action{
		Type: game.ActionQuest,
		Quest: &game.QuestAction{
			Faction:     game.Vagabond,
			QuestID:     9,
			ItemIndexes: []int{0, 1},
			Reward:      game.QuestRewardDrawCards,
		},
	})

	if len(next.Vagabond.CardsInHand) != 2 {
		t.Fatalf("expected draw-reward quest to add 2 cards to Vagabond hand, got %+v", next.Vagabond.CardsInHand)
	}
	if len(next.Deck) != 0 {
		t.Fatalf("expected quest draw reward to consume the deck cards, got %+v", next.Deck)
	}
}
