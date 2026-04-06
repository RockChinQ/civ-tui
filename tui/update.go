package tui

import (
	"github.com/RockChinQ/civ-tui/game"
	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/game/worldmap"
	"github.com/RockChinQ/civ-tui/i18n"
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
		if m.Game != nil {
			m.centerViewport()
		}
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Main menu / settings screens
	if m.CurrentScreen == ScreenMainMenu {
		return m.handleMainMenu(msg)
	}

	// In-game screens require a live game
	if m.Game == nil {
		return m, nil
	}

	if m.Game.State != game.StateRunning {
		if msg.String() == "q" || msg.String() == "Q" {
			return m, tea.Quit
		}
		return m, nil
	}

	// Range mode
	if m.RangeMode {
		return m.handleRangedMode(msg)
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
	case MenuDiplomacy:
		return m.handleDiplomacyMenu(msg)
	case MenuCity:
		return m.handleCityMenu(msg)
	case MenuPromotion:
		return m.handlePromotionMenu(msg)
	case MenuInspect:
		return m.handleInspectMenu(msg)
	}

	switch msg.String() {
	case "q", "Q":
		return m, tea.Quit
	case "?":
		m.ActiveMenu = MenuHelp
		return m, nil
	case "esc":
		m.SelectedUnit = nil
		m.ReachableTiles = nil
		m.ActiveMenu = MenuNone
		m.RangeMode = false
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
	case "r", "R":
		return m.enterRangeMode()
	case "d", "D":
		return m.openDiplomacyMenu()
	case "s", "S":
		return m.saveGame()
	case "i", "I":
		return m.startImprovement()
	case "v", "V":
		m.ActiveMenu = MenuInspect
		return m, nil
	case "enter":
		// If on own city with no unit, open city details
		city := m.Game.GetCityAt(m.CursorX, m.CursorY)
		unit := m.Game.GetUnitAt(m.CursorX, m.CursorY)
		if city != nil && city.CivID == 1 && unit == nil {
			m.ActiveMenu = MenuCity
			m.MenuCursor = 0
			return m, nil
		}
		return m.endTurn()
	}
	return m, nil
}

func (m Model) handleMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.InSettings {
		return m.handleSettingsMenu(msg)
	}

	items := 4 // New Game, Load Game, Settings, Quit
	switch msg.String() {
	case "up", "k":
		if m.MainMenuCursor > 0 {
			m.MainMenuCursor--
		}
	case "down", "j":
		if m.MainMenuCursor < items-1 {
			m.MainMenuCursor++
		}
	case "enter", " ":
		switch m.MainMenuCursor {
		case 0: // New Game
			m.startGame()
		case 1: // Load Game
			g, err := game.LoadFromFile(game.DefaultSavePath())
			if err == nil {
				m.Game = g
				m.CurrentScreen = ScreenGame
				for _, u := range g.Units {
					if u.CivID == 1 && u.IsAlive() {
						m.CursorX = u.X
						m.CursorY = u.Y
						m.SelectedUnit = u
						break
					}
				}
				m.updateReachable()
				m.centerViewport()
			} else {
				// No save file, just start new game
				m.startGame()
			}
		case 2: // Settings
			m.InSettings = true
			m.SettingsCursor = 0
		case 3: // Quit
			return m, tea.Quit
		}
	case "q", "Q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleSettingsMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	numItems := 5 // Language, Map Size, AI Civs, Difficulty, Back
	switch msg.String() {
	case "up", "k":
		if m.SettingsCursor > 0 {
			m.SettingsCursor--
		}
	case "down", "j":
		if m.SettingsCursor < numItems-1 {
			m.SettingsCursor++
		}
	case "left", "h":
		switch m.SettingsCursor {
		case 0: // Language
			lang := i18n.GetLang()
			if lang > 0 {
				i18n.SetLang(lang - 1)
			} else {
				i18n.SetLang(i18n.Lang(int(i18n.LangCount) - 1))
			}
		case 1: // Map Size
			if m.SettingsMapSize > worldmap.MapSizeSmall {
				m.SettingsMapSize--
			}
		case 2: // AI Civs
			if m.SettingsNumAICivs > 1 {
				m.SettingsNumAICivs--
			}
		case 3: // Difficulty
			if m.SettingsDifficulty > 1 {
				m.SettingsDifficulty--
			}
		}
	case "right", "l":
		switch m.SettingsCursor {
		case 0: // Language
			lang := i18n.GetLang()
			if lang < i18n.Lang(int(i18n.LangCount)-1) {
				i18n.SetLang(lang + 1)
			} else {
				i18n.SetLang(0)
			}
		case 1: // Map Size
			if m.SettingsMapSize < worldmap.MapSizeLarge {
				m.SettingsMapSize++
			}
		case 2: // AI Civs
			if m.SettingsNumAICivs < 4 {
				m.SettingsNumAICivs++
			}
		case 3: // Difficulty
			if m.SettingsDifficulty < 3 {
				m.SettingsDifficulty++
			}
		}
	case "enter":
		if m.SettingsCursor == numItems-1 {
			m.InSettings = false
		}
	case "esc", "b", "B":
		m.InSettings = false
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
				m.ReachableTiles = nil
				// Auto-select next unit with moves
				m = m.selectNextUnit()
			} else if !u.HasMoves() {
				// Unit exhausted moves, auto-select next unit
				units := m.Game.PlayerUnitsWithMoves()
				if len(units) > 0 {
					m.SelectedUnit = units[0]
					m.CursorX = units[0].X
					m.CursorY = units[0].Y
				}
				m.updateReachable()
			} else {
				m.updateReachable()
			}
		} else {
			m.moveCursor(dx, dy)
			// Auto-select player unit at cursor position
			cu := m.Game.GetUnitAt(m.CursorX, m.CursorY)
			if cu != nil && cu.CivID == 1 && cu.IsAlive() {
				m.SelectedUnit = cu
				m.updateReachable()
			}
		}
	} else {
		m.moveCursor(dx, dy)
		u := m.Game.GetUnitAt(m.CursorX, m.CursorY)
		if u != nil && u.CivID == 1 && u.IsAlive() {
			m.SelectedUnit = u
			m.updateReachable()
		}
	}
	m.scrollViewportToCursor()
	return m, nil
}

