// Copyright (c) 2025 Joshua Sing <joshua@Joshuasing.dev>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package calc

import (
	"fmt"
	"math"
	"testing"
)

func TestAbsoluteHumidity(t *testing.T) {
	tests := []struct {
		name        string
		tempC       float64
		relHumidity float64 // relative humidity as fraction (0.0–1.0)
		want        float64
		wantErr     bool
	}{
		{"Normal 25C 50RH", 25, 0.50, 11.518, false},
		{"Normal 30C 60RH", 30, 0.60, 18.213, false},
		{"Freezing 0C 100RH", 0, 1.00, 4.85, false},
		{"Cold -5C 80RH", -5, 0.80, 2.727, false},
		{"Hot 40C 20RH", 40, 0.20, 10.23, false},
		{"Zero RH", 20, 0.0, 0.0, false},
		{"Full RH", 20, 1.0, 17.27, false},

		// Invalid cases — want error
		{"RH > 1", 20, 1.1, 0, true},
		{"RH < 0", 20, -0.1, 0, true},
		{"Absolute zero temp", -273.16, 0.5, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AbsoluteHumidity(tt.tempC, tt.relHumidity)
			if (err != nil) != tt.wantErr {
				t.Errorf("AbsoluteHumidity(%.2f, %.2f) err = %v, want %v",
					tt.tempC, tt.relHumidity, err, tt.wantErr)
			}
			if !approxEqual(got, tt.want) {
				t.Errorf("AbsoluteHumidity(%.2f, %.2f) = %.2f; want %.2f",
					tt.tempC, tt.relHumidity, got, tt.want)
			}
		})
	}
}

func TestDewPoint(t *testing.T) {
	tests := []struct {
		name        string
		tempC       float64
		relHumidity float64 // relative humidity as fraction (0.0–1.0)
		want        float64
	}{
		// Normal conditions
		{"Mild 20C 50%", 20, 0.50, 9.26},
		{"Warm 30C 70%", 30, 0.70, 23.93},
		{"Hot 35C 40%", 35, 0.40, 19.39},
		{"Cool 10C 90%", 10, 0.90, 8.435},
		{"Freezing 0C 100%", 0, 1.00, 0.00},
		{"Very cold -10C 60%", -10, 0.60, -16.3},
		{"Humid tropics 28C 90%", 28, 0.90, 26.205},

		// Extremes
		{"Dry 25C 10%", 25, 0.10, -8.756},
		{"Super humid 25C 100%", 25, 1.00, 25.00},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DewPoint(tt.tempC, tt.relHumidity)
			if err != nil {
				t.Errorf("DewPoint(%.2f, %.2f) err = %v, want nil",
					tt.tempC, tt.relHumidity, err)
				return
			}
			if !approxEqual(got, tt.want) {
				t.Errorf("DewPoint(%.2f, %.2f) = %v, want %v",
					tt.tempC, tt.relHumidity, got, tt.want)
			}
		})
	}
}

func TestFeelsLike(t *testing.T) {
	tests := []struct {
		name         string
		tempC        float64
		relHumidity  float64
		windSpeedKPH float64
		want         float64
		wantErr      bool
	}{
		// Heat Index cases (temp ≥ 26.7°C)
		{"HeatIndex 30C 70% RH", 30, 0.70, 5, 35.03, false},
		{"HeatIndex 35C 60% RH", 35, 0.60, 10, 45.05, false},

		// Wind Chill cases (temp ≤ 10°C and wind ≥ 4.8 kph)
		{"WindChill 0C 70% RH 10kph", 0, 0.70, 10, -3.29, false},
		{"WindChill 10C 50% RH 20kph", 10, 0.50, 20, 7.38, false},

		// Neither Heat Index nor Wind Chill
		{"Mild 20C 60% RH 5kph", 20, 0.60, 5, 20.00, false},

		// Edge cases
		{"Exact Heat Index threshold", 26.7, 0.50, 5, 27.14, false},
		{"Exact Wind Chill threshold", 10.0, 0.50, 5, 9.76, false},
		{"Below Wind Chill wind", 5.0, 0.50, 3.0, 5.0, false},

		// Invalid input
		{"Invalid RH > 1", 25, 1.2, 5, 0, true},
		{"Invalid RH < 0", 25, -0.1, 5, 0, true},
		{"Negative wind", 25, 0.5, -1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FeelsLike(tt.tempC, tt.relHumidity, tt.windSpeedKPH)
			if (err != nil) != tt.wantErr {
				t.Errorf("FeelsLike(%.2f, %.2f, %.2f) err = %v, want %v",
					tt.tempC, tt.relHumidity, tt.windSpeedKPH, err, tt.wantErr)
				return
			}
			if !approxEqual(got, tt.want) {
				t.Errorf("FeelsLike(%.2f, %.2f, %.2f) = %v, want %v",
					tt.tempC, tt.relHumidity, tt.windSpeedKPH, got, tt.want)
			}
		})
	}
}

