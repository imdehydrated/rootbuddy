package engine

import "github.com/imdehydrated/rootbuddy/game"

const (
	maxMarquiseWarriors = 25
	maxEyrieWarriors    = 20
	maxAllianceWarriors = 10
	maxMarquiseWood     = 8
	maxMarquiseBuilding = 6
	maxEyrieRoosts      = 7
	maxAllianceBases    = 3
	maxAllianceSympathy = 10
)

func ValidateState(state game.GameState) error {
	if err := state.TurnWindow().Validate(); err != nil {
		return err
	}

	if err := validateTurnOrderState(state); err != nil {
		return err
	}

	if err := validateBoardState(state); err != nil {
		return err
	}

	if err := validateFactionCountersAndSupply(state); err != nil {
		return err
	}

	if err := validateItemState(state); err != nil {
		return err
	}

	if err := validateQuestState(state); err != nil {
		return err
	}

	if err := validateKnownCardZones(state); err != nil {
		return err
	}

	for _, count := range state.OtherHandCounts {
		if count < 0 {
			return errInvalidState("other hand counts cannot be negative")
		}
	}

	seenHiddenCardIDs := map[int]struct{}{}
	maxHiddenCardID := 0
	derivedHandCounts := map[game.Faction]int{}
	for _, hidden := range state.HiddenCards {
		if hidden.ID <= 0 {
			return errInvalidState("hidden card ids must be positive")
		}
		if !validFaction(hidden.OwnerFaction) {
			return errInvalidState("hidden cards must have valid owner factions")
		}
		if hidden.Zone != game.HiddenCardZoneHand && hidden.Zone != game.HiddenCardZoneSupporters {
			return errInvalidState("hidden cards must have valid zones")
		}
		if hidden.KnownCardID > 0 {
			if _, ok := CardByID(hidden.KnownCardID); !ok {
				return errInvalidState("known hidden card ids must refer to base deck cards")
			}
		}
		if _, exists := seenHiddenCardIDs[hidden.ID]; exists {
			return errInvalidState("hidden card ids must be unique")
		}
		seenHiddenCardIDs[hidden.ID] = struct{}{}
		if hidden.ID > maxHiddenCardID {
			maxHiddenCardID = hidden.ID
		}
		if hidden.Zone == game.HiddenCardZoneHand && hidden.OwnerFaction != state.PlayerFaction {
			derivedHandCounts[hidden.OwnerFaction]++
		}
	}
	if state.NextHiddenCardID > 0 && state.NextHiddenCardID <= maxHiddenCardID {
		return errInvalidState("next hidden card id must be greater than existing hidden card ids")
	}
	if state.GameMode == game.GameModeAssist && len(state.HiddenCards) > 0 {
		for faction, count := range derivedHandCounts {
			if state.OtherHandCounts[faction] != count {
				return errInvalidState("other hand counts must match hidden hand placeholders in assist mode")
			}
		}
		for faction, count := range state.OtherHandCounts {
			if derivedHandCounts[faction] != count {
				return errInvalidState("other hand counts must match hidden hand placeholders in assist mode")
			}
		}
	}

	for _, cardID := range state.Deck {
		if cardID < 0 {
			return errInvalidState("deck cannot contain negative card ids")
		}
	}

	for _, cardID := range state.DiscardPile {
		if cardID < 0 {
			return errInvalidState("discard pile cannot contain negative card ids")
		}
	}

	seenAvailableDominance := map[game.CardID]struct{}{}
	for _, cardID := range state.AvailableDominance {
		card, ok := CardByID(cardID)
		if !ok || card.Kind != game.DominanceCard {
			return errInvalidState("available dominance must contain only dominance cards")
		}
		if _, exists := seenAvailableDominance[cardID]; exists {
			return errInvalidState("available dominance cannot contain duplicates")
		}
		seenAvailableDominance[cardID] = struct{}{}
	}

	for _, cardID := range state.ActiveDominance {
		card, ok := CardByID(cardID)
		if !ok || card.Kind != game.DominanceCard {
			return errInvalidState("active dominance must contain only dominance cards")
		}
		if _, exists := seenAvailableDominance[cardID]; exists {
			return errInvalidState("dominance card cannot be both active and available")
		}
	}

	if state.CoalitionActive {
		if !hasActiveDominance(state, game.Vagabond) {
			return errInvalidState("coalition requires an active Vagabond dominance card")
		}
		if state.CoalitionPartner == game.Vagabond {
			return errInvalidState("Vagabond cannot form a coalition with itself")
		}
		if !validFaction(state.CoalitionPartner) {
			return errInvalidState("coalition partner must be a valid faction")
		}
	}
	if len(state.WinningCoalition) > 0 {
		if state.GamePhase != game.LifecycleGameOver {
			return errInvalidState("winning coalition requires game over")
		}
		seen := map[game.Faction]struct{}{}
		for _, faction := range state.WinningCoalition {
			if !validFaction(faction) {
				return errInvalidState("winning coalition must contain valid factions")
			}
			if _, exists := seen[faction]; exists {
				return errInvalidState("winning coalition cannot contain duplicates")
			}
			seen[faction] = struct{}{}
		}
	}

	return nil
}

