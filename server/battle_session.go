package server

import (
	"errors"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/imdehydrated/rootbuddy/game"
)

var (
	errBattleSessionNotFound         = errors.New("battle session not found")
	errBattleSessionPendingResponse  = errors.New("battle is awaiting defender response")
	errBattleSessionPendingAttacker  = errors.New("battle is awaiting attacker response")
	errBattleSessionStale            = errors.New("battle session is stale")
	errBattleResponseNotAvailable    = errors.New("battle response is not available")
	errBattleResponseAlreadyProvided = errors.New("battle response has already been provided")
	errBattleResponseForbidden       = errors.New("battle response is not available for this player")
	errBattleResolutionNotReady      = errors.New("battle resolution is not ready")
	errBattleResolutionMismatch      = errors.New("battle resolution does not match authoritative result")
)

type BattlePromptStage string

const (
	BattlePromptWaitingDefender BattlePromptStage = "waiting_defender"
	BattlePromptDefenderTurn    BattlePromptStage = "defender_response"
	BattlePromptWaitingAttacker BattlePromptStage = "waiting_attacker"
	BattlePromptAttackerTurn    BattlePromptStage = "attacker_response"
	BattlePromptReadyToResolve  BattlePromptStage = "ready_to_resolve"
)

type battleResponseKind int

const (
	battleResponseNone battleResponseKind = iota
	battleResponseDefenderAmbush
	battleResponseAttackerCounterAmbush
	battleResponseDefenderEffects
	battleResponseAttackerEffects
)

type BattlePrompt struct {
	GameID                    string             `json:"gameID"`
	Revision                  int64              `json:"revision"`
	Action                    game.Action        `json:"action"`
	Stage                     BattlePromptStage  `json:"stage"`
	WaitingOnFaction          game.Faction       `json:"waitingOnFaction"`
	BattleContext             game.BattleContext `json:"battleContext"`
	AttackerRoll              int                `json:"attackerRoll,omitempty"`
	DefenderRoll              int                `json:"defenderRoll,omitempty"`
	CanUseAmbush              bool               `json:"canUseAmbush,omitempty"`
	CanUseDefenderArmorers    bool               `json:"canUseDefenderArmorers,omitempty"`
	CanUseSappers             bool               `json:"canUseSappers,omitempty"`
	CanUseCounterAmbush       bool               `json:"canUseCounterAmbush,omitempty"`
	CanUseAttackerArmorers    bool               `json:"canUseAttackerArmorers,omitempty"`
	CanUseBrutalTactics       bool               `json:"canUseBrutalTactics,omitempty"`
	DefenderAmbush            bool               `json:"defenderAmbush,omitempty"`
	DefenderUsedArmorers      bool               `json:"defenderUsedArmorers,omitempty"`
	DefenderUsedSappers       bool               `json:"defenderUsedSappers,omitempty"`
	AttackerCounterAmbush     bool               `json:"attackerCounterAmbush,omitempty"`
	AttackerUsedArmorers      bool               `json:"attackerUsedArmorers,omitempty"`
	AttackerUsedBrutalTactics bool               `json:"attackerUsedBrutalTactics,omitempty"`
}

