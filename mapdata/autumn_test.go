package mapdata

import (
	"slices"
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func findClearingByID(t *testing.T, m game.Map, id int) game.Clearing {
	for _, c := range m.Clearings {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("Clearing with id %v doesn't exist", id)
	return game.Clearing{}
}

func TestAutumnClearingAdjacency(t *testing.T) {
	sampleAutumn := AutumnMap()
	if len(sampleAutumn.Clearings) != 12 {
		t.Fatalf("Map has %v clearings. Expected: 12", len(sampleAutumn.Clearings))
	}
	for _, clearing := range sampleAutumn.Clearings {
		for _, adjid := range clearing.Adj {
			neighbor := findClearingByID(t, sampleAutumn, adjid)
			if !slices.Contains(neighbor.Adj, clearing.ID) {
				t.Fatalf("clearing %d lists %d as adjacent, but clearing %d does not list %d", clearing.ID, adjid, adjid, clearing.ID)
			}
		}
	}
}

func TestAutumnSuits(t *testing.T) {
	sampleAutumn := AutumnMap()
	suitCounts := map[game.Suit]int{game.Fox: 0, game.Rabbit: 0, game.Mouse: 0}
	for _, c := range sampleAutumn.Clearings {
		suitCounts[c.Suit] += 1
	}
	for suit, num := range suitCounts {
		if num != 4 {
			t.Fatalf("Expected 4 %v clearings, found %v", suit, num)
		}
	}
}
