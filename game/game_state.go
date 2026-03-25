package game

type GameLifecycle int

const (
	LifecycleSetup GameLifecycle = iota
	LifecyclePlaying
	LifecycleGameOver
)

type GameMode int

const (
	GameModeOnline GameMode = iota
	GameModeAssist
)

type GameState struct {
	Map               Map
	GameMode          GameMode
	GamePhase         GameLifecycle
	PlayerFaction     Faction
	Winner            Faction
	RoundNumber       int
	FactionTurn       Faction
	CurrentPhase      Phase
	CurrentStep       TurnStep
	TurnOrder         []Faction
	VictoryPoints     map[Faction]int
	Deck              []CardID
	DiscardPile       []CardID
	ItemSupply        map[ItemType]int
	PersistentEffects map[Faction][]CardID
	QuestDeck         []QuestID
	QuestDiscard      []QuestID
	OtherHandCounts   map[Faction]int
	Marquise          MarquiseState
	Eyrie             EyrieState
	Alliance          AllianceState
	Vagabond          VagabondState
	TurnProgress      TurnProgress
}
