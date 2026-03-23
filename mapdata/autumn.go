package mapdata

import "github.com/imdehydrated/rootbuddy/game"

func AutumnMap() game.Map {
	return game.Map{
		ID: game.AutumnMapID,
		Clearings: []game.Clearing{
			{
				ID:         1,
				Suit:       game.Fox,
				BuildSlots: 1,
				Adj:        []int{5, 10, 9},
				Ruins:      false,
			},
			{
				ID:         2,
				Suit:       game.Mouse,
				BuildSlots: 2,
				Adj:        []int{5, 10, 6},
				Ruins:      false,
			},
			{
				ID:         3,
				Suit:       game.Rabbit,
				BuildSlots: 1,
				Adj:        []int{7, 11, 6},
				Ruins:      false,
			},
			{
				ID:         4,
				Suit:       game.Rabbit,
				BuildSlots: 1,
				Adj:        []int{9, 12, 8},
				Ruins:      false,
			},
			{
				ID:         5,
				Suit:       game.Rabbit,
				BuildSlots: 2,
				Adj:        []int{1, 2},
				Ruins:      false,
			},
			{
				ID:         6,
				Suit:       game.Fox,
				BuildSlots: 2,
				Adj:        []int{2, 11, 3},
				Ruins:      true,
			},
			{
				ID:         7,
				Suit:       game.Mouse,
				BuildSlots: 2,
				Adj:        []int{3, 12, 8},
				Ruins:      false,
			},
			{
				ID:         8,
				Suit:       game.Fox,
				BuildSlots: 2,
				Adj:        []int{7, 4},
				Ruins:      false,
			},
			{
				ID:         9,
				Suit:       game.Mouse,
				BuildSlots: 2,
				Adj:        []int{1, 12, 4},
				Ruins:      false,
			},
			{
				ID:         10,
				Suit:       game.Rabbit,
				BuildSlots: 2,
				Adj:        []int{1, 2, 12},
				Ruins:      true,
			},
			{
				ID:         11,
				Suit:       game.Mouse,
				BuildSlots: 3,
				Adj:        []int{6, 3, 12},
				Ruins:      true,
			},
			{
				ID:         12,
				Suit:       game.Fox,
				BuildSlots: 2,
				Adj:        []int{4, 9, 10, 11, 7},
				Ruins:      true,
			},
		},
	}
}
