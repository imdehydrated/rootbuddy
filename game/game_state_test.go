package game

import "testing"

func TestGameStateCreation(t *testing.T) {
	sampleClearing := Clearing{
		ID:         1,
		Suit:       Fox,
		BuildSlots: 2,
		Adj:        []int{2, 3},
		Ruins:      true,
	}
	sampleMap := Map{
		Clearings: []Clearing{sampleClearing},
	}
	sampleCats := MarquiseState{
		CardsInHand:      []int{1, 2, 3},
		WarriorSupply:    25,
		SawmillsPlaced:   1,
		WorkshopsPlaced:  1,
		RecruitersPlaced: 1,
	}
	sampleGame := GameState{
		Map:          sampleMap,
		FactionTurn:  Marquise,
		CurrentPhase: Birdsong,
		Marquise:     sampleCats,
	}
	if len(sampleGame.Map.Clearings) != 1 {
		t.Fatalf("Expected 1 clearing, got %v", len(sampleGame.Map.Clearings))
	}
	if sampleGame.Marquise.WarriorSupply != 25 {
		t.Fatalf("Expected 25 warriors, got %v", sampleGame.Marquise.WarriorSupply)
	}
	if sampleGame.CurrentPhase != Birdsong {
		t.Fatalf("Expected phase to be Birdsong, got %v", sampleGame.CurrentPhase)
	}
}
