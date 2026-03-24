package game

type TurnStep int

const (
	StepUnspecified TurnStep = iota
	StepBirdsong
	StepDaylightCraft
	StepDaylightActions
	StepEvening
)