func TestApparentTemperatureAU(t *testing.T) {
	tests := []struct {
		name           string
		tempC          float64
		relHumidity    float64
		windSpeedMPS   float64
		solarRadiation float64
		want           float64
		wantErr        bool
	}{
		// BoM reference cases (no solar)
		{"BoM 30C 70%RH 7.2km/h", 30, 0.70, 7.2, 0, 34.36, false},
		{"BoM 25C 50%RH 7.2km/h", 25, 0.50, 7.2, 0, 24.81, false},
		{"BoM 20C 70%RH 7.2km/h", 20, 0.70, 7.2, 0, 19.98, false},
		{"BoM 15C 50%RH 7.2km/h", 15, 0.50, 7.2, 0, 12.40, false},
		{"BoM 10C 50%RH 7.2km/h", 10, 0.50, 7.2, 0, 6.62, false},
		{"BoM 5C 50%RH 7.2km/h", 5, 0.50, 7.2, 0, 1.03, false},
		{"BoM 0C 50%RH 7.2km/h", 0, 0.50, 7.2, 0, -4.39, false},

		// Actual readings from BoM
		{"BoM 10.3C 78.9%RH 6km/h", 10.3, 0.789, 6.00, 0, 8.4, false},

		// Extended sub-zero cases (valid formula usage)
		{"Cold -5C 60%RH 7.2km/h", -5, 0.60, 7.2, 0, -9.56, false},
		{"Very cold -10C 80%RH 3m/s", -10, 0.80, 10.8, 0, -15.34, false},
		{"Extreme cold -20C 90%RH 4m/s", -20, 0.90, 14.4, 0, -26.42, false},
		{"Dry sub-zero -15C 10%RH 2m/s", -15, 0.10, 7.2, 0, -20.33, false},

		// Invalid input handling
		{"Invalid RH > 1", 25, 1.1, 2.0, 0, 0, true},
		{"Invalid RH < 0", 25, -0.1, 2.0, 0, 0, true},
		{"Negative wind", 25, 0.5, -1.0, 0, 0, true},
		{"Negative solar", 25, 0.5, 1.0, -200, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApparentTemperatureAU(tt.tempC, tt.relHumidity, tt.windSpeedMPS, tt.solarRadiation)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApparentTemperatureAU(%.2f, %.2f, %.2f, %.2f) err = %v, want %v",
					tt.tempC, tt.relHumidity, tt.windSpeedMPS, tt.solarRadiation, err, tt.wantErr)
				return
			}
			if !approxEqual(got, tt.want) {
				t.Errorf("ApparentTemperatureAU(%.2f, %.2f, %.2f, %.2f) = %v, want %v",
					tt.tempC, tt.relHumidity, tt.windSpeedMPS, tt.solarRadiation, got, tt.want)
			}
		})
	}
}

func TestHeatIndex(t *testing.T) {
	tests := []struct {
		tempC       float64
		relHumidity float64
		want        float64
		wantErr     bool
	}{
		{
			tempC:       26.66, // 80°F
			relHumidity: 0.4,
			want:        26.62, // 80°F
		},
		{
			tempC:       26.66, // 80°F
			relHumidity: 0.7,
			want:        28.29, // 83°F
		},
		{
			tempC:       31.11, // 88°F
			relHumidity: 0.4,
			want:        31.05, // 88°F
		},
		{
			tempC:       31.11, // 88°F
			relHumidity: 0.8,
			want:        41.24, // 106°F
		},
		{
			tempC:       35.55, // 96°F
			relHumidity: 0.75,
			want:        55.60, // 132°F
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.2fC %.2fRH", tt.tempC, tt.relHumidity), func(t *testing.T) {
			got, err := HeatIndex(tt.tempC, tt.relHumidity)
			if (err != nil) != tt.wantErr {
				t.Errorf("HeatIndex(%.2f, %.2f) err = %v, want %v",
					tt.tempC, tt.relHumidity, err, tt.wantErr)
				return
			}
			if !approxEqual(got, tt.want) {
				t.Errorf("HeatIndex(%.2f, %.2f) got = %v, want %v",
					tt.tempC, tt.relHumidity, got, tt.want)
			}
		})
	}
}

func TestWindChill(t *testing.T) {
	tests := []struct {
		tempC        float64
		windSpeedKPH float64
		want         float64
		wantErr      bool
	}{
		{
			tempC:        20,
			windSpeedKPH: 20,
			want:         20, // only applies <= 10C
		},
		{
			tempC:        10,
			windSpeedKPH: 18,
			want:         7.58,
		},
		{
			tempC:        8,
			windSpeedKPH: 5,
			want:         7.49,
		},
		{
			tempC:        5,
			windSpeedKPH: 5,
			want:         4.09,
		},
		{
			tempC:        5,
			windSpeedKPH: 25,
			want:         0.53,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.2fC %.2fKPH", tt.tempC, tt.windSpeedKPH), func(t *testing.T) {
			got, err := WindChill(tt.tempC, tt.windSpeedKPH)
			if (err != nil) != tt.wantErr {
				t.Errorf("WindChill(%.2f, %.2f) err = %v, want %v",
					tt.tempC, tt.windSpeedKPH, err, tt.wantErr)
				return
			}
			if !approxEqual(got, tt.want) {
				t.Errorf("WindChill(%.2f, %.2f) got = %v, want %v",
					tt.tempC, tt.windSpeedKPH, got, tt.want)
			}
		})
	}
}

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}
