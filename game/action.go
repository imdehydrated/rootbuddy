package game

type Action struct {
	Type     ActionType
	Movement *MovementAction
	Battle   *BattleAction
	Build    *BuildAction
	Recruit  *RecruitAction
}

type ActionType int

const (
	ActionMovement ActionType = iota
	ActionBattle
	ActionBuild
	ActionRecruit
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

type RecruitAction struct {
	Faction     Faction
	ClearingIDs []int
}
