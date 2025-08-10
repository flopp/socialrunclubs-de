package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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

	// get current time
	now := time.Now()
	nowStr := now.Format("2006-01-02 15:04:05")

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
		LastUpdate:     nowStr,
	}

	pages := []struct {
		Title     string
		Canonical string
		Template  string
		OutFile   string
	}{
		{
			Title:     "Social Run Clubs in Deutschland",
			Canonical: "https://socialrunclubs.de/",
			Template:  "index.html",
			OutFile:   "index.html",
		},
		{
			Title:     "Informationen",
			Canonical: "https://socialrunclubs.de/infos.html",
			Template:  "infos.html",
			OutFile:   "infos.html",
		},
		{
			Title:     "Deutsche Städte mit Social Run Clubs",
			Canonical: "https://socialrunclubs.de/cities.html",
			Template:  "cities.html",
			OutFile:   "cities.html",
		},
	}

	for _, page := range pages {
		data.Title = page.Title
		data.Canonical = page.Canonical
		if err := internal.ExecuteTemplate(page.Template, filepath.Join(config.OutputDir, page.OutFile), data); err != nil {
			log.Fatalf("Error rendering template %s: %v", page.Template, err)
		}
	}
}
