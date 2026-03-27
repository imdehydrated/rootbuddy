package engine

import "github.com/imdehydrated/rootbuddy/game"

func appendHiddenCard(state *game.GameState, faction game.Faction, zone game.HiddenCardZone, knownCardID game.CardID) int {
	if state.NextHiddenCardID <= 0 {
		state.NextHiddenCardID = 1
	}

	id := state.NextHiddenCardID
	state.NextHiddenCardID++
	state.HiddenCards = append(state.HiddenCards, game.HiddenCard{
		ID:           id,
		OwnerFaction: faction,
		Zone:         zone,
		KnownCardID:  knownCardID,
	})
	return id
}

func materializeAssistHandPlaceholders(state *game.GameState) {
	if state.GameMode != game.GameModeAssist || len(state.OtherHandCounts) == 0 {
		return
	}

	existing := map[game.Faction]int{}
	for _, hidden := range state.HiddenCards {
		if hidden.Zone != game.HiddenCardZoneHand || hidden.OwnerFaction == state.PlayerFaction {
			continue
		}
		existing[hidden.OwnerFaction]++
	}

	for faction, count := range state.OtherHandCounts {
		if faction == state.PlayerFaction || count <= existing[faction] {
			continue
		}
		for i := existing[faction]; i < count; i++ {
			appendHiddenCard(state, faction, game.HiddenCardZoneHand, 0)
		}
	}
}

func addHiddenCard(state *game.GameState, faction game.Faction, zone game.HiddenCardZone, knownCardID game.CardID) int {
	materializeAssistHandPlaceholders(state)
	id := appendHiddenCard(state, faction, zone, knownCardID)
	syncOtherHandCountsFromHiddenCards(state)
	return id
}

func hiddenCardCount(state game.GameState, faction game.Faction, zone game.HiddenCardZone) int {
	count := 0
	for _, hidden := range state.HiddenCards {
		if hidden.OwnerFaction == faction && hidden.Zone == zone {
			count++
		}
	}
	return count
}

func consumeHiddenCard(state *game.GameState, faction game.Faction, zone game.HiddenCardZone) bool {
	materializeAssistHandPlaceholders(state)
	for index, hidden := range state.HiddenCards {
		if hidden.OwnerFaction != faction || hidden.Zone != zone {
			continue
		}
		state.HiddenCards = append(state.HiddenCards[:index], state.HiddenCards[index+1:]...)
		syncOtherHandCountsFromHiddenCards(state)
		return true
	}
	return false
}

func consumeHiddenCards(state *game.GameState, faction game.Faction, zone game.HiddenCardZone, count int) int {
	consumed := 0
	for consumed < count {
		if !consumeHiddenCard(state, faction, zone) {
			break
		}
		consumed++
	}
	return consumed
}

func moveHiddenCard(state *game.GameState, faction game.Faction, fromZone game.HiddenCardZone, toZone game.HiddenCardZone) bool {
	materializeAssistHandPlaceholders(state)
	for index := range state.HiddenCards {
		if state.HiddenCards[index].OwnerFaction != faction || state.HiddenCards[index].Zone != fromZone {
			continue
		}
		state.HiddenCards[index].Zone = toZone
		syncOtherHandCountsFromHiddenCards(state)
		return true
	}
	return false
}

func syncOtherHandCountsFromHiddenCards(state *game.GameState) {
	if state.GameMode != game.GameModeAssist {
		return
	}

	counts := map[game.Faction]int{}
	for _, hidden := range state.HiddenCards {
		if hidden.Zone != game.HiddenCardZoneHand || hidden.OwnerFaction == state.PlayerFaction {
			continue
		}
		counts[hidden.OwnerFaction]++
	}
	state.OtherHandCounts = counts
}
