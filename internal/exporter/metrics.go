// Copyright (c) 2025 Joshua Sing <joshua@joshuasing.dev>
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

package exporter

import "github.com/prometheus/client_golang/prometheus"

const (
	stationSubsystem = "station"
)

type Metrics struct {
	BarometricPressure *prometheus.GaugeVec
	DewPoint           *prometheus.GaugeVec
	Humidity           *prometheus.GaugeVec
	IndoorHumidity     *prometheus.GaugeVec
	IndoorTemperature  *prometheus.GaugeVec
	RainPastHour       *prometheus.GaugeVec
	Rain               *prometheus.CounterVec
	Temperature        *prometheus.GaugeVec
	WindDirection      *prometheus.GaugeVec
	WindGustSpeed      *prometheus.GaugeVec
	WindSpeed          *prometheus.GaugeVec
}

func newMetrics(namespace string, reg prometheus.Registerer) *Metrics {
	labels := []string{"station_id"}

	m := &Metrics{
		BarometricPressure: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "barometric_pressure_hpa",
			Help:      "Barometric pressure in hectopascals",
		}, labels),
		DewPoint: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "dew_point_celsius",
			Help:      "Dew point in celsius",
		}, labels),
		Humidity: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "humidity_percent",
			Help:      "Humidity percentage (0-1)",
		}, labels),
		IndoorHumidity: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "indoor_humidity_percent",
			Help:      "Indoor humidity percentage (0-1)",
		}, labels),
		IndoorTemperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "indoor_temperature_celsius",
			Help:      "Indoor temperature in Celsius",
		}, labels),
		RainPastHour: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "rain_past_hour_mm",
			Help:      "Amount of rain over the past hour in millimeters",
		}, labels),
		Rain: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "rain_mm",
			Help:      "Rain in millimeters",
		}, labels),
		Temperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "temperature_celsius",
			Help:      "Temperature in Celsius",
		}, labels),
		WindDirection: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "wind_direction_degrees",
			Help:      "Wind direction in degrees",
		}, labels),
		WindGustSpeed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "wind_gust_speed_kph",
			Help:      "Wind gust speed in KM/h",
		}, labels),
		WindSpeed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "wind_speed_kph",
			Help:      "Wind speed in KM/h",
		}, labels),
	}
	reg.MustRegister(
		m.BarometricPressure,
		m.DewPoint,
		m.Humidity,
		m.IndoorHumidity,
		m.IndoorTemperature,
		m.RainPastHour,
		m.Rain,
		m.Temperature,
		m.WindDirection,
		m.WindGustSpeed,
		m.WindSpeed,
	)
	return m
}
