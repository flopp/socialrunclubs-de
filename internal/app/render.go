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
	Post           *Post
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

func FillDDescription(description string) string {
	return description
}

func createCanonicalURL(path string) string {
	b := "https://socialrunclubs.de"
	if !strings.HasPrefix(path, "/") {
		b += "/"
	}
	b += path

	if !strings.Contains(path, ".") && !strings.HasSuffix(path, "/") {
		b += "/"
	}

	return b
}

func processJSFiles(jsFiles []string) (umamiJS string, otherJS []string) {
	otherJS = make([]string, 0)
	for _, jsFile := range jsFiles {
		if strings.Contains(jsFile, "umami") {
			umamiJS = jsFile
		} else {
			otherJS = append(otherJS, jsFile)
		}
	}
	return
}

func createTemplateData(config Config, data *Data, title, description, canonical, submitUrl, reportUrl string, cssFiles, jsFiles []string, umamiJS string) TemplateData {
	return TemplateData{
		Config:         config,
		Data:           data,
		isRemoteTarget: config.IsRemoteTarget,
		basePath:       config.OutputDir,
		Title:          title,
		Description:    description,
		Canonical:      canonical,
		SubmitUrl:      submitUrl,
		ReportUrl:      reportUrl,
		CssFiles:       cssFiles,
		JSFiles:        jsFiles,
		UmamiJS:        umamiJS,
	}
}

func createTemplateDataWithEntities(config Config, data *Data, title, description, canonical, submitUrl, reportUrl string, cssFiles, jsFiles []string, umamiJS string, city *City, club *Club, tag *Tag, post *Post) TemplateData {
	tdata := createTemplateData(config, data, title, description, canonical, submitUrl, reportUrl, cssFiles, jsFiles, umamiJS)
	tdata.City = city
	tdata.Club = club
	tdata.Tag = tag
	tdata.Post = post
	return tdata
}

func renderStaticPages(data *Data, config Config, cssFiles, otherJS []string, umamiJS string, sitemapUrls *[]string) error {
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
			Title:       "Liste aller Social Run Clubs in Deutschland",
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
		{
			Title:       "Social Run Club Artikel",
			Description: "Eine Übersicht über alle Social Run Club Artikel.",
			Canonical:   "/post/",
			Template:    "posts.html",
			OutFile:     "post/index.html",
		},
	}

	for _, page := range pages {
		tdata := createTemplateData(config, data, page.Title, page.Description, createCanonicalURL(page.Canonical), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS)
		if err := utils.ExecuteTemplate(page.Template, filepath.Join(config.OutputDir, page.OutFile), tdata); err != nil {
			return fmt.Errorf("rendering template %s: %w", page.Template, err)
		}
		*sitemapUrls = append(*sitemapUrls, tdata.Canonical)
	}

	return nil
}

func renderCityPages(data *Data, config Config, cssFiles, otherJS []string, umamiJS string, sitemapUrls *[]string) error {
	for _, city := range data.Cities {
		tdata := createTemplateDataWithEntities(config, data, fmt.Sprintf("Social Run Clubs in %s", city.Name), city.MetaDescription(), createCanonicalURL(city.Slug()), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS, city, nil, nil, nil)
		fileName := filepath.Join(config.OutputDir, city.Slug(), "index.html")
		if err := utils.ExecuteTemplate("city.html", fileName, tdata); err != nil {
			return fmt.Errorf("rendering city template %q: %w", city.Name, err)
		}
		*sitemapUrls = append(*sitemapUrls, tdata.Canonical)

		for _, club := range city.Clubs {
			tdata := createTemplateDataWithEntities(config, data, fmt.Sprintf("%s - ein Social Run Club in %s", club.Name, city.Name), club.MetaDescription(), createCanonicalURL(club.Slug()), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS, city, club, nil, nil)
			fileName := filepath.Join(config.OutputDir, club.Slug(), "index.html")
			if err := utils.ExecuteTemplate("club.html", fileName, tdata); err != nil {
				return fmt.Errorf("rendering club template %q: %w", club.Name, err)
			}

			imgName := filepath.Join(config.OutputDir, club.Image())
			if err := copyClubImage(config, club, imgName); err != nil {
				return fmt.Errorf("copying club image for club %q: %w", club.Name, err)
			}

			*sitemapUrls = append(*sitemapUrls, tdata.Canonical)
		}
	}
	return nil
}

