package game

import (
"math/rand"
)

const (
MapWidth  = 60
MapHeight = 35
)

type MapSize int

const (
MapSizeSmall  MapSize = iota
MapSizeMedium
MapSizeLarge
)

type MapSizeConfig struct {
Width  int
Height int
}

var MapSizes = map[MapSize]MapSizeConfig{
MapSizeSmall:  {40, 25},
MapSizeMedium: {60, 35},
MapSizeLarge:  {80, 48},
}

type GameMap struct {
Width  int
Height int
Tiles  [][]Tile
}

func NewGameMap(seed int64, size MapSize) *GameMap {
cfg, ok := MapSizes[size]
if !ok {
cfg = MapSizes[MapSizeMedium]
}
m := &GameMap{
Width:  cfg.Width,
Height: cfg.Height,
Tiles:  make([][]Tile, cfg.Height),
}
for y := 0; y < cfg.Height; y++ {
m.Tiles[y] = make([]Tile, cfg.Width)
for x := 0; x < cfg.Width; x++ {
m.Tiles[y][x] = Tile{Terrain: generateTerrain(seed, x, y, cfg.Width, cfg.Height)}
}
}
return m
}

func generateTerrain(seed int64, x, y, w, h int) TerrainType {
edgeX := float64(x) / float64(w)
edgeY := float64(y) / float64(h)
distFromEdge := min4(edgeX, 1-edgeX, edgeY, 1-edgeY)

n := fractalNoise(x, y, seed)

if distFromEdge < 0.08 {
return TerrainOcean
}
if distFromEdge < 0.15 {
if n < 0.5 {
return TerrainCoast
}
return TerrainOcean
}

switch {
case n < 0.08:
return TerrainOcean
case n < 0.14:
return TerrainCoast
case n < 0.22:
return TerrainDesert
case n < 0.28:
return TerrainTundra
case n < 0.38:
return TerrainMountains
case n < 0.50:
return TerrainHills
case n < 0.60:
return TerrainForest
case n < 0.75:
return TerrainPlains
default:
return TerrainGrassland
}
}

func min4(a, b, c, d float64) float64 {
m := a
if b < m {
m = b
}
if c < m {
m = c
}
if d < m {
m = d
}
return m
}

func (m *GameMap) InBounds(x, y int) bool {
return x >= 0 && x < m.Width && y >= 0 && y < m.Height
}

func (m *GameMap) GetTile(x, y int) *Tile {
if !m.InBounds(x, y) {
return nil
}
return &m.Tiles[y][x]
}

func (m *GameMap) Reveal(x, y, radius int) {
for dy := -radius; dy <= radius; dy++ {
for dx := -radius; dx <= radius; dx++ {
nx, ny := x+dx, y+dy
if m.InBounds(nx, ny) {
dist := abs(dx) + abs(dy)
if dist <= radius {
m.Tiles[ny][nx].Revealed = true
m.Tiles[ny][nx].Visible = true
}
}
}
}
}

func (m *GameMap) ResetVisibility() {
for y := 0; y < m.Height; y++ {
for x := 0; x < m.Width; x++ {
m.Tiles[y][x].Visible = false
}
}
}

func abs(x int) int {
if x < 0 {
return -x
}
return x
}

func (m *GameMap) FindPassableTile(r *rand.Rand, attempts int) (int, int, bool) {
for i := 0; i < attempts; i++ {
x := r.Intn(m.Width)
y := r.Intn(m.Height)
t := Terrains[m.Tiles[y][x].Terrain]
if t.Passable {
return x, y, true
}
}
return 0, 0, false
}

// AbsDist returns the Manhattan distance between two points.
func AbsDist(x1, y1, x2, y2 int) int {
return abs(x2-x1) + abs(y2-y1)
}
