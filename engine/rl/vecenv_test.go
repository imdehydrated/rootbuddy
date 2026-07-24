package rl

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"

	rootengine "github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

func TestNewVecEnvAppliesDefaults(t *testing.T) {
	env, err := NewVecEnv(VecEnvConfig{
		NumEnvs:  3,
		BaseSeed: 707,
	})
	if err != nil {
		t.Fatalf("NewVecEnv failed: %v", err)
	}

	if env.Len() != 3 {
		t.Fatalf("Len = %d, want 3", env.Len())
	}
	config := env.Config()
	if config.MaxSteps != DefaultVecEnvMaxSteps {
		t.Fatalf("MaxSteps = %d, want %d", config.MaxSteps, DefaultVecEnvMaxSteps)
	}
	if config.MapID != game.AutumnMapID {
		t.Fatalf("MapID = %q, want %q", config.MapID, game.AutumnMapID)
	}
	if config.TerminalWinBonus != DefaultTerminalWinBonus {
		t.Fatalf("TerminalWinBonus = %f, want %f", config.TerminalWinBonus, DefaultTerminalWinBonus)
	}
	if got, want := config.Factions, []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond}; !sameFactions(got, want) {
		t.Fatalf("Factions = %+v, want %+v", got, want)
	}
	if env.slots[0].seed == env.slots[1].seed || env.slots[1].seed == env.slots[2].seed {
		t.Fatalf("expected distinct slot seeds, got %+v", []int64{env.slots[0].seed, env.slots[1].seed, env.slots[2].seed})
	}
}

func TestNewVecEnvRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config VecEnvConfig
	}{
		{
			name: "no envs",
			config: VecEnvConfig{
				BaseSeed: 1,
			},
		},
		{
			name: "zero seed",
			config: VecEnvConfig{
				NumEnvs: 1,
			},
		},
		{
			name: "negative max steps",
			config: VecEnvConfig{
				NumEnvs:  1,
				BaseSeed: 1,
				MaxSteps: -1,
			},
		},
		{
			name: "duplicate factions",
			config: VecEnvConfig{
				NumEnvs:  1,
				BaseSeed: 1,
				Factions: []game.Faction{game.Marquise, game.Marquise},
			},
		},
		{
			name: "too few factions",
			config: VecEnvConfig{
				NumEnvs:  1,
				BaseSeed: 1,
				Factions: []game.Faction{game.Marquise},
			},
		},
		{
			name: "unsupported map",
			config: VecEnvConfig{
				NumEnvs:  1,
				BaseSeed: 1,
				MapID:    game.MapID("winter"),
			},
		},
		{
			name: "explicit invalid player faction",
			config: VecEnvConfig{
				NumEnvs:       1,
				BaseSeed:      1,
				Factions:      []game.Faction{game.Eyrie, game.Alliance},
				PlayerFaction: game.Vagabond,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewVecEnv(tt.config)
			if !errors.Is(err, ErrInvalidVecEnvConfig) {
				t.Fatalf("NewVecEnv error = %v, want ErrInvalidVecEnvConfig", err)
			}
		})
	}
}

func TestNewVecEnvCopiesConfigSlices(t *testing.T) {
	factions := []game.Faction{game.Marquise, game.Eyrie}
	env, err := NewVecEnv(VecEnvConfig{
		NumEnvs:  1,
		BaseSeed: 10,
		Factions: factions,
	})
	if err != nil {
		t.Fatalf("NewVecEnv failed: %v", err)
	}

	factions[0] = game.Vagabond
	config := env.Config()
	config.Factions[0] = game.Vagabond

	fresh := env.Config()
	if got, want := fresh.Factions, []game.Faction{game.Marquise, game.Eyrie}; !sameFactions(got, want) {
		t.Fatalf("stored factions mutated, got %+v want %+v", got, want)
	}
}

