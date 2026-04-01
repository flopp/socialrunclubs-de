package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flopp/socialrunclubs-de/internal/app"
	"github.com/flopp/socialrunclubs-de/internal/utils"
)

func main() {
	// read config file from command line (e.g., config.json)
	configFile := flag.String("config", "config.json", "Path to the config file")
	flag.Parse()

	delayMin := 2
	delayMax := 5

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

	for _, item := range data.Clubs {
		profileName := item.InstagramProfile()
		if profileName != "" {
			targetProfileHtml := config.CacheDir + "/instagram/" + profileName + "/html"
			if !utils.FileExists(targetProfileHtml) {
				err := utils.Download(item.Instagram, targetProfileHtml)
				if err != nil {
					log.Printf("Error downloading Instagram profile: %v", err)
					break
				}

				// sleep for a random time between 1 and 3 seconds to avoid rate limiting
				utils.RandomSleep(delayMin, delayMax)
			}

			targetProfileImage := config.CacheDir + "/instagram/" + profileName + "/image.jpg"
			if !utils.FileExists(targetProfileImage) {
				// get <meta property="og:image" content="([^"]+)" /> from the downloaded HTML
				htmlBytes, err := os.ReadFile(targetProfileHtml)
				if err != nil {
					log.Printf("Error reading HTML file: %v", err)
					continue
				}
				htmlContent := string(htmlBytes)

				reOgImage := regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)
				matches := reOgImage.FindStringSubmatch(htmlContent)
				if len(matches) < 2 {
					log.Printf("Could not find og:image in profile HTML for %s", profileName)
					continue
				}

				imageURL := matches[1]
				imageURL = strings.ReplaceAll(imageURL, "&amp;", "&")
				err = utils.Download(imageURL, targetProfileImage)
				if err != nil {
					log.Printf("Error downloading Instagram profile image: %v", err)
					continue
				}
			}

			targetImage := config.ImageDir + "/" + profileName + ".jpg"
			if err := utils.CopyFile(targetProfileImage, targetImage); err != nil {
				log.Printf("Error copying Instagram profile image to target image: %v", err)
			}

			continue
		}

		directImage := filepath.Join("club-images", item.City.SanitizeName(), item.SanitizeName()+".jpg")
		if utils.FileExists(directImage) {
			targetImage := filepath.Join(config.ImageDir, item.City.SanitizeName(), item.SanitizeName()+".jpg")
			if err := utils.CopyFile(directImage, targetImage); err != nil {
				log.Printf("Error copying direct image to target image: %v", err)
			}
			continue
		}

		log.Printf("No image found for club %s in city %s -> %s", item.Name, item.City.Name, directImage)
	}
}
