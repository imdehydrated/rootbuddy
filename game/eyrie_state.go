package game

type EyrieLeader int

const (
	LeaderBuilder EyrieLeader = iota
	LeaderCharismatic
	LeaderCommander
	LeaderDespot
)

type Decree struct {
	Recruit []CardID
	Move    []CardID
	Battle  []CardID
	Build   []CardID
}

type EyrieState struct {
	CardsInHand      []Card
	WarriorSupply    int
	RoostsPlaced     int
	Leader           EyrieLeader
	AvailableLeaders []EyrieLeader
	Decree           Decree
	CraftedThisTurn  bool
}

const (
	LoyalVizier1 CardID = -1
	LoyalVizier2 CardID = -2
)
