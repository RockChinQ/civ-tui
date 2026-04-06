package tui

import (
	"github.com/RockChinQ/civ-tui/game"
	tea "github.com/charmbracelet/bubbletea"
)

type MenuType int

const (
	MenuNone MenuType = iota
	MenuBuild
	MenuTech
	MenuHelp
)

type Model struct {
	Game         *game.Game
	CursorX      int
	CursorY      int
	ViewportX    int
	ViewportY    int
	SelectedUnit *game.Unit
	ActiveMenu   MenuType
	MenuCursor   int
	Width        int
	Height       int
	MapWidth     int
	MapHeight    int
	InfoWidth    int
	MsgHeight    int
}

func NewModel() Model {
	g := game.NewGame()
	m := Model{
		Game:      g,
		Width:     120,
		Height:    35,
		InfoWidth: 28,
		MsgHeight: 6,
	}
	// Start cursor at first player unit
	for _, u := range g.Units {
		if u.CivID == 1 && u.IsAlive() {
			m.CursorX = u.X
			m.CursorY = u.Y
			m.SelectedUnit = u
			break
		}
	}
	m.centerViewport()
	return m
}

func (m *Model) centerViewport() {
	mapW, mapH := m.mapViewSize()
	m.ViewportX = m.CursorX - mapW/2
	m.ViewportY = m.CursorY - mapH/2
	m.clampViewport()
}

func (m *Model) clampViewport() {
	mapW, mapH := m.mapViewSize()
	maxVX := game.MapWidth - mapW
	maxVY := game.MapHeight - mapH
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

func (m *Model) mapViewSize() (int, int) {
	// Each cell is 2 chars wide
	headerH := 1
	msgH := m.MsgHeight + 2 // border
	availH := m.Height - headerH - msgH - 2 // 2 for map borders
	availW := (m.Width - m.InfoWidth - 2) / 2 // 2 chars per cell, minus border

	if availW < 5 {
		availW = 5
	}
	if availH < 5 {
		availH = 5
	}
	if availW > game.MapWidth {
		availW = game.MapWidth
	}
	if availH > game.MapHeight {
		availH = game.MapHeight
	}
	return availW, availH
}

func (m Model) Init() tea.Cmd {
	return nil
}
