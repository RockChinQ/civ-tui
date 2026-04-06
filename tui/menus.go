package tui

import (
	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/i18n"
	tea "github.com/charmbracelet/bubbletea"
)

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
