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

// Package exporter provides utilities and types for sensor data exporters.
package exporter

import (
	"time"

	"github.com/joshuasing/pws_exporter/internal/calc"
)

// DeviceMeasurement is a measurement taken from a PWS device.
type DeviceMeasurement struct {
	LastUpdated time.Time `json:"last_updated"` // Submission time.
	UpdateFreq  float32   `json:"update_freq"`  // Submission frequency in seconds

	BarometricPressure float32 `json:"barometric_pressure"` // Barometric pressure (hPA)
	DewPoint           float32 `json:"dew_point"`           // Dew point (°C)
	AbsoluteHumidity   float32 `json:"humidity_abs"`        // Outdoor absolute humidity (g/m³)
	RelativeHumidity   float32 `json:"humidity_rel"`        // Outdoor relative humidity (%)
	IndoorHumidity     float32 `json:"indoor_humidity"`     // Indoor relative humidity (%)
	IndoorTemp         float32 `json:"indoor_temp"`         // Indoor temperature (°C)
	RainPastHour       float32 `json:"rain_past_hour"`      // Rain over past hour (mm)
	RainToday          float32 `json:"rain_today"`          // Rain over the past 24 hours (mm)
	Temperature        float32 `json:"temperature"`         // Temperature (°C)
	WindDirection      float32 `json:"wind_direction"`      // Wind direction (0-360°)
	WindGustSpeed      float32 `json:"wind_gust_speed"`     // Wind gust speed (km/h, software-specific time period)
	WindSpeed          float32 `json:"wind_speed"`          // Wind speed (km/h)

	FeelsLikeTemp   float32 `json:"feels_like_temp"`   // NOAA apparent ("feels like") temperature (°C)
	AUSApparentTemp float32 `json:"aus_apparent_temp"` // Australia (BoM) apparent temperature (°C)
	HeatIndex       float32 `json:"heat_index"`        // Heat Index (NOAA, °C)
	WindChill       float32 `json:"wind_chill"`        // Wind Chill (NOAA, °C)
}

// DeriveMetrics calculates and adds additional metrics to the device
// measurement. Calculated metrics include "feels like" temperatures, heat
// index, wind chill, etc.
//
// Additionally, if AbsoluteHumidity or DewPoint are 0, they will be calculated
// from Temperature and RelativeHumidity.
func DeriveMetrics(dm *DeviceMeasurement) {
	relHumidity := dm.RelativeHumidity / 100 // 0-100 -> 0-1

	if dm.AbsoluteHumidity == 0 {
		dm.AbsoluteHumidity, _ = calc.AbsoluteHumidity(dm.Temperature, relHumidity)
	}
	if dm.DewPoint == 0 {
		dm.DewPoint, _ = calc.DewPoint(dm.Temperature, relHumidity)
	}

	dm.FeelsLikeTemp, _ = calc.FeelsLike(dm.Temperature, relHumidity, dm.WindSpeed)
	dm.AUSApparentTemp, _ = calc.ApparentTemperatureAU(dm.Temperature, relHumidity, dm.WindSpeed, 0)
	dm.HeatIndex, _ = calc.HeatIndex(dm.Temperature, relHumidity)
	dm.WindChill, _ = calc.WindChill(dm.Temperature, dm.WindSpeed)
}

// Exporter is an exporter that exposes metrics from a device.
type Exporter interface {
	ExporterID() string
	HandleDeviceMeasurement(deviceID string, dm *DeviceMeasurement) error

	Run() error
	Close() error
}
