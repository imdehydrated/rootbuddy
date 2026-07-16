package rl

import "github.com/imdehydrated/rootbuddy/game"

const (
	actionScalarCount = 40
	actionFlagCount   = 32
)

type actionFeatures struct {
	actor           game.Faction
	actorPresent    bool
	target          game.Faction
	targetPresent   bool
	clearings       [maxClearings]float32
	sources         [maxClearings]float32
	destinations    [maxClearings]float32
	forests         [maxForests]float32
	cards           [baseDeckCardCount + specialCardIDSlots]float32
	cardSuits       [suitCount]float32
	cardKinds       [cardKindCount]float32
	items           [itemTypeCount]float32
	buildings       [buildingTypeCount]float32
	decreeColumns   [decreeColumnCount]float32
	attackerRolls   [4]float32
	defenderRolls   [4]float32
	scalars         [actionScalarCount]float32
	flags           [actionFlagCount]float32
	nextScalarIndex int
	nextFlagIndex   int
}

func EncodeAction(state game.GameState, action game.Action) []float32 {
	features := actionFeatures{}
	features.collect(state, action)

	out := make([]float32, 0, 256)
	out = appendOneHot(out, int(action.Type), actionTypeCount)
	out = appendOptionalFactionOneHot(out, features.actor, features.actorPresent)
	out = appendOptionalFactionOneHot(out, features.target, features.targetPresent)
	out = appendFloatArray(out, features.clearings[:])
	out = appendFloatArray(out, features.sources[:])
	out = appendFloatArray(out, features.destinations[:])
	out = appendFloatArray(out, features.forests[:])
	out = appendFloatArray(out, features.cards[:])
	out = appendFloatArray(out, features.cardSuits[:])
	out = appendFloatArray(out, features.cardKinds[:])
	out = appendFloatArray(out, features.items[:])
	out = appendFloatArray(out, features.buildings[:])
	out = appendFloatArray(out, features.decreeColumns[:])
	out = appendFloatArray(out, features.attackerRolls[:])
	out = appendFloatArray(out, features.defenderRolls[:])
	out = appendFloatArray(out, features.scalars[:])
	out = appendFloatArray(out, features.flags[:])
	return out
}

func appendFloatArray(out []float32, values []float32) []float32 {
	return append(out, values...)
}

