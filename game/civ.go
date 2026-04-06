package game

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
}

func NewCiv(id int, name string, isPlayer bool) *Civ {
	return &Civ{
		ID:       id,
		Name:     name,
		Gold:     10,
		Science:  0,
		Techs:    make(map[string]bool),
		IsPlayer: isPlayer,
		IsAlive:  true,
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
