package tui

import "github.com/charmbracelet/lipgloss"

var (
	StyleBase = lipgloss.NewStyle()

	StyleHeader = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Padding(0, 1)

	StyleMapBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	StyleInfoPanel = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	StyleMsgPanel = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	// Terrain colors
	StyleOcean     = lipgloss.NewStyle().Foreground(lipgloss.Color("27"))
	StyleCoast     = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	StyleGrassland = lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
	StylePlains    = lipgloss.NewStyle().Foreground(lipgloss.Color("184"))
	StyleHills     = lipgloss.NewStyle().Foreground(lipgloss.Color("130"))
	StyleMountains = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	StyleForest    = lipgloss.NewStyle().Foreground(lipgloss.Color("28"))
	StyleDesert    = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	StyleTundra    = lipgloss.NewStyle().Foreground(lipgloss.Color("152"))

	// Unit/city colors
	StylePlayerUnit   = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	StyleEnemyUnit    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	StylePlayerCity   = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	StyleEnemyCity    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	StyleSelectedUnit = lipgloss.NewStyle().Background(lipgloss.Color("226")).Foreground(lipgloss.Color("0")).Bold(true)
	StyleCursor       = lipgloss.NewStyle().Reverse(true)
	StyleFog          = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	StyleSectionTitle = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("226"))
	StyleDim          = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	StyleBold         = lipgloss.NewStyle().Bold(true)
	StyleGreen        = lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
	StyleRed          = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	StyleYellow       = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
)

var (
	StyleRangeHighlight = lipgloss.NewStyle().Background(lipgloss.Color("52")).Foreground(lipgloss.Color("255"))
	StyleMoveHighlight  = lipgloss.NewStyle().Background(lipgloss.Color("22")).Foreground(lipgloss.Color("255"))
	// StyleBusyUnit: warm orange background – unit is performing a multi-turn task
	StyleBusyUnit = lipgloss.NewStyle().Background(lipgloss.Color("130")).Foreground(lipgloss.Color("255")).Bold(true)
	// StyleMovingUnit: teal background – unit is auto-moving toward a destination
	StyleMovingUnit = lipgloss.NewStyle().Background(lipgloss.Color("23")).Foreground(lipgloss.Color("255")).Bold(true)
	// StyleDestMarker: purple background – marks the current unit's movement destination
	StyleDestMarker = lipgloss.NewStyle().Background(lipgloss.Color("90")).Foreground(lipgloss.Color("255"))
)

var StylePopup = lipgloss.NewStyle().
	Border(lipgloss.DoubleBorder()).
	BorderForeground(lipgloss.Color("226")).
	Padding(1, 2)
