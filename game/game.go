package game

import (
	"math/rand"
	"time"
)

type GameState int

const (
	StateRunning GameState = iota
	StateVictory
	StateDefeat
	StateDraw
)

type Game struct {
	Map      *GameMap
	Civs     []*Civ
	Units    []*Unit
	Cities   []*City
	Turn     int
	MaxTurns int
	State    GameState
	Messages []string
	NextID   int
	Rand     *rand.Rand
}

func NewGame() *Game {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	g := &Game{
		Map:      NewGameMap(seed),
		MaxTurns: 200,
		Turn:     1,
		State:    StateRunning,
		Rand:     r,
		NextID:   1,
	}

	// Create civilizations
	player := NewCiv(1, "Roman Empire", true)
	ai := NewCiv(2, "Mongols", false)
	g.Civs = []*Civ{player, ai}

	// Place player units
	px, py, ok := g.Map.FindPassableTile(r, 200)
	if !ok {
		px, py = MapWidth/4, MapHeight/2
	}
	g.AddUnit(UnitSettler, 1, px, py)
	g.AddUnit(UnitWarrior, 1, px+1, py)

	// Place AI units
	ax, ay, ok := g.Map.FindPassableTile(r, 200)
	if !ok {
		ax, ay = 3*MapWidth/4, MapHeight/2
	}
	// Ensure AI is not too close to player
	for abs(ax-px)+abs(ay-py) < 15 {
		ax, ay, ok = g.Map.FindPassableTile(r, 200)
		if !ok {
			ax, ay = 3*MapWidth/4, MapHeight/2
			break
		}
	}
	g.AddUnit(UnitSettler, 2, ax, ay)
	g.AddUnit(UnitWarrior, 2, ax+1, ay)

	// Reveal initial area for player
	g.RevealForCiv(1)

	g.AddMessage("Turn 1: Welcome to Civ-TUI! Found a city with [F].")

	return g
}

func (g *Game) nextID() int {
	id := g.NextID
	g.NextID++
	return id
}

func (g *Game) AddUnit(utype UnitType, civID, x, y int) *Unit {
	// Make sure position is passable
	if t := g.Map.GetTile(x, y); t != nil {
		terrain := Terrains[t.Terrain]
		if !terrain.Passable {
			// Find nearby passable tile
			for d := 1; d < 5; d++ {
				for dx := -d; dx <= d; dx++ {
					for dy := -d; dy <= d; dy++ {
						nx, ny := x+dx, y+dy
						if g.Map.InBounds(nx, ny) {
							nt := Terrains[g.Map.Tiles[ny][nx].Terrain]
							if nt.Passable {
								x, y = nx, ny
								goto found
							}
						}
					}
				}
			}
		found:
		}
	}
	u := NewUnit(g.nextID(), utype, civID, x, y)
	g.Units = append(g.Units, u)
	return u
}

func (g *Game) AddCity(name string, civID, x, y int) *City {
	c := NewCity(g.nextID(), name, civID, x, y)
	g.Cities = append(g.Cities, c)
	g.Map.Tiles[y][x].Revealed = true
	return c
}

