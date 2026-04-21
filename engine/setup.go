package engine

import (
	"errors"

	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
	"github.com/imdehydrated/rootbuddy/mapdata"
)

type SetupRequest struct {
	GameMode      game.GameMode
	PlayerFaction game.Faction
	TrackAllHands bool
	Factions      []game.Faction
	MapID         game.MapID
	RandomSeed    int64
}

var baseQuestRegistry = buildQuestRegistry()

func buildQuestRegistry() map[game.QuestID]game.Quest {
	registry := make(map[game.QuestID]game.Quest, len(carddata.QuestDeck()))
	for _, quest := range carddata.QuestDeck() {
		registry[quest.ID] = quest
	}
	return registry
}

func questByID(id game.QuestID) (game.Quest, bool) {
	quest, ok := baseQuestRegistry[id]
	return quest, ok
}

func SetupGame(req SetupRequest) (game.GameState, error) {
	if len(req.Factions) < 2 || len(req.Factions) > 4 {
		return game.GameState{}, errors.New("setup requires between 2 and 4 factions")
	}
	if req.MapID != game.AutumnMapID {
		return game.GameState{}, errors.New("unsupported map")
	}

	present := map[game.Faction]bool{}
	for _, faction := range req.Factions {
		if present[faction] {
			return game.GameState{}, errors.New("setup factions must be unique")
		}
		present[faction] = true
	}
	if !present[req.PlayerFaction] {
		return game.GameState{}, errors.New("player faction must be included in setup factions")
	}

	state := game.GameState{
		Map:               mapdata.AutumnMap(),
		GameMode:          req.GameMode,
		RandomSeed:        req.RandomSeed,
		GamePhase:         game.LifecycleSetup,
		SetupStage:        game.SetupStageUnspecified,
		PlayerFaction:     req.PlayerFaction,
		TrackAllHands:     req.TrackAllHands,
		RoundNumber:       1,
		CurrentStep:       game.StepUnspecified,
		VictoryPoints:     make(map[game.Faction]int, len(req.Factions)),
		ItemSupply:        InitialItemSupply(),
		PersistentEffects: map[game.Faction][]game.CardID{},
		OtherHandCounts:   map[game.Faction]int{},
		HiddenCards:       []game.HiddenCard{},
		NextHiddenCardID:  1,
	}
	state.TurnOrder = randomizedTurnOrder(req.Factions, &state)

	if len(state.TurnOrder) == 0 {
		return game.GameState{}, errors.New("setup produced no turn order")
	}

	populateRuins(&state)
	initializeMarquiseSetupState(&state, present)
	initializeEyrieSetupState(&state, present)
	initializeAllianceSetupState(&state, present)
	initializeVagabondSetupState(&state, present)

	for _, faction := range state.TurnOrder {
		state.VictoryPoints[faction] = 0
	}
	setupDeckAndHands(&state, present)
	advanceSetupStage(&state)

	return state, nil
}

func randomizedTurnOrder(factions []game.Faction, state *game.GameState) []game.Faction {
	present := map[game.Faction]bool{}
	for _, faction := range factions {
		present[faction] = true
	}

	order := make([]game.Faction, 0, len(factions))
	for _, faction := range defaultTurnOrder {
		if present[faction] {
			order = append(order, faction)
		}
	}
	rng := nextShuffleRNG(state)
	rng.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})
	return order
}

func findClearing(state *game.GameState, clearingID int) *game.Clearing {
	for i := range state.Map.Clearings {
		if state.Map.Clearings[i].ID == clearingID {
			return &state.Map.Clearings[i]
		}
	}
	return nil
}

func ensureWarriors(clearing *game.Clearing) {
	if clearing.Warriors == nil {
		clearing.Warriors = map[game.Faction]int{}
	}
}

func placeWarriors(state *game.GameState, faction game.Faction, clearingID int, count int) {
	clearing := findClearing(state, clearingID)
	if clearing == nil {
		return
	}
	ensureWarriors(clearing)
	clearing.Warriors[faction] += count
}

func placeBuilding(state *game.GameState, faction game.Faction, clearingID int, buildingType game.BuildingType) {
	clearing := findClearing(state, clearingID)
	if clearing == nil {
		return
	}
	clearing.Buildings = append(clearing.Buildings, game.Building{
		Faction: faction,
		Type:    buildingType,
	})
}

func placeToken(state *game.GameState, faction game.Faction, clearingID int, tokenType game.TokenType) {
	clearing := findClearing(state, clearingID)
	if clearing == nil {
		return
	}
	clearing.Tokens = append(clearing.Tokens, game.Token{
		Faction: faction,
		Type:    tokenType,
	})
}

func populateRuins(state *game.GameState) {
	ruinClearings := []*game.Clearing{}
	for i := range state.Map.Clearings {
		if state.Map.Clearings[i].Ruins {
			ruinClearings = append(ruinClearings, &state.Map.Clearings[i])
		}
	}

	ruinItems := append([]game.ItemType(nil), RuinItems()...)
	rng := nextShuffleRNG(state)
	rng.Shuffle(len(ruinItems), func(i, j int) {
		ruinItems[i], ruinItems[j] = ruinItems[j], ruinItems[i]
	})

	for i, clearing := range ruinClearings {
		if i >= len(ruinItems) {
			clearing.RuinItems = nil
			continue
		}
		clearing.RuinItems = []game.ItemType{ruinItems[i]}
	}
}

func initializeMarquiseSetupState(state *game.GameState, present map[game.Faction]bool) {
	if !present[game.Marquise] {
		return
	}

	state.Marquise.WarriorSupply = 25
}

