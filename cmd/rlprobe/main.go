package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

var actionTypeNames = map[game.ActionType]string{
	game.ActionMovement:             "Movement",
	game.ActionBattle:               "Battle",
	game.ActionBattleResolution:     "BattleResolution",
	game.ActionBuild:                "Build",
	game.ActionRecruit:              "Recruit",
	game.ActionOverwork:             "Overwork",
	game.ActionCraft:                "Craft",
	game.ActionAddToDecree:          "AddToDecree",
	game.ActionSpreadSympathy:       "SpreadSympathy",
	game.ActionRevolt:               "Revolt",
	game.ActionMobilize:             "Mobilize",
	game.ActionTrain:                "Train",
	game.ActionOrganize:             "Organize",
	game.ActionExplore:              "Explore",
	game.ActionQuest:                "Quest",
	game.ActionAid:                  "Aid",
	game.ActionStrike:               "Strike",
	game.ActionRepair:               "Repair",
	game.ActionTurmoil:              "Turmoil",
	game.ActionDaybreak:             "Daybreak",
	game.ActionSlip:                 "Slip",
	game.ActionBirdsongWood:         "BirdsongWood",
	game.ActionEveningDraw:          "EveningDraw",
	game.ActionScoreRoosts:          "ScoreRoosts",
	game.ActionPassPhase:            "PassPhase",
	game.ActionAddCardToHand:        "AddCardToHand",
	game.ActionRemoveCardFromHand:   "RemoveCardFromHand",
	game.ActionOtherPlayerDraw:      "OtherPlayerDraw",
	game.ActionOtherPlayerPlay:      "OtherPlayerPlay",
	game.ActionDiscardEffect:        "DiscardEffect",
	game.ActionActivateDominance:    "ActivateDominance",
	game.ActionTakeDominance:        "TakeDominance",
	game.ActionMarquiseSetup:        "MarquiseSetup",
	game.ActionEyrieSetup:           "EyrieSetup",
	game.ActionVagabondSetup:        "VagabondSetup",
	game.ActionUsePersistentEffect:  "UsePersistentEffect",
	game.ActionEyrieEmergencyOrders: "EyrieEmergencyOrders",
	game.ActionEyrieNewRoost:        "EyrieNewRoost",
	game.ActionVagabondRest:         "VagabondRest",
	game.ActionVagabondDiscard:      "VagabondDiscard",
	game.ActionVagabondItemCapacity: "VagabondItemCapacity",
	game.ActionEveningDiscard:       "EveningDiscard",
	game.ActionFieldHospitals:       "FieldHospitals",
	game.ActionMarquiseExtraAction:  "MarquiseExtraAction",
	game.ActionVagabondSteal:        "VagabondSteal",
	game.ActionVagabondDayLabor:     "VagabondDayLabor",
	game.ActionVagabondHideout:      "VagabondHideout",
	game.ActionResolveOutrage:       "ResolveOutrage",
}

type config struct {
	games        int
	maxSteps     int
	seed         int64
	verbose      bool
	showTopTypes int
}

type probeStats struct {
	gamesStarted     int
	gamesCompleted   int
	gamesMaxSteps    int
	totalSteps       int
	minSteps         int
	maxSteps         int
	validActionCalls int
	validActionTotal int
	validActionMax   int
	validActionTime  time.Duration
	applyTime        time.Duration
	winners          map[game.Faction]int
	legalByType      map[game.ActionType]int
	chosenByType     map[game.ActionType]int
	windows          map[string]*windowStats
	battles          battleStats
	rejectedExamples []string
}

type windowStats struct {
	count int
	sum   int
	max   int
}

type battleStats struct {
	legalCandidates  int
	chosen           int
	applied          int
	rejected         int
	filledPieceLoss  int
	filledItemDamage int
	attackerRolls    map[int]int
	defenderRolls    map[int]int
	pairs            map[string]int
}

type actionResult struct {
	next     game.GameState
	applied  bool
	reason   string
	resolved game.Action
}

func main() {
	cfg := parseFlags()
	stats := newProbeStats()

	started := time.Now()
	for gameIndex := 0; gameIndex < cfg.games; gameIndex++ {
		seed := cfg.seed + int64(gameIndex)*100_003
		steps, completed, err := runGame(cfg, stats, gameIndex, seed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rlprobe failed: %v\n", err)
			os.Exit(1)
		}
		if cfg.verbose {
			fmt.Printf("game=%d seed=%d steps=%d completed=%t\n", gameIndex, seed, steps, completed)
		}
	}

	printReport(cfg, stats, time.Since(started))
}

