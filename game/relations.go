package game

func AreCoalitionPartners(state GameState, a Faction, b Faction) bool {
	if !state.CoalitionActive {
		return false
	}

	partner := state.CoalitionPartner
	if partner == Vagabond {
		return false
	}

	return (a == Vagabond && b == partner) || (b == Vagabond && a == partner)
}

func AreEnemies(state GameState, a Faction, b Faction) bool {
	if a == b {
		return false
	}

	return !AreCoalitionPartners(state, a, b)
}

func VagabondHostileTo(state GameState, faction Faction) bool {
	if faction == Vagabond || !AreEnemies(state, Vagabond, faction) {
		return false
	}
	if state.Vagabond.Relationships == nil {
		return false
	}

	relationship, ok := state.Vagabond.Relationships[faction]
	return ok && relationship == RelHostile
}
