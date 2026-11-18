package app

import (
	"encoding/xml"
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
	NoIndex        bool
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

func renderStaticPages(data *Data, config Config, cssFiles, otherJS []string, umamiJS string, sitemapUrls *[]SitemapEntry) error {
	pages := []struct {
		Title       string
		Description string
		Canonical   string
		Template    string
		OutFile     string
		LastMod     string
		NoIndex     bool
	}{
		{
			Title:       "Social Run Clubs in Deutschland",
			Description: "Eine Übersicht über alle Social Run Clubs in Deutschland.",
			Canonical:   "/",
			Template:    "index.html",
			OutFile:     "index.html",
			LastMod:     data.LastMod(),
		},
		{
			Title:       "Impressum - Social Run Clubs",
			Description: "Impressum von socialrunclubs.de.",
			Canonical:   "/impressum.html",
			Template:    "impressum.html",
			OutFile:     "impressum.html",
			NoIndex:     true,
		},
		{
			Title:       "Datenschutz - Social Run Clubs",
			Description: "Datenschutz von socialrunclubs.de.",
			Canonical:   "datenschutz.html",
			Template:    "datenschutz.html",
			OutFile:     "datenschutz.html",
			NoIndex:     true,
		},
		{
			Title:       "Deutsche Städte mit Social Run Clubs",
			Description: "Eine Übersicht über alle Städte mit Social Run Clubs.",
			Canonical:   "/cities.html",
			Template:    "cities.html",
			OutFile:     "cities.html",
			LastMod:     data.LastMod(),
		},
		{
			Title:       "Deutsche Städte ohne Social Run Clubs",
			Description: "Eine Übersicht über alle Städte ohne Social Run Clubs.",
			Canonical:   "/cities-no-club.html",
			Template:    "cities-no-club.html",
			OutFile:     "cities-no-club.html",
			NoIndex:     true,
		},
		{
			Title:       "Liste aller Social Run Clubs in Deutschland",
			Description: "Eine Übersicht über alle Social Run Clubs in Deutschland.",
			Canonical:   "/clubs.html",
			Template:    "clubs.html",
			OutFile:     "clubs.html",
			LastMod:     data.LastMod(),
		},
		{
			Title:       "Social Run Club Kategorien",
			Description: "Eine Übersicht über alle Social Run Club Kategorien.",
			Canonical:   "/tags.html",
			Template:    "tags.html",
			OutFile:     "tags.html",
			LastMod:     data.LastMod(),
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
		tdata.NoIndex = page.NoIndex
		if err := utils.ExecuteTemplate(page.Template, filepath.Join(config.OutputDir, page.OutFile), tdata); err != nil {
			return fmt.Errorf("rendering template %s: %w", page.Template, err)
		}
		if page.NoIndex {
			continue
		}
		*sitemapUrls = append(*sitemapUrls, SitemapEntry{URL: tdata.Canonical, LastMod: page.LastMod})
	}

	return nil
}

func renderCityPages(data *Data, config Config, cssFiles, otherJS []string, umamiJS string, sitemapUrls *[]SitemapEntry) error {
	for _, city := range data.Cities {
		tdata := createTemplateDataWithEntities(config, data, fmt.Sprintf("Run Clubs und Lauftreffs in %s", city.Name), city.MetaDescription(), createCanonicalURL(city.Slug()), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS, city, nil, nil, nil)
		fileName := filepath.Join(config.OutputDir, city.Slug(), "index.html")
		if err := utils.ExecuteTemplate("city.html", fileName, tdata); err != nil {
			return fmt.Errorf("rendering city template %q: %w", city.Name, err)
		}
		*sitemapUrls = append(*sitemapUrls, SitemapEntry{URL: tdata.Canonical, LastMod: city.LastMod()})

		for _, club := range city.Clubs {
			tdata := createTemplateDataWithEntities(config, data, fmt.Sprintf("%s - ein Run Club in %s", club.Name, city.Name), club.MetaDescription(), createCanonicalURL(club.Slug()), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS, city, club, nil, nil)
			fileName := filepath.Join(config.OutputDir, club.Slug(), "index.html")
			if err := utils.ExecuteTemplate("club.html", fileName, tdata); err != nil {
				return fmt.Errorf("rendering club template %q: %w", club.Name, err)
			}

			imgName := filepath.Join(config.OutputDir, club.Image())
			hasImage, err := copyClubImage(config, club, imgName)
			if err != nil {
				return fmt.Errorf("copying club image for club %q: %w", club.Name, err)
			}

			entry := SitemapEntry{URL: tdata.Canonical, LastMod: club.LastMod()}
			if hasImage {
				entry.Image = createCanonicalURL(club.Image())
			}
			*sitemapUrls = append(*sitemapUrls, entry)
		}
	}
	return nil
}

func renderTagPages(data *Data, config Config, cssFiles, otherJS []string, umamiJS string, sitemapUrls *[]SitemapEntry) error {
	for _, tag := range data.Tags {
		tdata := createTemplateDataWithEntities(config, data, fmt.Sprintf("Run Clubs und Lauftreffs in der Kategorie %s", tag.Name), fmt.Sprintf("Eine Übersicht über alle Run Clubs und Lauftreffs in der Kategorie %s.", tag.Name), createCanonicalURL(tag.Slug()), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS, nil, nil, tag, nil)
		fileName := filepath.Join(config.OutputDir, tag.Slug(), "index.html")
		if err := utils.ExecuteTemplate("tag.html", fileName, tdata); err != nil {
			return fmt.Errorf("rendering tag template %q: %w", tag.Name, err)
		}
		*sitemapUrls = append(*sitemapUrls, SitemapEntry{URL: tdata.Canonical, LastMod: tag.LastMod()})
	}
	return nil
}

func renderPostPages(data *Data, config Config, cssFiles, otherJS []string, umamiJS string, sitemapUrls *[]SitemapEntry) error {
	for _, post := range data.Posts {
		tdata := createTemplateDataWithEntities(config, data, post.Title, fmt.Sprintf("Artikel: %s", post.Title), createCanonicalURL(post.Slug), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS, nil, nil, nil, post)
		fileName := filepath.Join(config.OutputDir, post.Slug, "index.html")
		if err := utils.ExecuteTemplate(post.TemplateFile, fileName, tdata); err != nil {
			return fmt.Errorf("rendering post template %q: %w", post.Title, err)
		}
		*sitemapUrls = append(*sitemapUrls, SitemapEntry{URL: tdata.Canonical})
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

	// private helper page for formatting new club submissions
	tdata = createTemplateData(config, data, "Run Club eintragen - Social Run Clubs", "Interne Hilfe zum Formatieren neuer Run Club Einträge.", createCanonicalURL("/submit.html"), config.Google.SubmitUrl, config.Google.ReportUrl, cssFiles, otherJS, umamiJS)
	tdata.NoIndex = true
	if err := utils.ExecuteTemplate("submit.html", filepath.Join(config.OutputDir, "submit.html"), tdata); err != nil {
		return fmt.Errorf("rendering template %s: %w", "submit.html", err)
	}

	return nil
}

type SitemapEntry struct {
	URL     string
	LastMod string
	Image   string // optional image URL
}

type sitemapImage struct {
	Loc string `xml:"image:loc"`
}

type sitemapURL struct {
	Loc     string         `xml:"loc"`
	LastMod string         `xml:"lastmod,omitempty"`
	Images  []sitemapImage `xml:"image:image,omitempty"`
}

type sitemapURLSet struct {
	XMLName    xml.Name     `xml:"urlset"`
	Xmlns      string       `xml:"xmlns,attr"`
	XmlnsImage string       `xml:"xmlns:image,attr"`
	URLs       []sitemapURL `xml:"url"`
}

func buildSitemap(entries []SitemapEntry) ([]byte, error) {
	set := sitemapURLSet{
		Xmlns:      "http://www.sitemaps.org/schemas/sitemap/0.9",
		XmlnsImage: "http://www.google.com/schemas/sitemap-image/1.1",
	}
	for _, e := range entries {
		u := sitemapURL{Loc: e.URL, LastMod: e.LastMod}
		if e.Image != "" {
			u.Images = []sitemapImage{{Loc: e.Image}}
		}
		set.URLs = append(set.URLs, u)
	}
	body, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), append(body, '\n')...), nil
}

func createSiteFiles(data *Data, config Config, sitemapUrls []SitemapEntry) error {
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

	// create robots.txt
	robotsFile := filepath.Join(config.OutputDir, "robots.txt")
	robotsData := []byte("User-agent: *\nAllow: /\n\nSitemap: https://socialrunclubs.de/sitemap.xml\n")
	if err := os.WriteFile(robotsFile, robotsData, 0644); err != nil {
		return fmt.Errorf("writing robots.txt: %w", err)
	}

	// create sitemap.xml
	sitemapFile := filepath.Join(config.OutputDir, "sitemap.xml")
	sitemapData, err := buildSitemap(sitemapUrls)
	if err != nil {
		return fmt.Errorf("building sitemap: %w", err)
	}
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

// copyClubImage copies the club's image (or the placeholder) and reports whether a real image was used.
func copyClubImage(config Config, club *Club, targetPath string) (bool, error) {
	instagramProfile := club.InstagramProfile()
	if instagramProfile != "" {
		cachedImagName := filepath.Join(config.ImageDir, instagramProfile+".jpg")
		if utils.FileExists(cachedImagName) {
			if err := utils.CopyFile(cachedImagName, targetPath); err != nil {
				return false, fmt.Errorf("copying Instagram image for club %q: %w", club.Name, err)
			}
			return true, nil
		}
	}

	cachedImagName := filepath.Join(config.ImageDir, club.City.SanitizeName(), club.SanitizeName()+".jpg")
	if utils.FileExists(cachedImagName) {
		if err := utils.CopyFile(cachedImagName, targetPath); err != nil {
			return false, fmt.Errorf("copying direct image for club %q: %w", club.Name, err)
		}
		return true, nil
	}

	return false, copyPlaceholderImage(config, targetPath)
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
	sitemapUrls := make([]SitemapEntry, 0)

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