func TestSeedForSlotIsDeterministicAndDistinct(t *testing.T) {
	first := seedForSlot(707, 2, 3)
	second := seedForSlot(707, 2, 3)
	if first != second {
		t.Fatalf("seedForSlot not deterministic: %d != %d", first, second)
	}
	if first == seedForSlot(707, 2, 4) || first == seedForSlot(707, 3, 3) {
		t.Fatalf("seedForSlot should vary by slot and episode")
	}
}

func TestVecEnvResetInitializesSlotsAndDecisions(t *testing.T) {
	env, err := NewVecEnv(VecEnvConfig{
		NumEnvs:       2,
		BaseSeed:      707,
		TrackAllHands: true,
	})
	if err != nil {
		t.Fatalf("NewVecEnv failed: %v", err)
	}

	result, err := env.Reset()
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
	if len(result.Decisions) != 2 {
		t.Fatalf("decisions length = %d, want 2", len(result.Decisions))
	}

	for index, decision := range result.Decisions {
		if decision.EnvIndex != index {
			t.Fatalf("decision EnvIndex = %d, want %d", decision.EnvIndex, index)
		}
		if decision.Episode != 0 || decision.Step != 0 {
			t.Fatalf("decision episode/step = %d/%d, want 0/0", decision.Episode, decision.Step)
		}
		if decision.Done {
			t.Fatalf("reset decision unexpectedly done: %+v", decision)
		}
		if decision.CandidateCount == 0 || len(decision.CandidateActions) != decision.CandidateCount {
			t.Fatalf("candidate shape mismatch: count=%d vectors=%d", decision.CandidateCount, len(decision.CandidateActions))
		}
		if len(decision.LegalActionTypes) != decision.CandidateCount {
			t.Fatalf("legal action type length = %d, want %d", len(decision.LegalActionTypes), decision.CandidateCount)
		}
		if len(decision.Observation) != ObservationVectorLength() || decision.ObservationLength != ObservationVectorLength() {
			t.Fatalf("observation length = %d/%d, want %d", len(decision.Observation), decision.ObservationLength, ObservationVectorLength())
		}
		for actionIndex, encoded := range decision.CandidateActions {
			if len(encoded) != ActionVectorLength() {
				t.Fatalf("candidate %d length = %d, want %d", actionIndex, len(encoded), ActionVectorLength())
			}
		}

		slot := env.slots[index]
		if !slot.initialized {
			t.Fatalf("slot %d was not marked initialized", index)
		}
		if slot.seed != seedForSlot(707, index, 0) {
			t.Fatalf("slot %d seed = %d, want %d", index, slot.seed, seedForSlot(707, index, 0))
		}
		if len(slot.legalActions) != decision.CandidateCount {
			t.Fatalf("slot %d legal actions = %d, decision count = %d", index, len(slot.legalActions), decision.CandidateCount)
		}
		for _, faction := range []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond} {
			if got := slot.lastVictoryPoints[faction]; got != 0 {
				t.Fatalf("slot %d last VP for %v = %d, want 0", index, faction, got)
			}
		}
	}
}

func TestVecEnvResetIsDeterministicForFreshEnv(t *testing.T) {
	config := VecEnvConfig{
		NumEnvs:  1,
		BaseSeed: 909,
	}
	left, err := NewVecEnv(config)
	if err != nil {
		t.Fatalf("NewVecEnv left failed: %v", err)
	}
	right, err := NewVecEnv(config)
	if err != nil {
		t.Fatalf("NewVecEnv right failed: %v", err)
	}

	leftResult, err := left.Reset()
	if err != nil {
		t.Fatalf("left Reset failed: %v", err)
	}
	rightResult, err := right.Reset()
	if err != nil {
		t.Fatalf("right Reset failed: %v", err)
	}

	if !reflect.DeepEqual(leftResult.Decisions, rightResult.Decisions) {
		t.Fatalf("fresh env resets with same seed diverged")
	}
}

