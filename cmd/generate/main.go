package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/flopp/socialrunclubs-de/internal/app"
	"github.com/flopp/socialrunclubs-de/internal/utils"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func backup(config app.Config, backupFile string) error {
	fmt.Println("-- connecting to Google Drive service...")
	ctx := context.Background()
	service, err := drive.NewService(ctx, option.WithAPIKey(config.Google.APIKey))
	if err != nil {
		return fmt.Errorf("unable to connect to Google Drive: %w", err)
	}

	fmt.Printf("-- requesting file %s...\n", config.Google.SheetId)
	response, err := service.Files.Export(config.Google.SheetId, "application/vnd.oasis.opendocument.spreadsheet").Download()
	if err != nil {
		return fmt.Errorf("unable to download file: %w", err)
	}
	defer response.Body.Close()

	fmt.Printf("-- saving to %s...\n", backupFile)
	file, err := os.Create(backupFile)
	if err != nil {
		return fmt.Errorf("unable to create output file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("unable to write to output file: %w", err)
	}

	fmt.Println("-- done")
	return nil
}

func main() {
	// read config file from command line (e.g., config.json)
	configFile := flag.String("config", "config.json", "Path to the config file")
	backupFile := flag.String("backup", "", "backup sheets data to the specified file (optional)")
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
