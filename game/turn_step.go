package game

type TurnStep int

const (
	StepUnspecified TurnStep = iota
	StepRecruit
	StepDaylightActions
	StepEvening
)
