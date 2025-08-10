package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/flopp/socialrunclubs-de/internal"
	"github.com/flopp/socialrunclubs-de/internal/app"
)

func main() {
	// read config file from command line (e.g., config.json)
	configFile := flag.String("config", "config.json", "Path to the config file")
	flag.Parse()

	// load config from file
	config := app.Config{}
	if err := app.LoadConfig(*configFile, &config); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// get data from sheets
	data, err := app.GetData(config)
	if err != nil {
		log.Fatalf("Error processing sheets: %v", err)
	}

	// copy static files to output directory
	staticFiles := []string{"static/style.css"}
	for _, file := range staticFiles {
		dest := filepath.Join(config.OutputDir, "static", filepath.Base(file))
		if err := internal.CopyFile(file, dest); err != nil {
			log.Fatalf("Error copying static file %s to %s: %v", file, dest, err)
		}
	}

	// render pages
	if err := app.Render(data, config); err != nil {
		log.Fatalf("Error rendering data: %v", err)
	}
}