func renderTagPages(data *Data, config Config, cssFiles, otherJS []string, umamiJS string, sitemapUrls *[]string) error {
	for _, tag := range data.Tags {
		tdata := createTemplateDataWithEntities(config, data, fmt.Sprintf("Social Run Clubs in der Kategorie %s", tag.Name), fmt.Sprintf("Eine Übersicht über alle Social Run Clubs in der Kategorie %s.", tag.Name), createCanonicalURL(tag.Slug()), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS, nil, nil, tag, nil)
		fileName := filepath.Join(config.OutputDir, tag.Slug(), "index.html")
		if err := utils.ExecuteTemplate("tag.html", fileName, tdata); err != nil {
			return fmt.Errorf("rendering tag template %q: %w", tag.Name, err)
		}
		*sitemapUrls = append(*sitemapUrls, tdata.Canonical)
	}
	return nil
}

func renderPostPages(data *Data, config Config, cssFiles, otherJS []string, umamiJS string, sitemapUrls *[]string) error {
	for _, post := range data.Posts {
		tdata := createTemplateDataWithEntities(config, data, post.Title, fmt.Sprintf("Artikel: %s", post.Title), createCanonicalURL(post.Slug), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS, nil, nil, nil, post)
		fileName := filepath.Join(config.OutputDir, post.Slug, "index.html")
		if err := utils.ExecuteTemplate(post.TemplateFile, fileName, tdata); err != nil {
			return fmt.Errorf("rendering post template %q: %w", post.Title, err)
		}
		*sitemapUrls = append(*sitemapUrls, tdata.Canonical)
	}
	return nil
}

func renderSpecialPages(data *Data, config Config, cssFiles, otherJS []string, umamiJS string) error {
	// render 404 page
	tdata := createTemplateData(config, data, "404 - Seite nicht gefunden", "Die von dir angeforderte Seite konnte nicht gefunden werden.", createCanonicalURL("/404.html"), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS)
	if err := utils.ExecuteTemplate("404.html", filepath.Join(config.OutputDir, "404.html"), tdata); err != nil {
		return fmt.Errorf("rendering 404 template: %w", err)
	}

	// image grid
	tdata = createTemplateData(config, data, "Club Image Grid", "Club Image Grid", createCanonicalURL("/grid.html"), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS)
	if err := utils.ExecuteTemplate("grid.html", filepath.Join(config.OutputDir, "grid.html"), tdata); err != nil {
		return fmt.Errorf("rendering template %s: %w", "grid.html", err)
	}

	return nil
}

func createSiteFiles(data *Data, config Config, sitemapUrls []string) error {
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

func copyPlaceholderImage(config Config, targetPath string) error {
	placeholderImage := "static/placeholder.jpg"
	if err := utils.CopyFile(placeholderImage, targetPath); err != nil {
		return fmt.Errorf("copying placeholder image: %w", err)
	}
	return nil
}

func copyClubImage(config Config, club *Club, targetPath string) error {
	instagramProfile := club.InstagramProfile()
	if instagramProfile != "" {
		cachedImagName := filepath.Join(config.ImageDir, instagramProfile+".jpg")
		if utils.FileExists(cachedImagName) {
			if err := utils.CopyFile(cachedImagName, targetPath); err != nil {
				return fmt.Errorf("copying Instagram image for club %q: %w", club.Name, err)
			}
			return nil
		}
	}

	cachedImagName := filepath.Join(config.ImageDir, club.City.SanitizeName(), club.SanitizeName()+".jpg")
	if utils.FileExists(cachedImagName) {
		if err := utils.CopyFile(cachedImagName, targetPath); err != nil {
			return fmt.Errorf("copying direct image for club %q: %w", club.Name, err)
		}
		return nil
	}

	return copyPlaceholderImage(config, targetPath)
}

func Render(data *Data, cssFiles, jsFiles []string, config Config) error {
	umamiJS, otherJS := processJSFiles(jsFiles)

	// create indexnow.txt file
	if config.AHrefs.IndexNow != "" {
		filename := filepath.Join(config.OutputDir, config.AHrefs.IndexNow+".txt")
		if err := os.WriteFile(filename, []byte(config.AHrefs.IndexNow), 0644); err != nil {
			return fmt.Errorf("writing indexnow file %q: %w", filename, err)
		}
	}

	// collect all canonical URLs for creating a sitemap
	sitemapUrls := make([]string, 0)

	if err := renderStaticPages(data, config, cssFiles, otherJS, umamiJS, &sitemapUrls); err != nil {
		return err
	}

	if err := renderCityPages(data, config, cssFiles, otherJS, umamiJS, &sitemapUrls); err != nil {
		return err
	}

	if err := renderTagPages(data, config, cssFiles, otherJS, umamiJS, &sitemapUrls); err != nil {
		return err
	}

	if err := renderPostPages(data, config, cssFiles, otherJS, umamiJS, &sitemapUrls); err != nil {
		return err
	}

	if err := renderSpecialPages(data, config, cssFiles, otherJS, umamiJS); err != nil {
		return err
	}

	if err := createSiteFiles(data, config, sitemapUrls); err != nil {
		return err
	}

	return nil
}
