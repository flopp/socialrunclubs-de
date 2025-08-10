package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/flopp/socialrunclubs-de/internal"
)

type Config struct {
	IsRemoteTarget bool
	OutputDir      string
}

// loadConfig loads configuration from a JSON file into the given config struct.
func loadConfig(filename string, config *Config) error {
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

func main() {
	// read config file from command line (e.g., config.json)
	configFile := flag.String("config", "config.json", "Path to the config file")
	flag.Parse()

	// load config from file
	config := Config{}
	if err := loadConfig(*configFile, &config); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// use config
	fmt.Printf("Output directory: %s\n", config.OutputDir)

	// create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// copy static files to output directory
	staticFiles := []string{"static/style.css"}
	for _, file := range staticFiles {
		dest := filepath.Join(config.OutputDir, "static", filepath.Base(file))
		if err := internal.CopyFile(file, dest); err != nil {
			log.Fatalf("Error copying static file %s to %s: %v", file, dest, err)
		}
	}

	// render templates

	data := internal.TemplateData{
		IsRemoteTarget: config.IsRemoteTarget,
		BasePath:       config.OutputDir,
	}
	if err := internal.ExecuteTemplate("index.html", filepath.Join(config.OutputDir, "index.html"), data); err != nil {
		log.Fatalf("Error rendering template: %v", err)
	}
}
