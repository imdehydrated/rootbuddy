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

type Item struct {
	Type   ItemType
	Status ItemStatus
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
