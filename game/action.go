package game

type Action struct {
	Type             ActionType
	Movement         *MovementAction
	Battle           *BattleAction
	BattleResolution *BattleResolutionAction
	Build            *BuildAction
	Recruit          *RecruitAction
	Overwork         *OverworkAction
	Craft            *CraftAction
}

type ActionType int

const (
	ActionMovement ActionType = iota
	ActionBattle
	ActionBattleResolution
	ActionBuild
	ActionRecruit
	ActionOverwork
	ActionCraft
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

type BattleModifiers struct {
	AttackerHitModifier int
	DefenderHitModifier int
	IgnoreHitsToAttacker bool
	IgnoreHitsToDefender bool
}

type BattleResolutionAction struct {
	Faction              Faction
	ClearingID           int
	TargetFaction        Faction
	AttackerRoll         int
	DefenderRoll         int
	AttackerHitModifier  int
	DefenderHitModifier  int
	IgnoreHitsToAttacker bool
	IgnoreHitsToDefender bool
	AttackerLosses       int
	DefenderLosses       int
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

type OverworkAction struct {
	Faction    Faction
	ClearingID int
	CardID     CardID
}

type CraftAction struct {
	Faction               Faction
	CardID                CardID
	UsedWorkshopClearings []int
}
