package app

import (
	"fmt"
	"path/filepath"

	"github.com/flopp/socialrunclubs-de/internal"
)

type TemplateData struct {
	isRemoteTarget bool
	basePath       string
	LastUpdate     string
	Title          string
	Canonical      string
	SubmitUrl      string // URL to submit a new club
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

func Render(data *Data, config Config) error {
	// render templates
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
		tdata := TemplateData{
			Data:           data,
			City:           nil,
			Club:           nil,
			isRemoteTarget: config.IsRemoteTarget,
			basePath:       config.OutputDir,
			Title:          page.Title,
			Canonical:      page.Canonical,
		}
		if err := internal.ExecuteTemplate(page.Template, filepath.Join(config.OutputDir, page.OutFile), tdata); err != nil {
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
			Canonical:      fmt.Sprintf("https://socialrunclubs.de/%s", city.Slug()),
		}
		fileName := filepath.Join(config.OutputDir, city.Slug(), "index.html")
		if err := internal.ExecuteTemplate("city.html", fileName, tdata); err != nil {
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
				Canonical:      fmt.Sprintf("https://socialrunclubs.de/%s", club.Slug()),
			}
			fileName := filepath.Join(config.OutputDir, club.Slug(), "index.html")
			if err := internal.ExecuteTemplate("club.html", fileName, tdata); err != nil {
				return fmt.Errorf("rendering club template %q: %w", club.Name, err)
			}
		}
	}

	return nil
}
