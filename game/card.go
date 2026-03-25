package game

type DeckID int

const (
	BaseDeck DeckID = iota
	ExilesAndPartisansDeck
)

type CardID int
type CardKind int

const (
	ItemCard CardKind = iota
	PersistentEffectCard
	OneTimeEffectCard
	AmbushCard
	DominanceCard
)

type CraftingCost struct {
	Fox    int
	Rabbit int
	Mouse  int
	Any    int
}

type Card struct {
	ID           CardID
	Deck         DeckID
	Name         string
	Suit         Suit
	Kind         CardKind
	CraftingCost CraftingCost
	CraftedItem  *ItemType
	EffectID     string
	VP           int
}
