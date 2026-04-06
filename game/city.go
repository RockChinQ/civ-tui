package game

type BuildingType int

const (
	BuildingGranary BuildingType = iota
	BuildingBarracks
	BuildingMarket
	BuildingLibrary
	BuildingWalls
)

type Building struct {
	Type      BuildingType
	Name      string
	Cost      int
	FoodBonus int
	ProdBonus int
	GoldBonus int
	SciBonus  int
	DefBonus  int
}

var BuildingDefs = map[BuildingType]Building{
	BuildingGranary:  {BuildingGranary, "Granary", 30, 2, 0, 0, 0, 0},
	BuildingBarracks: {BuildingBarracks, "Barracks", 35, 0, 1, 0, 0, 0},
	BuildingMarket:   {BuildingMarket, "Market", 40, 0, 0, 2, 0, 0},
	BuildingLibrary:  {BuildingLibrary, "Library", 40, 0, 0, 0, 2, 0},
	BuildingWalls:    {BuildingWalls, "Walls", 60, 0, 0, 0, 0, 3},
}

type ProductionItem struct {
	IsUnit       bool
	UnitType     UnitType
	BuildingType BuildingType
	Name         string
	Cost         int
}

type City struct {
	ID          int
	Name        string
	CivID       int
	X, Y        int
	Population  int
	Food        int
	FoodNeeded  int
	Production  int
	Gold        int
	Buildings   map[BuildingType]bool
	ProductionQ []ProductionItem
	HP          int
	MaxHP       int
	Defense     int
}

func NewCity(id int, name string, civID, x, y int) *City {
	return &City{
		ID:         id,
		Name:       name,
		CivID:      civID,
		X:          x,
		Y:          y,
		Population: 1,
		Food:       0,
		FoodNeeded: 10,
		Production: 0,
		Buildings:  make(map[BuildingType]bool),
		HP:         20,
		MaxHP:      20,
		Defense:    3,
	}
}

func (c *City) FoodYield(m *GameMap) int {
	base := 2
	t := m.GetTile(c.X, c.Y)
	if t != nil {
		base += Terrains[t.Terrain].Food
	}
	if c.Buildings[BuildingGranary] {
		base += 2
	}
	return base
}

func (c *City) ProductionYield(m *GameMap) int {
	base := 1
	t := m.GetTile(c.X, c.Y)
	if t != nil {
		base += Terrains[t.Terrain].Production
	}
	if c.Buildings[BuildingBarracks] {
		base++
	}
	return base
}

func (c *City) GoldYield(m *GameMap) int {
	base := 1
	t := m.GetTile(c.X, c.Y)
	if t != nil {
		base += Terrains[t.Terrain].Gold
	}
	if c.Buildings[BuildingMarket] {
		base += 2
	}
	return base
}

func (c *City) ScienceYield() int {
	base := 1
	if c.Buildings[BuildingLibrary] {
		base += 2
	}
	return base * c.Population
}

func (c *City) ProcessTurn(m *GameMap) (unitBuilt *ProductionItem, msg string) {
	// Food
	c.Food += c.FoodYield(m)
	if c.Food >= c.FoodNeeded {
		c.Food -= c.FoodNeeded
		c.Population++
		c.FoodNeeded = c.Population * 10
		msg = c.Name + " grew to population " + itoa(c.Population)
	}

	// Production
	if len(c.ProductionQ) > 0 {
		item := &c.ProductionQ[0]
		c.Production += c.ProductionYield(m)
		if c.Production >= item.Cost {
			c.Production -= item.Cost
			built := *item
			c.ProductionQ = c.ProductionQ[1:]
			return &built, msg
		}
	}
	return nil, msg
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
