package model

type TerrainType int

const (
	TerrainOcean TerrainType = iota
	TerrainCoast
	TerrainGrassland
	TerrainPlains
	TerrainHills
	TerrainMountains
	TerrainForest
	TerrainDesert
	TerrainTundra
)

type Terrain struct {
	Type         TerrainType
	Name         string
	Symbol       string
	Food         int
	Production   int
	Gold         int
	MoveCost     int
	Passable     bool
	DefenseBonus int // percentage bonus (e.g. 25 = 25%)
}

var Terrains = map[TerrainType]Terrain{
	TerrainOcean:     {TerrainOcean, "Ocean", "~", 1, 0, 1, 1, false, 0},
	TerrainCoast:     {TerrainCoast, "Coast", "≈", 1, 0, 2, 1, false, 0},
	TerrainGrassland: {TerrainGrassland, "Grassland", ".", 2, 1, 0, 1, true, 0},
	TerrainPlains:    {TerrainPlains, "Plains", ",", 1, 1, 1, 1, true, 0},
	TerrainHills:     {TerrainHills, "Hills", "^", 1, 2, 0, 2, true, 25},
	TerrainMountains: {TerrainMountains, "Mountains", "M", 0, 1, 0, 3, false, 0},
	TerrainForest:    {TerrainForest, "Forest", "♣", 1, 2, 0, 2, true, 25},
	TerrainDesert:    {TerrainDesert, "Desert", "_", 0, 0, 1, 1, true, 0},
	TerrainTundra:    {TerrainTundra, "Tundra", ";", 1, 0, 0, 1, true, 0},
}

type ImprovementType int

const (
	ImprovementNone ImprovementType = iota
	ImprovementFarm
	ImprovementMine
	ImprovementRoad
	ImprovementLumberMill
)

type Improvement struct {
	Type         ImprovementType
	Name         string
	Symbol       string
	FoodBonus    int
	ProdBonus    int
	GoldBonus    int
	BuildTurns   int
	RequiresTech string
}

var Improvements = map[ImprovementType]Improvement{
	ImprovementFarm:       {ImprovementFarm, "Farm", "f", 1, 0, 0, 3, "Agriculture"},
	ImprovementMine:       {ImprovementMine, "Mine", "m", 0, 2, 0, 4, "Mining"},
	ImprovementRoad:       {ImprovementRoad, "Road", "r", 0, 0, 1, 2, ""},
	ImprovementLumberMill: {ImprovementLumberMill, "Lumber Mill", "l", 0, 1, 1, 3, ""},
}

type Tile struct {
	Terrain             TerrainType
	Revealed            bool
	Visible             bool
	Improvement         ImprovementType
	ImprovementProgress int
}
