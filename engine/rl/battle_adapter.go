package rl

import (
	"errors"
	"fmt"

	rootengine "github.com/imdehydrated/rootbuddy/engine"
	"github.com/imdehydrated/rootbuddy/game"
)

var ErrBattleResolutionRejected = errors.New("battle resolution was rejected")

func applyEnvAction(state game.GameState, action game.Action) (game.GameState, error) {
	if action.Type == game.ActionBattle {
		return applyBattleAction(state, action)
	}

	next, _, err := rootengine.ApplyLegalActionDetailed(state, action)
	return next, err
}

func applyBattleAction(state game.GameState, action game.Action) (game.GameState, error) {
	attackerRoll, defenderRoll, err := rootengine.RollBattleDice(state)
	if err != nil {
		return game.GameState{}, fmt.Errorf("roll battle dice: %w", err)
	}

	resolved := rootengine.ResolveBattle(state, action, attackerRoll, defenderRoll)
	completed := completeBattleResolution(state, resolved)
	next := rootengine.ApplyAction(state, completed)
	if next.BattleRollCount != state.BattleRollCount+1 {
		return game.GameState{}, fmt.Errorf("%w: %s", ErrBattleResolutionRejected, battleRejectionReason(action, completed))
	}
	return next, nil
}

func completeBattleResolution(state game.GameState, resolved game.Action) game.Action {
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
		}
	} else if len(resolution.AttackerPieceLosses) == 0 {
		resolution.AttackerPieceLosses = choosePieceLosses(clearing, resolution.Faction, resolution.AttackerLosses)
	}

	if resolution.TargetFaction == game.Vagabond {
		if len(resolution.DefenderDamagedItemIndexes) == 0 {
			resolution.DefenderDamagedItemIndexes = chooseVagabondDamage(state, resolution.DefenderLosses)
		}
	} else if len(resolution.DefenderPieceLosses) == 0 {
		resolution.DefenderPieceLosses = choosePieceLosses(clearing, resolution.TargetFaction, resolution.DefenderLosses)
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
		return fmt.Sprintf("battle in clearing %d resolved to empty battle resolution", action.Battle.ClearingID)
	}
	resolution := resolved.BattleResolution
	return fmt.Sprintf(
		"%d attacking %d in clearing %d rolls=%d/%d losses=%d/%d",
		resolution.Faction,
		resolution.TargetFaction,
		resolution.ClearingID,
		resolution.AttackerRoll,
		resolution.DefenderRoll,
		resolution.AttackerLosses,
		resolution.DefenderLosses,
	)
}
