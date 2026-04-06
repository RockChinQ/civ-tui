package tui

import (
	"github.com/RockChinQ/civ-tui/game"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.InfoWidth = 28
		if m.Width > 120 {
			m.InfoWidth = 32
		}
		m.MsgHeight = 6
		m.centerViewport()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Game.State != game.StateRunning {
		if msg.String() == "q" || msg.String() == "Q" {
			return m, tea.Quit
		}
		return m, nil
	}

	switch m.ActiveMenu {
	case MenuBuild:
		return m.handleBuildMenu(msg)
	case MenuTech:
		return m.handleTechMenu(msg)
	case MenuHelp:
		if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
			m.ActiveMenu = MenuNone
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "Q":
		return m, tea.Quit
	case "?":
		m.ActiveMenu = MenuHelp
		return m, nil
	case "esc":
		m.SelectedUnit = nil
		m.ActiveMenu = MenuNone
		return m, nil
	case "up", "k":
		return m.moveCursorOrUnit(0, -1)
	case "down", "j":
		return m.moveCursorOrUnit(0, 1)
	case "left", "h":
		return m.moveCursorOrUnit(-1, 0)
	case "right", "l":
		return m.moveCursorOrUnit(1, 0)
	case "n", "N":
		return m.selectNextUnit(), nil
	case "f", "F":
		return m.foundCity()
	case "w", "W":
		return m.waitUnit()
	case "b", "B":
		return m.openBuildMenu()
	case "t", "T":
		return m.openTechMenu()
	case "enter":
		return m.endTurn()
	}
	return m, nil
}

func (m Model) moveCursorOrUnit(dx, dy int) (tea.Model, tea.Cmd) {
	if m.SelectedUnit != nil && m.SelectedUnit.IsAlive() && m.SelectedUnit.HasMoves() {
		u := m.SelectedUnit
		msg, ok := m.Game.MoveUnit(u, dx, dy)
		if ok {
			m.CursorX = u.X
			m.CursorY = u.Y
			if msg != "" {
				m.Game.AddMessage(msg)
			}
			if !u.IsAlive() {
				m.SelectedUnit = nil
			}
		} else {
			// Just move cursor
			m.moveCursor(dx, dy)
		}
	} else {
		m.moveCursor(dx, dy)
		// Auto-select unit at cursor
		u := m.Game.GetUnitAt(m.CursorX, m.CursorY)
		if u != nil && u.CivID == 1 && u.IsAlive() {
			m.SelectedUnit = u
		}
	}
	m.scrollViewportToCursor()
	return m, nil
}

func (m *Model) moveCursor(dx, dy int) {
	m.CursorX += dx
	m.CursorY += dy
	if m.CursorX < 0 {
		m.CursorX = 0
	}
	if m.CursorX >= game.MapWidth {
		m.CursorX = game.MapWidth - 1
	}
	if m.CursorY < 0 {
		m.CursorY = 0
	}
	if m.CursorY >= game.MapHeight {
		m.CursorY = game.MapHeight - 1
	}
}

func (m *Model) scrollViewportToCursor() {
	mapW, mapH := m.mapViewSize()
	// Scroll if cursor near edge
	if m.CursorX < m.ViewportX+3 {
		m.ViewportX = m.CursorX - 3
	}
	if m.CursorX >= m.ViewportX+mapW-3 {
		m.ViewportX = m.CursorX - mapW + 4
	}
	if m.CursorY < m.ViewportY+2 {
		m.ViewportY = m.CursorY - 2
	}
	if m.CursorY >= m.ViewportY+mapH-2 {
		m.ViewportY = m.CursorY - mapH + 3
	}
	m.clampViewport()
}