func (features *actionFeatures) collect(state game.GameState, action game.Action) {
	switch action.Type {
	case game.ActionMovement:
		if payload := action.Movement; payload != nil {
			features.setActor(payload.Faction)
			features.addSource(payload.From)
			features.addDestination(payload.To)
			features.addForest(payload.FromForestID)
			features.addForest(payload.ToForestID)
			features.addScalar(payload.Count, 25)
			features.addScalar(payload.MaxCount, 25)
			features.addScalar(payload.AlliedWarriors, 25)
			features.addCard(payload.DecreeCardID)
			features.addStringFlag(payload.SourceEffectID)
			if payload.AlliedFaction != 0 || payload.AlliedWarriors > 0 {
				features.setTarget(payload.AlliedFaction)
				features.addFlag(true)
			} else {
				features.addFlag(false)
			}
		}
	case game.ActionBattle:
		if payload := action.Battle; payload != nil {
			features.setActor(payload.Faction)
			features.setTarget(payload.TargetFaction)
			features.addClearing(payload.ClearingID)
			features.addCard(payload.DecreeCardID)
			features.addFlag(payload.UseAlliedFaction)
			features.addStringFlag(payload.SourceEffectID)
		}
	case game.ActionBattleResolution:
		if payload := action.BattleResolution; payload != nil {
			features.setActor(payload.Faction)
			features.setTarget(payload.TargetFaction)
			features.addClearing(payload.ClearingID)
			features.addCard(payload.DecreeCardID)
			features.addRoll(features.attackerRolls[:], payload.AttackerRoll)
			features.addRoll(features.defenderRolls[:], payload.DefenderRoll)
			features.addScalar(payload.AttackerHitModifier, 4)
			features.addScalar(payload.DefenderHitModifier, 4)
			features.addScalar(payload.AmbushHitsToAttacker, 4)
			features.addScalar(payload.AttackerLosses, 8)
			features.addScalar(payload.DefenderLosses, 8)
			features.addScalar(payload.AlliedWarriorLosses, 8)
			features.addScalar(len(payload.AttackerDamagedItemIndexes), 8)
			features.addScalar(len(payload.DefenderDamagedItemIndexes), 8)
			features.addPieceLosses(payload.AttackerPieceLosses)
			features.addPieceLosses(payload.DefenderPieceLosses)
			features.addFlag(payload.IgnoreHitsToAttacker)
			features.addFlag(payload.IgnoreHitsToDefender)
			features.addFlag(payload.DefenderAmbushed)
			features.addFlag(payload.AttackerCounterAmbush)
			features.addFlag(payload.AttackerUsedArmorers)
			features.addFlag(payload.DefenderUsedArmorers)
			features.addFlag(payload.AttackerUsedBrutalTactics)
			features.addFlag(payload.DefenderUsedSappers)
			features.addFlag(payload.UseAlliedFaction)
			features.addStringFlag(payload.SourceEffectID)
		}
	case game.ActionBuild:
		if payload := action.Build; payload != nil {
			features.setActor(payload.Faction)
			features.addClearing(payload.ClearingID)
			features.addBuilding(payload.BuildingType)
			features.addCard(payload.DecreeCardID)
			features.addScalar(len(payload.WoodSources), maxClearings)
			for _, source := range payload.WoodSources {
				features.addSource(source.ClearingID)
				features.addScalar(source.Amount, 8)
			}
		}
	case game.ActionRecruit:
		if payload := action.Recruit; payload != nil {
			features.setActor(payload.Faction)
			features.addClearings(payload.ClearingIDs)
			features.addCard(payload.DecreeCardID)
			features.addScalar(len(payload.ClearingIDs), maxClearings)
		}
	case game.ActionOverwork:
		if payload := action.Overwork; payload != nil {
			features.setActor(payload.Faction)
			features.addClearing(payload.ClearingID)
			features.addCard(payload.CardID)
		}
	case game.ActionCraft:
		if payload := action.Craft; payload != nil {
			features.setActor(payload.Faction)
			features.addCard(payload.CardID)
			features.addClearings(payload.UsedWorkshopClearings)
			features.addScalar(len(payload.DamagedVagabondItemIndexes), 8)
		}
	case game.ActionAddToDecree:
		if payload := action.AddToDecree; payload != nil {
			features.setActor(payload.Faction)
			features.addCards(payload.CardIDs)
			for _, column := range payload.Columns {
				features.addDecreeColumn(column)
			}
			features.addScalar(len(payload.CardIDs), 4)
		}
	case game.ActionSpreadSympathy:
		if payload := action.SpreadSympathy; payload != nil {
			features.setActor(payload.Faction)
			features.addClearing(payload.ClearingID)
			features.addCards(payload.SupporterCardIDs)
		}
	case game.ActionRevolt:
		if payload := action.Revolt; payload != nil {
			features.setActor(payload.Faction)
			features.addClearing(payload.ClearingID)
			features.addCards(payload.SupporterCardIDs)
			features.addSuit(payload.BaseSuit)
			features.addScalar(len(payload.DamagedVagabondItemIndexes), 8)
		}
	case game.ActionMobilize:
		if payload := action.Mobilize; payload != nil {
			features.setActor(payload.Faction)
			features.addCard(payload.CardID)
		}
	case game.ActionTrain:
		if payload := action.Train; payload != nil {
			features.setActor(payload.Faction)
			features.addCard(payload.CardID)
		}
	case game.ActionOrganize:
		if payload := action.Organize; payload != nil {
			features.setActor(payload.Faction)
			features.addClearing(payload.ClearingID)
		}
	case game.ActionExplore:
		if payload := action.Explore; payload != nil {
			features.setActor(payload.Faction)
			features.addClearing(payload.ClearingID)
			features.addItem(payload.ItemType)
		}
	case game.ActionQuest:
		if payload := action.Quest; payload != nil {
			features.setActor(payload.Faction)
			features.addScalar(int(payload.QuestID), 15)
			features.addScalar(len(payload.ItemIndexes), 3)
			features.addScalar(int(payload.Reward), 2)
		}
	case game.ActionAid:
		if payload := action.Aid; payload != nil {
			features.setActor(payload.Faction)
			features.setTarget(payload.TargetFaction)
			features.addClearing(payload.ClearingID)
			features.addCard(payload.CardID)
			features.addScalar(payload.ItemIndex, 12)
			features.addFlag(payload.TakeItemIndex != nil)
			if payload.TakeItemIndex != nil {
				features.addScalar(*payload.TakeItemIndex, 12)
			}
		}
	case game.ActionStrike:
		if payload := action.Strike; payload != nil {
			features.setActor(payload.Faction)
			features.setTarget(payload.TargetFaction)
			features.addClearing(payload.ClearingID)
		}
	case game.ActionRepair:
		if payload := action.Repair; payload != nil {
			features.setActor(payload.Faction)
			features.addScalar(payload.ItemIndex, 12)
		}
	case game.ActionTurmoil:
		if payload := action.Turmoil; payload != nil {
			features.setActor(payload.Faction)
			features.addScalar(int(payload.NewLeader), leaderCount)
		}
	case game.ActionDaybreak:
		if payload := action.Daybreak; payload != nil {
			features.setActor(payload.Faction)
			features.addScalar(len(payload.RefreshedItemIndexes), 12)
		}
	case game.ActionSlip:
		if payload := action.Slip; payload != nil {
			features.setActor(payload.Faction)
			features.addSource(payload.From)
			features.addDestination(payload.To)
			features.addForest(payload.FromForestID)
			features.addForest(payload.ToForestID)
		}
	case game.ActionBirdsongWood:
		if payload := action.BirdsongWood; payload != nil {
			features.setActor(payload.Faction)
			features.addClearings(payload.ClearingIDs)
			features.addScalar(payload.Amount, 8)
		}
	case game.ActionEveningDraw:
		if payload := action.EveningDraw; payload != nil {
			features.setActor(payload.Faction)
			features.addScalar(payload.Count, 6)
		}
	case game.ActionScoreRoosts:
		if payload := action.ScoreRoosts; payload != nil {
			features.setActor(payload.Faction)
			features.addScalar(payload.Points, 10)
		}
	case game.ActionPassPhase:
		if payload := action.PassPhase; payload != nil {
			features.setActor(payload.Faction)
		}
	case game.ActionAddCardToHand:
		if payload := action.AddCardToHand; payload != nil {
			features.setActor(payload.Faction)
			features.addCard(payload.CardID)
		}
	case game.ActionRemoveCardFromHand:
		if payload := action.RemoveCardFromHand; payload != nil {
			features.setActor(payload.Faction)
			features.addCard(payload.CardID)
		}
	case game.ActionOtherPlayerDraw:
		if payload := action.OtherPlayerDraw; payload != nil {
			features.setActor(payload.Faction)
			features.addScalar(payload.Count, 6)
		}
	case game.ActionOtherPlayerPlay:
		if payload := action.OtherPlayerPlay; payload != nil {
			features.setActor(payload.Faction)
			features.addCard(payload.CardID)
		}
	case game.ActionDiscardEffect:
		if payload := action.DiscardEffect; payload != nil {
			features.setActor(payload.Faction)
			features.addCard(payload.CardID)
		}
	case game.ActionActivateDominance:
		if payload := action.ActivateDominance; payload != nil {
			features.setActor(payload.Faction)
			features.setTarget(payload.TargetFaction)
			features.addCard(payload.CardID)
		}
	case game.ActionTakeDominance:
		if payload := action.TakeDominance; payload != nil {
			features.setActor(payload.Faction)
			features.addCard(payload.DominanceCardID)
			features.addCard(payload.SpentCardID)
		}
	case game.ActionMarquiseSetup:
		if payload := action.MarquiseSetup; payload != nil {
			features.setActor(payload.Faction)
			features.addClearing(payload.KeepClearingID)
			features.addClearing(payload.SawmillClearingID)
			features.addClearing(payload.WorkshopClearingID)
			features.addClearing(payload.RecruiterClearingID)
			features.addBuilding(game.Sawmill)
			features.addBuilding(game.Workshop)
			features.addBuilding(game.Recruiter)
		}
	case game.ActionEyrieSetup:
		if payload := action.EyrieSetup; payload != nil {
			features.setActor(payload.Faction)
			features.addClearing(payload.ClearingID)
			features.addScalar(int(payload.Leader), leaderCount)
		}
	case game.ActionVagabondSetup:
		if payload := action.VagabondSetup; payload != nil {
			features.setActor(payload.Faction)
			features.addForest(payload.ForestID)
			features.addScalar(int(payload.Character), characterCount)
		}
	case game.ActionUsePersistentEffect:
		if payload := action.UsePersistentEffect; payload != nil {
			features.setActor(payload.Faction)
			features.setTarget(payload.TargetFaction)
			features.addClearing(payload.ClearingID)
			features.addCard(payload.ObservedCardID)
			features.addStringFlag(payload.EffectID)
		}
	case game.ActionEyrieEmergencyOrders:
		if payload := action.EyrieEmergency; payload != nil {
			features.setActor(payload.Faction)
			features.addScalar(payload.Count, 3)
		}
	case game.ActionEyrieNewRoost:
		if payload := action.EyrieNewRoost; payload != nil {
			features.setActor(payload.Faction)
			features.addClearing(payload.ClearingID)
		}
	case game.ActionVagabondRest:
		if payload := action.VagabondRest; payload != nil {
			features.setActor(payload.Faction)
		}
	case game.ActionVagabondDiscard:
		if payload := action.VagabondDiscard; payload != nil {
			features.setActor(payload.Faction)
			features.addCards(payload.CardIDs)
		}
	case game.ActionVagabondItemCapacity:
		if payload := action.VagabondCapacity; payload != nil {
			features.setActor(payload.Faction)
			features.addScalar(len(payload.ItemIndexes), 12)
		}
	case game.ActionEveningDiscard:
		if payload := action.EveningDiscard; payload != nil {
			features.setActor(payload.Faction)
			features.addCards(payload.CardIDs)
			features.addScalar(payload.Count, 12)
		}
	case game.ActionFieldHospitals:
		if payload := action.FieldHospitals; payload != nil {
			features.setActor(payload.Faction)
			features.addClearing(payload.ClearingID)
			features.addCard(payload.CardID)
			features.addFlag(payload.Decline)
		}
	case game.ActionMarquiseExtraAction:
		if payload := action.MarquiseExtraAction; payload != nil {
			features.setActor(payload.Faction)
			features.addCard(payload.CardID)
		}
	case game.ActionVagabondSteal:
		if payload := action.VagabondSteal; payload != nil {
			features.setActor(payload.Faction)
			features.setTarget(payload.TargetFaction)
			features.addClearing(payload.ClearingID)
			features.addCard(payload.ObservedCardID)
		}
	case game.ActionVagabondDayLabor:
		if payload := action.VagabondDayLabor; payload != nil {
			features.setActor(payload.Faction)
			features.addClearing(payload.ClearingID)
			features.addCard(payload.CardID)
		}
	case game.ActionVagabondHideout:
		if payload := action.VagabondHideout; payload != nil {
			features.setActor(payload.Faction)
			features.addScalar(len(payload.ItemIndexes), 12)
		}
	case game.ActionResolveOutrage:
		if payload := action.ResolveOutrage; payload != nil {
			features.setActor(payload.Faction)
			features.addSuit(payload.Suit)
			features.addCard(payload.CardID)
			features.addFlag(payload.DrawSupporter)
		}
	}

	features.addStateContext(state)
}

