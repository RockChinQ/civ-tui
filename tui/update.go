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

	// Save game slot picker
	if m.InSaveGame {
		return m.handleSaveGameMenu(msg)
	}

	// Destination (goto) mode
	if m.DestMode {
		return m.handleDestMode(msg)
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
		m.DestMode = false
		return m, nil
	case "up", "k":
		return m.moveCursorAndSelect(0, -1)
	case "down", "j":
		return m.moveCursorAndSelect(0, 1)
	case "left", "h":
		return m.moveCursorAndSelect(-1, 0)
	case "right", "l":
		return m.moveCursorAndSelect(1, 0)
	case "n", "N":
		return m.selectNextFocus(), nil
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
	case "g", "G":
		return m.enterDestMode()
	case "x", "X":
		return m.cancelDestination()
	case "d", "D":
		return m.openDiplomacyMenu()
	case "s", "S":
		return m.saveGame()
	case "i", "I":
		return m.startImprovement()
	case "v", "V":
		// If on own city, show city details; otherwise inspect tile
		city := m.Game.GetCityAt(m.CursorX, m.CursorY)
		if city != nil && city.CivID == 1 {
			m.ActiveMenu = MenuCity
			m.MenuCursor = 0
		} else {
			m.ActiveMenu = MenuInspect
		}
		return m, nil
	case "enter":
		return m.endTurn()
	}
	return m, nil
}

func (m Model) handleMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.InSettings {
		return m.handleSettingsMenu(msg)
	}
	if m.InNewGame {
		return m.handleNewGameMenu(msg)
	}
	if m.InLoadGame {
		return m.handleLoadGameMenu(msg)
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
			m.InNewGame = true
			m.NewGameCursor = 0
		case 1: // Load Game
			m.LoadGameSaves = game.ListExistingSaves()
			if len(m.LoadGameSaves) == 0 {
				// No saves — stay on main menu, nothing to do
				return m, nil
			}
			m.InLoadGame = true
			m.LoadGameCursor = 0
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
	numItems := 2 // Language, Back
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
		if m.SettingsCursor == 0 {
			lang := i18n.GetLang()
			if lang > 0 {
				i18n.SetLang(lang - 1)
			} else {
				i18n.SetLang(i18n.Lang(int(i18n.LangCount) - 1))
			}
			i18n.SaveConfig()
		}
	case "right", "l":
		if m.SettingsCursor == 0 {
			lang := i18n.GetLang()
			if lang < i18n.Lang(int(i18n.LangCount)-1) {
				i18n.SetLang(lang + 1)
			} else {
				i18n.SetLang(0)
			}
			i18n.SaveConfig()
		}
	case "enter":
		if m.SettingsCursor == numItems-1 {
			m.InSettings = false
		}
	case "esc":
		m.InSettings = false
	}
	return m, nil
}

func (m Model) handleNewGameMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	numItems := 5 // Map Size, AI Civs, Difficulty, Start Game, Back
	switch msg.String() {
	case "up", "k":
		if m.NewGameCursor > 0 {
			m.NewGameCursor--
		}
	case "down", "j":
		if m.NewGameCursor < numItems-1 {
			m.NewGameCursor++
		}
	case "left", "h":
		switch m.NewGameCursor {
		case 0: // Map Size
			if m.SettingsMapSize > worldmap.MapSizeSmall {
				m.SettingsMapSize--
			}
		case 1: // AI Civs
			if m.SettingsNumAICivs > 1 {
				m.SettingsNumAICivs--
			}
		case 2: // Difficulty
			if m.SettingsDifficulty > 1 {
				m.SettingsDifficulty--
			}
		}
	case "right", "l":
		switch m.NewGameCursor {
		case 0: // Map Size
			if m.SettingsMapSize < worldmap.MapSizeLarge {
				m.SettingsMapSize++
			}
		case 1: // AI Civs
			if m.SettingsNumAICivs < 4 {
				m.SettingsNumAICivs++
			}
		case 2: // Difficulty
			if m.SettingsDifficulty < 3 {
				m.SettingsDifficulty++
			}
		}
	case "enter":
		switch m.NewGameCursor {
		case 3: // Start Game
			m.InNewGame = false
			m.startGame()
		case 4: // Back
			m.InNewGame = false
		}
	case "esc":
		m.InNewGame = false
	}
	return m, nil
}

func (m Model) handleLoadGameMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	saves := m.LoadGameSaves
	// items = saves + Back
	numItems := len(saves) + 1
	switch msg.String() {
	case "up", "k":
		if m.LoadGameCursor > 0 {
			m.LoadGameCursor--
		}
	case "down", "j":
		if m.LoadGameCursor < numItems-1 {
			m.LoadGameCursor++
		}
	case "enter":
		if m.LoadGameCursor < len(saves) {
			// Load the selected save
			s := saves[m.LoadGameCursor]
			g, err := game.LoadFromFile(s.Path)
			if err == nil {
				m.Game = g
				m.CurrentScreen = ScreenGame
				m.InLoadGame = false
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
			}
		} else {
			// Back
			m.InLoadGame = false
		}
	case "d", "D":
		// Delete the currently highlighted save
		if m.LoadGameCursor < len(saves) {
			s := saves[m.LoadGameCursor]
			game.DeleteSave(s.Path)
			m.LoadGameSaves = game.ListExistingSaves()
			if len(m.LoadGameSaves) == 0 {
				m.InLoadGame = false
			} else if m.LoadGameCursor >= len(m.LoadGameSaves) {
				m.LoadGameCursor = len(m.LoadGameSaves) - 1
			}
		}
	case "esc":
		m.InLoadGame = false
	}
	return m, nil
}