func validFaction(faction game.Faction) bool {
	switch faction {
	case game.Marquise, game.Eyrie, game.Alliance, game.Vagabond:
		return true
	default:
		return false
	}
}

func validSuit(suit game.Suit) bool {
	switch suit {
	case game.Fox, game.Rabbit, game.Mouse, game.Bird:
		return true
	default:
		return false
	}
}

func validBuilding(building game.Building) bool {
	if !validFaction(building.Faction) {
		return false
	}
	switch building.Type {
	case game.Sawmill, game.Workshop, game.Recruiter:
		return building.Faction == game.Marquise
	case game.Roost:
		return building.Faction == game.Eyrie
	case game.Base:
		return building.Faction == game.Alliance
	default:
		return false
	}
}

func validToken(token game.Token) bool {
	if !validFaction(token.Faction) {
		return false
	}
	switch token.Type {
	case game.TokenKeep:
		return token.Faction == game.Marquise
	case game.TokenSympathy:
		return token.Faction == game.Alliance
	default:
		return false
	}
}

func validateTurnOrderState(state game.GameState) error {
	seen := map[game.Faction]struct{}{}
	for _, faction := range state.TurnOrder {
		if !validFaction(faction) {
			return errInvalidState("turn order must contain only valid factions")
		}
		if _, exists := seen[faction]; exists {
			return errInvalidState("turn order cannot contain duplicate factions")
		}
		seen[faction] = struct{}{}
	}
	if state.GamePhase == game.LifecyclePlaying && len(state.TurnOrder) > 0 {
		if _, ok := seen[state.FactionTurn]; !ok {
			return errInvalidState("active faction must be in turn order")
		}
	}
	if state.RoundNumber < 0 {
		return errInvalidState("round number cannot be negative")
	}
	for faction, points := range state.VictoryPoints {
		if !validFaction(faction) {
			return errInvalidState("victory points must use valid factions")
		}
		if points < 0 {
			return errInvalidState("victory points cannot be negative")
		}
	}
	return nil
}

type boardCounts struct {
	warriors          map[game.Faction]int
	wood              int
	keepTokens        int
	sympathyTokens    int
	marquiseBuildings map[game.BuildingType]int
	roosts            int
	baseBySuit        map[game.Suit]int
	baseTotal         int
}

func countBoardPieces(state game.GameState) boardCounts {
	counts := boardCounts{
		warriors:          map[game.Faction]int{},
		marquiseBuildings: map[game.BuildingType]int{},
		baseBySuit:        map[game.Suit]int{},
	}
	for _, clearing := range state.Map.Clearings {
		counts.wood += clearing.Wood
		for faction, count := range clearing.Warriors {
			counts.warriors[faction] += count
		}
		for _, building := range clearing.Buildings {
			switch {
			case building.Faction == game.Marquise:
				counts.marquiseBuildings[building.Type]++
			case building.Faction == game.Eyrie && building.Type == game.Roost:
				counts.roosts++
			case building.Faction == game.Alliance && building.Type == game.Base:
				counts.baseBySuit[clearing.Suit]++
				counts.baseTotal++
			}
		}
		for _, token := range clearing.Tokens {
			switch token.Type {
			case game.TokenKeep:
				counts.keepTokens++
			case game.TokenSympathy:
				counts.sympathyTokens++
			}
		}
	}
	return counts
}

