package utils

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	geo "github.com/codingsince1985/geo-golang"
)

type stubGeocoder struct {
	geocode func(address string) (*geo.Location, error)
	calls   int
}

func (s *stubGeocoder) Geocode(address string) (*geo.Location, error) {
	s.calls++
	if s.geocode == nil {
		return nil, fmt.Errorf("unexpected geocode call for %q", address)
	}
	return s.geocode(address)
}

func (s *stubGeocoder) ReverseGeocode(lat, lng float64) (*geo.Address, error) {
	return nil, fmt.Errorf("unexpected reverse geocode call for %f,%f", lat, lng)
}

func TestCachingGeocoderLookupReturnsCachedCoordinates(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "geocoder-cache.json")
	expected := LatLon{Lat: 52.52, Lon: 13.405}

	if err := saveCache(cacheFile, geocoderCacheData{"Berlin": expected}); err != nil {
		t.Fatalf("saveCache() error = %v", err)
	}

	geocoder := &stubGeocoder{}
	cache := &CachingGeocoder{
		cacheFile: cacheFile,
		geocoder:  geocoder,
	}

	coords, err := cache.Lookup("Berlin")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if coords != expected {
		t.Fatalf("Lookup() = %+v, want %+v", coords, expected)
	}
	if geocoder.calls != 0 {
		t.Fatalf("Geocode() calls = %d, want 0", geocoder.calls)
	}
}

func TestCachingGeocoderLookupCachesFetchedCoordinates(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "geocoder-cache.json")
	expected := LatLon{Lat: 48.1351, Lon: 11.5820}

	rateLimiter := make(chan time.Time, 1)
	rateLimiter <- time.Now()

	geocoder := &stubGeocoder{
		geocode: func(address string) (*geo.Location, error) {
			if address != "Munich, Germany" {
				return nil, fmt.Errorf("Geocode() address = %q, want %q", address, "Munich, Germany")
			}
			return &geo.Location{Lat: expected.Lat, Lng: expected.Lon}, nil
		},
	}

	cache := &CachingGeocoder{
		cacheFile:   cacheFile,
		rateLimiter: rateLimiter,
		geocoder:    geocoder,
	}

	coords, err := cache.Lookup("Munich")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if coords != expected {
		t.Fatalf("Lookup() = %+v, want %+v", coords, expected)
	}
	if geocoder.calls != 1 {
		t.Fatalf("Geocode() calls = %d, want 1", geocoder.calls)
	}

	data, err := loadCache(cacheFile)
	if err != nil {
		t.Fatalf("loadCache() error = %v", err)
	}
	if cached := data["Munich"]; cached != expected {
		t.Fatalf("cached coordinates = %+v, want %+v", cached, expected)
	}
}
