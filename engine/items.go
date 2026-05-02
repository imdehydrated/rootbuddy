package engine

import "github.com/imdehydrated/rootbuddy/game"

func InitialItemSupply() map[game.ItemType]int {
	return map[game.ItemType]int{
		game.ItemTea:      2,
		game.ItemCoin:     2,
		game.ItemCrossbow: 1,
		game.ItemHammer:   1,
		game.ItemSword:    2,
		game.ItemBoots:    2,
		game.ItemBag:      2,
	}
}

func RuinItems() []game.ItemType {
	return []game.ItemType{
		game.ItemSword,
		game.ItemHammer,
		game.ItemBag,
		game.ItemCoin,
	}
}

func VagabondStartingItems(character game.VagabondCharacter) []game.Item {
	switch character {
	case game.CharThief:
		return []game.Item{
			game.NormalizeItemZone(game.Item{Type: game.ItemBoots, Status: game.ItemReady}),
			game.NormalizeItemZone(game.Item{Type: game.ItemTorch, Status: game.ItemReady}),
			game.NormalizeItemZone(game.Item{Type: game.ItemTea, Status: game.ItemReady}),
			game.NormalizeItemZone(game.Item{Type: game.ItemSword, Status: game.ItemReady}),
		}
	case game.CharTinker:
		return []game.Item{
			game.NormalizeItemZone(game.Item{Type: game.ItemBoots, Status: game.ItemReady}),
			game.NormalizeItemZone(game.Item{Type: game.ItemTorch, Status: game.ItemReady}),
			game.NormalizeItemZone(game.Item{Type: game.ItemBag, Status: game.ItemReady}),
			game.NormalizeItemZone(game.Item{Type: game.ItemHammer, Status: game.ItemReady}),
		}
	case game.CharRanger:
		return []game.Item{
			game.NormalizeItemZone(game.Item{Type: game.ItemBoots, Status: game.ItemReady}),
			game.NormalizeItemZone(game.Item{Type: game.ItemTorch, Status: game.ItemReady}),
			game.NormalizeItemZone(game.Item{Type: game.ItemCrossbow, Status: game.ItemReady}),
			game.NormalizeItemZone(game.Item{Type: game.ItemSword, Status: game.ItemReady}),
		}
	default:
		return nil
	}
}

func ensureItemSupply(state *game.GameState) {
	if state.ItemSupply == nil {
		state.ItemSupply = InitialItemSupply()
	}
}

func DeductItem(state *game.GameState, item game.ItemType) bool {
	ensureItemSupply(state)
	if state.ItemSupply[item] <= 0 {
		return false
	}

	state.ItemSupply[item]--
	return true
}

func ReturnItem(state *game.GameState, item game.ItemType) {
	ensureItemSupply(state)
	state.ItemSupply[item]++
}