func validateBoardState(state game.GameState) error {
	seenClearings := map[int]struct{}{}
	for _, clearing := range state.Map.Clearings {
		if clearing.ID <= 0 {
			return errInvalidState("clearing ids must be positive")
		}
		if _, exists := seenClearings[clearing.ID]; exists {
			return errInvalidState("clearing ids must be unique")
		}
		seenClearings[clearing.ID] = struct{}{}
		if !validSuit(clearing.Suit) || clearing.Suit == game.Bird {
			return errInvalidState("clearings must have valid non-bird suits")
		}
		if clearing.BuildSlots < 0 {
			return errInvalidState("clearing build slots cannot be negative")
		}
		if clearing.BuildSlots > 0 && len(clearing.Buildings) > clearing.BuildSlots {
			return errInvalidState("clearing buildings cannot exceed build slots")
		}
		if clearing.Wood < 0 {
			return errInvalidState("clearing wood cannot be negative")
		}
		for faction, count := range clearing.Warriors {
			if !validFaction(faction) {
				return errInvalidState("clearing warriors must use valid factions")
			}
			if count < 0 {
				return errInvalidState("clearing warriors cannot be negative")
			}
		}
		for _, building := range clearing.Buildings {
			if !validBuilding(building) {
				return errInvalidState("buildings must have legal faction and type")
			}
		}
		for _, token := range clearing.Tokens {
			if !validToken(token) {
				return errInvalidState("tokens must have legal faction and type")
			}
		}
		for _, item := range clearing.RuinItems {
			if !validItemType(item) {
				return errInvalidState("ruin items must have valid item types")
			}
		}
	}

	for _, forest := range state.Map.Forests {
		if forest.ID <= 0 {
			return errInvalidState("forest ids must be positive")
		}
		for _, clearingID := range forest.AdjacentClearings {
			if _, ok := seenClearings[clearingID]; !ok && len(state.Map.Clearings) > 0 {
				return errInvalidState("forest adjacency must reference existing clearings")
			}
		}
	}

	return nil
}

func validateFactionCountersAndSupply(state game.GameState) error {
	counts := countBoardPieces(state)
	if err := validateSupplyTotal(state.Marquise.WarriorSupply, counts.warriors[game.Marquise], maxMarquiseWarriors, "Marquise warriors"); err != nil {
		return err
	}
	if err := validateSupplyTotal(state.Eyrie.WarriorSupply, counts.warriors[game.Eyrie], maxEyrieWarriors, "Eyrie warriors"); err != nil {
		return err
	}
	if err := validateSupplyTotal(state.Alliance.WarriorSupply, counts.warriors[game.Alliance]+state.Alliance.Officers, maxAllianceWarriors, "Alliance warriors"); err != nil {
		return err
	}
	if err := validateSupplyTotal(state.Marquise.WoodSupply, counts.wood, maxMarquiseWood, "Marquise wood"); err != nil {
		return err
	}
	if state.Alliance.Officers < 0 {
		return errInvalidState("Alliance officers cannot be negative")
	}

	if err := validatePlacedCounter(state.Marquise.SawmillsPlaced, counts.marquiseBuildings[game.Sawmill], maxMarquiseBuilding, "Marquise sawmills"); err != nil {
		return err
	}
	if err := validatePlacedCounter(state.Marquise.WorkshopsPlaced, counts.marquiseBuildings[game.Workshop], maxMarquiseBuilding, "Marquise workshops"); err != nil {
		return err
	}
	if err := validatePlacedCounter(state.Marquise.RecruitersPlaced, counts.marquiseBuildings[game.Recruiter], maxMarquiseBuilding, "Marquise recruiters"); err != nil {
		return err
	}
	if err := validatePlacedCounter(state.Eyrie.RoostsPlaced, counts.roosts, maxEyrieRoosts, "Eyrie roosts"); err != nil {
		return err
	}
	if err := validatePlacedCounter(state.Alliance.SympathyPlaced, counts.sympathyTokens, maxAllianceSympathy, "Alliance sympathy"); err != nil {
		return err
	}
	if counts.baseTotal > maxAllianceBases {
		return errInvalidState("Alliance bases cannot exceed piece limit")
	}
	if err := validateAllianceBaseFlag(state.Alliance.FoxBasePlaced, counts.baseBySuit[game.Fox], "fox"); err != nil {
		return err
	}
	if err := validateAllianceBaseFlag(state.Alliance.RabbitBasePlaced, counts.baseBySuit[game.Rabbit], "rabbit"); err != nil {
		return err
	}
	if err := validateAllianceBaseFlag(state.Alliance.MouseBasePlaced, counts.baseBySuit[game.Mouse], "mouse"); err != nil {
		return err
	}
	if counts.keepTokens > 1 {
		return errInvalidState("Marquise keep cannot appear more than once")
	}
	if state.Marquise.KeepClearingID != 0 && counts.keepTokens == 0 {
		return errInvalidState("Marquise keep clearing id requires a keep token")
	}
	return nil
}

