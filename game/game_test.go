package game

import (
	"testing"

	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/game/worldmap"
)

func defaultOpts() GameOptions {
	return GameOptions{NumAICivs: 1, MapSize: worldmap.MapSizeSmall, Difficulty: 1}
}

func TestCombat(t *testing.T) {
	g := NewGame(defaultOpts())

	// Find a player unit and an enemy unit
	var player, enemy *model.Unit
	for _, u := range g.Units {
		if u.CivID == 1 && u.Type == model.UnitWarrior {
			player = u
		}
		if u.CivID == 2 && u.Type == model.UnitWarrior {
			enemy = u
		}
	}
	if player == nil || enemy == nil {
		t.Skip("Could not find required units")
	}

	initialPlayerHP := player.HP
	initialEnemyHP := enemy.HP

	result := g.Combat(player, enemy)
	if result == "" {
		t.Error("Combat should return a result message")
	}

	// At least one unit should have taken damage or died
	if player.HP == initialPlayerHP && enemy.HP == initialEnemyHP {
		t.Error("Combat should have caused damage")
	}

	// XP should be awarded
	if player.IsAlive() && enemy.IsAlive() {
		if player.XP == 0 && enemy.XP == 0 {
			t.Error("XP should be awarded after combat")
		}
	}
}

func TestCityProductionYields(t *testing.T) {
	g := NewGame(defaultOpts())

	// Found a city for the player
	var settler *model.Unit
	for _, u := range g.Units {
		if u.CivID == 1 && u.Type == model.UnitSettler {
			settler = u
			break
		}
	}
	if settler == nil {
		t.Skip("No settler found")
	}

	// Ensure tile is passable before founding
	tile := g.Map.GetTile(settler.X, settler.Y)
	if tile == nil || !model.Terrains[tile.Terrain].Passable {
		t.Skip("Settler not on passable tile")
	}

	// Move settler away from other cities if needed
	_, ok := g.FoundCity(settler, nil)
	if !ok {
		t.Skip("Could not found city (too close to another city)")
	}

	// Find the city we just founded
	city := g.GetCityAt(settler.X, settler.Y)
	if city == nil {
		// settler position changes when it dies
		for _, c := range g.Cities {
			if c.CivID == 1 {
				city = c
			}
		}
	}
	if city == nil {
		t.Fatal("City not found after founding")
	}

	cityTile := g.Map.GetTile(city.X, city.Y)
	food := city.FoodYield(cityTile)
	if food <= 0 {
		t.Errorf("Food yield should be positive, got %d", food)
	}

	prod := city.ProductionYield(cityTile)
	if prod <= 0 {
		t.Errorf("Production yield should be positive, got %d", prod)
	}

	// Add granary and check food bonus
	city.Buildings[model.BuildingGranary] = true
	newFood := city.FoodYield(cityTile)
	if newFood <= food {
		t.Errorf("Granary should increase food yield: before=%d after=%d", food, newFood)
	}
}

func TestTechResearch(t *testing.T) {
	g := NewGame(defaultOpts())
	civ := g.GetCiv(1)
	if civ == nil {
		t.Fatal("Player civ not found")
	}

	available := model.AvailableTechs(civ.Techs)
	if len(available) == 0 {
		t.Fatal("No techs available at start")
	}

	tech := available[0]
	civ.ResearchTech(tech)
	if civ.Researching != tech.Name {
		t.Errorf("Expected researching %s, got %s", tech.Name, civ.Researching)
	}

	// Process enough research to complete
	completed := civ.ProcessResearch(tech.Cost, model.AllTechs)
	if completed != tech.Name {
		t.Errorf("Expected completed tech %s, got %s", tech.Name, completed)
	}
	if !civ.Techs[tech.Name] {
		t.Errorf("Tech %s should be in civ.Techs after completion", tech.Name)
	}
	if civ.Researching != "" {
		t.Error("civ.Researching should be empty after completion")
	}
}

func TestMapGeneration(t *testing.T) {
	for _, size := range []worldmap.MapSize{worldmap.MapSizeSmall, worldmap.MapSizeMedium, worldmap.MapSizeLarge} {
		cfg := worldmap.MapSizes[size]
		m := worldmap.NewGameMap(42, size)
		if m.Width != cfg.Width {
			t.Errorf("Map width: expected %d, got %d", cfg.Width, m.Width)
		}
		if m.Height != cfg.Height {
			t.Errorf("Map height: expected %d, got %d", cfg.Height, m.Height)
		}
		if len(m.Tiles) != m.Height {
			t.Errorf("Tiles rows: expected %d, got %d", m.Height, len(m.Tiles))
		}
		for y := 0; y < m.Height; y++ {
			if len(m.Tiles[y]) != m.Width {
				t.Errorf("Tiles cols at row %d: expected %d, got %d", y, m.Width, len(m.Tiles[y]))
			}
		}
		// Check that some passable tiles exist
		passable := 0
		for y := 0; y < m.Height; y++ {
			for x := 0; x < m.Width; x++ {
				if model.Terrains[m.Tiles[y][x].Terrain].Passable {
					passable++
				}
			}
		}
		if passable == 0 {
			t.Errorf("Map size %d has no passable tiles", size)
		}
	}
}

