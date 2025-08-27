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
	SizeIndexWithoutClub int
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

type Club struct {
	Name           string
	DescriptionRaw string
	Description    *template.HTML
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

func (c *Club) SanitizeName() string {
	return utils.SanitizeName(c.Name)
}

func (c *Club) Slug() string {
	return fmt.Sprintf("/%s/%s", c.City.SanitizeName(), c.SanitizeName())
}

type Data struct {
	Now         time.Time
	NowStr      string
	Cities      []*City
	CityMap     map[string]*City
	Clubs       []*Club
	LatestClubs []*Club
	TopCities   []*City
	NumberClubs int
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

	required := []string{"ID", "ADDED", "UPDATED", "STATUS", "NAME", "CITY", "COORDS", "DESCRIPTION", "INSTAGRAM_URL", "STRAVA_URL", "WEBSITE_URL"}
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

	// sorting
	sort.Slice(data.Cities, func(i, j int) bool {
		return data.Cities[i].Slug() < data.Cities[j].Slug()
	})
	for _, city := range data.Cities {
		sort.Slice(city.Clubs, func(i, j int) bool {
			return city.Clubs[i].Slug() < city.Clubs[j].Slug()
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
		return len(topCities[i].Clubs) > len(topCities[j].Clubs)
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
			log.Panicf("getting coordinates for city %q: %w", city.Name, err)
		} else {
			city.Coords = &coords
		}
	}
	return nil
}