func parseFlags() config {
	cfg := config{}
	flag.IntVar(&cfg.games, "games", 20, "number of random self-play games to run")
	flag.IntVar(&cfg.maxSteps, "max-steps", 5000, "maximum decision steps per game")
	flag.Int64Var(&cfg.seed, "seed", 707, "base nonzero seed for deterministic probe runs")
	flag.BoolVar(&cfg.verbose, "v", false, "print one-line per-game summaries")
	flag.IntVar(&cfg.showTopTypes, "top-types", 16, "number of action types to show in histograms")
	flag.Parse()

	if cfg.games <= 0 {
		fatalFlag("games must be positive")
	}
	if cfg.maxSteps <= 0 {
		fatalFlag("max-steps must be positive")
	}
	if cfg.seed == 0 {
		fatalFlag("seed must be nonzero")
	}
	return cfg
}

func fatalFlag(message string) {
	fmt.Fprintf(os.Stderr, "rlprobe: %s\n", message)
	os.Exit(2)
}

func newProbeStats() *probeStats {
	return &probeStats{
		minSteps:         int(^uint(0) >> 1),
		winners:          map[game.Faction]int{},
		legalByType:      map[game.ActionType]int{},
		chosenByType:     map[game.ActionType]int{},
		windows:          map[string]*windowStats{},
		rejectedExamples: []string{},
		battles: battleStats{
			attackerRolls: map[int]int{},
			defenderRolls: map[int]int{},
			pairs:         map[string]int{},
		},
	}
}

func runGame(cfg config, stats *probeStats, gameIndex int, seed int64) (int, bool, error) {
	state, err := engine.SetupTrainingGame(engine.SetupRequest{
		GameMode:      game.GameModeOnline,
		PlayerFaction: game.Marquise,
		TrackAllHands: true,
		Factions:      []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond},
		MapID:         game.AutumnMapID,
		RandomSeed:    seed,
	})
	if err != nil {
		return 0, false, fmt.Errorf("setup game %d seed %d: %w", gameIndex, seed, err)
	}
	if err := engine.ValidateState(state); err != nil {
		return 0, false, fmt.Errorf("initial state game %d seed %d invalid: %w", gameIndex, seed, err)
	}

	rng := rand.New(rand.NewSource(seed ^ 0x726c70726f6265))
	stats.gamesStarted++

	for step := 0; step < cfg.maxSteps; step++ {
		if state.GamePhase == game.LifecycleGameOver {
			recordGameEnd(stats, state, step, true)
			return step, true, nil
		}

		actions, validDuration := timedValidActions(state)
		stats.recordLegalActions(state, actions, validDuration)
		if len(actions) == 0 {
			return step, false, fmt.Errorf("game %d step %d produced no legal actions before game over: state=%+v", gameIndex, step, state)
		}

		startApply := time.Now()
		result, chosen, err := chooseAndApply(state, actions, rng, stats)
		stats.applyTime += time.Since(startApply)
		if err != nil {
			return step, false, fmt.Errorf("game %d step %d: %w", gameIndex, step, err)
		}
		stats.chosenByType[chosen.Type]++

		state = result.next
		if err := engine.ValidateState(state); err != nil {
			return step, false, fmt.Errorf("game %d step %d action %s produced invalid state: %w", gameIndex, step, actionTypeName(chosen.Type), err)
		}
		stats.totalSteps++
	}

	recordGameEnd(stats, state, cfg.maxSteps, false)
	return cfg.maxSteps, false, nil
}

func timedValidActions(state game.GameState) ([]game.Action, time.Duration) {
	start := time.Now()
	actions := engine.ValidActions(state)
	return actions, time.Since(start)
}

func chooseAndApply(state game.GameState, actions []game.Action, rng *rand.Rand, stats *probeStats) (actionResult, game.Action, error) {
	order := rng.Perm(len(actions))
	var lastRejected string

	for _, actionIndex := range order {
		action := actions[actionIndex]
		if action.Type == game.ActionBattle {
			stats.battles.chosen++
			result := applyBattleAction(state, action, stats)
			if result.applied {
				return result, action, nil
			}
			lastRejected = result.reason
			continue
		}

		next, _, err := engine.ApplyLegalActionDetailed(state, action)
		if err != nil {
			return actionResult{}, action, fmt.Errorf("apply legal action %s: %w", actionTypeName(action.Type), err)
		}
		return actionResult{next: next, applied: true}, action, nil
	}

	return actionResult{}, game.Action{}, fmt.Errorf("all %d candidate actions were rejected; last rejection: %s", len(actions), lastRejected)
}