func TestVecEnvResetEnvAdvancesEpisodeSeed(t *testing.T) {
	env, err := NewVecEnv(VecEnvConfig{
		NumEnvs:  1,
		BaseSeed: 707,
	})
	if err != nil {
		t.Fatalf("NewVecEnv failed: %v", err)
	}

	first, err := env.ResetEnv(0)
	if err != nil {
		t.Fatalf("first ResetEnv failed: %v", err)
	}
	firstSeed := env.slots[0].seed

	second, err := env.ResetEnv(0)
	if err != nil {
		t.Fatalf("second ResetEnv failed: %v", err)
	}
	if first.Episode != 0 || second.Episode != 1 {
		t.Fatalf("episodes = %d/%d, want 0/1", first.Episode, second.Episode)
	}
	if env.slots[0].seed == firstSeed {
		t.Fatalf("expected second reset to use a new seed")
	}
	if env.slots[0].seed != seedForSlot(707, 0, 1) {
		t.Fatalf("second seed = %d, want %d", env.slots[0].seed, seedForSlot(707, 0, 1))
	}
}

func TestVecEnvResetEnvRejectsInvalidIndex(t *testing.T) {
	env, err := NewVecEnv(VecEnvConfig{
		NumEnvs:  1,
		BaseSeed: 707,
	})
	if err != nil {
		t.Fatalf("NewVecEnv failed: %v", err)
	}

	_, err = env.ResetEnv(1)
	if !errors.Is(err, ErrInvalidEnvIndex) {
		t.Fatalf("ResetEnv error = %v, want ErrInvalidEnvIndex", err)
	}
}

func TestVecEnvStepEnvAppliesNonBattleAction(t *testing.T) {
	env, err := NewVecEnv(VecEnvConfig{
		NumEnvs:  1,
		BaseSeed: 707,
	})
	if err != nil {
		t.Fatalf("NewVecEnv failed: %v", err)
	}
	reset, err := env.ResetEnv(0)
	if err != nil {
		t.Fatalf("ResetEnv failed: %v", err)
	}

	actionIndex := firstNonBattleActionIndex(t, reset)
	decision, err := env.StepEnv(0, actionIndex)
	if err != nil {
		t.Fatalf("StepEnv failed: %v", err)
	}
	if decision.Step != 1 {
		t.Fatalf("Step = %d, want 1", decision.Step)
	}
	if decision.Episode != 0 {
		t.Fatalf("Episode = %d, want 0", decision.Episode)
	}
	if decision.CandidateCount == 0 {
		t.Fatalf("expected next decision candidates, got 0")
	}
	if env.slots[0].steps != 1 {
		t.Fatalf("slot steps = %d, want 1", env.slots[0].steps)
	}
	if len(env.slots[0].legalActions) != decision.CandidateCount {
		t.Fatalf("slot legal actions = %d, decision candidates = %d", len(env.slots[0].legalActions), decision.CandidateCount)
	}
}

func TestVecEnvStepAppliesBatch(t *testing.T) {
	env, err := NewVecEnv(VecEnvConfig{
		NumEnvs:  2,
		BaseSeed: 707,
	})
	if err != nil {
		t.Fatalf("NewVecEnv failed: %v", err)
	}
	reset, err := env.Reset()
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	result, err := env.Step(StepRequest{
		ActionIndices: []int{
			firstNonBattleActionIndex(t, reset.Decisions[0]),
			firstNonBattleActionIndex(t, reset.Decisions[1]),
		},
	})
	if err != nil {
		t.Fatalf("Step failed: %v", err)
	}
	if len(result.Decisions) != 2 {
		t.Fatalf("decisions length = %d, want 2", len(result.Decisions))
	}
	for index, decision := range result.Decisions {
		if decision.EnvIndex != index || decision.Step != 1 {
			t.Fatalf("decision %d env/step = %d/%d, want %d/1", index, decision.EnvIndex, decision.Step, index)
		}
	}
}

