package engine

import "github.com/imdehydrated/rootbuddy/game"

func hasMarquiseKeepToken(clearing game.Clearing) bool {
	for _, token := range clearing.Tokens {
		if token.Faction == game.Marquise && token.Type == game.TokenKeep {
			return true
		}
	}

	return false
}
