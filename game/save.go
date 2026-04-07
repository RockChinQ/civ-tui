package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const MaxSaveSlots = 10

// SaveInfo contains metadata about a save file for display in the UI.
type SaveInfo struct {
	Slot    int
	Path    string
	CivName string
	Turn    int
	ModTime string // formatted modification time
	IsEmpty bool
}

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

// SaveDir returns the directory where save files are stored.
func SaveDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".civ-tui", "saves")
}

// SlotPath returns the file path for the given save slot (1-based).
func SlotPath(slot int) string {
	return filepath.Join(SaveDir(), fmt.Sprintf("save_%d.json", slot))
}

// DefaultSavePath returns the path to save slot 1 (for backward compatibility).
func DefaultSavePath() string {
	return SlotPath(1)
}

// ListSaves returns save info for all slots (1..MaxSaveSlots).
// Empty slots have IsEmpty=true.
func ListSaves() []SaveInfo {
	var saves []SaveInfo

	for slot := 1; slot <= MaxSaveSlots; slot++ {
		path := SlotPath(slot)
		info := SaveInfo{Slot: slot, Path: path, IsEmpty: true}

		fi, err := os.Stat(path)
		if err != nil {
			saves = append(saves, info)
			continue
		}

		// Try to read metadata from the save
		data, err := os.ReadFile(path)
		if err != nil {
			saves = append(saves, info)
			continue
		}

		// Partial decode — only need Civs and Turn
		var partial struct {
			Civs []*struct {
				Name     string `json:"Name"`
				IsPlayer bool   `json:"IsPlayer"`
			} `json:"Civs"`
			Turn int `json:"Turn"`
		}
		if err := json.Unmarshal(data, &partial); err != nil {
			saves = append(saves, info)
			continue
		}

		civName := "?"
		for _, c := range partial.Civs {
			if c.IsPlayer {
				civName = c.Name
				break
			}
		}

		info.IsEmpty = false
		info.CivName = civName
		info.Turn = partial.Turn
		info.ModTime = fi.ModTime().Format("2006-01-02 15:04")
		saves = append(saves, info)
	}

	return saves
}

// ListExistingSaves returns only non-empty save slots, sorted by modification time (newest first).
func ListExistingSaves() []SaveInfo {
	all := ListSaves()
	var existing []SaveInfo
	for _, s := range all {
		if !s.IsEmpty {
			existing = append(existing, s)
		}
	}
	// Sort by mod time descending (newest first)
	sort.Slice(existing, func(i, j int) bool {
		return existing[i].ModTime > existing[j].ModTime
	})
	return existing
}

// DeleteSave removes the save file at the given path.
func DeleteSave(path string) error {
	return os.Remove(path)
}

// MigrateLegacySave moves the old single save.json to save_1.json if it exists.
func MigrateLegacySave() {
	dir := SaveDir()
	legacy := filepath.Join(dir, "save.json")
	if _, err := os.Stat(legacy); err != nil {
		return // no legacy save
	}
	slot1 := SlotPath(1)
	if _, err := os.Stat(slot1); err == nil {
		return // slot 1 already exists, don't overwrite
	}
	os.Rename(legacy, slot1)
}

// FindFirstEmptySlot returns the first empty slot number, or 0 if all full.
func FindFirstEmptySlot() int {
	dir := SaveDir()
	for slot := 1; slot <= MaxSaveSlots; slot++ {
		path := filepath.Join(dir, fmt.Sprintf("save_%d.json", slot))
		if _, err := os.Stat(path); err != nil {
			return slot
		}
	}
	return 0
}

// SlotLabel returns a display label for a save slot.
func (s SaveInfo) SlotLabel() string {
	if s.IsEmpty {
		return fmt.Sprintf("[%d] ---", s.Slot)
	}
	return fmt.Sprintf("[%d] %s  Turn %d  (%s)", s.Slot, s.CivName, s.Turn, s.ModTime)
}

// HasAnySaves returns true if there is at least one save file.
func HasAnySaves() bool {
	dir := SaveDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "save_") && strings.HasSuffix(e.Name(), ".json") {
			return true
		}
	}
	return false
}
