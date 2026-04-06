package tui

import (
	"github.com/RockChinQ/civ-tui/game"
	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/game/worldmap"
	tea "github.com/charmbracelet/bubbletea"
)

type Screen int

const (
	ScreenMainMenu Screen = iota
	ScreenGame
)

type MenuType int

const (
	MenuNone MenuType = iota
	MenuBuild
	MenuTech
	MenuHelp
	MenuCity
	MenuRanged
	MenuDiplomacy
	MenuPromotion
	MenuInspect
)

type Model struct {
	Game               *game.Game
	CursorX            int
	CursorY            int
	ViewportX          int
	ViewportY          int
	SelectedUnit       *model.Unit
	ActiveMenu         MenuType
	MenuCursor         int
	Width              int
	Height             int
	MapWidth           int
	MapHeight          int
	InfoWidth          int
	MsgHeight          int
	CurrentScreen      Screen
	MainMenuCursor     int
	InSettings         bool
	SettingsCursor     int
	SettingsMapSize    worldmap.MapSize
	SettingsNumAICivs  int
	SettingsDifficulty int
	RangeMode          bool
	PendingPromotion   *model.Unit
	ReachableTiles     map[[2]int]bool // cached reachable tiles for selected unit
}

func NewModel() Model {
	m := Model{
		Width:              120,
		Height:             35,
		InfoWidth:          28,
		MsgHeight:          6,
		CurrentScreen:      ScreenMainMenu,
		SettingsMapSize:    worldmap.MapSizeMedium,
		SettingsNumAICivs:  1,
		SettingsDifficulty: 1,
	}
	return m
}

func (m *Model) startGame() {
	opts := game.GameOptions{
		NumAICivs:  m.SettingsNumAICivs,
		MapSize:    m.SettingsMapSize,
		Difficulty: m.SettingsDifficulty,
	}
	g := game.NewGame(opts)
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
}

func (m *Model) centerViewport() {
	mapW, mapH := m.mapViewSize()
	m.ViewportX = m.CursorX - mapW/2
	m.ViewportY = m.CursorY - mapH/2
	m.clampViewport()
}

func (m *Model) clampViewport() {
	mapW, mapH := m.mapViewSize()
	mw, mh := m.gameMapSize()
	maxVX := mw - mapW
	maxVY := mh - mapH
	if maxVX < 0 {
		maxVX = 0
	}
	if maxVY < 0 {
		maxVY = 0
	}
	if m.ViewportX < 0 {
		m.ViewportX = 0
	}
	if m.ViewportX > maxVX {
		m.ViewportX = maxVX
	}
	if m.ViewportY < 0 {
		m.ViewportY = 0
	}
	if m.ViewportY > maxVY {
		m.ViewportY = maxVY
	}
}

func (m *Model) gameMapSize() (int, int) {
	if m.Game != nil {
		return m.Game.Map.Width, m.Game.Map.Height
	}
	return worldmap.MapWidth, worldmap.MapHeight
}

func (m *Model) mapViewSize() (int, int) {
	headerH := 1
	msgH := m.MsgHeight + 2
	availH := m.Height - headerH - msgH - 2
	availW := (m.Width - m.InfoWidth - 2) / 2

	if availW < 5 {
		availW = 5
	}
	if availH < 5 {
		availH = 5
	}
	mw, mh := m.gameMapSize()
	if availW > mw {
		availW = mw
	}
	if availH > mh {
		availH = mh
	}
	return availW, availH
}

func (m Model) Init() tea.Cmd {
	return nil
}

// updateReachable recalculates the reachable tile cache for the selected unit.
func (m *Model) updateReachable() {
	if m.Game != nil && m.SelectedUnit != nil && m.SelectedUnit.IsAlive() && m.SelectedUnit.HasMoves() {
		m.ReachableTiles = m.Game.ReachableTiles(m.SelectedUnit)
	} else {
		m.ReachableTiles = nil
	}
}