func (features *actionFeatures) addStateContext(state game.GameState) {
	features.addScalar(state.RoundNumber, 30)
	features.addScalar(state.VictoryPoints[game.Marquise], 30)
	features.addScalar(state.VictoryPoints[game.Eyrie], 30)
	features.addScalar(state.VictoryPoints[game.Alliance], 30)
	features.addScalar(state.VictoryPoints[game.Vagabond], 30)
}

func (features *actionFeatures) setActor(faction game.Faction) {
	features.actor = faction
	features.actorPresent = true
}

func (features *actionFeatures) setTarget(faction game.Faction) {
	features.target = faction
	features.targetPresent = true
}

func (features *actionFeatures) addClearing(clearingID int) {
	if slot := clearingSlot(clearingID); slot >= 0 {
		features.clearings[slot] = 1
	}
}

func (features *actionFeatures) addClearings(clearingIDs []int) {
	for _, clearingID := range clearingIDs {
		features.addClearing(clearingID)
	}
}

func (features *actionFeatures) addSource(clearingID int) {
	if slot := clearingSlot(clearingID); slot >= 0 {
		features.sources[slot] = 1
		features.addClearing(clearingID)
	}
}

func (features *actionFeatures) addDestination(clearingID int) {
	if slot := clearingSlot(clearingID); slot >= 0 {
		features.destinations[slot] = 1
		features.addClearing(clearingID)
	}
}