func (m *Model) moveCursor(dx, dy int) {
	mw, mh := m.gameMapSize()
	m.CursorX += dx
	m.CursorY += dy
	if m.CursorX < 0 {
		m.CursorX = 0
	}
	if m.CursorX >= mw {
		m.CursorX = mw - 1
	}
	if m.CursorY < 0 {
		m.CursorY = 0
	}
	if m.CursorY >= mh {
		m.CursorY = mh - 1
	}
}

func (m *Model) scrollViewportToCursor() {
	mapW, mapH := m.mapViewSize()
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
		m.ReachableTiles = nil
		return m
	}
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
	m.updateReachable()
	m.scrollViewportToCursor()
	return m
}

func (m Model) foundCity() (tea.Model, tea.Cmd) {
	if m.SelectedUnit == nil || m.SelectedUnit.Type != model.UnitSettler {
		return m, nil
	}
	msg, ok := m.Game.FoundCity(m.SelectedUnit, nil)
	m.Game.AddMessage(msg)
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
		m.Game.AddMessage(i18n.T("This unit cannot perform ranged attacks"))
		return m, nil
	}
	m.RangeMode = true
	m.Game.AddMessage(i18n.T("Ranged mode: select target with arrow keys, Enter to fire, Esc to cancel"))
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
				m.Game.AddMessage(i18n.T("Target out of range"))
			} else {
				target := m.Game.GetUnitAt(m.CursorX, m.CursorY)
				if target != nil && target.CivID != m.SelectedUnit.CivID {
					result := m.Game.RangedAttack(m.SelectedUnit, target)
					m.Game.AddMessage(result)
					if !m.SelectedUnit.IsAlive() {
						m.SelectedUnit = nil
					}
				} else {
					m.Game.AddMessage(i18n.T("No enemy unit at target"))
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

func (m Model) handleDiplomacyMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	civs := m.diplomacyCivs()
	switch msg.String() {
	case "esc", "d", "D":
		m.ActiveMenu = MenuNone
	case "up", "k":
		if m.MenuCursor > 0 {
			m.MenuCursor--
		}
	case "down", "j":
		if m.MenuCursor < len(civs)-1 {
			m.MenuCursor++
		}
	case "enter":
		if m.MenuCursor < len(civs) {
			target := civs[m.MenuCursor]
			player := m.Game.GetCiv(1)
			if player != nil {
				rel := m.Game.GetRelation(1, target.ID)
				if rel == model.RelationWar {
					m.Game.MakePeace(player, target)
					m.Game.AddMessage(i18n.Tf("Made peace with %s", i18n.T(target.Name)))
				} else {
					m.Game.DeclareWar(player, target)
					m.Game.AddMessage(i18n.Tf("Declared war on %s", i18n.T(target.Name)))
				}
			}
		}
		m.ActiveMenu = MenuNone
	}
	return m, nil
}

func (m Model) diplomacyCivs() []*model.Civ {
	var civs []*model.Civ
	for _, c := range m.Game.Civs {
		if !c.IsPlayer && c.IsAlive {
			civs = append(civs, c)
		}
	}
	return civs
}

func (m Model) handleCityMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "b", "B":
		m.ActiveMenu = MenuNone
	}
	return m, nil
}

