package tui

import (
	"fmt"
	"strings"

	"github.com/RockChinQ/civ-tui/game"
	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/i18n"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderMenu() string {
	var sb strings.Builder
	playerCiv := m.Game.GetCiv(1)
	var civTechs map[string]bool
	if playerCiv != nil {
		civTechs = playerCiv.Techs
	}

	switch m.ActiveMenu {
	case MenuBuild:
		city := m.Game.GetCityAt(m.CursorX, m.CursorY)
		items := buildMenuItems(city, civTechs)
		sb.WriteString(StyleSectionTitle.Render(i18n.T("BUILD MENU")) + "\n")
		for i, item := range items {
			line := i18n.Tf("%s (cost %d)", i18n.T(item.Name), item.Cost)
			if i == m.MenuCursor {
				sb.WriteString(StyleSelectedUnit.Render("> "+line) + "\n")
			} else {
				sb.WriteString("  " + line + "\n")
			}
		}
		sb.WriteString(StyleDim.Render(i18n.T("[Enter]=select  [Esc]=cancel")))
	case MenuTech:
		sb.WriteString(StyleSectionTitle.Render(i18n.T("TECH MENU")) + "\n")
		if playerCiv != nil {
			available := model.AvailableTechs(playerCiv.Techs)
			if len(available) == 0 {
				sb.WriteString(i18n.T("No techs available") + "\n")
			}
			for i, t := range available {
				// Show what this tech unlocks
				unlocks := techUnlocks(t.Name)
				line := i18n.Tf("%s (cost %d)", i18n.T(t.Name), t.Cost) + unlocks
				if i == m.MenuCursor {
					sb.WriteString(StyleSelectedUnit.Render("> "+line) + "\n")
				} else {
					sb.WriteString("  " + line + "\n")
				}
			}
		}
		sb.WriteString(StyleDim.Render(i18n.T("[Enter]=select  [Esc]=cancel")))
	}
	return sb.String()
}

func techUnlocks(techName string) string {
	var unlocks []string
	for _, ud := range model.UnitDefs {
		if ud.RequiresTech == techName {
			unlocks = append(unlocks, i18n.T(ud.Name))
		}
	}
	for _, bd := range model.BuildingDefs {
		if bd.RequiresTech == techName {
			unlocks = append(unlocks, i18n.T(bd.Name))
		}
	}
	if len(unlocks) == 0 {
		return ""
	}
	return " → " + strings.Join(unlocks, ", ")
}

func (m Model) renderCityDetails(city *model.City) string {
	if city == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render(i18n.Tf("CITY: %s", city.Name)) + "\n\n")
	sb.WriteString(i18n.Tf("Population: %d  HP: %d/%d", city.Population, city.HP, city.MaxHP) + "\n")
	sb.WriteString(i18n.Tf("Food: %d/%d  Production: %d",
		city.Food, city.FoodNeeded, city.Production) + "\n")
	tile := m.Game.Map.GetTile(city.X, city.Y)
	sb.WriteString(i18n.Tf("Yields: Food %d  Prod %d  Gold %d  Sci %d",
		city.FoodYield(tile),
		city.ProductionYield(tile),
		city.GoldYield(tile),
		city.ScienceYield()) + "\n")
	sb.WriteString("\n" + i18n.T("Buildings:") + "\n")
	if len(city.Buildings) == 0 {
		sb.WriteString("  " + i18n.T("(none)") + "\n")
	}
	for bt := range city.Buildings {
		sb.WriteString("  " + i18n.T(model.BuildingDefs[bt].Name) + "\n")
	}
	sb.WriteString("\n" + i18n.T("Production Queue:") + "\n")
	if len(city.ProductionQ) == 0 {
		sb.WriteString("  " + i18n.T("(empty)") + "\n")
	}
	for i, item := range city.ProductionQ {
		marker := "  "
		if i == 0 {
			marker = "→ "
		}
		sb.WriteString(fmt.Sprintf("%s%s (%d/%d)\n", marker, i18n.T(item.Name), city.Production, item.Cost))
	}
	sb.WriteString("\n" + StyleDim.Render(i18n.T("[Enter/Esc] Close")))
	return sb.String()
}

