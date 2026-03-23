package game

type Clearing struct {
	ID         int
	Suit       Suit
	BuildSlots int
	Adj        []int
	Ruins      bool
	RuinItems  []ItemType
	Wood       int
	Warriors   map[Faction]int
	Buildings  []Building
	Tokens     []Token
}
