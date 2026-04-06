package game

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/game/worldmap"
)

type GameState int

const (
	StateRunning GameState = iota
	StateVictory
	StateDefeat
	StateDraw
)

type CivDef struct {
	Name      string
	CityNames []string
}

var CivDefs = []CivDef{
	{Name: "Roman Empire", CityNames: []string{"Rome", "Antium", "Neapolis", "Capua", "Pompeii", "Ravenna", "Milan", "Venice"}},
	{Name: "Mongols", CityNames: []string{"Karakorum", "Samarkand", "Bukhara", "Urgench", "Tabriz", "Alamut"}},
	{Name: "Egypt", CityNames: []string{"Memphis", "Thebes", "Alexandria", "Heliopolis", "Luxor", "Giza"}},
	{Name: "China", CityNames: []string{"Beijing", "Shanghai", "Nanjing", "Xian", "Hangzhou", "Chengdu"}},
	{Name: "Greece", CityNames: []string{"Athens", "Sparta", "Corinth", "Thessaloniki", "Olympia", "Delphi"}},
}

type GameOptions struct {
	NumAICivs  int
	MapSize    worldmap.MapSize
	Difficulty int
}

type Game struct {
	Map      *worldmap.GameMap
	Civs     []*model.Civ
	Units    []*model.Unit
	Cities   []*model.City
	Turn     int
	MaxTurns int
	State    GameState
	Messages []string
	NextID   int
	RandSeed int64
	Rand     *rand.Rand `json:"-"`
}

func NewGame(opts GameOptions) *Game {
	if opts.NumAICivs == 0 {
		opts.NumAICivs = 1
	}
	if opts.Difficulty == 0 {
		opts.Difficulty = 1
	}

	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	g := &Game{
		Map:      worldmap.NewGameMap(seed, opts.MapSize),
		MaxTurns: 200,
		Turn:     1,
		State:    StateRunning,
		Rand:     r,
		RandSeed: seed,
		NextID:   1,
	}

	// Create player civ
	player := model.NewCiv(1, CivDefs[0].Name, true)
	player.CityNames = CivDefs[0].CityNames
	g.Civs = []*model.Civ{player}

	// Create AI civs
	numAI := opts.NumAICivs
	if numAI > len(CivDefs)-1 {
		numAI = len(CivDefs) - 1
	}
	for i := 0; i < numAI; i++ {
		def := CivDefs[i+1]
		ai := model.NewCiv(i+2, def.Name, false)
		ai.CityNames = def.CityNames
		g.Civs = append(g.Civs, ai)
	}

	// Set all civs at war with each other
	for i := 0; i < len(g.Civs); i++ {
		for j := i + 1; j < len(g.Civs); j++ {
			g.DeclareWar(g.Civs[i], g.Civs[j])
		}
	}

	// Place player units
	px, py, ok := g.Map.FindPassableTile(r, 200)
	if !ok {
		px, py = g.Map.Width/4, g.Map.Height/2
	}
	g.AddUnit(model.UnitSettler, 1, px, py)
	g.AddUnit(model.UnitWarrior, 1, px+1, py)

	// Place AI units
	for i := 0; i < numAI; i++ {
		civID := i + 2
		ax, ay, ok2 := g.Map.FindPassableTile(r, 200)
		if !ok2 {
			ax, ay = 3*g.Map.Width/4, g.Map.Height/2
		}
		for abs(ax-px)+abs(ay-py) < 15 {
			ax, ay, ok2 = g.Map.FindPassableTile(r, 200)
			if !ok2 {
				ax, ay = 3*g.Map.Width/4, g.Map.Height/2
				break
			}
		}
		g.AddUnit(model.UnitSettler, civID, ax, ay)
		g.AddUnit(model.UnitWarrior, civID, ax+1, ay)
	}

	g.RevealForCiv(1)
	g.AddMessage("Turn 1: Welcome to Civ-TUI! Found a city with [F].")

	return g
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *Game) nextID() int {
	id := g.NextID
	g.NextID++
	return id
}

