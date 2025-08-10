package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	IsRemoteTarget bool
	OutputDir      string
	Google         struct {
		APIKey    string
		SheetId   string
		SubmitUrl string
		ReportUrl string
	}
}

// loadConfig loads configuration from a JSON file into the given config struct.
func LoadConfig(filename string, config *Config) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return err
	}

	abs, err := filepath.Abs(config.OutputDir)
	if err != nil {
		return err
	}
	config.OutputDir = abs

	return nil
}
