package app

import (
	"fmt"
	"html/template"
	"log"
	"math/rand/v2"
	"sort"
	"strings"
	"time"

	"github.com/flopp/socialrunclubs-de/internal/utils"
)

type City struct {
	Name                 string
	Clubs                []*Club
	Coords               *utils.LatLon
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
	tagsRaw        []string
	Tags           []*Tag
	CityRaw        string
	City           *City
	LatLonRaw      string
	LatLon         *utils.LatLon
	Instagram      string
	StravaClub     string
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

func (c *Club) Slug() string {
	return fmt.Sprintf("/%s/%s", c.City.SanitizeName(), c.SanitizeName())
}

func (c *Club) Search() string {
	return strings.ToLower(fmt.Sprintf("%s %s", c.Name, c.City.Name))
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
	Redirects   map[string]string
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

func processClubsSheet(sheet utils.Sheet, data *Data) error {
	if len(sheet.Rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}

	required := []string{"ID", "ADDED", "UPDATED", "STATUS", "REDIRECT NAME", "REDIRECT CITY", "NAME", "OLD NAME", "CITY", "COORDS", "DESCRIPTION", "TAGS", "INSTAGRAM_URL", "STRAVA_URL", "WEBSITE_URL"}
	colIdx, err := utils.ValidateColumns(sheet.Rows[0], required)
	if err != nil {
		return err
	}

	hasCities := len(data.Cities) > 0
	for index, row := range sheet.Rows[1:] {
		club := &Club{}

		if club.Name, err = getVal("NAME", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if club.DescriptionRaw, err = getVal("DESCRIPTION", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		tagsRaw := ""
		if tagsRaw, err = getVal("TAGS", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		club.tagsRaw = utils.SplitAndTrim(tagsRaw, ",")
		if club.DescriptionRaw, err = getVal("DESCRIPTION", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if club.CityRaw, err = getVal("CITY", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if club.LatLonRaw, err = getVal("COORDS", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if club.Instagram, err = getVal("INSTAGRAM_URL", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if club.StravaClub, err = getVal("STRAVA_URL", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if club.Website, err = getVal("WEBSITE_URL", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if club.AddedRaw, err = getVal("ADDED", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if club.UpdatedRaw, err = getVal("UPDATED", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if club.StatusRaw, err = getVal("STATUS", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}

		// skip invalid clubs
		if club.Name == "" {
			log.Printf("CLUBS row %d: empty club name: %q", index+2, club.Name)
			continue
		}
		if club.CityRaw == "" {
			log.Printf("CLUBS row %d: empty city name: %q", index+2, club.CityRaw)
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
				from := fmt.Sprintf("/%s/%s", utils.SanitizeName(club.CityRaw), utils.SanitizeName(club.Name))
				to := fmt.Sprintf("/%s/%s", utils.SanitizeName(redirectCity), utils.SanitizeName(redirectName))
				data.redirect(from, to)
			}
			continue
		}

		// old name redirect
		oldName := ""
		if oldName, err = getVal("OLD NAME", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if oldName != "" {
			from := fmt.Sprintf("/%s/%s", utils.SanitizeName(club.CityRaw), utils.SanitizeName(oldName))
			to := fmt.Sprintf("/%s/%s", utils.SanitizeName(club.CityRaw), utils.SanitizeName(club.Name))
			data.redirect(from, to)
		}

		// process data
		descriptionHtml := template.HTML(club.DescriptionRaw)
		club.Description = &descriptionHtml

		if club.LatLonRaw != "" {
			latlon, err := utils.ParseLatLon(club.LatLonRaw)
			if err != nil {
				log.Printf("CLUBS row %d: invalid coords: %q", index+2, club.LatLonRaw)
				continue
			}
			club.LatLon = &latlon
		}

		if club.UpdatedRaw == club.AddedRaw {
			club.UpdatedRaw = ""
		}

		if city, found := data.CityMap[club.CityRaw]; found {
			city.Clubs = append(city.Clubs, club)
			club.City = city
		} else {
			if hasCities {
				log.Printf("CLUBS row %d: unknown city: %q", index+2, club.CityRaw)
			}
			city = &City{
				Name:                 club.CityRaw,
				Clubs:                []*Club{club},
				SizeIndexWithoutClub: 0,
			}
			data.Cities = append(data.Cities, city)
			data.CityMap[city.Name] = city
			club.City = city
		}

		// process tags
		club.Tags = make([]*Tag, 0)
		for _, tagName := range club.tagsRaw {
			tag := data.getOrAddTag(tagName)
			tag.Clubs = append(tag.Clubs, club)
			club.Tags = append(club.Tags, tag)
		}
	}

	return nil
}

func processCitiesSheet(sheet utils.Sheet, data *Data) error {
	if len(sheet.Rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}

	required := []string{"NAME"}
	colIdx, err := utils.ValidateColumns(sheet.Rows[0], required)
	if err != nil {
		return err
	}

	cities := make(map[string]struct{})
	cityList := make([]string, 0)

	for index, row := range sheet.Rows[1:] {
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

func processTagsSheet(sheet utils.Sheet, data *Data) error {
	if len(sheet.Rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}

	required := []string{"NAME", "FANCY", "DESCRIPTION"}
	colIdx, err := utils.ValidateColumns(sheet.Rows[0], required)
	if err != nil {
		return err
	}

	for index, row := range sheet.Rows[1:] {
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

func GetData(config Config) (*Data, error) {
	data := &Data{
		Now:         time.Now(),
		NowStr:      time.Now().Format("2006-01-02 15:04:05"),
		Cities:      make([]*City, 0),
		CityMap:     make(map[string]*City),
		NumberClubs: 0,
	}

	sheets, err := utils.Retry(3, 8*time.Second, func() ([]utils.Sheet, error) {
		return utils.GetSheets(config.Google.APIKey, config.Google.SheetId)
	})
	if err != nil {
		return nil, fmt.Errorf("getting sheets: %v", err)
	}

	// get the clubs and cities sheets
	clubsFound := false
	citiesFound := false
	tagsFound := false
	for _, sheet := range sheets {
		switch sheet.Name {
		case "CLUBS":
			clubsFound = true
			if err := processClubsSheet(sheet, data); err != nil {
				return nil, fmt.Errorf("processing clubs sheet: %v", err)
			}
		case "CITIES":
			citiesFound = true
			if err := processCitiesSheet(sheet, data); err != nil {
				return nil, fmt.Errorf("processing cities sheet: %v", err)
			}
		case "TAGS":
			tagsFound = true
			if err := processTagsSheet(sheet, data); err != nil {
				return nil, fmt.Errorf("processing tags sheet: %v", err)
			}
		case "SUBMIT":
			// ignore
		case "REPORT":
			// ignore
		default:
			if strings.Contains(sheet.Name, "IGNORE") {
				continue
			}
			return nil, fmt.Errorf("unknown sheet name: %s", sheet.Name)
		}
	}

	if !clubsFound {
		return nil, fmt.Errorf("missing clubs sheet")
	}
	if !citiesFound {
		return nil, fmt.Errorf("missing cities sheet")
	}
	if !tagsFound {
		return nil, fmt.Errorf("missing tags sheet")
	}

	// sorting of cities
	sort.Slice(data.Cities, func(i, j int) bool {
		return data.Cities[i].Slug() < data.Cities[j].Slug()
	})
	for _, city := range data.Cities {
		sort.Slice(city.Clubs, func(i, j int) bool {
			return city.Clubs[i].Slug() < city.Clubs[j].Slug()
		})
	}

	// sorting of tags
	sort.Slice(data.Tags, func(i, j int) bool {
		return data.Tags[i].Slug() < data.Tags[j].Slug()
	})
	for _, tag := range data.Tags {
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

	// all clubs
	data.Clubs = make([]*Club, 0)
	for _, city := range data.Cities {
		data.Clubs = append(data.Clubs, city.Clubs...)
	}
	sort.Slice(data.Clubs, func(i, j int) bool {
		// sort by name + city name
		n1 := data.Clubs[i].SanitizeName()
		n2 := data.Clubs[j].SanitizeName()
		if n1 != n2 {
			return n1 < n2
		}
		// fallback to city
		return data.Clubs[i].City.SanitizeName() < data.Clubs[j].City.SanitizeName()
	})

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

	// get selection of 5 latest clubs
	if len(addedClubs) > 0 {
		// sort by date, latest first
		sort.Slice(addedClubs, func(i, j int) bool {
			return addedClubs[i].AddedRaw > addedClubs[j].AddedRaw
		})

		// get áll clubs with latest date
		latestDate := addedClubs[0].AddedRaw
		candidates := make([]*Club, 0)
		for i := 0; i < len(addedClubs); i++ {
			if addedClubs[i].AddedRaw == latestDate {
				candidates = append(candidates, addedClubs[i])
			}
		}
		// not enough -> just take latest 5
		if len(candidates) < 5 {
			candidates = addedClubs[:5]
		}

		// get random selection of 5 clubs from candidates
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
		data.LatestClubs = candidates[:5]
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

	return data, nil
}

func AnnotateCityCoordinates(data *Data, geocoder *utils.CachingGeocoder) error {
	for _, city := range data.Cities {
		if city.Coords != nil {
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
			city.Coords = &coords
		}
	}
	return nil
}

func findNearestCities(city *City, cities []*City, maxResults int, withClubs bool) []*City {
	if city.Coords == nil {
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
		if other.Coords == nil {
			continue
		}
		if withClubs && len(other.Clubs) == 0 {
			continue
		}
		if !withClubs && len(other.Clubs) != 0 {
			continue
		}

		distance := utils.Distance(*city.Coords, *other.Coords)
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
