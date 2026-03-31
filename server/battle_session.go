package server

import (
	"errors"
	"sync"
	"time"

	"github.com/imdehydrated/rootbuddy/game"
)

var (
	errBattleSessionNotFound         = errors.New("battle session not found")
	errBattleSessionPendingResponse  = errors.New("battle is awaiting defender response")
	errBattleSessionStale            = errors.New("battle session is stale")
	errBattleResponseNotAvailable    = errors.New("battle response is not available")
	errBattleResponseAlreadyProvided = errors.New("battle response has already been provided")
	errBattleResponseForbidden       = errors.New("battle response is not available for this player")
)

type BattlePromptStage string

const (
	BattlePromptWaitingDefender BattlePromptStage = "waiting_defender"
	BattlePromptDefenderTurn    BattlePromptStage = "defender_response"
	BattlePromptReadyToResolve  BattlePromptStage = "ready_to_resolve"
)

type BattlePrompt struct {
	GameID               string             `json:"gameID"`
	Revision             int64              `json:"revision"`
	Action               game.Action        `json:"action"`
	Stage                BattlePromptStage  `json:"stage"`
	WaitingOnFaction     game.Faction       `json:"waitingOnFaction"`
	BattleContext        game.BattleContext `json:"battleContext"`
	CanUseAmbush         bool               `json:"canUseAmbush,omitempty"`
	CanUseArmorers       bool               `json:"canUseArmorers,omitempty"`
	CanUseSappers        bool               `json:"canUseSappers,omitempty"`
	DefenderAmbush       bool               `json:"defenderAmbush,omitempty"`
	DefenderUsedArmorers bool               `json:"defenderUsedArmorers,omitempty"`
	DefenderUsedSappers  bool               `json:"defenderUsedSappers,omitempty"`
}

type battleSession struct {
	GameID               string
	Revision             int64
	Action               game.Action
	BattleContext        game.BattleContext
	AttackerFaction      game.Faction
	DefenderFaction      game.Faction
	CreatedAt            time.Time
	DefenderCanAmbush    bool
	DefenderCanArmorers  bool
	DefenderCanSappers   bool
	DefenderResponded    bool
	DefenderAmbush       bool
	DefenderUsedArmorers bool
	DefenderUsedSappers  bool
}

type battleSessionStore struct {
	mu     sync.RWMutex
	byGame map[string]battleSession
}

func newBattleSessionStore() *battleSessionStore {
	return &battleSessionStore{
		byGame: map[string]battleSession{},
	}
}

var battleSessions = newBattleSessionStore()

