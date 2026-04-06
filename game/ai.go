package game

func (g *Game) RunAI() []string {
	var msgs []string
	for _, civ := range g.Civs {
		if civ.IsPlayer || !civ.IsAlive {
			continue
		}
		msgs = append(msgs, g.runCivAI(civ)...)
	}
	return msgs
}

func (g *Game) runCivAI(civ *Civ) []string {
	var msgs []string

	// Auto-research
	if civ.Researching == "" {
		available := AvailableTechs(civ.Techs)
		if len(available) > 0 {
			civ.ResearchTech(available[g.Rand.Intn(len(available))])
		}
	}

	for _, u := range g.Units {
		if u.CivID != civ.ID || !u.IsAlive() {
			continue
		}
		for u.HasMoves() {
			var acted bool
			switch u.Type {
			case UnitSettler:
				acted = g.aiSettlerAction(civ, u, &msgs)
			default:
				acted = g.aiMilitaryAction(civ, u, &msgs)
			}
			if !acted {
				u.MovesLeft = 0
			}
		}
	}

	// City production
	for _, city := range g.Cities {
		if city.CivID != civ.ID {
			continue
		}
		if len(city.ProductionQ) == 0 {
			unitCount := 0
			for _, u := range g.Units {
				if u.CivID == civ.ID && u.IsAlive() {
					unitCount++
				}
			}
			unitType := UnitWarrior
			if unitCount >= 5 {
				unitType = UnitSpearman
			} else if unitCount >= 3 {
				unitType = UnitArcher
			}
			stats := UnitDefs[unitType]
			city.ProductionQ = append(city.ProductionQ, ProductionItem{
				IsUnit:   true,
				UnitType: unitType,
				Name:     stats.Name,
				Cost:     stats.ProductionCost,
			})
		}
	}

	return msgs
}

func (g *Game) aiSettlerAction(civ *Civ, u *Unit, msgs *[]string) bool {
	// Check if good spot to found city
	tooClose := false
	for _, c := range g.Cities {
		if abs(c.X-u.X)+abs(c.Y-u.Y) < 4 {
			tooClose = true
			break
		}
	}
	t := g.Map.GetTile(u.X, u.Y)
	if !tooClose && t != nil && Terrains[t.Terrain].Passable {
		msg, ok := g.FoundCity(u, nil)
		if ok {
			*msgs = append(*msgs, msg)
			return false
		}
	}
	// Move randomly
	return g.aiMoveRandom(u)
}

func (g *Game) aiMilitaryAction(civ *Civ, u *Unit, msgs *[]string) bool {
	// Find closest player unit or city
	bestDist := 999
	bestX, bestY := -1, -1

	for _, pu := range g.Units {
		if pu.CivID == 1 && pu.IsAlive() {
			d := abs(pu.X-u.X) + abs(pu.Y-u.Y)
			if d < bestDist {
				bestDist = d
				bestX, bestY = pu.X, pu.Y
			}
		}
	}
	for _, pc := range g.Cities {
		if pc.CivID == 1 {
			d := abs(pc.X-u.X) + abs(pc.Y-u.Y)
			if d < bestDist {
				bestDist = d
				bestX, bestY = pc.X, pc.Y
			}
		}
	}

	if bestX == -1 {
		return g.aiMoveRandom(u)
	}

	// Move toward target
	dx, dy := 0, 0
	if bestX > u.X {
		dx = 1
	} else if bestX < u.X {
		dx = -1
	}
	if bestY > u.Y {
		dy = 1
	} else if bestY < u.Y {
		dy = -1
	}

	// Try horizontal first, then vertical
	var dirs [][2]int
	if dx != 0 && dy != 0 {
		dirs = [][2]int{{dx, 0}, {0, dy}, {dx, dy}, {-dx, 0}, {0, -dy}}
	} else if dx != 0 {
		dirs = [][2]int{{dx, 0}, {0, 1}, {0, -1}, {-dx, 0}}
	} else {
		dirs = [][2]int{{0, dy}, {1, 0}, {-1, 0}, {0, -dy}}
	}

	for _, d := range dirs {
		msg, ok := g.MoveUnit(u, d[0], d[1])
		if ok {
			if msg != "" {
				*msgs = append(*msgs, msg)
			}
			return true
		}
	}
	return false
}

func (g *Game) aiMoveRandom(u *Unit) bool {
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	g.Rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
	for _, d := range dirs {
		_, ok := g.MoveUnit(u, d[0], d[1])
		if ok {
			return true
		}
	}
	return false
}
