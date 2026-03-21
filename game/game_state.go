package game

type GameState struct {
	Map          Map
	FactionTurn  Faction
	CurrentPhase Phase
	Marquise     MarquiseState
	TurnProgress TurnProgress
}
