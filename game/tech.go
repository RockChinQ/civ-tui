package game

type Tech struct {
	Name     string
	Cost     int
	Requires []string
}

var AllTechs = []*Tech{
	{Name: "Agriculture", Cost: 20, Requires: []string{}},
	{Name: "Pottery", Cost: 25, Requires: []string{"Agriculture"}},
	{Name: "Mining", Cost: 35, Requires: []string{}},
	{Name: "Bronze Working", Cost: 50, Requires: []string{"Mining"}},
	{Name: "Iron Working", Cost: 80, Requires: []string{"Bronze Working"}},
	{Name: "Writing", Cost: 55, Requires: []string{"Pottery"}},
	{Name: "Archery", Cost: 35, Requires: []string{}},
	{Name: "Animal Husbandry", Cost: 35, Requires: []string{"Agriculture"}},
	{Name: "Horseback Riding", Cost: 60, Requires: []string{"Animal Husbandry"}},
	{Name: "Calendar", Cost: 45, Requires: []string{"Pottery"}},
	{Name: "The Wheel", Cost: 55, Requires: []string{"Mining"}},
}

func GetTech(name string) *Tech {
	for _, t := range AllTechs {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func AvailableTechs(civTechs map[string]bool) []*Tech {
	var result []*Tech
	for _, t := range AllTechs {
		if civTechs[t.Name] {
			continue
		}
		available := true
		for _, req := range t.Requires {
			if !civTechs[req] {
				available = false
				break
			}
		}
		if available {
			result = append(result, t)
		}
	}
	return result
}