func TestUnitMovement(t *testing.T) {
	g := NewGame(defaultOpts())

	var warrior *model.Unit
	for _, u := range g.Units {
		if u.CivID == 1 && u.Type == model.UnitWarrior {
			warrior = u
			break
		}
	}
	if warrior == nil {
		t.Skip("No warrior found")
	}

	origX, origY := warrior.X, warrior.Y
	origMoves := warrior.MovesLeft

	// Try all 4 directions until one works
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	moved := false
	for _, d := range dirs {
		msg, ok := g.MoveUnit(warrior, d[0], d[1])
		if ok {
			moved = true
			if warrior.X == origX && warrior.Y == origY {
				// Only same position if combat occurred
				if msg == "" {
					t.Error("Unit position should change after successful move")
				}
			}
			if msg == "" && warrior.MovesLeft >= origMoves {
				t.Error("Movement should deduct move points")
			}
			break
		}
	}
	if !moved {
		t.Skip("Could not find a passable adjacent tile")
	}
}

func TestMovementBlockedByImpassable(t *testing.T) {
	g := NewGame(defaultOpts())

	// Create a warrior and place it manually next to an ocean tile
	warrior := g.AddUnit(model.UnitWarrior, 1, 1, 1)
	warrior.MovesLeft = warrior.MaxMoves

	// Find an ocean tile adjacent to the warrior
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for _, d := range dirs {
		nx, ny := warrior.X+d[0], warrior.Y+d[1]
		if g.Map.InBounds(nx, ny) {
			tile := g.Map.GetTile(nx, ny)
			if !model.Terrains[tile.Terrain].Passable {
				msg, ok := g.MoveUnit(warrior, d[0], d[1])
				if ok {
					t.Error("Should not be able to move onto impassable terrain")
				}
				if msg != "Terrain not passable" {
					t.Errorf("Expected 'Terrain not passable', got %q", msg)
				}
				return
			}
		}
	}
	t.Skip("No impassable terrain adjacent to test position")
}

func TestFriendlyUnitStacking(t *testing.T) {
	g := NewGame(defaultOpts())

	// Find a passable tile
	var px, py int
	for y := 5; y < g.Map.Height-5; y++ {
		for x := 5; x < g.Map.Width-5; x++ {
			tile := g.Map.GetTile(x, y)
			if model.Terrains[tile.Terrain].Passable {
				// Check adjacent tile too
				adj := g.Map.GetTile(x+1, y)
				if adj != nil && model.Terrains[adj.Terrain].Passable {
					px, py = x, y
					goto found
				}
			}
		}
	}
found:

	// Place two friendly warriors adjacent to each other
	w1 := g.AddUnit(model.UnitWarrior, 1, px, py)
	g.AddUnit(model.UnitWarrior, 1, px+1, py)
	w1.MovesLeft = w1.MaxMoves

	// Try to move w1 onto w2's tile
	msg, ok := g.MoveUnit(w1, 1, 0)
	if ok {
		t.Error("Should not be able to stack friendly units")
	}
	if msg != "Tile occupied by friendly unit" {
		t.Errorf("Expected 'Tile occupied by friendly unit', got %q", msg)
	}
}

func TestReachableTiles(t *testing.T) {
	g := NewGame(defaultOpts())

	var warrior *model.Unit
	for _, u := range g.Units {
		if u.CivID == 1 && u.Type == model.UnitWarrior {
			warrior = u
			break
		}
	}
	if warrior == nil {
		t.Skip("No warrior found")
	}

	reachable := g.ReachableTiles(warrior)

	// Should not be empty (warrior has 2 moves, there should be at least some passable tiles)
	if len(reachable) == 0 {
		t.Skip("No reachable tiles from warrior position (surrounded by impassable)")
	}

	// Unit's own position should not be in reachable set
	if reachable[[2]int{warrior.X, warrior.Y}] {
		t.Error("Unit's current position should not be in reachable tiles")
	}

	// All reachable tiles should be passable
	for pos := range reachable {
		tile := g.Map.GetTile(pos[0], pos[1])
		if tile == nil || !model.Terrains[tile.Terrain].Passable {
			t.Errorf("Reachable tile at (%d,%d) is not passable", pos[0], pos[1])
		}
	}

	// All reachable tiles should be within reasonable distance
	for pos := range reachable {
		dist := abs(pos[0]-warrior.X) + abs(pos[1]-warrior.Y)
		if dist > warrior.MovesLeft {
			// Could be reachable via cheaper path, but Manhattan distance
			// is a lower bound for BFS cost, so this should hold
			t.Errorf("Reachable tile at (%d,%d) is Manhattan distance %d from warrior with %d moves",
				pos[0], pos[1], dist, warrior.MovesLeft)
		}
	}
}

