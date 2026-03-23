package game

type MapID string

const (
	AutumnMapID MapID = "autumn"
)

type Forest struct {
	ID                int
	AdjacentClearings []int
}

type Map struct {
	ID        MapID
	Clearings []Clearing
	Forests   []Forest
}
