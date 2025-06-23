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

import "time"

// DeviceMeasurement is a measurement taken from a PWS device.
type DeviceMeasurement struct {
	LastUpdated time.Time `json:"last_updated"` // Submission time.
	UpdateFreq  float32   `json:"update_freq"`  // Submission frequency in seconds

	BarometricPressure float32 `json:"barometric_pressure"` // Barometric pressure, hPA
	DewPoint           float32 `json:"dew_point"`           // Dew point, in Celsius
	Humidity           float32 `json:"humidity"`            // Outdoor humidity percentage
	IndoorHumidity     float32 `json:"indoor_humidity"`     // Indoor humidity, percentage
	IndoorTemp         float32 `json:"indoor_temp"`         // Indoor temperature in Celsius
	RainPastHour       float32 `json:"rain_past_hour"`      // Rain over past hour, millimeters
	RainToday          float32 `json:"rain_today"`          // Rain over the past 24 hours, millimeters
	Temperature        float32 `json:"temperature"`         // Temperature in Celsius
	WindDirection      float32 `json:"wind_direction"`      // Instantaneous wind direction, 0-360, degrees
	WindGustSpeed      float32 `json:"wind_gust_speed"`     // Current wind gust, KM/h (software-specific time period)
	WindSpeed          float32 `json:"wind_speed"`          // Instantaneous wind speed, KM/h
}

// Exporter is an exporter that exposes metrics from a device.
type Exporter interface {
	ExporterID() string
	HandleDeviceMeasurement(deviceID string, dm *DeviceMeasurement) error

	Run() error
	Close() error
}
