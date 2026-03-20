package game

type Action struct {
	Type     ActionType
	Movement *MovementAction
	Battle   *BattleAction
	Build    *BuildAction
}

type ActionType int

const (
	ActionMovement ActionType = iota
	ActionBattle
	ActionBuild
)

type MovementAction struct {
	Faction  Faction
	MaxCount int
	From     int
	To       int
}

type BattleAction struct {
	Faction       Faction
	ClearingID    int
	TargetFaction Faction
}

type BuildAction struct {
	Faction      Faction
	ClearingID   int
	BuildingType BuildingType
}
