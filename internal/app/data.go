package app

import (
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/flopp/socialrunclubs-de/internal"
)

type City struct {
	Name  string
	Clubs []*Club
}

func (c *City) Slug() string {
	return fmt.Sprintf("/%s", internal.SanitizeName(c.Name))
}

type Club struct {
	Name           string
	DescriptionRaw string
	Description    *template.HTML
	CityRaw        string
	City           *City
	LatLonRaw      string
	LatLon         *internal.LatLon
	Instagram      string
	StravaClub     string
	Website        string
	AddedRaw       string
	UpdatedRaw     string
	StatusRaw      string
}

func (c *Club) Slug() string {
	return fmt.Sprintf("/%s/%s", internal.SanitizeName(c.City.Name), internal.SanitizeName(c.Name))
}

type Data struct {
	Now     time.Time
	NowStr  string
	Cities  []*City
	CityMap map[string]*City
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

func processCitiesSheet(sheet internal.Sheet, data *Data) error {
	if len(sheet.Rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}

	required := []string{"NAME"}
	colIdx, err := internal.ValidateColumns(sheet.Rows[0], required)
	if err != nil {
		return err
	}

	for index, row := range sheet.Rows[1:] {
		city := &City{}
		if city.Name, err = getVal("NAME", row, colIdx); err != nil {
			return fmt.Errorf("row %d: %v", index+2, err)
		}
		if _, found := data.CityMap[city.Name]; found {
			continue
		}

		// skip invalid cities
		if city.Name == "" {
			log.Printf("CITIES row %d: empty city name: %q", index+2, city.Name)
			continue
		}

		city.Clubs = make([]*Club, 0)
		data.Cities = append(data.Cities, city)
		data.CityMap[city.Name] = city
	}

	return nil
}

func processClubsSheet(sheet internal.Sheet, data *Data) error {
	if len(sheet.Rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}

	required := []string{"ID", "ADDED", "UPDATED", "STATUS", "NAME", "CITY", "COORDS", "DESCRIPTION", "INSTAGRAM_URL", "STRAVA_URL", "WEBSITE_URL"}
	colIdx, err := internal.ValidateColumns(sheet.Rows[0], required)
	if err != nil {
		return err
	}

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
			latlon, err := internal.ParseLatLon(club.LatLonRaw)
			if err != nil {
				log.Printf("CLUBS row %d: invalid coords: %q", index+2, club.LatLonRaw)
				continue
			}
			club.LatLon = &latlon
		}

		if city, found := data.CityMap[club.CityRaw]; found {
			city.Clubs = append(city.Clubs, club)
			club.City = city
		} else {
			city = &City{
				Name:  club.CityRaw,
				Clubs: []*Club{club},
			}
			data.Cities = append(data.Cities, city)
			data.CityMap[city.Name] = city
			club.City = city
		}
	}

	return nil
}

func GetData(config Config) (*Data, error) {
	data := &Data{
		Now:     time.Now(),
		NowStr:  time.Now().Format("2006-01-02 15:04:05"),
		Cities:  make([]*City, 0),
		CityMap: make(map[string]*City),
	}

	sheets, err := internal.GetSheets(config.Google.APIKey, config.Google.SheetId)
	if err != nil {
		return nil, fmt.Errorf("getting sheets: %v", err)
	}

	// get the clubs and cities sheets
	if len(sheets) != 2 {
		return nil, fmt.Errorf("expected 2 sheets, got %d", len(sheets))
	}
	for _, sheet := range sheets {
		switch sheet.Name {
		case "CLUBS":
			if err := processClubsSheet(sheet, data); err != nil {
				return nil, fmt.Errorf("processing clubs sheet: %v", err)
			}
		case "CITIES":
			if err := processCitiesSheet(sheet, data); err != nil {
				return nil, fmt.Errorf("processing cities sheet: %v", err)
			}
		default:
			return nil, fmt.Errorf("unknown sheet name: %s", sheet.Name)
		}
	}

	return data, nil
}
