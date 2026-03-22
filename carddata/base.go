package carddata

import "github.com/imdehydrated/rootbuddy/game"

func BaseDeck() []game.Card {
	item := func(id game.CardID, name string, suit game.Suit, cost game.CraftingCost, vp int) game.Card {
		return game.Card{
			ID:           id,
			Deck:         game.BaseDeck,
			Name:         name,
			Suit:         suit,
			Kind:         game.ItemCard,
			CraftingCost: cost,
			VP:           vp,
		}
	}

	persistent := func(id game.CardID, name string, suit game.Suit, cost game.CraftingCost) game.Card {
		return game.Card{
			ID:           id,
			Deck:         game.BaseDeck,
			Name:         name,
			Suit:         suit,
			Kind:         game.PersistentEffectCard,
			CraftingCost: cost,
			VP:           0,
		}
	}

	oneTime := func(id game.CardID, name string, suit game.Suit, cost game.CraftingCost) game.Card {
		return game.Card{
			ID:           id,
			Deck:         game.BaseDeck,
			Name:         name,
			Suit:         suit,
			Kind:         game.OneTimeEffectCard,
			CraftingCost: cost,
			VP:           0,
		}
	}

	ambush := func(id game.CardID, suit game.Suit) game.Card {
		return game.Card{
			ID:           id,
			Deck:         game.BaseDeck,
			Name:         "Ambush",
			Suit:         suit,
			Kind:         game.AmbushCard,
			CraftingCost: game.CraftingCost{},
			VP:           0,
		}
	}

	dominance := func(id game.CardID, suit game.Suit) game.Card {
		return game.Card{
			ID:           id,
			Deck:         game.BaseDeck,
			Name:         "Dominance",
			Suit:         suit,
			Kind:         game.DominanceCard,
			CraftingCost: game.CraftingCost{},
			VP:           0,
		}
	}

	return []game.Card{
		// Bird suit (14)
		persistent(1, "Armorers", game.Bird, game.CraftingCost{Fox: 1}),
		persistent(2, "Armorers", game.Bird, game.CraftingCost{Fox: 1}),
		persistent(3, "Sappers", game.Bird, game.CraftingCost{Mouse: 1}),
		persistent(4, "Sappers", game.Bird, game.CraftingCost{Mouse: 1}),
		persistent(5, "Brutal Tactics", game.Bird, game.CraftingCost{Fox: 2}),
		persistent(6, "Brutal Tactics", game.Bird, game.CraftingCost{Fox: 2}),
		oneTime(7, "Royal Claim", game.Bird, game.CraftingCost{Any: 4}),
		item(8, "Birdy Bindle", game.Bird, game.CraftingCost{Mouse: 1}, 1),
		item(9, "Woodland Runners", game.Bird, game.CraftingCost{Rabbit: 1}, 1),
		item(10, "Arms Trader", game.Bird, game.CraftingCost{Fox: 2}, 2),
		item(11, "Crossbow", game.Bird, game.CraftingCost{Fox: 1}, 1),
		ambush(12, game.Bird),
		ambush(13, game.Bird),
		dominance(14, game.Bird),

		// Rabbit suit (13)
		persistent(15, "Better Burrow Bank", game.Rabbit, game.CraftingCost{Rabbit: 2}),
		persistent(16, "Better Burrow Bank", game.Rabbit, game.CraftingCost{Rabbit: 2}),
		persistent(17, "Cobbler", game.Rabbit, game.CraftingCost{Rabbit: 2}),
		persistent(18, "Cobbler", game.Rabbit, game.CraftingCost{Rabbit: 2}),
		persistent(19, "Command Warren", game.Rabbit, game.CraftingCost{Rabbit: 2}),
		persistent(20, "Command Warren", game.Rabbit, game.CraftingCost{Rabbit: 2}),
		item(21, "Bake Sale", game.Rabbit, game.CraftingCost{Rabbit: 2}, 3),
		item(22, "Smuggler's Trail", game.Rabbit, game.CraftingCost{Mouse: 1}, 1),
		item(23, "Root Tea", game.Rabbit, game.CraftingCost{Mouse: 1}, 2),
		item(24, "A Visit to Friends", game.Rabbit, game.CraftingCost{Rabbit: 1}, 1),
		oneTime(25, "Favor of the Rabbits", game.Rabbit, game.CraftingCost{Rabbit: 3}),
		ambush(26, game.Rabbit),
		dominance(27, game.Rabbit),

		// Mouse suit (13)
		persistent(28, "Codebreakers", game.Mouse, game.CraftingCost{Mouse: 1}),
		persistent(29, "Codebreakers", game.Mouse, game.CraftingCost{Mouse: 1}),
		persistent(30, "Scouting Party", game.Mouse, game.CraftingCost{Mouse: 2}),
		persistent(31, "Scouting Party", game.Mouse, game.CraftingCost{Mouse: 2}),
		item(32, "Crossbow", game.Mouse, game.CraftingCost{Fox: 1}, 1),
		item(33, "Sword", game.Mouse, game.CraftingCost{Fox: 2}, 2),
		item(34, "Travel Gear", game.Mouse, game.CraftingCost{Rabbit: 1}, 1),
		item(35, "Investments", game.Mouse, game.CraftingCost{Rabbit: 2}, 3),
		oneTime(36, "Favor of the Mice", game.Mouse, game.CraftingCost{Mouse: 3}),
		item(37, "Root Tea", game.Mouse, game.CraftingCost{Mouse: 1}, 2),
		item(38, "Mouse-in-a-Sack", game.Mouse, game.CraftingCost{Mouse: 1}, 1),
		ambush(39, game.Mouse),
		dominance(40, game.Mouse),

		// Fox suit (14)
		persistent(41, "Stand and Deliver!", game.Fox, game.CraftingCost{Mouse: 3}),
		persistent(42, "Stand and Deliver!", game.Fox, game.CraftingCost{Mouse: 3}),
		persistent(43, "Tax Collector", game.Fox, game.CraftingCost{Fox: 1, Mouse: 1, Rabbit: 1}),
		persistent(44, "Tax Collector", game.Fox, game.CraftingCost{Fox: 1, Mouse: 1, Rabbit: 1}),
		persistent(45, "Tax Collector", game.Fox, game.CraftingCost{Fox: 1, Mouse: 1, Rabbit: 1}),
		item(46, "Root Tea", game.Fox, game.CraftingCost{Mouse: 1}, 2),
		item(47, "Protection Racket", game.Fox, game.CraftingCost{Rabbit: 2}, 3),
		item(48, "Travel Gear", game.Fox, game.CraftingCost{Rabbit: 1}, 1),
		item(49, "Gently Used Knapsack", game.Fox, game.CraftingCost{Mouse: 1}, 1),
		oneTime(50, "Favor of the Foxes", game.Fox, game.CraftingCost{Fox: 3}),
		item(51, "Foxfolk Steel", game.Fox, game.CraftingCost{Fox: 2}, 2),
		item(52, "Anvil", game.Fox, game.CraftingCost{Fox: 1}, 2),
		ambush(53, game.Fox),
		dominance(54, game.Fox),
	}
}
