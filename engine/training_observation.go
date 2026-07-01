package engine

import "github.com/imdehydrated/rootbuddy/game"

type TrainingObservationOptions struct {
	Perspective game.Faction
	Omniscient  bool
}

type TrainingHiddenCounts struct {
	Hands              map[game.Faction]int `json:"hands,omitempty"`
	AllianceSupporters int                  `json:"allianceSupporters,omitempty"`
	Deck               int                  `json:"deck,omitempty"`
	QuestDeck          int                  `json:"questDeck,omitempty"`
}

type TrainingObservation struct {
	Perspective game.Faction         `json:"perspective"`
	Omniscient  bool                 `json:"omniscient,omitempty"`
	State       TrainingPublicState  `json:"state"`
	Hidden      TrainingHiddenCounts `json:"hidden,omitempty"`
	DebugState  *game.GameState      `json:"debugState,omitempty"`
}

type TrainingPublicState struct {
	Map                   game.Map                         `json:"map"`
	GameMode              game.GameMode                    `json:"gameMode"`
	GamePhase             game.GameLifecycle               `json:"gamePhase"`
	SetupStage            game.SetupStage                  `json:"setupStage"`
	PlayerFaction         game.Faction                     `json:"playerFaction"`
	Winner                game.Faction                     `json:"winner"`
	WinningCoalition      []game.Faction                   `json:"winningCoalition,omitempty"`
	RoundNumber           int                              `json:"roundNumber"`
	FactionTurn           game.Faction                     `json:"factionTurn"`
	CurrentPhase          game.Phase                       `json:"currentPhase"`
	CurrentStep           game.TurnStep                    `json:"currentStep"`
	TurnOrder             []game.Faction                   `json:"turnOrder,omitempty"`
	VictoryPoints         map[game.Faction]int             `json:"victoryPoints,omitempty"`
	DiscardPile           []game.CardID                    `json:"discardPile,omitempty"`
	AvailableDominance    []game.CardID                    `json:"availableDominance,omitempty"`
	ActiveDominance       map[game.Faction]game.CardID     `json:"activeDominance,omitempty"`
	CoalitionActive       bool                             `json:"coalitionActive,omitempty"`
	CoalitionPartner      game.Faction                     `json:"coalitionPartner"`
	ItemSupply            map[game.ItemType]int            `json:"itemSupply,omitempty"`
	CraftedItems          map[game.Faction][]game.ItemType `json:"craftedItems,omitempty"`
	PersistentEffects     map[game.Faction][]game.CardID   `json:"persistentEffects,omitempty"`
	QuestDiscard          []game.QuestID                   `json:"questDiscard,omitempty"`
	OtherHandCounts       map[game.Faction]int             `json:"otherHandCounts,omitempty"`
	PendingFieldHospitals []game.FieldHospitalsPending     `json:"pendingFieldHospitals,omitempty"`
	PendingOutrage        []game.OutragePending            `json:"pendingOutrage,omitempty"`
	Marquise              game.MarquiseState               `json:"marquise"`
	Eyrie                 game.EyrieState                  `json:"eyrie"`
	Alliance              game.AllianceState               `json:"alliance"`
	Vagabond              game.VagabondState               `json:"vagabond"`
	TurnProgress          game.TurnProgress                `json:"turnProgress"`
}

func NewTrainingObservation(state game.GameState, options TrainingObservationOptions) TrainingObservation {
	hidden := trainingHiddenCounts(state, options.Perspective)
	public := trainingPublicGameState(state, options.Perspective, hidden.Hands)
	observation := TrainingObservation{
		Perspective: options.Perspective,
		Omniscient:  options.Omniscient,
		State:       trainingPublicStateFromGameState(public),
		Hidden:      hidden,
	}
	if options.Omniscient {
		debug := CloneState(state)
		observation.DebugState = &debug
	}
	return observation
}

func trainingPublicGameState(state game.GameState, perspective game.Faction, hiddenHands map[game.Faction]int) game.GameState {
	public := CloneState(state)
	public.TrackAllHands = false
	public.PlayerFaction = perspective
	public.RandomSeed = 0
	public.ShuffleCount = 0
	public.BattleRollCount = 0
	public.Deck = nil
	public.QuestDeck = nil
	public.HiddenCards = nil
	public.NextHiddenCardID = 0
	public.OtherHandCounts = copyTrainingHandCounts(hiddenHands)

	for _, faction := range trainingObservationFactions(state) {
		if faction == perspective {
			continue
		}
		clearTrainingFactionHand(&public, faction)
	}
	if perspective != game.Alliance {
		public.Alliance.Supporters = nil
	}

	return public
}

