package web

import (
	"net/http"

	"github.com/RockChinQ/civ-tui/game"
	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/game/worldmap"
)

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.game == nil {
		writeJSON(w, http.StatusOK, map[string]any{"state": "none"})
		return
	}
	writeJSON(w, http.StatusOK, s.gameStateLocked())
}

type newGameReq struct {
	NumAI      int `json:"numAI"`
	MapSize    int `json:"mapSize"`
	Difficulty int `json:"difficulty"`
}

func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	var req newGameReq
	_ = decodeJSON(r, &req)
	if req.NumAI <= 0 {
		req.NumAI = 1
	}
	if req.Difficulty <= 0 {
		req.Difficulty = 1
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.game = game.NewGame(game.GameOptions{
		NumAICivs:  req.NumAI,
		MapSize:    worldmap.MapSize(req.MapSize),
		Difficulty: req.Difficulty,
	})
	writeJSON(w, http.StatusOK, s.gameStateLocked())
}

func (s *Server) handleSaves(w http.ResponseWriter, r *http.Request) {
	saves := game.ListSaves()
	out := make([]map[string]any, 0, len(saves))
	for _, sv := range saves {
		out = append(out, map[string]any{
			"slot":    sv.Slot,
			"empty":   sv.IsEmpty,
			"civ":     sv.CivName,
			"turn":    sv.Turn,
			"modTime": sv.ModTime,
			"label":   sv.SlotLabel(),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

type slotReq struct {
	Slot int `json:"slot"`
}

func (s *Server) handleSave(w http.ResponseWriter, r *http.Request) {
	var req slotReq
	if err := decodeJSON(r, &req); err != nil || req.Slot < 1 || req.Slot > game.MaxSaveSlots {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid slot"})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.game == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no game"})
		return
	}
	if err := s.game.SaveToFile(game.SlotPath(req.Slot)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleLoad(w http.ResponseWriter, r *http.Request) {
	var req slotReq
	if err := decodeJSON(r, &req); err != nil || req.Slot < 1 || req.Slot > game.MaxSaveSlots {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid slot"})
		return
	}
	g, err := game.LoadFromFile(game.SlotPath(req.Slot))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.game = g
	writeJSON(w, http.StatusOK, s.gameStateLocked())
}

func (s *Server) findPlayerUnit(g *game.Game, id int) *model.Unit {
	for _, u := range g.Units {
		if u.ID == id && u.CivID == 1 && u.IsAlive() {
			return u
		}
	}
	return nil
}

type moveReq struct {
	UnitID int `json:"unitId"`
	DX     int `json:"dx"`
	DY     int `json:"dy"`
}

func (s *Server) handleMove(w http.ResponseWriter, r *http.Request) {
	var req moveReq
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.withGame(w, func(g *game.Game) any {
		u := s.findPlayerUnit(g, req.UnitID)
		if u == nil {
			return map[string]string{"error": "unit not found"}
		}
		msg, ok := g.MoveUnit(u, req.DX, req.DY)
		if msg != "" {
			g.AddPlayerMessage(msg)
		}
		_ = ok
		return nil
	})
}

type unitReq struct {
	UnitID int `json:"unitId"`
}

func (s *Server) handleFoundCity(w http.ResponseWriter, r *http.Request) {
	var req unitReq
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.withGame(w, func(g *game.Game) any {
		u := s.findPlayerUnit(g, req.UnitID)
		if u == nil {
			return map[string]string{"error": "unit not found"}
		}
		msg, _ := g.FoundCity(u, nil)
		if msg != "" {
			g.AddPlayerMessage(msg)
		}
		return nil
	})
}

func (s *Server) handleWait(w http.ResponseWriter, r *http.Request) {
	var req unitReq
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.withGame(w, func(g *game.Game) any {
		u := s.findPlayerUnit(g, req.UnitID)
		if u != nil {
			u.Waiting = true
		}
		return nil
	})
}

func (s *Server) handleEndTurn(w http.ResponseWriter, r *http.Request) {
	s.withGame(w, func(g *game.Game) any {
		g.EndTurn()
		return nil
	})
}

type rangedReq struct {
	UnitID int `json:"unitId"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

func (s *Server) handleRanged(w http.ResponseWriter, r *http.Request) {
	var req rangedReq
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.withGame(w, func(g *game.Game) any {
		u := s.findPlayerUnit(g, req.UnitID)
		if u == nil {
			return map[string]string{"error": "unit not found"}
		}
		stats := model.UnitDefs[u.Type]
		if stats.Range <= 0 {
			g.AddPlayerMessage("This unit cannot perform ranged attacks")
			return nil
		}
		dist := worldmap.AbsDist(u.X, u.Y, req.X, req.Y)
		if dist > stats.Range {
			g.AddPlayerMessage("Target out of range")
			return nil
		}
		target := g.GetUnitAt(req.X, req.Y)
		if target == nil || target.CivID == u.CivID {
			g.AddPlayerMessage("No enemy unit at target")
			return nil
		}
		result := g.RangedAttack(u, target)
		g.AddPlayerMessage(result)
		return nil
	})
}

type impReq struct {
	UnitID int    `json:"unitId"`
	Type   string `json:"type"`
}

var impByName = map[string]model.ImprovementType{
	"Farm":        model.ImprovementFarm,
	"Mine":        model.ImprovementMine,
	"Road":        model.ImprovementRoad,
	"Lumber Mill": model.ImprovementLumberMill,
}

func (s *Server) handleImprovement(w http.ResponseWriter, r *http.Request) {
	var req impReq
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.withGame(w, func(g *game.Game) any {
		u := s.findPlayerUnit(g, req.UnitID)
		if u == nil || u.Type != model.UnitWorker {
			return map[string]string{"error": "not a worker"}
		}
		impType, ok := impByName[req.Type]
		if !ok {
			return map[string]string{"error": "bad imp"}
		}
		impDef := model.Improvements[impType]
		player := g.GetCiv(1)
		if impDef.RequiresTech != "" && (player == nil || !player.Techs[impDef.RequiresTech]) {
			g.AddPlayerMessage("Missing required tech: " + impDef.RequiresTech)
			return nil
		}
		u.BuildingImprovement = impType
		u.ImprovementTurnsLeft = impDef.BuildTurns
		u.Waiting = true
		g.AddPlayerMessage("Worker building " + impDef.Name)
		return nil
	})
}

type buildReq struct {
	CityID int    `json:"cityId"`
	IsUnit bool   `json:"isUnit"`
	Name   string `json:"name"`
}

func (s *Server) handleBuild(w http.ResponseWriter, r *http.Request) {
	var req buildReq
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.withGame(w, func(g *game.Game) any {
		var city *model.City
		for _, c := range g.Cities {
			if c.ID == req.CityID && c.CivID == 1 {
				city = c
				break
			}
		}
		if city == nil {
			return map[string]string{"error": "city not found"}
		}
		if req.IsUnit {
			for ut, def := range model.UnitDefs {
				if def.Name == req.Name {
					city.ProductionQ = append(city.ProductionQ, model.ProductionItem{
						IsUnit: true, UnitType: ut, Name: def.Name, Cost: def.ProductionCost,
					})
					g.AddPlayerMessage(city.Name + " queued " + def.Name)
					return nil
				}
			}
		} else {
			for bt, def := range model.BuildingDefs {
				if def.Name == req.Name {
					if city.Buildings[bt] {
						g.AddPlayerMessage(city.Name + " already has " + def.Name)
						return nil
					}
					city.ProductionQ = append(city.ProductionQ, model.ProductionItem{
						IsUnit: false, BuildingType: bt, Name: def.Name, Cost: def.Cost,
					})
					g.AddPlayerMessage(city.Name + " queued " + def.Name)
					return nil
				}
			}
		}
		return map[string]string{"error": "unknown item"}
	})
}

type techReq struct {
	Name string `json:"name"`
}

func (s *Server) handleResearch(w http.ResponseWriter, r *http.Request) {
	var req techReq
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.withGame(w, func(g *game.Game) any {
		player := g.GetCiv(1)
		if player == nil {
			return map[string]string{"error": "no player"}
		}
		t := model.GetTech(req.Name)
		if t == nil {
			return map[string]string{"error": "unknown tech"}
		}
		if player.Techs[t.Name] {
			return map[string]string{"error": "already known"}
		}
		for _, dep := range t.Requires {
			if !player.Techs[dep] {
				g.AddPlayerMessage("Need prerequisite: " + dep)
				return nil
			}
		}
		player.Researching = t.Name
		player.ResearchProgress = 0
		g.AddPlayerMessage("Now researching " + t.Name)
		return nil
	})
}

type destReq struct {
	UnitID int `json:"unitId"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

func (s *Server) handleSetDest(w http.ResponseWriter, r *http.Request) {
	var req destReq
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.withGame(w, func(g *game.Game) any {
		u := s.findPlayerUnit(g, req.UnitID)
		if u == nil {
			return map[string]string{"error": "unit not found"}
		}
		u.HasDest = true
		u.DestX = req.X
		u.DestY = req.Y
		g.AddPlayerMessage("Destination set")
		return nil
	})
}

func (s *Server) handleClearDest(w http.ResponseWriter, r *http.Request) {
	var req unitReq
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.withGame(w, func(g *game.Game) any {
		u := s.findPlayerUnit(g, req.UnitID)
		if u != nil {
			u.HasDest = false
		}
		return nil
	})
}
