package server

import "github.com/imdehydrated/rootbuddy/game"

func validateApplyActionRequest(req ApplyActionRequest) string {
	switch req.Action.Type {
	case game.ActionMovement:
		if req.Action.Movement == nil {
			return "movement payload is required"
		}
		count := req.Action.Movement.Count
		if count <= 0 {
			count = req.Action.Movement.MaxCount
		}
		if count <= 0 || req.Action.Movement.From <= 0 || req.Action.Movement.To <= 0 {
			return "movement action must have positive count and valid clearing IDs"
		}
	case game.ActionBattleResolution:
		if req.Action.BattleResolution == nil {
			return "battle resolution payload is required"
		}
		if req.Action.BattleResolution.ClearingID <= 0 {
			return "battle resolution must have a valid clearing ID"
		}
		if req.Action.BattleResolution.AttackerLosses < 0 || req.Action.BattleResolution.DefenderLosses < 0 {
			return "battle resolution losses cannot be negative"
		}
	case game.ActionBuild:
		if req.Action.Build == nil {
			return "build payload is required"
		}
		if req.Action.Build.ClearingID <= 0 {
			return "build action must have a valid clearing ID"
		}
	case game.ActionRecruit:
		if req.Action.Recruit == nil {
			return "recruit payload is required"
		}
		if len(req.Action.Recruit.ClearingIDs) == 0 {
			return "recruit action must target at least one clearing"
		}
	case game.ActionOverwork:
		if req.Action.Overwork == nil {
			return "overwork payload is required"
		}
		if req.Action.Overwork.ClearingID <= 0 || req.Action.Overwork.CardID <= 0 {
			return "overwork action must have a valid clearing ID and card ID"
		}
	case game.ActionCraft:
		if req.Action.Craft == nil {
			return "craft payload is required"
		}
		if req.Action.Craft.CardID <= 0 {
			return "craft action must have a valid card ID"
		}
	case game.ActionBattle:
		return "battle initiation cannot be applied directly; resolve it first"
	default:
		return "unsupported action type"
	}

	return ""
}

func validateResolveBattleRequest(req ResolveBattleRequest) string {
	if req.Action.Type != game.ActionBattle {
		return "battle resolution requires a battle action"
	}
	if req.Action.Battle == nil {
		return "battle payload is required"
	}
	if req.Action.Battle.ClearingID <= 0 {
		return "battle action must have a valid clearing ID"
	}
	if req.Action.Battle.Faction == req.Action.Battle.TargetFaction {
		return "battle action must target a different faction"
	}
	if req.AttackerRoll < 0 || req.AttackerRoll > 3 || req.DefenderRoll < 0 || req.DefenderRoll > 3 {
		return "battle rolls must be between 0 and 3"
	}

	return ""
}
