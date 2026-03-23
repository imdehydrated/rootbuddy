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

type VagabondState struct {
	CardsInHand     []Card
	Character       VagabondCharacter
	ClearingID      int
	InForest        bool
	Items           []Item
	Relationships   map[Faction]RelationshipLevel
	QuestsCompleted []Card
	QuestsAvailable []Card
}
