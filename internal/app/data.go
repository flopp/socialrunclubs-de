package app

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	googlesheetswrapper "github.com/flopp/go-googlesheetswrapper"
	"github.com/flopp/socialrunclubs-de/internal/utils"
)

type City struct {
	Name                 string
	Clubs                []*Club
	LatLon               *utils.LatLon
	NearestCities        []*City
	NearestCitiesNoClub  []*City
	SizeIndexWithoutClub int
}

func (c *City) MetaDescription() string {
	maxLength := 160
	desc := fmt.Sprintf("Eine Übersicht über alle Social Run Clubs in %s. ", c.Name)
	if len(c.Clubs) == 0 {
		desc += "Aktuell gibt es leider keine Einträge für diese Stadt. Du kannst aber gerne einen neuen Club hinzufügen!"
	} else if len(c.Clubs) == 1 {
		desc += fmt.Sprintf("Aktuell gibt es einen Eintrag: %s", c.Clubs[0].Name)
	} else {
		clubNames := make([]string, 0, len(c.Clubs))
		for _, club := range c.Clubs {
			clubNames = append(clubNames, club.Name)
		}
		desc += fmt.Sprintf("Aktuell gibt es %d Einträge:", len(c.Clubs))
		for i, name := range clubNames {
			if len(desc)+len(name)+2 >= maxLength {
				desc += "…"
				break
			}
			if i == 0 {
				desc += " " + name
			} else {
				desc += ", " + name
			}
		}
	}
	return desc
}

func (c *City) Show() bool {
	// show 10 biggest cities without clubs & cities with run clubs
	return c.SizeIndexWithoutClub <= 10 || len(c.Clubs) > 0
}

func (c *City) NumberOfClubs() int {
	return len(c.Clubs)
}

func (c *City) SanitizeName() string {
	return utils.SanitizeName(c.Name)
}

func (c *City) Slug() string {
	return fmt.Sprintf("/%s", c.SanitizeName())
}

func (c *City) Search() string {
	return strings.ToLower(c.Name)
}

type Tag struct {
	RawName     string
	Name        string
	Description *template.HTML
	Clubs       []*Club
}

func (t *Tag) Slug() string {
	return fmt.Sprintf("/tag/%s", utils.SanitizeName(t.RawName))
}

type Club struct {
	Name           string
	DescriptionRaw string
	Description    *template.HTML
	Tags           []*Tag
	City           *City
	LatLon         *utils.LatLon
	Instagram      string
	StravaClub     string
	Whatsapp       string
	Signal         string
	Website        string
	AddedRaw       string
	UpdatedRaw     string
	StatusRaw      string
}

func (c *Club) MetaDescription() string {
	maxLength := 160
	desc := fmt.Sprintf("Informationen und Links zum Social Run Club '%s' in %s", c.Name, c.City.Name)
	if c.Description != nil {
		desc += " - " + strings.ReplaceAll(string(*c.Description), "<br>", "; ")
		if len(desc) > maxLength {
			desc = desc[:maxLength-2] + "…"
		}
	}
	return desc
}

func (c *Club) SanitizeName() string {
	return utils.SanitizeName(c.Name)
}

var reInstagramProfile = regexp.MustCompile(`https?://(www\.)?instagram\.com/([^/?]+)/*`)

func (c *Club) InstagramProfile() string {
	if c.Instagram == "" {
		return ""
	}
	// Extract Instagram profile name from URL
	matches := reInstagramProfile.FindStringSubmatch(c.Instagram)
	if len(matches) > 1 {
		return matches[2]
	}
	return ""
}

func (c *Club) Slug() string {
	return fmt.Sprintf("/%s/%s", c.City.SanitizeName(), c.SanitizeName())
}

func (c *Club) Image() string {
	return fmt.Sprintf("/%s/%s/img.jpg", c.City.SanitizeName(), c.SanitizeName())
}

func (c *Club) Search() string {
	return strings.ToLower(fmt.Sprintf("%s %s", c.Name, c.City.Name))
}

type Post struct {
	Title        string
	Slug         string
	TemplateFile string
}

type Data struct {
	Now         time.Time
	NowStr      string
	Cities      []*City
	CityMap     map[string]*City
	Tags        []*Tag
	TagMap      map[string]*Tag
	Clubs       []*Club
	LatestClubs []*Club
	TopCities   []*City
	NumberClubs int
	Posts       []*Post
	Redirects   map[string]string
}

