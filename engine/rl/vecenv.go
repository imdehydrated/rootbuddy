package rl

import (
	"errors"
	"fmt"

	rootengine "github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

const (
	DefaultVecEnvMaxSteps                      = 5000
	DefaultTerminalWinBonus            float32 = 30
	DefaultStepPenalty                 float32 = 0.01
	DefaultTruncationPenalty           float32 = 10
	DefaultEyrieTurmoilPenalty         float32 = 1
	DefaultEyrieTurmoilVPLossPenalty   float32 = 0.5
	DefaultEyrieScoreRoostsBonus       float32 = 0.25
	DefaultEyrieRoostBuildBonus        float32 = 0.5
	DefaultEyrieRoostLossPenalty       float32 = 0.75
	DefaultEyrieBuildDecreeCardPenalty float32 = 0.05
)

var (
	ErrInvalidVecEnvConfig = errors.New("invalid vec env config")
	ErrInvalidEnvIndex     = errors.New("invalid env index")
	ErrEnvNotInitialized   = errors.New("env is not initialized")
	ErrEnvDone             = errors.New("env is done")
	ErrInvalidActionIndex  = errors.New("invalid action index")
	ErrActionCountMismatch = errors.New("action index count does not match env count")
)

type VecEnvConfig struct {
	NumEnvs                     int            `json:"numEnvs"`
	BaseSeed                    int64          `json:"baseSeed"`
	MaxSteps                    int            `json:"maxSteps"`
	Factions                    []game.Faction `json:"factions,omitempty"`
	PlayerFaction               game.Faction   `json:"playerFaction"`
	MapID                       game.MapID     `json:"mapId"`
	TrackAllHands               bool           `json:"trackAllHands"`
	TerminalWinBonus            float32        `json:"terminalWinBonus"`
	StepPenalty                 float32        `json:"stepPenalty"`
	TruncationPenalty           float32        `json:"truncationPenalty"`
	DisableEyrieRewardShaping   bool           `json:"disableEyrieRewardShaping"`
	EyrieTurmoilPenalty         float32        `json:"eyrieTurmoilPenalty"`
	EyrieTurmoilVPLossPenalty   float32        `json:"eyrieTurmoilVpLossPenalty"`
	EyrieScoreRoostsBonus       float32        `json:"eyrieScoreRoostsBonus"`
	EyrieRoostBuildBonus        float32        `json:"eyrieRoostBuildBonus"`
	EyrieRoostLossPenalty       float32        `json:"eyrieRoostLossPenalty"`
	EyrieBuildDecreeCardPenalty float32        `json:"eyrieBuildDecreeCardPenalty"`
}

type VecEnv struct {
	config VecEnvConfig
	slots  []envSlot
}

type envSlot struct {
	state             game.GameState
	seed              int64
	episode           int
	steps             int
	done              bool
	truncated         bool
	initialized       bool
	lastVictoryPoints map[game.Faction]int
	legalActions      []game.Action
}

type EnvDecision struct {
	EnvIndex          int               `json:"envIndex"`
	Episode           int               `json:"episode"`
	Step              int               `json:"step"`
	RoundNumber       int               `json:"roundNumber"`
	ActiveFaction     game.Faction      `json:"activeFaction"`
	CurrentPhase      game.Phase        `json:"currentPhase"`
	CurrentStep       game.TurnStep     `json:"currentStep"`
	Eyrie             EyrieDiagnostics  `json:"eyrie"`
	Observation       []float32         `json:"observation"`
	CandidateActions  [][]float32       `json:"candidateActions"`
	CandidateRewards  []float32         `json:"candidateRewards"`
	CandidateCount    int               `json:"candidateCount"`
	Reward            float32           `json:"reward"`
	Done              bool              `json:"done"`
	Truncated         bool              `json:"truncated"`
	Winner            game.Faction      `json:"winner"`
	WinningCoalition  []game.Faction    `json:"winningCoalition,omitempty"`
	VictoryPoints     []int             `json:"victoryPoints"`
	LegalActionTypes  []game.ActionType `json:"legalActionTypes,omitempty"`
	ObservationLength int               `json:"observationLength"`
	ActionLength      int               `json:"actionLength"`
}

type EyrieDiagnostics struct {
	VictoryPoints         int   `json:"victoryPoints"`
	RoostsPlaced          int   `json:"roostsPlaced"`
	WarriorSupply         int   `json:"warriorSupply"`
	DecreeColumnCounts    []int `json:"decreeColumnCounts"`
	CurrentDecreeColumn   int   `json:"currentDecreeColumn"`
	DecreeColumnsResolved int   `json:"decreeColumnsResolved"`
	DecreeCardsResolved   int   `json:"decreeCardsResolved"`
	CardsAddedToDecree    int   `json:"cardsAddedToDecree"`
}

type ResetResult struct {
	Decisions []EnvDecision `json:"decisions"`
}

type StepRequest struct {
	ActionIndices []int `json:"actionIndices"`
}

type StepResult struct {
	Decisions []EnvDecision `json:"decisions"`
}

func NewVecEnv(config VecEnvConfig) (*VecEnv, error) {
	normalized, err := normalizeVecEnvConfig(config)
	if err != nil {
		return nil, err
	}

	env := &VecEnv{
		config: normalized,
		slots:  make([]envSlot, normalized.NumEnvs),
	}
	for index := range env.slots {
		env.slots[index].seed = seedForSlot(normalized.BaseSeed, index, 0)
		env.slots[index].lastVictoryPoints = map[game.Faction]int{}
	}
	return env, nil
}

func (env *VecEnv) Config() VecEnvConfig {
	config := env.config
	config.Factions = append([]game.Faction(nil), env.config.Factions...)
	return config
}

func (env *VecEnv) Len() int {
	if env == nil {
		return 0
	}
	return len(env.slots)
}

func (env *VecEnv) Reset() (ResetResult, error) {
	decisions := make([]EnvDecision, len(env.slots))
	for index := range env.slots {
		decision, err := env.ResetEnv(index)
		if err != nil {
			return ResetResult{}, err
		}
		decisions[index] = decision
	}
	return ResetResult{Decisions: decisions}, nil
}

func (env *VecEnv) ResetEnv(index int) (EnvDecision, error) {
	if env == nil || index < 0 || index >= len(env.slots) {
		return EnvDecision{}, fmt.Errorf("%w: %d", ErrInvalidEnvIndex, index)
	}

	slot := &env.slots[index]
	if slot.initialized {
		slot.episode++
	}
	slot.seed = seedForSlot(env.config.BaseSeed, index, slot.episode)

	state, err := rootengine.SetupTrainingGame(rootengine.SetupRequest{
		GameMode:      game.GameModeOnline,
		PlayerFaction: env.config.PlayerFaction,
		TrackAllHands: env.config.TrackAllHands,
		Factions:      env.config.Factions,
		MapID:         env.config.MapID,
		RandomSeed:    slot.seed,
	})
	if err != nil {
		return EnvDecision{}, fmt.Errorf("reset env %d: %w", index, err)
	}
	if err := rootengine.ValidateState(state); err != nil {
		return EnvDecision{}, fmt.Errorf("reset env %d produced invalid state: %w", index, err)
	}

	slot.state = state
	slot.steps = 0
	slot.done = state.GamePhase == game.LifecycleGameOver
	slot.truncated = false
	slot.initialized = true
	slot.lastVictoryPoints = victoryPointSnapshot(state)
	slot.legalActions = rootengine.ValidActions(state)
	if !slot.done && len(slot.legalActions) == 0 {
		slot.done = true
		slot.truncated = true
	}

	return env.decision(index, 0)
}

func (env *VecEnv) Step(request StepRequest) (StepResult, error) {
	if env == nil {
		return StepResult{}, fmt.Errorf("%w: nil env", ErrInvalidEnvIndex)
	}
	if len(request.ActionIndices) != len(env.slots) {
		return StepResult{}, fmt.Errorf("%w: got %d want %d", ErrActionCountMismatch, len(request.ActionIndices), len(env.slots))
	}

	decisions := make([]EnvDecision, len(env.slots))
	for index, actionIndex := range request.ActionIndices {
		decision, err := env.StepEnv(index, actionIndex)
		if err != nil {
			return StepResult{}, err
		}
		decisions[index] = decision
	}
	return StepResult{Decisions: decisions}, nil
}

func (env *VecEnv) StepEnv(index int, actionIndex int) (EnvDecision, error) {
	if env == nil || index < 0 || index >= len(env.slots) {
		return EnvDecision{}, fmt.Errorf("%w: %d", ErrInvalidEnvIndex, index)
	}
	slot := &env.slots[index]
	if !slot.initialized {
		return EnvDecision{}, fmt.Errorf("%w: env %d", ErrEnvNotInitialized, index)
	}
	if slot.done {
		return EnvDecision{}, fmt.Errorf("%w: env %d requires reset", ErrEnvDone, index)
	}
	if actionIndex < 0 || actionIndex >= len(slot.legalActions) {
		return EnvDecision{}, fmt.Errorf("%w: env %d action %d legal count %d", ErrInvalidActionIndex, index, actionIndex, len(slot.legalActions))
	}

	action := slot.legalActions[actionIndex]
	actingFaction := slot.state.FactionTurn
	previousState := slot.state
	next, err := applyEnvAction(slot.state, action)
	if err != nil {
		return EnvDecision{}, fmt.Errorf("step env %d action %d: %w", index, actionIndex, err)
	}
	if err := rootengine.ValidateState(next); err != nil {
		return EnvDecision{}, fmt.Errorf("step env %d action %d produced invalid state: %w", index, actionIndex, err)
	}

	slot.state = next
	slot.steps++
	gameOver := next.GamePhase == game.LifecycleGameOver
	slot.truncated = !gameOver && slot.steps >= env.config.MaxSteps
	slot.done = gameOver || slot.truncated
	reward := rewardForTransition(
		slot.lastVictoryPoints,
		previousState,
		next,
		action,
		actingFaction,
		env.config,
		slot.truncated,
	)
	slot.lastVictoryPoints = victoryPointSnapshot(next)
	if slot.done {
		slot.legalActions = nil
	} else {
		slot.legalActions = rootengine.ValidActions(next)
		if len(slot.legalActions) == 0 {
			slot.done = true
			slot.truncated = true
		}
	}

	return env.decision(index, reward)
}

func (env *VecEnv) decision(index int, reward float32) (EnvDecision, error) {
	if index < 0 || index >= len(env.slots) {
		return EnvDecision{}, fmt.Errorf("%w: %d", ErrInvalidEnvIndex, index)
	}
	slot := &env.slots[index]
	if !slot.initialized {
		return EnvDecision{}, fmt.Errorf("%w: env %d", ErrEnvNotInitialized, index)
	}

	activeFaction := slot.state.FactionTurn
	obs := rootengine.NewTrainingObservation(slot.state, rootengine.TrainingObservationOptions{
		Perspective: activeFaction,
	})
	encodedObservation := EncodeObservation(obs)
	encodedActions := make([][]float32, len(slot.legalActions))
	actionTypes := make([]game.ActionType, len(slot.legalActions))
	for actionIndex, action := range slot.legalActions {
		encodedActions[actionIndex] = EncodeAction(slot.state, action)
		actionTypes[actionIndex] = action.Type
	}
	candidateRewards, err := candidateRewardVector(slot.state, slot.legalActions, env.config.TerminalWinBonus)
	if err != nil {
		return EnvDecision{}, fmt.Errorf("candidate rewards for env %d: %w", index, err)
	}

	return EnvDecision{
		EnvIndex:          index,
		Episode:           slot.episode,
		Step:              slot.steps,
		RoundNumber:       slot.state.RoundNumber,
		ActiveFaction:     activeFaction,
		CurrentPhase:      slot.state.CurrentPhase,
		CurrentStep:       slot.state.CurrentStep,
		Eyrie:             eyrieDiagnostics(slot.state),
		Observation:       encodedObservation,
		CandidateActions:  encodedActions,
		CandidateRewards:  candidateRewards,
		CandidateCount:    len(encodedActions),
		Reward:            reward,
		Done:              slot.done,
		Truncated:         slot.truncated,
		Winner:            slot.state.Winner,
		WinningCoalition:  append([]game.Faction(nil), slot.state.WinningCoalition...),
		VictoryPoints:     victoryPointVector(slot.state),
		LegalActionTypes:  actionTypes,
		ObservationLength: len(encodedObservation),
		ActionLength:      ActionVectorLength(),
	}, nil
}

func eyrieDiagnostics(state game.GameState) EyrieDiagnostics {
	currentColumn := -1
	if column, ok := currentEyrieDecreeColumn(state); ok {
		currentColumn = int(column)
	}

	return EyrieDiagnostics{
		VictoryPoints:         state.VictoryPoints[game.Eyrie],
		RoostsPlaced:          state.Eyrie.RoostsPlaced,
		WarriorSupply:         state.Eyrie.WarriorSupply,
		DecreeColumnCounts:    eyrieDecreeColumnCounts(state.Eyrie.Decree),
		CurrentDecreeColumn:   currentColumn,
		DecreeColumnsResolved: state.TurnProgress.DecreeColumnsResolved,
		DecreeCardsResolved:   state.TurnProgress.DecreeCardsResolved,
		CardsAddedToDecree:    state.TurnProgress.CardsAddedToDecree,
	}
}

func eyrieDecreeColumnCounts(decree game.Decree) []int {
	return []int{
		len(decree.Recruit),
		len(decree.Move),
		len(decree.Battle),
		len(decree.Build),
	}
}

func currentEyrieDecreeColumn(state game.GameState) (game.DecreeColumn, bool) {
	for columnIndex := state.TurnProgress.DecreeColumnsResolved; columnIndex < 4; columnIndex++ {
		column := game.DecreeColumn(columnIndex)
		for _, cardID := range eyrieDecreeCardsByColumn(state.Eyrie.Decree, column) {
			if !eyrieDecreeCardResolved(state, cardID) {
				return column, true
			}
		}
	}
	return 0, false
}

func eyrieDecreeCardsByColumn(decree game.Decree, column game.DecreeColumn) []game.CardID {
	switch column {
	case game.DecreeRecruit:
		return decree.Recruit
	case game.DecreeMove:
		return decree.Move
	case game.DecreeBattle:
		return decree.Battle
	case game.DecreeBuild:
		return decree.Build
	default:
		return nil
	}
}

func eyrieDecreeCardResolved(state game.GameState, cardID game.CardID) bool {
	for _, resolvedID := range state.TurnProgress.ResolvedDecreeCardIDs {
		if resolvedID == cardID {
			return true
		}
	}
	return false
}

func normalizeVecEnvConfig(config VecEnvConfig) (VecEnvConfig, error) {
	if config.NumEnvs <= 0 {
		return VecEnvConfig{}, fmt.Errorf("%w: numEnvs must be positive", ErrInvalidVecEnvConfig)
	}
	if config.BaseSeed == 0 {
		return VecEnvConfig{}, fmt.Errorf("%w: baseSeed must be nonzero", ErrInvalidVecEnvConfig)
	}
	if config.MaxSteps == 0 {
		config.MaxSteps = DefaultVecEnvMaxSteps
	}
	if config.MaxSteps < 0 {
		return VecEnvConfig{}, fmt.Errorf("%w: maxSteps must be positive", ErrInvalidVecEnvConfig)
	}
	if len(config.Factions) == 0 {
		config.Factions = []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond}
	}
	if len(config.Factions) < 2 || len(config.Factions) > factionCount {
		return VecEnvConfig{}, fmt.Errorf("%w: factions must contain 2 to 4 factions", ErrInvalidVecEnvConfig)
	}
	if hasDuplicateFaction(config.Factions) {
		return VecEnvConfig{}, fmt.Errorf("%w: factions must be unique", ErrInvalidVecEnvConfig)
	}
	if !containsFaction(config.Factions, config.PlayerFaction) {
		if config.PlayerFaction == 0 {
			config.PlayerFaction = config.Factions[0]
		} else {
			return VecEnvConfig{}, fmt.Errorf("%w: playerFaction must be in factions", ErrInvalidVecEnvConfig)
		}
	}
	if config.MapID == "" {
		config.MapID = game.AutumnMapID
	}
	if config.MapID != game.AutumnMapID {
		return VecEnvConfig{}, fmt.Errorf("%w: unsupported map %q", ErrInvalidVecEnvConfig, config.MapID)
	}
	config.TrackAllHands = true
	if config.TerminalWinBonus == 0 {
		config.TerminalWinBonus = DefaultTerminalWinBonus
	}
	if config.StepPenalty == 0 {
		config.StepPenalty = DefaultStepPenalty
	}
	if config.TruncationPenalty == 0 {
		config.TruncationPenalty = DefaultTruncationPenalty
	}
	if config.EyrieTurmoilPenalty == 0 {
		config.EyrieTurmoilPenalty = DefaultEyrieTurmoilPenalty
	}
	if config.EyrieTurmoilVPLossPenalty == 0 {
		config.EyrieTurmoilVPLossPenalty = DefaultEyrieTurmoilVPLossPenalty
	}
	if config.EyrieScoreRoostsBonus == 0 {
		config.EyrieScoreRoostsBonus = DefaultEyrieScoreRoostsBonus
	}
	if config.EyrieRoostBuildBonus == 0 {
		config.EyrieRoostBuildBonus = DefaultEyrieRoostBuildBonus
	}
	if config.EyrieRoostLossPenalty == 0 {
		config.EyrieRoostLossPenalty = DefaultEyrieRoostLossPenalty
	}
	if config.EyrieBuildDecreeCardPenalty == 0 {
		config.EyrieBuildDecreeCardPenalty = DefaultEyrieBuildDecreeCardPenalty
	}
	config.Factions = append([]game.Faction(nil), config.Factions...)
	return config, nil
}