func applyBattleAction(state game.GameState, action game.Action, stats *probeStats) actionResult {
	attackerRoll, defenderRoll, err := engine.RollBattleDice(state)
	if err != nil {
		stats.battles.rejected++
		return actionResult{next: state, reason: fmt.Sprintf("roll dice: %v", err)}
	}

	resolved := completeBattleResolution(state, engine.ResolveBattle(state, action, attackerRoll, defenderRoll), stats)
	next := engine.ApplyAction(state, resolved)
	if next.BattleRollCount != state.BattleRollCount+1 {
		stats.battles.rejected++
		reason := battleRejectionReason(action, resolved)
		stats.recordBattleRejection(reason)
		return actionResult{next: state, reason: reason, resolved: resolved}
	}

	stats.battles.applied++
	stats.battles.attackerRolls[attackerRoll]++
	stats.battles.defenderRolls[defenderRoll]++
	if action.Battle != nil {
		stats.battles.pairs[battlePairName(action.Battle.Faction, action.Battle.TargetFaction)]++
	}
	return actionResult{next: next, applied: true, resolved: resolved}
}

func completeBattleResolution(state game.GameState, resolved game.Action, stats *probeStats) game.Action {
	if resolved.BattleResolution == nil {
		return resolved
	}

	completed := resolved
	resolution := *resolved.BattleResolution
	clearing, ok := findClearing(state, resolution.ClearingID)
	if !ok {
		return resolved
	}

	if resolution.Faction == game.Vagabond {
		damageHits := resolution.AttackerLosses - resolution.AlliedWarriorLosses
		if len(resolution.AttackerDamagedItemIndexes) == 0 {
			resolution.AttackerDamagedItemIndexes = chooseVagabondDamage(state, damageHits)
			if len(resolution.AttackerDamagedItemIndexes) > 0 {
				stats.battles.filledItemDamage++
			}
		}
	} else if len(resolution.AttackerPieceLosses) == 0 {
		resolution.AttackerPieceLosses = choosePieceLosses(clearing, resolution.Faction, resolution.AttackerLosses)
		if len(resolution.AttackerPieceLosses) > 0 {
			stats.battles.filledPieceLoss++
		}
	}

	if resolution.TargetFaction == game.Vagabond {
		if len(resolution.DefenderDamagedItemIndexes) == 0 {
			resolution.DefenderDamagedItemIndexes = chooseVagabondDamage(state, resolution.DefenderLosses)
			if len(resolution.DefenderDamagedItemIndexes) > 0 {
				stats.battles.filledItemDamage++
			}
		}
	} else if len(resolution.DefenderPieceLosses) == 0 {
		resolution.DefenderPieceLosses = choosePieceLosses(clearing, resolution.TargetFaction, resolution.DefenderLosses)
		if len(resolution.DefenderPieceLosses) > 0 {
			stats.battles.filledPieceLoss++
		}
	}

	completed.BattleResolution = &resolution
	return completed
}

func findClearing(state game.GameState, clearingID int) (game.Clearing, bool) {
	for _, clearing := range state.Map.Clearings {
		if clearing.ID == clearingID {
			return clearing, true
		}
	}
	return game.Clearing{}, false
}

func chooseVagabondDamage(state game.GameState, hits int) []int {
	if hits <= 0 {
		return nil
	}
	indexes := []int{}
	for index, item := range state.Vagabond.Items {
		if item.Status == game.ItemDamaged {
			continue
		}
		indexes = append(indexes, index)
		if len(indexes) >= hits {
			break
		}
	}
	return indexes
}

func choosePieceLosses(clearing game.Clearing, faction game.Faction, losses int) []game.BattlePieceLoss {
	if losses <= 0 {
		return nil
	}

	warriors := 0
	if clearing.Warriors != nil {
		warriors = clearing.Warriors[faction]
	}
	remaining := losses - minInt(losses, warriors)
	if remaining <= 0 {
		return nil
	}

	pool := newPieceLossPool(clearing, faction)
	choices := []game.BattlePieceLoss{}
	for remaining > 0 && pool.total() > 0 {
		loss, ok := pool.takeFirst()
		if !ok {
			break
		}
		choices = append(choices, loss)
		remaining--
	}
	return choices
}

