package rl

import (
	"github.com/imdehydrated/rootbuddy/carddata"
	"github.com/imdehydrated/rootbuddy/game"
)

const (
	maxClearings = 12
	maxForests   = 7

	factionCount       = 4
	suitCount          = 4
	itemTypeCount      = 8
	itemStatusCount    = 3
	buildingTypeCount  = 5
	tokenTypeCount     = 2
	phaseCount         = 3
	stepCount          = 5
	lifecycleCount     = 3
	setupStageCount    = 5
	gameModeCount      = 2
	leaderCount        = 4
	characterCount     = 3
	relationshipCount  = 5
	decreeColumnCount  = 4
	cardKindCount      = 5
	baseDeckCardCount  = 54
	specialCardIDSlots = 2

	actionTypeCount = int(game.ActionResolveOutrage) + 1
)

var (
	orderedFactions = []game.Faction{
		game.Marquise,
		game.Eyrie,
		game.Alliance,
		game.Vagabond,
	}
	orderedBuildingTypes = []game.BuildingType{
		game.Sawmill,
		game.Workshop,
		game.Recruiter,
		game.Roost,
		game.Base,
	}
	orderedTokenTypes = []game.TokenType{
		game.TokenKeep,
		game.TokenSympathy,
	}
	orderedItemTypes = []game.ItemType{
		game.ItemTea,
		game.ItemCoin,
		game.ItemCrossbow,
		game.ItemHammer,
		game.ItemSword,
		game.ItemTorch,
		game.ItemBoots,
		game.ItemBag,
	}
	orderedItemStatuses = []game.ItemStatus{
		game.ItemReady,
		game.ItemExhausted,
		game.ItemDamaged,
	}
	cardByID = buildCardLookup()
)

func ObservationVectorLength() int {
	return len(EncodeObservationZero())
}

func ActionVectorLength() int {
	return len(EncodeAction(game.GameState{}, game.Action{}))
}

func EncodeObservationZero() []float32 {
	return encodeObservationState(observationInput{})
}

func buildCardLookup() map[game.CardID]game.Card {
	cards := carddata.BaseDeck()
	lookup := make(map[game.CardID]game.Card, len(cards))
	for _, card := range cards {
		lookup[card.ID] = card
	}
	return lookup
}

func appendBool(out []float32, value bool) []float32 {
	if value {
		return append(out, 1)
	}
	return append(out, 0)
}

func appendNorm(out []float32, value int, scale float32) []float32 {
	if scale <= 0 {
		return append(out, 0)
	}
	return append(out, float32(value)/scale)
}

func appendOneHot(out []float32, index int, count int) []float32 {
	for i := 0; i < count; i++ {
		if i == index {
			out = append(out, 1)
		} else {
			out = append(out, 0)
		}
	}
	return out
}

func appendFactionOneHot(out []float32, faction game.Faction) []float32 {
	return appendOneHot(out, factionIndex(faction), factionCount)
}

func appendOptionalFactionOneHot(out []float32, faction game.Faction, present bool) []float32 {
	if !present {
		return appendOneHot(out, 0, factionCount+1)
	}
	return appendOneHot(out, factionIndex(faction)+1, factionCount+1)
}

func factionIndex(faction game.Faction) int {
	for index, candidate := range orderedFactions {
		if candidate == faction {
			return index
		}
	}
	return -1
}

func boolMapFaction(values map[game.Faction]bool, faction game.Faction) bool {
	if values == nil {
		return false
	}
	return values[faction]
}

func clearingSlot(clearingID int) int {
	if clearingID < 1 || clearingID > maxClearings {
		return -1
	}
	return clearingID - 1
}

func forestSlot(forestID int) int {
	if forestID < 1 || forestID > maxForests {
		return -1
	}
	return forestID - 1
}

func itemTypeIndex(itemType game.ItemType) int {
	for index, candidate := range orderedItemTypes {
		if candidate == itemType {
			return index
		}
	}
	return -1
}

func buildingTypeIndex(buildingType game.BuildingType) int {
	for index, candidate := range orderedBuildingTypes {
		if candidate == buildingType {
			return index
		}
	}
	return -1
}

func tokenTypeIndex(tokenType game.TokenType) int {
	for index, candidate := range orderedTokenTypes {
		if candidate == tokenType {
			return index
		}
	}
	return -1
}

func cardSlot(cardID game.CardID) int {
	switch cardID {
	case game.LoyalVizier1:
		return baseDeckCardCount
	case game.LoyalVizier2:
		return baseDeckCardCount + 1
	default:
		if cardID < 1 || cardID > baseDeckCardCount {
			return -1
		}
		return int(cardID) - 1
	}
}
