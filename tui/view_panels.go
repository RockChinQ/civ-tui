package tui

import (
	"fmt"
	"strings"

	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/i18n"
)

func (m Model) renderInfo() string {
	var sb strings.Builder
	w := m.InfoWidth - 4

	sb.WriteString(StyleSectionTitle.Render(i18n.T("SELECTED UNIT")) + "\n")
	if m.SelectedUnit != nil && m.SelectedUnit.IsAlive() {
		u := m.SelectedUnit
		stats := model.UnitDefs[u.Type]
		sb.WriteString(fmt.Sprintf("%s (HP %d/%d)\n", i18n.T(stats.Name), u.HP, u.MaxHP))
		sb.WriteString(i18n.Tf("Move: %d/%d  XP: %d  Lv: %d", u.MovesLeft, u.MaxMoves, u.XP, u.Level) + "\n")
		sb.WriteString(i18n.Tf("Atk: %d  Def: %d", u.Attack, u.Defense) + "\n")
		sb.WriteString(i18n.Tf("Pos: (%d, %d)", u.X, u.Y) + "\n")
		if u.HasDest {
			sb.WriteString(StyleMovingUnit.Render(i18n.Tf("→ Dest: (%d, %d)", u.DestX, u.DestY)) + "\n")
		}
		if u.IsBusy() {
			impName := i18n.T(model.Improvements[u.BuildingImprovement].Name)
			sb.WriteString(StyleBusyUnit.Render(i18n.Tf("★ Building: %s", impName)) + "\n")
		}
		tile := m.Game.Map.GetTile(u.X, u.Y)
		if tile != nil {
			t := model.Terrains[tile.Terrain]
			sb.WriteString(i18n.Tf("Terrain: %s", i18n.T(t.Name)) + "\n")
			if t.DefenseBonus > 0 {
				sb.WriteString(i18n.Tf("Defense bonus: +%d%%", t.DefenseBonus) + "\n")
			}
		}
		// Show terrain info at cursor if cursor is not on unit
		if m.CursorX != u.X || m.CursorY != u.Y {
			curTile := m.Game.Map.GetTile(m.CursorX, m.CursorY)
			if curTile != nil && curTile.Revealed {
				ct := model.Terrains[curTile.Terrain]
				sb.WriteString(i18n.Tf("Cursor: %s (cost %d)", i18n.T(ct.Name), ct.MoveCost) + "\n")
			}
		}
	} else {
		tile := m.Game.Map.GetTile(m.CursorX, m.CursorY)
		sb.WriteString(i18n.Tf("Cursor: (%d, %d)", m.CursorX, m.CursorY) + "\n")
		if tile != nil && tile.Revealed {
			t := model.Terrains[tile.Terrain]
			sb.WriteString(i18n.Tf("Terrain: %s", i18n.T(t.Name)) + "\n")
			sb.WriteString(i18n.Tf("Yields: F%d P%d G%d", t.Food, t.Production, t.Gold) + "\n")
		} else {
			sb.WriteString(i18n.T("Unexplored") + "\n")
		}
		city := m.Game.GetCityAt(m.CursorX, m.CursorY)
		if city != nil {
			sb.WriteString(i18n.Tf("City: %s", city.Name) + "\n")
			sb.WriteString(i18n.Tf("Pop: %d  HP: %d/%d", city.Population, city.HP, city.MaxHP) + "\n")
			if len(city.ProductionQ) > 0 {
				first := city.ProductionQ[0]
				if len(city.ProductionQ) == 1 {
					sb.WriteString(i18n.Tf("Building: %s", i18n.T(first.Name)) + "\n")
				} else {
					sb.WriteString(i18n.Tf("Building: %s (+%d)", i18n.T(first.Name), len(city.ProductionQ)-1) + "\n")
				}
			} else {
				sb.WriteString(StyleYellow.Render(i18n.T("Production: idle")) + "\n")
			}
		}
	}

	sb.WriteString(StyleDim.Render(strings.Repeat("─", w)) + "\n")
	sb.WriteString(StyleSectionTitle.Render(i18n.T("ACTIONS")) + "\n")
	if m.SelectedUnit != nil {
		switch m.SelectedUnit.Type {
		case model.UnitSettler:
			sb.WriteString(i18n.T("[F] Found City") + "\n")
		case model.UnitWorker:
			sb.WriteString(i18n.T("[I] Build Improvement") + "\n")
		default:
			if model.UnitDefs[m.SelectedUnit.Type].Range > 0 {
				sb.WriteString(i18n.T("[R] Ranged Attack") + "\n")
			}
		}
		sb.WriteString(i18n.T("[G] Goto") + "\n")
		if m.SelectedUnit.HasDest {
			sb.WriteString(i18n.T("[X] Cancel Dest") + "\n")
		}
	}
	sb.WriteString(i18n.T("[W] Wait/Skip") + "\n")
	sb.WriteString(i18n.T("[N] Next Unit") + "\n")
	sb.WriteString(i18n.T("[B] Build Menu") + "\n")
	sb.WriteString(i18n.T("[T] Tech Menu") + "\n")
	sb.WriteString(i18n.T("[D] Diplomacy") + "\n")
	sb.WriteString(i18n.T("[S] Save Game") + "\n")
	sb.WriteString(i18n.T("[V] Inspect Tile") + "\n")
	sb.WriteString(i18n.T("[Enter] End Turn") + "\n")
	sb.WriteString(i18n.T("[?] Help") + "\n")

	sb.WriteString(StyleDim.Render(strings.Repeat("─", w)) + "\n")
	sb.WriteString(StyleSectionTitle.Render(i18n.T("RESEARCH")) + "\n")
	playerCiv := m.Game.GetCiv(1)
	if playerCiv != nil {
		if playerCiv.Researching != "" {
			tech := model.GetTech(playerCiv.Researching)
			sb.WriteString(i18n.Tf("Researching: %s", i18n.T(playerCiv.Researching)) + "\n")
			if tech != nil {
				sb.WriteString(i18n.Tf("Progress: %d/%d", playerCiv.ResearchProgress, tech.Cost) + "\n")
			}
		} else {
			sb.WriteString(StyleYellow.Render(i18n.T("Press [T] to research")) + "\n")
		}
		done := 0
		for _, t := range model.AllTechs {
			if playerCiv.Techs[t.Name] {
				done++
			}
		}
		sb.WriteString(i18n.Tf("Techs: %d/%d", done, len(model.AllTechs)) + "\n")
	}

	sb.WriteString(StyleDim.Render(strings.Repeat("─", w)) + "\n")
	sb.WriteString(StyleSectionTitle.Render(i18n.T("MINIMAP")) + "\n")
	sb.WriteString(m.renderMinimap())

	mapH := m.mapViewH()
	style := StyleInfoPanel.Width(m.InfoWidth).Height(mapH)
	return style.Render(sb.String())
}

