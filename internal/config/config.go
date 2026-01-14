package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Theme string `json:"theme"`
}

var (
	current = Config{
		Theme: "purple", // default theme
	}
	configPath string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	configDir := filepath.Join(home, ".config", "tsk")
	os.MkdirAll(configDir, 0755)
	configPath = filepath.Join(configDir, "config.json")
}

func Load() error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // use defaults
		}
		return err
	}
	return json.Unmarshal(data, &current)
}

func Save() error {
	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func Get() Config {
	return current
}

func SetTheme(theme string) {
	current.Theme = theme
}

func GetTheme() string {
	return current.Theme
}