func (m Model) renderDiplomacy() string {
	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render(i18n.T("DIPLOMACY")) + "\n\n")
	civs := m.diplomacyCivs()
	if len(civs) == 0 {
		sb.WriteString(i18n.T("No other civilizations") + "\n")
	}
	for i, c := range civs {
		rel := m.Game.GetRelation(1, c.ID)
		relStr := i18n.T("Peace")
		relStyle := StyleGreen
		if rel == model.RelationWar {
			relStr = i18n.T("War")
			relStyle = StyleRed
		}
		line := fmt.Sprintf("%s: %s", i18n.T(c.Name), relStr)
		action := i18n.T(" [Enter=declare war]")
		if rel == model.RelationWar {
			action = i18n.T(" [Enter=make peace]")
		}
		if i == m.MenuCursor {
			sb.WriteString(StyleSelectedUnit.Render("> "+line) + relStyle.Render(action) + "\n")
		} else {
			sb.WriteString("  " + line + "\n")
		}
	}
	sb.WriteString("\n" + StyleDim.Render(i18n.T("[Enter]=toggle war/peace  [Esc]=close")))
	return sb.String()
}

func (m Model) renderGameOver() string {
	var msg string
	switch m.Game.State {
	case game.StateVictory:
		msg = StyleGreen.Render(i18n.T("VICTORY! Press Q to quit."))
	case game.StateDefeat:
		msg = StyleRed.Render(i18n.T("DEFEAT! Press Q to quit."))
	case game.StateDraw:
		msg = StyleYellow.Render(i18n.T("DRAW - Turn limit reached! Press Q to quit."))
	}
	return StyleBold.Render(msg)
}

// overlayCenter places the popup string centered on top of the background string.
// It uses lipgloss.Place to correctly center the popup (handling multi-byte UTF-8
// and wide characters), then replaces background lines with the centered popup lines
// only for the rows where the popup appears.
func overlayCenter(bg, popup string, screenW, screenH int) string {
	bgLines := strings.Split(bg, "\n")

	// Pad background to fill screen height
	for len(bgLines) < screenH {
		bgLines = append(bgLines, "")
	}

	// Use lipgloss.Place to center the popup in a full-screen frame.
	// This correctly handles multi-byte characters and ANSI sequences.
	placed := lipgloss.Place(screenW, screenH, lipgloss.Center, lipgloss.Center, popup)
	placedLines := strings.Split(placed, "\n")

	// Determine which rows contain the popup content.
	popupLines := strings.Split(popup, "\n")
	popupH := len(popupLines)
	startY := (screenH - popupH) / 2
	if startY < 0 {
		startY = 0
	}

	// Replace only the popup rows in the background with the placed (centered) lines.
	for i := startY; i < startY+popupH && i < len(bgLines) && i < len(placedLines); i++ {
		bgLines[i] = placedLines[i]
	}

	return strings.Join(bgLines, "\n")
}

func (m Model) renderConfirmEndTurn() string {
	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render(i18n.T("END TURN")) + "\n\n")
	sb.WriteString(i18n.T("Are you sure you want to end the turn?") + "\n\n")

	items := []string{i18n.T("Yes"), i18n.T("No")}
	for i, item := range items {
		if i == m.ConfirmCursor {
			sb.WriteString(StyleSelectedUnit.Render("> "+item) + "\n")
		} else {
			sb.WriteString("  " + item + "\n")
		}
	}
	sb.WriteString("\n" + StyleDim.Render(i18n.T("[Enter] Confirm  [Esc] Cancel")))
	return StylePopup.Width(40).Render(sb.String())
}

