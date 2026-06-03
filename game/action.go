package game

type Action struct {
	Type                ActionType
	Movement            *MovementAction
	Battle              *BattleAction
	BattleResolution    *BattleResolutionAction
	Build               *BuildAction
	Recruit             *RecruitAction
	Overwork            *OverworkAction
	Craft               *CraftAction
	AddToDecree         *AddToDecreeAction
	SpreadSympathy      *SpreadSympathyAction
	Revolt              *RevoltAction
	Mobilize            *MobilizeAction
	Train               *TrainAction
	Organize            *OrganizeAction
	Explore             *ExploreAction
	Quest               *QuestAction
	Aid                 *AidAction
	Strike              *StrikeAction
	Repair              *RepairAction
	VagabondSteal       *VagabondStealAction
	VagabondDayLabor    *VagabondDayLaborAction
	VagabondHideout     *VagabondHideoutAction
	Turmoil             *TurmoilAction
	Daybreak            *DaybreakAction
	Slip                *SlipAction
	BirdsongWood        *BirdsongWoodAction
	EveningDraw         *EveningDrawAction
	ScoreRoosts         *ScoreRoostsAction
	PassPhase           *PassPhaseAction
	AddCardToHand       *AddCardToHandAction
	RemoveCardFromHand  *RemoveCardFromHandAction
	OtherPlayerDraw     *OtherPlayerDrawAction
	OtherPlayerPlay     *OtherPlayerPlayAction
	DiscardEffect       *DiscardEffectAction
	ActivateDominance   *ActivateDominanceAction
	TakeDominance       *TakeDominanceAction
	MarquiseSetup       *MarquiseSetupAction
	EyrieSetup          *EyrieSetupAction
	VagabondSetup       *VagabondSetupAction
	UsePersistentEffect *UsePersistentEffectAction
	EyrieEmergency      *EyrieEmergencyOrdersAction
	EyrieNewRoost       *EyrieNewRoostAction
	EveningDiscard      *EveningDiscardAction
	VagabondRest        *VagabondRestAction
	VagabondDiscard     *VagabondDiscardAction
	VagabondCapacity    *VagabondItemCapacityAction
	FieldHospitals      *FieldHospitalsAction
	MarquiseExtraAction *MarquiseExtraActionAction
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
	ActionAddToDecree
	ActionSpreadSympathy
	ActionRevolt
	ActionMobilize
	ActionTrain
	ActionOrganize
	ActionExplore
	ActionQuest
	ActionAid
	ActionStrike
	ActionRepair
	ActionTurmoil
	ActionDaybreak
	ActionSlip
	ActionBirdsongWood
	ActionEveningDraw
	ActionScoreRoosts
	ActionPassPhase
	ActionAddCardToHand
	ActionRemoveCardFromHand
	ActionOtherPlayerDraw
	ActionOtherPlayerPlay
	ActionDiscardEffect
	ActionActivateDominance
	ActionTakeDominance
	ActionMarquiseSetup
	ActionEyrieSetup
	ActionVagabondSetup
	ActionUsePersistentEffect
	ActionEyrieEmergencyOrders
	ActionEyrieNewRoost
	ActionVagabondRest
	ActionVagabondDiscard
	ActionVagabondItemCapacity
	ActionEveningDiscard
	ActionFieldHospitals
	ActionMarquiseExtraAction
	ActionVagabondSteal
	ActionVagabondDayLabor
	ActionVagabondHideout
)

type MovementAction struct {
	Faction        Faction
	Count          int
	MaxCount       int
	From           int
	To             int
	FromForestID   int
	ToForestID     int
	AlliedFaction  Faction
	AlliedWarriors int
	DecreeCardID   CardID
	SourceEffectID string
}

type BattleAction struct {
	Faction          Faction
	ClearingID       int
	TargetFaction    Faction
	UseAlliedFaction bool
	AlliedFaction    Faction
	DecreeCardID     CardID
	SourceEffectID   string
}

type BattleModifiers struct {
	AttackerHitModifier         int
	DefenderHitModifier         int
	IgnoreHitsToAttacker        bool
	IgnoreHitsToDefender        bool
	DefenderAmbush              bool
	DefenderAmbushCardID        CardID
	AttackerCounterAmbush       bool
	AttackerCounterAmbushCardID CardID
	AttackerUsesArmorers        bool
	DefenderUsesArmorers        bool
	AttackerUsesBrutalTactics   bool
	DefenderUsesSappers         bool
}

type BattlePieceKind int

const (
	BattlePieceBuilding BattlePieceKind = iota
	BattlePieceToken
	BattlePieceWood
)

type BattlePieceLoss struct {
	Kind         BattlePieceKind
	BuildingType BuildingType
	TokenType    TokenType
}

type BattleResolutionAction struct {
	Faction                     Faction
	ClearingID                  int
	TargetFaction               Faction
	DecreeCardID                CardID
	AttackerRoll                int
	DefenderRoll                int
	AttackerHitModifier         int
	DefenderHitModifier         int
	IgnoreHitsToAttacker        bool
	IgnoreHitsToDefender        bool
	DefenderAmbushed            bool
	DefenderAmbushCardID        CardID
	AttackerCounterAmbush       bool
	AttackerCounterAmbushCardID CardID
	AttackerUsedArmorers        bool
	DefenderUsedArmorers        bool
	AttackerUsedBrutalTactics   bool
	DefenderUsedSappers         bool
	AmbushHitsToAttacker        int
	AttackerLosses              int
	DefenderLosses              int
	UseAlliedFaction            bool
	AlliedFaction               Faction
	AlliedWarriorLosses         int
	AttackerDamagedItemIndexes  []int
	DefenderDamagedItemIndexes  []int
	AttackerPieceLosses         []BattlePieceLoss
	DefenderPieceLosses         []BattlePieceLoss
	SourceEffectID              string
}

