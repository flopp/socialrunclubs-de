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
	Config         Config
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
	Tag            *Tag
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

	canonical := func(p string) string {
		b := "https://socialrunclubs.de"
		if !strings.HasPrefix(p, "/") {
			b += "/"
		}
		b += p

		if !strings.Contains(p, ".") && !strings.HasSuffix(p, "/") {
			b += "/"
		}

		return b
	}

	// create indexnow.txt file
	if config.AHrefs.IndexNow != "" {
		filename := filepath.Join(config.OutputDir, config.AHrefs.IndexNow+".txt")
		if err := os.WriteFile(filename, []byte(config.AHrefs.IndexNow), 0644); err != nil {
			return fmt.Errorf("writing indexnow file %q: %w", filename, err)
		}
	}

	// collect all canonical URLs for creating a sitemap
	sitemapUrls := make([]string, 0)

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
			Canonical:   "/",
			Template:    "index.html",
			OutFile:     "index.html",
		},
		{
			Title:       "Impressum - Social Run Clubs",
			Description: "Impressum von socialrunclubs.de.",
			Canonical:   "/impressum.html",
			Template:    "impressum.html",
			OutFile:     "impressum.html",
		},
		{
			Title:       "Datenschutz - Social Run Clubs",
			Description: "Datenschutz von socialrunclubs.de.",
			Canonical:   "datenschutz.html",
			Template:    "datenschutz.html",
			OutFile:     "datenschutz.html",
		},
		{
			Title:       "Deutsche Städte mit Social Run Clubs",
			Description: "Eine Übersicht über alle Städte mit Social Run Clubs.",
			Canonical:   "/cities.html",
			Template:    "cities.html",
			OutFile:     "cities.html",
		},
		{
			Title:       "Deutsche Städte ohne Social Run Clubs",
			Description: "Eine Übersicht über alle Städte ohne Social Run Clubs.",
			Canonical:   "/cities-no-club.html",
			Template:    "cities-no-club.html",
			OutFile:     "cities-no-club.html",
		},
		{
			Title:       "Alphabetische Liste deutscher Social Run Clubs",
			Description: "Eine Übersicht über alle Social Run Clubs in Deutschland.",
			Canonical:   "/clubs.html",
			Template:    "clubs.html",
			OutFile:     "clubs.html",
		},
		{
			Title:       "Social Run Club Kategorien",
			Description: "Eine Übersicht über alle Social Run Club Kategorien.",
			Canonical:   "/tags.html",
			Template:    "tags.html",
			OutFile:     "tags.html",
		},
	}
	for _, page := range pages {
		tdata := TemplateData{
			Config:         config,
			Data:           data,
			City:           nil,
			Club:           nil,
			isRemoteTarget: config.IsRemoteTarget,
			basePath:       config.OutputDir,
			Title:          page.Title,
			Description:    page.Description,
			Canonical:      canonical(page.Canonical),
			SubmitUrl:      config.Google.SubmitUrl,
			ReportUrl:      config.Google.ReportUrl,
			CssFiles:       cssFiles,
			JSFiles:        otherJS,
			UmamiJS:        umamiJS,
		}
		if err := utils.ExecuteTemplate(page.Template, filepath.Join(config.OutputDir, page.OutFile), tdata); err != nil {
			return fmt.Errorf("rendering template %s: %w", page.Template, err)
		}

		sitemapUrls = append(sitemapUrls, tdata.Canonical)
	}

	// render city pages
	for _, city := range data.Cities {
		if len(city.Clubs) == 0 {
			continue
		}
		tdata := TemplateData{
			Config:         config,
			Data:           data,
			City:           city,
			Club:           nil,
			isRemoteTarget: config.IsRemoteTarget,
			basePath:       config.OutputDir,
			Title:          fmt.Sprintf("Social Run Clubs in %s", city.Name),
			Description:    fmt.Sprintf("Eine Übersicht über alle Social Run Clubs in %s.", city.Name),
			Canonical:      canonical(city.Slug()),
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
		sitemapUrls = append(sitemapUrls, tdata.Canonical)

		for _, club := range city.Clubs {
			tdata := TemplateData{
				Config:         config,
				Data:           data,
				City:           city,
				Club:           club,
				isRemoteTarget: config.IsRemoteTarget,
				basePath:       config.OutputDir,
				Title:          fmt.Sprintf("%s - ein Social Run Club in %s", club.Name, city.Name),
				Description:    fmt.Sprintf("Informationen und Links zum Social Run Club '%s' in %s.", club.Name, city.Name),
				Canonical:      canonical(club.Slug()),
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
			sitemapUrls = append(sitemapUrls, tdata.Canonical)
		}
	}
	for _, city := range data.Cities {
		if len(city.Clubs) != 0 {
			continue
		}
		tdata := TemplateData{
			Config:         config,
			Data:           data,
			City:           city,
			Club:           nil,
			isRemoteTarget: config.IsRemoteTarget,
			basePath:       config.OutputDir,
			Title:          fmt.Sprintf("Social Run Clubs in %s", city.Name),
			Description:    fmt.Sprintf("Eine Übersicht über alle Social Run Clubs in %s.", city.Name),
			Canonical:      canonical(city.Slug()),
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
		sitemapUrls = append(sitemapUrls, tdata.Canonical)
	}

	for _, tag := range data.Tags {
		tdata := TemplateData{
			Config:         config,
			Data:           data,
			Tag:            tag,
			isRemoteTarget: config.IsRemoteTarget,
			basePath:       config.OutputDir,
			Title:          fmt.Sprintf("Social Run Clubs in der Kategorie %s", tag.Name),
			Description:    fmt.Sprintf("Eine Übersicht über alle Social Run Clubs in der Kategorie %s.", tag.Name),
			Canonical:      canonical(tag.Slug()),
			SubmitUrl:      config.Google.SubmitUrl,
			ReportUrl:      config.Google.ReportUrl,
			CssFiles:       cssFiles,
			JSFiles:        otherJS,
			UmamiJS:        umamiJS,
		}
		fileName := filepath.Join(config.OutputDir, tag.Slug(), "index.html")
		if err := utils.ExecuteTemplate("tag.html", fileName, tdata); err != nil {
			return fmt.Errorf("rendering tag template %q: %w", tag.Name, err)
		}
		sitemapUrls = append(sitemapUrls, tdata.Canonical)
	}

	// render 404 page
	tdata := TemplateData{
		Config:         config,
		Data:           data,
		City:           nil,
		Club:           nil,
		isRemoteTarget: config.IsRemoteTarget,
		basePath:       config.OutputDir,
		Title:          "404 - Seite nicht gefunden",
		Description:    "Die von dir angeforderte Seite konnte nicht gefunden werden.",
		Canonical:      canonical("/404.html"),
		SubmitUrl:      config.Google.SubmitUrl,
		ReportUrl:      config.Google.ReportUrl,
		CssFiles:       cssFiles,
		JSFiles:        otherJS,
		UmamiJS:        umamiJS,
	}
	if err := utils.ExecuteTemplate("404.html", filepath.Join(config.OutputDir, "404.html"), tdata); err != nil {
		return fmt.Errorf("rendering 404 template: %w", err)
	}

	// create htaccess with error page & redirects
	htaccessFile := filepath.Join(config.OutputDir, ".htaccess")
	htaccessData := make([]byte, 0)
	// add 404 error document
	htaccessData = append(htaccessData, []byte("ErrorDocument 404 /404.html\n")...)
	htaccessData = append(htaccessData, []byte("\n")...)
	// add redirects
	for from, to := range data.Redirects {
		htaccessData = append(htaccessData, []byte(fmt.Sprintf("Redirect 301 %s %s\n", from, to))...)
	}
	if err := os.WriteFile(htaccessFile, htaccessData, 0644); err != nil {
		return fmt.Errorf("writing htaccess file: %w", err)
	}

	// create sitemap.xml
	sitemapFile := filepath.Join(config.OutputDir, "sitemap.xml")
	sitemapData := make([]byte, 0)
	sitemapData = append(sitemapData, []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")...)
	sitemapData = append(sitemapData, []byte("<urlset xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xsi:schemaLocation=\"http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd\" xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")...)
	for _, url := range sitemapUrls {
		sitemapData = append(sitemapData, []byte("  <url>\n")...)
		sitemapData = append(sitemapData, []byte(fmt.Sprintf("    <loc>%s</loc>\n", url))...)
		sitemapData = append(sitemapData, []byte("  </url>\n")...)
	}
	sitemapData = append(sitemapData, []byte("</urlset>\n")...)
	if err := os.WriteFile(sitemapFile, sitemapData, 0644); err != nil {
		return fmt.Errorf("writing sitemap file: %w", err)
	}

	return nil
}
