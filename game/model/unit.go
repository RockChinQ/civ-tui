package model

type UnitType int

const (
	UnitSettler UnitType = iota
	UnitScout
	UnitWarrior
	UnitArcher
	UnitSpearman
	UnitSwordsman
	UnitHorseman
	UnitWorker
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
	RequiresTech   string
	Range          int
}

var UnitDefs = map[UnitType]UnitStats{
	UnitSettler:   {UnitSettler, "Settler", "S", 10, 0, 1, 2, 30, "", 0},
	UnitScout:     {UnitScout, "Scout", "C", 10, 2, 1, 3, 15, "", 0},
	UnitWarrior:   {UnitWarrior, "Warrior", "W", 15, 4, 2, 2, 20, "", 0},
	UnitArcher:    {UnitArcher, "Archer", "A", 12, 5, 1, 2, 25, "Archery", 2},
	UnitSpearman:  {UnitSpearman, "Spearman", "P", 18, 3, 4, 2, 30, "Bronze Working", 0},
	UnitSwordsman: {UnitSwordsman, "Swordsman", "X", 20, 7, 3, 2, 40, "Iron Working", 0},
	UnitHorseman:  {UnitHorseman, "Horseman", "H", 15, 6, 2, 4, 35, "Horseback Riding", 0},
	UnitWorker:    {UnitWorker, "Worker", "K", 10, 0, 1, 2, 20, "", 0},
}

type Unit struct {
	ID                   int
	Type                 UnitType
	CivID                int
	X, Y                 int
	HP                   int
	MaxHP                int
	Attack               int
	Defense              int
	MovesLeft            int
	MaxMoves             int
	Waiting              bool
	XP                   int
	Level                int
	BuildingImprovement  ImprovementType
	ImprovementTurnsLeft int
	HasDest              bool // whether unit has a movement destination set
	DestX                int  // movement destination X coordinate
	DestY                int  // movement destination Y coordinate
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

// IsBusy returns true when the unit is performing a multi-turn task (e.g. building an improvement).
func (u *Unit) IsBusy() bool {
	return u.BuildingImprovement != ImprovementNone
}

// IsMovingToDest returns true when the unit has an active movement destination.
func (u *Unit) IsMovingToDest() bool {
	return u.HasDest
}

func (u *Unit) ResetMoves() {
	u.MovesLeft = u.MaxMoves
	u.Waiting = false
}

func AvailableUnits(civTechs map[string]bool) []UnitType {
	var result []UnitType
	order := []UnitType{UnitSettler, UnitWorker, UnitScout, UnitWarrior, UnitArcher, UnitSpearman, UnitSwordsman, UnitHorseman}
	for _, ut := range order {
		udef := UnitDefs[ut]
		if udef.RequiresTech == "" || civTechs[udef.RequiresTech] {
			result = append(result, ut)
		}
	}
	return result
}
