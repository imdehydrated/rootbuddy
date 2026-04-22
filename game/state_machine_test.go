package game

import "testing"

func TestTurnWindowDefaultsDaylightToFactionEntryStep(t *testing.T) {
	tests := []struct {
		name     string
		faction  Faction
		wantStep TurnStep
	}{
		{name: "marquise", faction: Marquise, wantStep: StepDaylightCraft},
		{name: "eyrie", faction: Eyrie, wantStep: StepDaylightCraft},
		{name: "alliance", faction: Alliance, wantStep: StepDaylightCraft},
		{name: "vagabond", faction: Vagabond, wantStep: StepDaylightActions},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window := GameState{
				GamePhase:    LifecyclePlaying,
				FactionTurn:  tt.faction,
				CurrentPhase: Daylight,
			}.TurnWindow()

			if window.Step != tt.wantStep {
				t.Fatalf("expected daylight entry step %v, got %v", tt.wantStep, window.Step)
			}
		})
	}
}