func victoryPointSnapshot(state game.GameState) map[game.Faction]int {
	snapshot := make(map[game.Faction]int, len(orderedFactions))
	for _, faction := range orderedFactions {
		snapshot[faction] = state.VictoryPoints[faction]
	}
	return snapshot
}

func victoryPointVector(state game.GameState) []int {
	points := make([]int, len(orderedFactions))
	for index, faction := range orderedFactions {
		points[index] = state.VictoryPoints[faction]
	}
	return points
}

func candidateRewardVector(state game.GameState, actions []game.Action, terminalWinBonus float32) ([]float32, error) {
	rewards := make([]float32, len(actions))
	previous := victoryPointSnapshot(state)
	actingFaction := state.FactionTurn
	for index, action := range actions {
		next, err := applyEnvAction(rootengine.CloneState(state), action)
		if err != nil {
			return nil, fmt.Errorf("action %d: %w", index, err)
		}
		rewards[index] = rewardForTransition(
			previous,
			state,
			next,
			action,
			actingFaction,
			VecEnvConfig{
				TerminalWinBonus:          terminalWinBonus,
				DisableEyrieRewardShaping: true,
			},
			false,
		)
	}
	return rewards, nil
}

func rewardForTransition(
	previous map[game.Faction]int,
	previousState game.GameState,
	next game.GameState,
	action game.Action,
	actingFaction game.Faction,
	config VecEnvConfig,
	truncated bool,
) float32 {
	reward := float32(next.VictoryPoints[actingFaction]-previous[actingFaction]) - config.StepPenalty
	reward += eyrieRewardShaping(previous, previousState, next, action, actingFaction, config)
	if truncated {
		reward -= config.TruncationPenalty
	}
	if next.GamePhase != game.LifecycleGameOver {
		return reward
	}
	if next.Winner == actingFaction || factionInSlice(next.WinningCoalition, actingFaction) {
		reward += config.TerminalWinBonus
	} else {
		reward -= config.TerminalWinBonus
	}
	return reward
}