func (m Model) renderInspect() string {
	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render(i18n.T("TILE INSPECT")) + "\n\n")

	tile := m.Game.Map.GetTile(m.CursorX, m.CursorY)
	sb.WriteString(i18n.Tf("Position: (%d, %d)", m.CursorX, m.CursorY) + "\n")

	if tile == nil || !tile.Revealed {
		sb.WriteString("\n" + StyleDim.Render(i18n.T("Unexplored territory")) + "\n")
		sb.WriteString("\n" + StyleDim.Render(i18n.T("[Esc/V] Close")))
		return sb.String()
	}

	// Terrain
	terrain := model.Terrains[tile.Terrain]
	sb.WriteString("\n" + StyleBold.Render(i18n.T("Terrain")) + "\n")
	tStyle := terrainStyle(tile.Terrain)
	sb.WriteString(fmt.Sprintf("  %s %s\n", tStyle.Render(terrain.Symbol), i18n.T(terrain.Name)))
	sb.WriteString(i18n.Tf("  Food: %d  Prod: %d  Gold: %d", terrain.Food, terrain.Production, terrain.Gold) + "\n")
	sb.WriteString(i18n.Tf("  Move Cost: %d", terrain.MoveCost))
	if terrain.DefenseBonus > 0 {
		sb.WriteString(i18n.Tf("  Defense: +%d%%", terrain.DefenseBonus))
	}
	if !terrain.Passable {
		sb.WriteString("  " + StyleRed.Render(i18n.T("Impassable")))
	}
	sb.WriteString("\n")

	// Improvement
	if tile.Improvement != model.ImprovementNone {
		imp := model.Improvements[tile.Improvement]
		sb.WriteString("\n" + StyleBold.Render(i18n.T("Improvement")) + "\n")
		sb.WriteString("  " + i18n.T(imp.Name) + "\n")
		sb.WriteString(i18n.Tf("  Food +%d  Prod +%d  Gold +%d", imp.FoodBonus, imp.ProdBonus, imp.GoldBonus) + "\n")
	}

	// Unit
	unit := m.Game.GetUnitAt(m.CursorX, m.CursorY)
	if unit != nil && (tile.Visible || unit.CivID == 1) {
		stats := model.UnitDefs[unit.Type]
		civ := m.Game.GetCiv(unit.CivID)
		owner := ""
		if civ != nil {
			owner = " - " + i18n.T(civ.Name)
		}
		sb.WriteString("\n" + StyleBold.Render(i18n.T("Unit")) + "\n")
		sb.WriteString(fmt.Sprintf("  %s (%s%s)\n", i18n.T(stats.Name), stats.Symbol, owner))
		sb.WriteString(i18n.Tf("  HP: %d/%d  Move: %d/%d", unit.HP, unit.MaxHP, unit.MovesLeft, unit.MaxMoves) + "\n")
		sb.WriteString(i18n.Tf("  Atk: %d  Def: %d  XP: %d  Lv: %d", unit.Attack, unit.Defense, unit.XP, unit.Level) + "\n")
		if stats.Range > 0 {
			sb.WriteString(i18n.Tf("  Range: %d", stats.Range) + "\n")
		}
		if unit.BuildingImprovement != model.ImprovementNone {
			impName := i18n.T(model.Improvements[unit.BuildingImprovement].Name)
			sb.WriteString(i18n.Tf("  Building: %s (%d turns left)", impName, unit.ImprovementTurnsLeft) + "\n")
		}
	}

	// City
	city := m.Game.GetCityAt(m.CursorX, m.CursorY)
	if city != nil {
		civ := m.Game.GetCiv(city.CivID)
		owner := ""
		if civ != nil {
			owner = " - " + i18n.T(civ.Name)
		}
		sb.WriteString("\n" + StyleBold.Render(i18n.T("City")) + "\n")
		sb.WriteString(fmt.Sprintf("  %s%s\n", city.Name, owner))
		sb.WriteString(i18n.Tf("  Pop: %d  HP: %d/%d  Def: %d", city.Population, city.HP, city.MaxHP, city.Defense) + "\n")
		if city.CivID == 1 || tile.Visible {
			sb.WriteString(i18n.Tf("  Food: %d  Prod: %d  Gold: %d  Sci: %d",
				city.FoodYield(tile), city.ProductionYield(tile), city.GoldYield(tile), city.ScienceYield()) + "\n")
			if len(city.Buildings) > 0 {
				sb.WriteString(i18n.T("  Buildings: "))
				first := true
				for bt := range city.Buildings {
					if !first {
						sb.WriteString(", ")
					}
					sb.WriteString(i18n.T(model.BuildingDefs[bt].Name))
					first = false
				}
				sb.WriteString("\n")
			}
		}
	}

	// Visibility
	visibility := i18n.T("Visible")
	if !tile.Visible {
		visibility = i18n.T("Revealed (not in sight)")
	}
	sb.WriteString("\n" + StyleDim.Render(i18n.T("Visibility: ")+visibility))
	sb.WriteString("\n" + StyleDim.Render(i18n.T("[Esc/V] Close")))
	return sb.String()
}
