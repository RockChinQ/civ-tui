package game

import (
	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/i18n"
)

func (g *Game) EndTurn() []GameMessage {
	var msgs []GameMessage

	for _, city := range g.Cities {
		civ := g.GetCiv(city.CivID)
		if civ == nil {
			continue
		}
		tile := g.Map.GetTile(city.X, city.Y)
		prevPop := city.Population
		built, _ := city.ProcessTurn(tile)
		isPlayer := city.CivID == 1
		if city.Population > prevPop {
			msgs = append(msgs, GameMessage{Text: i18n.Tf("%s grew to population %d", city.Name, city.Population), IsPlayer: isPlayer})
		}
		if built != nil {
			if built.IsUnit {
				g.AddUnit(built.UnitType, city.CivID, city.X, city.Y)
				msgs = append(msgs, GameMessage{Text: i18n.Tf("%s trained %s", city.Name, i18n.T(built.Name)), IsPlayer: isPlayer})
			} else {
				city.Buildings[built.BuildingType] = true
				msgs = append(msgs, GameMessage{Text: i18n.Tf("%s built %s", city.Name, i18n.T(built.Name)), IsPlayer: isPlayer})
			}
		}
		civ.Gold += city.GoldYield(tile)
		civ.Science += city.ScienceYield()
	}

	// Apply building maintenance costs
	for _, civ := range g.Civs {
		if !civ.IsAlive {
			continue
		}
		for _, city := range g.Cities {
			if city.CivID != civ.ID {
				continue
			}
			for bt := range city.Buildings {
				bdef := model.BuildingDefs[bt]
				civ.Gold -= bdef.Maintenance
			}
		}
	}

	for _, civ := range g.Civs {
		if civ.Researching != "" {
			numCities := 0
			for _, c := range g.Cities {
				if c.CivID == civ.ID {
					numCities++
				}
			}
			if numCities < 1 {
				numCities = 1
			}
			completed := civ.ProcessResearch(civ.Science/numCities, model.AllTechs)
			if completed != "" {
				msgs = append(msgs, GameMessage{Text: i18n.Tf("%s discovered %s!", i18n.T(civ.Name), i18n.T(completed)), IsPlayer: civ.ID == 1})
			}
		}
	}

	// Healing
	for _, u := range g.Units {
		if !u.IsAlive() {
			continue
		}
		if u.HP < u.MaxHP {
			city := g.GetCityAt(u.X, u.Y)
			if city != nil && city.CivID == u.CivID {
				u.HP += 2
				if u.HP > u.MaxHP {
					u.HP = u.MaxHP
				}
			} else if u.Waiting {
				u.HP++
				if u.HP > u.MaxHP {
					u.HP = u.MaxHP
				}
			}
		}
	}

	// Worker improvement processing
	for _, u := range g.Units {
		if !u.IsAlive() || u.Type != model.UnitWorker {
			continue
		}
		if u.BuildingImprovement != model.ImprovementNone {
			u.ImprovementTurnsLeft--
			if u.ImprovementTurnsLeft <= 0 {
				t := g.Map.GetTile(u.X, u.Y)
				if t != nil {
					t.Improvement = u.BuildingImprovement
					impName := model.Improvements[u.BuildingImprovement].Name
					msgs = append(msgs, GameMessage{Text: i18n.Tf("Worker built %s at (%d,%d)", i18n.T(impName), u.X, u.Y), IsPlayer: u.CivID == 1})
				}
				u.BuildingImprovement = model.ImprovementNone
				u.ImprovementTurnsLeft = 0
			}
		}
	}

	for _, u := range g.Units {
		if u.IsAlive() {
			u.ResetMoves()
		}
	}

	// Re-apply the waiting flag for units that are still performing a multi-turn
	// task (e.g. workers building improvements). ResetMoves clears Waiting, but
	// busy units must not be manually movable or auto-moved until the task ends.
	for _, u := range g.Units {
		if u.IsAlive() && u.IsBusy() {
			u.Waiting = true
		}
	}

	// Auto-movement: advance player units that have a movement destination set.
	// Each unit uses its fresh movement points to travel as far as possible along
	// the shortest path toward the destination this turn.
	for _, u := range g.Units {
		if !u.IsAlive() || u.CivID != 1 || !u.HasDest {
			continue
		}
		if u.IsBusy() {
			continue // task in progress – movement deferred
		}
		path := g.FindPath(u, u.DestX, u.DestY)
		if len(path) == 0 {
			// Already at destination or no path exists
			if u.X != u.DestX || u.Y != u.DestY {
				msgs = append(msgs, GameMessage{Text: i18n.Tf("Unit %s cannot reach destination", i18n.T(model.UnitDefs[u.Type].Name)), IsPlayer: true})
			}
			u.HasDest = false
			continue
		}
		for i := 0; i < len(path) && u.MovesLeft > 0; i++ {
			next := path[i]
			prevX, prevY := u.X, u.Y
			moveMsg, ok := g.MoveUnit(u, next[0]-prevX, next[1]-prevY)
			if !ok {
				u.HasDest = false
				break
			}
			if moveMsg != "" {
				msgs = append(msgs, GameMessage{Text: moveMsg, IsPlayer: true})
			}
			// Combat may have fired but the unit did not physically advance
			// (e.g. both units survived). Cancel the destination so the unit
			// does not attack indefinitely.
			if u.X == prevX && u.Y == prevY {
				u.HasDest = false
				break
			}
			if u.X == u.DestX && u.Y == u.DestY {
				u.HasDest = false
				msgs = append(msgs, GameMessage{Text: i18n.Tf("Unit %s arrived at (%d,%d)", i18n.T(model.UnitDefs[u.Type].Name), u.DestX, u.DestY), IsPlayer: true})
				break
			}
		}
	}

	aiMsgs := g.RunAI()
	for _, m := range aiMsgs {
		msgs = append(msgs, GameMessage{Text: m})
	}

	g.Turn++

	if g.Turn > g.MaxTurns {
		g.State = StateDraw
		msgs = append(msgs, GameMessage{Text: i18n.T("Turn limit reached! Game over."), IsPlayer: true})
	}

	g.CheckVictory()

	g.Map.ResetVisibility()
	g.RevealForCiv(1)

	g.PurgeDeadUnits()

	for _, m := range msgs {
		g.Messages = append(g.Messages, m)
	}
	if len(g.Messages) > 50 {
		g.Messages = g.Messages[len(g.Messages)-50:]
	}

	return msgs
}

