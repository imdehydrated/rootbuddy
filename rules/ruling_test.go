package rules

import (
	"testing"

	"github.com/imdehydrated/rootbuddy/game"
)

func TestRuler(t *testing.T) {
	tests := []struct {
		name         string
		clearing     game.Clearing
		wantFaction  game.Faction
		wantHasRuler bool
	}{
		{
			name: "marquise rules with more warriors",
			clearing: game.Clearing{
				Warriors: map[game.Faction]int{
					game.Marquise:         4,
					game.Eyrie:            1,
					game.WoodlandAlliance: 1,
				},
				Buildings: []game.Building{},
			},
			wantFaction:  game.Marquise,
			wantHasRuler: true,
		},
		{
			name: "buildings count toward rule",
			clearing: game.Clearing{
				Warriors: map[game.Faction]int{
					game.Marquise:         2,
					game.WoodlandAlliance: 2,
				},
				Buildings: []game.Building{
					{Faction: game.Marquise, Type: game.Sawmill},
				},
			},
			wantFaction:  game.Marquise,
			wantHasRuler: true,
		},
		{
			name: "tie means nobody rules",
			clearing: game.Clearing{
				Warriors: map[game.Faction]int{
					game.Marquise:         1,
					game.WoodlandAlliance: 2,
				},
				Buildings: []game.Building{
					{Faction: game.Marquise, Type: game.Sawmill},
				},
			},
			wantHasRuler: false,
		},
		{
			name: "eyrie rules tied clearing with lords of the forest",
			clearing: game.Clearing{
				Warriors: map[game.Faction]int{
					game.Marquise: 2,
					game.Eyrie:    2,
				},
			},
			wantFaction:  game.Eyrie,
			wantHasRuler: true,
		},
		{
			name: "eyrie buildings count for lords of the forest tie",
			clearing: game.Clearing{
				Warriors: map[game.Faction]int{
					game.Marquise: 1,
				},
				Buildings: []game.Building{
					{Faction: game.Eyrie, Type: game.Roost},
				},
			},
			wantFaction:  game.Eyrie,
			wantHasRuler: true,
		},
		{
			name: "single faction present rules",
			clearing: game.Clearing{
				Warriors: map[game.Faction]int{
					game.Eyrie: 1,
				},
				Buildings: []game.Building{},
			},
			wantFaction:  game.Eyrie,
			wantHasRuler: true,
		},
		{
			name: "empty clearing has no ruler",
			clearing: game.Clearing{
				Warriors:  map[game.Faction]int{},
				Buildings: []game.Building{},
			},
			wantHasRuler: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFaction, gotHasRuler := Ruler(tt.clearing)
			if gotHasRuler != tt.wantHasRuler {
				t.Fatalf("got hasRuler %v, want %v", gotHasRuler, tt.wantHasRuler)
			}
			if gotHasRuler && gotFaction != tt.wantFaction {
				t.Fatalf("got ruler %v, want %v", gotFaction, tt.wantFaction)
			}
		})
	}
}
