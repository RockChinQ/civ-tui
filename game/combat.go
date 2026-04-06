package game

import (
	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/game/worldmap"
	"github.com/RockChinQ/civ-tui/i18n"
)

// civName returns the translated civilization name for the given civ ID.
func (g *Game) civName(civID int) string {
	civ := g.GetCiv(civID)
	if civ != nil {
		return i18n.T(civ.Name)
	}
	return "?"
}

func (g *Game) Combat(attacker, defender *model.Unit) string {
	// Apply terrain defense bonus to reduce attacker damage
	defBonus := 0
	t := g.Map.GetTile(defender.X, defender.Y)
	if t != nil {
		defBonus = model.Terrains[t.Terrain].DefenseBonus
	}

	atkDmg := attacker.Attack + g.Rand.Intn(3)
	defDmg := defender.Defense + g.Rand.Intn(3)

	// Reduce attacker damage by terrain defense bonus percentage
	if defBonus > 0 {
		atkDmg = atkDmg - atkDmg*defBonus/100
		if atkDmg < 1 {
			atkDmg = 1
		}
	}

	defender.HP -= atkDmg
	attacker.HP -= defDmg

	atkCiv := g.civName(attacker.CivID)
	defCiv := g.civName(defender.CivID)
	atkName := i18n.T(model.UnitDefs[attacker.Type].Name)
	defName := i18n.T(model.UnitDefs[defender.Type].Name)
	x, y := defender.X, defender.Y
	var result string

	if !defender.IsAlive() {
		g.RemoveUnit(defender)
		result = i18n.Tf("[%s] %s attacks [%s] %s (%d,%d) → killed!", atkCiv, atkName, defCiv, defName, x, y)
		attacker.XP += 2
		g.levelUp(attacker)
		g.CheckAlive()
	} else if !attacker.IsAlive() {
		g.RemoveUnit(attacker)
		result = i18n.Tf("[%s] %s attacks [%s] %s (%d,%d) → attacker killed!", atkCiv, atkName, defCiv, defName, x, y)
		defender.XP++
		g.levelUp(defender)
		g.CheckAlive()
	} else {
		result = i18n.Tf("[%s] %s attacks [%s] %s (%d,%d) → both damaged", atkCiv, atkName, defCiv, defName, x, y)
		attacker.XP++
		defender.XP++
	}
	return result
}

func (g *Game) levelUp(u *model.Unit) {
	if u.XP >= 5 {
		u.XP -= 5
		u.Level++
		u.Attack++
	}
}

func (g *Game) RangedAttack(attacker, target *model.Unit) string {
	atkDmg := attacker.Attack + g.Rand.Intn(3)
	target.HP -= atkDmg
	attacker.MovesLeft = 0

	atkCiv := g.civName(attacker.CivID)
	tgtCiv := g.civName(target.CivID)
	atkName := i18n.T(model.UnitDefs[attacker.Type].Name)
	tgtName := i18n.T(model.UnitDefs[target.Type].Name)
	x, y := target.X, target.Y
	var result string
	if !target.IsAlive() {
		g.RemoveUnit(target)
		result = i18n.Tf("[%s] %s ranged attacks [%s] %s (%d,%d) → killed!", atkCiv, atkName, tgtCiv, tgtName, x, y)
		attacker.XP += 2
		g.levelUp(attacker)
		g.CheckAlive()
	} else {
		result = i18n.Tf("[%s] %s ranged attacks [%s] %s (%d,%d) → hit!", atkCiv, atkName, tgtCiv, tgtName, x, y)
		attacker.XP++
	}
	return result
}

func (g *Game) AttackCity(u *model.Unit, city *model.City) string {
	dmg := u.Attack + g.Rand.Intn(3)
	dmg = max(1, dmg-city.Defense/2)
	city.HP -= dmg
	u.MovesLeft = 0

	atkCiv := g.civName(u.CivID)
	defCiv := g.civName(city.CivID)
	atkName := i18n.T(model.UnitDefs[u.Type].Name)

	if city.HP <= 0 {
		city.CivID = u.CivID
		city.HP = city.MaxHP / 2
		g.CheckAlive()
		return i18n.Tf("[%s] %s captured [%s] %s (%d,%d)!", atkCiv, atkName, defCiv, city.Name, city.X, city.Y)
	}
	return i18n.Tf("[%s] %s attacked [%s] %s (%d,%d)", atkCiv, atkName, defCiv, city.Name, city.X, city.Y)
}

// ReachableTiles returns the set of tiles reachable by the given unit this turn.
// Uses BFS flood fill considering terrain movement costs and passability.
func (g *Game) ReachableTiles(u *model.Unit) map[[2]int]bool {
	reachable := make(map[[2]int]bool)
	if u == nil || !u.IsAlive() || !u.HasMoves() {
		return reachable
	}

	type state struct {
		x, y      int
		movesLeft int
	}

	visited := make(map[[2]int]int) // position -> best movesLeft seen
	queue := []state{{u.X, u.Y, u.MovesLeft}}
	visited[[2]int{u.X, u.Y}] = u.MovesLeft

	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, d := range dirs {
			nx, ny := cur.x+d[0], cur.y+d[1]
			if !g.Map.InBounds(nx, ny) {
				continue
			}
			t := model.Terrains[g.Map.Tiles[ny][nx].Terrain]
			if !t.Passable {
				continue
			}
			cost := t.MoveCost
			if cost > cur.movesLeft {
				continue
			}
			remaining := cur.movesLeft - cost

			pos := [2]int{nx, ny}
			if prev, ok := visited[pos]; ok && prev >= remaining {
				continue
			}

			// Check for blocking units
			blocker := g.GetUnitAt(nx, ny)
			if blocker != nil && blocker.CivID == u.CivID && blocker.ID != u.ID {
				continue // Can't move through friendly units
			}

			visited[pos] = remaining
			reachable[pos] = true
			if remaining > 0 {
				queue = append(queue, state{nx, ny, remaining})
			}
		}
	}

	// Remove the unit's current position from reachable
	delete(reachable, [2]int{u.X, u.Y})

	return reachable
}

// AbsDist is a convenience re-export for use within the game package.
func AbsDist(x1, y1, x2, y2 int) int {
	return worldmap.AbsDist(x1, y1, x2, y2)
}