func (m Model) handleSaveGameMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	slots := m.SaveGameSlots
	numItems := len(slots) + 1 // slots + Back
	switch msg.String() {
	case "up", "k":
		if m.SaveGameCursor > 0 {
			m.SaveGameCursor--
		}
	case "down", "j":
		if m.SaveGameCursor < numItems-1 {
			m.SaveGameCursor++
		}
	case "enter":
		if m.SaveGameCursor < len(slots) {
			s := slots[m.SaveGameCursor]
			err := m.Game.SaveToFile(s.Path)
			if err != nil {
				m.Game.AddPlayerMessage(i18n.Tf("Failed to save: %s", err.Error()))
			} else {
				m.Game.AddPlayerMessage(i18n.Tf("Game saved to slot %d!", s.Slot))
			}
			m.InSaveGame = false
			m.ActiveMenu = MenuNone
		} else {
			// Back
			m.InSaveGame = false
		}
	case "esc", "s", "S":
		m.InSaveGame = false
	}
	return m, nil
}

func (m Model) moveCursorAndSelect(dx, dy int) (tea.Model, tea.Cmd) {
	m.moveCursor(dx, dy)
	u := m.Game.GetUnitAt(m.CursorX, m.CursorY)
	if u != nil && u.CivID == 1 && u.IsAlive() {
		m.SelectedUnit = u
		m.updateReachable()
	} else if m.SelectedUnit != nil {
		// Deselect when cursor moves away from any player unit
		m.SelectedUnit = nil
		m.ReachableTiles = nil
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
	units := m.Game.PlayerUnitsNeedingAttention()
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

// selectNextFocus cycles through units needing attention AND player cities.
// Used by the N key to let the player review all their assets.
func (m Model) selectNextFocus() Model {
	units := m.Game.PlayerUnitsNeedingAttention()
	var cities []*model.City
	for _, c := range m.Game.Cities {
		if c.CivID == 1 {
			cities = append(cities, c)
		}
	}

	total := len(units) + len(cities)
	if total == 0 {
		m.SelectedUnit = nil
		m.ReachableTiles = nil
		return m
	}

	// Find current position in the combined list
	currentIdx := -1
	if m.SelectedUnit != nil {
		for i, u := range units {
			if u.ID == m.SelectedUnit.ID {
				currentIdx = i
				break
			}
		}
	}
	if currentIdx == -1 {
		// Check if cursor is on a player city
		for i, c := range cities {
			if c.X == m.CursorX && c.Y == m.CursorY {
				currentIdx = len(units) + i
				break
			}
		}
	}

	nextIdx := (currentIdx + 1) % total

	if nextIdx < len(units) {
		u := units[nextIdx]
		m.SelectedUnit = u
		m.CursorX = u.X
		m.CursorY = u.Y
	} else {
		c := cities[nextIdx-len(units)]
		m.SelectedUnit = nil
		m.CursorX = c.X
		m.CursorY = c.Y
	}
	m.updateReachable()
	m.scrollViewportToCursor()
	return m
}

// handleDestMode processes key events while the player is choosing a movement
// destination for the selected unit.
func (m Model) handleDestMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "g", "G":
		m.DestMode = false
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
		if m.SelectedUnit != nil {
			tile := m.Game.Map.GetTile(m.CursorX, m.CursorY)
			if tile != nil && model.Terrains[tile.Terrain].Passable {
				// Same tile as unit – treat as cancel
				if m.CursorX == m.SelectedUnit.X && m.CursorY == m.SelectedUnit.Y {
					m.SelectedUnit.HasDest = false
				} else {
					m.SelectedUnit.HasDest = true
					m.SelectedUnit.DestX = m.CursorX
					m.SelectedUnit.DestY = m.CursorY
					m.Game.AddPlayerMessage(i18n.Tf("Set destination to (%d,%d)", m.CursorX, m.CursorY))
				}
			} else {
				m.Game.AddPlayerMessage(i18n.T("Cannot set destination on impassable terrain"))
			}
		}
		m.DestMode = false
		return m, nil
	}
	return m, nil
}

// enterDestMode activates destination-selection mode for the selected unit.
func (m Model) enterDestMode() (tea.Model, tea.Cmd) {
	if m.SelectedUnit == nil || !m.SelectedUnit.IsAlive() {
		return m, nil
	}
	m.DestMode = true
	m.Game.AddPlayerMessage(i18n.T("Goto mode: move cursor to destination, Enter to confirm, Esc to cancel"))
	return m, nil
}

// cancelDestination clears the movement destination of the selected unit.
func (m Model) cancelDestination() (tea.Model, tea.Cmd) {
	if m.SelectedUnit != nil && m.SelectedUnit.HasDest {
		m.SelectedUnit.HasDest = false
		m.Game.AddPlayerMessage(i18n.T("Destination cancelled"))
		m.updateReachable()
	}
	return m, nil
}
