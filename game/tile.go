package game

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
	Type       TerrainType
	Name       string
	Symbol     string
	Food       int
	Production int
	Gold       int
	MoveCost   int
	Passable   bool
}

var Terrains = map[TerrainType]Terrain{
	TerrainOcean:     {TerrainOcean, "Ocean", "~", 1, 0, 1, 1, false},
	TerrainCoast:     {TerrainCoast, "Coast", "≈", 1, 0, 2, 1, false},
	TerrainGrassland: {TerrainGrassland, "Grassland", ".", 2, 1, 0, 1, true},
	TerrainPlains:    {TerrainPlains, "Plains", ",", 1, 1, 1, 1, true},
	TerrainHills:     {TerrainHills, "Hills", "^", 1, 2, 0, 2, true},
	TerrainMountains: {TerrainMountains, "Mountains", "M", 0, 1, 0, 3, false},
	TerrainForest:    {TerrainForest, "Forest", "♣", 1, 2, 0, 2, true},
	TerrainDesert:    {TerrainDesert, "Desert", "_", 0, 0, 1, 1, true},
	TerrainTundra:    {TerrainTundra, "Tundra", ";", 1, 0, 0, 1, true},
}

type Tile struct {
	Terrain  TerrainType
	Revealed bool
	Visible  bool
}
