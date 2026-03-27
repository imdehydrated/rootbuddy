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
	Map                Map
	GameMode           GameMode
	TrackAllHands      bool `json:"-"`
	RandomSeed         int64
	ShuffleCount       int64
	GamePhase          GameLifecycle
	SetupStage         SetupStage
	PlayerFaction      Faction
	Winner             Faction
	WinningCoalition   []Faction
	RoundNumber        int
	FactionTurn        Faction
	CurrentPhase       Phase
	CurrentStep        TurnStep
	TurnOrder          []Faction
	VictoryPoints      map[Faction]int
	Deck               []CardID
	DiscardPile        []CardID
	AvailableDominance []CardID
	ActiveDominance    map[Faction]CardID
	CoalitionActive    bool
	CoalitionPartner   Faction
	ItemSupply         map[ItemType]int
	PersistentEffects  map[Faction][]CardID
	QuestDeck          []QuestID
	QuestDiscard       []QuestID
	OtherHandCounts    map[Faction]int
	HiddenCards        []HiddenCard
	NextHiddenCardID   int
	Marquise           MarquiseState
	Eyrie              EyrieState
	Alliance           AllianceState
	Vagabond           VagabondState
	TurnProgress       TurnProgress
}
