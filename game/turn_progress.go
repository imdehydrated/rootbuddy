package game

type TurnProgress struct {
	ActionsUsed             int
	BonusActions            int
	MarchesUsed             int
	RecruitUsed             bool
	UsedWorkshopClearings   []int
	HasCrafted              bool
	DecreeColumnsResolved   int
	DecreeCardsResolved     int
	ResolvedDecreeCardIDs   []CardID
	CardsAddedToDecree      int
	OfficerActionsUsed      int
	HasOrganized            bool
	HasRefreshed            bool
	HasSlipped              bool
	UsedPersistentEffectIDs []string
	BirdsongMainActionTaken bool
	SpreadSympathyStarted   bool
	DaylightMainActionTaken bool
	EveningMainActionTaken  bool
}
