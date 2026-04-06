package game

type UnitType int

const (
	UnitSettler UnitType = iota
	UnitScout
	UnitWarrior
	UnitArcher
	UnitSpearman
	UnitSwordsman
	UnitHorseman
)

type UnitStats struct {
	Type           UnitType
	Name           string
	Symbol         string
	MaxHP          int
	Attack         int
	Defense        int
	MaxMoves       int
	ProductionCost int
}

var UnitDefs = map[UnitType]UnitStats{
	UnitSettler:   {UnitSettler, "Settler", "S", 10, 0, 1, 2, 30},
	UnitScout:     {UnitScout, "Scout", "C", 10, 2, 1, 3, 15},
	UnitWarrior:   {UnitWarrior, "Warrior", "W", 15, 4, 2, 2, 20},
	UnitArcher:    {UnitArcher, "Archer", "A", 12, 5, 1, 2, 25},
	UnitSpearman:  {UnitSpearman, "Spearman", "P", 18, 3, 4, 2, 30},
	UnitSwordsman: {UnitSwordsman, "Swordsman", "X", 20, 7, 3, 2, 40},
	UnitHorseman:  {UnitHorseman, "Horseman", "H", 15, 6, 2, 4, 35},
}

type Unit struct {
	ID        int
	Type      UnitType
	CivID     int
	X, Y      int
	HP        int
	MaxHP     int
	Attack    int
	Defense   int
	MovesLeft int
	MaxMoves  int
	Waiting   bool
}

func NewUnit(id int, utype UnitType, civID, x, y int) *Unit {
	stats := UnitDefs[utype]
	return &Unit{
		ID:        id,
		Type:      utype,
		CivID:     civID,
		X:         x,
		Y:         y,
		HP:        stats.MaxHP,
		MaxHP:     stats.MaxHP,
		Attack:    stats.Attack,
		Defense:   stats.Defense,
		MovesLeft: stats.MaxMoves,
		MaxMoves:  stats.MaxMoves,
	}
}

func (u *Unit) IsAlive() bool {
	return u.HP > 0
}

func (u *Unit) HasMoves() bool {
	return u.MovesLeft > 0 && !u.Waiting
}

func (u *Unit) ResetMoves() {
	u.MovesLeft = u.MaxMoves
	u.Waiting = false
}
