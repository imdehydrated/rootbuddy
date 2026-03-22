package game

type GameState struct {
	Map          Map
	FactionTurn  Faction
	CurrentPhase Phase
	CurrentStep  TurnStep
	Marquise     MarquiseState
	TurnProgress TurnProgress
}