func (g *Game) PurgeDeadUnits() {
	alive := make([]*model.Unit, 0, len(g.Units))
	for _, u := range g.Units {
		if u.IsAlive() {
			alive = append(alive, u)
		}
	}
	g.Units = alive
}

func (g *Game) RevealForCiv(civID int) {
	for _, u := range g.Units {
		if u.CivID == civID && u.IsAlive() {
			g.Map.Reveal(u.X, u.Y, model.VisionRadius)
		}
	}
	for _, c := range g.Cities {
		if c.CivID == civID {
			g.Map.Reveal(c.X, c.Y, model.VisionRadius)
		}
	}
}

func (g *Game) CheckAlive() {
	for _, civ := range g.Civs {
		if !civ.IsAlive {
			continue
		}
		hasUnits := false
		hasCities := false
		for _, u := range g.Units {
			if u.CivID == civ.ID && u.IsAlive() {
				hasUnits = true
				break
			}
		}
		for _, c := range g.Cities {
			if c.CivID == civ.ID {
				hasCities = true
				break
			}
		}
		if !hasUnits && !hasCities {
			civ.IsAlive = false
		}
	}
	g.CheckVictory()
}

func (g *Game) CheckVictory() {
	playerAlive := false
	enemyAlive := false
	for _, c := range g.Civs {
		if c.IsPlayer && c.IsAlive {
			playerAlive = true
		}
		if !c.IsPlayer && c.IsAlive {
			enemyAlive = true
		}
	}
	if !playerAlive {
		g.State = StateDefeat
		g.AddPlayerMessage(i18n.T("You have been defeated!"))
		return
	}
	if !enemyAlive {
		g.State = StateVictory
		g.AddPlayerMessage(i18n.T("Domination Victory! You conquered all enemies!"))
		return
	}
	playerCiv := g.GetCiv(1)
	if playerCiv != nil {
		allDone := true
		for _, t := range model.AllTechs {
			if !playerCiv.Techs[t.Name] {
				allDone = false
				break
			}
		}
		if allDone {
			g.State = StateVictory
			g.AddPlayerMessage(i18n.T("Science Victory! You researched all technologies!"))
		}
	}
}