func (m Model) selectNextUnit() Model {
	units := m.Game.PlayerUnitsWithMoves()
	if len(units) == 0 {
		return m
	}
	// Find next unit after current selection
	idx := 0
	if m.SelectedUnit != nil {
		for i, u := range units {
			if u.ID == m.SelectedUnit.ID {
				idx = (i + 1) % len(units)
				break
			}
		}
	}
	u := units[idx]
	m.SelectedUnit = u
	m.CursorX = u.X
	m.CursorY = u.Y
	m.scrollViewportToCursor()
	return m
}

func (m Model) foundCity() (tea.Model, tea.Cmd) {
	if m.SelectedUnit == nil || m.SelectedUnit.Type != game.UnitSettler {
		return m, nil
	}
	msg, ok := m.Game.FoundCity(m.SelectedUnit, nil)
	m.Game.AddMessage(msg)
	if ok {
		m.SelectedUnit = nil
	}
	return m, nil
}

func (m Model) waitUnit() (tea.Model, tea.Cmd) {
	if m.SelectedUnit != nil {
		m.SelectedUnit.Waiting = true
		m.SelectedUnit = nil
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

func (m Model) handleBuildMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	city := m.Game.GetCityAt(m.CursorX, m.CursorY)
	items := buildMenuItems(city)

	switch msg.String() {
	case "esc", "b", "B":
		m.ActiveMenu = MenuNone
	case "up", "k":
		if m.MenuCursor > 0 {
			m.MenuCursor--
		}
	case "down", "j":
		if m.MenuCursor < len(items)-1 {
			m.MenuCursor++
		}
	case "enter":
		if city != nil && m.MenuCursor < len(items) {
			item := items[m.MenuCursor]
			city.ProductionQ = append(city.ProductionQ, item)
			m.Game.AddMessage("Queued: " + item.Name + " in " + city.Name)
		}
		m.ActiveMenu = MenuNone
	}
	return m, nil
}

func buildMenuItems(city *game.City) []game.ProductionItem {
	var items []game.ProductionItem
	// Units
	for _, ut := range []game.UnitType{game.UnitSettler, game.UnitScout, game.UnitWarrior, game.UnitArcher, game.UnitSpearman, game.UnitSwordsman, game.UnitHorseman} {
		stats := game.UnitDefs[ut]
		items = append(items, game.ProductionItem{
			IsUnit:   true,
			UnitType: ut,
			Name:     stats.Name,
			Cost:     stats.ProductionCost,
		})
	}
	// Buildings
	if city != nil {
		for bt, bdef := range game.BuildingDefs {
			if !city.Buildings[bt] {
				items = append(items, game.ProductionItem{
					IsUnit:       false,
					BuildingType: bt,
					Name:         bdef.Name,
					Cost:         bdef.Cost,
				})
			}
		}
	}
	return items
}

func (m Model) handleTechMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	playerCiv := m.Game.GetCiv(1)
	if playerCiv == nil {
		m.ActiveMenu = MenuNone
		return m, nil
	}
	available := game.AvailableTechs(playerCiv.Techs)

	switch msg.String() {
	case "esc", "t", "T":
		m.ActiveMenu = MenuNone
	case "up", "k":
		if m.MenuCursor > 0 {
			m.MenuCursor--
		}
	case "down", "j":
		if m.MenuCursor < len(available)-1 {
			m.MenuCursor++
		}
	case "enter":
		if m.MenuCursor < len(available) {
			playerCiv.ResearchTech(available[m.MenuCursor])
			playerCiv.ResearchProgress = 0
			m.Game.AddMessage("Researching: " + available[m.MenuCursor].Name)
		}
		m.ActiveMenu = MenuNone
	}
	return m, nil
}

func (m Model) endTurn() (tea.Model, tea.Cmd) {
	m.Game.EndTurn()
	// Select first unit with moves
	units := m.Game.PlayerUnitsWithMoves()
	m.SelectedUnit = nil
	if len(units) > 0 {
		m.SelectedUnit = units[0]
		m.CursorX = units[0].X
		m.CursorY = units[0].Y
		m.scrollViewportToCursor()
	}
	return m, nil
}