func TestVecEnvStepRejectsActionCountMismatch(t *testing.T) {
	env, err := NewVecEnv(VecEnvConfig{
		NumEnvs:  2,
		BaseSeed: 707,
	})
	if err != nil {
		t.Fatalf("NewVecEnv failed: %v", err)
	}

	_, err = env.Step(StepRequest{ActionIndices: []int{0}})
	if !errors.Is(err, ErrActionCountMismatch) {
		t.Fatalf("Step error = %v, want ErrActionCountMismatch", err)
	}
}

func TestVecEnvStepEnvRejectsUninitializedDoneAndInvalidAction(t *testing.T) {
	env, err := NewVecEnv(VecEnvConfig{
		NumEnvs:  1,
		BaseSeed: 707,
	})
	if err != nil {
		t.Fatalf("NewVecEnv failed: %v", err)
	}

	_, err = env.StepEnv(0, 0)
	if !errors.Is(err, ErrEnvNotInitialized) {
		t.Fatalf("uninitialized StepEnv error = %v, want ErrEnvNotInitialized", err)
	}

	if _, err := env.ResetEnv(0); err != nil {
		t.Fatalf("ResetEnv failed: %v", err)
	}
	_, err = env.StepEnv(0, len(env.slots[0].legalActions))
	if !errors.Is(err, ErrInvalidActionIndex) {
		t.Fatalf("invalid index StepEnv error = %v, want ErrInvalidActionIndex", err)
	}

	env.slots[0].done = true
	_, err = env.StepEnv(0, 0)
	if !errors.Is(err, ErrEnvDone) {
		t.Fatalf("done StepEnv error = %v, want ErrEnvDone", err)
	}
}

func TestVecEnvStepEnvAppliesBattleAction(t *testing.T) {
	env, err := NewVecEnv(VecEnvConfig{
		NumEnvs:          1,
		BaseSeed:         707,
		TerminalWinBonus: 30,
	})
	if err != nil {
		t.Fatalf("NewVecEnv failed: %v", err)
	}

	state := testBattleState()
	action := game.Action{
		Type: game.ActionBattle,
		Battle: &game.BattleAction{
			Faction:       game.Marquise,
			ClearingID:    1,
			TargetFaction: game.Eyrie,
		},
	}
	env.slots[0] = envSlot{
		state:             state,
		seed:              707,
		initialized:       true,
		lastVictoryPoints: victoryPointSnapshot(state),
		legalActions:      []game.Action{action},
	}

	decision, err := env.StepEnv(0, 0)
	if err != nil {
		t.Fatalf("StepEnv battle failed: %v", err)
	}
	if decision.Step != 1 {
		t.Fatalf("Step = %d, want 1", decision.Step)
	}
	if env.slots[0].state.BattleRollCount != 1 {
		t.Fatalf("BattleRollCount = %d, want 1", env.slots[0].state.BattleRollCount)
	}
	if env.slots[0].state.Eyrie.RoostsPlaced != 0 {
		t.Fatalf("Eyrie roosts placed = %d, want 0", env.slots[0].state.Eyrie.RoostsPlaced)
	}
	if got := env.slots[0].state.VictoryPoints[game.Marquise]; got != 1 {
		t.Fatalf("Marquise victory points = %d, want 1", got)
	}
	if decision.Reward != 1 {
		t.Fatalf("Reward = %f, want 1", decision.Reward)
	}
}

