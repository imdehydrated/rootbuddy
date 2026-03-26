package engine

import (
	"github.com/imdehydrated/rootbuddy/game"
	"github.com/imdehydrated/rootbuddy/rules"
)

func effectActions(state game.GameState) []game.Action {
	actions := []game.Action{}
	actions = append(actions, birdsongEffectActions(state)...)
	actions = append(actions, daylightEffectActions(state)...)
	actions = append(actions, eveningEffectActions(state)...)
	return actions
}

func birdsongEffectActions(state game.GameState) []game.Action {
	if state.CurrentPhase != game.Birdsong || state.TurnProgress.BirdsongMainActionTaken {
		return nil
	}

	faction := state.FactionTurn
	actions := []game.Action{}

	if hasPersistentEffect(state, faction, "better_burrow_bank") && !persistentEffectUsedThisTurn(state, "better_burrow_bank") {
		for _, targetFaction := range otherFactionsInTurnOrder(state, faction) {
			actions = append(actions, game.Action{
				Type: game.ActionUsePersistentEffect,
				UsePersistentEffect: &game.UsePersistentEffectAction{
					Faction:       faction,
					EffectID:      "better_burrow_bank",
					TargetFaction: targetFaction,
				},
			})
		}
	}

	if hasPersistentEffect(state, faction, "royal_claim") {
		actions = append(actions, game.Action{
			Type: game.ActionUsePersistentEffect,
			UsePersistentEffect: &game.UsePersistentEffectAction{
				Faction:  faction,
				EffectID: "royal_claim",
			},
		})
	}

	if hasPersistentEffect(state, faction, "stand_and_deliver") && !persistentEffectUsedThisTurn(state, "stand_and_deliver") {
		for _, targetFaction := range otherFactionsInTurnOrder(state, faction) {
			if factionHandSize(state, targetFaction) == 0 {
				continue
			}
			actions = append(actions, game.Action{
				Type: game.ActionUsePersistentEffect,
				UsePersistentEffect: &game.UsePersistentEffectAction{
					Faction:       faction,
					EffectID:      "stand_and_deliver",
					TargetFaction: targetFaction,
				},
			})
		}
	}

	return actions
}

func daylightEffectActions(state game.GameState) []game.Action {
	if state.CurrentPhase != game.Daylight {
		return nil
	}

	faction := state.FactionTurn
	actions := []game.Action{}

	if !state.TurnProgress.DaylightMainActionTaken &&
		hasPersistentEffect(state, faction, "command_warren") &&
		!persistentEffectUsedThisTurn(state, "command_warren") {
		for _, action := range commandWarrenBattles(state) {
			actions = append(actions, action)
		}
	}

	if hasPersistentEffect(state, faction, "tax_collector") && !persistentEffectUsedThisTurn(state, "tax_collector") {
		for _, clearingID := range factionsWithWarriors(state, faction) {
			actions = append(actions, game.Action{
				Type: game.ActionUsePersistentEffect,
				UsePersistentEffect: &game.UsePersistentEffectAction{
					Faction:    faction,
					EffectID:   "tax_collector",
					ClearingID: clearingID,
				},
			})
		}
	}

	return actions
}

func eveningEffectActions(state game.GameState) []game.Action {
	if state.CurrentPhase != game.Evening || state.TurnProgress.EveningMainActionTaken {
		return nil
	}

	faction := state.FactionTurn
	actions := []game.Action{}

	if hasPersistentEffect(state, faction, "codebreakers") && !persistentEffectUsedThisTurn(state, "codebreakers") {
		for _, targetFaction := range otherFactionsInTurnOrder(state, faction) {
			actions = append(actions, game.Action{
				Type: game.ActionUsePersistentEffect,
				UsePersistentEffect: &game.UsePersistentEffectAction{
					Faction:       faction,
					EffectID:      "codebreakers",
					TargetFaction: targetFaction,
				},
			})
		}
	}

	if hasPersistentEffect(state, faction, "cobbler") && !persistentEffectUsedThisTurn(state, "cobbler") {
		actions = append(actions, cobblerMoves(state)...)
	}

	return actions
}

func commandWarrenBattles(state game.GameState) []game.Action {
	var actions []game.Action
	if state.FactionTurn == game.Vagabond {
		actions = rules.ValidVagabondBattleActions(state)
	} else {
		actions = rules.ValidBattlesInState(state.FactionTurn, state)
	}

	for index := range actions {
		if actions[index].Battle != nil {
			actions[index].Battle.SourceEffectID = "command_warren"
		}
	}

	return actions
}

func cobblerMoves(state game.GameState) []game.Action {
	var actions []game.Action
	if state.FactionTurn == game.Vagabond {
		actions = rules.ValidVagabondMoveActions(state)
	} else {
		actions = rules.ValidMovementActions(state.FactionTurn, state.Map)
	}

	for index := range actions {
		if actions[index].Movement != nil {
			actions[index].Movement.SourceEffectID = "cobbler"
		}
	}

	return actions
}

