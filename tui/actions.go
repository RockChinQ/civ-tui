package tui

import (
	"github.com/RockChinQ/civ-tui/game"
	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/game/worldmap"
	"github.com/RockChinQ/civ-tui/i18n"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) foundCity() (tea.Model, tea.Cmd) {
	if m.SelectedUnit == nil || m.SelectedUnit.Type != model.UnitSettler {
		return m, nil
	}
	msg, ok := m.Game.FoundCity(m.SelectedUnit, nil)
	m.Game.AddPlayerMessage(msg)
	if ok {
		m.SelectedUnit = nil
		m.ReachableTiles = nil
		// Auto-select next unit with moves
		m = m.selectNextUnit()
	}
	return m, nil
}

func (m Model) waitUnit() (tea.Model, tea.Cmd) {
	if m.SelectedUnit != nil {
		m.SelectedUnit.Waiting = true
		m.SelectedUnit = nil
		m.ReachableTiles = nil
		// Auto-select next unit with moves
		m = m.selectNextUnit()
	}
	return m, nil
}

func (m Model) openBuildMenu() (tea.Model, tea.Cmd) {
	city := m.Game.GetCityAt(m.CursorX, m.CursorY)
	if city == nil || city.CivID != 1 {
		return m, nil
	}
	m.ActiveMenu = MenuBuild
	m.MenuCursor = 0
	return m, nil
}

func (m Model) openTechMenu() (tea.Model, tea.Cmd) {
	m.ActiveMenu = MenuTech
	m.MenuCursor = 0
	return m, nil
}

func (m Model) enterRangeMode() (tea.Model, tea.Cmd) {
	if m.SelectedUnit == nil || !m.SelectedUnit.IsAlive() {
		return m, nil
	}
	stats := model.UnitDefs[m.SelectedUnit.Type]
	if stats.Range <= 0 {
		m.Game.AddPlayerMessage(i18n.T("This unit cannot perform ranged attacks"))
		return m, nil
	}
	m.RangeMode = true
	m.Game.AddPlayerMessage(i18n.T("Ranged mode: select target with arrow keys, Enter to fire, Esc to cancel"))
	return m, nil
}

func (m Model) handleRangedMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.RangeMode = false
		return m, nil
	case "up", "k":
		m.moveCursor(0, -1)
		m.scrollViewportToCursor()
	case "down", "j":
		m.moveCursor(0, 1)
		m.scrollViewportToCursor()
	case "left", "h":
		m.moveCursor(-1, 0)
		m.scrollViewportToCursor()
	case "right", "l":
		m.moveCursor(1, 0)
		m.scrollViewportToCursor()
	case "enter":
		if m.SelectedUnit != nil && m.SelectedUnit.IsAlive() {
			stats := model.UnitDefs[m.SelectedUnit.Type]
			dist := worldmap.AbsDist(m.SelectedUnit.X, m.SelectedUnit.Y, m.CursorX, m.CursorY)
			if dist > stats.Range {
				m.Game.AddPlayerMessage(i18n.T("Target out of range"))
			} else {
				target := m.Game.GetUnitAt(m.CursorX, m.CursorY)
				if target != nil && target.CivID != m.SelectedUnit.CivID {
					result := m.Game.RangedAttack(m.SelectedUnit, target)
					m.Game.AddPlayerMessage(result)
					if !m.SelectedUnit.IsAlive() {
						m.SelectedUnit = nil
					}
				} else {
					m.Game.AddPlayerMessage(i18n.T("No enemy unit at target"))
				}
			}
			m.RangeMode = false
		}
	}
	return m, nil
}

func (m Model) openDiplomacyMenu() (tea.Model, tea.Cmd) {
	m.ActiveMenu = MenuDiplomacy
	m.MenuCursor = 0
	return m, nil
}

func (m Model) saveGame() (tea.Model, tea.Cmd) {
	if m.Game == nil {
		return m, nil
	}
	err := m.Game.SaveToFile(game.DefaultSavePath())
	if err != nil {
		m.Game.AddPlayerMessage(i18n.Tf("Failed to save: %s", err.Error()))
	} else {
		m.Game.AddPlayerMessage(i18n.T("Game saved!"))
	}
	return m, nil
}

func (m Model) startImprovement() (tea.Model, tea.Cmd) {
	if m.SelectedUnit == nil || m.SelectedUnit.Type != model.UnitWorker {
		return m, nil
	}
	u := m.SelectedUnit
	tile := m.Game.Map.GetTile(u.X, u.Y)
	if tile == nil {
		return m, nil
	}
	// Pick best improvement for terrain
	var imp model.ImprovementType
	switch tile.Terrain {
	case model.TerrainGrassland, model.TerrainPlains:
		imp = model.ImprovementFarm
	case model.TerrainHills, model.TerrainMountains:
		imp = model.ImprovementMine
	case model.TerrainForest:
		imp = model.ImprovementLumberMill
	default:
		imp = model.ImprovementRoad
	}
	impDef := model.Improvements[imp]
	// Check tech requirement
	playerCiv := m.Game.GetCiv(1)
	if impDef.RequiresTech != "" && (playerCiv == nil || !playerCiv.Techs[impDef.RequiresTech]) {
		m.Game.AddPlayerMessage(i18n.Tf("Need %s to build %s", i18n.T(impDef.RequiresTech), i18n.T(impDef.Name)))
		return m, nil
	}
	u.BuildingImprovement = imp
	u.ImprovementTurnsLeft = impDef.BuildTurns
	u.Waiting = true
	m.Game.AddPlayerMessage(i18n.Tf("Worker building %s (%d turns)", i18n.T(impDef.Name), impDef.BuildTurns))
	return m, nil
}

func (m Model) endTurn() (tea.Model, tea.Cmd) {
	m.DestMode = false
	prev := m.SelectedUnit
	m.Game.EndTurn()

	// Keep previous unit selected if it still needs player attention
	if prev != nil && prev.IsAlive() && prev.HasMoves() && !prev.HasDest {
		m.SelectedUnit = prev
		m.updateReachable()
		return m, nil
	}

	// Otherwise fall back to first unit needing attention
	m.SelectedUnit = nil
	m.ReachableTiles = nil
	units := m.Game.PlayerUnitsNeedingAttention()
	if len(units) > 0 {
		m.SelectedUnit = units[0]
		m.CursorX = units[0].X
		m.CursorY = units[0].Y
		m.updateReachable()
		m.scrollViewportToCursor()
	}
	return m, nil
}