func trainingPublicStateFromGameState(state game.GameState) TrainingPublicState {
	return TrainingPublicState{
		Map:                   state.Map,
		GameMode:              state.GameMode,
		GamePhase:             state.GamePhase,
		SetupStage:            state.SetupStage,
		PlayerFaction:         state.PlayerFaction,
		Winner:                state.Winner,
		WinningCoalition:      state.WinningCoalition,
		RoundNumber:           state.RoundNumber,
		FactionTurn:           state.FactionTurn,
		CurrentPhase:          state.CurrentPhase,
		CurrentStep:           state.CurrentStep,
		TurnOrder:             state.TurnOrder,
		VictoryPoints:         state.VictoryPoints,
		DiscardPile:           state.DiscardPile,
		AvailableDominance:    state.AvailableDominance,
		ActiveDominance:       state.ActiveDominance,
		CoalitionActive:       state.CoalitionActive,
		CoalitionPartner:      state.CoalitionPartner,
		ItemSupply:            state.ItemSupply,
		CraftedItems:          state.CraftedItems,
		PersistentEffects:     state.PersistentEffects,
		QuestDiscard:          state.QuestDiscard,
		OtherHandCounts:       state.OtherHandCounts,
		PendingFieldHospitals: state.PendingFieldHospitals,
		PendingOutrage:        state.PendingOutrage,
		Marquise:              state.Marquise,
		Eyrie:                 state.Eyrie,
		Alliance:              state.Alliance,
		Vagabond:              state.Vagabond,
		TurnProgress:          state.TurnProgress,
	}
}

func trainingHiddenCounts(state game.GameState, perspective game.Faction) TrainingHiddenCounts {
	handCounts := map[game.Faction]int{}
	for _, faction := range trainingObservationFactions(state) {
		if faction == perspective {
			continue
		}
		if count := trainingFactionHandCount(state, faction); count > 0 {
			handCounts[faction] = count
		}
	}
	if len(handCounts) == 0 {
		handCounts = nil
	}

	return TrainingHiddenCounts{
		Hands:              handCounts,
		AllianceSupporters: trainingAllianceSupporterCount(state, perspective),
		Deck:               len(state.Deck),
		QuestDeck:          len(state.QuestDeck),
	}
}

func trainingObservationFactions(state game.GameState) []game.Faction {
	if len(state.TurnOrder) > 0 {
		factions := make([]game.Faction, len(state.TurnOrder))
		copy(factions, state.TurnOrder)
		return factions
	}
	return []game.Faction{game.Marquise, game.Eyrie, game.Alliance, game.Vagabond}
}

func trainingFactionHandCount(state game.GameState, faction game.Faction) int {
	if count := len(trainingFactionHand(state, faction)); count > 0 {
		return count
	}
	if count := state.OtherHandCounts[faction]; count > 0 {
		return count
	}
	count := 0
	for _, hidden := range state.HiddenCards {
		if hidden.OwnerFaction == faction && hidden.Zone == game.HiddenCardZoneHand {
			count++
		}
	}
	return count
}

func trainingFactionHand(state game.GameState, faction game.Faction) []game.Card {
	switch faction {
	case game.Marquise:
		return state.Marquise.CardsInHand
	case game.Eyrie:
		return state.Eyrie.CardsInHand
	case game.Alliance:
		return state.Alliance.CardsInHand
	case game.Vagabond:
		return state.Vagabond.CardsInHand
	default:
		return nil
	}
}

func trainingAllianceSupporterCount(state game.GameState, perspective game.Faction) int {
	if perspective == game.Alliance {
		return 0
	}
	if count := len(state.Alliance.Supporters); count > 0 {
		return count
	}
	count := 0
	for _, hidden := range state.HiddenCards {
		if hidden.OwnerFaction == game.Alliance && hidden.Zone == game.HiddenCardZoneSupporters {
			count++
		}
	}
	return count
}

func clearTrainingFactionHand(state *game.GameState, faction game.Faction) {
	switch faction {
	case game.Marquise:
		state.Marquise.CardsInHand = nil
	case game.Eyrie:
		state.Eyrie.CardsInHand = nil
	case game.Alliance:
		state.Alliance.CardsInHand = nil
	case game.Vagabond:
		state.Vagabond.CardsInHand = nil
	}
}

func copyTrainingHandCounts(source map[game.Faction]int) map[game.Faction]int {
	if len(source) == 0 {
		return nil
	}
	cloned := make(map[game.Faction]int, len(source))
	for faction, count := range source {
		cloned[faction] = count
	}
	return cloned
}
