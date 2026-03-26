package game

type BattleTimingStep string

const (
	BattleTimingAmbush        BattleTimingStep = "ambush"
	BattleTimingCounterAmbush BattleTimingStep = "counter_ambush"
	BattleTimingModifiers     BattleTimingStep = "modifiers"
	BattleTimingRolls         BattleTimingStep = "rolls"
)

type BattleContext struct {
	Action                             Action
	ClearingSuit                       Suit
	Timing                             []BattleTimingStep
	AttackerHasScoutingParty           bool
	CanDefenderAmbush                  bool
	CanAttackerCounterAmbush           bool
	CanAttackerArmorers                bool
	CanDefenderArmorers                bool
	CanAttackerBrutalTactics           bool
	CanDefenderSappers                 bool
	AssistDefenderAmbushPromptRequired bool
}
