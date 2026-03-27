package game

type HiddenCardZone string

const (
	HiddenCardZoneHand       HiddenCardZone = "hand"
	HiddenCardZoneSupporters HiddenCardZone = "supporters"
)

type HiddenCard struct {
	ID           int
	OwnerFaction Faction
	Zone         HiddenCardZone
	KnownCardID  CardID
}
