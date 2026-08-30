package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/flopp/socialrunclubs-de/internal/app"
	"github.com/flopp/socialrunclubs-de/internal/utils"
)

func getStravaClubId(url string) (string, error) {
	re := regexp.MustCompile(`^https?://www\.strava\.com/clubs/([^/]+)/?$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid Strava club URL: %s", url)
	}
	return matches[1], nil
}

func extractStravaClubImageUrl(htmlFile string) (string, error) {
	htmlBytes, err := os.ReadFile(htmlFile)
	if err != nil {
		return "", fmt.Errorf("error reading Strava club HTML file: %w", err)
	}
	htmlContent := string(htmlBytes)

	reOgImage := regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)
	matches := reOgImage.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find og:image in Strava club HTML")
	}

	imageURL := matches[1]
	imageURL = strings.ReplaceAll(imageURL, "&amp;", "&")
	return imageURL, nil
}

func getDirectImagePath(item *app.Club) string {
	return filepath.Join("club-images", item.City.SanitizeName(), item.SanitizeName()+".jpg")
}

func getUrlImagePath(config app.Config, item *app.Club) string {
	return filepath.Join(config.CacheDir, "url_image", item.City.SanitizeName(), item.SanitizeName()+".jpg")
}

func getStravaImagePath(config app.Config, item *app.Club) string {
	stravaClubId, err := getStravaClubId(item.StravaClub)
	if err != nil {
		return ""
	}
	return filepath.Join(config.CacheDir, "strava", stravaClubId, "image.jpg")
}

func getInstagramImagePath(config app.Config, item *app.Club) string {
	profileName := item.InstagramProfile()
	if profileName == "" {
		return ""
	}
	return filepath.Join(config.CacheDir, "instagram", profileName, "image.jpg")
}

func getDirectImage(config app.Config, item *app.Club, directImage string, targetFile string) error {
	if !utils.FileExists(directImage) {
		return fmt.Errorf("direct image file does not exist: %s", directImage)
	}

	if err := utils.CopyFile(directImage, targetFile); err != nil {
		return fmt.Errorf("error copying direct image to target file: %w", err)
	}

	return nil
}

func getUrlImage(config app.Config, item *app.Club, urlImage string, targetFile string) error {
	if item.ImageURL == "" {
		return nil
	}

	if !utils.FileExists(urlImage) {
		err := utils.DownloadAgent(item.ImageURL, urlImage)
		if err != nil {
			return fmt.Errorf("error downloading URL image: %w", err)
		}
	}

	if err := utils.CopyFile(urlImage, targetFile); err != nil {
		return fmt.Errorf("error copying URL image to target file: %w", err)
	}

	return nil
}

func getStravaImage(config app.Config, item *app.Club, stravaImage string, targetFile string) error {
	stravaClubId, err := getStravaClubId(item.StravaClub)
	if err != nil {
		return fmt.Errorf("error getting Strava club ID: %w", err)
	}

	if !utils.FileExists(stravaImage) {
		targetStravaHtml := config.CacheDir + "/strava/" + stravaClubId + "/html"
		if !utils.FileExists(targetStravaHtml) {
			err := utils.DownloadAgent(item.StravaClub, targetStravaHtml)
			if err != nil {
				return fmt.Errorf("error downloading Strava club HTML: %w", err)
			}
		}

		imageUrl, err := extractStravaClubImageUrl(targetStravaHtml)
		if err != nil {
			return fmt.Errorf("error extracting Strava club image URL: %w", err)
		}

		err = utils.DownloadAgent(imageUrl, stravaImage)
		if err != nil {
			return fmt.Errorf("error downloading Strava club image: %w", err)
		}
	}

	if err := utils.CopyFile(stravaImage, targetFile); err != nil {
		return fmt.Errorf("error copying Strava club image to target file: %w", err)
	}

	return nil
}

type profileInfoResponse struct {
	Data struct {
		User struct {
			ProfilePicURLHD string `json:"profile_pic_url_hd"`
			ProfilePicURL   string `json:"profile_pic_url"`
		} `json:"user"`
	} `json:"data"`
}

func fetchProfilePictureURL(username string) (string, error) {
	if username == "" {
		return "", errors.New("username is empty")
	}

	endpoint := "https://www.instagram.com/api/v1/users/web_profile_info/?username=" + url.QueryEscape(username)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-IG-App-ID", "936619743392459")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request profile metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("Instagram API returned status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload profileInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode profile metadata: %w", err)
	}

	if payload.Data.User.ProfilePicURLHD != "" {
		return payload.Data.User.ProfilePicURLHD, nil
	}
	if payload.Data.User.ProfilePicURL != "" {
		return payload.Data.User.ProfilePicURL, nil
	}

	return "", errors.New("no profile image URL found for this user")
}

func getInsta1(targetImage string, cacheDir string, profileName string) error {
	targetProfileHtml := cacheDir + "/instagram/" + profileName + "/html"
	if !utils.FileExists(targetProfileHtml) {
		profileUrl := "https://www.instagram.com/" + profileName + "/"
		err := utils.DownloadAgent(profileUrl, targetProfileHtml)
		if err != nil {
			return fmt.Errorf("error downloading Instagram profile HTML: %w", err)
		}
	}

	// get <meta property="og:image" content="([^"]+)" /> from the downloaded HTML
	htmlBytes, err := os.ReadFile(targetProfileHtml)
	if err != nil {
		log.Printf("Error reading HTML file: %v", err)
		return fmt.Errorf("error reading Instagram profile HTML: %w", err)
	}
	htmlContent := string(htmlBytes)

	reOgImage := regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)
	matches := reOgImage.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		log.Printf("Could not find og:image in profile HTML for %s", profileName)
		return fmt.Errorf("could not find og:image in Instagram profile HTML for %s", profileName)
	}

	imageURL := matches[1]
	imageURL = strings.ReplaceAll(imageURL, "&amp;", "&")
	err = utils.Download(imageURL, targetImage)
	if err != nil {
		log.Printf("Error downloading Instagram profile image: %v", err)
		return fmt.Errorf("error downloading Instagram profile image: %w", err)
	}
	return nil
}

func getInsta2(targetImage string, cacheDir string, profileName string) error {
	imageURL, err := fetchProfilePictureURL(profileName)
	if err != nil {
		log.Printf("Error fetching Instagram profile picture URL for %s: %v", profileName, err)
		return fmt.Errorf("error fetching Instagram profile picture URL for %s: %w", profileName, err)
	}

	err = utils.Download(imageURL, targetImage)
	if err != nil {
		log.Printf("Error downloading Instagram profile image: %v", err)
		return fmt.Errorf("error downloading Instagram profile image: %w", err)
	}
	return nil
}

func getInstagramImage(config app.Config, item *app.Club, instagramImage string, targetFile string) error {
	profileName := item.InstagramProfile()
	if profileName == "" {
		return fmt.Errorf("no Instagram profile found for club: %s", item.Name)
	}

	if !utils.FileExists(instagramImage) {
		/*
			targetProfileHtml := config.CacheDir + "/instagram/" + profileName + "/html"
			if !utils.FileExists(targetProfileHtml) {
				profileUrl := "https://www.instagram.com/" + profileName + "/"
				err := utils.DownloadAgent(profileUrl, targetProfileHtml)
				if err != nil {
					return fmt.Errorf("error downloading Instagram profile HTML: %w", err)
				}
			}

			// get <meta property="og:image" content="([^"]+)" /> from the downloaded HTML
			htmlBytes, err := os.ReadFile(targetProfileHtml)
			if err != nil {
				log.Printf("Error reading HTML file: %v", err)
				return fmt.Errorf("error reading Instagram profile HTML: %w", err)
			}
			htmlContent := string(htmlBytes)

			reOgImage := regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)
			matches := reOgImage.FindStringSubmatch(htmlContent)
			if len(matches) < 2 {
				log.Printf("Could not find og:image in profile HTML for %s", profileName)
				return fmt.Errorf("could not find og:image in Instagram profile HTML for %s", profileName)
			}

			imageURL := matches[1]
			imageURL = strings.ReplaceAll(imageURL, "&amp;", "&")
		*/
		/*
			imageURL, err := fetchProfilePictureURL(profileName)
			if err != nil {
				log.Printf("Error fetching Instagram profile picture URL for %s: %v", profileName, err)
				return fmt.Errorf("error fetching Instagram profile picture URL for %s: %w", profileName, err)
			}

			err = utils.Download(imageURL, instagramImage)
			if err != nil {
				log.Printf("Error downloading Instagram profile image: %v", err)
				return fmt.Errorf("error downloading Instagram profile image: %w", err)
			}
		*/

		if err1 := getInsta1(instagramImage, config.CacheDir, profileName); err1 != nil {
			if err2 := getInsta2(instagramImage, config.CacheDir, profileName); err2 != nil {
				return fmt.Errorf("error downloading Instagram profile image for %s: %v, %v", profileName, err1, err2)
			}
		}
	}

	if err := utils.CopyFile(instagramImage, targetFile); err != nil {
		return fmt.Errorf("error copying Instagram profile image to target file: %w", err)
	}

	return nil
}

func main() {
	// read config file from command line (e.g., config.json)
	configFile := flag.String("config", "config.json", "Path to the config file")
	flag.Parse()

	//delayMin := 2
	//delayMax := 5

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
		targetImage := filepath.Join(config.ImageDir, item.City.SanitizeName(), item.SanitizeName()+".jpg")
		directImage := getDirectImagePath(item)
		urlImage := getUrlImagePath(config, item)
		stravaImage := getStravaImagePath(config, item)
		instagramImage := getInstagramImagePath(config, item)

		methods := []struct {
			name      string
			imagePath string
			fn        func() error
		}{
			{
				name:      "direct",
				imagePath: directImage,
				fn: func() error {
					return getDirectImage(config, item, directImage, targetImage)
				},
			},
			{
				name:      "url",
				imagePath: urlImage,
				fn: func() error {
					return getUrlImage(config, item, urlImage, targetImage)
				},
			},
			{
				name:      "strava",
				imagePath: stravaImage,
				fn: func() error {
					return getStravaImage(config, item, stravaImage, targetImage)
				},
			},
			{
				name:      "instagram",
				imagePath: instagramImage,
				fn: func() error {
					return getInstagramImage(config, item, instagramImage, targetImage)
				},
			},
		}

		// check if we already have any image in our caches
		existingImage := ""
		for _, method := range methods {
			if method.imagePath != "" && utils.FileExists(method.imagePath) {
				if err := method.fn(); err == nil {
					existingImage = method.imagePath
					break
				}
			}
		}
		if existingImage != "" {
			if err := utils.CopyFile(existingImage, targetImage); err != nil {
				log.Printf("Error copying existing image for club %s in city %s: %v", item.Name, item.City.Name, err)
			}
			continue
		}

		// try vto get new image
		for _, method := range methods {
			if method.imagePath != "" {
				if err := method.fn(); err != nil {
					log.Printf("Error downloading image using method %s for club %s in city %s: %v", method.name, item.Name, item.City.Name, err)
					continue
				}
				if utils.FileExists(targetImage) {
					log.Printf("Successfully downloaded image for club %s in city %s: %s", item.Name, item.City.Name, targetImage)
					break
				}
			}
		}

		if !utils.FileExists(targetImage) {
			log.Printf("No image found for club %s in city %s -> %s", item.Name, item.City.Name, directImage)
		}
	}
}
