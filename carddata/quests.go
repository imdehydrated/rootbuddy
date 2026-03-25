package carddata

import "github.com/imdehydrated/rootbuddy/game"

func QuestDeck() []game.Quest {
	return []game.Quest{
		{ID: 1, Name: "Fundraising", Suit: game.Fox, RequiredItems: []game.ItemType{game.ItemCoin, game.ItemTea}},
		{ID: 2, Name: "Errand", Suit: game.Fox, RequiredItems: []game.ItemType{game.ItemBag, game.ItemBoots}},
		{ID: 3, Name: "Logistics Help", Suit: game.Fox, RequiredItems: []game.ItemType{game.ItemBag, game.ItemBoots}},
		{ID: 4, Name: "Repair a Shed", Suit: game.Fox, RequiredItems: []game.ItemType{game.ItemHammer, game.ItemTorch}},
		{ID: 5, Name: "Give a Speech", Suit: game.Fox, RequiredItems: []game.ItemType{game.ItemTea, game.ItemCoin}},
		{ID: 6, Name: "Guard Duty", Suit: game.Rabbit, RequiredItems: []game.ItemType{game.ItemSword, game.ItemTorch}},
		{ID: 7, Name: "Errand", Suit: game.Rabbit, RequiredItems: []game.ItemType{game.ItemBag, game.ItemBoots}},
		{ID: 8, Name: "Give a Speech", Suit: game.Rabbit, RequiredItems: []game.ItemType{game.ItemTea, game.ItemCoin}},
		{ID: 9, Name: "Fend off a Bear", Suit: game.Rabbit, RequiredItems: []game.ItemType{game.ItemCrossbow, game.ItemSword}},
		{ID: 10, Name: "Expel Bandits", Suit: game.Rabbit, RequiredItems: []game.ItemType{game.ItemSword, game.ItemBoots}},
		{ID: 11, Name: "Guard Duty", Suit: game.Mouse, RequiredItems: []game.ItemType{game.ItemSword, game.ItemTorch}},
		{ID: 12, Name: "Escort", Suit: game.Mouse, RequiredItems: []game.ItemType{game.ItemBoots, game.ItemTea}},
		{ID: 13, Name: "Logistics Help", Suit: game.Mouse, RequiredItems: []game.ItemType{game.ItemBag, game.ItemBoots}},
		{ID: 14, Name: "Fend off a Bear", Suit: game.Mouse, RequiredItems: []game.ItemType{game.ItemCrossbow, game.ItemSword}},
		{ID: 15, Name: "Expel Bandits", Suit: game.Mouse, RequiredItems: []game.ItemType{game.ItemSword, game.ItemBoots}},
	}
}