func eyrieRewardShaping(
	previous map[game.Faction]int,
	previousState game.GameState,
	next game.GameState,
	action game.Action,
	actingFaction game.Faction,
	config VecEnvConfig,
) float32 {
	if config.DisableEyrieRewardShaping || actingFaction != game.Eyrie {
		return 0
	}

	reward := float32(0)
	beforeVP := previous[game.Eyrie]
	afterVP := next.VictoryPoints[game.Eyrie]
	vpDelta := afterVP - beforeVP

	if action.Type == game.ActionTurmoil {
		reward -= config.EyrieTurmoilPenalty
		if vpDelta < 0 {
			reward -= float32(-vpDelta) * config.EyrieTurmoilVPLossPenalty
		}
	}
	if action.Type == game.ActionScoreRoosts && vpDelta > 0 {
		reward += float32(vpDelta) * config.EyrieScoreRoostsBonus
	}

	roostDelta := next.Eyrie.RoostsPlaced - previousState.Eyrie.RoostsPlaced
	if roostDelta > 0 {
		reward += float32(roostDelta) * config.EyrieRoostBuildBonus
	} else if roostDelta < 0 {
		reward -= float32(-roostDelta) * config.EyrieRoostLossPenalty
	}

	if action.Type == game.ActionAddToDecree && action.AddToDecree != nil {
		for _, column := range action.AddToDecree.Columns {
			if column == game.DecreeBuild {
				reward -= config.EyrieBuildDecreeCardPenalty
			}
		}
	}

	return reward
}

func factionInSlice(factions []game.Faction, target game.Faction) bool {
	for _, faction := range factions {
		if faction == target {
			return true
		}
	}
	return false
}

func seedForSlot(baseSeed int64, slotIndex int, episode int) int64 {
	return baseSeed + int64(slotIndex)*100_003 + int64(episode)*1_000_003
}

func containsFaction(factions []game.Faction, target game.Faction) bool {
	for _, faction := range factions {
		if faction == target {
			return true
		}
	}
	return false
}

func hasDuplicateFaction(factions []game.Faction) bool {
	seen := map[game.Faction]bool{}
	for _, faction := range factions {
		if seen[faction] {
			return true
		}
		seen[faction] = true
	}
	return false
}