func TestVecEnvStepEnvIsDeterministicForSameActionIndex(t *testing.T) {
	left, err := NewVecEnv(VecEnvConfig{
		NumEnvs:  1,
		BaseSeed: 808,
	})
	if err != nil {
		t.Fatalf("NewVecEnv left failed: %v", err)
	}
	right, err := NewVecEnv(VecEnvConfig{
		NumEnvs:  1,
		BaseSeed: 808,
	})
	if err != nil {
		t.Fatalf("NewVecEnv right failed: %v", err)
	}

	leftReset, err := left.ResetEnv(0)
	if err != nil {
		t.Fatalf("left reset failed: %v", err)
	}
	rightReset, err := right.ResetEnv(0)
	if err != nil {
		t.Fatalf("right reset failed: %v", err)
	}
	actionIndex := firstNonBattleActionIndex(t, leftReset)
	if actionIndex != firstNonBattleActionIndex(t, rightReset) {
		t.Fatalf("fresh envs chose different first non-battle action index")
	}

	leftDecision, err := left.StepEnv(0, actionIndex)
	if err != nil {
		t.Fatalf("left StepEnv failed: %v", err)
	}
	rightDecision, err := right.StepEnv(0, actionIndex)
	if err != nil {
		t.Fatalf("right StepEnv failed: %v", err)
	}
	if !reflect.DeepEqual(leftDecision, rightDecision) {
		t.Fatalf("same seed/action StepEnv decisions diverged")
	}
}

func TestVecEnvRandomPolicyRolloutMaintainsValidState(t *testing.T) {
	env, err := NewVecEnv(VecEnvConfig{
		NumEnvs:       2,
		BaseSeed:      1777,
		MaxSteps:      5000,
		Factions:      []game.Faction{game.Marquise, game.Eyrie},
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
	})
	if err != nil {
		t.Fatalf("NewVecEnv failed: %v", err)
	}
	if _, err := env.Reset(); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	rng := rand.New(rand.NewSource(1777))
	completedGames := 0
	maxedEpisodes := 0
	observedBattle := false
	countedDone := make([]bool, env.Len())
	const targetCompletedGames = 2
	const maxTransitions = 12_000

	for transition := 0; transition < maxTransitions && completedGames < targetCompletedGames; transition++ {
		envIndex := transition % env.Len()
		slot := &env.slots[envIndex]
		if slot.done {
			if !countedDone[envIndex] {
				if slot.state.GamePhase == game.LifecycleGameOver {
					completedGames++
				} else {
					maxedEpisodes++
				}
				countedDone[envIndex] = true
			}
			if completedGames >= targetCompletedGames {
				break
			}
			if _, err := env.ResetEnv(envIndex); err != nil {
				t.Fatalf("ResetEnv(%d) after done failed: %v", envIndex, err)
			}
			countedDone[envIndex] = false
			slot = &env.slots[envIndex]
		}
		if len(slot.legalActions) == 0 {
			t.Fatalf("env %d has no legal actions before done at transition %d", envIndex, transition)
		}

		actionIndex := rng.Intn(len(slot.legalActions))
		if slot.legalActions[actionIndex].Type == game.ActionBattle {
			observedBattle = true
		}
		decision, err := env.StepEnv(envIndex, actionIndex)
		if err != nil {
			t.Fatalf("StepEnv(%d, %d) transition %d failed: %v", envIndex, actionIndex, transition, err)
		}
		if err := rootengine.ValidateState(env.slots[envIndex].state); err != nil {
			t.Fatalf("env %d transition %d produced invalid state: %v", envIndex, transition, err)
		}
		assertDecisionShape(t, decision)
		if decision.Done && !countedDone[envIndex] {
			if env.slots[envIndex].state.GamePhase == game.LifecycleGameOver {
				completedGames++
			} else {
				maxedEpisodes++
			}
			countedDone[envIndex] = true
		}
	}
	if completedGames < targetCompletedGames {
		t.Fatalf("completed games = %d, want at least %d; maxed episodes=%d", completedGames, targetCompletedGames, maxedEpisodes)
	}
	if !observedBattle {
		t.Fatalf("random rollout did not exercise battle adapter")
	}
}

