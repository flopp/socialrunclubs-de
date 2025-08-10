package internal

import "github.com/flopp/go-coordsparser"

type LatLon struct {
	Lat float64
	Lon float64
}

func ParseLatLon(input string) (LatLon, error) {
	lat, lon, err := coordsparser.Parse(input)
	if err != nil {
		return LatLon{}, err
	}
	return LatLon{Lat: lat, Lon: lon}, nil
}
