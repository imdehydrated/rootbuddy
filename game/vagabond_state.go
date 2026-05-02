package game

type VagabondCharacter int

const (
	CharThief VagabondCharacter = iota
	CharTinker
	CharRanger
)

type ItemType int

const (
	ItemTea ItemType = iota
	ItemCoin
	ItemCrossbow
	ItemHammer
	ItemSword
	ItemTorch
	ItemBoots
	ItemBag
)

type ItemStatus int

const (
	ItemReady ItemStatus = iota
	ItemExhausted
	ItemDamaged
)

type ItemZone int

const (
	ItemZoneUnspecified ItemZone = iota
	ItemZoneTrack
	ItemZoneSatchel
	ItemZoneDamaged
)

type Item struct {
	Type   ItemType
	Status ItemStatus
	Zone   ItemZone
}

func IsTrackItemType(itemType ItemType) bool {
	switch itemType {
	case ItemTea, ItemCoin, ItemBag:
		return true
	default:
		return false
	}
}

func ItemZoneForStatus(itemType ItemType, status ItemStatus) ItemZone {
	if status == ItemDamaged {
		return ItemZoneDamaged
	}
	if IsTrackItemType(itemType) && status == ItemReady {
		return ItemZoneTrack
	}
	return ItemZoneSatchel
}

func NormalizeItemZone(item Item) Item {
	item.Zone = ItemZoneForStatus(item.Type, item.Status)
	return item
}

func ItemCurrentZone(item Item) ItemZone {
	return NormalizeItemZone(item).Zone
}

type RelationshipLevel int

const (
	RelHostile RelationshipLevel = iota
	RelIndifferent
	RelFriendly
	RelAllied
)

type QuestID int

type Quest struct {
	ID            QuestID
	Name          string
	Suit          Suit
	RequiredItems []ItemType
}

type VagabondState struct {
	CardsInHand     []Card
	Character       VagabondCharacter
	ClearingID      int
	ForestID        int
	InForest        bool
	Items           []Item
	Relationships   map[Faction]RelationshipLevel
	QuestsCompleted []Quest
	QuestsAvailable []Quest
}
