package game

type Clearing struct {
	ID         int
	Suit       Suit
	BuildSlots int
	Adj        []int
	Ruins      bool
	Wood       int
	Warriors   map[Faction]int
	Buildings  []Building
}
