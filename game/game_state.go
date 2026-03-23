package game

type GameState struct {
	Map           Map
	FactionTurn   Faction
	CurrentPhase  Phase
	CurrentStep   TurnStep
	TurnOrder     []Faction
	VictoryPoints map[Faction]int
	Marquise      MarquiseState
	Eyrie         EyrieState
	Alliance      AllianceState
	Vagabond      VagabondState
	TurnProgress  TurnProgress
}