func (g *Game) GetCiv(id int) *Civ {
	for _, c := range g.Civs {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func (g *Game) GetUnitAt(x, y int) *Unit {
	for _, u := range g.Units {
		if u.IsAlive() && u.X == x && u.Y == y {
			return u
		}
	}
	return nil
}

func (g *Game) GetCityAt(x, y int) *City {
	for _, c := range g.Cities {
		if c.X == x && c.Y == y {
			return c
		}
	}
	return nil
}

func (g *Game) MoveUnit(u *Unit, dx, dy int) (msg string, ok bool) {
	nx, ny := u.X+dx, u.Y+dy
	if !g.Map.InBounds(nx, ny) {
		return "Can't move there", false
	}
	t := Terrains[g.Map.Tiles[ny][nx].Terrain]
	if !t.Passable {
		return "Terrain not passable", false
	}
	if t.MoveCost > u.MovesLeft {
		return "Not enough movement", false
	}

	// Check for enemy unit at destination
	targetUnit := g.GetUnitAt(nx, ny)
	if targetUnit != nil && targetUnit.CivID != u.CivID {
		msg = g.Combat(u, targetUnit)
		if u.IsAlive() {
			u.MovesLeft = 0
		}
		return msg, true
	}

	// Check for enemy city at destination
	targetCity := g.GetCityAt(nx, ny)
	if targetCity != nil && targetCity.CivID != u.CivID {
		msg = g.AttackCity(u, targetCity)
		return msg, true
	}

	u.X, u.Y = nx, ny
	u.MovesLeft -= t.MoveCost
	if u.MovesLeft < 0 {
		u.MovesLeft = 0
	}

	if u.CivID == 1 {
		g.Map.Reveal(nx, ny, 3)
	}

	return "", true
}

func (g *Game) Combat(attacker, defender *Unit) string {
	atkDmg := attacker.Attack + g.Rand.Intn(3)
	defDmg := defender.Defense + g.Rand.Intn(3)

	defender.HP -= atkDmg
	attacker.HP -= defDmg

	result := UnitDefs[attacker.Type].Name + " attacks " + UnitDefs[defender.Type].Name

	if !defender.IsAlive() {
		g.RemoveUnit(defender)
		result += " → killed!"
		g.CheckAlive()
	} else if !attacker.IsAlive() {
		g.RemoveUnit(attacker)
		result += " → attacker killed!"
		g.CheckAlive()
	} else {
		result += " → both damaged"
	}
	return result
}

func (g *Game) AttackCity(u *Unit, city *City) string {
	dmg := u.Attack + g.Rand.Intn(3)
	city.HP -= dmg
	u.MovesLeft = 0

	if city.HP <= 0 {
		// Capture city
		city.CivID = u.CivID
		city.HP = city.MaxHP / 2
		g.CheckAlive()
		return "Captured " + city.Name + "!"
	}
	return "Attacked " + city.Name
}

func (g *Game) RemoveUnit(u *Unit) {
	u.HP = 0
}

func (g *Game) RevealForCiv(civID int) {
	for _, u := range g.Units {
		if u.CivID == civID && u.IsAlive() {
			g.Map.Reveal(u.X, u.Y, 3)
		}
	}
	for _, c := range g.Cities {
		if c.CivID == civID {
			g.Map.Reveal(c.X, c.Y, 3)
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
		g.AddMessage("You have been defeated!")
		return
	}
	if !enemyAlive {
		g.State = StateVictory
		g.AddMessage("Domination Victory! You conquered all enemies!")
		return
	}
	// Science victory
	playerCiv := g.GetCiv(1)
	if playerCiv != nil {
		allDone := true
		for _, t := range AllTechs {
			if !playerCiv.Techs[t.Name] {
				allDone = false
				break
			}
		}
		if allDone {
			g.State = StateVictory
			g.AddMessage("Science Victory! You researched all technologies!")
		}
	}
}

func (g *Game) EndTurn() []string {
	var msgs []string

	// Process cities
	for _, city := range g.Cities {
		civ := g.GetCiv(city.CivID)
		if civ == nil {
			continue
		}
		built, msg := city.ProcessTurn(g.Map)
		if msg != "" {
			msgs = append(msgs, msg)
		}
		if built != nil {
			if built.IsUnit {
				g.AddUnit(built.UnitType, city.CivID, city.X, city.Y)
				msgs = append(msgs, city.Name+" trained "+built.Name)
			} else {
				city.Buildings[built.BuildingType] = true
				msgs = append(msgs, city.Name+" built "+built.Name)
			}
		}
		civ.Gold += city.GoldYield(g.Map)
		civ.Science += city.ScienceYield()
	}

	// Process research
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
			completed := civ.ProcessResearch(civ.Science/numCities, AllTechs)
			if completed != "" {
				msgs = append(msgs, civ.Name+" discovered "+completed+"!")
			}
		}
	}

	// Reset unit moves
	for _, u := range g.Units {
		if u.IsAlive() {
			u.ResetMoves()
		}
	}

	// AI turn
	aiMsgs := g.RunAI()
	msgs = append(msgs, aiMsgs...)

	g.Turn++

	if g.Turn > g.MaxTurns {
		g.State = StateDraw
		msgs = append(msgs, "Turn limit reached! Game over.")
	}

	g.CheckVictory()

	// Reset map visibility
	g.Map.ResetVisibility()
	g.RevealForCiv(1)

	for _, m := range msgs {
		g.AddMessage(m)
	}

	return msgs
}

func (g *Game) AddMessage(msg string) {
	g.Messages = append(g.Messages, msg)
	if len(g.Messages) > 50 {
		g.Messages = g.Messages[len(g.Messages)-50:]
	}
}

func (g *Game) PlayerUnitsWithMoves() []*Unit {
	var result []*Unit
	for _, u := range g.Units {
		if u.CivID == 1 && u.IsAlive() && u.HasMoves() {
			result = append(result, u)
		}
	}
	return result
}

func (g *Game) FoundCity(u *Unit, cityNames []string) (string, bool) {
	if u.Type != UnitSettler {
		return "Only Settlers can found cities", false
	}
	// Check not on water
	t := g.Map.GetTile(u.X, u.Y)
	if t == nil || !Terrains[t.Terrain].Passable {
		return "Cannot found city here", false
	}
	// Check no other city nearby
	for _, c := range g.Cities {
		if abs(c.X-u.X)+abs(c.Y-u.Y) < 3 {
			return "Too close to another city", false
		}
	}
	civ := g.GetCiv(u.CivID)
	name := "New City"
	if civ != nil {
		if civ.IsPlayer {
			names := []string{"Rome", "Antium", "Neapolis", "Capua", "Pompeii", "Ravenna", "Milan", "Venice"}
			idx := len(g.Cities)
			if idx < len(names) {
				name = names[idx]
			}
		} else {
			names := []string{"Karakorum", "Samarkand", "Bukhara", "Urgench", "Tabriz", "Alamut"}
			idx := 0
			for _, c := range g.Cities {
				if c.CivID == u.CivID {
					idx++
				}
			}
			if idx < len(names) {
				name = names[idx]
			}
		}
	}
	city := g.AddCity(name, u.CivID, u.X, u.Y)
	g.Map.Reveal(u.X, u.Y, 3)
	// Remove settler
	u.HP = 0
	return "Founded " + city.Name + "!", true
}