func TestRewardForTransitionUsesActingFactionVPAndWinBonus(t *testing.T) {
	previous := map[game.Faction]int{
		game.Marquise: 2,
		game.Eyrie:    5,
	}
	next := game.GameState{
		GamePhase: game.LifecyclePlaying,
		VictoryPoints: map[game.Faction]int{
			game.Marquise: 4,
			game.Eyrie:    10,
		},
	}
	if got := rewardForTransition(previous, next, game.Marquise, 30); got != 2 {
		t.Fatalf("reward = %f, want 2", got)
	}

	next.GamePhase = game.LifecycleGameOver
	next.Winner = game.Marquise
	if got := rewardForTransition(previous, next, game.Marquise, 30); got != 32 {
		t.Fatalf("terminal winner reward = %f, want 32", got)
	}
}

func assertDecisionShape(t *testing.T, decision EnvDecision) {
	t.Helper()

	if len(decision.Observation) != ObservationVectorLength() {
		t.Fatalf("observation length = %d, want %d", len(decision.Observation), ObservationVectorLength())
	}
	if decision.ObservationLength != ObservationVectorLength() {
		t.Fatalf("observation length field = %d, want %d", decision.ObservationLength, ObservationVectorLength())
	}
	if decision.ActionLength != ActionVectorLength() {
		t.Fatalf("action length field = %d, want %d", decision.ActionLength, ActionVectorLength())
	}
	if decision.CandidateCount != len(decision.CandidateActions) {
		t.Fatalf("candidate count = %d, vectors = %d", decision.CandidateCount, len(decision.CandidateActions))
	}
	if len(decision.LegalActionTypes) != decision.CandidateCount {
		t.Fatalf("legal action types = %d, candidates = %d", len(decision.LegalActionTypes), decision.CandidateCount)
	}
	for index, encoded := range decision.CandidateActions {
		if len(encoded) != ActionVectorLength() {
			t.Fatalf("candidate %d length = %d, want %d", index, len(encoded), ActionVectorLength())
		}
	}
}

func firstNonBattleActionIndex(t *testing.T, decision EnvDecision) int {
	t.Helper()

	for index, actionType := range decision.LegalActionTypes {
		if actionType != game.ActionBattle {
			return index
		}
	}
	t.Fatalf("decision has no non-battle actions: %+v", decision.LegalActionTypes)
	return -1
}

func testBattleState() game.GameState {
	return game.GameState{
		Map: game.Map{
			ID: game.AutumnMapID,
			Clearings: []game.Clearing{
				{
					ID:         1,
					Suit:       game.Fox,
					BuildSlots: 1,
					Warriors: map[game.Faction]int{
						game.Marquise: 3,
					},
					Buildings: []game.Building{
						{Faction: game.Eyrie, Type: game.Roost},
					},
				},
			},
		},
		GameMode:          game.GameModeOnline,
		RandomSeed:        707,
		GamePhase:         game.LifecyclePlaying,
		SetupStage:        game.SetupStageComplete,
		PlayerFaction:     game.Marquise,
		RoundNumber:       1,
		FactionTurn:       game.Marquise,
		CurrentPhase:      game.Daylight,
		CurrentStep:       game.StepDaylightActions,
		TurnOrder:         []game.Faction{game.Marquise, game.Eyrie},
		VictoryPoints:     map[game.Faction]int{game.Marquise: 0, game.Eyrie: 0},
		ActiveDominance:   map[game.Faction]game.CardID{},
		ItemSupply:        rootengine.InitialItemSupply(),
		CraftedItems:      map[game.Faction][]game.ItemType{},
		PersistentEffects: map[game.Faction][]game.CardID{},
		OtherHandCounts:   map[game.Faction]int{},
		Marquise: game.MarquiseState{
			WarriorSupply: 22,
			WoodSupply:    8,
		},
		Eyrie: game.EyrieState{
			WarriorSupply: 20,
			RoostsPlaced:  1,
		},
		Alliance: game.AllianceState{
			WarriorSupply: 10,
		},
	}
}

func sameFactions(left []game.Faction, right []game.Faction) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