func initializeEyrieSetupState(state *game.GameState, present map[game.Faction]bool) {
	if !present[game.Eyrie] {
		return
	}

	state.Eyrie.WarriorSupply = 20
	state.Eyrie.RoostsPlaced = 0
	state.Eyrie.Leader = -1
	state.Eyrie.AvailableLeaders = []game.EyrieLeader{
		game.LeaderBuilder,
		game.LeaderCharismatic,
		game.LeaderCommander,
		game.LeaderDespot,
	}
}

func initializeAllianceSetupState(state *game.GameState, present map[game.Faction]bool) {
	if !present[game.Alliance] {
		return
	}

	state.Alliance.WarriorSupply = 10
	state.Alliance.Officers = 0
	state.Alliance.SympathyPlaced = 0
}

func initializeVagabondSetupState(state *game.GameState, present map[game.Faction]bool) {
	if !present[game.Vagabond] {
		return
	}

	state.Vagabond.Character = -1
	state.Vagabond.Relationships = map[game.Faction]game.RelationshipLevel{}
	for _, faction := range state.TurnOrder {
		if faction == game.Vagabond {
			continue
		}
		state.Vagabond.Relationships[faction] = game.RelIndifferent
	}

	questDeck := carddata.QuestDeck()
	state.QuestDeck = shuffleQuestIDs(state, questDeck)
	drawSetupQuests(state, 3)
}

func factionPresentInTurnOrder(state game.GameState, faction game.Faction) bool {
	for _, turnFaction := range state.TurnOrder {
		if turnFaction == faction {
			return true
		}
	}
	return false
}

func advanceSetupStage(state *game.GameState) {
	if state.GamePhase != game.LifecycleSetup {
		return
	}

	switch {
	case factionPresentInTurnOrder(*state, game.Marquise) && state.Marquise.KeepClearingID == 0:
		state.SetupStage = game.SetupStageMarquise
		state.FactionTurn = game.Marquise
	case factionPresentInTurnOrder(*state, game.Eyrie) && state.Eyrie.RoostsPlaced == 0:
		state.SetupStage = game.SetupStageEyrie
		state.FactionTurn = game.Eyrie
	case factionPresentInTurnOrder(*state, game.Vagabond) && !state.Vagabond.InForest && state.Vagabond.ForestID == 0:
		state.SetupStage = game.SetupStageVagabond
		state.FactionTurn = game.Vagabond
	default:
		finalizeSetup(state)
	}
}

func finalizeSetup(state *game.GameState) {
	state.SetupStage = game.SetupStageComplete
	state.GamePhase = game.LifecyclePlaying
	if len(state.TurnOrder) > 0 {
		state.FactionTurn = state.TurnOrder[0]
	}
	state.CurrentPhase = game.Birdsong
	state.CurrentStep = game.StepBirdsong
}

func shuffleQuestIDs(state *game.GameState, quests []game.Quest) []game.QuestID {
	ids := make([]game.QuestID, 0, len(quests))
	for _, quest := range quests {
		ids = append(ids, quest.ID)
	}

	rng := nextShuffleRNG(state)
	rng.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})
	return ids
}

func drawSetupQuests(state *game.GameState, count int) {
	for i := 0; i < count && len(state.QuestDeck) > 0; i++ {
		questID := state.QuestDeck[0]
		state.QuestDeck = state.QuestDeck[1:]
		quest, ok := questByID(questID)
		if !ok {
			continue
		}
		state.Vagabond.QuestsAvailable = append(state.Vagabond.QuestsAvailable, quest)
	}
}

func setupDeckAndHands(state *game.GameState, present map[game.Faction]bool) {
	baseCards := carddata.BaseDeck()
	if len(state.TurnOrder) == 2 {
		filtered := make([]game.Card, 0, len(baseCards)-4)
		for _, card := range baseCards {
			if card.Kind == game.DominanceCard {
				continue
			}
			filtered = append(filtered, card)
		}
		baseCards = filtered
	}

	if state.GameMode == game.GameModeOnline {
		state.Deck = ShuffleDeck(state, BuildDeck(baseCards))
	}

	for _, faction := range state.TurnOrder {
		if faction == state.PlayerFaction {
			if state.GameMode == game.GameModeOnline {
				DrawCards(state, faction, 3)
			}
			continue
		}
		if state.GameMode == game.GameModeOnline {
			DrawCards(state, faction, 3)
		} else {
			for i := 0; i < 3; i++ {
				addHiddenCard(state, faction, game.HiddenCardZoneHand, 0)
			}
		}
	}

	if present[game.Alliance] {
		setupAllianceSupporters(state)
	}
}

func setupAllianceSupporters(state *game.GameState) {
	switch state.GameMode {
	case game.GameModeOnline:
		if tracksHandForFaction(*state, game.Alliance) {
			state.Alliance.Supporters = drawCardsToSlice(state, 3)
			return
		}
		consumeDeckCards(state, 3)
		for i := 0; i < 3; i++ {
			addHiddenCard(state, game.Alliance, game.HiddenCardZoneSupporters, 0)
		}
	case game.GameModeAssist:
		for i := 0; i < 3; i++ {
			addHiddenCard(state, game.Alliance, game.HiddenCardZoneSupporters, 0)
		}
	}
}

func consumeDeckCards(state *game.GameState, count int) {
	for i := 0; i < count; i++ {
		if _, ok := drawOneCardID(state); !ok {
			return
		}
	}
}

func drawCardsToSlice(state *game.GameState, count int) []game.Card {
	if state.GameMode != game.GameModeOnline || count <= 0 {
		return nil
	}

	drawn := make([]game.Card, 0, count)
	for i := 0; i < count; i++ {
		cardID, ok := drawOneCardID(state)
		if !ok {
			break
		}

		card, found := CardByID(cardID)
		if !found {
			continue
		}
		drawn = append(drawn, card)
	}

	return drawn
}