type pieceLossPool struct {
	buildings map[game.BuildingType]int
	tokens    map[game.TokenType]int
	wood      int
}

func newPieceLossPool(clearing game.Clearing, faction game.Faction) pieceLossPool {
	pool := pieceLossPool{
		buildings: map[game.BuildingType]int{},
		tokens:    map[game.TokenType]int{},
	}
	for _, building := range clearing.Buildings {
		if building.Faction == faction {
			pool.buildings[building.Type]++
		}
	}
	for _, token := range clearing.Tokens {
		if token.Faction == faction {
			pool.tokens[token.Type]++
		}
	}
	if faction == game.Marquise {
		pool.wood = clearing.Wood
	}
	return pool
}

func (pool pieceLossPool) total() int {
	total := pool.wood
	for _, count := range pool.buildings {
		total += count
	}
	for _, count := range pool.tokens {
		total += count
	}
	return total
}

func (pool *pieceLossPool) takeFirst() (game.BattlePieceLoss, bool) {
	for _, buildingType := range []game.BuildingType{game.Sawmill, game.Workshop, game.Recruiter, game.Roost, game.Base} {
		if pool.buildings[buildingType] <= 0 {
			continue
		}
		pool.buildings[buildingType]--
		return game.BattlePieceLoss{
			Kind:         game.BattlePieceBuilding,
			BuildingType: buildingType,
		}, true
	}
	for _, tokenType := range []game.TokenType{game.TokenKeep, game.TokenSympathy} {
		if pool.tokens[tokenType] <= 0 {
			continue
		}
		pool.tokens[tokenType]--
		return game.BattlePieceLoss{
			Kind:      game.BattlePieceToken,
			TokenType: tokenType,
		}, true
	}
	if pool.wood > 0 {
		pool.wood--
		return game.BattlePieceLoss{
			Kind: game.BattlePieceWood,
		}, true
	}
	return game.BattlePieceLoss{}, false
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func battleRejectionReason(action game.Action, resolved game.Action) string {
	if action.Battle == nil {
		return "battle action missing payload"
	}
	if resolved.BattleResolution == nil {
		return fmt.Sprintf("%s in clearing %d resolved to empty battle resolution", battlePairName(action.Battle.Faction, action.Battle.TargetFaction), action.Battle.ClearingID)
	}
	resolution := resolved.BattleResolution
	parts := []string{
		fmt.Sprintf("%s in clearing %d", battlePairName(resolution.Faction, resolution.TargetFaction), resolution.ClearingID),
		fmt.Sprintf("rolls=%d/%d", resolution.AttackerRoll, resolution.DefenderRoll),
		fmt.Sprintf("losses=%d/%d", resolution.AttackerLosses, resolution.DefenderLosses),
	}
	if resolution.AttackerLosses > 0 && resolution.Faction == game.Vagabond && len(resolution.AttackerDamagedItemIndexes) == 0 {
		parts = append(parts, "attacker Vagabond damage choices absent")
	}
	if resolution.DefenderLosses > 0 && resolution.TargetFaction == game.Vagabond && len(resolution.DefenderDamagedItemIndexes) == 0 {
		parts = append(parts, "defender Vagabond damage choices absent")
	}
	if resolution.AttackerLosses > 0 && resolution.Faction != game.Vagabond && len(resolution.AttackerPieceLosses) == 0 {
		parts = append(parts, "attacker piece-loss choices absent")
	}
	if resolution.DefenderLosses > 0 && resolution.TargetFaction != game.Vagabond && len(resolution.DefenderPieceLosses) == 0 {
		parts = append(parts, "defender piece-loss choices absent")
	}
	return strings.Join(parts, "; ")
}

func (stats *probeStats) recordLegalActions(state game.GameState, actions []game.Action, duration time.Duration) {
	stats.validActionCalls++
	stats.validActionTotal += len(actions)
	stats.validActionTime += duration
	if len(actions) > stats.validActionMax {
		stats.validActionMax = len(actions)
	}

	window := stats.window(state)
	window.count++
	window.sum += len(actions)
	if len(actions) > window.max {
		window.max = len(actions)
	}

	for _, action := range actions {
		stats.legalByType[action.Type]++
		if action.Type == game.ActionBattle {
			stats.battles.legalCandidates++
		}
	}
}

func (stats *probeStats) window(state game.GameState) *windowStats {
	key := stateWindowName(state)
	window := stats.windows[key]
	if window == nil {
		window = &windowStats{}
		stats.windows[key] = window
	}
	return window
}

func (stats *probeStats) recordBattleRejection(reason string) {
	if len(stats.rejectedExamples) >= 8 {
		return
	}
	stats.rejectedExamples = append(stats.rejectedExamples, reason)
}

func recordGameEnd(stats *probeStats, state game.GameState, steps int, completed bool) {
	if completed {
		stats.gamesCompleted++
		stats.winners[state.Winner]++
	} else {
		stats.gamesMaxSteps++
	}
	if steps < stats.minSteps {
		stats.minSteps = steps
	}
	if steps > stats.maxSteps {
		stats.maxSteps = steps
	}
}

func printReport(cfg config, stats *probeStats, elapsed time.Duration) {
	fmt.Println("RL probe report")
	fmt.Printf("games: started=%d completed=%d max_steps=%d\n", stats.gamesStarted, stats.gamesCompleted, stats.gamesMaxSteps)
	fmt.Printf("config: games=%d max_steps=%d seed=%d\n", cfg.games, cfg.maxSteps, cfg.seed)
	fmt.Printf("elapsed: %s\n", elapsed.Round(time.Millisecond))
	if stats.gamesStarted > 0 {
		fmt.Printf("steps/game: avg=%.1f min=%d max=%d total=%d\n", float64(stats.totalSteps)/float64(stats.gamesStarted), stats.minSteps, stats.maxSteps, stats.totalSteps)
	}
	if stats.totalSteps > 0 {
		fmt.Printf("throughput: %.1f applied_steps/sec\n", float64(stats.totalSteps)/elapsed.Seconds())
	}
	if stats.validActionCalls > 0 {
		fmt.Printf("valid actions: calls=%d avg=%.2f max=%d avg_latency=%s total_latency=%s\n",
			stats.validActionCalls,
			float64(stats.validActionTotal)/float64(stats.validActionCalls),
			stats.validActionMax,
			(stats.validActionTime / time.Duration(stats.validActionCalls)).Round(time.Microsecond),
			stats.validActionTime.Round(time.Millisecond),
		)
	}
	if stats.totalSteps > 0 {
		fmt.Printf("apply latency: avg=%s total=%s\n", (stats.applyTime / time.Duration(stats.totalSteps)).Round(time.Microsecond), stats.applyTime.Round(time.Millisecond))
	}

	printWinnerHistogram(stats)
	printBattleReport(stats)
	printTopActionHistogram("legal action-type histogram", stats.legalByType, cfg.showTopTypes)
	printTopActionHistogram("chosen action-type histogram", stats.chosenByType, cfg.showTopTypes)
	printWindowStats(stats)
}

func printWinnerHistogram(stats *probeStats) {
	if len(stats.winners) == 0 {
		fmt.Println("winners: none")
		return
	}
	fmt.Println("winners:")
	for _, faction := range sortedFactions(stats.winners) {
		fmt.Printf("  %s: %d\n", factionName(faction), stats.winners[faction])
	}
}

func printBattleReport(stats *probeStats) {
	fmt.Println("battles:")
	fmt.Printf("  legal_candidates=%d chosen=%d applied=%d rejected=%d filled_piece_loss=%d filled_item_damage=%d\n",
		stats.battles.legalCandidates,
		stats.battles.chosen,
		stats.battles.applied,
		stats.battles.rejected,
		stats.battles.filledPieceLoss,
		stats.battles.filledItemDamage,
	)
	if len(stats.battles.attackerRolls) > 0 || len(stats.battles.defenderRolls) > 0 {
		fmt.Printf("  attacker_rolls=%s defender_rolls=%s\n", intHistogram(stats.battles.attackerRolls), intHistogram(stats.battles.defenderRolls))
	}
	if len(stats.battles.pairs) > 0 {
		fmt.Println("  pairs:")
		for _, key := range sortedStringKeys(stats.battles.pairs) {
			fmt.Printf("    %s: %d\n", key, stats.battles.pairs[key])
		}
	}
	if len(stats.rejectedExamples) > 0 {
		fmt.Println("  rejected examples:")
		for _, example := range stats.rejectedExamples {
			fmt.Printf("    - %s\n", example)
		}
	}
}

func printTopActionHistogram(title string, counts map[game.ActionType]int, limit int) {
	fmt.Printf("%s:\n", title)
	if len(counts) == 0 {
		fmt.Println("  none")
		return
	}
	entries := make([]actionCount, 0, len(counts))
	for actionType, count := range counts {
		entries = append(entries, actionCount{actionType: actionType, count: count})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].actionType < entries[j].actionType
	})
	if limit <= 0 || limit > len(entries) {
		limit = len(entries)
	}
	for _, entry := range entries[:limit] {
		fmt.Printf("  %s: %d\n", actionTypeName(entry.actionType), entry.count)
	}
}