func TestFindPath(t *testing.T) {
g := NewGame(defaultOpts())

var warrior *model.Unit
for _, u := range g.Units {
if u.CivID == 1 && u.Type == model.UnitWarrior {
warrior = u
break
}
}
if warrior == nil {
t.Skip("No warrior found")
}

// Find a passable tile that is not the warrior's current position
destX, destY := -1, -1
for y := 0; y < g.Map.Height && destX == -1; y++ {
for x := 0; x < g.Map.Width && destX == -1; x++ {
if (x != warrior.X || y != warrior.Y) && model.Terrains[g.Map.Tiles[y][x].Terrain].Passable {
destX, destY = x, y
}
}
}
if destX == -1 {
t.Skip("No passable destination found")
}

path := g.FindPath(warrior, destX, destY)
if path == nil {
t.Skip("No path found to destination (map may be isolated)")
}

// Path should end at the destination
last := path[len(path)-1]
if last[0] != destX || last[1] != destY {
t.Errorf("Path last step should be (%d,%d), got (%d,%d)", destX, destY, last[0], last[1])
}

// Each step in path should be passable and adjacent to the previous step
prev := [2]int{warrior.X, warrior.Y}
for i, step := range path {
tile := g.Map.GetTile(step[0], step[1])
if tile == nil || !model.Terrains[tile.Terrain].Passable {
t.Errorf("Path step %d at (%d,%d) is not passable", i, step[0], step[1])
}
dist := abs(step[0]-prev[0]) + abs(step[1]-prev[1])
if dist != 1 {
t.Errorf("Path step %d at (%d,%d) is not adjacent to previous (%d,%d)", i, step[0], step[1], prev[0], prev[1])
}
prev = step
}
}

func TestAutoMovement(t *testing.T) {
g := NewGame(defaultOpts())

// Find a passable region for a fresh test warrior
var px, py int
for y := 3; y < g.Map.Height-3; y++ {
for x := 3; x < g.Map.Width-3; x++ {
tile := g.Map.GetTile(x, y)
dest := g.Map.GetTile(x+3, y)
if tile != nil && model.Terrains[tile.Terrain].Passable &&
dest != nil && model.Terrains[dest.Terrain].Passable {
px, py = x, y
goto placed
}
}
}
t.Skip("No suitable passable area found")
placed:

warrior := g.AddUnit(model.UnitWarrior, 1, px, py)
destX, destY := px+3, py
warrior.HasDest = true
warrior.DestX = destX
warrior.DestY = destY

origX := warrior.X

g.EndTurn()

// Warrior should have moved closer to destination
if warrior.X <= origX && warrior.X != destX {
t.Errorf("Warrior should have moved toward destination (%d,%d), still at (%d,%d)", destX, destY, warrior.X, warrior.Y)
}
}

func TestPlayerUnitsNeedingAttention(t *testing.T) {
g := NewGame(defaultOpts())

// All freshly-created player units should initially need attention
units := g.PlayerUnitsNeedingAttention()
if len(units) == 0 {
t.Fatal("Expected at least one unit needing attention at game start")
}

// Set a destination on the first unit – it should no longer need attention
u := units[0]
u.HasDest = true
u.DestX = u.X + 1
u.DestY = u.Y

unitsAfter := g.PlayerUnitsNeedingAttention()
for _, au := range unitsAfter {
if au.ID == u.ID {
t.Errorf("Unit with destination should not appear in PlayerUnitsNeedingAttention")
}
}
}

func TestBusyWorkerSkippedInAutoMove(t *testing.T) {
g := NewGame(defaultOpts())

var worker *model.Unit
for y := 3; y < g.Map.Height-3 && worker == nil; y++ {
for x := 3; x < g.Map.Width-3 && worker == nil; x++ {
tile := g.Map.GetTile(x, y)
if tile != nil && model.Terrains[tile.Terrain].Passable {
worker = g.AddUnit(model.UnitWorker, 1, x, y)
}
}
}
if worker == nil {
t.Skip("Could not place worker")
}

// Start building and also set a destination
worker.BuildingImprovement = model.ImprovementFarm
worker.ImprovementTurnsLeft = 3
worker.Waiting = true
worker.HasDest = true
worker.DestX = worker.X + 2
worker.DestY = worker.Y

origX := worker.X

g.EndTurn()

// Worker should NOT have moved – busy workers are skipped
if worker.X != origX {
t.Errorf("Busy worker should not auto-move; started at x=%d, now at x=%d", origX, worker.X)
}
}
