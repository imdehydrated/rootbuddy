package game

type SetupStage int

const (
	SetupStageUnspecified SetupStage = iota
	SetupStageMarquise
	SetupStageEyrie
	SetupStageVagabond
	SetupStageComplete
)
