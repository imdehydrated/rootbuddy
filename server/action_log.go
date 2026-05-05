package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/imdehydrated/rootbuddy/game"
)

const maxActionLogEntries = 50

type ActionLogEntry struct {
	RoundNumber int             `json:"roundNumber"`
	Faction     game.Faction    `json:"faction"`
	ActionType  game.ActionType `json:"actionType"`
	Summary     string          `json:"summary"`
	Timestamp   int64           `json:"timestamp"`
}

type actionLogStore struct {
	mu     sync.RWMutex
	byGame map[string][]ActionLogEntry
}

func newActionLogStore() *actionLogStore {
	return &actionLogStore{
		byGame: map[string][]ActionLogEntry{},
	}
}

var actionLogs = newActionLogStore()

func (s *actionLogStore) ensureGame(gameID string) {
	if gameID == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byGame[gameID]; !ok {
		s.byGame[gameID] = []ActionLogEntry{}
	}
}

func (s *actionLogStore) append(gameID string, entry ActionLogEntry) []ActionLogEntry {
	if gameID == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entries := append([]ActionLogEntry(nil), s.byGame[gameID]...)
	entries = append(entries, entry)
	if len(entries) > maxActionLogEntries {
		entries = append([]ActionLogEntry(nil), entries[len(entries)-maxActionLogEntries:]...)
	}
	s.byGame[gameID] = entries
	return append([]ActionLogEntry(nil), entries...)
}

