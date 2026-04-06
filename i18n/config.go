package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type config struct {
	Language string `json:"language"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".civ-tui", "config.json")
}

// LoadConfig loads language from config file.
// If no config exists, detects from system locale.
func LoadConfig() {
	data, err := os.ReadFile(configPath())
	if err != nil {
		SetLang(DetectSystemLang())
		return
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		SetLang(DetectSystemLang())
		return
	}
	switch cfg.Language {
	case "zh":
		SetLang(ZH)
	default:
		SetLang(EN)
	}
}

// SaveConfig persists the current language to the config file.
func SaveConfig() {
	lang := "en"
	if current == ZH {
		lang = "zh"
	}
	cfg := config{Language: lang}
	data, _ := json.Marshal(cfg)
	os.MkdirAll(filepath.Dir(configPath()), 0755)
	os.WriteFile(configPath(), data, 0644)
}

// DetectSystemLang checks environment variables for language hints.
func DetectSystemLang() Lang {
	for _, env := range []string{"LANG", "LC_ALL", "LANGUAGE"} {
		val := os.Getenv(env)
		if strings.HasPrefix(strings.ToLower(val), "zh") {
			return ZH
		}
	}
	return EN
}
