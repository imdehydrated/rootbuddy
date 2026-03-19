package game

type GameState struct {
	Clearings    []Clearing
	FactionTurn  Faction
	CurrentPhase Phase
	Marquise     MarquiseState
}
