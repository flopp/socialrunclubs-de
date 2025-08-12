package app

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/flopp/socialrunclubs-de/internal/utils"
)

type TemplateData struct {
	isRemoteTarget bool
	basePath       string
	LastUpdate     string
	Title          string
	Description    string
	Canonical      string
	SubmitUrl      string // URL to submit a new club
	ReportUrl      string // URL to report an issue with a club
	CssFiles       []string
	JSFiles        []string
	UmamiJS        string
	Data           *Data
	City           *City
	Club           *Club
}

func (t TemplateData) IsRemoteTarget() bool {
	return t.isRemoteTarget
}
func (t TemplateData) BasePath() string {
	return t.basePath
}
func (t TemplateData) ReportLink() string {
	if t.Club == nil {
		return strings.ReplaceAll(t.ReportUrl, "NAME", "")
	}

	encodedCanonical := url.PathEscape(t.Canonical)
	return strings.ReplaceAll(t.ReportUrl, "NAME", encodedCanonical)
}

func Render(data *Data, cssFiles, jsFiles []string, config Config) error {
	umamiJS := ""
	otherJS := make([]string, 0)
	// find umami.js file
	for _, jsFile := range jsFiles {
		if strings.Contains(jsFile, "umami") {
			umamiJS = jsFile
		} else {
			otherJS = append(otherJS, jsFile)
		}
	}

	// create indexnow.txt file
	if config.AHrefs.IndexNow != "" {
		filename := filepath.Join(config.OutputDir, config.AHrefs.IndexNow+".txt")
		if err := os.WriteFile(filename, []byte(config.AHrefs.IndexNow), 0644); err != nil {
			return fmt.Errorf("writing indexnow file %q: %w", filename, err)
		}
	}

	// render templates
	pages := []struct {
		Title       string
		Description string
		Canonical   string
		Template    string
		OutFile     string
	}{
		{
			Title:       "Social Run Clubs in Deutschland",
			Description: "Eine Übersicht über alle Social Run Clubs in Deutschland.",
			Canonical:   "https://socialrunclubs.de/",
			Template:    "index.html",
			OutFile:     "index.html",
		},
		{
			Title:       "Impressum",
			Description: "Impressum von socialrunclubs.de.",
			Canonical:   "https://socialrunclubs.de/impressum.html",
			Template:    "impressum.html",
			OutFile:     "impressum.html",
		},
		{
			Title:       "Datenschutz",
			Description: "Datenschutz von socialrunclubs.de.",
			Canonical:   "https://socialrunclubs.de/datenschutz.html",
			Template:    "datenschutz.html",
			OutFile:     "datenschutz.html",
		},
		{
			Title:       "Deutsche Städte mit Social Run Clubs",
			Description: "Eine Übersicht über alle Städte mit Social Run Clubs.",
			Canonical:   "https://socialrunclubs.de/cities.html",
			Template:    "cities.html",
			OutFile:     "cities.html",
		},
	}
	for _, page := range pages {
		tdata := TemplateData{
			Data:           data,
			City:           nil,
			Club:           nil,
			isRemoteTarget: config.IsRemoteTarget,
			basePath:       config.OutputDir,
			Title:          page.Title,
			Description:    page.Description,
			Canonical:      page.Canonical,
			SubmitUrl:      config.Google.SubmitUrl,
			ReportUrl:      config.Google.ReportUrl,
			CssFiles:       cssFiles,
			JSFiles:        otherJS,
			UmamiJS:        umamiJS,
		}
		if err := utils.ExecuteTemplate(page.Template, filepath.Join(config.OutputDir, page.OutFile), tdata); err != nil {
			return fmt.Errorf("rendering template %s: %w", page.Template, err)
		}
	}

	// render city pages
	for _, city := range data.Cities {
		tdata := TemplateData{
			Data:           data,
			City:           city,
			Club:           nil,
			isRemoteTarget: config.IsRemoteTarget,
			basePath:       config.OutputDir,
			Title:          fmt.Sprintf("Social Run Clubs in %s", city.Name),
			Description:    fmt.Sprintf("Eine Übersicht über alle Social Run Clubs in %s.", city.Name),
			Canonical:      fmt.Sprintf("https://socialrunclubs.de/%s", city.Slug()),
			SubmitUrl:      config.Google.SubmitUrl,
			ReportUrl:      config.Google.ReportUrl,
			CssFiles:       cssFiles,
			JSFiles:        otherJS,
			UmamiJS:        umamiJS,
		}
		fileName := filepath.Join(config.OutputDir, city.Slug(), "index.html")
		if err := utils.ExecuteTemplate("city.html", fileName, tdata); err != nil {
			return fmt.Errorf("rendering city template %q: %w", city.Name, err)
		}

		for _, club := range city.Clubs {
			tdata := TemplateData{
				Data:           data,
				City:           city,
				Club:           club,
				isRemoteTarget: config.IsRemoteTarget,
				basePath:       config.OutputDir,
				Title:          club.Name,
				Description:    fmt.Sprintf("Informationen zum Social Run Club '%s' in %s.", club.Name, city.Name),
				Canonical:      fmt.Sprintf("https://socialrunclubs.de/%s", club.Slug()),
				SubmitUrl:      config.Google.SubmitUrl,
				ReportUrl:      config.Google.ReportUrl,
				CssFiles:       cssFiles,
				JSFiles:        otherJS,
				UmamiJS:        umamiJS,
			}
			fileName := filepath.Join(config.OutputDir, club.Slug(), "index.html")
			if err := utils.ExecuteTemplate("club.html", fileName, tdata); err != nil {
				return fmt.Errorf("rendering club template %q: %w", club.Name, err)
			}
		}
	}

	return nil
}
