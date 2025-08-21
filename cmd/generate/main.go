package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/flopp/socialrunclubs-de/internal/app"
	"github.com/flopp/socialrunclubs-de/internal/utils"
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

	// annotate city coordinates
	geocoder := utils.NewCachingGeocoder(utils.Download, fmt.Sprintf("%s/geocoder.json", config.CacheDir))
	if err := app.AnnotateCityCoordinates(data, geocoder); err != nil {
		log.Fatalf("Error annotating city coordinates: %v", err)
	}

	// copy static files to output directory
	cssFiles, jsFiles, err := app.CopyAssets(config)
	if err != nil {
		log.Fatalf("Error copying assets: %v", err)
	}

	// render pages
	if err := app.Render(data, cssFiles, jsFiles, config); err != nil {
		log.Fatalf("Error rendering data: %v", err)
	}
}
