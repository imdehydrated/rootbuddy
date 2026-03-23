package game

type MapID string

const (
	AutumnMapID MapID = "autumn"
)

type Map struct {
	ID        MapID
	Clearings []Clearing
}
