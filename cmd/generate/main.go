package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/flopp/socialrunclubs-de/internal/app"
	"github.com/flopp/socialrunclubs-de/internal/utils"
)

func backup(config app.Config, backupFile string) error {
	fmt.Printf("-- requesting file %s...\n", config.Google.SheetId)
	exportURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=ods", config.Google.SheetId)
	_, err := utils.Retry(5, 2*time.Second, func() (struct{}, error) {
		return struct{}{}, utils.Download(exportURL, backupFile)
	})
	if err != nil {
		return fmt.Errorf("unable to download file from %s: %w", exportURL, err)
	}
	fmt.Printf("-- saved to %s...\n", backupFile)

	fmt.Println("-- done")
	return nil
}

func main() {
	// read config file from command line (e.g., config.json)
	configFile := flag.String("config", "config.json", "Path to the config file")
	backupFile := flag.String("backup", "", "backup sheets data to the specified file (optional)")
	linkCheck := flag.Bool("link-check", false, "check if all club links are reachable (optional)")
	flag.Parse()

	// load config from file
	config := app.Config{}
	if err := app.LoadConfig(*configFile, &config); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// backup sheets data if requested
	if *backupFile != "" {
		if err := backup(config, *backupFile); err != nil {
			log.Fatalf("Error backing up data: %v", err)
		}
		return
	}

	// get data from sheets
	data, err := app.GetData(config)
	if err != nil {
		log.Fatalf("Error processing sheets: %v", err)
	}

	if *linkCheck {
		if err := app.CheckLinks(data); err != nil {
			log.Fatalf("Error checking links: %v", err)
		}
		return
	}

	// annotate city coordinates
	geocoder := utils.NewCachingGeocoder(utils.Download, fmt.Sprintf("%s/geocoder.json", config.CacheDir))
	if err := app.AnnotateCityCoordinates(data, geocoder); err != nil {
		log.Fatalf("Error annotating city coordinates: %v", err)
	}
	if err := app.AnnotateNearestCities(data); err != nil {
		log.Fatalf("Error annotating nearest cities: %v", err)
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
