package engine

import (
	"errors"
	"math/rand"
	"time"

	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
	"github.com/imdehydrated/rootbuddy/mapdata"
)

type SetupRequest struct {
	GameMode          game.GameMode
	PlayerFaction     game.Faction
	Factions          []game.Faction
	MapID             game.MapID
	VagabondCharacter game.VagabondCharacter
	EyrieLeader       game.EyrieLeader
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
		GamePhase:         game.LifecyclePlaying,
		PlayerFaction:     req.PlayerFaction,
		RoundNumber:       1,
		CurrentPhase:      game.Birdsong,
		CurrentStep:       game.StepBirdsong,
		TurnOrder:         filteredTurnOrder(req.Factions),
		VictoryPoints:     make(map[game.Faction]int, len(req.Factions)),
		ItemSupply:        InitialItemSupply(),
		PersistentEffects: map[game.Faction][]game.CardID{},
		OtherHandCounts:   map[game.Faction]int{},
	}

	if len(state.TurnOrder) == 0 {
		return game.GameState{}, errors.New("setup produced no turn order")
	}
	state.FactionTurn = state.TurnOrder[0]

	populateRuins(&state)
	setupMarquise(&state, present)
	setupEyrie(&state, present, req.EyrieLeader)
	setupAlliance(&state, present)
	setupVagabond(&state, present, req.VagabondCharacter)
	setupDeckAndHands(&state, present)

	for _, faction := range state.TurnOrder {
		state.VictoryPoints[faction] = 0
	}

	return state, nil
}

func filteredTurnOrder(factions []game.Faction) []game.Faction {
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
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
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

func setupMarquise(state *game.GameState, present map[game.Faction]bool) {
	if !present[game.Marquise] {
		return
	}

	state.Marquise.WarriorSupply = 14
	state.Marquise.SawmillsPlaced = 1
	state.Marquise.WorkshopsPlaced = 1
	state.Marquise.RecruitersPlaced = 1
	state.Marquise.KeepClearingID = 1

	placeToken(state, game.Marquise, 1, game.TokenKeep)
	for _, clearing := range state.Map.Clearings {
		if clearing.ID == 8 {
			continue
		}
		placeWarriors(state, game.Marquise, clearing.ID, 1)
	}

	placeBuilding(state, game.Marquise, 1, game.Sawmill)
	placeBuilding(state, game.Marquise, 2, game.Workshop)
	placeBuilding(state, game.Marquise, 10, game.Recruiter)
}

func availableCorner(exclude ...int) int {
	blocked := map[int]bool{}
	for _, id := range exclude {
		if id != 0 {
			blocked[id] = true
		}
	}

	for _, corner := range []int{8, 1, 3, 7} {
		if !blocked[corner] {
			return corner
		}
	}
	return 1
}

func setupEyrie(state *game.GameState, present map[game.Faction]bool, leader game.EyrieLeader) {
	if !present[game.Eyrie] {
		return
	}

	startClearing := availableCorner(state.Marquise.KeepClearingID)
	if state.Marquise.KeepClearingID == 1 {
		startClearing = 8
	}

	state.Eyrie.WarriorSupply = 14
	state.Eyrie.RoostsPlaced = 1
	state.Eyrie.Leader = leader
	state.Eyrie.AvailableLeaders = []game.EyrieLeader{
		game.LeaderBuilder,
		game.LeaderCharismatic,
		game.LeaderCommander,
		game.LeaderDespot,
	}
	state.Eyrie.AvailableLeaders = removeLeader(state.Eyrie.AvailableLeaders, leader)

	placeBuilding(state, game.Eyrie, startClearing, game.Roost)
	placeWarriors(state, game.Eyrie, startClearing, 6)

	vizierColumns := vizierColumnsForLeader(leader)
	appendCardToDecree(&state.Eyrie.Decree, vizierColumns[0], game.LoyalVizier1)
	appendCardToDecree(&state.Eyrie.Decree, vizierColumns[1], game.LoyalVizier2)
}

func setupAlliance(state *game.GameState, present map[game.Faction]bool) {
	if !present[game.Alliance] {
		return
	}

	state.Alliance.WarriorSupply = 10
	state.Alliance.Officers = 0
	state.Alliance.SympathyPlaced = 0
}

func setupVagabond(state *game.GameState, present map[game.Faction]bool, character game.VagabondCharacter) {
	if !present[game.Vagabond] {
		return
	}

	state.Vagabond.Character = character
	state.Vagabond.ForestID = 7
	state.Vagabond.InForest = true
	state.Vagabond.Items = VagabondStartingItems(character)
	state.Vagabond.Relationships = map[game.Faction]game.RelationshipLevel{}
	for _, faction := range state.TurnOrder {
		if faction == game.Vagabond {
			continue
		}
		state.Vagabond.Relationships[faction] = game.RelIndifferent
	}

	questDeck := carddata.QuestDeck()
	state.QuestDeck = shuffleQuestIDs(questDeck)
	drawSetupQuests(state, 3)
}

func shuffleQuestIDs(quests []game.Quest) []game.QuestID {
	ids := make([]game.QuestID, 0, len(quests))
	for _, quest := range quests {
		ids = append(ids, quest.ID)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
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
		state.Deck = ShuffleDeck(BuildDeck(baseCards))
	}

	for _, faction := range state.TurnOrder {
		if faction == game.Alliance {
			if state.GameMode == game.GameModeOnline {
				if state.PlayerFaction == game.Alliance {
					state.Alliance.Supporters = DrawCards(state, game.Alliance, 3)
				} else {
					consumeDeckCards(state, 3)
				}
			}
			continue
		}
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
			state.OtherHandCounts[faction] = 3
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