func (d *Data) RandomizedClubs() []*Club {
	// Filter out clubs based on name patterns (limit to 1 per pattern)
	excludedPatterns := map[string]int{
		"parkrun": 1,
		"kraft":   1,
	}

	clubs := make([]*Club, 0, len(d.Clubs))
	patternCounts := make(map[string]int)

	for _, club := range d.Clubs {
		clubNameLower := strings.ToLower(club.Name)
		shouldExclude := false

		for pattern, maxCount := range excludedPatterns {
			if strings.Contains(clubNameLower, pattern) {
				if patternCounts[pattern] >= maxCount {
					shouldExclude = true
					break
				}
				patternCounts[pattern]++
			}
		}

		if !shouldExclude {
			clubs = append(clubs, club)
		}
	}

	// Shuffle the clubs slice
	rand.Shuffle(len(clubs), func(i, j int) {
		clubs[i], clubs[j] = clubs[j], clubs[i]
	})

	return clubs
}

func (d *Data) getOrAddTag(name string) *Tag {
	if d.TagMap == nil {
		d.TagMap = make(map[string]*Tag)
	}
	if tag, found := d.TagMap[name]; found {
		return tag
	}
	tag := &Tag{
		RawName: name,
		Name:    name,
	}
	d.Tags = append(d.Tags, tag)
	d.TagMap[name] = tag
	return tag
}

func (d *Data) redirect(from string, to string) {
	if d.Redirects == nil {
		d.Redirects = make(map[string]string)
	}
	d.Redirects[from] = to
}

func getVal(colName string, row []string, colIdx map[string]int) (string, error) {
	col, ok := colIdx[colName]
	if !ok {
		return "", fmt.Errorf("unknown column: %s", colName)
	}
	if col >= len(row) {
		return "", nil
	}
	return strings.TrimSpace(row[col]), nil
}

func sortCitiesAndClubs(cities []*City) {
	sort.Slice(cities, func(i, j int) bool {
		return cities[i].Slug() < cities[j].Slug()
	})
	for _, city := range cities {
		sort.Slice(city.Clubs, func(i, j int) bool {
			return city.Clubs[i].Slug() < city.Clubs[j].Slug()
		})
	}
}

func sortTagsAndClubs(tags []*Tag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Slug() < tags[j].Slug()
	})
	for _, tag := range tags {
		sort.Slice(tag.Clubs, func(i, j int) bool {
			// sort by name + city name
			n1 := tag.Clubs[i].SanitizeName()
			n2 := tag.Clubs[j].SanitizeName()
			if n1 != n2 {
				return n1 < n2
			}
			// fallback to city
			return tag.Clubs[i].City.SanitizeName() < tag.Clubs[j].City.SanitizeName()
		})
	}
}

func sortClubs(clubs []*Club) {
	sort.Slice(clubs, func(i, j int) bool {
		// sort by name + city name
		n1 := clubs[i].SanitizeName()
		n2 := clubs[j].SanitizeName()
		if n1 != n2 {
			return n1 < n2
		}
		// fallback to city
		return clubs[i].City.SanitizeName() < clubs[j].City.SanitizeName()
	})
}

func checkForDuplicateClubs(clubs []*Club) {
	seen := make(map[string]*Club)
	for _, club := range clubs {
		key := club.Slug()
		if _, exists := seen[key]; exists {
			log.Printf("duplicate club found with slug: %s", key)
		} else {
			seen[key] = club
		}
	}
}

