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
		hasSource := req.Action.Movement.From > 0 || req.Action.Movement.FromForestID > 0
		hasDestination := req.Action.Movement.To > 0 || req.Action.Movement.ToForestID > 0
		if count <= 0 || !hasSource || !hasDestination {
			return "movement action must have positive count and valid source and destination"
		}
	case game.ActionBattleResolution:
		if req.Action.BattleResolution == nil {
			return "battle resolution payload is required"
		}
		if req.Action.BattleResolution.ClearingID <= 0 {
			return "battle resolution must have a valid clearing ID"
		}
		if req.Action.BattleResolution.Faction == req.Action.BattleResolution.TargetFaction {
			return "battle resolution must target a different faction"
		}
		if !game.AreEnemies(req.State, req.Action.BattleResolution.Faction, req.Action.BattleResolution.TargetFaction) {
			return "battle resolution must target an enemy faction"
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
	case game.ActionAddToDecree:
		if req.Action.AddToDecree == nil {
			return "add-to-decree payload is required"
		}
	case game.ActionSpreadSympathy:
		if req.Action.SpreadSympathy == nil {
			return "spread sympathy payload is required"
		}
	case game.ActionRevolt:
		if req.Action.Revolt == nil {
			return "revolt payload is required"
		}
	case game.ActionMobilize:
		if req.Action.Mobilize == nil {
			return "mobilize payload is required"
		}
	case game.ActionTrain:
		if req.Action.Train == nil {
			return "train payload is required"
		}
	case game.ActionOrganize:
		if req.Action.Organize == nil {
			return "organize payload is required"
		}
	case game.ActionExplore:
		if req.Action.Explore == nil {
			return "explore payload is required"
		}
	case game.ActionQuest:
		if req.Action.Quest == nil {
			return "quest payload is required"
		}
	case game.ActionAid:
		if req.Action.Aid == nil {
			return "aid payload is required"
		}
	case game.ActionStrike:
		if req.Action.Strike == nil {
			return "strike payload is required"
		}
	case game.ActionRepair:
		if req.Action.Repair == nil {
			return "repair payload is required"
		}
	case game.ActionTurmoil:
		if req.Action.Turmoil == nil {
			return "turmoil payload is required"
		}
	case game.ActionDaybreak:
		if req.Action.Daybreak == nil {
			return "daybreak payload is required"
		}
	case game.ActionSlip:
		if req.Action.Slip == nil {
			return "slip payload is required"
		}
	case game.ActionBirdsongWood:
		if req.Action.BirdsongWood == nil {
			return "birdsong wood payload is required"
		}
	case game.ActionEveningDraw:
		if req.Action.EveningDraw == nil {
			return "evening draw payload is required"
		}
	case game.ActionScoreRoosts:
		if req.Action.ScoreRoosts == nil {
			return "score roosts payload is required"
		}
	case game.ActionPassPhase:
		if req.Action.PassPhase == nil {
			return "pass phase payload is required"
		}
	case game.ActionAddCardToHand:
		if req.Action.AddCardToHand == nil {
			return "add card to hand payload is required"
		}
		if req.Action.AddCardToHand.CardID <= 0 {
			return "add card to hand action must have a valid card ID"
		}
	case game.ActionRemoveCardFromHand:
		if req.Action.RemoveCardFromHand == nil {
			return "remove card from hand payload is required"
		}
		if req.Action.RemoveCardFromHand.CardID <= 0 {
			return "remove card from hand action must have a valid card ID"
		}
	case game.ActionOtherPlayerDraw:
		if req.Action.OtherPlayerDraw == nil {
			return "other player draw payload is required"
		}
		if req.Action.OtherPlayerDraw.Count <= 0 {
			return "other player draw action must have a positive count"
		}
	case game.ActionOtherPlayerPlay:
		if req.Action.OtherPlayerPlay == nil {
			return "other player play payload is required"
		}
		if req.Action.OtherPlayerPlay.CardID <= 0 {
			return "other player play action must have a valid card ID"
		}
	case game.ActionDiscardEffect:
		if req.Action.DiscardEffect == nil {
			return "discard effect payload is required"
		}
		if req.Action.DiscardEffect.CardID <= 0 {
			return "discard effect action must have a valid card ID"
		}
	case game.ActionActivateDominance:
		if req.Action.ActivateDominance == nil {
			return "activate dominance payload is required"
		}
		if req.Action.ActivateDominance.CardID <= 0 {
			return "activate dominance action must have a valid card ID"
		}
	case game.ActionTakeDominance:
		if req.Action.TakeDominance == nil {
			return "take dominance payload is required"
		}
		if req.Action.TakeDominance.DominanceCardID <= 0 || req.Action.TakeDominance.SpentCardID <= 0 {
			return "take dominance action must have valid dominance and spent card IDs"
		}
	case game.ActionMarquiseSetup:
		if req.Action.MarquiseSetup == nil {
			return "marquise setup payload is required"
		}
		if req.Action.MarquiseSetup.KeepClearingID <= 0 {
			return "marquise setup action must have a valid keep clearing ID"
		}
	case game.ActionEyrieSetup:
		if req.Action.EyrieSetup == nil {
			return "eyrie setup payload is required"
		}
		if req.Action.EyrieSetup.ClearingID <= 0 {
			return "eyrie setup action must have a valid clearing ID"
		}
	case game.ActionVagabondSetup:
		if req.Action.VagabondSetup == nil {
			return "vagabond setup payload is required"
		}
		if req.Action.VagabondSetup.ForestID <= 0 {
			return "vagabond setup action must have a valid forest ID"
		}
	case game.ActionUsePersistentEffect:
		if req.Action.UsePersistentEffect == nil {
			return "use persistent effect payload is required"
		}
		if req.Action.UsePersistentEffect.EffectID == "" {
			return "use persistent effect action must have an effect ID"
		}
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
	if !game.AreEnemies(req.State, req.Action.Battle.Faction, req.Action.Battle.TargetFaction) {
		return "battle action must target an enemy faction"
	}
	if req.AttackerRoll < 0 || req.AttackerRoll > 3 || req.DefenderRoll < 0 || req.DefenderRoll > 3 {
		return "battle rolls must be between 0 and 3"
	}

	return ""
}

func validateBattleContextRequest(req BattleContextRequest) string {
	if req.Action.Type != game.ActionBattle {
		return "battle context requires a battle action"
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
	if !game.AreEnemies(req.State, req.Action.Battle.Faction, req.Action.Battle.TargetFaction) {
		return "battle action must target an enemy faction"
	}

	return ""
}
