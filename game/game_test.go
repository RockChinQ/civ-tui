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