func (s *battleSessionStore) open(gameID string, revision int64, action game.Action, context game.BattleContext) (battleSession, bool) {
	session := battleSession{
		GameID:              gameID,
		Revision:            revision,
		Action:              action,
		BattleContext:       context,
		AttackerFaction:     action.Battle.Faction,
		DefenderFaction:     action.Battle.TargetFaction,
		CreatedAt:           time.Now().UTC(),
		DefenderCanAmbush:   context.CanDefenderAmbush,
		DefenderCanArmorers: context.CanDefenderArmorers,
		DefenderCanSappers:  context.CanDefenderSappers,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.byGame[gameID]
	if ok && battleSessionsMatch(existing, session) {
		return existing, false
	}

	s.byGame[gameID] = session
	return session, true
}

func (s *battleSessionStore) get(gameID string) (battleSession, bool) {
	if gameID == "" {
		return battleSession{}, false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.byGame[gameID]
	return session, ok
}

func (s *battleSessionStore) clear(gameID string) {
	if gameID == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byGame, gameID)
}

func (s *battleSessionStore) applyDefenderResponse(gameID string, revision int64, perspective game.Faction, req BattleResponseRequest) (battleSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.byGame[gameID]
	if !ok {
		return battleSession{}, errBattleSessionNotFound
	}
	if session.Revision != revision {
		return battleSession{}, errBattleSessionStale
	}
	if perspective != session.DefenderFaction {
		return battleSession{}, errBattleResponseForbidden
	}
	if session.DefenderResponded {
		return battleSession{}, errBattleResponseAlreadyProvided
	}
	if !requiresDefenderResponse(session.BattleContext) {
		return battleSession{}, errBattleResponseNotAvailable
	}

	if req.UseAmbush != nil {
		if !session.DefenderCanAmbush && *req.UseAmbush {
			return battleSession{}, errBattleResponseNotAvailable
		}
		session.DefenderAmbush = session.DefenderCanAmbush && *req.UseAmbush
	}
	if req.UseArmorers != nil {
		if !session.DefenderCanArmorers && *req.UseArmorers {
			return battleSession{}, errBattleResponseNotAvailable
		}
		session.DefenderUsedArmorers = session.DefenderCanArmorers && *req.UseArmorers
	}
	if req.UseSappers != nil {
		if !session.DefenderCanSappers && *req.UseSappers {
			return battleSession{}, errBattleResponseNotAvailable
		}
		session.DefenderUsedSappers = session.DefenderCanSappers && *req.UseSappers
	}

	session.DefenderResponded = true
	s.byGame[gameID] = session
	return session, nil
}

func requiresDefenderResponse(context game.BattleContext) bool {
	return context.CanDefenderAmbush || context.CanDefenderArmorers || context.CanDefenderSappers
}

func battleSessionsMatch(left battleSession, right battleSession) bool {
	if left.GameID != right.GameID || left.Revision != right.Revision {
		return false
	}
	return battleActionsMatch(left.Action, right.Action)
}

func battleActionsMatch(left game.Action, right game.Action) bool {
	if left.Type != game.ActionBattle || right.Type != game.ActionBattle {
		return false
	}
	if left.Battle == nil || right.Battle == nil {
		return false
	}
	return left.Battle.Faction == right.Battle.Faction &&
		left.Battle.ClearingID == right.Battle.ClearingID &&
		left.Battle.TargetFaction == right.Battle.TargetFaction &&
		left.Battle.DecreeCardID == right.Battle.DecreeCardID &&
		left.Battle.SourceEffectID == right.Battle.SourceEffectID
}

func battlePromptView(session battleSession, perspective game.Faction) *BattlePrompt {
	if session.GameID == "" {
		return nil
	}

	stage := BattlePromptReadyToResolve
	waitingOn := session.AttackerFaction
	if !session.DefenderResponded {
		stage = BattlePromptWaitingDefender
		waitingOn = session.DefenderFaction
		if perspective == session.DefenderFaction {
			stage = BattlePromptDefenderTurn
		}
	}

	prompt := &BattlePrompt{
		GameID:           session.GameID,
		Revision:         session.Revision,
		Action:           session.Action,
		Stage:            stage,
		WaitingOnFaction: waitingOn,
		BattleContext:    redactBattleContextForPerspective(session.BattleContext, perspective),
	}

	if perspective == session.DefenderFaction && !session.DefenderResponded {
		prompt.CanUseAmbush = session.DefenderCanAmbush
		prompt.CanUseArmorers = session.DefenderCanArmorers
		prompt.CanUseSappers = session.DefenderCanSappers
	}
	if session.DefenderResponded {
		prompt.DefenderAmbush = session.DefenderAmbush
		prompt.DefenderUsedArmorers = session.DefenderUsedArmorers
		prompt.DefenderUsedSappers = session.DefenderUsedSappers
	}

	return prompt
}

func redactBattleContextForPerspective(context game.BattleContext, perspective game.Faction) game.BattleContext {
	redacted := context
	if context.Action.Battle == nil {
		return redacted
	}

	attacker := context.Action.Battle.Faction
	defender := context.Action.Battle.TargetFaction

	if perspective != defender {
		redacted.CanDefenderAmbush = false
	}
	if perspective != attacker {
		redacted.CanAttackerCounterAmbush = false
	}

	return redacted
}