func (m Model) renderMinimap() string {
	if m.Game == nil {
		return ""
	}
	mmW := 20
	mmH := 10
	if mmW > m.Game.Map.Width {
		mmW = m.Game.Map.Width
	}
	if mmH > m.Game.Map.Height {
		mmH = m.Game.Map.Height
	}
	scaleX := float64(m.Game.Map.Width) / float64(mmW)
	scaleY := float64(m.Game.Map.Height) / float64(mmH)

	var sb strings.Builder
	for my := 0; my < mmH; my++ {
		for mx := 0; mx < mmW; mx++ {
			mapX := int(float64(mx) * scaleX)
			mapY := int(float64(my) * scaleY)
			tile := m.Game.Map.GetTile(mapX, mapY)
			if tile == nil || !tile.Revealed {
				sb.WriteString(StyleFog.Render("░"))
				continue
			}
			// Check city
			hasPlayerCity := false
			hasEnemyCity := false
			for _, c := range m.Game.Cities {
				cx := int(float64(c.X) / scaleX)
				cy := int(float64(c.Y) / scaleY)
				if cx == mx && cy == my {
					if c.CivID == 1 {
						hasPlayerCity = true
					} else {
						hasEnemyCity = true
					}
				}
			}
			if hasPlayerCity {
				sb.WriteString(StylePlayerCity.Render("*"))
			} else if hasEnemyCity {
				sb.WriteString(StyleEnemyCity.Render("*"))
			} else {
				terrain := model.Terrains[tile.Terrain]
				style := terrainStyle(tile.Terrain)
				if !tile.Visible {
					style = style.Faint(true)
				}
				sb.WriteString(style.Render(terrain.Symbol))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (m Model) renderMessages() string {
	var sb strings.Builder
	msgs := m.Game.Messages
	start := len(msgs) - 8
	if start < 0 {
		start = 0
	}
	msgs = msgs[start:]
	sb.WriteString(StyleSectionTitle.Render(i18n.T("Messages:")) + "\n")
	for _, msg := range msgs {
		text := msg.Text
		if len(text) > m.Width-6 {
			text = text[:m.Width-6]
		}
		if msg.IsPlayer {
			sb.WriteString(StylePlayerMsg.Render("▸ "+text) + "\n")
		} else {
			sb.WriteString(StyleDim.Render("  "+text) + "\n")
		}
	}
	w := m.Width - 2
	if w < 10 {
		w = 10
	}
	return StyleMsgPanel.Width(w).Height(m.MsgHeight).Render(sb.String())
}
