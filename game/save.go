package game

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
)

func (g *Game) SaveToFile(path string) error {
	data, err := json.Marshal(g)
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, data, 0644)
}

func LoadFromFile(path string) (*Game, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var g Game
	if err = json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	g.Rand = rand.New(rand.NewSource(g.RandSeed))
	return &g, nil
}

func DefaultSavePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".civ-tui", "saves", "save.json")
}