func processClubsSheet(sheetName string, rows [][]string, data *Data) error {
	if len(rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}

	required := []string{"ID", "ADDED", "UPDATED", "STATUS", "REDIRECT NAME", "REDIRECT CITY", "NAME", "OLD NAME", "CITY", "COORDS", "DESCRIPTION", "TAGS", "INSTAGRAM_URL", "STRAVA_URL", "WHATSAPP_URL", "WEBSITE_URL"}
	colIdx, err := googlesheetswrapper.ExtractHeader(rows[:1], required, false)
	if err != nil {
		return err
	}

	hasCities := len(data.Cities) > 0
	for index, row := range rows[1:] {
		club := &Club{}

		// Define field mappings for direct assignment
		type fieldMapping struct {
			field *string // pointer to string field
			col   string
		}

		cityRaw := ""
		latLonRaw := ""
		tagsRaw := ""

		mappings := []fieldMapping{
			{&club.Name, "NAME"},
			{&club.DescriptionRaw, "DESCRIPTION"},
			{&cityRaw, "CITY"},
			{&latLonRaw, "COORDS"},
			{&club.Instagram, "INSTAGRAM_URL"},
			{&club.StravaClub, "STRAVA_URL"},
			{&club.Whatsapp, "WHATSAPP_URL"},
			{&club.Website, "WEBSITE_URL"},
			{&club.AddedRaw, "ADDED"},
			{&club.UpdatedRaw, "UPDATED"},
			{&club.StatusRaw, "STATUS"},
			{&tagsRaw, "TAGS"},
		}

		// Process direct field assignments
		for _, mapping := range mappings {
			if val, err := getVal(mapping.col, row, colIdx); err != nil {
				return fmt.Errorf("row %d: %v", index+2, err)
			} else {
				*mapping.field = val
			}
		}

		// skip invalid clubs
		if club.Name == "" {
			log.Printf("CLUBS row %d: empty club name: %q", index+2, club.Name)
			continue
		}
		if cityRaw == "" {
			log.Printf("CLUBS row %d: empty city name: %q", index+2, cityRaw)
			continue
		}

		if club.StatusRaw == "obsolete" || club.StatusRaw == "duplicate" {
			// redirects
			redirectName := ""
			redirectCity := ""
			if redirectName, err = getVal("REDIRECT NAME", row, colIdx); err != nil {
				return fmt.Errorf("row %d: %v", index+2, err)
			}
			if redirectCity, err = getVal("REDIRECT CITY", row, colIdx); err != nil {
				return fmt.Errorf("row %d: %v", index+2, err)
			}
			if redirectName != "" && redirectCity != "" {
				// redirect from "old city/old club" to "new city/new club"
				to := fmt.Sprintf("/%s/%s", utils.SanitizeName(redirectCity), utils.SanitizeName(redirectName))
				data.redirect(fmt.Sprintf("/%s/%s", utils.SanitizeName(cityRaw), utils.SanitizeName(club.Name)), to)
				data.redirect(fmt.Sprintf("/%s/%s/", utils.SanitizeName(cityRaw), utils.SanitizeName(club.Name)), to)
				data.redirect(fmt.Sprintf("/%s/%s/index.html", utils.SanitizeName(cityRaw), utils.SanitizeName(club.Name)), to)
			} else if club.StatusRaw == "obsolete" && redirectName == "" && redirectCity == "" {
				// redirect from "city/obsolete club" to "city"
				to := fmt.Sprintf("/%s", utils.SanitizeName(cityRaw))
				data.redirect(fmt.Sprintf("/%s/%s", utils.SanitizeName(cityRaw), utils.SanitizeName(club.Name)), to)
				data.redirect(fmt.Sprintf("/%s/%s/", utils.SanitizeName(cityRaw), utils.SanitizeName(club.Name)), to)
				data.redirect(fmt.Sprintf("/%s/%s/index.html", utils.SanitizeName(cityRaw), utils.SanitizeName(club.Name)), to)
			} else {
				log.Printf("CLUBS row %d: invalid redirect for obsolete/duplicate club: %q / %q", index+2, redirectCity, redirectName)
			}

			continue
		}

		// old name redirect
		oldName := ""
		if oldName, err = getVal("OLD NAME", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if oldName != "" {
			from := fmt.Sprintf("/%s/%s", utils.SanitizeName(cityRaw), utils.SanitizeName(oldName))
			to := fmt.Sprintf("/%s/%s", utils.SanitizeName(cityRaw), utils.SanitizeName(club.Name))
			data.redirect(from, to)
		}

		// process data
		descriptionHtml := template.HTML(club.DescriptionRaw)
		club.Description = &descriptionHtml

		if latLonRaw != "" {
			latlon, err := utils.ParseLatLon(latLonRaw)
			if err != nil {
				log.Printf("CLUBS row %d: invalid coords: %q", index+2, latLonRaw)
				continue
			}
			club.LatLon = &latlon
		}

		if strings.Contains(club.Whatsapp, "signal") {
			club.Signal = club.Whatsapp
			club.Whatsapp = ""
		}

		if club.UpdatedRaw == club.AddedRaw {
			club.UpdatedRaw = ""
		}

		if city, found := data.CityMap[cityRaw]; found {
			city.Clubs = append(city.Clubs, club)
			club.City = city
		} else {
			if hasCities {
				log.Printf("CLUBS row %d: unknown city: %q", index+2, cityRaw)
			}
			city = &City{
				Name:                 cityRaw,
				Clubs:                []*Club{club},
				SizeIndexWithoutClub: 0,
			}
			data.Cities = append(data.Cities, city)
			data.CityMap[city.Name] = city
			club.City = city
		}

		// process tags
		club.Tags = make([]*Tag, 0)
		for _, tagName := range utils.SplitAndTrim(tagsRaw, ",") {
			tag := data.getOrAddTag(tagName)
			tag.Clubs = append(tag.Clubs, club)
			club.Tags = append(club.Tags, tag)
		}
	}

	return nil
}

func processCitiesSheet(sheetName string, rows [][]string, data *Data) error {
	if len(rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}

	required := []string{"NAME"}
	colIdx, err := googlesheetswrapper.ExtractHeader(rows[:1], required, false)
	if err != nil {
		return err
	}

	cities := make(map[string]struct{})
	cityList := make([]string, 0)

	for index, row := range rows[1:] {
		name := ""

		if name, err = getVal("NAME", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if name == "" {
			continue
		}

		if _, found := cities[name]; !found {
			cities[name] = struct{}{}
			cityList = append(cityList, name)
		} else {
			log.Printf("CITIES row %d: duplicate city name: %q", index+2, name)
		}
	}

	// if there are already city objects (from clubs), check the are all in the cities list
	for _, city := range data.Cities {
		if _, found := cities[city.Name]; !found {
			log.Printf("CITIES: missing city from sheet: %q", city.Name)
		}
	}

	// add cities to data

	indexWithoutClub := 1
	for _, name := range cityList {
		if _, found := data.CityMap[name]; !found {
			city := &City{
				Name:                 name,
				SizeIndexWithoutClub: indexWithoutClub,
			}
			data.Cities = append(data.Cities, city)
			data.CityMap[city.Name] = city
			indexWithoutClub++
		}
	}

	return nil
}

func processTagsSheet(sheetName string, rows [][]string, data *Data) error {
	if len(rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}

	required := []string{"NAME", "FANCY", "DESCRIPTION"}
	colIdx, err := googlesheetswrapper.ExtractHeader(rows[:1], required, false)
	if err != nil {
		return err
	}

	for index, row := range rows[1:] {
		name := ""

		if name, err = getVal("NAME", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if name == "" {
			continue
		}

		fancy := ""
		if fancy, err = getVal("FANCY", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if fancy == "" {
			fancy = name
		}

		tag := data.getOrAddTag(name)
		tag.Name = fancy
		descriptionRaw := ""
		if descriptionRaw, err = getVal("DESCRIPTION", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if len(descriptionRaw) > 0 {
			descriptionHtml := template.HTML(descriptionRaw)
			tag.Description = &descriptionHtml
		}
	}

	return nil
}

func collectPosts(data *Data) error {
	data.Posts = make([]*Post, 0)

	files, err := filepath.Glob("templates/posts/*.html")
	if err != nil {
		return err
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading post file %s: %v", file, err)
		}

		title := ""
		re := regexp.MustCompile(`<!--\s*TITLE\s+"([^"]+)"\s*-->`)
		matches := re.FindStringSubmatch(string(content))
		if len(matches) == 0 {
			return fmt.Errorf("post file %s: no TITLE found", file)
		}

		if len(matches) > 1 {
			title = matches[1]
		}

		filename := filepath.Base(file)
		slug := "/post/" + strings.TrimSuffix(filename, filepath.Ext(filename)) + "/"
		templateName := "posts/" + filename

		data.Posts = append(data.Posts, &Post{
			Title:        title,
			Slug:         slug,
			TemplateFile: templateName,
		})
	}

	return nil
}

func GetData(config Config) (*Data, error) {
	data := &Data{
		Now:         time.Now(),
		NowStr:      time.Now().Format("2006-01-02 15:04:05"),
		Cities:      make([]*City, 0),
		CityMap:     make(map[string]*City),
		NumberClubs: 0,
	}

	sheetData, err := utils.Retry(3, 8*time.Second, func() (map[string][][]string, error) {
		ctx := context.Background()
		client, err := googlesheetswrapper.New(config.Google.APIKey, config.Google.SheetId)
		if err != nil {
			return nil, fmt.Errorf("creating sheets client: %w", err)
		}
		all, err := client.ReadAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("reading all sheets: %w", err)
		}
		return all, nil
	})
	if err != nil {
		return nil, fmt.Errorf("getting sheets: %v", err)
	}

	// get the clubs and cities sheets
	var clubsFound, citiesFound, tagsFound bool

	// Define sheet processors
	type sheetProcessor struct {
		processFunc func(string, [][]string, *Data) error
		found       *bool
	}

	processors := map[string]*sheetProcessor{
		"CLUBS":  {processFunc: processClubsSheet, found: &clubsFound},
		"CITIES": {processFunc: processCitiesSheet, found: &citiesFound},
		"TAGS":   {processFunc: processTagsSheet, found: &tagsFound},
	}

	// Process sheets
	for name, rows := range sheetData {
		if processor, exists := processors[name]; exists {
			*processor.found = true
			if err := processor.processFunc(name, rows, data); err != nil {
				return nil, fmt.Errorf("processing %s sheet: %v", name, err)
			}
		} else if name == "SUBMIT" || name == "REPORT" || strings.Contains(name, "IGNORE") {
			// ignore these sheets
			continue
		} else {
			return nil, fmt.Errorf("unknown sheet name: %s", name)
		}
	}

	// Check required sheets
	requiredSheets := []string{"CLUBS", "CITIES", "TAGS"}
	for _, sheetName := range requiredSheets {
		if processor := processors[sheetName]; !*processor.found {
			return nil, fmt.Errorf("missing %s sheet", sheetName)
		}
	}

	// sorting of cities
	sortCitiesAndClubs(data.Cities)

	// sorting of tags
	sortTagsAndClubs(data.Tags)

	// all clubs
	data.Clubs = make([]*Club, 0)
	for _, city := range data.Cities {
		data.Clubs = append(data.Clubs, city.Clubs...)
	}
	sortClubs(data.Clubs)

	// check for duplicates (via slugs)
	checkForDuplicateClubs(data.Clubs)

	// collect clubs by added date
	var addedClubs []*Club
	for _, city := range data.Cities {
		for _, club := range city.Clubs {
			if club.AddedRaw != "" {
				addedClubs = append(addedClubs, club)
			}
			data.NumberClubs++
		}
	}

	// get selection of 6 latest clubs
	if len(addedClubs) > 0 {
		numberOfLatest := 6
		// sort by date, latest first
		sort.Slice(addedClubs, func(i, j int) bool {
			return addedClubs[i].AddedRaw > addedClubs[j].AddedRaw
		})

		// get all clubs with latest date
		latestDate := addedClubs[0].AddedRaw
		candidates := make([]*Club, 0)
		for i := 0; i < len(addedClubs); i++ {
			if addedClubs[i].AddedRaw == latestDate {
				candidates = append(candidates, addedClubs[i])
			}
		}
		if len(candidates) >= numberOfLatest {
			data.LatestClubs = candidates[:numberOfLatest]
		} else {
			// not enough -> just take latest numberOfLatest
			data.LatestClubs = addedClubs[:numberOfLatest]
		}
	}

	// get the 5 cities with the most clubs:
	var topCities []*City
	for _, city := range data.Cities {
		if len(city.Clubs) > 0 {
			topCities = append(topCities, city)
		}
	}
	sort.Slice(topCities, func(i, j int) bool {
		ic := len(topCities[i].Clubs)
		jc := len(topCities[j].Clubs)
		if ic != jc {
			return ic > jc
		}
		return topCities[i].Name < topCities[j].Name
	})
	if len(topCities) > 5 {
		data.TopCities = topCities[:5]
	} else {
		data.TopCities = topCities
	}

	if err := collectPosts(data); err != nil {
		return nil, fmt.Errorf("collecting posts: %v", err)
	}

	return data, nil
}

func AnnotateCityCoordinates(data *Data, geocoder *utils.CachingGeocoder) error {
	for _, city := range data.Cities {
		if city.LatLon != nil {
			continue
		}
		/*
			if city.Show() == false {
				continue
			}
		*/
		coords, err := geocoder.Lookup(city.Name)
		if err != nil {
			log.Panicf("getting coordinates for city %q: %v", city.Name, err)
		} else {
			city.LatLon = &coords
		}
	}

	return nil
}

func findNearestCities(city *City, cities []*City, maxResults int, withClubs bool) []*City {
	if city.LatLon == nil {
		return nil
	}

	type cityDistance struct {
		city     *City
		distance float64
	}
	var distances []cityDistance
	for _, other := range cities {
		if other == city {
			continue
		}
		if other.LatLon == nil {
			continue
		}
		if withClubs && len(other.Clubs) == 0 {
			continue
		}
		if !withClubs && len(other.Clubs) != 0 {
			continue
		}

		distance := utils.Distance(*city.LatLon, *other.LatLon)
		distances = append(distances, cityDistance{city: other, distance: distance})
	}
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].distance < distances[j].distance
	})
	if len(distances) > maxResults {
		distances = distances[:maxResults]
	}
	var result []*City
	for _, d := range distances {
		result = append(result, d.city)
	}
	return result
}

func AnnotateNearestCities(data *Data) error {
	for _, city := range data.Cities {
		city.NearestCities = findNearestCities(city, data.Cities, 3, true)
		city.NearestCitiesNoClub = findNearestCities(city, data.Cities, 3, false)
	}
	return nil
}
