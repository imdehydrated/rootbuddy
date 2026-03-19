package game

type Move struct {
	Type   MoveType
	March  *MarchDetail
	Battle *BattleDetail
	Build  *BuildDetail
}

type MoveType int

const (
	MoveMarch MoveType = iota
	MoveBattle
	MoveBuild
)

type MarchDetail struct {
	From int
	To   int
}

type BattleDetail struct {
	ClearingID    int
	TargetFaction Faction
}

type BuildDetail struct {
	ClearingID   int
	BuildingType BuildingType
}