func usePersistentEffect(state *game.GameState, action game.Action) *game.EffectResult {
	if action.UsePersistentEffect == nil {
		return nil
	}

	effect := action.UsePersistentEffect
	switch effect.EffectID {
	case "better_burrow_bank":
		effectDrawCards(state, effect.Faction, 1)
		effectDrawCards(state, effect.TargetFaction, 1)
		markPersistentEffectUsed(state, effect.EffectID)
		return &game.EffectResult{
			EffectID: effect.EffectID,
			Message:  "Better Burrow Bank drew 1 card for each faction.",
		}
	case "royal_claim":
		cardID, ok := persistentEffectCardID(*state, effect.Faction, effect.EffectID)
		if !ok {
			return nil
		}

		points := ruledClearingCount(*state, effect.Faction)
		addVictoryPoints(state, effect.Faction, points)
		removePersistentEffect(state, effect.Faction, cardID)
		DiscardCard(state, cardID)
		return &game.EffectResult{
			EffectID: effect.EffectID,
			Message:  "Royal Claim scored ruled clearings.",
		}
	case "tax_collector":
		index := findClearingIndex(state.Map, effect.ClearingID)
		if index == -1 {
			return nil
		}

		if state.Map.Clearings[index].Warriors[effect.Faction] <= 0 {
			return nil
		}
		state.Map.Clearings[index].Warriors[effect.Faction]--
		effectDrawCards(state, effect.Faction, 1)
		markPersistentEffectUsed(state, effect.EffectID)
		return &game.EffectResult{
			EffectID: effect.EffectID,
			Message:  "Tax Collector removed 1 warrior and drew 1 card.",
		}
	case "codebreakers":
		result := resolveCodebreakers(state, effect.Faction, effect.TargetFaction)
		markPersistentEffectUsed(state, effect.EffectID)
		return result
	case "stand_and_deliver":
		result := resolveStandAndDeliver(state, *effect)
		if result != nil {
			markPersistentEffectUsed(state, effect.EffectID)
		}
		return result
	}

	return nil
}

func effectDrawCards(state *game.GameState, faction game.Faction, count int) {
	if count <= 0 {
		return
	}

	if state.GameMode == game.GameModeOnline {
		DrawCards(state, faction, count)
		return
	}

	if faction == state.PlayerFaction {
		return
	}

	incrementOtherHandCount(state, faction, count)
}

func ruledClearingCount(state game.GameState, faction game.Faction) int {
	count := 0
	for _, clearing := range state.Map.Clearings {
		ruler, ruled := rules.Ruler(clearing)
		if ruled && ruler == faction {
			count++
		}
	}

	return count
}

func factionHandSize(state game.GameState, faction game.Faction) int {
	if tracksHandForFaction(state, faction) {
		return len(factionHand(state, faction))
	}

	return state.OtherHandCounts[faction]
}

func resolveCodebreakers(state *game.GameState, faction game.Faction, targetFaction game.Faction) *game.EffectResult {
	if !tracksHandForFaction(*state, targetFaction) {
		return &game.EffectResult{
			EffectID: "codebreakers",
			Message:  "Codebreakers: inspect the target faction's physical hand.",
		}
	}

	revealed := append([]game.Card(nil), factionHand(*state, targetFaction)...)
	message := "Codebreakers revealed no cards."
	if len(revealed) > 0 {
		names := make([]string, 0, len(revealed))
		for _, card := range revealed {
			names = append(names, card.Name)
		}
		message = "Codebreakers revealed: " + joinCardNames(names)
	}

	return &game.EffectResult{
		EffectID: "codebreakers",
		Message:  message,
		Cards:    revealed,
	}
}

func resolveStandAndDeliver(state *game.GameState, effect game.UsePersistentEffectAction) *game.EffectResult {
	targetFaction := effect.TargetFaction
	if factionHandSize(*state, targetFaction) == 0 {
		return nil
	}

	addVictoryPoints(state, targetFaction, 1)

	if state.GameMode == game.GameModeOnline {
		card, ok := removeRandomCardFromFactionHand(state, targetFaction)
		if !ok {
			return nil
		}
		appendCardToFactionHand(state, effect.Faction, card)
		return &game.EffectResult{
			EffectID: "stand_and_deliver",
			Message:  "Stand and Deliver! transferred " + card.Name + ".",
			Cards:    []game.Card{card},
		}
	}

	if effect.Faction == state.PlayerFaction {
		if effect.ObservedCardID > 0 {
			card, ok := CardByID(effect.ObservedCardID)
			if ok {
				appendCardToFactionHand(state, effect.Faction, card)
				decrementOtherHandCount(state, targetFaction, 1)
				return &game.EffectResult{
					EffectID: "stand_and_deliver",
					Message:  "Stand and Deliver! recorded " + card.Name + " in your hand.",
					Cards:    []game.Card{card},
				}
			}
		}

		decrementOtherHandCount(state, targetFaction, 1)
		return &game.EffectResult{
			EffectID: "stand_and_deliver",
			Message:  "Stand and Deliver! stole a card. Record the stolen card in your hand manually.",
		}
	}

	if targetFaction == state.PlayerFaction {
		card, ok := removeRandomCardFromFactionHand(state, targetFaction)
		if !ok {
			return nil
		}
		incrementOtherHandCount(state, effect.Faction, 1)
		return &game.EffectResult{
			EffectID: "stand_and_deliver",
			Message:  "Stand and Deliver! took " + card.Name + " from your hand.",
			Cards:    []game.Card{card},
		}
	}

	decrementOtherHandCount(state, targetFaction, 1)
	incrementOtherHandCount(state, effect.Faction, 1)
	return &game.EffectResult{
		EffectID: "stand_and_deliver",
		Message:  "Stand and Deliver! transferred 1 hidden card.",
	}
}

func removeRandomCardFromFactionHand(state *game.GameState, faction game.Faction) (game.Card, bool) {
	hand := factionHand(*state, faction)
	if len(hand) == 0 {
		return game.Card{}, false
	}

	rng := nextShuffleRNG(state)
	index := rng.Intn(len(hand))
	card := hand[index]
	_, ok := removeCardFromFactionHand(state, faction, card.ID)
	return card, ok
}

func joinCardNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return names[0]
	}
	if len(names) == 2 {
		return names[0] + " and " + names[1]
	}

	result := ""
	for index, name := range names {
		switch {
		case index == 0:
			result = name
		case index == len(names)-1:
			result += ", and " + name
		default:
			result += ", " + name
		}
	}
	return result
}
