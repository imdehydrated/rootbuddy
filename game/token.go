package game

type TokenType int

const (
	TokenKeep TokenType = iota
	TokenSympathy
)

type Token struct {
	Faction Faction
	Type    TokenType
}
