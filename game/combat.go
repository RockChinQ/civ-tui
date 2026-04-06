package game

import (
	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/game/worldmap"
	"github.com/RockChinQ/civ-tui/i18n"
)

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

	atkName := i18n.T(model.UnitDefs[attacker.Type].Name)
	defName := i18n.T(model.UnitDefs[defender.Type].Name)
	var result string

	if !defender.IsAlive() {
		g.RemoveUnit(defender)
		result = i18n.Tf("%s attacks %s → killed!", atkName, defName)
		attacker.XP += 2
		g.levelUp(attacker)
		g.CheckAlive()
	} else if !attacker.IsAlive() {
		g.RemoveUnit(attacker)
		result = i18n.Tf("%s attacks %s → attacker killed!", atkName, defName)
		defender.XP++
		g.levelUp(defender)
		g.CheckAlive()
	} else {
		result = i18n.Tf("%s attacks %s → both damaged", atkName, defName)
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

	atkName := i18n.T(model.UnitDefs[attacker.Type].Name)
	tgtName := i18n.T(model.UnitDefs[target.Type].Name)
	var result string
	if !target.IsAlive() {
		g.RemoveUnit(target)
		result = i18n.Tf("%s ranged attacks %s → killed!", atkName, tgtName)
		attacker.XP += 2
		g.levelUp(attacker)
		g.CheckAlive()
	} else {
		result = i18n.Tf("%s ranged attacks %s → hit!", atkName, tgtName)
		attacker.XP++
	}
	return result
}

func (g *Game) AttackCity(u *model.Unit, city *model.City) string {
	dmg := u.Attack + g.Rand.Intn(3)
	dmg = max(1, dmg-city.Defense/2)
	city.HP -= dmg
	u.MovesLeft = 0

	if city.HP <= 0 {
		city.CivID = u.CivID
		city.HP = city.MaxHP / 2
		g.CheckAlive()
		return i18n.Tf("Captured %s!", city.Name)
	}
	return i18n.Tf("Attacked %s", city.Name)
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