type battleSession struct {
	GameID                    string
	Revision                  int64
	Action                    game.Action
	BattleContext             game.BattleContext
	AttackerFaction           game.Faction
	DefenderFaction           game.Faction
	CreatedAt                 time.Time
	DefenderCanAmbush         bool
	DefenderCanArmorers       bool
	DefenderCanSappers        bool
	DefenderAmbushResponded   bool
	DefenderAmbush            bool
	DefenderEffectsResponded  bool
	DefenderUsedArmorers      bool
	DefenderUsedSappers       bool
	AttackerCanCounterAmbush  bool
	AttackerCanArmorers       bool
	AttackerCanBrutalTactics  bool
	AttackerCounterResponded  bool
	AttackerCounterAmbush     bool
	AttackerEffectsResponded  bool
	AttackerUsedArmorers      bool
	AttackerUsedBrutalTactics bool
	RollsResolved             bool
	AttackerRoll              int
	DefenderRoll              int
	ResolvedAction            *game.Action
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
		GameID:                   gameID,
		Revision:                 revision,
		Action:                   action,
		BattleContext:            context,
		AttackerFaction:          action.Battle.Faction,
		DefenderFaction:          action.Battle.TargetFaction,
		CreatedAt:                time.Now().UTC(),
		DefenderCanAmbush:        context.CanDefenderAmbush,
		DefenderCanArmorers:      context.CanDefenderArmorers,
		DefenderCanSappers:       context.CanDefenderSappers,
		AttackerCanCounterAmbush: context.CanAttackerCounterAmbush,
		AttackerCanArmorers:      context.CanAttackerArmorers,
		AttackerCanBrutalTactics: context.CanAttackerBrutalTactics,
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

func (s *battleSessionStore) clearIfPresent(gameID string) bool {
	if gameID == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byGame[gameID]; !ok {
		return false
	}
	delete(s.byGame, gameID)
	return true
}

func (s *battleSessionStore) recordResolution(gameID string, revision int64, action game.Action, context game.BattleContext, resolved game.Action) (battleSession, error) {
	if gameID == "" {
		return battleSession{}, errBattleSessionNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.byGame[gameID]
	if ok {
		if session.Revision != revision {
			return battleSession{}, errBattleSessionStale
		}
		if !battleActionsMatch(session.Action, action) {
			return battleSession{}, errBattleSessionStale
		}
	} else {
		session = battleSession{
			GameID:          gameID,
			Revision:        revision,
			Action:          action,
			BattleContext:   context,
			AttackerFaction: action.Battle.Faction,
			DefenderFaction: action.Battle.TargetFaction,
			CreatedAt:       time.Now().UTC(),
		}
	}

	resolvedCopy := resolved
	session.ResolvedAction = &resolvedCopy
	s.byGame[gameID] = session
	return session, nil
}

func (s *battleSessionStore) resolvedActionForApply(gameID string, revision int64, action game.Action) (game.Action, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.byGame[gameID]
	if !ok {
		return game.Action{}, errBattleResolutionNotReady
	}
	if session.Revision != revision {
		return game.Action{}, errBattleSessionStale
	}
	if session.ResolvedAction == nil {
		return game.Action{}, errBattleResolutionNotReady
	}
	if !reflect.DeepEqual(*session.ResolvedAction, action) {
		return game.Action{}, errBattleResolutionMismatch
	}

	return *session.ResolvedAction, nil
}

func (s *battleSessionStore) applyResponse(gameID string, revision int64, perspective game.Faction, req BattleResponseRequest) (battleSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.byGame[gameID]
	if !ok {
		return battleSession{}, errBattleSessionNotFound
	}
	if session.Revision != revision {
		return battleSession{}, errBattleSessionStale
	}
	responseKind := currentBattleResponseKind(session)
	if perspective != currentBattleResponder(session) {
		return battleSession{}, errBattleResponseForbidden
	}
	switch responseKind {
	case battleResponseDefenderAmbush:
		if session.DefenderAmbushResponded {
			return battleSession{}, errBattleResponseAlreadyProvided
		}
		if req.UseAmbush != nil {
			if !session.DefenderCanAmbush && *req.UseAmbush {
				return battleSession{}, errBattleResponseNotAvailable
			}
			session.DefenderAmbush = session.DefenderCanAmbush && *req.UseAmbush
		}
		session.DefenderAmbushResponded = true
	case battleResponseDefenderEffects:
		if session.DefenderEffectsResponded {
			return battleSession{}, errBattleResponseAlreadyProvided
		}
		if req.UseDefenderArmorers != nil {
			if !session.DefenderCanArmorers && *req.UseDefenderArmorers {
				return battleSession{}, errBattleResponseNotAvailable
			}
			session.DefenderUsedArmorers = session.DefenderCanArmorers && *req.UseDefenderArmorers
		}
		if req.UseSappers != nil {
			if !session.DefenderCanSappers && *req.UseSappers {
				return battleSession{}, errBattleResponseNotAvailable
			}
			session.DefenderUsedSappers = session.DefenderCanSappers && *req.UseSappers
		}
		session.DefenderEffectsResponded = true
	case battleResponseAttackerCounterAmbush:
		if session.AttackerCounterResponded {
			return battleSession{}, errBattleResponseAlreadyProvided
		}
		if req.UseCounterAmbush != nil {
			if !canSessionUseCounterAmbush(session) && *req.UseCounterAmbush {
				return battleSession{}, errBattleResponseNotAvailable
			}
			session.AttackerCounterAmbush = canSessionUseCounterAmbush(session) && *req.UseCounterAmbush
		}
		session.AttackerCounterResponded = true
	case battleResponseAttackerEffects:
		if session.AttackerEffectsResponded {
			return battleSession{}, errBattleResponseAlreadyProvided
		}
		if req.UseAttackerArmorers != nil {
			if !session.AttackerCanArmorers && *req.UseAttackerArmorers {
				return battleSession{}, errBattleResponseNotAvailable
			}
			session.AttackerUsedArmorers = session.AttackerCanArmorers && *req.UseAttackerArmorers
		}
		if req.UseBrutalTactics != nil {
			if !session.AttackerCanBrutalTactics && *req.UseBrutalTactics {
				return battleSession{}, errBattleResponseNotAvailable
			}
			session.AttackerUsedBrutalTactics = session.AttackerCanBrutalTactics && *req.UseBrutalTactics
		}
		session.AttackerEffectsResponded = true
	default:
		return battleSession{}, errBattleResponseNotAvailable
	}

	s.byGame[gameID] = session
	return session, nil
}

func (s *battleSessionStore) storeRolls(gameID string, revision int64, attackerRoll int, defenderRoll int) (battleSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.byGame[gameID]
	if !ok {
		return battleSession{}, errBattleSessionNotFound
	}
	if session.Revision != revision {
		return battleSession{}, errBattleSessionStale
	}
	if session.RollsResolved {
		return session, nil
	}

	session.AttackerRoll = attackerRoll
	session.DefenderRoll = defenderRoll
	session.RollsResolved = true
	s.byGame[gameID] = session
	return session, nil
}

func requiresBattleSession(context game.BattleContext) bool {
	return context.CanDefenderAmbush ||
		context.CanAttackerCounterAmbush ||
		context.CanDefenderArmorers ||
		context.CanDefenderSappers ||
		context.CanAttackerArmorers ||
		context.CanAttackerBrutalTactics
}

func requiresDefenderAmbushResponse(context game.BattleContext) bool {
	return context.CanDefenderAmbush
}

func requiresPostRollResponse(context game.BattleContext) bool {
	return context.CanDefenderArmorers ||
		context.CanDefenderSappers ||
		context.CanAttackerArmorers ||
		context.CanAttackerBrutalTactics
}

func requiresAttackerBaseResponse(context game.BattleContext) bool {
	return context.CanAttackerArmorers || context.CanAttackerBrutalTactics
}

func requiresDefenderEffectsResponse(session battleSession) bool {
	return session.RollsResolved && (session.DefenderCanArmorers || session.DefenderCanSappers)
}

func requiresAttackerEffectsResponse(session battleSession) bool {
	return session.RollsResolved && (session.AttackerCanArmorers || session.AttackerCanBrutalTactics)
}

func canSessionUseCounterAmbush(session battleSession) bool {
	return session.DefenderAmbush && session.AttackerCanCounterAmbush
}

func currentBattleResponseKind(session battleSession) battleResponseKind {
	if !session.DefenderAmbushResponded && requiresDefenderAmbushResponse(session.BattleContext) {
		return battleResponseDefenderAmbush
	}
	if !session.AttackerCounterResponded && canSessionUseCounterAmbush(session) {
		return battleResponseAttackerCounterAmbush
	}
	if !session.RollsResolved {
		return battleResponseNone
	}
	if !session.DefenderEffectsResponded && requiresDefenderEffectsResponse(session) {
		return battleResponseDefenderEffects
	}
	if !session.AttackerEffectsResponded && requiresAttackerEffectsResponse(session) {
		return battleResponseAttackerEffects
	}
	return battleResponseNone
}

func currentBattleResponder(session battleSession) game.Faction {
	switch currentBattleResponseKind(session) {
	case battleResponseDefenderAmbush, battleResponseDefenderEffects:
		return session.DefenderFaction
	case battleResponseAttackerCounterAmbush, battleResponseAttackerEffects:
		return session.AttackerFaction
	default:
		return 0
	}
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
	responseKind := currentBattleResponseKind(session)
	if responseKind == battleResponseDefenderAmbush || responseKind == battleResponseDefenderEffects {
		stage = BattlePromptWaitingDefender
		waitingOn = session.DefenderFaction
		if perspective == session.DefenderFaction {
			stage = BattlePromptDefenderTurn
		}
	} else if responseKind == battleResponseAttackerCounterAmbush || responseKind == battleResponseAttackerEffects {
		stage = BattlePromptWaitingAttacker
		waitingOn = session.AttackerFaction
		if perspective == session.AttackerFaction {
			stage = BattlePromptAttackerTurn
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
	if session.RollsResolved {
		prompt.AttackerRoll = session.AttackerRoll
		prompt.DefenderRoll = session.DefenderRoll
	}

	if perspective == session.DefenderFaction && responseKind == battleResponseDefenderAmbush {
		prompt.CanUseAmbush = session.DefenderCanAmbush
	}
	if perspective == session.DefenderFaction && responseKind == battleResponseDefenderEffects {
		prompt.CanUseDefenderArmorers = session.DefenderCanArmorers
		prompt.CanUseSappers = session.DefenderCanSappers
	}
	if perspective == session.AttackerFaction && responseKind == battleResponseAttackerCounterAmbush {
		prompt.CanUseCounterAmbush = canSessionUseCounterAmbush(session)
	}
	if perspective == session.AttackerFaction && responseKind == battleResponseAttackerEffects {
		prompt.CanUseAttackerArmorers = session.AttackerCanArmorers
		prompt.CanUseBrutalTactics = session.AttackerCanBrutalTactics
	}
	if session.DefenderAmbushResponded {
		prompt.DefenderAmbush = session.DefenderAmbush
	}
	if session.DefenderEffectsResponded {
		prompt.DefenderUsedArmorers = session.DefenderUsedArmorers
		prompt.DefenderUsedSappers = session.DefenderUsedSappers
	}
	if session.AttackerCounterResponded {
		prompt.AttackerCounterAmbush = session.AttackerCounterAmbush
	}
	if session.AttackerEffectsResponded {
		prompt.AttackerUsedArmorers = session.AttackerUsedArmorers
		prompt.AttackerUsedBrutalTactics = session.AttackerUsedBrutalTactics
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
		redacted.CanDefenderArmorers = false
		redacted.CanDefenderSappers = false
	}
	if perspective != attacker {
		redacted.CanAttackerCounterAmbush = false
		redacted.CanAttackerArmorers = false
		redacted.CanAttackerBrutalTactics = false
	}

	return redacted
}

func pendingBattleSessionError(gameID string, revision int64, context game.BattleContext) ErrorResponse {
	err := errBattleSessionPendingResponse
	if !requiresDefenderAmbushResponse(context) && requiresAttackerBaseResponse(context) {
		err = errBattleSessionPendingAttacker
	}

	return ErrorResponse{
		Error:    err.Error(),
		GameID:   gameID,
		Revision: revision,
	}
}

func battleSessionErrorStatus(err error) int {
	switch {
	case errors.Is(err, errBattleSessionNotFound):
		return http.StatusNotFound
	case errors.Is(err, errBattleSessionPendingResponse), errors.Is(err, errBattleSessionPendingAttacker), errors.Is(err, errBattleSessionStale), errors.Is(err, errBattleResponseAlreadyProvided), errors.Is(err, errBattleResolutionNotReady):
		return http.StatusConflict
	case errors.Is(err, errBattleResponseForbidden):
		return http.StatusForbidden
	case errors.Is(err, errBattleResolutionMismatch):
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}