func validateSupplyTotal(supply int, placed int, limit int, label string) error {
	if supply < 0 {
		return errInvalidState(label + " supply cannot be negative")
	}
	if placed > limit || supply > limit || placed+supply > limit {
		return errInvalidState(label + " exceed piece limit")
	}
	return nil
}

func validatePlacedCounter(counter int, boardCount int, limit int, label string) error {
	if counter < 0 {
		return errInvalidState(label + " counter cannot be negative")
	}
	if counter > limit || boardCount > limit {
		return errInvalidState(label + " exceed piece limit")
	}
	if counter != boardCount {
		return errInvalidState(label + " counter must match board pieces")
	}
	return nil
}

func validateAllianceBaseFlag(flag bool, boardCount int, suitLabel string) error {
	if boardCount > 1 {
		return errInvalidState("Alliance cannot have multiple " + suitLabel + " bases")
	}
	if flag != (boardCount == 1) {
		return errInvalidState("Alliance base flags must match board bases")
	}
	return nil
}

func validItemType(item game.ItemType) bool {
	switch item {
	case game.ItemTea, game.ItemCoin, game.ItemCrossbow, game.ItemHammer, game.ItemSword, game.ItemTorch, game.ItemBoots, game.ItemBag:
		return true
	default:
		return false
	}
}

func validItemStatus(status game.ItemStatus) bool {
	switch status {
	case game.ItemReady, game.ItemExhausted, game.ItemDamaged:
		return true
	default:
		return false
	}
}

func validItemZone(zone game.ItemZone) bool {
	switch zone {
	case game.ItemZoneUnspecified, game.ItemZoneTrack, game.ItemZoneSatchel, game.ItemZoneDamaged:
		return true
	default:
		return false
	}
}

func validateItemState(state game.GameState) error {
	initialSupply := InitialItemSupply()
	for item, count := range state.ItemSupply {
		if !validItemType(item) {
			return errInvalidState("item supply must use valid item types")
		}
		if count < 0 {
			return errInvalidState("item supply cannot be negative")
		}
		if count > initialSupply[item] {
			return errInvalidState("item supply cannot exceed initial supply")
		}
	}
	for faction, items := range state.CraftedItems {
		if !validFaction(faction) || faction == game.Vagabond {
			return errInvalidState("crafted items must belong to non-Vagabond factions")
		}
		for _, item := range items {
			if !validItemType(item) {
				return errInvalidState("crafted items must use valid item types")
			}
		}
	}
	for _, item := range state.Vagabond.Items {
		if !validItemType(item.Type) || !validItemStatus(item.Status) || !validItemZone(item.Zone) {
			return errInvalidState("Vagabond items must have valid type, status, and zone")
		}
		expected := game.ItemZoneForStatus(item.Type, item.Status)
		if item.Zone != game.ItemZoneUnspecified && item.Zone != expected {
			return errInvalidState("Vagabond item zones must match status")
		}
		if item.Status == game.ItemDamaged && item.DamagedSide != game.ItemReady && item.DamagedSide != game.ItemExhausted {
			return errInvalidState("damaged Vagabond items must remember a valid side")
		}
	}
	for faction, relationship := range state.Vagabond.Relationships {
		if !validFaction(faction) || faction == game.Vagabond {
			return errInvalidState("Vagabond relationships must target non-Vagabond factions")
		}
		if relationship < game.RelHostile || relationship > game.RelAllied {
			return errInvalidState("Vagabond relationships must use valid levels")
		}
	}
	if state.Vagabond.ClearingID != 0 && state.Vagabond.ForestID != 0 {
		return errInvalidState("Vagabond cannot be in a clearing and forest at once")
	}
	if state.Vagabond.InForest && state.Vagabond.ForestID == 0 {
		return errInvalidState("Vagabond in-forest state requires a forest id")
	}
	if !state.Vagabond.InForest && state.Vagabond.ForestID != 0 {
		return errInvalidState("Vagabond forest id requires in-forest state")
	}
	return nil
}