type BuildAction struct {
	Faction      Faction
	ClearingID   int
	BuildingType BuildingType
	WoodSources  []WoodSource
	DecreeCardID CardID
}

type WoodSource struct {
	ClearingID int
	Amount     int
}

type RecruitAction struct {
	Faction      Faction
	ClearingIDs  []int
	DecreeCardID CardID
}

type OverworkAction struct {
	Faction    Faction
	ClearingID int
	CardID     CardID
}

type CraftAction struct {
	Faction                    Faction
	CardID                     CardID
	UsedWorkshopClearings      []int
	DamagedVagabondItemIndexes []int
}

type DecreeColumn int

const (
	DecreeRecruit DecreeColumn = iota
	DecreeMove
	DecreeBattle
	DecreeBuild
)

type AddToDecreeAction struct {
	Faction Faction
	CardIDs []CardID
	Columns []DecreeColumn
}

type SpreadSympathyAction struct {
	Faction          Faction
	ClearingID       int
	SupporterCardIDs []CardID
}

type RevoltAction struct {
	Faction                    Faction
	ClearingID                 int
	BaseSuit                   Suit
	SupporterCardIDs           []CardID
	DamagedVagabondItemIndexes []int
}

type MobilizeAction struct {
	Faction Faction
	CardID  CardID
}

type TrainAction struct {
	Faction Faction
	CardID  CardID
}

type OrganizeAction struct {
	Faction    Faction
	ClearingID int
}

type ExploreAction struct {
	Faction    Faction
	ClearingID int
	ItemType   ItemType
}

type QuestReward int

const (
	QuestRewardVictoryPoints QuestReward = iota
	QuestRewardDrawCards
)

type QuestAction struct {
	Faction     Faction
	QuestID     QuestID
	ItemIndexes []int
	Reward      QuestReward
}

type AidAction struct {
	Faction       Faction
	TargetFaction Faction
	ClearingID    int
	CardID        CardID
	ItemIndex     int
	TakeItemIndex *int
}

type StrikeAction struct {
	Faction       Faction
	ClearingID    int
	TargetFaction Faction
}

type RepairAction struct {
	Faction   Faction
	ItemIndex int
}

type VagabondStealAction struct {
	Faction        Faction
	ClearingID     int
	TargetFaction  Faction
	ObservedCardID CardID
}

type VagabondDayLaborAction struct {
	Faction    Faction
	ClearingID int
	CardID     CardID
}

type VagabondHideoutAction struct {
	Faction     Faction
	ItemIndexes []int
}

type TurmoilAction struct {
	Faction   Faction
	NewLeader EyrieLeader
}

type DaybreakAction struct {
	Faction              Faction
	RefreshedItemIndexes []int
}

type SlipAction struct {
	Faction      Faction
	From         int
	To           int
	FromForestID int
	ToForestID   int
}

type BirdsongWoodAction struct {
	Faction     Faction
	ClearingIDs []int
	Amount      int
}

type EveningDrawAction struct {
	Faction Faction
	Count   int
}

type EveningDiscardAction struct {
	Faction Faction
	CardIDs []CardID
	Count   int
}

type ScoreRoostsAction struct {
	Faction Faction
	Points  int
}

type PassPhaseAction struct {
	Faction Faction
}

type AddCardToHandAction struct {
	Faction Faction
	CardID  CardID
}

type RemoveCardFromHandAction struct {
	Faction Faction
	CardID  CardID
}

type OtherPlayerDrawAction struct {
	Faction Faction
	Count   int
}

type OtherPlayerPlayAction struct {
	Faction Faction
	CardID  CardID
}

type DiscardEffectAction struct {
	Faction Faction
	CardID  CardID
}

type ActivateDominanceAction struct {
	Faction       Faction
	CardID        CardID
	TargetFaction Faction
}

type TakeDominanceAction struct {
	Faction         Faction
	DominanceCardID CardID
	SpentCardID     CardID
}

type MarquiseSetupAction struct {
	Faction             Faction
	KeepClearingID      int
	SawmillClearingID   int
	WorkshopClearingID  int
	RecruiterClearingID int
}

type EyrieSetupAction struct {
	Faction    Faction
	Leader     EyrieLeader
	ClearingID int
}

type VagabondSetupAction struct {
	Faction   Faction
	Character VagabondCharacter
	ForestID  int
}

type UsePersistentEffectAction struct {
	Faction        Faction
	EffectID       string
	TargetFaction  Faction
	ClearingID     int
	ObservedCardID CardID
}

type EyrieEmergencyOrdersAction struct {
	Faction Faction
	Count   int
}

type EyrieNewRoostAction struct {
	Faction    Faction
	ClearingID int
}

type VagabondRestAction struct {
	Faction Faction
}

type VagabondDiscardAction struct {
	Faction Faction
	CardIDs []CardID
}

type VagabondItemCapacityAction struct {
	Faction     Faction
	ItemIndexes []int
}

type FieldHospitalsPending struct {
	ClearingID   int
	Suit         Suit
	WarriorCount int
}

type FieldHospitalsAction struct {
	Faction    Faction
	ClearingID int
	CardID     CardID
	Decline    bool
}

type MarquiseExtraActionAction struct {
	Faction Faction
	CardID  CardID
}
