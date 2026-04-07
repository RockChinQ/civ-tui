package model

type RelationType int

const (
	RelationPeace RelationType = 0
	RelationWar   RelationType = 1
)

type Civ struct {
	ID               int
	Name             string
	Gold             int
	Science          int
	GoldPerTurn      int
	SciPerTurn       int
	Techs            map[string]bool
	Researching      string
	ResearchProgress int
	IsPlayer         bool
	IsAlive          bool
	Relations        map[int]RelationType
	CityNames        []string
	PeaceTurns       map[int]int // civID → turn when peace was made
}

func NewCiv(id int, name string, isPlayer bool) *Civ {
	return &Civ{
		ID:         id,
		Name:       name,
		Gold:       InitialGold,
		Science:    0,
		Techs:      make(map[string]bool),
		IsPlayer:   isPlayer,
		IsAlive:    true,
		Relations:  make(map[int]RelationType),
		PeaceTurns: make(map[int]int),
	}
}

func (c *Civ) AddGold(n int) {
	c.Gold += n
}

func (c *Civ) HasTech(name string) bool {
	return c.Techs[name]
}

func (c *Civ) ResearchTech(tech *Tech) bool {
	if c.Techs[tech.Name] {
		return false
	}
	c.Researching = tech.Name
	return true
}

func (c *Civ) ProcessResearch(amount int, allTechs []*Tech) (completed string) {
	if c.Researching == "" {
		return ""
	}
	c.ResearchProgress += amount
	for _, t := range allTechs {
		if t.Name == c.Researching {
			if c.ResearchProgress >= t.Cost {
				c.Techs[t.Name] = true
				c.ResearchProgress = 0
				c.Researching = ""
				return t.Name
			}
			break
		}
	}
	return ""
}

// SetPeaceTurn records the turn when peace was made with the given civ.
func (c *Civ) SetPeaceTurn(otherID, turn int) {
	if c.PeaceTurns == nil {
		c.PeaceTurns = make(map[int]int)
	}
	c.PeaceTurns[otherID] = turn
}

// GetPeaceTurn returns the turn when peace was made with the given civ.
func (c *Civ) GetPeaceTurn(otherID int) (int, bool) {
	if c.PeaceTurns == nil {
		return 0, false
	}
	t, ok := c.PeaceTurns[otherID]
	return t, ok
}