func (g *Game) AddUnit(utype model.UnitType, civID, x, y int) *model.Unit {
	if t := g.Map.GetTile(x, y); t != nil {
		terrain := model.Terrains[t.Terrain]
		if !terrain.Passable {
			for d := 1; d < 5; d++ {
				for dx := -d; dx <= d; dx++ {
					for dy := -d; dy <= d; dy++ {
						nx, ny := x+dx, y+dy
						if g.Map.InBounds(nx, ny) {
							nt := model.Terrains[g.Map.Tiles[ny][nx].Terrain]
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
	u := model.NewUnit(g.nextID(), utype, civID, x, y)
	g.Units = append(g.Units, u)
	return u
}

func (g *Game) AddCity(name string, civID, x, y int) *model.City {
	c := model.NewCity(g.nextID(), name, civID, x, y)
	g.Cities = append(g.Cities, c)
	g.Map.Tiles[y][x].Revealed = true
	return c
}

func (g *Game) GetCiv(id int) *model.Civ {
	for _, c := range g.Civs {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func (g *Game) GetUnitAt(x, y int) *model.Unit {
	for _, u := range g.Units {
		if u.IsAlive() && u.X == x && u.Y == y {
			return u
		}
	}
	return nil
}

func (g *Game) GetCityAt(x, y int) *model.City {
	for _, c := range g.Cities {
		if c.X == x && c.Y == y {
			return c
		}
	}
	return nil
}

func (g *Game) GetRelation(civA, civB int) model.RelationType {
	civ := g.GetCiv(civA)
	if civ == nil {
		return model.RelationPeace
	}
	return civ.Relations[civB]
}

func (g *Game) DeclareWar(civA, civB *model.Civ) {
	if civA != nil {
		civA.Relations[civB.ID] = model.RelationWar
	}
	if civB != nil {
		civB.Relations[civA.ID] = model.RelationWar
	}
}

func (g *Game) MakePeace(civA, civB *model.Civ) {
	if civA != nil {
		civA.Relations[civB.ID] = model.RelationPeace
	}
	if civB != nil {
		civB.Relations[civA.ID] = model.RelationPeace
	}
}

func (g *Game) MoveUnit(u *model.Unit, dx, dy int) (msg string, ok bool) {
	nx, ny := u.X+dx, u.Y+dy
	if !g.Map.InBounds(nx, ny) {
		return "Can't move there", false
	}
	t := model.Terrains[g.Map.Tiles[ny][nx].Terrain]
	if !t.Passable {
		return "Terrain not passable", false
	}
	if t.MoveCost > u.MovesLeft {
		return "Not enough movement", false
	}

	targetUnit := g.GetUnitAt(nx, ny)
	if targetUnit != nil && targetUnit.CivID != u.CivID {
		if g.GetRelation(u.CivID, targetUnit.CivID) == model.RelationWar {
			msg = g.Combat(u, targetUnit)
			if u.IsAlive() {
				u.MovesLeft = 0
			}
			return msg, true
		}
		return "Not at war", false
	}

	targetCity := g.GetCityAt(nx, ny)
	if targetCity != nil && targetCity.CivID != u.CivID {
		if g.GetRelation(u.CivID, targetCity.CivID) == model.RelationWar {
			msg = g.AttackCity(u, targetCity)
			return msg, true
		}
		return "Not at war", false
	}

	u.X, u.Y = nx, ny
	u.MovesLeft -= t.MoveCost
	if u.MovesLeft < 0 {
		u.MovesLeft = 0
	}

	if u.CivID == 1 {
		g.Map.Reveal(nx, ny, model.VisionRadius)
	}

	return "", true
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

	result := model.UnitDefs[attacker.Type].Name + " attacks " + model.UnitDefs[defender.Type].Name

	if !defender.IsAlive() {
		g.RemoveUnit(defender)
		result += " → killed!"
		attacker.XP += 2
		g.levelUp(attacker)
		g.CheckAlive()
	} else if !attacker.IsAlive() {
		g.RemoveUnit(attacker)
		result += " → attacker killed!"
		defender.XP++
		g.levelUp(defender)
		g.CheckAlive()
	} else {
		result += " → both damaged"
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

	result := model.UnitDefs[attacker.Type].Name + " ranged attacks " + model.UnitDefs[target.Type].Name
	if !target.IsAlive() {
		g.RemoveUnit(target)
		result += " → killed!"
		attacker.XP += 2
		g.levelUp(attacker)
		g.CheckAlive()
	} else {
		result += " → hit!"
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
		return "Captured " + city.Name + "!"
	}
	return "Attacked " + city.Name
}

func (g *Game) RemoveUnit(u *model.Unit) {
	u.HP = 0
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
		g.AddMessage("You have been defeated!")
		return
	}
	if !enemyAlive {
		g.State = StateVictory
		g.AddMessage("Domination Victory! You conquered all enemies!")
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
			g.AddMessage("Science Victory! You researched all technologies!")
		}
	}
}

func (g *Game) EndTurn() []string {
	var msgs []string

	for _, city := range g.Cities {
		civ := g.GetCiv(city.CivID)
		if civ == nil {
			continue
		}
		tile := g.Map.GetTile(city.X, city.Y)
		built, msg := city.ProcessTurn(tile)
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
				msgs = append(msgs, civ.Name+" discovered "+completed+"!")
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
					msgs = append(msgs, "Worker built "+impName+" at ("+strconv.Itoa(u.X)+","+strconv.Itoa(u.Y)+")")
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

	aiMsgs := g.RunAI()
	msgs = append(msgs, aiMsgs...)

	g.Turn++

	if g.Turn > g.MaxTurns {
		g.State = StateDraw
		msgs = append(msgs, "Turn limit reached! Game over.")
	}

	g.CheckVictory()

	g.Map.ResetVisibility()
	g.RevealForCiv(1)

	g.PurgeDeadUnits()

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

func (g *Game) PlayerUnitsWithMoves() []*model.Unit {
	var result []*model.Unit
	for _, u := range g.Units {
		if u.CivID == 1 && u.IsAlive() && u.HasMoves() {
			result = append(result, u)
		}
	}
	return result
}

func (g *Game) FoundCity(u *model.Unit, cityNames []string) (string, bool) {
	if u.Type != model.UnitSettler {
		return "Only Settlers can found cities", false
	}
	t := g.Map.GetTile(u.X, u.Y)
	if t == nil || !model.Terrains[t.Terrain].Passable {
		return "Cannot found city here", false
	}
	for _, c := range g.Cities {
		if abs(c.X-u.X)+abs(c.Y-u.Y) < model.MinCityDistance {
			return "Too close to another city", false
		}
	}

	name := g.civCityName(u.CivID)
	city := g.AddCity(name, u.CivID, u.X, u.Y)
	g.Map.Reveal(u.X, u.Y, model.VisionRadius)
	u.HP = 0
	return "Founded " + city.Name + "!", true
}

func (g *Game) civCityName(civID int) string {
	civ := g.GetCiv(civID)
	idx := 0
	for _, c := range g.Cities {
		if c.CivID == civID {
			idx++
		}
	}
	if civ != nil && idx < len(civ.CityNames) {
		return civ.CityNames[idx]
	}
	return "City " + strconv.Itoa(idx+1)
}