func validateQuestState(state game.GameState) error {
	seen := map[game.QuestID]struct{}{}
	addQuestID := func(id game.QuestID) error {
		if _, ok := questByID(id); !ok {
			return errInvalidState("quest ids must refer to base quests")
		}
		if _, exists := seen[id]; exists {
			return errInvalidState("quests cannot appear in multiple zones")
		}
		seen[id] = struct{}{}
		return nil
	}
	for _, id := range state.QuestDeck {
		if err := addQuestID(id); err != nil {
			return err
		}
	}
	for _, id := range state.QuestDiscard {
		if err := addQuestID(id); err != nil {
			return err
		}
	}
	for _, quest := range state.Vagabond.QuestsAvailable {
		if registered, ok := questByID(quest.ID); !ok || registered.Name != quest.Name || registered.Suit != quest.Suit {
			return errInvalidState("available quests must match base quest data")
		}
		if err := addQuestID(quest.ID); err != nil {
			return err
		}
	}
	for _, quest := range state.Vagabond.QuestsCompleted {
		if registered, ok := questByID(quest.ID); !ok || registered.Name != quest.Name || registered.Suit != quest.Suit {
			return errInvalidState("completed quests must match base quest data")
		}
		if err := addQuestID(quest.ID); err != nil {
			return err
		}
	}
	return nil
}

func validateKnownCardZones(state game.GameState) error {
	seen := map[game.CardID]string{}
	add := func(cardID game.CardID, zone string) error {
		if cardID <= 0 {
			return nil
		}
		if _, ok := CardByID(cardID); !ok {
			return errInvalidState("known card ids must refer to base deck cards")
		}
		if existing, exists := seen[cardID]; exists {
			return errInvalidState("known card " + existing + " duplicates another public card zone")
		}
		seen[cardID] = zone
		return nil
	}
	for _, cardID := range state.Deck {
		if err := add(cardID, "deck"); err != nil {
			return err
		}
	}
	for _, cardID := range state.DiscardPile {
		if err := add(cardID, "discard"); err != nil {
			return err
		}
	}
	for _, card := range state.Marquise.CardsInHand {
		if err := add(card.ID, "Marquise hand"); err != nil {
			return err
		}
	}
	for _, card := range state.Eyrie.CardsInHand {
		if err := add(card.ID, "Eyrie hand"); err != nil {
			return err
		}
	}
	for _, card := range state.Alliance.CardsInHand {
		if err := add(card.ID, "Alliance hand"); err != nil {
			return err
		}
	}
	for _, card := range state.Alliance.Supporters {
		if err := add(card.ID, "Alliance supporters"); err != nil {
			return err
		}
	}
	for _, card := range state.Vagabond.CardsInHand {
		if err := add(card.ID, "Vagabond hand"); err != nil {
			return err
		}
	}
	for faction, cardIDs := range state.PersistentEffects {
		if !validFaction(faction) {
			return errInvalidState("persistent effects must use valid factions")
		}
		for _, cardID := range cardIDs {
			if err := add(cardID, "persistent effects"); err != nil {
				return err
			}
		}
	}
	for _, cardID := range decreeCardIDs(state.Eyrie.Decree) {
		if cardID == game.LoyalVizier1 || cardID == game.LoyalVizier2 {
			continue
		}
		if err := add(cardID, "Eyrie decree"); err != nil {
			return err
		}
	}
	for _, hidden := range state.HiddenCards {
		if err := add(hidden.KnownCardID, "hidden card"); err != nil {
			return err
		}
	}
	return nil
}

func decreeCardIDs(decree game.Decree) []game.CardID {
	cardIDs := []game.CardID{}
	cardIDs = append(cardIDs, decree.Recruit...)
	cardIDs = append(cardIDs, decree.Move...)
	cardIDs = append(cardIDs, decree.Battle...)
	cardIDs = append(cardIDs, decree.Build...)
	return cardIDs
}

type invalidStateError string

func (msg invalidStateError) Error() string {
	return string(msg)
}

func errInvalidState(message string) error {
	return invalidStateError(message)
}