func (s *actionLogStore) get(gameID string) []ActionLogEntry {
	if gameID == "" {
		return []ActionLogEntry{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := s.byGame[gameID]
	if len(entries) == 0 {
		return []ActionLogEntry{}
	}
	return append([]ActionLogEntry(nil), entries...)
}

func (s *actionLogStore) delete(gameID string) {
	if gameID == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byGame, gameID)
}

func newActionLogEntry(roundNumber int, faction game.Faction, action game.Action) ActionLogEntry {
	return ActionLogEntry{
		RoundNumber: roundNumber,
		Faction:     faction,
		ActionType:  action.Type,
		Summary:     summarizeAction(action),
		Timestamp:   time.Now().Unix(),
	}
}

func summarizeAction(action game.Action) string {
	switch action.Type {
	case game.ActionMovement:
		return fmt.Sprintf("Move from clearing %d to clearing %d", action.Movement.From, action.Movement.To)
	case game.ActionBattle:
		return fmt.Sprintf("Battle %s in clearing %d", factionLabel(action.Battle.TargetFaction), action.Battle.ClearingID)
	case game.ActionBattleResolution:
		return fmt.Sprintf("Resolve battle in clearing %d", action.BattleResolution.ClearingID)
	case game.ActionBuild:
		return fmt.Sprintf("Build %s in clearing %d", buildingLabel(action.Build.BuildingType), action.Build.ClearingID)
	case game.ActionRecruit:
		return fmt.Sprintf("Recruit in clearings %v", action.Recruit.ClearingIDs)
	case game.ActionOverwork:
		return fmt.Sprintf("Overwork in clearing %d", action.Overwork.ClearingID)
	case game.ActionCraft:
		return fmt.Sprintf("Craft card %d", action.Craft.CardID)
	case game.ActionAddToDecree:
		return fmt.Sprintf("Add decree cards %v", action.AddToDecree.CardIDs)
	case game.ActionSpreadSympathy:
		return fmt.Sprintf("Spread sympathy to clearing %d", action.SpreadSympathy.ClearingID)
	case game.ActionRevolt:
		return fmt.Sprintf("Revolt in clearing %d", action.Revolt.ClearingID)
	case game.ActionMobilize:
		return fmt.Sprintf("Mobilize card %d", action.Mobilize.CardID)
	case game.ActionTrain:
		return fmt.Sprintf("Train with card %d", action.Train.CardID)
	case game.ActionOrganize:
		return fmt.Sprintf("Organize in clearing %d", action.Organize.ClearingID)
	case game.ActionExplore:
		return fmt.Sprintf("Explore clearing %d", action.Explore.ClearingID)
	case game.ActionQuest:
		return fmt.Sprintf("Complete quest %d", action.Quest.QuestID)
	case game.ActionAid:
		return fmt.Sprintf("Aid %s in clearing %d with item %d", factionLabel(action.Aid.TargetFaction), action.Aid.ClearingID, action.Aid.ItemIndex)
	case game.ActionStrike:
		return fmt.Sprintf("Strike %s in clearing %d", factionLabel(action.Strike.TargetFaction), action.Strike.ClearingID)
	case game.ActionRepair:
		return fmt.Sprintf("Repair item %d", action.Repair.ItemIndex)
	case game.ActionTurmoil:
		return "Go into turmoil"
	case game.ActionDaybreak:
		return fmt.Sprintf("Refresh %d item(s)", len(action.Daybreak.RefreshedItemIndexes))
	case game.ActionSlip:
		if action.Slip.From > 0 && action.Slip.To == action.Slip.From {
			return fmt.Sprintf("Stay in clearing %d", action.Slip.From)
		}
		if action.Slip.FromForestID > 0 && action.Slip.ToForestID == action.Slip.FromForestID {
			return fmt.Sprintf("Stay in forest %d", action.Slip.FromForestID)
		}
		if action.Slip.ToForestID > 0 {
			return fmt.Sprintf("Slip to forest %d", action.Slip.ToForestID)
		}
		return fmt.Sprintf("Slip to clearing %d", action.Slip.To)
	case game.ActionBirdsongWood:
		return fmt.Sprintf("Place wood in clearings %v", action.BirdsongWood.ClearingIDs)
	case game.ActionEveningDraw:
		return fmt.Sprintf("Draw %d card(s)", action.EveningDraw.Count)
	case game.ActionScoreRoosts:
		return fmt.Sprintf("Score %d roost point(s)", action.ScoreRoosts.Points)
	case game.ActionPassPhase:
		return "Pass phase"
	case game.ActionAddCardToHand:
		return fmt.Sprintf("Add card %d to hand", action.AddCardToHand.CardID)
	case game.ActionRemoveCardFromHand:
		return fmt.Sprintf("Remove card %d from hand", action.RemoveCardFromHand.CardID)
	case game.ActionOtherPlayerDraw:
		return fmt.Sprintf("Record %s drawing %d card(s)", factionLabel(action.OtherPlayerDraw.Faction), action.OtherPlayerDraw.Count)
	case game.ActionOtherPlayerPlay:
		return fmt.Sprintf("Record %s playing card %d", factionLabel(action.OtherPlayerPlay.Faction), action.OtherPlayerPlay.CardID)
	case game.ActionDiscardEffect:
		return fmt.Sprintf("Discard effect card %d", action.DiscardEffect.CardID)
	case game.ActionActivateDominance:
		return fmt.Sprintf("Activate dominance card %d", action.ActivateDominance.CardID)
	case game.ActionTakeDominance:
		return fmt.Sprintf("Take dominance card %d", action.TakeDominance.DominanceCardID)
	case game.ActionMarquiseSetup:
		return fmt.Sprintf(
			"Marquise setup: keep %d, sawmill %d, workshop %d, recruiter %d",
			action.MarquiseSetup.KeepClearingID,
			action.MarquiseSetup.SawmillClearingID,
			action.MarquiseSetup.WorkshopClearingID,
			action.MarquiseSetup.RecruiterClearingID,
		)
	case game.ActionEyrieSetup:
		return fmt.Sprintf("Eyrie setup: leader %d, start in clearing %d", action.EyrieSetup.Leader, action.EyrieSetup.ClearingID)
	case game.ActionVagabondSetup:
		return fmt.Sprintf("Vagabond setup: character %d, start in forest %d", action.VagabondSetup.Character, action.VagabondSetup.ForestID)
	case game.ActionUsePersistentEffect:
		switch action.UsePersistentEffect.EffectID {
		case "better_burrow_bank":
			return fmt.Sprintf("Use Better Burrow Bank with %s", factionLabel(action.UsePersistentEffect.TargetFaction))
		case "codebreakers":
			return fmt.Sprintf("Use Codebreakers on %s", factionLabel(action.UsePersistentEffect.TargetFaction))
		case "royal_claim":
			return "Use Royal Claim"
		case "stand_and_deliver":
			return fmt.Sprintf("Use Stand and Deliver! on %s", factionLabel(action.UsePersistentEffect.TargetFaction))
		case "tax_collector":
			return fmt.Sprintf("Use Tax Collector in clearing %d", action.UsePersistentEffect.ClearingID)
		default:
			return fmt.Sprintf("Use persistent effect %s", action.UsePersistentEffect.EffectID)
		}
	case game.ActionFieldHospitals:
		if action.FieldHospitals.Decline {
			return fmt.Sprintf("Decline Field Hospitals for clearing %d", action.FieldHospitals.ClearingID)
		}
		return fmt.Sprintf("Use Field Hospitals for clearing %d with card %d", action.FieldHospitals.ClearingID, action.FieldHospitals.CardID)
	case game.ActionMarquiseExtraAction:
		return fmt.Sprintf("Spend bird card %d for an extra Marquise action", action.MarquiseExtraAction.CardID)
	default:
		return actionTypeLabel(action.Type)
	}
}

func factionLabel(faction game.Faction) string {
	switch faction {
	case game.Marquise:
		return "Marquise"
	case game.Alliance:
		return "Woodland Alliance"
	case game.Eyrie:
		return "Eyrie"
	case game.Vagabond:
		return "Vagabond"
	default:
		return fmt.Sprintf("Faction %d", faction)
	}
}

func buildingLabel(buildingType game.BuildingType) string {
	switch buildingType {
	case game.Sawmill:
		return "Sawmill"
	case game.Workshop:
		return "Workshop"
	case game.Recruiter:
		return "Recruiter"
	case game.Roost:
		return "Roost"
	case game.Base:
		return "Base"
	default:
		return fmt.Sprintf("Building %d", buildingType)
	}
}

func actionTypeLabel(actionType game.ActionType) string {
	switch actionType {
	case game.ActionMovement:
		return "Movement"
	case game.ActionBattle:
		return "Battle"
	case game.ActionBattleResolution:
		return "Battle Resolution"
	case game.ActionBuild:
		return "Build"
	case game.ActionRecruit:
		return "Recruit"
	case game.ActionOverwork:
		return "Overwork"
	case game.ActionCraft:
		return "Craft"
	case game.ActionAddToDecree:
		return "Add To Decree"
	case game.ActionSpreadSympathy:
		return "Spread Sympathy"
	case game.ActionRevolt:
		return "Revolt"
	case game.ActionMobilize:
		return "Mobilize"
	case game.ActionTrain:
		return "Train"
	case game.ActionOrganize:
		return "Organize"
	case game.ActionExplore:
		return "Explore"
	case game.ActionQuest:
		return "Quest"
	case game.ActionAid:
		return "Aid"
	case game.ActionStrike:
		return "Strike"
	case game.ActionRepair:
		return "Repair"
	case game.ActionTurmoil:
		return "Turmoil"
	case game.ActionDaybreak:
		return "Daybreak"
	case game.ActionSlip:
		return "Slip"
	case game.ActionBirdsongWood:
		return "Birdsong Wood"
	case game.ActionEveningDraw:
		return "Evening Draw"
	case game.ActionScoreRoosts:
		return "Score Roosts"
	case game.ActionPassPhase:
		return "Pass Phase"
	case game.ActionAddCardToHand:
		return "Add Card To Hand"
	case game.ActionRemoveCardFromHand:
		return "Remove Card From Hand"
	case game.ActionOtherPlayerDraw:
		return "Other Player Draw"
	case game.ActionOtherPlayerPlay:
		return "Other Player Play"
	case game.ActionDiscardEffect:
		return "Discard Effect"
	case game.ActionActivateDominance:
		return "Activate Dominance"
	case game.ActionTakeDominance:
		return "Take Dominance"
	case game.ActionMarquiseSetup:
		return "Marquise Setup"
	case game.ActionEyrieSetup:
		return "Eyrie Setup"
	case game.ActionVagabondSetup:
		return "Vagabond Setup"
	case game.ActionUsePersistentEffect:
		return "Use Persistent Effect"
	case game.ActionFieldHospitals:
		return "Field Hospitals"
	case game.ActionMarquiseExtraAction:
		return "Marquise Extra Action"
	default:
		return fmt.Sprintf("Action %d", actionType)
	}
}