func (features *actionFeatures) addForest(forestID int) {
	if slot := forestSlot(forestID); slot >= 0 {
		features.forests[slot] = 1
	}
}

func (features *actionFeatures) addCard(cardID game.CardID) {
	if slot := cardSlot(cardID); slot >= 0 {
		features.cards[slot] = 1
	}
	card, ok := cardByID[cardID]
	if !ok {
		return
	}
	features.addSuit(card.Suit)
	if int(card.Kind) >= 0 && int(card.Kind) < cardKindCount {
		features.cardKinds[card.Kind] += 1
	}
}

func (features *actionFeatures) addCards(cardIDs []game.CardID) {
	for _, cardID := range cardIDs {
		features.addCard(cardID)
	}
}

func (features *actionFeatures) addSuit(suit game.Suit) {
	if int(suit) >= 0 && int(suit) < suitCount {
		features.cardSuits[suit] += 1
	}
}

func (features *actionFeatures) addItem(itemType game.ItemType) {
	if slot := itemTypeIndex(itemType); slot >= 0 {
		features.items[slot] += 1
	}
}

func (features *actionFeatures) addBuilding(buildingType game.BuildingType) {
	if slot := buildingTypeIndex(buildingType); slot >= 0 {
		features.buildings[slot] += 1
	}
}

