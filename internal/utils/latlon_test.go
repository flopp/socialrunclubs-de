package utils

import (
	"math"
	"testing"
)

func TestDistance(t *testing.T) {
	tests := []struct {
		name        string
		aa          LatLon
		bb          LatLon
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "same location",
			aa:          LatLon{Lat: 52.5200, Lon: 13.4050},
			bb:          LatLon{Lat: 52.5200, Lon: 13.4050},
			expectedMin: 0,
			expectedMax: 0.001,
		},
		{
			name:        "berlin to munich",
			aa:          LatLon{Lat: 52.5200, Lon: 13.4050},
			bb:          LatLon{Lat: 48.1351, Lon: 11.5820},
			expectedMin: 450,
			expectedMax: 550,
		},
		{
			name:        "new york to london",
			aa:          LatLon{Lat: 40.7128, Lon: -74.0060},
			bb:          LatLon{Lat: 51.5074, Lon: -0.1278},
			expectedMin: 5570,
			expectedMax: 5580,
		},
		{
			name:        "equator crossing",
			aa:          LatLon{Lat: 1.0, Lon: 0.0},
			bb:          LatLon{Lat: -1.0, Lon: 0.0},
			expectedMin: 222,
			expectedMax: 223,
		},
		{
			name:        "north pole to south pole",
			aa:          LatLon{Lat: 90.0, Lon: 0.0},
			bb:          LatLon{Lat: -90.0, Lon: 0.0},
			expectedMin: 20015,
			expectedMax: 20016,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Distance(tt.aa, tt.bb)
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("Distance(%+v, %+v) = %f km, want between %f and %f km",
					tt.aa, tt.bb, result, tt.expectedMin, tt.expectedMax)
			}
		})
	}

	t.Run("symmetry", func(t *testing.T) {
		a := LatLon{Lat: 52.5200, Lon: 13.4050}
		b := LatLon{Lat: 48.1351, Lon: 11.5820}
		d1 := Distance(a, b)
		d2 := Distance(b, a)
		if math.Abs(d1-d2) > 0.001 {
			t.Errorf("Distance is not symmetric: Distance(A, B) = %f, Distance(B, A) = %f", d1, d2)
		}
	})
}

func TestParseLatLon(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
		checkFunc func(LatLon) bool
	}{
		{
			name:      "decimal format",
			input:     "52.5200, 13.4050",
			expectErr: false,
			checkFunc: func(ll LatLon) bool {
				return math.Abs(ll.Lat-52.5200) < 0.0001 && math.Abs(ll.Lon-13.4050) < 0.0001
			},
		},
		{
			name:      "invalid format",
			input:     "invalid",
			expectErr: true,
		},
		{
			name:      "empty string",
			input:     "",
			expectErr: true,
		},
		{
			name:      "only latitude",
			input:     "52.5200",
			expectErr: true,
		},
		{
			name:      "negative coordinates",
			input:     "-33.8688, 151.2093",
			expectErr: false,
			checkFunc: func(ll LatLon) bool {
				return math.Abs(ll.Lat-(-33.8688)) < 0.0001 && math.Abs(ll.Lon-151.2093) < 0.0001
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLatLon(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ParseLatLon(%q) expected error, got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseLatLon(%q) got unexpected error: %v", tt.input, err)
				}
				if tt.checkFunc != nil && !tt.checkFunc(result) {
					t.Errorf("ParseLatLon(%q) returned %+v, check failed", tt.input, result)
				}
			}
		})
	}
}

func TestDeg2Rad(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"0 degrees", 0, 0},
		{"180 degrees", 180, math.Pi},
		{"90 degrees", 90, math.Pi / 2},
		{"360 degrees", 360, 2 * math.Pi},
		{"-45 degrees", -45, -math.Pi / 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deg2rad(tt.input)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("deg2rad(%f) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}
