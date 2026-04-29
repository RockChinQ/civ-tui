package web

import (
	"github.com/RockChinQ/civ-tui/game"
	"github.com/RockChinQ/civ-tui/game/model"
)

// View structs are flat, JSON-friendly snapshots of game state shaped for the
// front-end. They intentionally avoid exposing internal pointers so the client
// can render purely from JSON.

type tileView struct {
	Terrain    string `json:"terrain"`
	Symbol     string `json:"symbol"`
	Revealed   bool   `json:"revealed"`
	Visible    bool   `json:"visible"`
	Imp        string `json:"imp,omitempty"`
	ImpSymbol  string `json:"impSymbol,omitempty"`
	Passable   bool   `json:"passable"`
	MoveCost   int    `json:"moveCost"`
	DefBonus   int    `json:"defBonus"`
	Food       int    `json:"food"`
	Production int    `json:"prod"`
	Gold       int    `json:"gold"`
}

type unitView struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Symbol    string `json:"symbol"`
	CivID     int    `json:"civId"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	HP        int    `json:"hp"`
	MaxHP     int    `json:"maxHp"`
	Attack    int    `json:"attack"`
	Defense   int    `json:"defense"`
	MovesLeft int    `json:"movesLeft"`
	MaxMoves  int    `json:"maxMoves"`
	Range     int    `json:"range"`
	Waiting   bool   `json:"waiting"`
	XP        int    `json:"xp"`
	Level     int    `json:"level"`
	Building  string `json:"building,omitempty"`
	BuildLeft int    `json:"buildLeft,omitempty"`
	HasDest   bool   `json:"hasDest"`
	DestX     int    `json:"destX"`
	DestY     int    `json:"destY"`
}

type prodItemView struct {
	IsUnit bool   `json:"isUnit"`
	Name   string `json:"name"`
	Cost   int    `json:"cost"`
}

type cityView struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	CivID      int            `json:"civId"`
	X          int            `json:"x"`
	Y          int            `json:"y"`
	Population int            `json:"pop"`
	Food       int            `json:"food"`
	FoodNeeded int            `json:"foodNeeded"`
	Production int            `json:"prod"`
	HP         int            `json:"hp"`
	MaxHP      int            `json:"maxHp"`
	Buildings  []string       `json:"buildings"`
	Queue      []prodItemView `json:"queue"`
}

type civView struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	IsPlayer    bool           `json:"isPlayer"`
	IsAlive     bool           `json:"isAlive"`
	Gold        int            `json:"gold"`
	Science     int            `json:"science"`
	Researching string         `json:"researching"`
	Progress    int            `json:"progress"`
	Techs       []string       `json:"techs"`
	Relations   map[int]string `json:"relations"`
}

type messageView struct {
	Text     string `json:"text"`
	IsPlayer bool   `json:"isPlayer"`
}

type buildOption struct {
	IsUnit bool   `json:"isUnit"`
	Key    string `json:"key"`
	Name   string `json:"name"`
	Cost   int    `json:"cost"`
}

type techOption struct {
	Name     string `json:"name"`
	Cost     int    `json:"cost"`
	Progress int    `json:"progress"`
}

type gameStateView struct {
	Turn         int           `json:"turn"`
	MaxTurns     int           `json:"maxTurns"`
	State        string        `json:"state"`
	MapWidth     int           `json:"mapWidth"`
	MapHeight    int           `json:"mapHeight"`
	Tiles        [][]tileView  `json:"tiles"`
	Units        []unitView    `json:"units"`
	Cities       []cityView    `json:"cities"`
	Civs         []civView     `json:"civs"`
	Messages     []messageView `json:"messages"`
	BuildOptions []buildOption `json:"buildOptions"`
	TechOptions  []techOption  `json:"techOptions"`
}

func stateToView(g *game.Game) gameStateView {
	v := gameStateView{
		Turn:      g.Turn,
		MaxTurns:  g.MaxTurns,
		MapWidth:  g.Map.Width,
		MapHeight: g.Map.Height,
	}
	switch g.State {
	case game.StateRunning:
		v.State = "running"
	case game.StateVictory:
		v.State = "victory"
	case game.StateDefeat:
		v.State = "defeat"
	case game.StateDraw:
		v.State = "draw"
	}

	v.Tiles = make([][]tileView, g.Map.Height)
	for y := 0; y < g.Map.Height; y++ {
		row := make([]tileView, g.Map.Width)
		for x := 0; x < g.Map.Width; x++ {
			t := g.Map.Tiles[y][x]
			ter := model.Terrains[t.Terrain]
			tv := tileView{
				Terrain:    ter.Name,
				Symbol:     ter.Symbol,
				Revealed:   t.Revealed,
				Visible:    t.Visible,
				Passable:   ter.Passable,
				MoveCost:   ter.MoveCost,
				DefBonus:   ter.DefenseBonus,
				Food:       ter.Food,
				Production: ter.Production,
				Gold:       ter.Gold,
			}
			if t.Improvement != model.ImprovementNone {
				imp := model.Improvements[t.Improvement]
				tv.Imp = imp.Name
				tv.ImpSymbol = imp.Symbol
			}
			row[x] = tv
		}
		v.Tiles[y] = row
	}

	for _, u := range g.Units {
		if !u.IsAlive() {
			continue
		}
		stats := model.UnitDefs[u.Type]
		uv := unitView{
			ID: u.ID, Type: stats.Name, Symbol: stats.Symbol,
			CivID: u.CivID, X: u.X, Y: u.Y,
			HP: u.HP, MaxHP: u.MaxHP, Attack: u.Attack, Defense: u.Defense,
			MovesLeft: u.MovesLeft, MaxMoves: u.MaxMoves, Range: stats.Range,
			Waiting: u.Waiting, XP: u.XP, Level: u.Level,
			HasDest: u.HasDest, DestX: u.DestX, DestY: u.DestY,
		}
		if u.BuildingImprovement != model.ImprovementNone {
			uv.Building = model.Improvements[u.BuildingImprovement].Name
			uv.BuildLeft = u.ImprovementTurnsLeft
		}
		v.Units = append(v.Units, uv)
	}

	for _, c := range g.Cities {
		cv := cityView{
			ID: c.ID, Name: c.Name, CivID: c.CivID, X: c.X, Y: c.Y,
			Population: c.Population, Food: c.Food, FoodNeeded: c.FoodNeeded,
			Production: c.Production, HP: c.HP, MaxHP: c.MaxHP,
		}
		for bt := range c.Buildings {
			cv.Buildings = append(cv.Buildings, model.BuildingDefs[bt].Name)
		}
		for _, q := range c.ProductionQ {
			cv.Queue = append(cv.Queue, prodItemView{IsUnit: q.IsUnit, Name: q.Name, Cost: q.Cost})
		}
		v.Cities = append(v.Cities, cv)
	}

	for _, c := range g.Civs {
		civ := civView{
			ID: c.ID, Name: c.Name, IsPlayer: c.IsPlayer, IsAlive: c.IsAlive,
			Gold: c.Gold, Science: c.Science,
			Researching: c.Researching, Progress: c.ResearchProgress,
			Relations: make(map[int]string),
		}
		for name := range c.Techs {
			civ.Techs = append(civ.Techs, name)
		}
		for id, r := range c.Relations {
			if r == model.RelationWar {
				civ.Relations[id] = "war"
			} else {
				civ.Relations[id] = "peace"
			}
		}
		v.Civs = append(v.Civs, civ)
	}

	for _, m := range g.Messages {
		v.Messages = append(v.Messages, messageView{Text: m.Text, IsPlayer: m.IsPlayer})
	}

	// Build options for the player
	player := g.GetCiv(1)
	if player != nil {
		for _, ut := range model.AvailableUnits(player.Techs) {
			def := model.UnitDefs[ut]
			v.BuildOptions = append(v.BuildOptions, buildOption{
				IsUnit: true, Key: def.Name, Name: def.Name, Cost: def.ProductionCost,
			})
		}
		for _, bt := range model.AvailableBuildings(player.Techs) {
			def := model.BuildingDefs[bt]
			v.BuildOptions = append(v.BuildOptions, buildOption{
				IsUnit: false, Key: def.Name, Name: def.Name, Cost: def.Cost,
			})
		}
		for _, t := range model.AvailableTechs(player.Techs) {
			progress := 0
			if player.Researching == t.Name {
				progress = player.ResearchProgress
			}
			v.TechOptions = append(v.TechOptions, techOption{
				Name: t.Name, Cost: t.Cost, Progress: progress,
			})
		}
	}

	return v
}

func (s *Server) gameStateLocked() gameStateView {
	if s.game == nil {
		return gameStateView{State: "none"}
	}
	return stateToView(s.game)
}
