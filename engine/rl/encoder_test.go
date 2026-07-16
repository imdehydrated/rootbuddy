package rl

import (
	"reflect"
	"testing"

	rootengine "github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

func TestEncodeObservationConstantLengthAndCloneStable(t *testing.T) {
	state := testTrainingState(t)

	obs := rootengine.NewTrainingObservation(state, rootengine.TrainingObservationOptions{
		Perspective: game.Marquise,
	})
	encoded := EncodeObservation(obs)
	if len(encoded) == 0 {
		t.Fatal("expected non-empty observation encoding")
	}
	if got, want := len(encoded), ObservationVectorLength(); got != want {
		t.Fatalf("observation length = %d, want %d", got, want)
	}

	cloned := rootengine.CloneState(state)
	cloneObs := rootengine.NewTrainingObservation(cloned, rootengine.TrainingObservationOptions{
		Perspective: game.Marquise,
	})
	cloneEncoded := EncodeObservation(cloneObs)
	if !reflect.DeepEqual(encoded, cloneEncoded) {
		t.Fatalf("expected clone-equivalent observations to encode identically")
	}
}

func TestEncodeObservationLengthStableAcrossPerspectives(t *testing.T) {
	state := testTrainingState(t)
	want := ObservationVectorLength()

	for _, faction := range []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond} {
		obs := rootengine.NewTrainingObservation(state, rootengine.TrainingObservationOptions{
			Perspective: faction,
		})
		if got := len(EncodeObservation(obs)); got != want {
			t.Fatalf("observation length for %v = %d, want %d", faction, got, want)
		}
	}
}

func TestEncodeActionConstantLengthForRepresentativeFamilies(t *testing.T) {
	state := game.GameState{
		RoundNumber: 1,
		VictoryPoints: map[game.Faction]int{
			game.Marquise: 3,
			game.Eyrie:    2,
			game.Alliance: 1,
			game.Vagabond: 4,
		},
	}
	actions := []game.Action{
		{
			Type: game.ActionMovement,
			Movement: &game.MovementAction{
				Faction:  game.Marquise,
				Count:    2,
				MaxCount: 4,
				From:     1,
				To:       5,
			},
		},
		{
			Type: game.ActionBattle,
			Battle: &game.BattleAction{
				Faction:       game.Eyrie,
				ClearingID:    3,
				TargetFaction: game.Marquise,
			},
		},
		{
			Type: game.ActionBattleResolution,
			BattleResolution: &game.BattleResolutionAction{
				Faction:        game.Marquise,
				ClearingID:     1,
				TargetFaction:  game.Eyrie,
				AttackerRoll:   3,
				DefenderRoll:   1,
				DefenderLosses: 2,
				DefenderPieceLosses: []game.BattlePieceLoss{
					{Kind: game.BattlePieceBuilding, BuildingType: game.Roost},
				},
			},
		},
		{
			Type: game.ActionBuild,
			Build: &game.BuildAction{
				Faction:      game.Marquise,
				ClearingID:   1,
				BuildingType: game.Sawmill,
				WoodSources:  []game.WoodSource{{ClearingID: 1, Amount: 1}},
			},
		},
		{
			Type: game.ActionRecruit,
			Recruit: &game.RecruitAction{
				Faction:     game.Alliance,
				ClearingIDs: []int{2, 6},
			},
		},
		{
			Type: game.ActionAddToDecree,
			AddToDecree: &game.AddToDecreeAction{
				Faction: game.Eyrie,
				CardIDs: []game.CardID{24},
				Columns: []game.DecreeColumn{game.DecreeMove},
			},
		},
		{
			Type: game.ActionCraft,
			Craft: &game.CraftAction{
				Faction:               game.Marquise,
				CardID:                52,
				UsedWorkshopClearings: []int{1},
			},
		},
		{
			Type: game.ActionSpreadSympathy,
			SpreadSympathy: &game.SpreadSympathyAction{
				Faction:          game.Alliance,
				ClearingID:       8,
				SupporterCardIDs: []game.CardID{53},
			},
		},
		{
			Type: game.ActionQuest,
			Quest: &game.QuestAction{
				Faction:     game.Vagabond,
				QuestID:     4,
				ItemIndexes: []int{0, 1},
				Reward:      game.QuestRewardVictoryPoints,
			},
		},
		{
			Type: game.ActionAid,
			Aid: &game.AidAction{
				Faction:       game.Vagabond,
				TargetFaction: game.Marquise,
				ClearingID:    1,
				CardID:        52,
				ItemIndex:     0,
			},
		},
		{
			Type:      game.ActionPassPhase,
			PassPhase: &game.PassPhaseAction{Faction: game.Eyrie},
		},
	}

	want := ActionVectorLength()
	for _, action := range actions {
		encoded := EncodeAction(state, action)
		if got := len(encoded); got != want {
			t.Fatalf("%v action length = %d, want %d", action.Type, got, want)
		}
		if !hasNonZero(encoded) {
			t.Fatalf("%v action encoding is all zeros", action.Type)
		}
	}
}

func TestEncodeActionCloneStable(t *testing.T) {
	state := testTrainingState(t)
	action := game.Action{
		Type: game.ActionBuild,
		Build: &game.BuildAction{
			Faction:      game.Marquise,
			ClearingID:   1,
			BuildingType: game.Workshop,
			WoodSources:  []game.WoodSource{{ClearingID: 1, Amount: 2}},
			DecreeCardID: 24,
		},
	}

	encoded := EncodeAction(state, action)
	cloneEncoded := EncodeAction(rootengine.CloneState(state), action)
	if !reflect.DeepEqual(encoded, cloneEncoded) {
		t.Fatalf("expected clone-equivalent action encodings to match")
	}
}

func TestEncodeGeneratedLegalActionsHaveConstantLength(t *testing.T) {
	state := testTrainingState(t)
	actions := rootengine.ValidActions(state)
	if len(actions) == 0 {
		t.Fatal("expected training setup to produce legal actions")
	}

	want := ActionVectorLength()
	for index, action := range actions {
		if got := len(EncodeAction(state, action)); got != want {
			t.Fatalf("generated action %d length = %d, want %d", index, got, want)
		}
	}
}

func testTrainingState(t *testing.T) game.GameState {
	t.Helper()

	state, err := rootengine.SetupTrainingGame(rootengine.SetupRequest{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
		Factions:      []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		MapID:         game.AutumnMapID,
		RandomSeed:    707,
	})
	if err != nil {
		t.Fatalf("setup training game: %v", err)
	}
	if err := rootengine.ValidateState(state); err != nil {
		t.Fatalf("setup training game produced invalid state: %v", err)
	}
	return state
}

func hasNonZero(values []float32) bool {
	for _, value := range values {
		if value != 0 {
			return true
		}
	}
	return false
}
