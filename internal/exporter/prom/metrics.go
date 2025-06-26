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

package prom

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/joshuasing/pws_exporter/internal/exporter"
)

const (
	namespace        = "weather"
	stationSubsystem = "station"
)

// Metrics stores measurements from devices as Prometheus metrics.
type Metrics struct {
	BarometricPressure *prometheus.GaugeVec
	DewPoint           *prometheus.GaugeVec
	AbsoluteHumidity   *prometheus.GaugeVec
	Humidity           *prometheus.GaugeVec
	IndoorHumidity     *prometheus.GaugeVec
	IndoorTemperature  *prometheus.GaugeVec
	RainPastHour       *prometheus.GaugeVec
	Rain               *prometheus.CounterVec
	Temperature        *prometheus.GaugeVec
	WindDirection      *prometheus.GaugeVec
	WindGustSpeed      *prometheus.GaugeVec
	WindSpeed          *prometheus.GaugeVec

	FeelsLikeTemp   *prometheus.GaugeVec
	AUSApparentTemp *prometheus.GaugeVec
	HeatIndex       *prometheus.GaugeVec
	WindChill       *prometheus.GaugeVec
}

// NewMetrics creates Prometheus metrics to store measurements in.
func NewMetrics(reg prometheus.Registerer) *Metrics {
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
			Help:      "Dew point in Celsius",
		}, labels),
		AbsoluteHumidity: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "humidity_abs_grams_per_cubic_meter",
			Help:      "Absolute humidity in grams per cubic meter",
		}, labels),
		Humidity: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "humidity_percent",
			Help:      "Relative humidity percentage (0-1)",
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
		FeelsLikeTemp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "feels_like_celsius",
			Help:      "NOAA feels like temperature in Celsius",
		}, labels),
		AUSApparentTemp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "aus_apparent_temp_celsius",
			Help:      "Australia (BoM) apparent temperature in Celsius",
		}, labels),
		HeatIndex: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "heat_index_celsius",
			Help:      "NOAA Heat Index in Celsius",
		}, labels),
		WindChill: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: stationSubsystem,
			Name:      "wind_chill_celsius",
			Help:      "NOAA Wind Chill in Celsius",
		}, labels),
	}
	reg.MustRegister(
		m.BarometricPressure,
		m.DewPoint,
		m.AbsoluteHumidity,
		m.Humidity,
		m.IndoorHumidity,
		m.IndoorTemperature,
		m.RainPastHour,
		m.Rain,
		m.Temperature,
		m.WindDirection,
		m.WindGustSpeed,
		m.WindSpeed,
		m.FeelsLikeTemp,
		m.AUSApparentTemp,
		m.HeatIndex,
		m.WindChill,
	)
	return m
}

// HandleDeviceMeasurement updates the Prometheus metrics with new values from
// the device measurement.
func (m *Metrics) HandleDeviceMeasurement(deviceID string, dm *exporter.DeviceMeasurement) error {
	l := prometheus.Labels{"station_id": deviceID}

	m.BarometricPressure.With(l).Set(float64(dm.BarometricPressure))
	m.DewPoint.With(l).Set(float64(dm.DewPoint))
	m.AbsoluteHumidity.With(l).Set(float64(dm.AbsoluteHumidity))
	m.Humidity.With(l).Set(float64(dm.RelativeHumidity / 100))
	m.IndoorHumidity.With(l).Set(float64(dm.IndoorHumidity / 100))
	m.IndoorTemperature.With(l).Set(float64(dm.IndoorTemp))
	m.RainPastHour.With(l).Set(float64(dm.RainPastHour))
	m.Rain.Delete(l) // Counter state is stored on the station, not in the exporter.
	m.Rain.With(l).Add(float64(dm.RainToday))
	m.Temperature.With(l).Set(float64(dm.Temperature))
	m.WindDirection.With(l).Set(float64(dm.WindDirection))
	m.WindGustSpeed.With(l).Set(float64(dm.WindGustSpeed))
	m.WindSpeed.With(l).Set(float64(dm.WindSpeed))
	m.FeelsLikeTemp.With(l).Set(float64(dm.FeelsLikeTemp))
	m.AUSApparentTemp.With(l).Set(float64(dm.AUSApparentTemp))
	m.HeatIndex.With(l).Set(float64(dm.HeatIndex))
	m.WindChill.With(l).Set(float64(dm.WindChill))

	return nil
}
