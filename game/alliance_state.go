package game

type AllianceState struct {
	CardsInHand      []Card
	WarriorSupply    int
	Supporters       []Card
	Officers         int
	FoxBasePlaced    bool
	RabbitBasePlaced bool
	MouseBasePlaced  bool
	SympathyPlaced   int
}