type actionCount struct {
	actionType game.ActionType
	count      int
}

func printWindowStats(stats *probeStats) {
	fmt.Println("valid actions by faction/phase/step:")
	if len(stats.windows) == 0 {
		fmt.Println("  none")
		return
	}
	keys := sortedWindowKeys(stats.windows)
	for _, key := range keys {
		window := stats.windows[key]
		fmt.Printf("  %s: calls=%d avg=%.2f max=%d\n", key, window.count, float64(window.sum)/float64(window.count), window.max)
	}
}

func sortedWindowKeys(windows map[string]*windowStats) []string {
	keys := make([]string, 0, len(windows))
	for key := range windows {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedStringKeys(values map[string]int) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedFactions(values map[game.Faction]int) []game.Faction {
	factions := make([]game.Faction, 0, len(values))
	for faction := range values {
		factions = append(factions, faction)
	}
	sort.Slice(factions, func(i, j int) bool {
		return factions[i] < factions[j]
	})
	return factions
}

func intHistogram(values map[int]int) string {
	if len(values) == 0 {
		return "{}"
	}
	keys := make([]int, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%d:%d", key, values[key]))
	}
	return "{" + strings.Join(parts, " ") + "}"
}

func stateWindowName(state game.GameState) string {
	if state.GamePhase == game.LifecycleSetup {
		return fmt.Sprintf("setup/%s", setupStageName(state.SetupStage))
	}
	return fmt.Sprintf("%s/%s/%s", factionName(state.FactionTurn), phaseName(state.CurrentPhase), stepName(state.CurrentStep))
}

func battlePairName(attacker game.Faction, defender game.Faction) string {
	return fmt.Sprintf("%s->%s", factionName(attacker), factionName(defender))
}

func actionTypeName(actionType game.ActionType) string {
	if name, ok := actionTypeNames[actionType]; ok {
		return name
	}
	return fmt.Sprintf("ActionType(%d)", actionType)
}

func factionName(faction game.Faction) string {
	switch faction {
	case game.Marquise:
		return "Marquise"
	case game.Eyrie:
		return "Eyrie"
	case game.Alliance:
		return "Alliance"
	case game.Vagabond:
		return "Vagabond"
	default:
		return fmt.Sprintf("Faction(%d)", faction)
	}
}

func phaseName(phase game.Phase) string {
	switch phase {
	case game.Birdsong:
		return "Birdsong"
	case game.Daylight:
		return "Daylight"
	case game.Evening:
		return "Evening"
	default:
		return fmt.Sprintf("Phase(%d)", phase)
	}
}

func stepName(step game.TurnStep) string {
	switch step {
	case game.StepUnspecified:
		return "Unspecified"
	case game.StepBirdsong:
		return "Birdsong"
	case game.StepDaylightCraft:
		return "DaylightCraft"
	case game.StepDaylightActions:
		return "DaylightActions"
	case game.StepEvening:
		return "Evening"
	default:
		return fmt.Sprintf("Step(%d)", step)
	}
}

func setupStageName(stage game.SetupStage) string {
	switch stage {
	case game.SetupStageUnspecified:
		return "Unspecified"
	case game.SetupStageMarquise:
		return "Marquise"
	case game.SetupStageEyrie:
		return "Eyrie"
	case game.SetupStageVagabond:
		return "Vagabond"
	case game.SetupStageComplete:
		return "Complete"
	default:
		return fmt.Sprintf("SetupStage(%d)", stage)
	}
}
