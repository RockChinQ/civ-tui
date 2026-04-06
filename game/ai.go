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

// AI research: prefer techs that unlock units
if civ.Researching == "" {
available := AvailableTechs(civ.Techs)
if len(available) > 0 {
chosen := available[0]
for _, t := range available {
for _, ud := range UnitDefs {
if ud.RequiresTech == t.Name {
chosen = t
break
}
}
}
civ.ResearchTech(chosen)
}
}

// AI diplomacy
g.aiDiplomacy(civ)

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

for _, city := range g.Cities {
if city.CivID != civ.ID {
continue
}
if len(city.ProductionQ) == 0 {
unitTypes := AvailableUnits(civ.Techs)
// Remove settler and worker from AI production choices
var militaryTypes []UnitType
for _, ut := range unitTypes {
if ut != UnitSettler && ut != UnitWorker {
militaryTypes = append(militaryTypes, ut)
}
}
if len(militaryTypes) == 0 {
militaryTypes = []UnitType{UnitWarrior}
}
unitCount := 0
for _, u := range g.Units {
if u.CivID == civ.ID && u.IsAlive() {
unitCount++
}
}
// Pick a unit type based on count
idx := 0
if unitCount >= 5 && len(militaryTypes) > 2 {
idx = len(militaryTypes) - 1
} else if unitCount >= 3 && len(militaryTypes) > 1 {
idx = len(militaryTypes) / 2
}
unitType := militaryTypes[idx]
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

func (g *Game) aiDiplomacy(civ *Civ) {
myUnits := 0
for _, u := range g.Units {
if u.CivID == civ.ID && u.IsAlive() {
myUnits++
}
}

for _, other := range g.Civs {
if other.ID == civ.ID || !other.IsAlive {
continue
}
relation := civ.Relations[other.ID]
if relation == RelationWar {
enemyUnits := 0
for _, u := range g.Units {
if u.CivID == other.ID && u.IsAlive() {
enemyUnits++
}
}
// 5% chance to make peace if outnumbered
if myUnits < enemyUnits && g.Rand.Intn(100) < 5 {
g.MakePeace(civ, other)
}
} else {
enemyUnits := 0
for _, u := range g.Units {
if u.CivID == other.ID && u.IsAlive() {
enemyUnits++
}
}
// 2% chance to declare war if we have more units
if myUnits > enemyUnits && g.Rand.Intn(100) < 2 {
g.DeclareWar(civ, other)
}
}
}
}

func (g *Game) aiSettlerAction(civ *Civ, u *Unit, msgs *[]string) bool {
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
return g.aiMoveRandom(u)
}

func (g *Game) aiMilitaryAction(civ *Civ, u *Unit, msgs *[]string) bool {
bestDist := 999
bestX, bestY := -1, -1

for _, pu := range g.Units {
if pu.CivID == civ.ID || !pu.IsAlive() {
continue
}
if g.GetRelation(civ.ID, pu.CivID) != RelationWar {
continue
}
d := abs(pu.X-u.X) + abs(pu.Y-u.Y)
if d < bestDist {
bestDist = d
bestX, bestY = pu.X, pu.Y
}
}
for _, pc := range g.Cities {
if g.GetRelation(civ.ID, pc.CivID) != RelationWar {
continue
}
d := abs(pc.X-u.X) + abs(pc.Y-u.Y)
if d < bestDist {
bestDist = d
bestX, bestY = pc.X, pc.Y
}
}

if bestX == -1 {
return g.aiMoveRandom(u)
}

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