func (m Model) handleInspectMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "v", "V":
		m.ActiveMenu = MenuNone
	}
	return m, nil
}

func (m Model) handlePromotionMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.ActiveMenu = MenuNone
		m.PendingPromotion = nil
	case "enter":
		if m.PendingPromotion != nil {
			m.PendingPromotion.Attack++
			m.PendingPromotion.XP -= 5
			m.Game.AddMessage(i18n.Tf("%s promoted!", i18n.T(model.UnitDefs[m.PendingPromotion.Type].Name)))
			m.PendingPromotion = nil
		}
		m.ActiveMenu = MenuNone
	}
	return m, nil
}

func (m Model) saveGame() (tea.Model, tea.Cmd) {
	if m.Game == nil {
		return m, nil
	}
	err := m.Game.SaveToFile(game.DefaultSavePath())
	if err != nil {
		m.Game.AddMessage(i18n.Tf("Failed to save: %s", err.Error()))
	} else {
		m.Game.AddMessage(i18n.T("Game saved!"))
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
		m.Game.AddMessage(i18n.Tf("Need %s to build %s", i18n.T(impDef.RequiresTech), i18n.T(impDef.Name)))
		return m, nil
	}
	u.BuildingImprovement = imp
	u.ImprovementTurnsLeft = impDef.BuildTurns
	u.Waiting = true
	m.Game.AddMessage(i18n.Tf("Worker building %s (%d turns)", i18n.T(impDef.Name), impDef.BuildTurns))
	return m, nil
}

func (m Model) handleBuildMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	city := m.Game.GetCityAt(m.CursorX, m.CursorY)
	playerCiv := m.Game.GetCiv(1)
	var civTechs map[string]bool
	if playerCiv != nil {
		civTechs = playerCiv.Techs
	}
	items := buildMenuItems(city, civTechs)

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
			m.Game.AddMessage(i18n.Tf("Queued: %s in %s", i18n.T(item.Name), city.Name))
		}
		m.ActiveMenu = MenuNone
	}
	return m, nil
}

func buildMenuItems(city *model.City, civTechs map[string]bool) []model.ProductionItem {
	var items []model.ProductionItem
	for _, ut := range model.AvailableUnits(civTechs) {
		stats := model.UnitDefs[ut]
		items = append(items, model.ProductionItem{
			IsUnit:   true,
			UnitType: ut,
			Name:     stats.Name,
			Cost:     stats.ProductionCost,
		})
	}
	if city != nil {
		for _, bt := range model.AvailableBuildings(civTechs) {
			if !city.Buildings[bt] {
				bdef := model.BuildingDefs[bt]
				items = append(items, model.ProductionItem{
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
	available := model.AvailableTechs(playerCiv.Techs)

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
			m.Game.AddMessage(i18n.Tf("Researching: %s", i18n.T(available[m.MenuCursor].Name)))
		}
		m.ActiveMenu = MenuNone
	}
	return m, nil
}

func (m Model) endTurn() (tea.Model, tea.Cmd) {
	m.Game.EndTurn()
	units := m.Game.PlayerUnitsWithMoves()
	m.SelectedUnit = nil
	m.ReachableTiles = nil
	if len(units) > 0 {
		m.SelectedUnit = units[0]
		m.CursorX = units[0].X
		m.CursorY = units[0].Y
		m.updateReachable()
		m.scrollViewportToCursor()
	}
	return m, nil
}