func (features *actionFeatures) addDecreeColumn(column game.DecreeColumn) {
	if int(column) >= 0 && int(column) < decreeColumnCount {
		features.decreeColumns[column] += 1
	}
}

func (features *actionFeatures) addPieceLosses(losses []game.BattlePieceLoss) {
	for _, loss := range losses {
		switch loss.Kind {
		case game.BattlePieceBuilding:
			features.addBuilding(loss.BuildingType)
		case game.BattlePieceToken:
			if slot := tokenTypeIndex(loss.TokenType); slot >= 0 {
				features.addScalar(slot+1, tokenTypeCount)
			}
		case game.BattlePieceWood:
			features.addFlag(true)
		}
	}
}

func (features *actionFeatures) addRoll(target []float32, roll int) {
	if roll >= 0 && roll < len(target) {
		target[roll] = 1
	}
}

func (features *actionFeatures) addScalar(value int, scale float32) {
	if features.nextScalarIndex >= len(features.scalars) {
		return
	}
	if scale <= 0 {
		features.scalars[features.nextScalarIndex] = 0
	} else {
		features.scalars[features.nextScalarIndex] = float32(value) / scale
	}
	features.nextScalarIndex++
}

func (features *actionFeatures) addFlag(value bool) {
	if features.nextFlagIndex >= len(features.flags) {
		return
	}
	if value {
		features.flags[features.nextFlagIndex] = 1
	}
	features.nextFlagIndex++
}

func (features *actionFeatures) addStringFlag(value string) {
	features.addFlag(value != "")
}
