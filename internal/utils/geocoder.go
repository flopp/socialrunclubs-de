package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/codingsince1985/geo-golang"
	"github.com/codingsince1985/geo-golang/openstreetmap"
)

type CachingGeocoder struct {
	download    func(url string, targetPath string) error
	cacheFile   string
	rateLimiter <-chan time.Time
	geocoder    geo.Geocoder
}

func NewCachingGeocoder(download func(url string, targetPath string) error, cacheFile string) *CachingGeocoder {
	return &CachingGeocoder{download, cacheFile, time.Tick(time.Second), openstreetmap.Geocoder()}
}

type geocoderCacheData map[string]LatLon

func (g *CachingGeocoder) LookupOSM(address string) (LatLon, error) {
	<-g.rateLimiter

	location, err := g.geocoder.Geocode(address)
	if err != nil {
		return LatLon{}, err
	}
	if location == nil {
		return LatLon{}, fmt.Errorf("cannot get coordinates of '%s'", address)
	}

	return LatLon{location.Lat, location.Lng}, nil
}

func (g *CachingGeocoder) Lookup(city string) (LatLon, error) {
	data := make(geocoderCacheData)
	filePath := g.cacheFile
	if FileExists(filePath) {
		if buf, err := os.ReadFile(filePath); err != nil {
			return LatLon{}, err
		} else {
			if err := json.Unmarshal(buf, &data); err != nil {
				return LatLon{}, err
			}
		}
	}
	if coords, found := data[city]; found {
		return coords, nil
	}

	log.Printf("geocoder: looking up '%s'", city)
	coords, err := g.LookupOSM(fmt.Sprintf("%s, Germany", city))
	if err != nil {
		return LatLon{}, err
	}

	data[city] = coords

	if buf, err := json.Marshal(data); err != nil {
		return coords, err
	} else {
		if err := MakeDir(filepath.Dir(filePath)); err != nil {
			return coords, err
		}
		if err := os.WriteFile(filePath, buf, 0644); err != nil {
			return coords, err
		}
	}

	return coords, nil
}
