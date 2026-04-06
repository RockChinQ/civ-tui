package game

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/game/worldmap"
	"github.com/RockChinQ/civ-tui/i18n"
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
	g.AddMessage(i18n.T("Turn 1: Welcome to Civ-TUI! Found a city with [F]."))

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
		return i18n.T("Can't move there"), false
	}
	t := model.Terrains[g.Map.Tiles[ny][nx].Terrain]
	if !t.Passable {
		return i18n.T("Terrain not passable"), false
	}
	if t.MoveCost > u.MovesLeft {
		return i18n.T("Not enough movement"), false
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
		return i18n.T("Not at war"), false
	}

	// Prevent friendly unit stacking
	if targetUnit != nil && targetUnit.CivID == u.CivID {
		return i18n.T("Tile occupied by friendly unit"), false
	}

	targetCity := g.GetCityAt(nx, ny)
	if targetCity != nil && targetCity.CivID != u.CivID {
		if g.GetRelation(u.CivID, targetCity.CivID) == model.RelationWar {
			msg = g.AttackCity(u, targetCity)
			return msg, true
		}
		return i18n.T("Not at war"), false
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

func (g *Game) RemoveUnit(u *model.Unit) {
	u.HP = 0
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
		return i18n.T("Only Settlers can found cities"), false
	}
	t := g.Map.GetTile(u.X, u.Y)
	if t == nil || !model.Terrains[t.Terrain].Passable {
		return i18n.T("Cannot found city here"), false
	}
	for _, c := range g.Cities {
		if abs(c.X-u.X)+abs(c.Y-u.Y) < model.MinCityDistance {
			return i18n.T("Too close to another city"), false
		}
	}

	name := g.civCityName(u.CivID)
	city := g.AddCity(name, u.CivID, u.X, u.Y)
	g.Map.Reveal(u.X, u.Y, model.VisionRadius)
	u.HP = 0
	return i18n.Tf("Founded %s!", city.Name), true
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
